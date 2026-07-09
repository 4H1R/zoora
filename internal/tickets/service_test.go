package tickets_test

import (
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/tickets"
)

// ---- in-memory fakes ----

type fakeStore struct {
	mu sync.Mutex

	tickets  map[uuid.UUID]domain.Ticket
	msgs     map[uuid.UUID][]domain.TicketMessage // ticketID -> oldest first
	classes  map[uuid.UUID]domain.Class
	members  map[uuid.UUID]map[uuid.UUID]bool // classID -> userID -> enrolled
	sessions map[uuid.UUID]domain.ClassSession
	rooms    map[uuid.UUID]domain.QuizRoom
	columns  map[uuid.UUID]domain.GradebookColumn
}

func newFakeStore() *fakeStore {
	return &fakeStore{
		tickets:  map[uuid.UUID]domain.Ticket{},
		msgs:     map[uuid.UUID][]domain.TicketMessage{},
		classes:  map[uuid.UUID]domain.Class{},
		members:  map[uuid.UUID]map[uuid.UUID]bool{},
		sessions: map[uuid.UUID]domain.ClassSession{},
		rooms:    map[uuid.UUID]domain.QuizRoom{},
		columns:  map[uuid.UUID]domain.GradebookColumn{},
	}
}

type fakeTicketRepo struct{ s *fakeStore }

func (r fakeTicketRepo) Create(_ context.Context, t *domain.Ticket) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	now := time.Now()
	t.CreatedAt, t.UpdatedAt = now, now
	r.s.tickets[t.ID] = *t
	return nil
}

func (r fakeTicketRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Ticket, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	t, ok := r.s.tickets[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &t, nil
}

func (r fakeTicketRepo) List(_ context.Context, scope domain.TicketListScope, q domain.ListTicketsQuery) ([]domain.Ticket, int64, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	var out []domain.Ticket
	for _, t := range r.s.tickets {
		if scope.OrganizationID != nil && t.OrganizationID != *scope.OrganizationID {
			continue
		}
		if scope.UserID != nil {
			ownsClass := false
			if c, ok := r.s.classes[t.ClassID]; ok && c.UserID == *scope.UserID {
				ownsClass = true
			}
			if t.UserID != *scope.UserID && !ownsClass {
				continue
			}
		}
		if q.ClassID != nil && t.ClassID != *q.ClassID {
			continue
		}
		if q.Status != nil && t.Status != *q.Status {
			continue
		}
		if q.Type != nil && t.Type != *q.Type {
			continue
		}
		out = append(out, t)
	}
	return out, int64(len(out)), nil
}

func (r fakeTicketRepo) SetStatus(_ context.Context, id uuid.UUID, status domain.TicketStatus, closedBy *uuid.UUID, closedAt *time.Time) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	t, ok := r.s.tickets[id]
	if !ok {
		return domain.ErrNotFound
	}
	t.Status = status
	t.ClosedBy = closedBy
	t.ClosedAt = closedAt
	t.UpdatedAt = time.Now()
	r.s.tickets[id] = t
	return nil
}

type fakeMsgRepo struct{ s *fakeStore }

func (r fakeMsgRepo) Create(_ context.Context, m *domain.TicketMessage) error {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	m.CreatedAt = time.Now()
	r.s.msgs[m.TicketID] = append(r.s.msgs[m.TicketID], *m)
	return nil
}

func (r fakeMsgRepo) ListByTicket(_ context.Context, ticketID uuid.UUID) ([]domain.TicketMessage, error) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	return append([]domain.TicketMessage{}, r.s.msgs[ticketID]...), nil
}

type fakeClassLookup struct{ s *fakeStore }

