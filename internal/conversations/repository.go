package conversations

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

// ---- Conversation repository ----

type conversationRepository struct{ db *gorm.DB }

func NewConversationRepository(db *gorm.DB) domain.ConversationRepository {
	return &conversationRepository{db: db}
}

func (r *conversationRepository) Create(ctx context.Context, c *domain.Conversation) error {
	if err := database.DB(ctx, r.db).Create(c).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("conversations.repository.Create: %w", err)
	}
	return nil
}

func (r *conversationRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Conversation, error) {
	var c domain.Conversation
	if err := database.DB(ctx, r.db).First(&c, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("conversations.repository.FindByID: %w", err)
	}
	return &c, nil
}

func (r *conversationRepository) FindDirect(ctx context.Context, orgID uuid.UUID, dk string) (*domain.Conversation, error) {
	var c domain.Conversation
	err := database.DB(ctx, r.db).
		Where("organization_id = ? AND direct_key = ?", orgID, dk).
		First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("conversations.repository.FindDirect: %w", err)
	}
	return &c, nil
}

func (r *conversationRepository) Update(ctx context.Context, c *domain.Conversation) error {
	res := database.DB(ctx, r.db).Save(c)
	if res.Error != nil {
		if database.IsUniqueViolation(res.Error) {
			return domain.ErrConflict
		}
		return fmt.Errorf("conversations.repository.Update: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *conversationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := database.DB(ctx, r.db).Delete(&domain.Conversation{}, "id = ?", id)
	if res.Error != nil {
		return fmt.Errorf("conversations.repository.Delete: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *conversationRepository) Touch(ctx context.Context, id uuid.UUID) error {
	res := database.DB(ctx, r.db).Model(&domain.Conversation{}).
		Where("id = ?", id).Update("updated_at", time.Now())
	if res.Error != nil {
		return fmt.Errorf("conversations.repository.Touch: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ListForUser: conversations where the caller is a member, org-scoped.
// Ordering/paging comes from ListParams via the repo-wide listparams.Paginate
// idiom (see internal/classes/repository.go) — the handler's ListConfig
// defaults to updated_at DESC (newest activity first).
func (r *conversationRepository) ListForUser(ctx context.Context, orgID, userID uuid.UUID, q domain.ListConversationsQuery) ([]domain.Conversation, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.Conversation{}).
		Where("conversations.organization_id = ?", orgID).
		Where("EXISTS (SELECT 1 FROM conversation_members m WHERE m.conversation_id = conversations.id AND m.user_id = ?)", userID)
	if q.Type != nil {
		base = base.Where("conversations.type = ?", *q.Type)
	}
	var out []domain.Conversation
	total, err := listparams.Paginate(base, q.ListParams, &out)
	if err != nil {
		return nil, 0, fmt.Errorf("conversations.repository.ListForUser: %w", err)
	}
	return out, total, nil
}

// ---- Member repository ----

type memberRepository struct{ db *gorm.DB }

func NewMemberRepository(db *gorm.DB) domain.ConversationMemberRepository {
	return &memberRepository{db: db}
}

func (r *memberRepository) Create(ctx context.Context, m *domain.ConversationMember) error {
	if err := database.DB(ctx, r.db).Create(m).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("conversations.repository.member.Create: %w", err)
	}
	return nil
}

func (r *memberRepository) CreateMany(ctx context.Context, members []domain.ConversationMember) error {
	if len(members) == 0 {
		return nil
	}
	if err := database.DB(ctx, r.db).Create(&members).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("conversations.repository.member.CreateMany: %w", err)
	}
	return nil
}

func (r *memberRepository) FindByConversationAndUser(ctx context.Context, convID, userID uuid.UUID) (*domain.ConversationMember, error) {
	var m domain.ConversationMember
	err := database.DB(ctx, r.db).
		Where("conversation_id = ? AND user_id = ?", convID, userID).First(&m).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("conversations.repository.member.Find: %w", err)
	}
	return &m, nil
}

func (r *memberRepository) Delete(ctx context.Context, convID, userID uuid.UUID) error {
	res := database.DB(ctx, r.db).
		Where("conversation_id = ? AND user_id = ?", convID, userID).
		Delete(&domain.ConversationMember{})
	if res.Error != nil {
		return fmt.Errorf("conversations.repository.member.Delete: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *memberRepository) ListByConversation(ctx context.Context, convID uuid.UUID) ([]domain.ConversationMember, error) {
	var ms []domain.ConversationMember
	if err := database.DB(ctx, r.db).Preload("User").
		Where("conversation_id = ?", convID).Find(&ms).Error; err != nil {
		return nil, fmt.Errorf("conversations.repository.member.List: %w", err)
	}
	return ms, nil
}

func (r *memberRepository) ListPageMembers(ctx context.Context, convIDs, directIDs []uuid.UUID, viewerID uuid.UUID) ([]domain.ConversationMember, error) {
	if len(convIDs) == 0 {
		return nil, nil
	}
	q := database.DB(ctx, r.db).Preload("User").
		Where("conversation_id IN ?", convIDs)
	if len(directIDs) > 0 {
		q = q.Where("user_id = ? OR conversation_id IN ?", viewerID, directIDs)
	} else {
		q = q.Where("user_id = ?", viewerID)
	}
	var ms []domain.ConversationMember
	if err := q.Find(&ms).Error; err != nil {
		return nil, fmt.Errorf("conversations.repository.member.ListPageMembers: %w", err)
	}
	return ms, nil
}

func (r *memberRepository) ListUserIDs(ctx context.Context, convID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	if err := database.DB(ctx, r.db).Model(&domain.ConversationMember{}).
		Where("conversation_id = ?", convID).Pluck("user_id", &ids).Error; err != nil {
		return nil, fmt.Errorf("conversations.repository.member.ListUserIDs: %w", err)
	}
	return ids, nil
}

func (r *memberRepository) SetLastRead(ctx context.Context, convID, userID, messageID uuid.UUID, at time.Time) error {
	res := database.DB(ctx, r.db).Model(&domain.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", convID, userID).
		Updates(map[string]any{"last_read_message_id": messageID, "last_read_at": at})
	if res.Error != nil {
		return fmt.Errorf("conversations.repository.member.SetLastRead: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *memberRepository) SetMuted(ctx context.Context, convID, userID uuid.UUID, until *time.Time) error {
	res := database.DB(ctx, r.db).Model(&domain.ConversationMember{}).
		Where("conversation_id = ? AND user_id = ?", convID, userID).
		Update("muted_until", until)
	if res.Error != nil {
		return fmt.Errorf("conversations.repository.member.SetMuted: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// UnreadCounts: per-conversation counts of messages newer than the member's
// last_read pointer (keyset on id), excluding the member's own messages
// (sending doesn't bump last_read) — one grouped query for a whole page.
func (r *memberRepository) UnreadCounts(ctx context.Context, userID uuid.UUID, convIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	out := make(map[uuid.UUID]int64, len(convIDs))
	if len(convIDs) == 0 {
		return out, nil
	}
	type row struct {
		ConversationID uuid.UUID
		N              int64
	}
	var rows []row
	err := database.DB(ctx, r.db).Model(&domain.ConversationMessage{}).
		Select("conversation_messages.conversation_id, COUNT(*) AS n").
		Joins(`JOIN conversation_members m ON m.conversation_id = conversation_messages.conversation_id AND m.user_id = ?`, userID).
		Where("conversation_messages.conversation_id IN ?", convIDs).
		Where("conversation_messages.sender_id IS DISTINCT FROM ?", userID).
		Where(`conversation_messages.id > COALESCE(m.last_read_message_id, '00000000-0000-0000-0000-000000000000'::uuid)`).
		Group("conversation_messages.conversation_id").
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("conversations.repository.member.UnreadCounts: %w", err)
	}
	for _, x := range rows {
		out[x.ConversationID] = x.N
	}
	return out, nil
}

// ---- Message repository ----

type messageRepository struct{ db *gorm.DB }

func NewMessageRepository(db *gorm.DB) domain.ConversationMessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, m *domain.ConversationMessage) error {
	if err := database.DB(ctx, r.db).Create(m).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict // idempotency: caller re-fetches
		}
		return fmt.Errorf("conversations.repository.message.Create: %w", err)
	}
	return nil
}

func (r *messageRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.ConversationMessage, error) {
	var m domain.ConversationMessage
	if err := database.DB(ctx, r.db).Preload("Sender").First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("conversations.repository.message.FindByID: %w", err)
	}
	return &m, nil
}

func (r *messageRepository) Update(ctx context.Context, m *domain.ConversationMessage) error {
	res := database.DB(ctx, r.db).Save(m)
	if res.Error != nil {
		return fmt.Errorf("conversations.repository.message.Update: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *messageRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res := database.DB(ctx, r.db).Delete(&domain.ConversationMessage{}, "id = ?", id)
	if res.Error != nil {
		return fmt.Errorf("conversations.repository.message.Delete: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ListWindow implements before/after/around keyset paging on the time-ordered
// uuidv7 id. Results are always returned newest-first (DESC) for the client.
func (r *messageRepository) ListWindow(ctx context.Context, convID uuid.UUID, cur domain.MessageCursor) ([]domain.ConversationMessage, error) {
	limit := cur.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	base := database.DB(ctx, r.db).Preload("Sender").Where("conversation_id = ?", convID)

	var out []domain.ConversationMessage
	switch {
	case cur.Around != nil:
		// older (incl. anchor) + newer, together exactly `limit` rows, then merge
		// desc. older takes limit-half so an even limit isn't over-fetched by one.
		half := limit / 2
		var older, newer []domain.ConversationMessage
		if err := base.Session(&gorm.Session{}).
			Where("id <= ?", *cur.Around).Order("id DESC").Limit(limit - half).Find(&older).Error; err != nil {
			return nil, fmt.Errorf("conversations.repository.message.ListWindow around/older: %w", err)
		}
		if err := base.Session(&gorm.Session{}).
			Where("id > ?", *cur.Around).Order("id ASC").Limit(half).Find(&newer).Error; err != nil {
			return nil, fmt.Errorf("conversations.repository.message.ListWindow around/newer: %w", err)
		}
		// newer is ASC; reverse into DESC and prepend before older (which is DESC).
		for i := len(newer) - 1; i >= 0; i-- {
			out = append(out, newer[i])
		}
		out = append(out, older...)
		return out, nil
	case cur.After != nil:
		if err := base.Where("id > ?", *cur.After).Order("id ASC").Limit(limit).Find(&out).Error; err != nil {
			return nil, fmt.Errorf("conversations.repository.message.ListWindow after: %w", err)
		}
		// return DESC for client consistency
		for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
			out[i], out[j] = out[j], out[i]
		}
		return out, nil
	case cur.Before != nil:
		if err := base.Where("id < ?", *cur.Before).Order("id DESC").Limit(limit).Find(&out).Error; err != nil {
			return nil, fmt.Errorf("conversations.repository.message.ListWindow before: %w", err)
		}
		return out, nil
	default: // latest page
		if err := base.Order("id DESC").Limit(limit).Find(&out).Error; err != nil {
			return nil, fmt.Errorf("conversations.repository.message.ListWindow latest: %w", err)
		}
		return out, nil
	}
}

// LatestByConversation: each conversation's newest message (max uuidv7 id) in
// one DISTINCT ON query.
func (r *messageRepository) LatestByConversation(ctx context.Context, convIDs []uuid.UUID) (map[uuid.UUID]domain.ConversationMessage, error) {
	out := make(map[uuid.UUID]domain.ConversationMessage, len(convIDs))
	if len(convIDs) == 0 {
		return out, nil
	}
	var msgs []domain.ConversationMessage
	err := database.DB(ctx, r.db).
		Select("DISTINCT ON (conversation_id) *").
		Where("conversation_id IN ?", convIDs).
		Order("conversation_id, id DESC").
		Find(&msgs).Error
	if err != nil {
		return nil, fmt.Errorf("conversations.repository.message.LatestByConversation: %w", err)
	}
	for _, m := range msgs {
		out[m.ConversationID] = m
	}
	return out, nil
}

func (r *messageRepository) ListPinned(ctx context.Context, convID uuid.UUID) ([]domain.ConversationMessage, error) {
	var out []domain.ConversationMessage
	if err := database.DB(ctx, r.db).Preload("Sender").
		Where("conversation_id = ? AND is_pinned", convID).Order("pinned_at DESC").Find(&out).Error; err != nil {
		return nil, fmt.Errorf("conversations.repository.message.ListPinned: %w", err)
	}
	return out, nil
}

func (r *messageRepository) SetPinned(ctx context.Context, id uuid.UUID, pinned bool, by *uuid.UUID, at *time.Time) error {
	res := database.DB(ctx, r.db).Model(&domain.ConversationMessage{}).Where("id = ?", id).
		Updates(map[string]any{"is_pinned": pinned, "pinned_by": by, "pinned_at": at})
	if res.Error != nil {
		return fmt.Errorf("conversations.repository.message.SetPinned: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *messageRepository) SearchInConversation(ctx context.Context, convID uuid.UUID, q string, limit int) ([]domain.ConversationMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var out []domain.ConversationMessage
	if err := database.DB(ctx, r.db).
		Where("conversation_id = ? AND content ILIKE ?", convID, "%"+q+"%").
		Order("id DESC").Limit(limit).Find(&out).Error; err != nil {
		return nil, fmt.Errorf("conversations.repository.message.Search: %w", err)
	}
	return out, nil
}

// ---- Reaction repository ----

type reactionRepository struct{ db *gorm.DB }

func NewReactionRepository(db *gorm.DB) domain.ConversationReactionRepository {
	return &reactionRepository{db: db}
}

func (r *reactionRepository) Create(ctx context.Context, x *domain.ConversationMessageReaction) error {
	if err := database.DB(ctx, r.db).Create(x).Error; err != nil {
		if database.IsUniqueViolation(err) {
			return domain.ErrConflict
		}
		return fmt.Errorf("conversations.repository.reaction.Create: %w", err)
	}
	return nil
}

func (r *reactionRepository) Delete(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	if err := database.DB(ctx, r.db).
		Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).
		Delete(&domain.ConversationMessageReaction{}).Error; err != nil {
		return fmt.Errorf("conversations.repository.reaction.Delete: %w", err)
	}
	return nil
}

func (r *reactionRepository) FindByMessageAndUser(ctx context.Context, messageID, userID uuid.UUID, emoji string) (*domain.ConversationMessageReaction, error) {
	var x domain.ConversationMessageReaction
	err := database.DB(ctx, r.db).
		Where("message_id = ? AND user_id = ? AND emoji = ?", messageID, userID, emoji).First(&x).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("conversations.repository.reaction.Find: %w", err)
	}
	return &x, nil
}

func (r *reactionRepository) CountByMessage(ctx context.Context, messageID uuid.UUID) (map[string]int, error) {
	type row struct {
		Emoji string
		N     int
	}
	var rows []row
	if err := database.DB(ctx, r.db).Model(&domain.ConversationMessageReaction{}).
		Select("emoji, COUNT(*) AS n").Where("message_id = ?", messageID).
		Group("emoji").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("conversations.repository.reaction.Count: %w", err)
	}
	out := make(map[string]int, len(rows))
	for _, x := range rows {
		out[x.Emoji] = x.N
	}
	return out, nil
}

// CountByMessages: emoji counts for a whole page of messages in one
// GROUP BY (message_id, emoji) query.
func (r *reactionRepository) CountByMessages(ctx context.Context, messageIDs []uuid.UUID) (map[uuid.UUID]map[string]int, error) {
	out := make(map[uuid.UUID]map[string]int, len(messageIDs))
	if len(messageIDs) == 0 {
		return out, nil
	}
	type row struct {
		MessageID uuid.UUID
		Emoji     string
		N         int
	}
	var rows []row
	if err := database.DB(ctx, r.db).Model(&domain.ConversationMessageReaction{}).
		Select("message_id, emoji, COUNT(*) AS n").
		Where("message_id IN ?", messageIDs).
		Group("message_id, emoji").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("conversations.repository.reaction.CountByMessages: %w", err)
	}
	for _, x := range rows {
		if out[x.MessageID] == nil {
			out[x.MessageID] = map[string]int{}
		}
		out[x.MessageID][x.Emoji] = x.N
	}
	return out, nil
}
