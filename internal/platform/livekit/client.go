package livekit

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"

	"github.com/4H1R/zoora/internal/config"
)

type Client struct {
	roomClient *lksdk.RoomServiceClient
	host      string
	publicURL string
	apiKey     string
	apiSecret  string
	logger     *slog.Logger
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
		roomClient: roomClient,
		host:       cfg.LiveKitHost,
		publicURL:  publicURL,
		apiKey:     cfg.LiveKitAPIKey,
		apiSecret:  cfg.LiveKitSecret,
		logger:     logger,
	}
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

func (c *Client) GenerateToken(roomName, identity, name string, canPublish, roomAdmin bool) (string, error) {
	at := auth.NewAccessToken(c.apiKey, c.apiSecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}
	grant.SetCanPublish(canPublish)
	grant.SetCanSubscribe(true)
	if roomAdmin {
		grant.RoomAdmin = true
	}

	at.SetVideoGrant(grant).
		SetIdentity(identity).
		SetName(name).
		SetValidFor(24 * time.Hour)

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
		return nil, fmt.Errorf("listing participants: %w", err)
	}
	return resp.Participants, nil
}

func (c *Client) Host() string {
	return c.host
}

func (c *Client) PublicURL() string {
	return c.publicURL
}
