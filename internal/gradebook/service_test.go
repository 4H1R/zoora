package gradebook_test

import (
	"context"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/gradebook"
	"github.com/4H1R/zoora/internal/platform/authz"
)

type mColRepo struct{ mock.Mock }

func (m *mColRepo) Create(ctx context.Context, col *domain.GradebookColumn) error {
	return m.Called(ctx, col).Error(0)
}

func (m *mColRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.GradebookColumn, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.GradebookColumn), a.Error(1)
}

func (m *mColRepo) Update(ctx context.Context, col *domain.GradebookColumn) error {
	return m.Called(ctx, col).Error(0)
}

func (m *mColRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mColRepo) ListByClass(ctx context.Context, classID uuid.UUID, q domain.ListGradebookColumnsQuery) ([]domain.GradebookColumn, int64, error) {
	a := m.Called(ctx, classID, q)
	cols, _ := a.Get(0).([]domain.GradebookColumn)
	return cols, a.Get(1).(int64), a.Error(2)
}

func (m *mColRepo) ListAllByClass(ctx context.Context, classID uuid.UUID) ([]domain.GradebookColumn, error) {
	a := m.Called(ctx, classID)
	cols, _ := a.Get(0).([]domain.GradebookColumn)
	return cols, a.Error(1)
}

type mCellRepo struct{ mock.Mock }

func (m *mCellRepo) Upsert(ctx context.Context, cell *domain.GradebookCell) error {
	return m.Called(ctx, cell).Error(0)
}

func (m *mCellRepo) ListByColumns(ctx context.Context, columnIDs []uuid.UUID) ([]domain.GradebookCell, error) {
	a := m.Called(ctx, columnIDs)
	cells, _ := a.Get(0).([]domain.GradebookCell)
	return cells, a.Error(1)
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
	cs, _ := a.Get(0).([]domain.Class)
	return cs, a.Get(1).(int64), a.Error(2)
}

func (m *mClassRepo) ListByNames(ctx context.Context, orgID uuid.UUID, names []string) ([]domain.Class, error) {
	a := m.Called(ctx, orgID, names)
	cs, _ := a.Get(0).([]domain.Class)
	return cs, a.Error(1)
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
	cs, _ := a.Get(0).([]domain.Class)
	return cs, a.Get(1).(int64), a.Error(2)
}

type mMemberRepo struct{ mock.Mock }

func (m *mMemberRepo) Create(ctx context.Context, cm *domain.ClassMember) error {
	return m.Called(ctx, cm).Error(0)
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
	ms, _ := a.Get(0).([]domain.ClassMember)
	return ms, a.Get(1).(int64), a.Error(2)
}

func (m *mMemberRepo) ListAllByClass(ctx context.Context, classID uuid.UUID) ([]domain.ClassMember, error) {
	a := m.Called(ctx, classID)
	ms, _ := a.Get(0).([]domain.ClassMember)
	return ms, a.Error(1)
}

type mAttendanceRepo struct{ mock.Mock }

func (m *mAttendanceRepo) Create(ctx context.Context, a *domain.Attendance) error {
	return m.Called(ctx, a).Error(0)
}

func (m *mAttendanceRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Attendance, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Attendance), a.Error(1)
}

func (m *mAttendanceRepo) Update(ctx context.Context, a *domain.Attendance) error {
	return m.Called(ctx, a).Error(0)
}

func (m *mAttendanceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mAttendanceRepo) ListBySession(ctx context.Context, sessionID uuid.UUID, q domain.ListAttendanceQuery) ([]domain.Attendance, int64, error) {
	a := m.Called(ctx, sessionID, q)
	items, _ := a.Get(0).([]domain.Attendance)
	return items, a.Get(1).(int64), a.Error(2)
}

func (m *mAttendanceRepo) ListByUser(ctx context.Context, userID uuid.UUID, q domain.ListMyAttendanceQuery) ([]domain.Attendance, int64, error) {
	a := m.Called(ctx, userID, q)
	items, _ := a.Get(0).([]domain.Attendance)
	return items, a.Get(1).(int64), a.Error(2)
}

