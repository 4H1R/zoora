package tickets

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

type ticketRepository struct{ db *gorm.DB }

func NewRepository(db *gorm.DB) domain.TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) Create(ctx context.Context, t *domain.Ticket) error {
	if err := database.DB(ctx, r.db).Create(t).Error; err != nil {
		return fmt.Errorf("tickets.repository.Create: %w", err)
	}
	return nil
}

func (r *ticketRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	var t domain.Ticket
	err := database.DB(ctx, r.db).
		Preload("User").Preload("Class").
		Preload("QuizRoom.Quiz").Preload("GradebookColumn").
		First(&t, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("tickets.repository.FindByID: %w", err)
	}
	return &t, nil
}

// List translates the role-resolved scope into SQL: UserID set = tickets the
// user created OR tickets of classes the user owns (teacher inbox); nil =
// every org ticket (platform admin).
func (r *ticketRepository) List(ctx context.Context, scope domain.TicketListScope, q domain.ListTicketsQuery) ([]domain.Ticket, int64, error) {
	base := database.DB(ctx, r.db).Model(&domain.Ticket{}).
		Preload("User").Preload("Class")
	if scope.OrganizationID != nil {
		base = base.Where("tickets.organization_id = ?", *scope.OrganizationID)
	}
	if scope.UserID != nil {
		base = base.Where(
			"(tickets.user_id = ? OR tickets.class_id IN (SELECT id FROM classes WHERE user_id = ? AND deleted_at IS NULL))",
			*scope.UserID, *scope.UserID,
		)
	}
	if q.ClassID != nil {
		base = base.Where("tickets.class_id = ?", *q.ClassID)
	}
	if q.Status != nil {
		base = base.Where("tickets.status = ?", *q.Status)
	}
	if q.Type != nil {
		base = base.Where("tickets.type = ?", *q.Type)
	}
	var out []domain.Ticket
	total, err := listparams.Paginate(base, q.ListParams, &out)
	if err != nil {
		return nil, 0, fmt.Errorf("tickets.repository.List: %w", err)
	}
	return out, total, nil
}

func (r *ticketRepository) SetStatus(ctx context.Context, id uuid.UUID, status domain.TicketStatus, closedBy *uuid.UUID, closedAt *time.Time) error {
	res := database.DB(ctx, r.db).Model(&domain.Ticket{}).Where("id = ?", id).
		Updates(map[string]any{
			"status":     status,
			"closed_by":  closedBy,
			"closed_at":  closedAt,
			"updated_at": time.Now(),
		})
	if res.Error != nil {
		return fmt.Errorf("tickets.repository.SetStatus: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// ---- Message repository ----

type messageRepository struct{ db *gorm.DB }

func NewMessageRepository(db *gorm.DB) domain.TicketMessageRepository {
	return &messageRepository{db: db}
}

func (r *messageRepository) Create(ctx context.Context, m *domain.TicketMessage) error {
	if err := database.DB(ctx, r.db).Create(m).Error; err != nil {
		return fmt.Errorf("tickets.repository.message.Create: %w", err)
	}
	return nil
}

func (r *messageRepository) ListByTicket(ctx context.Context, ticketID uuid.UUID) ([]domain.TicketMessage, error) {
	var out []domain.TicketMessage
	if err := database.DB(ctx, r.db).Preload("User").
		Where("ticket_id = ?", ticketID).Order("id ASC").Find(&out).Error; err != nil {
		return nil, fmt.Errorf("tickets.repository.message.ListByTicket: %w", err)
	}
	return out, nil
}
