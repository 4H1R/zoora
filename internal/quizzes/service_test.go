package quizzes_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/quizzes"
)

// --- Mock implementations ---

type mQuizRepo struct{ mock.Mock }

func (m *mQuizRepo) Create(ctx context.Context, q *domain.Quiz) error {
	return m.Called(ctx, q).Error(0)
}
func (m *mQuizRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Quiz, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Quiz), a.Error(1)
}
func (m *mQuizRepo) Update(ctx context.Context, q *domain.Quiz) error {
	return m.Called(ctx, q).Error(0)
}
func (m *mQuizRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mQuizRepo) List(ctx context.Context, scope domain.QuizListScope, p domain.ListParams) ([]domain.Quiz, int64, error) {
	a := m.Called(ctx, scope, p)
	qs, _ := a.Get(0).([]domain.Quiz)
	return qs, a.Get(1).(int64), a.Error(2)
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

type mRuleRepo struct{ mock.Mock }

func (m *mRuleRepo) Create(ctx context.Context, r *domain.QuizRule) error {
	return m.Called(ctx, r).Error(0)
}
func (m *mRuleRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.QuizRule, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.QuizRule), a.Error(1)
}
func (m *mRuleRepo) Update(ctx context.Context, r *domain.QuizRule) error {
	return m.Called(ctx, r).Error(0)
}
func (m *mRuleRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mRuleRepo) ListByQuiz(ctx context.Context, quizID uuid.UUID, p domain.ListParams) ([]domain.QuizRule, int64, error) {
	a := m.Called(ctx, quizID, p)
	rs, _ := a.Get(0).([]domain.QuizRule)
	return rs, a.Get(1).(int64), a.Error(2)
}

type mRoomRepo struct{ mock.Mock }

func (m *mRoomRepo) Create(ctx context.Context, r *domain.QuizRoom) error {
	return m.Called(ctx, r).Error(0)
}
func (m *mRoomRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.QuizRoom, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.QuizRoom), a.Error(1)
}
func (m *mRoomRepo) Update(ctx context.Context, r *domain.QuizRoom) error {
	return m.Called(ctx, r).Error(0)
}
func (m *mRoomRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mRoomRepo) ListByQuiz(ctx context.Context, quizID uuid.UUID, p domain.ListParams) ([]domain.QuizRoom, int64, error) {
	a := m.Called(ctx, quizID, p)
	rs, _ := a.Get(0).([]domain.QuizRoom)
	return rs, a.Get(1).(int64), a.Error(2)
}
func (m *mRoomRepo) ListBySessionID(ctx context.Context, sessionID uuid.UUID) ([]domain.QuizRoom, error) {
	a := m.Called(ctx, sessionID)
	rs, _ := a.Get(0).([]domain.QuizRoom)
	return rs, a.Error(1)
}
func (m *mRoomRepo) FindOpenByQuizID(ctx context.Context, quizID uuid.UUID) (*domain.QuizRoom, error) {
	a := m.Called(ctx, quizID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.QuizRoom), a.Error(1)
}

type mSubRepo struct{ mock.Mock }

func (m *mSubRepo) Create(ctx context.Context, s *domain.QuizSubmission) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mSubRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.QuizSubmission, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.QuizSubmission), a.Error(1)
}
func (m *mSubRepo) Update(ctx context.Context, s *domain.QuizSubmission) error {
	return m.Called(ctx, s).Error(0)
}
func (m *mSubRepo) FindByQuizAndUser(ctx context.Context, quizID, userID uuid.UUID) (*domain.QuizSubmission, error) {
	a := m.Called(ctx, quizID, userID)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.QuizSubmission), a.Error(1)
}
func (m *mSubRepo) ListByQuiz(ctx context.Context, quizID uuid.UUID, q domain.ListSubmissionsQuery) ([]domain.QuizSubmission, int64, error) {
	a := m.Called(ctx, quizID, q)
	ss, _ := a.Get(0).([]domain.QuizSubmission)
	return ss, a.Get(1).(int64), a.Error(2)
}

type mQRepo struct{ mock.Mock }