func (m *mAttendanceRepo) SummarizeByUser(ctx context.Context, userID uuid.UUID, q domain.ListMyAttendanceQuery) (domain.MyAttendanceSummary, error) {
	a := m.Called(ctx, userID, q)
	res, _ := a.Get(0).(domain.MyAttendanceSummary)
	return res, a.Error(1)
}

func (m *mAttendanceRepo) FindBySessionAndUser(ctx context.Context, sessionID, userID uuid.UUID) (*domain.Attendance, error) {
	a := m.Called(ctx, sessionID, userID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Attendance), a.Error(1)
}

func (m *mAttendanceRepo) AdminList(ctx context.Context, q domain.AdminListAttendanceQuery) ([]domain.Attendance, int64, error) {
	a := m.Called(ctx, q)
	items, _ := a.Get(0).([]domain.Attendance)
	return items, a.Get(1).(int64), a.Error(2)
}

func (m *mAttendanceRepo) ListByClassAndUsers(ctx context.Context, classID uuid.UUID, userIDs []uuid.UUID) ([]domain.Attendance, error) {
	a := m.Called(ctx, classID, userIDs)
	items, _ := a.Get(0).([]domain.Attendance)
	return items, a.Error(1)
}

type mPracticeSubRepo struct{ mock.Mock }

func (m *mPracticeSubRepo) Create(ctx context.Context, sub *domain.PracticeSubmission) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *mPracticeSubRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.PracticeSubmission, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.PracticeSubmission), a.Error(1)
}

func (m *mPracticeSubRepo) Update(ctx context.Context, sub *domain.PracticeSubmission) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *mPracticeSubRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mPracticeSubRepo) FindByRoomAndUser(ctx context.Context, roomID, userID uuid.UUID) (*domain.PracticeSubmission, error) {
	a := m.Called(ctx, roomID, userID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.PracticeSubmission), a.Error(1)
}

func (m *mPracticeSubRepo) ListByRoom(ctx context.Context, roomID uuid.UUID, p domain.ListParams) ([]domain.PracticeSubmission, int64, error) {
	a := m.Called(ctx, roomID, p)
	subs, _ := a.Get(0).([]domain.PracticeSubmission)
	return subs, a.Get(1).(int64), a.Error(2)
}

func (m *mPracticeSubRepo) ListByRoomsAndUser(ctx context.Context, roomIDs []uuid.UUID, userID uuid.UUID) ([]domain.PracticeSubmission, error) {
	a := m.Called(ctx, roomIDs, userID)
	subs, _ := a.Get(0).([]domain.PracticeSubmission)
	return subs, a.Error(1)
}

func (m *mPracticeSubRepo) CountsByRooms(ctx context.Context, roomIDs []uuid.UUID) (map[uuid.UUID]domain.PracticeRoomStats, error) {
	a := m.Called(ctx, roomIDs)
	rs, _ := a.Get(0).(map[uuid.UUID]domain.PracticeRoomStats)
	return rs, a.Error(1)
}

type mQuizSubRepo struct{ mock.Mock }

func (m *mQuizSubRepo) Create(ctx context.Context, sub *domain.QuizSubmission) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *mQuizSubRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.QuizSubmission, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.QuizSubmission), a.Error(1)
}

func (m *mQuizSubRepo) Update(ctx context.Context, sub *domain.QuizSubmission) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *mQuizSubRepo) FindByQuizAndUser(ctx context.Context, quizID, userID uuid.UUID) (*domain.QuizSubmission, error) {
	a := m.Called(ctx, quizID, userID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.QuizSubmission), a.Error(1)
}

func (m *mQuizSubRepo) ListByQuiz(ctx context.Context, quizID uuid.UUID, q domain.ListSubmissionsQuery) ([]domain.QuizSubmission, int64, error) {
	a := m.Called(ctx, quizID, q)
	subs, _ := a.Get(0).([]domain.QuizSubmission)
	return subs, a.Get(1).(int64), a.Error(2)
}
func (m *mQuizSubRepo) FindByQuizID(ctx context.Context, quizID uuid.UUID) ([]domain.QuizSubmission, error) {
	a := m.Called(ctx, quizID)
	subs, _ := a.Get(0).([]domain.QuizSubmission)
	return subs, a.Error(1)
}

