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
func (m *mQuizRepo) ListByMemberWithRooms(ctx context.Context, userID uuid.UUID, p domain.ListParams) ([]domain.Quiz, int64, error) {
	a := m.Called(ctx, userID, p)
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

func teacherCtx(userID uuid.UUID) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID:      userID,
		Permissions: []string{"quizzes:update_any", "quizzes:create", "quizzes:view", "quizzes:delete"},
		Ent:         domain.PlanCatalog[domain.PlanKey(domain.TierPro, 50)],
	})
}

// freeTeacherCtx is a teacher on the Free plan (no advanced anti-cheat).
func freeTeacherCtx(userID uuid.UUID) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID:      userID,
		Permissions: []string{"quizzes:update_any", "quizzes:create", "quizzes:view", "quizzes:delete"},
		Ent:         domain.PlanCatalog[domain.PlanFree],
	})
}

func studentCtx(userID uuid.UUID) context.Context {
	return domain.WithCaller(context.Background(), domain.Caller{
		UserID: userID,
	})
}

type testDeps struct {
	quizRepo     *mQuizRepo
	ruleRepo     *mRuleRepo
	roomRepo     *mRoomRepo
	subRepo      *mSubRepo
	questionRepo *mQRepo
	classRepo    *mClassRepo
	memberRepo   *mMemberRepo
}

func newDeps() testDeps {
	return testDeps{
		quizRepo:     &mQuizRepo{},
		ruleRepo:     &mRuleRepo{},
		roomRepo:     &mRoomRepo{},
		subRepo:      &mSubRepo{},
		questionRepo: &mQRepo{},
		classRepo:    &mClassRepo{},
		memberRepo:   &mMemberRepo{},
	}
}

func (d testDeps) service() domain.QuizService {
	return quizzes.NewService(
		d.quizRepo, d.ruleRepo, d.roomRepo, d.subRepo,
		d.questionRepo, d.classRepo, d.memberRepo, nil, slog.Default(),
	)
}

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
	d.ruleRepo.On("ListByQuiz", ctx, quizID, mock.Anything).
		Return([]domain.QuizRule{}, int64(0), nil)
	d.subRepo.On("Create", ctx, mock.AnythingOfType("*domain.QuizSubmission")).Return(nil)

	svc := d.service()
	sub, err := svc.StartSubmission(ctx, quizID, domain.StartQuizSubmissionDTO{QuizRoomID: roomID})
	assert.NoError(t, err)
	assert.Equal(t, domain.SubmissionStatusInProgress, sub.Status)
	assert.Equal(t, studentID, sub.UserID)
	assert.Equal(t, roomID, *sub.QuizRoomID)
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
		QuestionSet: []domain.SubmissionQuestion{
			{QuestionID: q1ID}, {QuestionID: q2ID}, {QuestionID: q3ID},
		},
	}

	d.subRepo.On("FindByID", ctx, subID).Return(sub, nil)
	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, DurationMinutes: 60}, nil)
	d.roomRepo.On("FindOpenByQuizID", ctx, quizID).
		Return(&domain.QuizRoom{ID: uuid.New(), QuizID: quizID}, nil)
	d.ruleRepo.On("ListByQuiz", ctx, quizID, domain.ListParams{Page: 1, PageSize: 10000}).
		Return([]domain.QuizRule{}, int64(0), nil)

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
	graded := captureGraded(d, ctx)

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

	// Grading runs on the persisted record; the quiz does not opt into
	// ShowResults, so the response back to the student is masked.
	// choice: selected "a" → score 2
	assert.Equal(t, 2.0, graded.Answers[0].EarnedScore)
	// short_answer: "  paris  " matches "Paris" → score 3
	assert.Equal(t, 3.0, graded.Answers[1].EarnedScore)
	// descriptive: always 0
	assert.Equal(t, 0.0, graded.Answers[2].EarnedScore)
	// total = 2 + 3 + 0 = 5
	assert.Equal(t, 5.0, graded.TotalScore)

	// Student's own response is stripped until results are revealed.
	assert.False(t, result.ResultsRevealed)
	assert.Equal(t, 0.0, result.TotalScore)
	assert.Equal(t, 0.0, result.Answers[0].EarnedScore)
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
	quizID := uuid.New()
	subID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	// In-progress and still within the deadline → finalizeIfExpired is a no-op,
	// so grading a not-yet-submitted attempt is a validation error.
	d.subRepo.On("FindByID", ctx, subID).
		Return(&domain.QuizSubmission{ID: subID, QuizID: quizID, Status: domain.SubmissionStatusInProgress, StartedAt: time.Now()}, nil)
	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, DurationMinutes: 60}, nil)

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
		Return(&domain.QuizSubmission{ID: subID, QuizID: quizID, UserID: studentID, Status: domain.SubmissionStatusInProgress, StartedAt: startedAt, QuestionSet: []domain.SubmissionQuestion{{QuestionID: qID}}}, nil)
	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, DurationMinutes: 60}, nil)
	d.roomRepo.On("FindOpenByQuizID", ctx, quizID).
		Return(&domain.QuizRoom{ID: uuid.New(), QuizID: quizID}, nil)
	d.ruleRepo.On("ListByQuiz", ctx, quizID, domain.ListParams{Page: 1, PageSize: 10000}).
		Return([]domain.QuizRule{}, int64(0), nil)

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
	graded := captureGraded(d, ctx)

	svc := d.service()
	_, err := svc.SubmitQuiz(ctx, subID, domain.SubmitQuizDTO{
		Answers: []domain.SubmitAnswerDTO{
			{QuestionID: qID, SelectedOptionIDs: []string{"a", "b", "c"}, SpentSeconds: 45},
		},
	})

	assert.NoError(t, err)
	// Sign-only correctness with mode=none: positives 1 + 1.5, the -0.5 option
	// is a distractor contributing 0 (no penalty). 1 + 1.5 + 0 = 2.5
	assert.Equal(t, 2.5, graded.Answers[0].EarnedScore)
	assert.Equal(t, 2.5, graded.TotalScore)
}

