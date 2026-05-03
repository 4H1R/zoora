//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/4H1R/zoora/internal/classes"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/questionbanks"
	"github.com/4H1R/zoora/internal/quizzes"
	"github.com/4H1R/zoora/internal/users"
	"github.com/4H1R/zoora/tests/testutil"
)

type quizRepos struct {
	quizzes     domain.QuizRepository
	rules       domain.QuizRuleRepository
	rooms       domain.QuizRoomRepository
	submissions domain.QuizSubmissionRepository
	classes     domain.ClassRepository
	sessions    domain.ClassSessionRepository
	members     domain.ClassMemberRepository
	banks       domain.QuestionBankRepository
	questions   domain.QuestionRepository
	users       domain.UserRepository
	orgs        domain.OrganizationRepository
}

func setupQuizzesDB(t *testing.T) quizRepos {
	t.Helper()
	db := testutil.SetupPostgres(t)
	require.NoError(t, db.AutoMigrate(
		&domain.Organization{},
		&domain.User{},
		&domain.Class{},
		&domain.ClassSession{},
		&domain.ClassMember{},
		&domain.QuestionBank{},
		&domain.Question{},
		&domain.Quiz{},
		&domain.QuizRule{},
		&domain.QuizRoom{},
		&domain.QuizSubmission{},
	))
	return quizRepos{
		quizzes:     quizzes.NewRepository(db),
		rules:       quizzes.NewRuleRepository(db),
		rooms:       quizzes.NewRoomRepository(db),
		submissions: quizzes.NewSubmissionRepository(db),
		classes:     classes.NewRepository(db),
		sessions:    classes.NewSessionRepository(db),
		members:     classes.NewMemberRepository(db),
		banks:       questionbanks.NewRepository(db),
		questions:   questionbanks.NewQuestionRepository(db),
		users:       users.NewRepository(db),
		orgs:        organizations.NewRepository(db),
	}
}

type quizFixture struct {
	org     *domain.Organization
	teacher *domain.User
	student *domain.User
	class   *domain.Class
	session *domain.ClassSession
	bank    *domain.QuestionBank
	quiz    *domain.Quiz
}

func seedQuizFixture(t *testing.T, r quizRepos) quizFixture {
	t.Helper()
	ctx := context.Background()
	org := seedOrg(t, r.orgs, "Acme")
	teacher := seedTeacher(t, r.users, org.ID, "teacher")
	student := seedTeacher(t, r.users, org.ID, "student")

	cls := &domain.Class{OrganizationID: org.ID, UserID: teacher.ID, Name: "Math", TotalUsers: 30}
	require.NoError(t, r.classes.Create(ctx, cls))

	require.NoError(t, r.members.Create(ctx, &domain.ClassMember{ClassID: cls.ID, UserID: student.ID}))

	start := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	sess := &domain.ClassSession{ClassID: cls.ID, Name: "Quiz Session", StartTime: start, Type: domain.ClassSessionTypeQuiz}
	require.NoError(t, r.sessions.Create(ctx, sess))

	bank := seedBank(t, r.banks, org.ID, "Math Bank")

	quiz := &domain.Quiz{
		OrganizationID:  org.ID,
		UserID:          teacher.ID,
		ClassID:         cls.ID,
		Title:           "Midterm",
		Description:     "Mid-term exam",
		DurationMinutes: 60,
	}
	require.NoError(t, r.quizzes.Create(ctx, quiz))

	return quizFixture{org: org, teacher: teacher, student: student, class: cls, session: sess, bank: bank, quiz: quiz}
}

// --- Quiz Repository ---