func (l fakeClassLookup) FindByID(_ context.Context, id uuid.UUID) (*domain.Class, error) {
	l.s.mu.Lock()
	defer l.s.mu.Unlock()
	c, ok := l.s.classes[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &c, nil
}

type fakeMemberLookup struct{ s *fakeStore }

func (l fakeMemberLookup) Exists(_ context.Context, classID, userID uuid.UUID) (bool, error) {
	l.s.mu.Lock()
	defer l.s.mu.Unlock()
	return l.s.members[classID][userID], nil
}

type fakeSessionLookup struct{ s *fakeStore }

func (l fakeSessionLookup) FindByID(_ context.Context, id uuid.UUID) (*domain.ClassSession, error) {
	l.s.mu.Lock()
	defer l.s.mu.Unlock()
	sess, ok := l.s.sessions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &sess, nil
}

type fakeRoomLookup struct{ s *fakeStore }

func (l fakeRoomLookup) FindByID(_ context.Context, id uuid.UUID) (*domain.QuizRoom, error) {
	l.s.mu.Lock()
	defer l.s.mu.Unlock()
	room, ok := l.s.rooms[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &room, nil
}

type fakeColumnLookup struct{ s *fakeStore }

func (l fakeColumnLookup) FindByID(_ context.Context, id uuid.UUID) (*domain.GradebookColumn, error) {
	l.s.mu.Lock()
	defer l.s.mu.Unlock()
	col, ok := l.s.columns[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return &col, nil
}

type fakeTransactor struct{}

func (fakeTransactor) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

type notifyCall struct {
	kind      string // "created" | "replied"
	ticketID  uuid.UUID
	recipient uuid.UUID
}

type fakeNotifier struct {
	mu    sync.Mutex
	calls []notifyCall
}

func (n *fakeNotifier) TicketCreated(_ context.Context, t *domain.Ticket, _ string, teacherID uuid.UUID) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls = append(n.calls, notifyCall{kind: "created", ticketID: t.ID, recipient: teacherID})
	return nil
}

func (n *fakeNotifier) TicketReplied(_ context.Context, t *domain.Ticket, _ *domain.TicketMessage, recipientID uuid.UUID) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls = append(n.calls, notifyCall{kind: "replied", ticketID: t.ID, recipient: recipientID})
	return nil
}

// ---- harness ----

type env struct {
	store    *fakeStore
	notifier *fakeNotifier
	svc      domain.TicketService

	orgID   uuid.UUID
	classID uuid.UUID
	teacher uuid.UUID
	student uuid.UUID
}

func newEnv(t *testing.T) *env {
	t.Helper()
	s := newFakeStore()
	n := &fakeNotifier{}
	e := &env{
		store: s, notifier: n,
		orgID: uuid.New(), classID: uuid.New(),
		teacher: uuid.New(), student: uuid.New(),
	}
	s.classes[e.classID] = domain.Class{ID: e.classID, OrganizationID: e.orgID, UserID: e.teacher, Name: "Math"}
	s.members[e.classID] = map[uuid.UUID]bool{e.student: true}
	e.svc = tickets.NewService(
		fakeTicketRepo{s}, fakeMsgRepo{s},
		fakeClassLookup{s}, fakeMemberLookup{s},
		fakeSessionLookup{s}, fakeRoomLookup{s}, fakeColumnLookup{s},
		nil, // media lookup: nil skips attachment validation in unit tests
		fakeTransactor{}, n,
		slog.New(slog.DiscardHandler),
	)
	return e
}

func (e *env) studentCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID: e.student, OrgID: &e.orgID,
		Permissions: []string{string(domain.PermTicketsView)},
	})
}

func (e *env) teacherCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID: e.teacher, OrgID: &e.orgID,
		Permissions: []string{string(domain.PermTicketsView), string(domain.PermTicketsManage)},
	})
}

func (e *env) outsiderCtx() context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID: uuid.New(), OrgID: &e.orgID,
		Permissions: []string{string(domain.PermTicketsView)},
	})
}

func validCreate(e *env) domain.CreateTicketDTO {
	return domain.CreateTicketDTO{
		ClassID: e.classID.String(),
		Title:   "Why is my grade low?",
		Type:    domain.TicketTypeQuestion,
		Body:    "Please explain question 3.",
	}
}