func TestQuizService_SubmitQuiz_AppliesResolvedNegativeMarking(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	subID := uuid.New()
	qID := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	startedAt := time.Now().Add(-5 * time.Minute)
	d.subRepo.On("FindByID", ctx, subID).
		Return(&domain.QuizSubmission{ID: subID, QuizID: quizID, UserID: studentID, Status: domain.SubmissionStatusInProgress, StartedAt: startedAt, QuestionSet: []domain.SubmissionQuestion{{QuestionID: qID}}}, nil)
	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, DurationMinutes: 60}, nil)
	d.roomRepo.On("FindOpenByQuizID", ctx, quizID).
		Return(&domain.QuizRoom{ID: uuid.New(), QuizID: quizID}, nil)
	d.ruleRepo.On("ListByQuiz", ctx, quizID, domain.ListParams{Page: 1, PageSize: 10000}).
		Return([]domain.QuizRule{
			{
				ID: uuid.New(), QuizID: quizID, Type: domain.QuizRuleTypeManual, QuestionIDs: []uuid.UUID{qID},
				NegativeOverrides: []domain.QuizQuestionNegativeOverride{
					{QuestionID: qID, Mode: domain.NegativeMarkPerWrong, NegativeValue: 0.5},
				},
			},
		}, int64(1), nil)

	questions := []domain.Question{
		{
			ID: qID, Type: domain.QuestionTypeChoice,
			Options: []domain.QuestionOption{
				{ID: "a", Value: "correct", Score: 1},
				{ID: "b", Value: "wrong", Score: 0},
			},
		},
	}
	d.questionRepo.On("FindByIDs", ctx, mock.Anything).Return(questions, nil)
	graded := captureGraded(d, ctx)

	svc := d.service()
	_, err := svc.SubmitQuiz(ctx, subID, domain.SubmitQuizDTO{
		Answers: []domain.SubmitAnswerDTO{
			{QuestionID: qID, SelectedOptionIDs: []string{"a", "b"}, SpentSeconds: 30},
		},
	})
	assert.NoError(t, err)
	// positive 1 - per_wrong penalty 0.5*1 = 0.5
	assert.Equal(t, 0.5, graded.Answers[0].EarnedScore)
	assert.Equal(t, 0.5, graded.TotalScore)
}

func TestQuizService_ListQuestionsForTaking_AttachesNegativeConfig(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	classID := uuid.New()
	qID := uuid.New()
	// Manager preview path (listQuestionsComposed): composes rules live.
	ctx := teacherCtx(studentID)
	d := newDeps()

	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, ClassID: classID, UserID: uuid.New()}, nil)
	d.memberRepo.On("Exists", ctx, classID, studentID).Return(true, nil)
	d.ruleRepo.On("ListByQuiz", ctx, quizID, domain.ListParams{Page: 1, PageSize: 10000}).
		Return([]domain.QuizRule{
			{
				ID: uuid.New(), QuizID: quizID, Type: domain.QuizRuleTypeManual, QuestionIDs: []uuid.UUID{qID},
				NegativeOverrides: []domain.QuizQuestionNegativeOverride{
					{QuestionID: qID, Mode: domain.NegativeMarkPerWrong, NegativeValue: 0.5},
				},
			},
		}, int64(1), nil)
	d.questionRepo.On("FindByIDs", ctx, []uuid.UUID{qID}).
		Return([]domain.Question{
			{
				ID: qID, Type: domain.QuestionTypeChoice,
				Options: []domain.QuestionOption{
					{ID: "a", Value: "A", Score: 1},
					{ID: "b", Value: "B", Score: 0},
				},
			},
		}, nil)

	svc := d.service()
	questions, err := svc.ListQuestionsForTaking(ctx, quizID)
	assert.NoError(t, err)
	assert.Len(t, questions, 1)
	if assert.NotNil(t, questions[0].NegativeConfig) {
		assert.Equal(t, domain.NegativeMarkPerWrong, questions[0].NegativeConfig.Mode)
		assert.Equal(t, 0.5, questions[0].NegativeConfig.Fraction)
	}
	// answer key still stripped
	assert.Equal(t, 0.0, questions[0].Options[0].Score)
}