type mQuizRepo struct{ mock.Mock }

func (m *mQuizRepo) Create(ctx context.Context, quiz *domain.Quiz) error {
	return m.Called(ctx, quiz).Error(0)
}

func (m *mQuizRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Quiz, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Quiz), a.Error(1)
}

func (m *mQuizRepo) Update(ctx context.Context, quiz *domain.Quiz) error {
	return m.Called(ctx, quiz).Error(0)
}

func (m *mQuizRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mQuizRepo) List(ctx context.Context, scope domain.QuizListScope, p domain.ListParams) ([]domain.Quiz, int64, error) {
	a := m.Called(ctx, scope, p)
	qs, _ := a.Get(0).([]domain.Quiz)
	return qs, a.Get(1).(int64), a.Error(2)
}

func (m *mQuizRepo) CountPendingSubmissionsByQuizIDs(ctx context.Context, quizIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	a := m.Called(ctx, quizIDs)
	res, _ := a.Get(0).(map[uuid.UUID]int64)
	return res, a.Error(1)
}

func (m *mQuizRepo) ListByMemberWithRooms(ctx context.Context, userID uuid.UUID, classID *uuid.UUID, p domain.ListParams) ([]domain.Quiz, error) {
	a := m.Called(ctx, userID, classID, p)
	qs, _ := a.Get(0).([]domain.Quiz)
	return qs, a.Error(1)
}

func (m *mQuizRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mQuizRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.Quiz, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Quiz), a.Error(1)
}

func (m *mQuizRepo) AdminList(ctx context.Context, q domain.AdminListQuizzesQuery) ([]domain.Quiz, int64, error) {
	a := m.Called(ctx, q)
	qs, _ := a.Get(0).([]domain.Quiz)
	return qs, a.Get(1).(int64), a.Error(2)
}

type mPracticeRoomRepo struct{ mock.Mock }

func (m *mPracticeRoomRepo) Create(ctx context.Context, room *domain.PracticeRoom) error {
	return m.Called(ctx, room).Error(0)
}

func (m *mPracticeRoomRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.PracticeRoom, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.PracticeRoom), a.Error(1)
}

func (m *mPracticeRoomRepo) Update(ctx context.Context, room *domain.PracticeRoom) error {
	return m.Called(ctx, room).Error(0)
}

func (m *mPracticeRoomRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mPracticeRoomRepo) List(ctx context.Context, scope domain.PracticeRoomListScope, q domain.ListPracticeRoomsQuery) ([]domain.PracticeRoom, int64, error) {
	a := m.Called(ctx, scope, q)
	rs, _ := a.Get(0).([]domain.PracticeRoom)
	return rs, a.Get(1).(int64), a.Error(2)
}

func (m *mPracticeRoomRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mPracticeRoomRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.PracticeRoom, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.PracticeRoom), a.Error(1)
}

func (m *mPracticeRoomRepo) AdminList(ctx context.Context, q domain.AdminListPracticeRoomsQuery) ([]domain.PracticeRoom, int64, error) {
	a := m.Called(ctx, q)
	rs, _ := a.Get(0).([]domain.PracticeRoom)
	return rs, a.Get(1).(int64), a.Error(2)
}

func (m *mPracticeRoomRepo) MemberCountsByClasses(ctx context.Context, classIDs []uuid.UUID) (map[uuid.UUID]int64, error) {
	a := m.Called(ctx, classIDs)
	res, _ := a.Get(0).(map[uuid.UUID]int64)
	return res, a.Error(1)
}

func (m *mPracticeRoomRepo) ViewerMemberClasses(ctx context.Context, userID uuid.UUID, classIDs []uuid.UUID) ([]uuid.UUID, error) {
	a := m.Called(ctx, userID, classIDs)
	res, _ := a.Get(0).([]uuid.UUID)
	return res, a.Error(1)
}

type mSessionRepo struct{ mock.Mock }

func (m *mSessionRepo) Create(ctx context.Context, session *domain.ClassSession) error {
	return m.Called(ctx, session).Error(0)
}

func (m *mSessionRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.ClassSession, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.ClassSession), a.Error(1)
}