func TestIntegration_QuizRepo_CRUD(t *testing.T) {
	r := setupQuizzesDB(t)
	ctx := context.Background()
	f := seedQuizFixture(t, r)

	got, err := r.quizzes.FindByID(ctx, f.quiz.ID)
	require.NoError(t, err)
	assert.Equal(t, "Midterm", got.Title)
	assert.Equal(t, 60, got.DurationMinutes)

	got.Title = "Final"
	require.NoError(t, r.quizzes.Update(ctx, got))
	updated, err := r.quizzes.FindByID(ctx, f.quiz.ID)
	require.NoError(t, err)
	assert.Equal(t, "Final", updated.Title)

	require.NoError(t, r.quizzes.Delete(ctx, f.quiz.ID))
	_, err = r.quizzes.FindByID(ctx, f.quiz.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)

	soft, err := r.quizzes.FindByIDIncludingDeleted(ctx, f.quiz.ID)
	require.NoError(t, err)
	assert.True(t, soft.DeletedAt.Valid)

	require.NoError(t, r.quizzes.HardDelete(ctx, f.quiz.ID))
	_, err = r.quizzes.FindByIDIncludingDeleted(ctx, f.quiz.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestIntegration_QuizRepo_List_OwnerScope(t *testing.T) {
	r := setupQuizzesDB(t)
	ctx := context.Background()
	f := seedQuizFixture(t, r)

	quiz2 := &domain.Quiz{
		OrganizationID:  f.org.ID,
		UserID:          f.teacher.ID,
		ClassID:         f.class.ID,
		Title:           "Quiz2",
		DurationMinutes: 30,
	}
	require.NoError(t, r.quizzes.Create(ctx, quiz2))

	bigPage := domain.ListParams{Page: 1, PageSize: 50}
	scope := domain.QuizListScope{OwnerID: &f.teacher.ID}
	got, total, err := r.quizzes.List(ctx, scope, bigPage)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, got, 2)
}

func TestIntegration_QuizRepo_List_AllOrgs(t *testing.T) {
	r := setupQuizzesDB(t)
	ctx := context.Background()
	f := seedQuizFixture(t, r)

	orgB := seedOrg(t, r.orgs, "B")
	teacherB := seedTeacher(t, r.users, orgB.ID, "tb")
	clsB := &domain.Class{OrganizationID: orgB.ID, UserID: teacherB.ID, Name: "ClassB", TotalUsers: 10}
	require.NoError(t, r.classes.Create(ctx, clsB))
	quizB := &domain.Quiz{OrganizationID: orgB.ID, UserID: teacherB.ID, ClassID: clsB.ID, Title: "QuizB", DurationMinutes: 30}
	require.NoError(t, r.quizzes.Create(ctx, quizB))

	bigPage := domain.ListParams{Page: 1, PageSize: 50}
	_, total, err := r.quizzes.List(ctx, domain.QuizListScope{All: true}, bigPage)
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)

	_ = f
}

func TestIntegration_QuizRepo_AdminList_FilterByClass(t *testing.T) {
	r := setupQuizzesDB(t)
	ctx := context.Background()
	f := seedQuizFixture(t, r)

	q := domain.AdminListQuizzesQuery{
		ClassID:    &f.class.ID,
		ListParams: domain.ListParams{Page: 1, PageSize: 50},
	}
	got, total, err := r.quizzes.AdminList(ctx, q)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, f.quiz.ID, got[0].ID)
}

// --- Rule Repository ---