func TestQuizService_TakePreview_EnrolledStudentGetsCountAndNegativeFlag(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	classID := uuid.New()
	qID := uuid.New()
	shortQID := uuid.New()
	// Enrolled student (no manage/view_any) — must resolve preview without a
	// submission and without leaking question bodies.
	ctx := studentCtx(studentID)
	d := newDeps()

	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, ClassID: classID, UserID: uuid.New()}, nil)
	d.memberRepo.On("Exists", ctx, classID, studentID).Return(true, nil)
	d.ruleRepo.On("ListByQuiz", ctx, quizID, domain.ListParams{Page: 1, PageSize: 10000}).
		Return([]domain.QuizRule{
			{
				ID: uuid.New(), QuizID: quizID, Type: domain.QuizRuleTypeManual, QuestionIDs: []uuid.UUID{qID, shortQID},
				NegativeOverrides: []domain.QuizQuestionNegativeOverride{
					{QuestionID: qID, Mode: domain.NegativeMarkPerWrong, NegativeValue: 0.5},
				},
			},
		}, int64(1), nil)
	d.questionRepo.On("FindByIDs", ctx, []uuid.UUID{qID, shortQID}).
		Return([]domain.Question{
			{
				ID: qID, Type: domain.QuestionTypeChoice,
				Options: []domain.QuestionOption{
					{ID: "a", Value: "A", Score: 1},
					{ID: "b", Value: "B", Score: 0},
				},
			},
			{ID: shortQID, Type: domain.QuestionTypeShortAnswer, Options: []domain.QuestionOption{{ID: "answer", Value: "secret", Score: 5}}},
		}, nil)

	svc := d.service()
	preview, err := svc.TakePreview(ctx, quizID)
	assert.NoError(t, err)
	if assert.NotNil(t, preview) {
		assert.Equal(t, 2, preview.QuestionCount)
		assert.True(t, preview.HasNegativeMarking)
	}
}

func TestQuizService_TakePreview_NoNegativeMarking(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	classID := uuid.New()
	qID := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, ClassID: classID, UserID: uuid.New()}, nil)
	d.memberRepo.On("Exists", ctx, classID, studentID).Return(true, nil)
	d.ruleRepo.On("ListByQuiz", ctx, quizID, domain.ListParams{Page: 1, PageSize: 10000}).
		Return([]domain.QuizRule{
			{ID: uuid.New(), QuizID: quizID, Type: domain.QuizRuleTypeManual, QuestionIDs: []uuid.UUID{qID}},
		}, int64(1), nil)
	d.questionRepo.On("FindByIDs", ctx, []uuid.UUID{qID}).
		Return([]domain.Question{
			{ID: qID, Type: domain.QuestionTypeChoice, Options: []domain.QuestionOption{{ID: "a", Value: "A", Score: 1}, {ID: "b", Value: "B", Score: 0}}},
		}, nil)

	svc := d.service()
	preview, err := svc.TakePreview(ctx, quizID)
	assert.NoError(t, err)
	if assert.NotNil(t, preview) {
		assert.Equal(t, 1, preview.QuestionCount)
		assert.False(t, preview.HasNegativeMarking)
	}
}