func (m *mSessionRepo) Update(ctx context.Context, session *domain.ClassSession) error {
	return m.Called(ctx, session).Error(0)
}

func (m *mSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mSessionRepo) ListByClass(ctx context.Context, classID uuid.UUID, q domain.ListClassSessionsQuery) ([]domain.ClassSession, int64, error) {
	a := m.Called(ctx, classID, q)
	ss, _ := a.Get(0).([]domain.ClassSession)
	return ss, a.Get(1).(int64), a.Error(2)
}

func (m *mSessionRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mSessionRepo) FindByIDIncludingDeleted(ctx context.Context, id uuid.UUID) (*domain.ClassSession, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.ClassSession), a.Error(1)
}

func (m *mSessionRepo) AdminList(ctx context.Context, q domain.AdminListClassSessionsQuery) ([]domain.ClassSession, int64, error) {
	a := m.Called(ctx, q)
	ss, _ := a.Get(0).([]domain.ClassSession)
	return ss, a.Get(1).(int64), a.Error(2)
}

func teacherCtx(userID uuid.UUID) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID:      userID,
		Permissions: []string{"classes:update", "gradebook:create", "gradebook:update", "gradebook:delete"},
	})
}

func studentCtx(userID uuid.UUID) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID: userID,
	})
}

type deps struct {
	colRepo          *mColRepo
	cellRepo         *mCellRepo
	classRepo        *mClassRepo
	memberRepo       *mMemberRepo
	attendanceRepo   *mAttendanceRepo
	practiceRepo     *mPracticeSubRepo
	quizSubRepo      *mQuizSubRepo
	quizRepo         *mQuizRepo
	practiceRoomRepo *mPracticeRoomRepo
	sessionRepo      *mSessionRepo
	audit            *auditSpy
}

// fakeTransactor runs fn inline with no real DB — unit tests exercise the audit
// same-tx wiring without a database.
type fakeTransactor struct{}

func (fakeTransactor) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return fn(ctx)
}

// auditSpy captures the records a service emits so tests can assert on them.
type auditSpy struct{ records []domain.AuditRecord }

func (a *auditSpy) Record(_ context.Context, r domain.AuditRecord) error {
	a.records = append(a.records, r)
	return nil
}

func (a *auditSpy) RecordDenied(_ context.Context, _ domain.AuditRecord) error { return nil }

func newDeps() deps {
	return deps{
		colRepo:          &mColRepo{},
		cellRepo:         &mCellRepo{},
		classRepo:        &mClassRepo{},
		memberRepo:       &mMemberRepo{},
		attendanceRepo:   &mAttendanceRepo{},
		practiceRepo:     &mPracticeSubRepo{},
		quizSubRepo:      &mQuizSubRepo{},
		quizRepo:         &mQuizRepo{},
		practiceRoomRepo: &mPracticeRoomRepo{},
		sessionRepo:      &mSessionRepo{},
		audit:            &auditSpy{},
	}
}

func (d deps) service() domain.GradebookService {
	return gradebook.NewService(
		d.colRepo, d.cellRepo, d.classRepo, d.memberRepo,
		d.attendanceRepo, d.practiceRepo, d.quizSubRepo,
		d.quizRepo, d.practiceRoomRepo, d.sessionRepo,
		authz.NewResolver(d.memberRepo),
		fakeTransactor{}, d.audit,
		slog.Default(),
	)
}

func TestCreateColumn_Success(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: teacherID}, nil)
	d.colRepo.On("Create", ctx, mock.AnythingOfType("*domain.GradebookColumn")).Return(nil)

	svc := d.service()
	col, err := svc.CreateColumn(ctx, classID, domain.CreateGradebookColumnDTO{
		Title:      "Session 1",
		Type:       domain.GradebookColumnManualGrade,
		OrderIndex: 0,
	})
	assert.NoError(t, err)
	assert.Equal(t, "Session 1", col.Title)
	assert.Equal(t, domain.GradebookColumnManualGrade, col.Type)
}