func TestIntegration_QuizRuleRepo_CRUD(t *testing.T) {
	r := setupQuizzesDB(t)
	ctx := context.Background()
	f := seedQuizFixture(t, r)

	seedQuestion(t, r.questions, f.bank.ID, f.org.ID, "Q1", domain.QuestionTypeChoice)
	seedQuestion(t, r.questions, f.bank.ID, f.org.ID, "Q2", domain.QuestionTypeChoice)

	rule := &domain.QuizRule{
		QuizID:    f.quiz.ID,
		Type:      domain.QuizRuleTypeRandom,
		BankID:    &f.bank.ID,
		Count:     2,
		IsDynamic: true,
	}
	require.NoError(t, r.rules.Create(ctx, rule))

	got, err := r.rules.FindByID(ctx, rule.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.QuizRuleTypeRandom, got.Type)
	assert.Equal(t, 2, got.Count)

	got.Count = 5
	require.NoError(t, r.rules.Update(ctx, got))

	bigPage := domain.ListParams{Page: 1, PageSize: 50}
	rules, total, err := r.rules.ListByQuiz(ctx, f.quiz.ID, bigPage)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, 5, rules[0].Count)

	require.NoError(t, r.rules.Delete(ctx, rule.ID))
	_, err = r.rules.FindByID(ctx, rule.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestIntegration_QuizRuleRepo_ManualWithQuestionIDs(t *testing.T) {
	r := setupQuizzesDB(t)
	ctx := context.Background()
	f := seedQuizFixture(t, r)

	q1 := seedQuestion(t, r.questions, f.bank.ID, f.org.ID, "Q1", domain.QuestionTypeChoice)
	q2 := seedQuestion(t, r.questions, f.bank.ID, f.org.ID, "Q2", domain.QuestionTypeShortAnswer)

	rule := &domain.QuizRule{
		QuizID:      f.quiz.ID,
		Type:        domain.QuizRuleTypeManual,
		QuestionIDs: []uuid.UUID{q1.ID, q2.ID},
		Count:       2,
		IsDynamic:   false,
	}
	require.NoError(t, r.rules.Create(ctx, rule))

	got, err := r.rules.FindByID(ctx, rule.ID)
	require.NoError(t, err)
	assert.Len(t, got.QuestionIDs, 2)
	assert.Contains(t, got.QuestionIDs, q1.ID)
	assert.Contains(t, got.QuestionIDs, q2.ID)
}

// --- Room Repository ---

func TestIntegration_QuizRoomRepo_CRUD(t *testing.T) {
	r := setupQuizzesDB(t)
	ctx := context.Background()
	f := seedQuizFixture(t, r)

	room := &domain.QuizRoom{
		QuizID:         f.quiz.ID,
		ClassSessionID: f.session.ID,
	}
	require.NoError(t, r.rooms.Create(ctx, room))

	got, err := r.rooms.FindByID(ctx, room.ID)
	require.NoError(t, err)
	assert.Nil(t, got.StartedAt)
	assert.Nil(t, got.EndedAt)

	now := time.Now().UTC()
	got.StartedAt = &now
	require.NoError(t, r.rooms.Update(ctx, got))

	open, err := r.rooms.FindOpenByQuizID(ctx, f.quiz.ID)
	require.NoError(t, err)
	assert.NotNil(t, open.StartedAt)
	assert.Nil(t, open.EndedAt)
	assert.True(t, open.IsRoomOpen())

	bySession, err := r.rooms.ListBySessionID(ctx, f.session.ID)
	require.NoError(t, err)
	require.Len(t, bySession, 1)
	assert.Equal(t, room.ID, bySession[0].ID)

	bigPage := domain.ListParams{Page: 1, PageSize: 50}
	rooms, total, err := r.rooms.ListByQuiz(ctx, f.quiz.ID, bigPage)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, rooms, 1)

	require.NoError(t, r.rooms.Delete(ctx, room.ID))
	_, err = r.rooms.FindByID(ctx, room.ID)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// --- Submission Repository ---

func TestIntegration_SubmissionRepo_CRUD(t *testing.T) {
	r := setupQuizzesDB(t)
	ctx := context.Background()
	f := seedQuizFixture(t, r)

	sub := &domain.QuizSubmission{
		QuizID:    f.quiz.ID,
		UserID:    f.student.ID,
		Status:    domain.SubmissionStatusInProgress,
		StartedAt: time.Now().UTC(),
	}
	require.NoError(t, r.submissions.Create(ctx, sub))

	got, err := r.submissions.FindByID(ctx, sub.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.SubmissionStatusInProgress, got.Status)
	assert.Equal(t, f.student.ID, got.UserID)

	now := time.Now().UTC()
	got.Status = domain.SubmissionStatusSubmitted
	got.SubmittedAt = &now
	got.TotalScore = 42.5
	got.Answers = []domain.SubmissionAnswer{
		{QuestionID: uuid.New(), SelectedOptionIDs: []string{"a"}, EarnedScore: 42.5},
	}
	require.NoError(t, r.submissions.Update(ctx, got))

	updated, err := r.submissions.FindByID(ctx, sub.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.SubmissionStatusSubmitted, updated.Status)
	assert.Equal(t, 42.5, updated.TotalScore)
	assert.Len(t, updated.Answers, 1)
}

func TestIntegration_SubmissionRepo_FindByQuizAndUser(t *testing.T) {
	r := setupQuizzesDB(t)
	ctx := context.Background()
	f := seedQuizFixture(t, r)

	sub := &domain.QuizSubmission{
		QuizID:    f.quiz.ID,
		UserID:    f.student.ID,
		Status:    domain.SubmissionStatusInProgress,
		StartedAt: time.Now().UTC(),
	}
	require.NoError(t, r.submissions.Create(ctx, sub))

	got, err := r.submissions.FindByQuizAndUser(ctx, f.quiz.ID, f.student.ID)
	require.NoError(t, err)
	assert.Equal(t, sub.ID, got.ID)

	_, err = r.submissions.FindByQuizAndUser(ctx, f.quiz.ID, uuid.New())
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestIntegration_SubmissionRepo_UniqueConstraint(t *testing.T) {
	r := setupQuizzesDB(t)
	ctx := context.Background()
	f := seedQuizFixture(t, r)

	sub1 := &domain.QuizSubmission{
		QuizID:    f.quiz.ID,
		UserID:    f.student.ID,
		Status:    domain.SubmissionStatusInProgress,
		StartedAt: time.Now().UTC(),
	}
	require.NoError(t, r.submissions.Create(ctx, sub1))

	sub2 := &domain.QuizSubmission{
		QuizID:    f.quiz.ID,
		UserID:    f.student.ID,
		Status:    domain.SubmissionStatusInProgress,
		StartedAt: time.Now().UTC(),
	}
	err := r.submissions.Create(ctx, sub2)
	assert.ErrorIs(t, err, domain.ErrConflict)
}

func TestIntegration_SubmissionRepo_ListByQuiz_FilterStatus(t *testing.T) {
	r := setupQuizzesDB(t)
	ctx := context.Background()
	f := seedQuizFixture(t, r)

	student2 := seedTeacher(t, r.users, f.org.ID, "stu2")
	require.NoError(t, r.members.Create(ctx, &domain.ClassMember{ClassID: f.class.ID, UserID: student2.ID}))

	sub1 := &domain.QuizSubmission{
		QuizID:    f.quiz.ID,
		UserID:    f.student.ID,
		Status:    domain.SubmissionStatusSubmitted,
		StartedAt: time.Now().UTC(),
	}
	sub2 := &domain.QuizSubmission{
		QuizID:    f.quiz.ID,
		UserID:    student2.ID,
		Status:    domain.SubmissionStatusInProgress,
		StartedAt: time.Now().UTC(),
	}
	require.NoError(t, r.submissions.Create(ctx, sub1))
	require.NoError(t, r.submissions.Create(ctx, sub2))

	bigPage := domain.ListParams{Page: 1, PageSize: 50}

	_, total, err := r.submissions.ListByQuiz(ctx, f.quiz.ID, domain.ListSubmissionsQuery{ListParams: bigPage})
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)

	status := "submitted"
	got, total, err := r.submissions.ListByQuiz(ctx, f.quiz.ID, domain.ListSubmissionsQuery{Status: &status, ListParams: bigPage})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, domain.SubmissionStatusSubmitted, got[0].Status)

	got, total, err = r.submissions.ListByQuiz(ctx, f.quiz.ID, domain.ListSubmissionsQuery{UserID: &f.student.ID, ListParams: bigPage})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Equal(t, f.student.ID, got[0].UserID)
}

