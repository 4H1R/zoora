package tickets

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// Narrow read ports keep this package decoupled from other feature packages:
// main.go satisfies them with the existing domain repositories.

type classLookup interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Class, error)
}

type memberLookup interface {
	Exists(ctx context.Context, classID, userID uuid.UUID) (bool, error)
}

type sessionLookup interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.ClassSession, error)
}

type quizRoomLookup interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.QuizRoom, error)
}

type columnLookup interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.GradebookColumn, error)
}

// mediaLookup validates attachments; nil = skip (unit tests).
type mediaLookup interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Media, error)
}

// notifier is the notifications port; nil = no-op.
type notifier interface {
	TicketCreated(ctx context.Context, t *domain.Ticket, body string, teacherID uuid.UUID) error
	TicketReplied(ctx context.Context, t *domain.Ticket, m *domain.TicketMessage, recipientID uuid.UUID) error
}

type service struct {
	tickets    domain.TicketRepository
	messages   domain.TicketMessageRepository
	classes    classLookup
	members    memberLookup
	sessions   sessionLookup
	quizRooms  quizRoomLookup
	columns    columnLookup
	media      mediaLookup // may be nil
	transactor domain.Transactor
	notif      notifier // may be nil
	logger     *slog.Logger
}

func NewService(
	ticketRepo domain.TicketRepository,
	msgRepo domain.TicketMessageRepository,
	classes classLookup,
	members memberLookup,
	sessions sessionLookup,
	quizRooms quizRoomLookup,
	columns columnLookup,
	media mediaLookup,
	transactor domain.Transactor,
	notif notifier,
	logger *slog.Logger,
) domain.TicketService {
	return &service{ticketRepo, msgRepo, classes, members, sessions, quizRooms, columns, media, transactor, notif, logger}
}

func (s *service) caller(ctx context.Context) (domain.Caller, error) {
	c, ok := domain.CallerFromCtx(ctx)
	if !ok || c.OrgID == nil {
		return domain.Caller{}, domain.ErrForbidden
	}
	return c, nil
}

// isHandler: the class teacher holding tickets:manage, or a platform admin.
// Membership is NOT required — handling rights derive from class ownership.
func isHandler(caller domain.Caller, class *domain.Class) bool {
	if caller.IsAdmin {
		return true
	}
	return class != nil && class.UserID == caller.UserID &&
		caller.HasPermission(domain.PermTicketsManage)
}

// marshalMediaIDs renders the jsonb payload; empty input stays "[]".
func marshalMediaIDs(ids []string) json.RawMessage {
	if len(ids) == 0 {
		return json.RawMessage(`[]`)
	}
	b, err := json.Marshal(ids)
	if err != nil {
		return json.RawMessage(`[]`)
	}
	return b
}

// validateMedia checks each attachment exists, belongs to the caller's org,
// and was presigned for tickets of this class (model_type=ticket,
// model_id=<class id> — see domain.MediaModelTicket). Nil port = skip.
func (s *service) validateMedia(ctx context.Context, caller domain.Caller, classID uuid.UUID, ids []string) error {
	if len(ids) == 0 || s.media == nil {
		return nil
	}
	for _, idStr := range ids {
		mid, err := uuid.Parse(idStr)
		if err != nil {
			return domain.NewValidationError(map[string]string{"media_ids": "invalid uuid"})
		}
		med, err := s.media.FindByID(ctx, mid)
		if err != nil {
			return domain.NewValidationError(map[string]string{"media_ids": "attachment not found"})
		}
		if med.OrganizationID == nil || *med.OrganizationID != *caller.OrgID ||
			med.ModelType != domain.MediaModelTicket || med.ModelID != classID {
			return domain.NewValidationError(map[string]string{"media_ids": "attachment does not belong to this class's tickets"})
		}
	}
	return nil
}