func TestCreateColumn_AutoWithoutSourceID_ValidationError(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: teacherID}, nil)

	svc := d.service()
	_, err := svc.CreateColumn(ctx, classID, domain.CreateGradebookColumnDTO{
		Title: "Auto Attendance",
		Type:  domain.GradebookColumnAutoAttendance,
	})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestCreateColumn_AutoWithSourceID_Success(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	sourceID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: teacherID}, nil)
	d.quizRepo.On("FindByID", ctx, sourceID).
		Return(&domain.Quiz{ID: sourceID, ClassID: classID}, nil)
	d.colRepo.On("Create", ctx, mock.AnythingOfType("*domain.GradebookColumn")).Return(nil)

	svc := d.service()
	col, err := svc.CreateColumn(ctx, classID, domain.CreateGradebookColumnDTO{
		Title:    "Auto Quiz Score",
		Type:     domain.GradebookColumnAutoQuiz,
		SourceID: &sourceID,
	})
	assert.NoError(t, err)
	assert.Equal(t, &sourceID, col.SourceID)
}

// TestCreateColumn_AutoQuiz_ForeignClass_ValidationError guards the horizontal
// grade-disclosure path: an auto column whose quiz belongs to a different class
// must be rejected before any column is created.
func TestCreateColumn_AutoQuiz_ForeignClass_ValidationError(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	otherClassID := uuid.New()
	sourceID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: teacherID}, nil)
	d.quizRepo.On("FindByID", ctx, sourceID).
		Return(&domain.Quiz{ID: sourceID, ClassID: otherClassID}, nil)

	svc := d.service()
	_, err := svc.CreateColumn(ctx, classID, domain.CreateGradebookColumnDTO{
		Title:    "Foreign Quiz",
		Type:     domain.GradebookColumnAutoQuiz,
		SourceID: &sourceID,
	})
	assert.ErrorIs(t, err, domain.ErrValidation)
	d.colRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreateColumn_AutoPractice_ForeignClass_ValidationError(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	otherClassID := uuid.New()
	sourceID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: teacherID}, nil)
	d.practiceRoomRepo.On("FindByID", ctx, sourceID).
		Return(&domain.PracticeRoom{ID: sourceID, ClassID: otherClassID}, nil)

	svc := d.service()
	_, err := svc.CreateColumn(ctx, classID, domain.CreateGradebookColumnDTO{
		Title:    "Foreign Practice",
		Type:     domain.GradebookColumnAutoPractice,
		SourceID: &sourceID,
	})
	assert.ErrorIs(t, err, domain.ErrValidation)
	d.colRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreateColumn_AutoAttendance_ForeignClass_ValidationError(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	otherClassID := uuid.New()
	sourceID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: teacherID}, nil)
	d.sessionRepo.On("FindByID", ctx, sourceID).
		Return(&domain.ClassSession{ID: sourceID, ClassID: otherClassID}, nil)

	svc := d.service()
	_, err := svc.CreateColumn(ctx, classID, domain.CreateGradebookColumnDTO{
		Title:    "Foreign Attendance",
		Type:     domain.GradebookColumnAutoAttendance,
		SourceID: &sourceID,
	})
	assert.ErrorIs(t, err, domain.ErrValidation)
	d.colRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestCreateColumn_AutoPractice_SameClass_Success(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	sourceID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: teacherID}, nil)
	d.practiceRoomRepo.On("FindByID", ctx, sourceID).
		Return(&domain.PracticeRoom{ID: sourceID, ClassID: classID}, nil)
	d.colRepo.On("Create", ctx, mock.AnythingOfType("*domain.GradebookColumn")).Return(nil)

	svc := d.service()
	col, err := svc.CreateColumn(ctx, classID, domain.CreateGradebookColumnDTO{
		Title:    "Practice",
		Type:     domain.GradebookColumnAutoPractice,
		SourceID: &sourceID,
	})
	assert.NoError(t, err)
	assert.Equal(t, &sourceID, col.SourceID)
}

