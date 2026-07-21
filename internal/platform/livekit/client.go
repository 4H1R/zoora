package livekit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/webhook"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/twitchtv/twirp"

	"github.com/4H1R/zoora/internal/config"
)

// ErrRoomNotFound reports that the LiveKit server no longer knows the room
// (already torn down by its empty_timeout, or never created). Callers use it
// to distinguish "room gone" from transient API failures.
var ErrRoomNotFound = errors.New("livekit room not found")

// ErrEgressNotActive reports that an egress can no longer be stopped because it
// already reached a terminal state (aborted or completed) on LiveKit's side —
// e.g. it aborted before the user hit stop. Callers treat StopEgress as a no-op
// and reconcile their own record instead of surfacing a hard failure.
var ErrEgressNotActive = errors.New("livekit egress not active")

type Client struct {
	roomClient      *lksdk.RoomServiceClient
	host            string
	publicURL       string
	apiKey          string
	apiSecret       string
	webhookProvider auth.KeyProvider
	// s3* configure where recording egress uploads the finished MP4. Passed to
	// LiveKit in the egress request (S3Upload) so the self-hosted egress worker
	// writes directly to our bucket without its own config file. The endpoint is
	// the internal S3 host (reachable from the egress container), not the public
	// browser-facing one.
	s3Endpoint  string
	s3Region    string
	s3Bucket    string
	s3AccessKey string
	s3Secret    string
	logger      *slog.Logger
}

func NewClient(cfg *config.Config, logger *slog.Logger) *Client {
	roomClient := lksdk.NewRoomServiceClient(cfg.LiveKitHost, cfg.LiveKitAPIKey, cfg.LiveKitSecret)
	logger.Info("LiveKit client initialized",
		"host", cfg.LiveKitHost,
		"public_url_config", cfg.LiveKitPublicURL,
	)

	publicURL := cfg.LiveKitPublicURL
	if publicURL == "" {
		publicURL = cfg.LiveKitHost
	}

	return &Client{
		roomClient:      roomClient,
		host:            cfg.LiveKitHost,
		publicURL:       publicURL,
		apiKey:          cfg.LiveKitAPIKey,
		apiSecret:       cfg.LiveKitSecret,
		webhookProvider: auth.NewSimpleKeyProvider(cfg.LiveKitAPIKey, cfg.LiveKitSecret),
		s3Endpoint:      cfg.S3Endpoint,
		s3Region:        cfg.S3Region,
		s3Bucket:        cfg.S3Bucket,
		s3AccessKey:     cfg.S3AccessKey,
		s3Secret:        cfg.S3SecretKey,
		logger:          logger,
	}
}

// ParseWebhook verifies a LiveKit webhook request's signature (signed with the
// same API key/secret pair) and returns the decoded event. A non-nil error
// means the request is unauthenticated or malformed and must be rejected.
func (c *Client) ParseWebhook(r *http.Request) (*livekit.WebhookEvent, error) {
	event, err := webhook.ReceiveWebhookEvent(r, c.webhookProvider)
	if err != nil {
		return nil, fmt.Errorf("verifying LiveKit webhook: %w", err)
	}
	return event, nil
}

func (c *Client) CreateRoom(ctx context.Context, roomName string, maxParticipants uint32) (*livekit.Room, error) {
	if maxParticipants == 0 {
		maxParticipants = 100
	}
	room, err := c.roomClient.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:            roomName,
		EmptyTimeout:    600,
		MaxParticipants: maxParticipants,
	})
	if err != nil {
		return nil, fmt.Errorf("creating LiveKit room: %w", err)
	}
	c.logger.Info("LiveKit room created", "room", roomName)
	return room, nil
}

func (c *Client) DeleteRoom(ctx context.Context, roomName string) error {
	_, err := c.roomClient.DeleteRoom(ctx, &livekit.DeleteRoomRequest{
		Room: roomName,
	})
	if err != nil {
		return fmt.Errorf("deleting LiveKit room: %w", err)
	}
	return nil
}

// GenerateToken mints a join token. Publishable sources are granted explicitly
// via CanPublishSources so room config (mic/camera/screen) is actually enforced
// — an empty slice means subscribe-only. Passing sources also guarantees screen
// share is authorized for moderators (previously CanPublish-only left it implicit).
func (c *Client) GenerateToken(roomName, identity, name, metadata string, sources []livekit.TrackSource, roomAdmin bool) (string, error) {
	at := auth.NewAccessToken(c.apiKey, c.apiSecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}
	if len(sources) > 0 {
		grant.SetCanPublish(true)
		grant.SetCanPublishSources(sources)
	} else {
		grant.SetCanPublish(false)
	}
	grant.SetCanSubscribe(true)
	grant.SetCanPublishData(true)
	// Lets each client publish its own device/OS/browser + live network stats
	// into its participant attributes, which hosts read in the People panel.
	grant.SetCanUpdateOwnMetadata(true)
	if roomAdmin {
		grant.RoomAdmin = true
	}

	at.SetVideoGrant(grant).
		SetIdentity(identity).
		SetName(name).
		SetValidFor(24 * time.Hour)
	if metadata != "" {
		at.SetMetadata(metadata)
	}

	token, err := at.ToJWT()
	if err != nil {
		return "", fmt.Errorf("generating LiveKit token: %w", err)
	}
	return token, nil
}

