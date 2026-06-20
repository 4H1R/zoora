package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type service struct {
	chatRepo     domain.ChatRepository
	memberRepo   domain.ChatMemberRepository
	messageRepo  domain.MessageRepository
	reactionRepo domain.MessageReactionRepository
	transactor   domain.Transactor
	logger       *slog.Logger
}

func NewService(
	chatRepo domain.ChatRepository,
	memberRepo domain.ChatMemberRepository,
	messageRepo domain.MessageRepository,
	reactionRepo domain.MessageReactionRepository,
	transactor domain.Transactor,
	logger *slog.Logger,
) domain.ChatService {
	return &service{
		chatRepo:     chatRepo,
		memberRepo:   memberRepo,
		messageRepo:  messageRepo,
		reactionRepo: reactionRepo,
		transactor:   transactor,
		logger:       logger,
	}
}

func (s *service) CreateChat(ctx context.Context, dto domain.CreateChatDTO) (*domain.Chat, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}
	if !caller.IsAdmin && !caller.HasPermission(domain.PermChatsManage) {
		return nil, domain.ErrForbidden
	}

	modelID, _ := uuid.Parse(dto.ModelID)

	chat := &domain.Chat{
		Name:        dto.Name,
		Description: dto.Description,
		ModelType:   dto.ModelType,
		ModelID:     modelID,
		Status:      domain.ChatStatusActive,
	}

	if err := s.chatRepo.Create(ctx, chat); err != nil {
		return nil, err
	}

	s.logger.Info("chat created",
		"chat_id", chat.ID.String(),
		"model_type", chat.ModelType,
		"model_id", chat.ModelID.String(),
		"created_by", caller.UserID.String(),
	)
	return chat, nil
}

func (s *service) GetChat(ctx context.Context, id uuid.UUID) (*domain.Chat, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	chat, err := s.chatRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.checkAccess(ctx, chat, caller); err != nil {
		return nil, err
	}
	return chat, nil
}

func (s *service) UpdateChat(ctx context.Context, id uuid.UUID, dto domain.UpdateChatDTO) (*domain.Chat, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	chat, err := s.chatRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if !caller.IsAdmin && !caller.HasPermission(domain.PermChatsManage) {
		member, mErr := s.memberRepo.FindByChatAndUser(ctx, id, caller.UserID)
		if mErr != nil || member.Role != domain.ChatMemberRoleAdmin {
			return nil, domain.ErrForbidden
		}
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

func (s *service) ListChats(ctx context.Context, q domain.ListChatsQuery) ([]domain.Chat, int64, error) {
	_, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}
	return s.chatRepo.List(ctx, q)
}

// Members

func (s *service) AddMember(ctx context.Context, chatID uuid.UUID, dto domain.AddChatMemberDTO) (*domain.ChatMember, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	if _, err := s.chatRepo.FindByID(ctx, chatID); err != nil {
		return nil, err
	}

	if !caller.IsAdmin && !caller.HasPermission(domain.PermChatsManage) {
		member, mErr := s.memberRepo.FindByChatAndUser(ctx, chatID, caller.UserID)
		if mErr != nil || member.Role != domain.ChatMemberRoleAdmin {
			return nil, domain.ErrForbidden
		}
	}

	if !dto.Role.Valid() {
		return nil, domain.NewValidationError(map[string]string{"role": "invalid role"})
	}

	userID, _ := uuid.Parse(dto.UserID)
	m := &domain.ChatMember{
		ChatID:   chatID,
		UserID:   userID,
		Role:     dto.Role,
		JoinedAt: time.Now(),
	}

	if err := s.memberRepo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *service) RemoveMember(ctx context.Context, chatID, userID uuid.UUID) error {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return domain.ErrForbidden
	}

	if !caller.IsAdmin && !caller.HasPermission(domain.PermChatsManage) && caller.UserID != userID {
		member, mErr := s.memberRepo.FindByChatAndUser(ctx, chatID, caller.UserID)
		if mErr != nil || member.Role != domain.ChatMemberRoleAdmin {
			return domain.ErrForbidden
		}
	}

	return s.memberRepo.Delete(ctx, chatID, userID)
}

func (s *service) ListMembers(ctx context.Context, chatID uuid.UUID) ([]domain.ChatMember, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	chat, err := s.chatRepo.FindByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if err := s.checkAccess(ctx, chat, caller); err != nil {
		return nil, err
	}
	return s.memberRepo.ListByChat(ctx, chatID)
}

// Messages