func TestQuizService_TakePreview_NotEnrolledForbidden(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	classID := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, ClassID: classID, UserID: uuid.New()}, nil)
	d.memberRepo.On("Exists", ctx, classID, studentID).Return(false, nil)

	svc := d.service()
	_, err := svc.TakePreview(ctx, quizID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestQuizService_ListQuestionsForTaking_SanitizesAnswersAndDeduplicates(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	classID := uuid.New()
	manualQID := uuid.New()
	shortQID := uuid.New()
	randomQID := uuid.New()
	bankID := uuid.New()
	// Manager preview path (listQuestionsComposed): composes + dedups rules live.
	ctx := teacherCtx(studentID)
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

func TestQuizService_ListMine_DerivesStates(t *testing.T) {
	studentID := uuid.New()
	q1, q2 := uuid.New(), uuid.New()
	c1 := uuid.New()
	r1, s1 := uuid.New(), uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	openQuiz := domain.Quiz{ID: q1, Title: "Open", ClassID: c1, Class: &domain.Class{Name: "Math"}, DurationMinutes: 30, TotalScore: 20}
	// gradedQuiz opts into showing results; its room has already closed so the
	// score is revealed to the student.
	gradedQuiz := domain.Quiz{ID: q2, Title: "Done", ClassID: c1, Class: &domain.Class{Name: "Math"}, DurationMinutes: 30, TotalScore: 20, ShowResults: true}

	d.quizRepo.On("ListByMemberWithRooms", mock.Anything, studentID, mock.Anything).
		Return([]domain.Quiz{openQuiz, gradedQuiz}, int64(2), nil)

	start := time.Now().Add(-time.Minute)
	end := time.Now().Add(time.Hour)
	d.roomRepo.On("ListByQuiz", mock.Anything, q1, mock.Anything).
		Return([]domain.QuizRoom{{ID: r1, ClassSessionID: s1, StartedAt: &start, EndedAt: &end}}, int64(1), nil)
	d.roomRepo.On("ListByQuiz", mock.Anything, q2, mock.Anything).
		Return([]domain.QuizRoom{}, int64(0), nil)

	q2RoomID := uuid.New()
	pastEnd := time.Now().Add(-time.Minute)
	d.roomRepo.On("FindByID", mock.Anything, q2RoomID).
		Return(&domain.QuizRoom{ID: q2RoomID, QuizID: q2, EndedAt: &pastEnd}, nil)

	submittedAt := time.Now()
	d.subRepo.On("FindByQuizAndUser", mock.Anything, q1, studentID).
		Return(nil, domain.ErrNotFound)
	d.subRepo.On("FindByQuizAndUser", mock.Anything, q2, studentID).
		Return(&domain.QuizSubmission{Status: domain.SubmissionStatusGraded, TotalScore: 18, SubmittedAt: &submittedAt, QuizRoomID: &q2RoomID}, nil)

	svc := d.service()
	exams, total, err := svc.ListMine(ctx, domain.ListParams{Page: 1, PageSize: 20})
	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, exams, 2)

	assert.Equal(t, domain.MyExamStateOpen, exams[0].State)
	assert.NotNil(t, exams[0].Room)
	assert.True(t, exams[0].Room.IsOpen)
	assert.Equal(t, "Math", exams[0].ClassName)

	assert.Equal(t, domain.MyExamStateGraded, exams[1].State)
	assert.NotNil(t, exams[1].Score)
	assert.Equal(t, 18.0, *exams[1].Score)
}

func TestQuizService_ListMine_NoCaller_Forbidden(t *testing.T) {
	d := newDeps()
	svc := d.service()
	_, _, err := svc.ListMine(context.Background(), domain.ListParams{Page: 1, PageSize: 20})
	assert.ErrorIs(t, err, domain.ErrForbidden)
}

// captureGraded records the submission as persisted by SubmitQuiz's grading,
// before the response-time score masking mutates the same pointer. Assert on
// the returned value to verify grading math independent of result visibility.
func captureGraded(d testDeps, ctx context.Context) *domain.QuizSubmission {
	graded := &domain.QuizSubmission{}
	d.subRepo.On("Update", ctx, mock.AnythingOfType("*domain.QuizSubmission")).
		Run(func(args mock.Arguments) {
			s := args.Get(1).(*domain.QuizSubmission)
			*graded = *s
			graded.Answers = append([]domain.SubmissionAnswer(nil), s.Answers...)
		}).Return(nil)
	return graded
}

func optionIDs(opts []domain.QuestionOption) []string {
	out := make([]string, len(opts))
	for i, o := range opts {
		out[i] = o.ID
	}
	return out
}

func TestQuizService_Create_CopiesAntiCheatToggles(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	orgID := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, OrganizationID: orgID, UserID: teacherID}, nil)
	d.quizRepo.On("Create", ctx, mock.AnythingOfType("*domain.Quiz")).
		Run(func(args mock.Arguments) {
			q := args.Get(1).(*domain.Quiz)
			assert.True(t, q.ShuffleOptions)
			assert.True(t, q.TrackTabSwitches)
			assert.True(t, q.DisableCopyPaste)
		}).Return(nil)

	svc := d.service()
	_, err := svc.Create(ctx, domain.CreateQuizDTO{
		ClassID:          classID,
		Title:            "Exam",
		DurationMinutes:  30,
		ShuffleOptions:   true,
		TrackTabSwitches: true,
		DisableCopyPaste: true,
	})
	assert.NoError(t, err)
}

func TestQuizService_Create_FreePlanRejectsAdvancedAntiCheat(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	orgID := uuid.New()
	ctx := freeTeacherCtx(teacherID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, OrganizationID: orgID, UserID: teacherID}, nil)

	svc := d.service()
	// TrackTabSwitches is an advanced (backend-cost) anti-cheat signal → gated.
	_, err := svc.Create(ctx, domain.CreateQuizDTO{
		ClassID:          classID,
		Title:            "Exam",
		DurationMinutes:  30,
		TrackTabSwitches: true,
	})
	assert.ErrorIs(t, err, domain.ErrFeatureNotInPlan)
	d.quizRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestQuizService_Create_FreePlanAllowsBasicAntiCheat(t *testing.T) {
	teacherID := uuid.New()
	classID := uuid.New()
	orgID := uuid.New()
	ctx := freeTeacherCtx(teacherID)
	d := newDeps()

	d.classRepo.On("FindByID", ctx, classID).
		Return(&domain.Class{ID: classID, OrganizationID: orgID, UserID: teacherID}, nil)
	d.quizRepo.On("Create", ctx, mock.AnythingOfType("*domain.Quiz")).Return(nil)

	svc := d.service()
	// Frontend-only deterrents are not gated.
	_, err := svc.Create(ctx, domain.CreateQuizDTO{
		ClassID:          classID,
		Title:            "Exam",
		DurationMinutes:  30,
		DisableCopyPaste: true,
	})
	assert.NoError(t, err)
}

