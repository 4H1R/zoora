package conversations

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/domain"
)

// mentionRepository is a Phase 1 stub — replaced by the real impl in Phase 3.
// It satisfies domain.ConversationMentionRepository so NewService can be
// wired up before message-mention persistence exists.
type mentionRepository struct{ db *gorm.DB }

// NewMentionRepository returns a no-op domain.ConversationMentionRepository.
// Phase 1 stub — replaced by the real impl in Phase 3.
func NewMentionRepository(db *gorm.DB) domain.ConversationMentionRepository {
	return &mentionRepository{db: db}
}

// CreateMany is a no-op. Phase 1 stub — replaced by the real impl in Phase 3.
func (r *mentionRepository) CreateMany(ctx context.Context, messageID uuid.UUID, userIDs []uuid.UUID) error {
	return nil
}