// StartRecording launches a Room Composite egress: LiveKit renders the room via
// a headless Chrome template, encodes it to a single MP4, and uploads it straight
// to our S3 bucket at s3Path. The 720p30 preset keeps files (and encode cost)
// down versus the 1080p default. Returns the egress ID used to stop/track it.
func (c *Client) StartRecording(ctx context.Context, roomName, s3Path string) (string, error) {
	egressClient := lksdk.NewEgressClient(c.host, c.apiKey, c.apiSecret)

	info, err := egressClient.StartRoomCompositeEgress(ctx, &livekit.RoomCompositeEgressRequest{
		RoomName: roomName,
		// H264 720p30 — smaller output and lighter encode than the 1080p default,
		// which matters since each composite runs Chrome + an encoder.
		Options: &livekit.RoomCompositeEgressRequest_Preset{
			Preset: livekit.EncodingOptionsPreset_H264_720P_30,
		},
		FileOutputs: []*livekit.EncodedFileOutput{
			{
				FileType: livekit.EncodedFileType_MP4,
				Filepath: s3Path,
				// Upload directly to our bucket. ForcePathStyle matches RustFS/MinIO
				// (bucket in the path, not the host). Endpoint is the internal S3
				// host the egress worker can reach.
				Output: &livekit.EncodedFileOutput_S3{
					S3: &livekit.S3Upload{
						AccessKey:      c.s3AccessKey,
						Secret:         c.s3Secret,
						Region:         c.s3Region,
						Endpoint:       c.s3Endpoint,
						Bucket:         c.s3Bucket,
						ForcePathStyle: true,
					},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("starting recording: %w", err)
	}

	c.logger.Info("recording started", "room", roomName, "egress_id", info.EgressId)
	return info.EgressId, nil
}

func (c *Client) StopRecording(ctx context.Context, egressID string) error {
	egressClient := lksdk.NewEgressClient(c.host, c.apiKey, c.apiSecret)
	_, err := egressClient.StopEgress(ctx, &livekit.StopEgressRequest{
		EgressId: egressID,
	})
	if err != nil {
		// A terminal egress (already aborted/completed) rejects StopEgress with
		// failed_precondition ("egress with status ... cannot be stopped").
		// Surface a typed sentinel so the caller can reconcile instead of 500ing.
		var terr twirp.Error
		if errors.As(err, &terr) && terr.Code() == twirp.FailedPrecondition {
			return fmt.Errorf("stopping recording %s: %w", egressID, ErrEgressNotActive)
		}
		return fmt.Errorf("stopping recording: %w", err)
	}
	c.logger.Info("recording stopped", "egress_id", egressID)
	return nil
}

func (c *Client) ListParticipants(ctx context.Context, roomName string) ([]*livekit.ParticipantInfo, error) {
	resp, err := c.roomClient.ListParticipants(ctx, &livekit.ListParticipantsRequest{
		Room: roomName,
	})
	if err != nil {
		var terr twirp.Error
		if errors.As(err, &terr) && terr.Code() == twirp.NotFound {
			return nil, fmt.Errorf("listing participants in %s: %w", roomName, ErrRoomNotFound)
		}
		return nil, fmt.Errorf("listing participants: %w", err)
	}
	return resp.Participants, nil
}

func (c *Client) UpdateParticipant(ctx context.Context, roomName, identity, metadata string, sources []livekit.TrackSource) error {
	canPublish := len(sources) > 0
	req := &livekit.UpdateParticipantRequest{
		Room:     roomName,
		Identity: identity,
		Permission: &livekit.ParticipantPermission{
			CanSubscribe:      true,
			CanPublish:        canPublish,
			CanPublishData:    true,
			CanPublishSources: sources,
			// Preserve self-attribute publishing (device/network presence) across
			// role changes — this permission set replaces the token grant.
			CanUpdateMetadata: true,
		},
	}
	if metadata != "" {
		req.Metadata = metadata
	}
	_, err := c.roomClient.UpdateParticipant(ctx, req)
	if err != nil {
		return fmt.Errorf("updating participant %s in %s: %w", identity, roomName, err)
	}
	return nil
}

func (c *Client) RemoveParticipant(ctx context.Context, roomName, identity string) error {
	_, err := c.roomClient.RemoveParticipant(ctx, &livekit.RoomParticipantIdentity{
		Room:     roomName,
		Identity: identity,
	})
	if err != nil {
		var terr twirp.Error
		if errors.As(err, &terr) && terr.Code() == twirp.NotFound {
			// Room or participant already gone — treat as success (idempotent kick).
			return nil
		}
		return fmt.Errorf("removing participant %s from %s: %w", identity, roomName, err)
	}
	return nil
}

func (c *Client) MutePublishedTrack(ctx context.Context, roomName, identity, trackSID string, muted bool) error {
	_, err := c.roomClient.MutePublishedTrack(ctx, &livekit.MuteRoomTrackRequest{
		Room:     roomName,
		Identity: identity,
		TrackSid: trackSID,
		Muted:    muted,
	})
	if err != nil {
		return fmt.Errorf("muting track %s of %s: %w", trackSID, identity, err)
	}
	return nil
}

func (c *Client) SendData(ctx context.Context, roomName string, payload []byte, destinationIdentities []string) error {
	_, err := c.roomClient.SendData(ctx, &livekit.SendDataRequest{
		Room:                  roomName,
		Data:                  payload,
		Kind:                  livekit.DataPacket_RELIABLE,
		DestinationIdentities: destinationIdentities,
	})
	if err != nil {
		return fmt.Errorf("sending data to %s: %w", roomName, err)
	}
	return nil
}

func (c *Client) Host() string {
	return c.host
}

func (c *Client) PublicURL() string {
	return c.publicURL
}
