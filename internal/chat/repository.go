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

func NewChatRepository(db *gorm.DB) domain.ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) Create(ctx context.Context, chat *domain.Chat) error {
	if err := database.DB(ctx, r.db).Create(chat).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("chat.repository.Create: %w", err)
	}
	return nil
}

func (r *chatRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Chat, error) {
	var chat domain.Chat
	if err := database.DB(ctx, r.db).First(&chat, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("chat.repository.FindByID: %w", err)
	}
	return &chat, nil
}

func (r *chatRepository) Update(ctx context.Context, chat *domain.Chat) error {
	if err := database.DB(ctx, r.db).Save(chat).Error; err != nil {
		return fmt.Errorf("chat.repository.Update: %w", err)
	}
	return nil
}

func (r *chatRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.Chat{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("chat.repository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *chatRepository) List(ctx context.Context, q domain.ListChatsQuery) ([]domain.Chat, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.Chat{})
	if q.ModelType != "" {
		base = base.Where("model_type = ?", q.ModelType)
	}
	if q.ModelID != nil {
		base = base.Where("model_id = ?", *q.ModelID)
	}
	if q.Status != nil {
		base = base.Where("status = ?", *q.Status)
	}

	var chats []domain.Chat
	total, err := listparams.Paginate(base, q.ListParams, &chats)
	if err != nil {
		return nil, 0, fmt.Errorf("chat.repository.List: %w", err)
	}
	return chats, total, nil
}

func (r *chatRepository) FindByModel(ctx context.Context, modelType string, modelID uuid.UUID) (*domain.Chat, error) {
	var chat domain.Chat
	if err := database.DB(ctx, r.db).
		Where("model_type = ? AND model_id = ?", modelType, modelID).
		First(&chat).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("chat.repository.FindByModel: %w", err)
	}
	return &chat, nil
}

type memberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) domain.ChatMemberRepository {
	return &memberRepository{db: db}
}

func (r *memberRepository) Create(ctx context.Context, member *domain.ChatMember) error {
	if err := database.DB(ctx, r.db).Create(member).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("chat.memberRepository.Create: %w", err)
	}
	return nil
}

func (r *memberRepository) FindByChatAndUser(ctx context.Context, chatID, userID uuid.UUID) (*domain.ChatMember, error) {
	var member domain.ChatMember
	if err := database.DB(ctx, r.db).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("chat.memberRepository.FindByChatAndUser: %w", err)
	}
	return &member, nil
}

func (r *memberRepository) Delete(ctx context.Context, chatID, userID uuid.UUID) error {
	result := database.DB(ctx, r.db).
		Where("chat_id = ? AND user_id = ?", chatID, userID).
		Delete(&domain.ChatMember{})
	if result.Error != nil {
		return fmt.Errorf("chat.memberRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *memberRepository) ListByChat(ctx context.Context, chatID uuid.UUID) ([]domain.ChatMember, error) {
	var members []domain.ChatMember
	if err := database.DB(ctx, r.db).
		Where("chat_id = ?", chatID).
		Order("joined_at ASC").
		Find(&members).Error; err != nil {
		return nil, fmt.Errorf("chat.memberRepository.ListByChat: %w", err)
	}
	return members, nil
}

type messageRepository struct {
	db *gorm.DB
}

func NewMessageRepository(db *gorm.DB) domain.MessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, msg *domain.Message) error {
	if err := database.DB(ctx, r.db).Create(msg).Error; err != nil {
		if database.IsForeignKeyViolation(err) {
			return domain.ErrNotFound
		}
		return fmt.Errorf("chat.messageRepository.Create: %w", err)
	}
	return nil
}

func (r *messageRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Message, error) {
	var msg domain.Message
	if err := database.DB(ctx, r.db).First(&msg, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("chat.messageRepository.FindByID: %w", err)
	}
	return &msg, nil
}

func (r *messageRepository) Update(ctx context.Context, msg *domain.Message) error {
	if err := database.DB(ctx, r.db).Save(msg).Error; err != nil {
		return fmt.Errorf("chat.messageRepository.Update: %w", err)
	}
	return nil
}

func (r *messageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := database.DB(ctx, r.db).Delete(&domain.Message{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("chat.messageRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *messageRepository) List(ctx context.Context, chatID uuid.UUID, q domain.ListMessagesQuery) ([]domain.Message, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.Message{}).Where("chat_id = ?", chatID)
	if q.ParentMessageID != nil {
		base = base.Where("parent_message_id = ?", *q.ParentMessageID)
	} else {
		base = base.Where("parent_message_id IS NULL")
	}

	var messages []domain.Message
	total, err := listparams.Paginate(base, q.ListParams, &messages)
	if err != nil {
		return nil, 0, fmt.Errorf("chat.messageRepository.List: %w", err)
	}
	return messages, total, nil
}

type reactionRepository struct {
	db *gorm.DB
}

func NewReactionRepository(db *gorm.DB) domain.MessageReactionRepository {
	return &reactionRepository{db: db}
}

func (r *reactionRepository) Create(ctx context.Context, reaction *domain.MessageReaction) error {
	if err := database.DB(ctx, r.db).Create(reaction).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("chat.reactionRepository.Create: %w", err)
	}
	return nil
}

func (r *reactionRepository) Delete(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	result := database.DB(ctx, r.db).
		Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).
		Delete(&domain.MessageReaction{})
	if result.Error != nil {
		return fmt.Errorf("chat.reactionRepository.Delete: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *reactionRepository) FindByMessageAndUser(ctx context.Context, messageID, userID uuid.UUID, emoji string) (*domain.MessageReaction, error) {
	var reaction domain.MessageReaction
	if err := database.DB(ctx, r.db).
		Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).
		First(&reaction).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("chat.reactionRepository.FindByMessageAndUser: %w", err)
	}
	return &reaction, nil
}

func (r *reactionRepository) CountByMessage(ctx context.Context, messageID uuid.UUID) (map[string]int, error) {
	type emojiCount struct {
		Emoji string
		Count int
	}
	var results []emojiCount
	if err := database.DB(ctx, r.db).
		Model(&domain.MessageReaction{}).
		Select("emoji, COUNT(*) as count").
		Where("message_id = ?", messageID).
		Group("emoji").
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("chat.reactionRepository.CountByMessage: %w", err)
	}
	counts := make(map[string]int, len(results))
	for _, r := range results {
		counts[r.Emoji] = r.Count
	}
	return counts, nil
}
