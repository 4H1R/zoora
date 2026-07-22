//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/4H1R/zoora/internal/ai"
	"github.com/4H1R/zoora/internal/audit"
	"github.com/4H1R/zoora/internal/classes"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/organizations"
	"github.com/4H1R/zoora/internal/platform/database"
	"github.com/4H1R/zoora/internal/platform/llm"
	"github.com/4H1R/zoora/internal/platform/queue"
	"github.com/4H1R/zoora/internal/questionbanks"
	"github.com/4H1R/zoora/internal/quizzes"
	"github.com/4H1R/zoora/internal/users"
	"github.com/4H1R/zoora/tests/testutil"
)

// scriptedLLM is a stub domain.LLM returning a fixed JSON body for every call,
// standing in for a real provider so the DB path can be exercised deterministically.
type scriptedLLM struct{ text string }

func (s scriptedLLM) Generate(context.Context, domain.LLMRequest) (domain.LLMResponse, error) {
	return domain.LLMResponse{Text: s.text, Model: "gemini-2.0-flash"}, nil
}

func setupAIGradingDB(t *testing.T) (*gorm.DB, quizRepos) {
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
		&domain.AIGradingJob{},
		&domain.AIUsageEvent{},
	))
	return db, quizRepos{
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

// TestAIGradingEndToEnd exercises the real DB path with a stub LLM: a Pro-plan
// teacher triggers apply-mode AI grading, the worker grading task runs, and the
// durable job, the per-answer AI fields, and the metering row are all persisted.
func TestAIGradingEndToEnd(t *testing.T) {
	db, r := setupAIGradingDB(t)
	ctx := context.Background()

	// Arrange: org + class + quiz (owned by the teacher) via the shared fixture.
	f := seedQuizFixture(t, r)

	// A descriptive question worth 5 points (max score = its single scored option).
	question := &domain.Question{
		BankID:         f.bank.ID,
		OrganizationID: f.org.ID,
		Text:           "مفهوم را توضیح دهید",
		ModelAnswer:    "پاسخ مدل",
		Type:           domain.QuestionTypeDescriptive,
		Options:        []domain.QuestionOption{{ID: "o1", Value: "", Score: 5}},
	}
	require.NoError(t, r.questions.Create(ctx, question))

	// A submitted submission with one ungraded descriptive answer.
	sub := &domain.QuizSubmission{
		QuizID:    f.quiz.ID,
		UserID:    f.student.ID,
		Status:    domain.SubmissionStatusSubmitted,
		StartedAt: time.Now().UTC(),
		Answers: []domain.SubmissionAnswer{
			{QuestionID: question.ID, Value: "پاسخ دانش‌آموز"},
		},
	}
	require.NoError(t, r.submissions.Create(ctx, sub))

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// AI infra: job + usage repos; the scripted LLM is wrapped in the metering
	// decorator so a successful grade also writes an ai_usage_events row.
	jobRepo := ai.NewJobRepository(db)
	usageRepo := ai.NewUsageRepository(db)
	batch := `{"scores":[{"question_id":"` + question.ID.String() + `","score":4,"rationale":"خوب"}]}`
	llmClient := llm.NewMetered(scriptedLLM{text: batch}, usageRepo, "gemini")

	// Real queue client (StartAIGrading fans out one task per submission).
	rdb := testutil.SetupRedis(t)
	qc, err := queue.NewClient("redis://"+rdb.Options().Addr, logger)
	require.NoError(t, err)
	t.Cleanup(func() { _ = qc.Close() })

	transactor := database.NewTransactor(db)
	auditSvc := audit.NewService(audit.NewRepository(db), logger)
	svc := quizzes.NewAIGradingWorker(
		r.quizzes, r.rules, r.rooms, r.submissions, r.questions,
		r.classes, r.members, qc, llmClient, jobRepo, transactor, auditSvc, logger,
	)

	// Act 1: teacher on the Pro plan (grants FeatureAI) triggers an apply-mode run.
	ent := domain.PlanCatalog[domain.PlanKey(domain.TierPro, 50)]
	caller := domain.Caller{UserID: f.teacher.ID, OrgID: &f.org.ID, Ent: ent}
	callerCtx := domain.WithCaller(ctx, caller)

	job, err := svc.StartAIGrading(callerCtx, f.quiz.ID, domain.StartAIGradingDTO{Mode: domain.AIGradingModeApply})
	require.NoError(t, err)
	require.NotNil(t, job)
	assert.Equal(t, 1, job.Total, "one eligible descriptive submission")

	// Act 2: run the worker grading path for the submission (as the Asynq handler would).
	handler := quizzes.NewAIGradeSubmissionHandler(svc)
	payload, err := json.Marshal(domain.QuizAIGradeSubmissionPayload{
		JobID:          job.ID,
		SubmissionID:   sub.ID,
		OrganizationID: f.org.ID,
		Mode:           domain.AIGradingModeApply,
	})
	require.NoError(t, err)
	require.NoError(t, handler(ctx, asynq.NewTask(domain.TypeQuizAIGradeSubmission, payload)))

	// Assert: the job advanced to completed.
	gotJob, err := jobRepo.FindByID(ctx, job.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, gotJob.Done)
	assert.Equal(t, domain.AIGradingStatusCompleted, gotJob.Status)

	// Assert: the descriptive answer was AI-scored and applied.
	gotSub, err := r.submissions.FindByID(ctx, sub.ID)
	require.NoError(t, err)
	require.Len(t, gotSub.Answers, 1)
	ans := gotSub.Answers[0]
	assert.Equal(t, float64(4), ans.EarnedScore)
	assert.Equal(t, domain.GradedByAI, ans.GradedBy)
	assert.Equal(t, domain.AIAnswerStatusScored, ans.AIStatus)
	require.NotNil(t, ans.SuggestedScore)
	assert.Equal(t, float64(4), *ans.SuggestedScore)

	// Assert: a metering row was recorded for the org.
	var usageCount int64
	require.NoError(t, db.Model(&domain.AIUsageEvent{}).
		Where("organization_id = ?", f.org.ID).Count(&usageCount).Error)
	assert.GreaterOrEqual(t, usageCount, int64(1))
}