// ---- create tests ----

func TestCreate_HappyPath(t *testing.T) {
	e := newEnv(t)
	tk, err := e.svc.Create(e.studentCtx(), validCreate(e))
	require.NoError(t, err)
	assert.Equal(t, domain.TicketStatusOpen, tk.Status)
	assert.Equal(t, e.student, tk.UserID)
	assert.Equal(t, e.orgID, tk.OrganizationID)

	msgs, _ := fakeMsgRepo{e.store}.ListByTicket(context.Background(), tk.ID)
	require.Len(t, msgs, 1)
	assert.Equal(t, "Please explain question 3.", msgs[0].Body)

	require.Len(t, e.notifier.calls, 1)
	assert.Equal(t, "created", e.notifier.calls[0].kind)
	assert.Equal(t, e.teacher, e.notifier.calls[0].recipient)
}

func TestCreate_RequiresMembership(t *testing.T) {
	e := newEnv(t)
	_, err := e.svc.Create(e.outsiderCtx(), validCreate(e))
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreate_TargetsOnlyForGradeObjection(t *testing.T) {
	e := newEnv(t)
	roomID := uuid.New().String()
	dto := validCreate(e)
	dto.Type = domain.TicketTypeQuestion
	dto.QuizRoomID = &roomID
	_, err := e.svc.Create(e.studentCtx(), dto)
	var verr *domain.ValidationError
	assert.ErrorAs(t, err, &verr)
}

func TestCreate_AtMostOneTarget(t *testing.T) {
	e := newEnv(t)
	roomID, colID := uuid.New().String(), uuid.New().String()
	dto := validCreate(e)
	dto.Type = domain.TicketTypeGradeObjection
	dto.QuizRoomID, dto.GradebookColumnID = &roomID, &colID
	_, err := e.svc.Create(e.studentCtx(), dto)
	var verr *domain.ValidationError
	assert.ErrorAs(t, err, &verr)
}

func TestCreate_QuizRoomMustBelongToClass(t *testing.T) {
	e := newEnv(t)
	otherClassSession := uuid.New()
	e.store.sessions[otherClassSession] = domain.ClassSession{ID: otherClassSession, ClassID: uuid.New()}
	roomID := uuid.New()
	e.store.rooms[roomID] = domain.QuizRoom{ID: roomID, ClassSessionID: otherClassSession}

	rid := roomID.String()
	dto := validCreate(e)
	dto.Type = domain.TicketTypeGradeObjection
	dto.QuizRoomID = &rid
	_, err := e.svc.Create(e.studentCtx(), dto)
	var verr *domain.ValidationError
	assert.ErrorAs(t, err, &verr)
}

func TestCreate_GradeObjectionWithValidColumn(t *testing.T) {
	e := newEnv(t)
	colID := uuid.New()
	e.store.columns[colID] = domain.GradebookColumn{ID: colID, ClassID: e.classID, Title: "Midterm"}

	cid := colID.String()
	dto := validCreate(e)
	dto.Type = domain.TicketTypeGradeObjection
	dto.GradebookColumnID = &cid
	tk, err := e.svc.Create(e.studentCtx(), dto)
	require.NoError(t, err)
	require.NotNil(t, tk.GradebookColumnID)
	assert.Equal(t, colID, *tk.GradebookColumnID)
}

func TestCreate_ColumnMustBelongToClass(t *testing.T) {
	e := newEnv(t)
	colID := uuid.New()
	e.store.columns[colID] = domain.GradebookColumn{ID: colID, ClassID: uuid.New(), Title: "Other"}

	cid := colID.String()
	dto := validCreate(e)
	dto.Type = domain.TicketTypeGradeObjection
	dto.GradebookColumnID = &cid
	_, err := e.svc.Create(e.studentCtx(), dto)
	var verr *domain.ValidationError
	assert.ErrorAs(t, err, &verr)
}
