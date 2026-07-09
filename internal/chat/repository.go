package chat

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) domain.LiveRoomChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) Create(ctx context.Context, chat *domain.LiveRoomChat) error {
	if err := database.DB(ctx, r.db).Create(chat).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		if database.IsForeignKeyViolation(err) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("chat.repository.Create: %w", err)
	}
	return nil
}

func (r *chatRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.LiveRoomChat, error) {
	var chat domain.LiveRoomChat
	if err := database.DB(ctx, r.db).First(&chat, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("chat.repository.FindByID: %w", err)
	}
	return &chat, nil
}

func (r *chatRepository) Update(ctx context.Context, chat *domain.LiveRoomChat) error {
	if err := database.DB(ctx, r.db).Save(chat).Error; err != nil {
		return fmt.Errorf("chat.repository.Update: %w", err)
	}
	return nil
}

func (r *chatRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.LiveRoomChat{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("chat.repository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *chatRepository) List(ctx context.Context, q domain.ListChatsQuery) ([]domain.LiveRoomChat, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.LiveRoomChat{})
	if q.LiveRoomID != nil {
		base = base.Where("live_room_id = ?", *q.LiveRoomID)
	}
	if q.Status != nil {
		base = base.Where("status = ?", *q.Status)
	}

	var chats []domain.LiveRoomChat
	total, err := listparams.Paginate(base, q.ListParams, &chats)
	if err != nil {
		return nil, 0, fmt.Errorf("chat.repository.List: %w", err)
	}
	return chats, total, nil
}

func (r *chatRepository) FindByRoom(ctx context.Context, liveRoomID uuid.UUID) (*domain.LiveRoomChat, error) {
	var chat domain.LiveRoomChat
	if err := database.DB(ctx, r.db).
		Where("live_room_id = ?", liveRoomID).
		First(&chat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("chat.repository.FindByRoom: %w", err)
	}
	return &chat, nil
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) domain.LiveRoomMessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, msg *domain.LiveRoomMessage) error {
	if err := database.DB(ctx, r.db).Create(msg).Error; err != nil {
		if database.IsForeignKeyViolation(err) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("chat.messageRepository.Create: %w", err)
	}
	return nil
}

func (r *messageRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.LiveRoomMessage, error) {
	var msg domain.LiveRoomMessage
	if err := database.DB(ctx, r.db).First(&msg, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("chat.messageRepository.FindByID: %w", err)
	}
	return &msg, nil
}

func (r *messageRepository) Update(ctx context.Context, msg *domain.LiveRoomMessage) error {
	if err := database.DB(ctx, r.db).Save(msg).Error; err != nil {
		return fmt.Errorf("chat.messageRepository.Update: %w", err)
	}
	return nil
}

func (r *messageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.LiveRoomMessage{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("chat.messageRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *messageRepository) List(ctx context.Context, chatID uuid.UUID, q domain.ListMessagesQuery) ([]domain.LiveRoomMessage, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.LiveRoomMessage{}).Preload("Sender").Where("chat_id = ?", chatID)
	if q.ParentMessageID != nil {
		base = base.Where("parent_message_id = ?", *q.ParentMessageID)
	} else {
		base = base.Where("parent_message_id IS NULL")
	}

	var messages []domain.LiveRoomMessage
	total, err := listparams.Paginate(base, q.ListParams, &messages)
	if err != nil {
		return nil, 0, fmt.Errorf("chat.messageRepository.List: %w", err)
	}
	return messages, total, nil
}