func TestCreateColumn_AutoAttendance_SameClass_Success(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	sourceID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: teacherID}, nil)
	d.sessionRepo.On("FindByID", ctx, sourceID).
		Return(&domain.ClassSession{ID: sourceID, ClassID: classID}, nil)
	d.colRepo.On("Create", ctx, mock.AnythingOfType("*domain.GradebookColumn")).Return(nil)

	svc := d.service()
	col, err := svc.CreateColumn(ctx, classID, domain.CreateGradebookColumnDTO{
		Title:    "Attendance",
		Type:     domain.GradebookColumnAutoAttendance,
		SourceID: &sourceID,
	})
	assert.NoError(t, err)
	assert.Equal(t, &sourceID, col.SourceID)
}

func TestCreateColumn_NoCaller_Forbidden(t *testing.T) {
	d := newDeps()
	svc := d.service()
	_, err := svc.CreateColumn(context.Background(), uuid.New(), domain.CreateGradebookColumnDTO{
		Title: "X", Type: domain.GradebookColumnManualGrade,
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestCreateColumn_NotOwner_Forbidden(t *testing.T) {
	studentID := uuid.New()
	classID := uuid.New()
	otherTeacher := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: otherTeacher}, nil)

	svc := d.service()
	_, err := svc.CreateColumn(ctx, classID, domain.CreateGradebookColumnDTO{
		Title: "X", Type: domain.GradebookColumnManualGrade,
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestUpdateColumn_Success(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	colID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.colRepo.On("FindByID", ctx, colID).
		Return(&domain.GradebookColumn{ID: colID, ClassID: classID, Title: "Old"}, nil)
	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: teacherID}, nil)
	d.colRepo.On("Update", ctx, mock.AnythingOfType("*domain.GradebookColumn")).Return(nil)

	svc := d.service()
	newTitle := "New Title"
	col, err := svc.UpdateColumn(ctx, colID, domain.UpdateGradebookColumnDTO{Title: &newTitle})
	assert.NoError(t, err)
	assert.Equal(t, "New Title", col.Title)
}

func TestDeleteColumn_Success(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	colID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.colRepo.On("FindByID", ctx, colID).
		Return(&domain.GradebookColumn{ID: colID, ClassID: classID}, nil)
	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: teacherID}, nil)
	d.colRepo.On("Delete", ctx, colID).Return(nil)

	svc := d.service()
	err := svc.DeleteColumn(ctx, colID)
	assert.NoError(t, err)
}

func TestDeleteColumn_NotOwner_Forbidden(t *testing.T) {
	studentID := uuid.New()
	classID := uuid.New()
	colID := uuid.New()
	otherTeacher := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	d.colRepo.On("FindByID", ctx, colID).
		Return(&domain.GradebookColumn{ID: colID, ClassID: classID}, nil)
	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: otherTeacher}, nil)

	svc := d.service()
	err := svc.DeleteColumn(ctx, colID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestUpsertCell_ManualColumn_Success(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	colID := uuid.New()
	studentID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.colRepo.On("FindByID", ctx, colID).
		Return(&domain.GradebookColumn{ID: colID, ClassID: classID, Type: domain.GradebookColumnManualGrade}, nil)
	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: teacherID}, nil)
	d.cellRepo.On("Upsert", ctx, mock.AnythingOfType("*domain.GradebookCell")).Return(nil)

	svc := d.service()
	cell, err := svc.UpsertCell(ctx, classID, colID, domain.UpsertGradebookCellDTO{
		StudentID: studentID,
		Value:     "95",
	})
	assert.NoError(t, err)
	assert.Equal(t, "95", cell.Value)
	assert.Equal(t, studentID, cell.StudentID)
}

func TestUpsertCell_RecordsGradedAudit(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	colID := uuid.New()
	studentID := uuid.New()
	orgID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.colRepo.On("FindByID", ctx, colID).
		Return(&domain.GradebookColumn{ID: colID, ClassID: classID, Title: "Midterm", Type: domain.GradebookColumnManualGrade}, nil)
	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: teacherID, OrganizationID: orgID}, nil)
	d.cellRepo.On("Upsert", ctx, mock.AnythingOfType("*domain.GradebookCell")).Return(nil)

	svc := d.service()
	_, err := svc.UpsertCell(ctx, classID, colID, domain.UpsertGradebookCellDTO{StudentID: studentID, Value: "95"})
	assert.NoError(t, err)
	assert.Len(t, d.audit.records, 1)
	assert.Equal(t, domain.AuditGraded, d.audit.records[0].Action)
	assert.Equal(t, domain.AuditTargetGradebook, d.audit.records[0].TargetType)
	assert.Equal(t, "Midterm", d.audit.records[0].TargetLabel)
	assert.Equal(t, studentID.String(), d.audit.records[0].Metadata["student_id"])
}

