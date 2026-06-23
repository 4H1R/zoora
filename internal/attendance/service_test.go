package attendance_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/attendance"
	"github.com/4H1R/zoora/internal/platform/authz"
	"github.com/4H1R/zoora/internal/domain"
)

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
func (m *mAttRepo) ListByUser(ctx context.Context, userID uuid.UUID, p domain.ListParams) ([]domain.Attendance, int64, error) {
	a := m.Called(ctx, userID, p)
	res, _ := a.Get(0).([]domain.Attendance)
	return res, a.Get(1).(int64), a.Error(2)
}
func (m *mAttRepo) AdminList(ctx context.Context, q domain.AdminListAttendanceQuery) ([]domain.Attendance, int64, error) {
	a := m.Called(ctx, q)
	res, _ := a.Get(0).([]domain.Attendance)
	return res, a.Get(1).(int64), a.Error(2)
}
func (m *mAttRepo) ListByClassAndUsers(ctx context.Context, classID uuid.UUID, userIDs []uuid.UUID) ([]domain.Attendance, error) {
	a := m.Called(ctx, classID, userIDs)
	res, _ := a.Get(0).([]domain.Attendance)
	return res, a.Error(1)
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

type mMemberRepo struct{ mock.Mock }

func (m *mMemberRepo) Create(ctx context.Context, mem *domain.ClassMember) error {
	return m.Called(ctx, mem).Error(0)
}
func (m *mMemberRepo) Delete(ctx context.Context, classID, userID uuid.UUID) error {
	return m.Called(ctx, classID, userID).Error(0)
}
func (m *mMemberRepo) Exists(ctx context.Context, classID, userID uuid.UUID) (bool, error) {
	a := m.Called(ctx, classID, userID)
	return a.Bool(0), a.Error(1)
}
func (m *mMemberRepo) CountByClass(ctx context.Context, classID uuid.UUID) (int64, error) {
	a := m.Called(ctx, classID)
	return a.Get(0).(int64), a.Error(1)
}
func (m *mMemberRepo) ListByClass(ctx context.Context, classID uuid.UUID, p domain.ListParams) ([]domain.ClassMember, int64, error) {
	a := m.Called(ctx, classID, p)
	res, _ := a.Get(0).([]domain.ClassMember)
	return res, a.Get(1).(int64), a.Error(2)
}
func (m *mMemberRepo) ListAllByClass(ctx context.Context, classID uuid.UUID) ([]domain.ClassMember, error) {
	a := m.Called(ctx, classID)
	res, _ := a.Get(0).([]domain.ClassMember)
	return res, a.Error(1)
}

func newSvc(repo domain.AttendanceRepository, classes domain.ClassRepository, sessions domain.ClassSessionRepository) domain.AttendanceService {
	return attendance.NewService(repo, classes, sessions, nil, nil, nil, nil, nil, authz.NewResolver(nil), slog.Default())
}

func newSvcWithMembers(repo domain.AttendanceRepository, classes domain.ClassRepository, sessions domain.ClassSessionRepository, members domain.ClassMemberRepository) domain.AttendanceService {
	return attendance.NewService(repo, classes, sessions, members, nil, nil, nil, nil, authz.NewResolver(nil), slog.Default())
}

func ownerCtx(userID uuid.UUID) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{UserID: userID})
}

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