func (m *mQRepo) Create(ctx context.Context, q *domain.Question) error {
	return m.Called(ctx, q).Error(0)
}
func (m *mQRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Question, error) {
	a := m.Called(ctx, id)
	if a.Get(0) == nil {
		return nil, a.Error(1)
	}
	return a.Get(0).(*domain.Question), a.Error(1)
}
func (m *mQRepo) Update(ctx context.Context, q *domain.Question) error {
	return m.Called(ctx, q).Error(0)
}
func (m *mQRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mQRepo) ListByBank(ctx context.Context, bankID uuid.UUID, q domain.ListQuestionsQuery) ([]domain.Question, int64, error) {
	a := m.Called(ctx, bankID, q)
	qs, _ := a.Get(0).([]domain.Question)
	return qs, a.Get(1).(int64), a.Error(2)
}
func (m *mQRepo) FindByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.Question, error) {
	a := m.Called(ctx, ids)
	qs, _ := a.Get(0).([]domain.Question)
	return qs, a.Error(1)
}
func (m *mQRepo) ListAllByBank(ctx context.Context, bankID uuid.UUID) ([]domain.Question, error) {
	a := m.Called(ctx, bankID)
	qs, _ := a.Get(0).([]domain.Question)
	return qs, a.Error(1)
}
func (m *mQRepo) CountByBank(ctx context.Context, bankID uuid.UUID) (int64, error) {
	a := m.Called(ctx, bankID)
	return a.Get(0).(int64), a.Error(1)
}
func (m *mQRepo) RandomByBank(ctx context.Context, bankID uuid.UUID, count int) ([]domain.Question, error) {
	a := m.Called(ctx, bankID, count)
	qs, _ := a.Get(0).([]domain.Question)
	return qs, a.Error(1)
}
func (m *mQRepo) HardDelete(ctx context.Context, id uuid.UUID) error {
	return m.Called(ctx, id).Error(0)
}
func (m *mQRepo) AdminList(ctx context.Context, q domain.AdminListQuestionsQuery) ([]domain.Question, int64, error) {
	a := m.Called(ctx, q)
	qs, _ := a.Get(0).([]domain.Question)
	return qs, a.Get(1).(int64), a.Error(2)
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

// --- Helpers ---

func teacherCtx(userID uuid.UUID) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID:      userID,
		Permissions: []string{"quizzes:update_any", "quizzes:create", "quizzes:view", "quizzes:delete"},
	})
}

func studentCtx(userID uuid.UUID) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID: userID,
	})
}

type testDeps struct {
	quizRepo    *mQuizRepo
	ruleRepo    *mRuleRepo
	roomRepo    *mRoomRepo
	subRepo     *mSubRepo
	questionRepo *mQRepo
	classRepo   *mClassRepo
	memberRepo  *mMemberRepo
}

func newDeps() testDeps {
	return testDeps{
		quizRepo:    &mQuizRepo{},
		ruleRepo:    &mRuleRepo{},
		roomRepo:    &mRoomRepo{},
		subRepo:     &mSubRepo{},
		questionRepo: &mQRepo{},
		classRepo:   &mClassRepo{},
		memberRepo:  &mMemberRepo{},
	}
}

func (d testDeps) service() domain.QuizService {
	return quizzes.NewService(
		d.quizRepo, d.ruleRepo, d.roomRepo, d.subRepo,
		d.questionRepo, d.classRepo, d.memberRepo, slog.Default(),
	)
}

// --- Quiz CRUD tests ---

func TestQuizService_Create_AsTeacher(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: teacherID}, nil)
	d.quizRepo.On("Create", ctx, mock.AnythingOfType("*domain.Quiz")).Return(nil)

	svc := d.service()
	quiz, err := svc.Create(ctx, domain.CreateQuizDTO{
		ClassID:         classID,
		Title:           "Midterm",
		DurationMinutes: 60,
	})
	assert.NoError(t, err)
	assert.Equal(t, "Midterm", quiz.Title)
	assert.Equal(t, teacherID, quiz.UserID)
}

func TestQuizService_Create_NoCaller_Forbidden(t *testing.T) {
	d := newDeps()
	svc := d.service()
	_, err := svc.Create(context.Background(), domain.CreateQuizDTO{ClassID: uuid.New(), Title: "X", DurationMinutes: 10})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestQuizService_Create_NotClassOwner_Forbidden(t *testing.T) {
	teacherID := uuid.New()
	otherTeacher := uuid.New()
	classID := uuid.New()
	ctx := studentCtx(teacherID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, UserID: otherTeacher}, nil)

	svc := d.service()
	_, err := svc.Create(ctx, domain.CreateQuizDTO{ClassID: classID, Title: "X", DurationMinutes: 10})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

// --- Submission tests ---

func TestQuizService_StartSubmission_Success(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	classID := uuid.New()
	roomID := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	now := time.Now()
	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, ClassID: classID}, nil)
	d.memberRepo.On("Exists", ctx, classID, studentID).Return(true, nil)
	d.roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.QuizRoom{ID: roomID, QuizID: quizID, StartedAt: &now}, nil)
	d.subRepo.On("FindByQuizAndUser", ctx, quizID, studentID).
		Return((*domain.QuizSubmission)(nil), domain.ErrNotFound)
	d.subRepo.On("Create", ctx, mock.AnythingOfType("*domain.QuizSubmission")).Return(nil)

	svc := d.service()
	sub, err := svc.StartSubmission(ctx, quizID, domain.StartQuizSubmissionDTO{QuizRoomID: roomID})
	assert.NoError(t, err)
	assert.Equal(t, domain.SubmissionStatusInProgress, sub.Status)
	assert.Equal(t, studentID, sub.UserID)
}