func (s *service) SendMessage(ctx context.Context, chatID uuid.UUID, dto domain.SendMessageDTO) (*domain.Message, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	chat, err := s.chatRepo.FindByID(ctx, chatID)
	if err != nil {
		return nil, err
	}
	if chat.Status != domain.ChatStatusActive {
		return nil, domain.NewValidationError(map[string]string{"chat": "chat is archived"})
	}

	if err := s.checkWriteAccess(ctx, chat, caller); err != nil {
		return nil, err
	}

	if !dto.MessageType.Valid() {
		return nil, domain.NewValidationError(map[string]string{"message_type": "invalid message type"})
	}

	msg := &domain.Message{
		ChatID:      chatID,
		SenderID:    &caller.UserID,
		MessageType: dto.MessageType,
		Content:     dto.Content,
		Attachments: json.RawMessage(`[]`),
		EmojiCounts: json.RawMessage(`{}`),
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
	return msg, nil
}

func (s *service) GetMessage(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	msg, err := s.messageRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	chat, err := s.chatRepo.FindByID(ctx, msg.ChatID)
	if err != nil {
		return nil, err
	}

	if err := s.checkAccess(ctx, chat, caller); err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *service) UpdateMessage(ctx context.Context, id uuid.UUID, dto domain.UpdateMessageDTO) (*domain.Message, error) {
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
			chat, cErr := s.chatRepo.FindByID(ctx, msg.ChatID)
			if cErr != nil {
				return cErr
			}
			member, mErr := s.memberRepo.FindByChatAndUser(ctx, chat.ID, caller.UserID)
			if mErr != nil || member.Role != domain.ChatMemberRoleAdmin {
				return domain.ErrForbidden
			}
		}
	}

	return s.messageRepo.Delete(ctx, id)
}

func (s *service) ListMessages(ctx context.Context, chatID uuid.UUID, q domain.ListMessagesQuery) ([]domain.Message, int64, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, 0, domain.ErrForbidden
	}

	chat, err := s.chatRepo.FindByID(ctx, chatID)
	if err != nil {
		return nil, 0, err
	}

	if err := s.checkAccess(ctx, chat, caller); err != nil {
		return nil, 0, err
	}
	return s.messageRepo.List(ctx, chatID, q)
}

// Reactions

func (s *service) ToggleReaction(ctx context.Context, messageID uuid.UUID, dto domain.ToggleReactionDTO) (*domain.Message, error) {
	caller, ok := domain.CallerFromCtx(ctx)
	if !ok {
		return nil, domain.ErrForbidden
	}

	msg, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	chat, err := s.chatRepo.FindByID(ctx, msg.ChatID)
	if err != nil {
		return nil, err
	}

	if err := s.checkAccess(ctx, chat, caller); err != nil {
		return nil, err
	}

	err = s.transactor.RunInTx(ctx, func(txCtx context.Context) error {
		existing, findErr := s.reactionRepo.FindByMessageAndUser(txCtx, messageID, caller.UserID, dto.Emoji)
		if findErr != nil && !errors.Is(findErr, domain.ErrNotFound) {
			return findErr
		}

		if existing != nil {
			if err := s.reactionRepo.Delete(txCtx, messageID, caller.UserID, dto.Emoji); err != nil {
				return err
			}
		} else {
			reaction := &domain.MessageReaction{
				MessageID: messageID,
				UserID:    caller.UserID,
				Emoji:     dto.Emoji,
				CreatedAt: time.Now(),
			}
			if err := s.reactionRepo.Create(txCtx, reaction); err != nil {
				return err
			}
		}

		counts, err := s.reactionRepo.CountByMessage(txCtx, messageID)
		if err != nil {
			return err
		}

		countsJSON, err := json.Marshal(counts)
		if err != nil {
			return fmt.Errorf("chat.service.ToggleReaction marshal: %w", err)
		}

		msg.EmojiCounts = countsJSON
		return s.messageRepo.Update(txCtx, msg)
	})
	if err != nil {
		return nil, err
	}

	return msg, nil
}

// Cross-feature helpers

func (s *service) FindChatByModel(ctx context.Context, modelType string, modelID uuid.UUID) (*domain.Chat, error) {
	return s.chatRepo.FindByModel(ctx, modelType, modelID)
}

func (s *service) ArchiveByModel(ctx context.Context, modelType string, modelID uuid.UUID) error {
	chat, err := s.chatRepo.FindByModel(ctx, modelType, modelID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil
		}
		return err
	}
	chat.Status = domain.ChatStatusArchived
	return s.chatRepo.Update(ctx, chat)
}

// Access control helpers

func (s *service) checkAccess(ctx context.Context, chat *domain.Chat, caller domain.Caller) error {
	if caller.IsAdmin || caller.HasPermission(domain.PermChatsManage) {
		return nil
	}
	if chat.ModelType == "live_session" {
		return nil
	}
	_, err := s.memberRepo.FindByChatAndUser(ctx, chat.ID, caller.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrForbidden
		}
		return err
	}
	return nil
}

func (s *service) checkWriteAccess(ctx context.Context, chat *domain.Chat, caller domain.Caller) error {
	if caller.IsAdmin || caller.HasPermission(domain.PermChatsManage) {
		return nil
	}
	if chat.ModelType == "live_session" {
		return nil
	}
	member, err := s.memberRepo.FindByChatAndUser(ctx, chat.ID, caller.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrForbidden
		}
		return err
	}
	if member.Role == domain.ChatMemberRoleReadOnly {
		return domain.ErrForbidden
	}
	return nil
}
