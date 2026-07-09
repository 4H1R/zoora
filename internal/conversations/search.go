package conversations

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
)

// SearchGlobal performs a member-scoped, ranked full-text search across all
// conversations the user belongs to within the org.
func (r *messageRepository) SearchGlobal(ctx context.Context, orgID, userID uuid.UUID, q string, limit int) ([]domain.ConversationMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	var out []domain.ConversationMessage
	// Rank is selected as an alias because gorm's Order() only accepts
	// strings / clause.OrderBy — passing gorm.Expr to Order is silently
	// ignored (no ordering at all).
	err := database.DB(ctx, r.db).
		Preload("Sender").
		Select("conversation_messages.*, ts_rank(to_tsvector('simple', conversation_messages.content), plainto_tsquery('simple', ?)) AS rank", q).
		Joins("JOIN conversations c ON c.id = conversation_messages.conversation_id").
		Joins("JOIN conversation_members m ON m.conversation_id = c.id AND m.user_id = ?", userID).
		Where("c.organization_id = ?", orgID).
		Where("to_tsvector('simple', conversation_messages.content) @@ plainto_tsquery('simple', ?)", q).
		Order("rank DESC").
		Limit(limit).
		Find(&out).Error
	if err != nil {
		return nil, fmt.Errorf("conversations.repository.message.SearchGlobal: %w", err)
	}
	return out, nil
}