func TestQuizService_StartSubmission_NotEnrolled_Forbidden(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	classID := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, ClassID: classID}, nil)
	d.memberRepo.On("Exists", ctx, classID, studentID).Return(false, nil)

	svc := d.service()
	_, err := svc.StartSubmission(ctx, quizID, domain.StartQuizSubmissionDTO{QuizRoomID: uuid.New()})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestQuizService_StartSubmission_RoomNotOpen_ValidationError(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	classID := uuid.New()
	roomID := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, ClassID: classID}, nil)
	d.memberRepo.On("Exists", ctx, classID, studentID).Return(true, nil)
	d.roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.QuizRoom{ID: roomID, QuizID: quizID, StartedAt: nil}, nil)

	svc := d.service()
	_, err := svc.StartSubmission(ctx, quizID, domain.StartQuizSubmissionDTO{QuizRoomID: roomID})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestQuizService_StartSubmission_RoomForDifferentQuiz_NotFound(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	otherQuizID := uuid.New()
	classID := uuid.New()
	roomID := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	now := time.Now()
	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, ClassID: classID}, nil)
	d.memberRepo.On("Exists", ctx, classID, studentID).Return(true, nil)
	d.roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.QuizRoom{ID: roomID, QuizID: otherQuizID, StartedAt: &now}, nil)

	svc := d.service()
	_, err := svc.StartSubmission(ctx, quizID, domain.StartQuizSubmissionDTO{QuizRoomID: roomID})
	assert.ErrorIs(t, err, domain.ErrNotFound)
	d.subRepo.AssertNotCalled(t, "FindByQuizAndUser")
	d.subRepo.AssertNotCalled(t, "Create")
}

func TestQuizService_StartSubmission_AlreadyExists_Conflict(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	classID := uuid.New()
	roomID := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	now := time.Now()
	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, ClassID: classID}, nil)
	d.memberRepo.On("Exists", ctx, classID, studentID).Return(true, nil)
	d.roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.QuizRoom{ID: roomID, QuizID: quizID, StartedAt: &now}, nil)
	d.subRepo.On("FindByQuizAndUser", ctx, quizID, studentID).
		Return(&domain.QuizSubmission{ID: uuid.New()}, nil)

	svc := d.service()
	_, err := svc.StartSubmission(ctx, quizID, domain.StartQuizSubmissionDTO{QuizRoomID: roomID})
	assert.ErrorIs(t, err, domain.ErrConflict)
}