func TestQuizService_StartSubmission_FreezesQuestionSetAndGPS(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	classID := uuid.New()
	roomID := uuid.New()
	bankID := uuid.New()
	q1 := domain.Question{ID: uuid.New(), Type: domain.QuestionTypeChoice,
		Options: []domain.QuestionOption{{ID: "o1"}, {ID: "o2"}, {ID: "o3"}}}
	ctx := studentCtx(studentID)
	d := newDeps()

	now := time.Now()
	lat, lng, acc := 35.7, 51.4, 12.0
	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, ClassID: classID, ShuffleOptions: true, RequireGPS: true}, nil)
	d.memberRepo.On("Exists", ctx, classID, studentID).Return(true, nil)
	d.roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.QuizRoom{ID: roomID, QuizID: quizID, StartedAt: &now}, nil)
	d.subRepo.On("FindByQuizAndUser", ctx, quizID, studentID).
		Return((*domain.QuizSubmission)(nil), domain.ErrNotFound)
	d.ruleRepo.On("ListByQuiz", ctx, quizID, mock.Anything).
		Return([]domain.QuizRule{{Type: domain.QuizRuleTypeRandom, BankID: &bankID, Count: 1}}, int64(1), nil)
	d.questionRepo.On("RandomByBank", ctx, bankID, 1).Return([]domain.Question{q1}, nil)
	d.subRepo.On("Create", ctx, mock.AnythingOfType("*domain.QuizSubmission")).
		Run(func(args mock.Arguments) {
			sub := args.Get(1).(*domain.QuizSubmission)
			assert.Len(t, sub.QuestionSet, 1)
			assert.Equal(t, q1.ID, sub.QuestionSet[0].QuestionID)
			assert.ElementsMatch(t, []string{"o1", "o2", "o3"}, sub.QuestionSet[0].OptionIDOrder)
			assert.Equal(t, lat, *sub.GPSLat)
			assert.Equal(t, acc, *sub.GPSAccuracy)
			assert.Equal(t, roomID, *sub.QuizRoomID)
		}).Return(nil)

	svc := d.service()
	_, err := svc.StartSubmission(ctx, quizID, domain.StartQuizSubmissionDTO{
		QuizRoomID: roomID, GPSLat: &lat, GPSLng: &lng, GPSAccuracy: &acc,
	})
	assert.NoError(t, err)
}

func TestQuizService_StartSubmission_RequireGPSRejectsMissingCoords(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	classID := uuid.New()
	roomID := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	now := time.Now()
	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, ClassID: classID, RequireGPS: true}, nil)
	d.memberRepo.On("Exists", ctx, classID, studentID).Return(true, nil)
	d.roomRepo.On("FindByID", ctx, roomID).
		Return(&domain.QuizRoom{ID: roomID, QuizID: quizID, StartedAt: &now}, nil)
	d.subRepo.On("FindByQuizAndUser", ctx, quizID, studentID).
		Return((*domain.QuizSubmission)(nil), domain.ErrNotFound)

	svc := d.service()
	_, err := svc.StartSubmission(ctx, quizID, domain.StartQuizSubmissionDTO{QuizRoomID: roomID})
	assert.ErrorIs(t, err, domain.ErrValidation)
	d.subRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestQuizService_ListQuestionsForTaking_UsesFrozenSetAndOrder(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	classID := uuid.New()
	qid := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	frozen := []domain.SubmissionQuestion{{QuestionID: qid, OptionIDOrder: []string{"o3", "o1", "o2"}}}
	d.quizRepo.On("FindByID", ctx, quizID).Return(&domain.Quiz{ID: quizID, ClassID: classID, UserID: uuid.New()}, nil)
	d.memberRepo.On("Exists", ctx, classID, studentID).Return(true, nil)
	d.subRepo.On("FindByQuizAndUser", ctx, quizID, studentID).
		Return(&domain.QuizSubmission{ID: uuid.New(), QuizID: quizID, UserID: studentID,
			Status: domain.SubmissionStatusInProgress, QuestionSet: frozen}, nil)
	d.ruleRepo.On("ListByQuiz", ctx, quizID, mock.Anything).Return([]domain.QuizRule{}, int64(0), nil)
	d.questionRepo.On("FindByIDs", ctx, []uuid.UUID{qid}).
		Return([]domain.Question{{ID: qid, Type: domain.QuestionTypeChoice,
			Options: []domain.QuestionOption{{ID: "o1", Value: "A", Score: 1}, {ID: "o2", Value: "B"}, {ID: "o3", Value: "C"}}}}, nil)

	svc := d.service()
	qs, err := svc.ListQuestionsForTaking(ctx, quizID)
	assert.NoError(t, err)
	assert.Len(t, qs, 1)
	assert.Equal(t, []string{"o3", "o1", "o2"}, optionIDs(qs[0].Options))
	for _, o := range qs[0].Options {
		assert.Zero(t, o.Score)
	}
}

func TestQuizService_SaveAnswer_UpsertsAndKeepsInProgress(t *testing.T) {
	studentID := uuid.New()
	subID := uuid.New()
	quizID := uuid.New()
	qid := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	existing := &domain.QuizSubmission{ID: subID, QuizID: quizID, UserID: studentID, Status: domain.SubmissionStatusInProgress,
		StartedAt: time.Now(), QuestionSet: []domain.SubmissionQuestion{{QuestionID: qid}}, Answers: []domain.SubmissionAnswer{}}
	d.subRepo.On("FindByID", ctx, subID).Return(existing, nil)
	d.quizRepo.On("FindByID", ctx, quizID).Return(&domain.Quiz{ID: quizID, DurationMinutes: 60}, nil)
	d.subRepo.On("Update", ctx, mock.AnythingOfType("*domain.QuizSubmission")).
		Run(func(args mock.Arguments) {
			sub := args.Get(1).(*domain.QuizSubmission)
			assert.Equal(t, domain.SubmissionStatusInProgress, sub.Status)
			assert.Len(t, sub.Answers, 1)
			assert.Equal(t, []string{"o1"}, sub.Answers[0].SelectedOptionIDs)
			assert.Zero(t, sub.Answers[0].EarnedScore)
			assert.Equal(t, 2, sub.TabHiddenCount)
		}).Return(nil)

	svc := d.service()
	err := svc.SaveAnswer(ctx, subID, domain.SaveAnswerDTO{
		QuestionID: qid, SelectedOptionIDs: []string{"o1"}, SpentSeconds: 9, TabHiddenCount: 2, TabHiddenSeconds: 40,
	})
	assert.NoError(t, err)
}