// resolveObjectionTargets validates the grade_objection target rules and
// returns the parsed ids. Non-objection tickets must carry no targets.
func (s *service) resolveObjectionTargets(ctx context.Context, dto domain.CreateTicketDTO, classID uuid.UUID) (quizRoomID, columnID *uuid.UUID, err error) {
	if dto.Type != domain.TicketTypeGradeObjection {
		if dto.QuizRoomID != nil || dto.GradebookColumnID != nil {
			return nil, nil, domain.NewValidationError(map[string]string{"type": "targets are only allowed for grade_objection tickets"})
		}
		return nil, nil, nil
	}
	if dto.QuizRoomID != nil && dto.GradebookColumnID != nil {
		return nil, nil, domain.NewValidationError(map[string]string{"quiz_room_id": "set at most one of quiz_room_id and gradebook_column_id"})
	}
	if dto.QuizRoomID != nil {
		rid, perr := uuid.Parse(*dto.QuizRoomID)
		if perr != nil {
			return nil, nil, domain.NewValidationError(map[string]string{"quiz_room_id": "invalid uuid"})
		}
		room, rerr := s.quizRooms.FindByID(ctx, rid)
		if rerr != nil {
			return nil, nil, domain.NewValidationError(map[string]string{"quiz_room_id": "quiz room not found"})
		}
		sess, serr := s.sessions.FindByID(ctx, room.ClassSessionID)
		if serr != nil || sess.ClassID != classID {
			return nil, nil, domain.NewValidationError(map[string]string{"quiz_room_id": "quiz room does not belong to this class"})
		}
		quizRoomID = &rid
	}
	if dto.GradebookColumnID != nil {
		cid, perr := uuid.Parse(*dto.GradebookColumnID)
		if perr != nil {
			return nil, nil, domain.NewValidationError(map[string]string{"gradebook_column_id": "invalid uuid"})
		}
		col, cerr := s.columns.FindByID(ctx, cid)
		if cerr != nil || col.ClassID != classID {
			return nil, nil, domain.NewValidationError(map[string]string{"gradebook_column_id": "column does not belong to this class"})
		}
		columnID = &cid
	}
	return quizRoomID, columnID, nil
}

func (s *service) Create(ctx context.Context, dto domain.CreateTicketDTO) (*domain.Ticket, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	classID, err := uuid.Parse(dto.ClassID)
	if err != nil {
		return nil, domain.NewValidationError(map[string]string{"class_id": "invalid uuid"})
	}
	class, err := s.classes.FindByID(ctx, classID)
	if err != nil {
		return nil, err
	}
	if class.OrganizationID != *caller.OrgID {
		return nil, domain.ErrForbidden
	}
	// Membership gate — only at creation. Tickets survive later unenroll.
	enrolled, err := s.members.Exists(ctx, classID, caller.UserID)
	if err != nil {
		return nil, err
	}
	if !enrolled {
		return nil, domain.ErrForbidden
	}
	if !dto.Type.Valid() {
		return nil, domain.NewValidationError(map[string]string{"type": "invalid ticket type"})
	}
	quizRoomID, columnID, err := s.resolveObjectionTargets(ctx, dto, classID)
	if err != nil {
		return nil, err
	}
	if err := s.validateMedia(ctx, caller, classID, dto.MediaIDs); err != nil {
		return nil, err
	}

	var t *domain.Ticket
	err = s.transactor.RunInTx(ctx, func(txCtx context.Context) error {
		t = &domain.Ticket{
			OrganizationID:    *caller.OrgID,
			ClassID:           classID,
			UserID:            caller.UserID,
			Title:             dto.Title,
			Type:              dto.Type,
			QuizRoomID:        quizRoomID,
			GradebookColumnID: columnID,
			Status:            domain.TicketStatusOpen,
		}
		if cerr := s.tickets.Create(txCtx, t); cerr != nil {
			return cerr
		}
		return s.messages.Create(txCtx, &domain.TicketMessage{
			TicketID: t.ID,
			UserID:   caller.UserID,
			Body:     dto.Body,
			MediaIDs: marshalMediaIDs(dto.MediaIDs),
		})
	})
	if err != nil {
		return nil, err
	}
	if s.notif != nil && class.UserID != caller.UserID {
		if nerr := s.notif.TicketCreated(ctx, t, dto.Body, class.UserID); nerr != nil {
			s.logger.Error("tickets.Create notify", "ticket_id", t.ID, "error", nerr)
		}
	}
	return t, nil
}