func TestDeleteColumn_RecordsAudit(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	colID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.colRepo.On("FindByID", ctx, colID).
		Return(&domain.GradebookColumn{ID: colID, ClassID: classID, Title: "Final"}, nil)
	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: teacherID}, nil)
	d.colRepo.On("Delete", ctx, colID).Return(nil)

	svc := d.service()
	err := svc.DeleteColumn(ctx, colID)
	assert.NoError(t, err)
	assert.Len(t, d.audit.records, 1)
	assert.Equal(t, domain.AuditDeleted, d.audit.records[0].Action)
	assert.Equal(t, domain.AuditTargetGradebook, d.audit.records[0].TargetType)
	assert.Equal(t, "Final", d.audit.records[0].TargetLabel)
	assert.NotNil(t, d.audit.records[0].TargetID)
	assert.Equal(t, colID, *d.audit.records[0].TargetID)
}

func TestUpsertCell_AutoColumn_ValidationError(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	colID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.colRepo.On("FindByID", ctx, colID).
		Return(&domain.GradebookColumn{ID: colID, ClassID: classID, Type: domain.GradebookColumnAutoAttendance}, nil)

	svc := d.service()
	_, err := svc.UpsertCell(ctx, classID, colID, domain.UpsertGradebookCellDTO{
		StudentID: uuid.New(),
		Value:     "present",
	})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestUpsertCell_WrongClass_NotFound(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	otherClassID := uuid.New()
	colID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.colRepo.On("FindByID", ctx, colID).
		Return(&domain.GradebookColumn{ID: colID, ClassID: otherClassID, Type: domain.GradebookColumnManualGrade}, nil)

	svc := d.service()
	_, err := svc.UpsertCell(ctx, classID, colID, domain.UpsertGradebookCellDTO{
		StudentID: uuid.New(),
		Value:     "90",
	})
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestGetMatrix_Teacher_Success(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	student1 := uuid.New()
	student2 := uuid.New()
	colManual := uuid.New()
	colAuto := uuid.New()
	sessionID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: teacherID}, nil)
	d.colRepo.On("ListAllByClass", ctx, classID).
		Return([]domain.GradebookColumn{
			{ID: colManual, ClassID: classID, Title: "Grade", Type: domain.GradebookColumnManualGrade, OrderIndex: 0},
			{ID: colAuto, ClassID: classID, Title: "Attendance", Type: domain.GradebookColumnAutoAttendance, SourceID: &sessionID, OrderIndex: 1},
		}, nil)
	d.memberRepo.On("ListAllByClass", ctx, classID).
		Return([]domain.ClassMember{
			{UserID: student1},
			{UserID: student2},
		}, nil)
	d.cellRepo.On("ListByColumns", ctx, []uuid.UUID{colManual}).
		Return([]domain.GradebookCell{
			{ColumnID: colManual, StudentID: student1, Value: "85"},
			{ColumnID: colManual, StudentID: student2, Value: "92"},
		}, nil)
	d.attendanceRepo.On("ListBySession", ctx, sessionID, mock.Anything).
		Return([]domain.Attendance{
			{UserID: student1, Status: domain.AttendanceStatusPresent},
			{UserID: student2, Status: domain.AttendanceStatusLate},
		}, int64(2), nil)

	svc := d.service()
	matrix, err := svc.GetMatrix(ctx, classID)
	assert.NoError(t, err)
	assert.Len(t, matrix.Columns, 2)
	assert.Len(t, matrix.Rows, 2)

	// Check student1 row
	for _, row := range matrix.Rows {
		if row.StudentID == student1 {
			assert.Equal(t, "85", row.Cells[colManual.String()])
			assert.Equal(t, "present", row.Cells[colAuto.String()])
		}
		if row.StudentID == student2 {
			assert.Equal(t, "92", row.Cells[colManual.String()])
			assert.Equal(t, "late", row.Cells[colAuto.String()])
		}
	}
}