func TestMark_UpdatesExisting(t *testing.T) {
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
	dto := domain.CreateAttendanceDTO{UserID: userID, Status: domain.AttendanceStatusPresent}

	a, err := svc.Mark(ownerCtx(ownerID), classID, sessionID, dto)

	assert.NoError(t, err)
	assert.NotNil(t, a)
	assert.Equal(t, domain.AttendanceStatusPresent, a.Status)
	repo.AssertCalled(t, "Update", mock.Anything, mock.AnythingOfType("*domain.Attendance"))
	repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestMark_CreatesWhenMissing(t *testing.T) {
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
	dto := domain.CreateAttendanceDTO{UserID: userID, Status: domain.AttendanceStatusPresent}

	a, err := svc.Mark(ownerCtx(ownerID), classID, sessionID, dto)

	assert.NoError(t, err)
	assert.NotNil(t, a)
	repo.AssertCalled(t, "Create", mock.Anything, mock.AnythingOfType("*domain.Attendance"))
	repo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestListMine_Summarizes(t *testing.T) {
	studentID := uuid.New()
	repo := &mAttRepo{}
	classes := &mClassRepo{}
	sessions := &mSessRepo{}

	repo.On("ListByUser", mock.Anything, studentID, mock.Anything).
		Return([]domain.Attendance{
			{Status: domain.AttendanceStatusPresent},
			{Status: domain.AttendanceStatusPresent},
			{Status: domain.AttendanceStatusAbsent},
			{Status: domain.AttendanceStatusLate},
		}, int64(4), nil)

	svc := newSvc(repo, classes, sessions)
	res, err := svc.ListMine(ownerCtx(studentID), domain.ListParams{Page: 1, PageSize: 50})

	assert.NoError(t, err)
	assert.Equal(t, 2, res.Summary.Present)
	assert.Equal(t, 1, res.Summary.Absent)
	assert.Equal(t, 1, res.Summary.Late)
	assert.Equal(t, 0, res.Summary.Excused)
	assert.Len(t, res.Items, 4)
}

func TestListMine_NoCaller_Forbidden(t *testing.T) {
	svc := newSvc(&mAttRepo{}, &mClassRepo{}, &mSessRepo{})
	_, err := svc.ListMine(context.Background(), domain.ListParams{Page: 1, PageSize: 50})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestMatrix_BuildsCellsAndSummary(t *testing.T) {
	ownerID := uuid.New()
	classID := uuid.New()
	orgID := uuid.New()
	userID := uuid.New()
	pastSessID := uuid.New()
	futureSessID := uuid.New()
	attID := uuid.New()

	now := time.Now()
	past := now.Add(-24 * time.Hour)
	future := now.Add(24 * time.Hour)

	class := &domain.Class{ID: classID, OrganizationID: orgID, UserID: ownerID}

	repo := &mAttRepo{}
	classes := &mClassRepo{}
	sessions := &mSessRepo{}
	members := &mMemberRepo{}

	classes.On("FindByID", mock.Anything, classID).Return(class, nil)
	sessions.On("ListByClass", mock.Anything, classID, mock.Anything).Return(
		[]domain.ClassSession{
			{ID: pastSessID, ClassID: classID, Name: "S1", StartTime: past},
			{ID: futureSessID, ClassID: classID, Name: "S2", StartTime: future},
		}, int64(2), nil)
	members.On("ListByClass", mock.Anything, classID, mock.Anything).Return(
		[]domain.ClassMember{{ClassID: classID, UserID: userID, User: &domain.User{ID: userID, Name: "Stu"}}},
		int64(1), nil)
	repo.On("ListByClassAndUsers", mock.Anything, classID, []uuid.UUID{userID}).Return(
		[]domain.Attendance{
			{ID: attID, ClassID: classID, ClassSessionID: pastSessID, UserID: userID, Status: domain.AttendanceStatusPresent, IsAutoMarked: true},
		}, nil)

	svc := newSvcWithMembers(repo, classes, sessions, members)
	res, err := svc.Matrix(ownerCtx(ownerID), classID, domain.ListAttendanceMatrixQuery{ListParams: domain.ListParams{Page: 1, PageSize: 20}})

	require.NoError(t, err)
	require.Len(t, res.Sessions, 2)
	require.Len(t, res.Students, 1)
	require.Equal(t, int64(1), res.Total)

	stu := res.Students[0]
	cell, ok := stu.Cells[pastSessID]
	require.True(t, ok)
	require.Equal(t, attID, cell.ID)
	require.Equal(t, domain.AttendanceStatusPresent, cell.Status)
	require.True(t, cell.IsAutoMarked)
	_, hasFuture := stu.Cells[futureSessID]
	require.False(t, hasFuture)

	require.Equal(t, 1, stu.Summary.Present)
	require.Equal(t, 1, stu.Summary.StartedCount)
	require.InDelta(t, 1.0, stu.Summary.Rate, 0.001)
}

func TestMatrix_ForbiddenForNonManager(t *testing.T) {
	classID := uuid.New()
	stranger := uuid.New()
	class := &domain.Class{ID: classID, UserID: uuid.New()}

	repo := &mAttRepo{}
	classes := &mClassRepo{}
	classes.On("FindByID", mock.Anything, classID).Return(class, nil)

	svc := newSvcWithMembers(repo, classes, &mSessRepo{}, &mMemberRepo{})
	_, err := svc.Matrix(ownerCtx(stranger), classID, domain.ListAttendanceMatrixQuery{ListParams: domain.ListParams{Page: 1, PageSize: 20}})
	require.ErrorIs(t, err, domain.ErrForbidden)
}