func TestQuizService_SubmitQuiz_AutoGrading(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	subID := uuid.New()
	q1ID := uuid.New()
	q2ID := uuid.New()
	q3ID := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	startedAt := time.Now().Add(-5 * time.Minute)
	sub := &domain.QuizSubmission{
		ID:        subID,
		QuizID:    quizID,
		UserID:    studentID,
		Status:    domain.SubmissionStatusInProgress,
		StartedAt: startedAt,
	}

	d.subRepo.On("FindByID", ctx, subID).Return(sub, nil)
	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, DurationMinutes: 60}, nil)
	d.roomRepo.On("FindOpenByQuizID", ctx, quizID).
		Return(&domain.QuizRoom{ID: uuid.New(), QuizID: quizID}, nil)

	questions := []domain.Question{
		{
			ID: q1ID, Type: domain.QuestionTypeChoice,
			Options: []domain.QuestionOption{
				{ID: "a", Value: "4", Score: 2},
				{ID: "b", Value: "3", Score: 0},
				{ID: "c", Value: "5", Score: 0},
			},
		},
		{
			ID: q2ID, Type: domain.QuestionTypeShortAnswer,
			Options: []domain.QuestionOption{
				{ID: "x", Value: "Paris", Score: 3},
			},
		},
		{
			ID: q3ID, Type: domain.QuestionTypeDescriptive,
			Options: []domain.QuestionOption{},
		},
	}
	d.questionRepo.On("FindByIDs", ctx, mock.Anything).Return(questions, nil)
	d.subRepo.On("Update", ctx, mock.AnythingOfType("*domain.QuizSubmission")).Return(nil)

	svc := d.service()
	result, err := svc.SubmitQuiz(ctx, subID, domain.SubmitQuizDTO{
		Answers: []domain.SubmitAnswerDTO{
			{QuestionID: q1ID, SelectedOptionIDs: []string{"a"}, SpentSeconds: 30},
			{QuestionID: q2ID, Value: "  paris  ", SpentSeconds: 20},
			{QuestionID: q3ID, Value: "Long essay answer", SpentSeconds: 120},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, domain.SubmissionStatusSubmitted, result.Status)
	assert.NotNil(t, result.SubmittedAt)

	// choice: selected "a" → score 2
	assert.Equal(t, 2.0, result.Answers[0].EarnedScore)
	// short_answer: "  paris  " matches "Paris" → score 3
	assert.Equal(t, 3.0, result.Answers[1].EarnedScore)
	// descriptive: always 0
	assert.Equal(t, 0.0, result.Answers[2].EarnedScore)
	// total = 2 + 3 + 0 = 5
	assert.Equal(t, 5.0, result.TotalScore)
}

func TestQuizService_SubmitQuiz_WrongUser_Forbidden(t *testing.T) {
	studentID := uuid.New()
	otherUser := uuid.New()
	subID := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	d.subRepo.On("FindByID", ctx, subID).
		Return(&domain.QuizSubmission{ID: subID, UserID: otherUser, Status: domain.SubmissionStatusInProgress}, nil)

	svc := d.service()
	_, err := svc.SubmitQuiz(ctx, subID, domain.SubmitQuizDTO{Answers: []domain.SubmitAnswerDTO{}})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestQuizService_SubmitQuiz_AlreadySubmitted_Conflict(t *testing.T) {
	studentID := uuid.New()
	subID := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	d.subRepo.On("FindByID", ctx, subID).
		Return(&domain.QuizSubmission{ID: subID, UserID: studentID, Status: domain.SubmissionStatusSubmitted}, nil)

	svc := d.service()
	_, err := svc.SubmitQuiz(ctx, subID, domain.SubmitQuizDTO{Answers: []domain.SubmitAnswerDTO{}})
	assert.ErrorIs(t, err, domain.ErrConflict)
}

func TestQuizService_GradeSubmission_Success(t *testing.T) {
	teacherID := uuid.New()
	quizID := uuid.New()
	subID := uuid.New()
	q1ID := uuid.New()
	q2ID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	sub := &domain.QuizSubmission{
		ID:     subID,
		QuizID: quizID,
		Status: domain.SubmissionStatusSubmitted,
		Answers: []domain.SubmissionAnswer{
			{QuestionID: q1ID, EarnedScore: 2},
			{QuestionID: q2ID, EarnedScore: 0},
		},
		TotalScore: 2,
	}

	d.subRepo.On("FindByID", ctx, subID).Return(sub, nil)
	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, UserID: teacherID}, nil)
	d.subRepo.On("Update", ctx, mock.AnythingOfType("*domain.QuizSubmission")).Return(nil)

	svc := d.service()
	result, err := svc.GradeSubmission(ctx, subID, domain.GradeSubmissionDTO{
		Grades: []domain.GradeAnswerDTO{
			{QuestionID: q2ID, EarnedScore: 7.5},
		},
	})

	assert.NoError(t, err)
	assert.Equal(t, domain.SubmissionStatusGraded, result.Status)
	assert.Equal(t, 9.5, result.TotalScore) // 2 + 7.5
}

func TestQuizService_GradeSubmission_NotSubmitted_ValidationError(t *testing.T) {
	teacherID := uuid.New()
	subID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.subRepo.On("FindByID", ctx, subID).
		Return(&domain.QuizSubmission{ID: subID, Status: domain.SubmissionStatusInProgress}, nil)

	svc := d.service()
	_, err := svc.GradeSubmission(ctx, subID, domain.GradeSubmissionDTO{
		Grades: []domain.GradeAnswerDTO{},
	})
	assert.ErrorIs(t, err, domain.ErrValidation)
}

