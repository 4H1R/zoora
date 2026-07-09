package chat

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// roomDataSender delivers realtime data packets to a LiveKit room. Satisfied by
// *livekit.Client. Kept as a narrow local interface so chat depends only on the
// one method it needs and stays trivially mockable in tests.
type roomDataSender interface {
	SendData(ctx context.Context, roomName string, payload []byte, destinationIdentities []string) error
}

type service struct {
	chatRepo    domain.LiveRoomChatRepository
	messageRepo domain.LiveRoomMessageRepository
	transactor  domain.Transactor
	logger      *slog.Logger
	// livekit and liveRooms power realtime delivery for chat messages. Both may
	// be nil (worker, unit tests) — broadcasts become no-ops in that case.
	livekit   roomDataSender
	liveRooms domain.LiveRoomRepository
}

func NewService(
	chatRepo domain.LiveRoomChatRepository,
	messageRepo domain.LiveRoomMessageRepository,
	transactor domain.Transactor,
	logger *slog.Logger,
	livekit roomDataSender,
	liveRooms domain.LiveRoomRepository,
) domain.LiveRoomChatService {
	return &service{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
		transactor:  transactor,
		logger:      logger,
		livekit:     livekit,
		liveRooms:   liveRooms,
	}
}

const (
	chatEventMessage        = "chat_message"
	chatEventMessageDeleted = "chat_message_deleted"
)

// broadcastToLiveRoom pushes a realtime chat event over the LiveKit data channel
// of the room backing the chat, so participants receive messages instantly
// instead of waiting for the next poll. Server-side fanout means every
// participant (any role) receives — no per-client publish grant required.
//
// Best-effort: the message is already persisted, so failures are logged and
// never surfaced. No-op when LiveKit wiring is absent (e.g. the worker).
func (s *service) broadcastToLiveRoom(ctx context.Context, chat *domain.LiveRoomChat, eventType string, data any) {
	if s.livekit == nil || s.liveRooms == nil {
		return
	}
	room, err := s.liveRooms.FindByID(ctx, chat.LiveRoomID)
	if err != nil {
		s.logger.Error("chat.broadcastToLiveRoom: resolve room", "chat_id", chat.ID.String(), "error", err)
		return
	}
	payload, err := json.Marshal(map[string]any{"type": eventType, "data": data})
	if err != nil {
		s.logger.Error("chat.broadcastToLiveRoom: marshal", "event", eventType, "error", err)
		return
	}
	if err := s.livekit.SendData(ctx, room.LiveKitRoomName, payload, nil); err != nil {
		s.logger.Error("chat.broadcastToLiveRoom: send", "room", room.LiveKitRoomName, "event", eventType, "error", err)
	}
}

// chatMessagePayload mirrors the message shape the frontend renders. Sender name
// comes from the caller, since the freshly-created row has no preloaded user.
func chatMessagePayload(msg *domain.LiveRoomMessage, caller domain.Caller) map[string]any {
	return map[string]any{
		"id":                msg.ID.String(),
		"chat_id":           msg.ChatID.String(),
		"sender_id":         caller.UserID.String(),
		"sender":            map[string]any{"id": caller.UserID.String(), "name": caller.Name},
		"message_type":      string(msg.MessageType),
		"content":           msg.Content,
		"created_at":        msg.CreatedAt,
		"parent_message_id": msg.ParentMessageID,
	}
}

func (s *service) CreateChat(ctx context.Context, dto domain.CreateChatDTO) (*domain.LiveRoomChat, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin && !caller.HasPermission(domain.PermChatsManage) {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin && !caller.HasFeature(domain.FeatureChat) {
		return nil, domain.NewFeatureError(caller.Ent.Plan, domain.FeatureChat)
	}

	liveRoomID, _ := uuid.Parse(dto.LiveRoomID)

	chat := &domain.LiveRoomChat{
		Name:        dto.Name,
		Description: dto.Description,
		LiveRoomID:  liveRoomID,
		Status:      domain.LiveRoomChatStatusActive,
	}

	if err := s.chatRepo.Create(ctx, chat); err != nil {
		return nil, err
	}

	s.logger.Info("chat created",
		"chat_id", chat.ID.String(),
		"live_room_id", chat.LiveRoomID.String(),
		"created_by", caller.UserID.String(),
	)
	return chat, nil
}

func (s *service) GetChat(ctx context.Context, id uuid.UUID) (*domain.LiveRoomChat, error) {
	if _, ok := domain.CallerFromCtx(ctx); !ok {
		return nil, domain.ErrForbidden
	}

	chat, err := s.chatRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return chat, nil
}

func (s *service) UpdateChat(ctx context.Context, id uuid.UUID, dto domain.UpdateChatDTO) (*domain.LiveRoomChat, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	chat, err := s.chatRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !caller.IsAdmin && !caller.HasPermission(domain.PermChatsManage) {
		return nil, domain.ErrForbidden
	}

	if dto.Name != nil {
		chat.Name = *dto.Name
	}
	if dto.Description != nil {
		chat.Description = *dto.Description
	}
	if dto.Status != nil {
		if !dto.Status.Valid() {
			return nil, domain.NewValidationError(map[string]string{"status": "invalid status"})
		}
		chat.Status = *dto.Status
	}

	if err := s.chatRepo.Update(ctx, chat); err != nil {
		return nil, err
	}
	return chat, nil
}

func (s *service) DeleteChat(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}
	if !caller.IsAdmin && !caller.HasPermission(domain.PermChatsManage) {
		return domain.ErrForbidden
	}

	if err := s.chatRepo.Delete(ctx, id); err != nil {
		return err
	}
	s.logger.Info("chat deleted", "chat_id", id.String(), "deleted_by", caller.UserID.String())
	return nil
}