func TestIntegration_SubmissionRepo_StatusTransitions(t *testing.T) {
	r := setupQuizzesDB(t)
	ctx := context.Background()
	f := seedQuizFixture(t, r)

	sub := &domain.QuizSubmission{
		QuizID:    f.quiz.ID,
		UserID:    f.student.ID,
		Status:    domain.SubmissionStatusInProgress,
		StartedAt: time.Now().UTC(),
	}
	require.NoError(t, r.submissions.Create(ctx, sub))

	got, _ := r.submissions.FindByID(ctx, sub.ID)
	assert.Equal(t, domain.SubmissionStatusInProgress, got.Status)

	now := time.Now().UTC()
	got.Status = domain.SubmissionStatusSubmitted
	got.SubmittedAt = &now
	got.TotalScore = 85
	got.Answers = []domain.SubmissionAnswer{
		{QuestionID: uuid.New(), Value: "answer", EarnedScore: 85},
	}
	require.NoError(t, r.submissions.Update(ctx, got))

	got, _ = r.submissions.FindByID(ctx, sub.ID)
	assert.Equal(t, domain.SubmissionStatusSubmitted, got.Status)

	got.Status = domain.SubmissionStatusGraded
	got.TotalScore = 90
	got.Answers[0].EarnedScore = 90
	require.NoError(t, r.submissions.Update(ctx, got))

	got, _ = r.submissions.FindByID(ctx, sub.ID)
	assert.Equal(t, domain.SubmissionStatusGraded, got.Status)
	assert.Equal(t, float64(90), got.TotalScore)
}
