package conversations

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/database"
)

type mentionRepository struct{ db *gorm.DB }

// NewMentionRepository returns the real domain.ConversationMentionRepository,
// persisting mention rows for "mentions of me" + notification fan-out.
func NewMentionRepository(db *gorm.DB) domain.ConversationMentionRepository {
	return &mentionRepository{db: db}
}

func (r *mentionRepository) CreateMany(ctx context.Context, messageID uuid.UUID, userIDs []uuid.UUID) error {
	if len(userIDs) == 0 {
		return nil
	}
	now := time.Now()
	rows := make([]domain.ConversationMention, 0, len(userIDs))
	seen := map[uuid.UUID]bool{}
	for _, uid := range userIDs {
		if seen[uid] {
			continue
		}
		seen[uid] = true
		rows = append(rows, domain.ConversationMention{MessageID: messageID, UserID: uid, CreatedAt: now})
	}
	// ON CONFLICT DO NOTHING: idempotent under message resend.
	if err := database.DB(ctx, r.db).
		Clauses(clauseOnConflictDoNothing()).Create(&rows).Error; err != nil {
		return fmt.Errorf("conversations.repository.mention.CreateMany: %w", err)
	}
	return nil
}

func clauseOnConflictDoNothing() clause.Expression { return clause.OnConflict{DoNothing: true} }