func (s *service) ListChats(ctx context.Context, q domain.ListChatsQuery) ([]domain.LiveRoomChat, int64, error) {
	if _, ok := domain.CallerFromCtx(ctx); !ok {
		return nil, 0, domain.ErrForbidden
	}
	return s.chatRepo.List(ctx, q)
}

func (s *service) SendMessage(ctx context.Context, chatID uuid.UUID, dto domain.SendMessageDTO) (*domain.LiveRoomMessage, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin && !caller.HasFeature(domain.FeatureChat) {
		return nil, domain.NewFeatureError(caller.Ent.Plan, domain.FeatureChat)
	}

	chat, err := s.chatRepo.FindByID(ctx, chatID)
	if err != nil {
		return nil, err
	}
	if chat.Status != domain.LiveRoomChatStatusActive {
		return nil, domain.NewValidationError(map[string]string{"chat": "chat is archived"})
	}

	if !dto.MessageType.Valid() {
		return nil, domain.NewValidationError(map[string]string{"message_type": "invalid message type"})
	}

	msg := &domain.LiveRoomMessage{
		ChatID:      chatID,
		SenderID:    &caller.UserID,
		MessageType: dto.MessageType,
		Content:     dto.Content,
		Attachments: json.RawMessage(`[]`),
	}

	if dto.Attachments != nil {
		msg.Attachments = *dto.Attachments
	}

	if dto.ParentMessageID != nil {
		parentID, _ := uuid.Parse(*dto.ParentMessageID)
		parent, pErr := s.messageRepo.FindByID(ctx, parentID)
		if pErr != nil {
			return nil, domain.NewValidationError(map[string]string{"parent_message_id": "parent message not found"})
		}
		if parent.ChatID != chatID {
			return nil, domain.NewValidationError(map[string]string{"parent_message_id": "parent message belongs to different chat"})
		}
		msg.ParentMessageID = &parentID
	}

	if err := s.messageRepo.Create(ctx, msg); err != nil {
		return nil, err
	}

	s.logger.Info("message sent",
		"message_id", msg.ID.String(),
		"chat_id", chatID.String(),
		"sender_id", caller.UserID.String(),
	)

	s.broadcastToLiveRoom(ctx, chat, chatEventMessage, chatMessagePayload(msg, caller))
	return msg, nil
}

func (s *service) GetMessage(ctx context.Context, id uuid.UUID) (*domain.LiveRoomMessage, error) {
	if _, ok := domain.CallerFromCtx(ctx); !ok {
		return nil, domain.ErrForbidden
	}

	msg, err := s.messageRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *service) UpdateMessage(ctx context.Context, id uuid.UUID, dto domain.UpdateMessageDTO) (*domain.LiveRoomMessage, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	msg, err := s.messageRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if msg.SenderID == nil || *msg.SenderID != caller.UserID {
		if !caller.IsAdmin {
			return nil, domain.ErrForbidden
		}
	}

	msg.Content = dto.Content
	msg.IsEdited = true

	if err := s.messageRepo.Update(ctx, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *service) DeleteMessage(ctx context.Context, id uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}

	msg, err := s.messageRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if msg.SenderID == nil || *msg.SenderID != caller.UserID {
		if !caller.IsAdmin && !caller.HasPermission(domain.PermChatsManage) {
			return domain.ErrForbidden
		}
	}

	if err := s.messageRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Only reload the chat when realtime broadcast is actually wired, to avoid an
	// extra query on the delete path when it would be a no-op anyway.
	if s.livekit != nil && s.liveRooms != nil {
		if chat, cErr := s.chatRepo.FindByID(ctx, msg.ChatID); cErr == nil {
			s.broadcastToLiveRoom(ctx, chat, chatEventMessageDeleted, map[string]any{"id": id.String()})
		}
	}
	return nil
}

func (s *service) ListMessages(ctx context.Context, chatID uuid.UUID, q domain.ListMessagesQuery) ([]domain.LiveRoomMessage, int64, error) {
	if _, ok := domain.CallerFromCtx(ctx); !ok {
		return nil, 0, domain.ErrForbidden
	}

	if _, err := s.chatRepo.FindByID(ctx, chatID); err != nil {
		return nil, 0, err
	}
	return s.messageRepo.List(ctx, chatID, q)
}

// FindChatByRoom and ArchiveByRoom back the live-room lifecycle in
// livesessions (join/end room); the caller has already been authorized to
// act on the room itself, so no additional authz happens here.
func (s *service) FindChatByRoom(ctx context.Context, liveRoomID uuid.UUID) (*domain.LiveRoomChat, error) {
	return s.chatRepo.FindByRoom(ctx, liveRoomID)
}

func (s *service) ArchiveByRoom(ctx context.Context, liveRoomID uuid.UUID) error {
	chat, err := s.chatRepo.FindByRoom(ctx, liveRoomID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}
	chat.Status = domain.LiveRoomChatStatusArchived
	return s.chatRepo.Update(ctx, chat)
}