func TestQuizService_SaveAnswer_NoBackNavigation_RejectsOverwrite(t *testing.T) {
	studentID := uuid.New()
	subID := uuid.New()
	quizID := uuid.New()
	qid := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	existing := &domain.QuizSubmission{ID: subID, QuizID: quizID, UserID: studentID, Status: domain.SubmissionStatusInProgress,
		StartedAt:   time.Now(),
		QuestionSet: []domain.SubmissionQuestion{{QuestionID: qid}},
		Answers:     []domain.SubmissionAnswer{{QuestionID: qid, SelectedOptionIDs: []string{"o1"}}}}
	d.subRepo.On("FindByID", ctx, subID).Return(existing, nil)
	d.quizRepo.On("FindByID", ctx, quizID).Return(&domain.Quiz{ID: quizID, DurationMinutes: 60, NoBackNavigation: true}, nil)

	svc := d.service()
	err := svc.SaveAnswer(ctx, subID, domain.SaveAnswerDTO{QuestionID: qid, SelectedOptionIDs: []string{"o2"}})
	assert.ErrorIs(t, err, domain.ErrValidation)
	d.subRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestQuizService_SaveAnswer_RejectsPastDeadline(t *testing.T) {
	studentID := uuid.New()
	subID := uuid.New()
	quizID := uuid.New()
	qid := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	sub := &domain.QuizSubmission{ID: subID, QuizID: quizID, UserID: studentID,
		Status: domain.SubmissionStatusInProgress, StartedAt: time.Now().Add(-2 * time.Hour),
		QuestionSet: []domain.SubmissionQuestion{{QuestionID: qid}}, Answers: []domain.SubmissionAnswer{}}
	d.subRepo.On("FindByID", ctx, subID).Return(sub, nil)
	d.quizRepo.On("FindByID", ctx, quizID).Return(&domain.Quiz{ID: quizID, DurationMinutes: 30}, nil)
	d.ruleRepo.On("ListByQuiz", ctx, quizID, mock.Anything).Return([]domain.QuizRule{}, int64(0), nil)
	d.subRepo.On("Update", ctx, mock.AnythingOfType("*domain.QuizSubmission")).Return(nil)

	svc := d.service()
	err := svc.SaveAnswer(ctx, subID, domain.SaveAnswerDTO{QuestionID: qid, SelectedOptionIDs: []string{"o1"}})
	assert.ErrorIs(t, err, domain.ErrConflict)
}

func TestQuizService_GetSubmission_LazyFinalizesPastDeadline(t *testing.T) {
	teacherID := uuid.New()
	subID := uuid.New()
	quizID := uuid.New()
	qid := uuid.New()
	ctx := teacherCtx(teacherID)
	d := newDeps()

	started := time.Now().Add(-2 * time.Hour)
	sub := &domain.QuizSubmission{ID: subID, QuizID: quizID, UserID: uuid.New(), Status: domain.SubmissionStatusInProgress,
		StartedAt: started, QuestionSet: []domain.SubmissionQuestion{{QuestionID: qid}},
		Answers: []domain.SubmissionAnswer{{QuestionID: qid, SelectedOptionIDs: []string{"o1"}}}}
	d.subRepo.On("FindByID", ctx, subID).Return(sub, nil)
	d.quizRepo.On("FindByID", ctx, quizID).Return(&domain.Quiz{ID: quizID, UserID: teacherID, DurationMinutes: 30}, nil)
	d.questionRepo.On("FindByIDs", ctx, []uuid.UUID{qid}).
		Return([]domain.Question{{ID: qid, Type: domain.QuestionTypeChoice,
			Options: []domain.QuestionOption{{ID: "o1", Score: 1}}}}, nil)
	d.ruleRepo.On("ListByQuiz", ctx, quizID, mock.Anything).Return([]domain.QuizRule{}, int64(0), nil)
	d.subRepo.On("Update", ctx, mock.AnythingOfType("*domain.QuizSubmission")).Return(nil)

	svc := d.service()
	got, err := svc.GetSubmission(ctx, subID)
	assert.NoError(t, err)
	assert.Equal(t, domain.SubmissionStatusSubmitted, got.Status)
	assert.NotNil(t, got.SubmittedAt)
	assert.Equal(t, float64(1), got.TotalScore)
}

func TestQuizService_SubmitQuiz_MergesSavedAnswersAndFiltersFrozenSet(t *testing.T) {
	studentID := uuid.New()
	subID := uuid.New()
	quizID := uuid.New()
	q1, q2, qOutside := uuid.New(), uuid.New(), uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	sub := &domain.QuizSubmission{ID: subID, QuizID: quizID, UserID: studentID,
		Status: domain.SubmissionStatusInProgress, StartedAt: time.Now(),
		QuestionSet: []domain.SubmissionQuestion{{QuestionID: q1}, {QuestionID: q2}},
		Answers:     []domain.SubmissionAnswer{{QuestionID: q1, SelectedOptionIDs: []string{"o1"}}}}
	now := time.Now()
	d.subRepo.On("FindByID", ctx, subID).Return(sub, nil)
	d.quizRepo.On("FindByID", ctx, quizID).Return(&domain.Quiz{ID: quizID, DurationMinutes: 30}, nil)
	d.roomRepo.On("FindOpenByQuizID", ctx, quizID).Return(&domain.QuizRoom{ID: uuid.New(), QuizID: quizID, StartedAt: &now}, nil)
	d.ruleRepo.On("ListByQuiz", ctx, quizID, mock.Anything).Return([]domain.QuizRule{}, int64(0), nil)
	d.questionRepo.On("FindByIDs", ctx, mock.Anything).
		Return([]domain.Question{
			{ID: q1, Type: domain.QuestionTypeChoice, Options: []domain.QuestionOption{{ID: "o1", Score: 1}}},
			{ID: q2, Type: domain.QuestionTypeChoice, Options: []domain.QuestionOption{{ID: "o2", Score: 1}}},
		}, nil)
	d.subRepo.On("Update", ctx, mock.AnythingOfType("*domain.QuizSubmission")).
		Run(func(args mock.Arguments) {
			got := args.Get(1).(*domain.QuizSubmission)
			assert.Equal(t, domain.SubmissionStatusSubmitted, got.Status)
			assert.Len(t, got.Answers, 2)
			for _, a := range got.Answers {
				assert.NotEqual(t, qOutside, a.QuestionID)
			}
		}).Return(nil)

	svc := d.service()
	_, err := svc.SubmitQuiz(ctx, subID, domain.SubmitQuizDTO{Answers: []domain.SubmitAnswerDTO{
		{QuestionID: q2, SelectedOptionIDs: []string{"o2"}},
		{QuestionID: qOutside, SelectedOptionIDs: []string{"oX"}},
	}})
	assert.NoError(t, err)
}

func TestQuizService_SubmitQuiz_NoBackNavigation_IgnoresOverwriteOfCommittedAnswer(t *testing.T) {
	studentID := uuid.New()
	subID := uuid.New()
	quizID := uuid.New()
	q1, q2 := uuid.New(), uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	sub := &domain.QuizSubmission{ID: subID, QuizID: quizID, UserID: studentID,
		Status: domain.SubmissionStatusInProgress, StartedAt: time.Now(),
		QuestionSet: []domain.SubmissionQuestion{{QuestionID: q1}, {QuestionID: q2}},
		Answers:     []domain.SubmissionAnswer{{QuestionID: q1, SelectedOptionIDs: []string{"o1"}}}}
	now := time.Now()
	d.subRepo.On("FindByID", ctx, subID).Return(sub, nil)
	d.quizRepo.On("FindByID", ctx, quizID).Return(&domain.Quiz{ID: quizID, DurationMinutes: 30, NoBackNavigation: true}, nil)
	d.roomRepo.On("FindOpenByQuizID", ctx, quizID).Return(&domain.QuizRoom{ID: uuid.New(), QuizID: quizID, StartedAt: &now}, nil)
	d.ruleRepo.On("ListByQuiz", ctx, quizID, mock.Anything).Return([]domain.QuizRule{}, int64(0), nil)
	d.questionRepo.On("FindByIDs", ctx, mock.Anything).
		Return([]domain.Question{
			{ID: q1, Type: domain.QuestionTypeChoice, Options: []domain.QuestionOption{{ID: "o1", Score: 1}, {ID: "oBad"}}},
			{ID: q2, Type: domain.QuestionTypeChoice, Options: []domain.QuestionOption{{ID: "o2", Score: 1}}},
		}, nil)
	d.subRepo.On("Update", ctx, mock.AnythingOfType("*domain.QuizSubmission")).
		Run(func(args mock.Arguments) {
			got := args.Get(1).(*domain.QuizSubmission)
			assert.Len(t, got.Answers, 2)
			for _, a := range got.Answers {
				if a.QuestionID == q1 {
					// committed answer must survive the back-nav overwrite attempt
					assert.Equal(t, []string{"o1"}, a.SelectedOptionIDs)
				}
			}
		}).Return(nil)

	svc := d.service()
	// Client tries to rewrite the already-committed q1 and add q2.
	_, err := svc.SubmitQuiz(ctx, subID, domain.SubmitQuizDTO{Answers: []domain.SubmitAnswerDTO{
		{QuestionID: q1, SelectedOptionIDs: []string{"oBad"}},
		{QuestionID: q2, SelectedOptionIDs: []string{"o2"}},
	}})
	assert.NoError(t, err)
}

func TestQuizService_AntiCheatReport_ForbiddenForNonManager(t *testing.T) {
	otherTeacherID := uuid.New()
	quizID := uuid.New()
	ctx := studentCtx(otherTeacherID)
	d := newDeps()

	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, UserID: uuid.New()}, nil)

	svc := d.service()
	_, err := svc.AntiCheatReport(ctx, quizID)
	assert.ErrorIs(t, err, domain.ErrForbidden)
	d.subRepo.AssertNotCalled(t, "ListByQuiz", mock.Anything, mock.Anything, mock.Anything)
}