func TestQuizService_GradeSubmission_NonOwner_Forbidden(t *testing.T) {
	studentID := uuid.New()
	otherTeacher := uuid.New()
	quizID := uuid.New()
	subID := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	d.subRepo.On("FindByID", ctx, subID).
		Return(&domain.QuizSubmission{ID: subID, QuizID: quizID, Status: domain.SubmissionStatusSubmitted}, nil)
	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, UserID: otherTeacher}, nil)
	d.memberRepo.On("Exists", ctx, mock.Anything, studentID).Return(true, nil)

	svc := d.service()
	_, err := svc.GradeSubmission(ctx, subID, domain.GradeSubmissionDTO{
		Grades: []domain.GradeAnswerDTO{},
	})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestQuizService_SubmitQuiz_MultipleChoiceScoring(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	subID := uuid.New()
	qID := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	startedAt := time.Now().Add(-5 * time.Minute)
	d.subRepo.On("FindByID", ctx, subID).
		Return(&domain.QuizSubmission{ID: subID, QuizID: quizID, UserID: studentID, Status: domain.SubmissionStatusInProgress, StartedAt: startedAt}, nil)
	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, DurationMinutes: 60}, nil)
	d.roomRepo.On("FindOpenByQuizID", ctx, quizID).
		Return(&domain.QuizRoom{ID: uuid.New(), QuizID: quizID}, nil)

	questions := []domain.Question{
		{
			ID: qID, Type: domain.QuestionTypeChoice,
			Options: []domain.QuestionOption{
				{ID: "a", Value: "Option A", Score: 1},
				{ID: "b", Value: "Option B", Score: 1.5},
				{ID: "c", Value: "Option C", Score: -0.5},
				{ID: "d", Value: "Option D", Score: 0},
			},
		},
	}
	d.questionRepo.On("FindByIDs", ctx, mock.Anything).Return(questions, nil)
	d.subRepo.On("Update", ctx, mock.AnythingOfType("*domain.QuizSubmission")).Return(nil)

	svc := d.service()
	result, err := svc.SubmitQuiz(ctx, subID, domain.SubmitQuizDTO{
		Answers: []domain.SubmitAnswerDTO{
			{QuestionID: qID, SelectedOptionIDs: []string{"a", "b", "c"}, SpentSeconds: 45},
		},
	})

	assert.NoError(t, err)
	// 1 + 1.5 + (-0.5) = 2.0
	assert.Equal(t, 2.0, result.Answers[0].EarnedScore)
	assert.Equal(t, 2.0, result.TotalScore)
}

func TestQuizService_ListQuestionsForTaking_SanitizesAnswersAndDeduplicates(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	classID := uuid.New()
	manualQID := uuid.New()
	shortQID := uuid.New()
	randomQID := uuid.New()
	bankID := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, ClassID: classID, UserID: uuid.New()}, nil)
	d.memberRepo.On("Exists", ctx, classID, studentID).Return(true, nil)
	d.ruleRepo.On("ListByQuiz", ctx, quizID, domain.ListParams{Page: 1, PageSize: 10000}).
		Return([]domain.QuizRule{
			{ID: uuid.New(), QuizID: quizID, Type: domain.QuizRuleTypeManual, QuestionIDs: []uuid.UUID{manualQID, shortQID, manualQID}},
			{ID: uuid.New(), QuizID: quizID, Type: domain.QuizRuleTypeRandom, BankID: &bankID, Count: 1},
		}, int64(2), nil)
	d.questionRepo.On("FindByIDs", ctx, []uuid.UUID{manualQID, shortQID, manualQID}).
		Return([]domain.Question{
			{
				ID: manualQID, Type: domain.QuestionTypeChoice,
				Options: []domain.QuestionOption{
					{ID: "a", Value: "A", Score: 10},
					{ID: "b", Value: "B", Score: 0},
				},
			},
			{
				ID: shortQID, Type: domain.QuestionTypeShortAnswer,
				Options: []domain.QuestionOption{{ID: "answer", Value: "secret", Score: 5}},
			},
		}, nil)
	d.questionRepo.On("RandomByBank", ctx, bankID, 1).
		Return([]domain.Question{
			{
				ID: randomQID, Type: domain.QuestionTypeChoice,
				Options: []domain.QuestionOption{{ID: "r", Value: "Random", Score: 7}},
			},
		}, nil)

	svc := d.service()
	questions, err := svc.ListQuestionsForTaking(ctx, quizID)

	assert.NoError(t, err)
	assert.Len(t, questions, 3)
	assert.Equal(t, manualQID, questions[0].ID)
	assert.Equal(t, shortQID, questions[1].ID)
	assert.Equal(t, randomQID, questions[2].ID)
	assert.Equal(t, 0.0, questions[0].Options[0].Score)
	assert.Equal(t, "a", questions[0].Options[0].ID)
	assert.Equal(t, "A", questions[0].Options[0].Value)
	assert.Equal(t, []domain.QuestionOption{}, questions[1].Options)
	assert.Equal(t, 0.0, questions[2].Options[0].Score)
}