func TestGetMatrix_Student_Enrolled_Success(t *testing.T) {
	studentID := uuid.New()
	classID := uuid.New()
	otherTeacher := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: otherTeacher}, nil)
	d.memberRepo.On("Exists", ctx, classID, studentID).Return(true, nil)
	d.colRepo.On("ListAllByClass", ctx, classID).Return([]domain.GradebookColumn{}, nil)
	d.memberRepo.On("ListAllByClass", ctx, classID).Return([]domain.ClassMember{}, nil)
	d.cellRepo.On("ListByColumns", ctx, ([]uuid.UUID)(nil)).Return([]domain.GradebookCell{}, nil)

	svc := d.service()
	matrix, err := svc.GetMatrix(ctx, classID)
	assert.NoError(t, err)
	assert.NotNil(t, matrix)
}

func TestGetMatrix_Student_Enrolled_OwnRowOnly(t *testing.T) {
	studentID := uuid.New()
	classmateID := uuid.New()
	classID := uuid.New()
	otherTeacher := uuid.New()
	colManual := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: otherTeacher}, nil)
	d.memberRepo.On("Exists", ctx, classID, studentID).Return(true, nil)
	d.colRepo.On("ListAllByClass", ctx, classID).
		Return([]domain.GradebookColumn{
			{ID: colManual, ClassID: classID, Title: "Grade", Type: domain.GradebookColumnManualGrade, OrderIndex: 0},
		}, nil)
	d.memberRepo.On("ListAllByClass", ctx, classID).
		Return([]domain.ClassMember{
			{UserID: studentID},
			{UserID: classmateID},
		}, nil)
	d.cellRepo.On("ListByColumns", ctx, []uuid.UUID{colManual}).
		Return([]domain.GradebookCell{
			{ColumnID: colManual, StudentID: studentID, Value: "85"},
			{ColumnID: colManual, StudentID: classmateID, Value: "92"},
		}, nil)

	svc := d.service()
	matrix, err := svc.GetMatrix(ctx, classID)
	assert.NoError(t, err)
	// Enrolled student must see ONLY their own row, never classmates' grades.
	assert.Len(t, matrix.Rows, 1)
	assert.Equal(t, studentID, matrix.Rows[0].StudentID)
	assert.Equal(t, "85", matrix.Rows[0].Cells[colManual.String()])
}

func TestGetMatrix_Student_NotEnrolled_Forbidden(t *testing.T) {
	studentID := uuid.New()
	classID := uuid.New()
	otherTeacher := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: otherTeacher}, nil)
	d.memberRepo.On("Exists", ctx, classID, studentID).Return(false, nil)

	svc := d.service()
	_, err := svc.GetMatrix(ctx, classID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestGetMatrix_NoCaller_Forbidden(t *testing.T) {
	d := newDeps()
	svc := d.service()
	_, err := svc.GetMatrix(context.Background(), uuid.New())
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestGradebookColumnType_Valid(t *testing.T) {
	assert.True(t, domain.GradebookColumnAutoAttendance.Valid())
	assert.True(t, domain.GradebookColumnAutoPractice.Valid())
	assert.True(t, domain.GradebookColumnAutoQuiz.Valid())
	assert.True(t, domain.GradebookColumnManualGrade.Valid())
	assert.True(t, domain.GradebookColumnManualAttendance.Valid())
	assert.True(t, domain.GradebookColumnManualText.Valid())
	assert.False(t, domain.GradebookColumnType("invalid").Valid())
}

func TestGradebookColumnType_IsAuto(t *testing.T) {
	assert.True(t, domain.GradebookColumnAutoAttendance.IsAuto())
	assert.True(t, domain.GradebookColumnAutoPractice.IsAuto())
	assert.True(t, domain.GradebookColumnAutoQuiz.IsAuto())
	assert.False(t, domain.GradebookColumnManualGrade.IsAuto())
	assert.False(t, domain.GradebookColumnManualAttendance.IsAuto())
	assert.False(t, domain.GradebookColumnManualText.IsAuto())
}