func TestQuizService_SubmitQuiz_DescriptiveSimilarity_PersistedButStrippedForStudent(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	subID := uuid.New()
	qID := uuid.New()
	ctx := studentCtx(studentID)
	d := newDeps()

	sub := &domain.QuizSubmission{
		ID:          subID,
		QuizID:      quizID,
		UserID:      studentID,
		Status:      domain.SubmissionStatusInProgress,
		StartedAt:   time.Now().Add(-5 * time.Minute),
		QuestionSet: []domain.SubmissionQuestion{{QuestionID: qID}},
	}

	d.subRepo.On("FindByID", ctx, subID).Return(sub, nil)
	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, DurationMinutes: 60}, nil)
	d.roomRepo.On("FindOpenByQuizID", ctx, quizID).
		Return(&domain.QuizRoom{ID: uuid.New(), QuizID: quizID}, nil)
	d.ruleRepo.On("ListByQuiz", ctx, quizID, domain.ListParams{Page: 1, PageSize: 10000}).
		Return([]domain.QuizRule{}, int64(0), nil)
	d.questionRepo.On("FindByIDs", ctx, mock.Anything).Return([]domain.Question{
		{
			ID: qID, Type: domain.QuestionTypeDescriptive,
			ModelAnswer: "گیاهان با فتوسنتز غذا می‌سازند",
			Options: []domain.QuestionOption{
				{ID: "c1", Score: 3},
			},
		},
	}, nil)

	var persisted []domain.SubmissionAnswer
	d.subRepo.On("Update", ctx, mock.AnythingOfType("*domain.QuizSubmission")).
		Run(func(args mock.Arguments) {
			got := args.Get(1).(*domain.QuizSubmission)
			persisted = append([]domain.SubmissionAnswer(nil), got.Answers...)
		}).Return(nil)

	svc := d.service()
	result, err := svc.SubmitQuiz(ctx, subID, domain.SubmitQuizDTO{
		Answers: []domain.SubmitAnswerDTO{
			{QuestionID: qID, Value: "گیاهان از طریق فتوسنتز غذا تولید می‌کنند", SpentSeconds: 60},
		},
	})

	assert.NoError(t, err)
	// persisted answer carries the advisory similarity hint
	if assert.Len(t, persisted, 1) {
		assert.NotNil(t, persisted[0].SimilarityPct)
	}
	// earned score still 0 and total unaffected
	assert.Equal(t, 0.0, persisted[0].EarnedScore)
	assert.Equal(t, 0.0, result.TotalScore)
	// student-facing return is stripped
	assert.Nil(t, result.Answers[0].SimilarityPct)
}

