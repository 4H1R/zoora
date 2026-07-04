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

type Client struct {
	roomClient      *lksdk.RoomServiceClient
	host            string
	publicURL       string
	apiKey          string
	apiSecret       string
	webhookProvider auth.KeyProvider
	logger          *slog.Logger
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

func (c *Client) StartRecording(ctx context.Context, roomName, s3Path string) (string, error) {
	egressClient := lksdk.NewEgressClient(c.host, c.apiKey, c.apiSecret)

	info, err := egressClient.StartRoomCompositeEgress(ctx, &livekit.RoomCompositeEgressRequest{
		RoomName: roomName,
		Output: &livekit.RoomCompositeEgressRequest_File{
			File: &livekit.EncodedFileOutput{
				FileType: livekit.EncodedFileType_MP4,
				Filepath: s3Path,
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
