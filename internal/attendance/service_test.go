package attendance_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/attendance"
	"github.com/4H1R/zoora/internal/domain"
)

// --- Mocks ---

type mAttRepo struct{ mock.Mock }

func (m *mAttRepo) Create(ctx context.Context, a *domain.Attendance) error {
	return m.Called(ctx, a).Error(0)
}
func (m *mAttRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Attendance, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Attendance), a.Error(1)
}
func (m *mAttRepo) Update(ctx context.Context, a *domain.Attendance) error {
	return m.Called(ctx, a).Error(0)
}
func (m *mAttRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mAttRepo) ListBySession(ctx context.Context, sessionID uuid.UUID, q domain.ListAttendanceQuery) ([]domain.Attendance, int64, error) {
	a := m.Called(ctx, sessionID, q)
	res, _ := a.Get(0).([]domain.Attendance)
	return res, a.Get(1).(int64), a.Error(2)
}
func (m *mAttRepo) FindBySessionAndUser(ctx context.Context, sessionID, userID uuid.UUID) (*domain.Attendance, error) {
	a := m.Called(ctx, sessionID, userID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Attendance), a.Error(1)
}
func (m *mAttRepo) AdminList(ctx context.Context, q domain.AdminListAttendanceQuery) ([]domain.Attendance, int64, error) {
	a := m.Called(ctx, q)
	res, _ := a.Get(0).([]domain.Attendance)
	return res, a.Get(1).(int64), a.Error(2)
}

type mClassRepo struct{ mock.Mock }

func (m *mClassRepo) Create(ctx context.Context, c *domain.Class) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mClassRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Class), a.Error(1)
}
func (m *mClassRepo) Update(ctx context.Context, c *domain.Class) error {
	return m.Called(ctx, c).Error(0)
}
func (m *mClassRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mClassRepo) List(ctx context.Context, scope domain.ClassListScope, p domain.ListParams) ([]domain.Class, int64, error) {
	a := m.Called(ctx, scope, p)
	res, _ := a.Get(0).([]domain.Class)
	return res, a.Get(1).(int64), a.Error(2)
}
func (m *mClassRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mClassRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.Class, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Class), a.Error(1)
}
func (m *mClassRepo) AdminList(ctx context.Context, q domain.AdminListClassesQuery) ([]domain.Class, int64, error) {
	a := m.Called(ctx, q)
	res, _ := a.Get(0).([]domain.Class)
	return res, a.Get(1).(int64), a.Error(2)
}

type mSessRepo struct{ mock.Mock }

func (m *mSessRepo) Create(ctx context.Context, s *domain.ClassSession) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mSessRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.ClassSession, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.ClassSession), a.Error(1)
}
func (m *mSessRepo) Update(ctx context.Context, s *domain.ClassSession) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mSessRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mSessRepo) ListByClass(ctx context.Context, classID uuid.UUID, q domain.ListClassSessionsQuery) ([]domain.ClassSession, int64, error) {
	a := m.Called(ctx, classID, q)
	res, _ := a.Get(0).([]domain.ClassSession)
	return res, a.Get(1).(int64), a.Error(2)
}
func (m *mSessRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mSessRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.ClassSession, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.ClassSession), a.Error(1)
}
func (m *mSessRepo) AdminList(ctx context.Context, q domain.AdminListClassSessionsQuery) ([]domain.ClassSession, int64, error) {
	a := m.Called(ctx, q)
	res, _ := a.Get(0).([]domain.ClassSession)
	return res, a.Get(1).(int64), a.Error(2)
}

// --- Helpers ---

func newSvc(repo domain.AttendanceRepository, classes domain.ClassRepository, sessions domain.ClassSessionRepository) domain.AttendanceService {
	return attendance.NewService(repo, classes, sessions, nil, nil, nil, nil, nil, slog.Default())
}

func ownerCtx(userID uuid.UUID) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: userID})
}

// --- Tests ---

func TestBulkMark_UpdatesExistingNoCreate(t *testing.T) {
	ownerID := uuid.New()
	classID := uuid.New()
	sessionID := uuid.New()
	userID := uuid.New()
	orgID := uuid.New()

	repo := &mAttRepo{}
	classes := &mClassRepo{}
	sessions := &mSessRepo{}

	classes.On("FindByID", mock.Anything, classID).
		Return(&domain.Class{ID: classID, OrganizationID: orgID, UserID: ownerID}, nil)
	sessions.On("FindByID", mock.Anything, sessionID).
		Return(&domain.ClassSession{ID: sessionID, ClassID: classID}, nil)
	existing := &domain.Attendance{ID: uuid.New(), ClassSessionID: sessionID, UserID: userID, Status: domain.AttendanceStatusAbsent}
	repo.On("FindBySessionAndUser", mock.Anything, sessionID, userID).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Attendance")).Return(nil)

	svc := newSvc(repo, classes, sessions)
	dto := domain.BulkCreateAttendanceDTO{Entries: []domain.CreateAttendanceDTO{{UserID: userID, Status: domain.AttendanceStatusPresent}}}

	res, err := svc.BulkMark(ownerCtx(ownerID), classID, sessionID, dto)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, domain.AttendanceStatusPresent, res[0].Status)
	repo.AssertCalled(t, "Update", mock.Anything, mock.AnythingOfType("*domain.Attendance"))
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestBulkMark_CreatesWhenMissing(t *testing.T) {
	ownerID := uuid.New()
	classID := uuid.New()
	sessionID := uuid.New()
	userID := uuid.New()
	orgID := uuid.New()

	repo := &mAttRepo{}
	classes := &mClassRepo{}
	sessions := &mSessRepo{}

	classes.On("FindByID", mock.Anything, classID).
		Return(&domain.Class{ID: classID, OrganizationID: orgID, UserID: ownerID}, nil)
	sessions.On("FindByID", mock.Anything, sessionID).
		Return(&domain.ClassSession{ID: sessionID, ClassID: classID}, nil)
	repo.On("FindBySessionAndUser", mock.Anything, sessionID, userID).Return(nil, domain.ErrNotFound)
	repo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Attendance")).Return(nil)

	svc := newSvc(repo, classes, sessions)
	dto := domain.BulkCreateAttendanceDTO{Entries: []domain.CreateAttendanceDTO{{UserID: userID, Status: domain.AttendanceStatusPresent}}}

	res, err := svc.BulkMark(ownerCtx(ownerID), classID, sessionID, dto)

	assert.NoError(t, err)
	assert.Len(t, res, 1)
	repo.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("*domain.Attendance"))
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}