func TestQuizService_GetSubmission_SimilarityVisibilityByRole(t *testing.T) {
	studentID := uuid.New()
	teacherID := uuid.New()
	quizID := uuid.New()
	subID := uuid.New()
	sim := 80.0

	makeSub := func() *domain.QuizSubmission {
		return &domain.QuizSubmission{
			ID: subID, QuizID: quizID, UserID: studentID,
			Status: domain.SubmissionStatusSubmitted,
			Answers: []domain.SubmissionAnswer{{
				QuestionID:    uuid.New(),
				SimilarityPct: &sim,
			}},
		}
	}

	t.Run("owner student gets stripped", func(t *testing.T) {
		ctx := studentCtx(studentID)
		d := newDeps()
		d.subRepo.On("FindByID", ctx, subID).Return(makeSub(), nil)
		d.quizRepo.On("FindByID", ctx, quizID).Return(&domain.Quiz{ID: quizID}, nil)
		svc := d.service()
		got, err := svc.GetSubmission(ctx, subID)
		assert.NoError(t, err)
		assert.Nil(t, got.Answers[0].SimilarityPct)
	})

	t.Run("manager sees similarity", func(t *testing.T) {
		ctx := teacherCtx(teacherID)
		d := newDeps()
		d.subRepo.On("FindByID", ctx, subID).Return(makeSub(), nil)
		d.quizRepo.On("FindByID", ctx, quizID).
			Return(&domain.Quiz{ID: quizID, UserID: teacherID}, nil)
		svc := d.service()
		got, err := svc.GetSubmission(ctx, subID)
		assert.NoError(t, err)
		if assert.NotNil(t, got.Answers[0].SimilarityPct) {
			assert.Equal(t, 80.0, *got.Answers[0].SimilarityPct)
		}
	})
}

func TestQuizService_ListSubmissions_StripsSimilarityForStudent(t *testing.T) {
	studentID := uuid.New()
	quizID := uuid.New()
	sim := 80.0
	ctx := studentCtx(studentID)
	d := newDeps()

	d.quizRepo.On("FindByID", ctx, quizID).
		Return(&domain.Quiz{ID: quizID, UserID: uuid.New()}, nil)
	d.memberRepo.On("Exists", ctx, mock.Anything, studentID).Return(true, nil)
	d.subRepo.On("ListByQuiz", ctx, quizID, mock.Anything).
		Return([]domain.QuizSubmission{{
			ID: uuid.New(), QuizID: quizID, UserID: studentID,
			Status:  domain.SubmissionStatusSubmitted,
			Answers: []domain.SubmissionAnswer{{QuestionID: uuid.New(), SimilarityPct: &sim}},
		}}, int64(1), nil)

	svc := d.service()
	subs, _, err := svc.ListSubmissions(ctx, quizID, domain.ListSubmissionsQuery{})
	assert.NoError(t, err)
	assert.Nil(t, subs[0].Answers[0].SimilarityPct)
}