// loadForAccess fetches the ticket + its class and enforces the view tier:
// creator, handler (class teacher with tickets:manage), or platform admin.
func (s *service) loadForAccess(ctx context.Context, caller domain.Caller, id uuid.UUID) (*domain.Ticket, *domain.Class, error) {
	t, err := s.tickets.FindByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if !caller.IsAdmin && t.OrganizationID != *caller.OrgID {
		return nil, nil, domain.ErrForbidden
	}
	class, err := s.classes.FindByID(ctx, t.ClassID)
	if err != nil {
		return nil, nil, err
	}
	if t.UserID != caller.UserID && !isHandler(caller, class) {
		return nil, nil, domain.ErrForbidden
	}
	return t, class, nil
}

func (s *service) Get(ctx context.Context, id uuid.UUID) (*domain.Ticket, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	t, _, err := s.loadForAccess(ctx, caller, id)
	if err != nil {
		return nil, err
	}
	msgs, err := s.messages.ListByTicket(ctx, id)
	if err != nil {
		return nil, err
	}
	t.Messages = msgs
	return t, nil
}

func (s *service) List(ctx context.Context, q domain.ListTicketsQuery) ([]domain.Ticket, int64, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, 0, err
	}
	scope := domain.TicketListScope{OrganizationID: caller.OrgID, UserID: &caller.UserID}
	if caller.IsAdmin {
		scope.UserID = nil // platform admin: every ticket in the org
	}
	return s.tickets.List(ctx, scope, q)
}

func (s *service) AddMessage(ctx context.Context, ticketID uuid.UUID, dto domain.AddTicketMessageDTO) (*domain.TicketMessage, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	t, class, err := s.loadForAccess(ctx, caller, ticketID)
	if err != nil {
		return nil, err
	}
	if t.Status == domain.TicketStatusClosed {
		return nil, domain.NewValidationError(map[string]string{"ticket": "ticket is closed"})
	}
	if err := s.validateMedia(ctx, caller, t.ClassID, dto.MediaIDs); err != nil {
		return nil, err
	}

	msg := &domain.TicketMessage{
		TicketID: ticketID,
		UserID:   caller.UserID,
		Body:     dto.Body,
		MediaIDs: marshalMediaIDs(dto.MediaIDs),
	}
	if err := s.messages.Create(ctx, msg); err != nil {
		return nil, err
	}

	// Status derives from who spoke last: creator reply re-opens, handler
	// reply marks answered. (A creator who is also the teacher counts as
	// creator — their ticket stays open for a real counterparty reply.)
	isCreator := t.UserID == caller.UserID
	newStatus := domain.TicketStatusAnswered
	if isCreator {
		newStatus = domain.TicketStatusOpen
	}
	if err := s.tickets.SetStatus(ctx, ticketID, newStatus, nil, nil); err != nil {
		return nil, err
	}

	recipient := t.UserID // handler replied -> notify creator
	if isCreator {
		recipient = class.UserID // creator replied -> notify teacher
	}
	if s.notif != nil && recipient != caller.UserID {
		if nerr := s.notif.TicketReplied(ctx, t, msg, recipient); nerr != nil {
			s.logger.Error("tickets.AddMessage notify", "ticket_id", ticketID, "error", nerr)
		}
	}
	return msg, nil
}

func (s *service) Close(ctx context.Context, ticketID uuid.UUID) (*domain.Ticket, error) {
	caller, err := s.caller(ctx)
	if err != nil {
		return nil, err
	}
	t, _, err := s.loadForAccess(ctx, caller, ticketID)
	if err != nil {
		return nil, err
	}
	if t.Status == domain.TicketStatusClosed {
		return nil, domain.NewValidationError(map[string]string{"ticket": "ticket is already closed"})
	}
	now := time.Now()
	if err := s.tickets.SetStatus(ctx, ticketID, domain.TicketStatusClosed, &caller.UserID, &now); err != nil {
		return nil, err
	}
	t.Status = domain.TicketStatusClosed
	t.ClosedBy = &caller.UserID
	t.ClosedAt = &now
	return t, nil
}
