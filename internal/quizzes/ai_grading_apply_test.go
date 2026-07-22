package quizzes

import (
	"testing"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func TestApplyAIScoreSuggestModeNeverTouchesEarned(t *testing.T) {
	qid := uuid.New()
	ans := &domain.SubmissionAnswer{QuestionID: qid, EarnedScore: 0}
	applyAIScore(ans, aiScore{Score: 4, Rationale: "خوب"}, domain.AIGradingModeSuggest, false)

	if ans.EarnedScore != 0 {
		t.Fatal("suggest mode must not change earned_score")
	}
	if ans.SuggestedScore == nil || *ans.SuggestedScore != 4 || ans.AIRationale != "خوب" {
		t.Fatal("suggestion + rationale must be stored")
	}
	if ans.AIStatus != domain.AIAnswerStatusScored || ans.GradedBy != "" {
		t.Fatalf("graded_by must stay unset in suggest mode, got %q", ans.GradedBy)
	}
}

func TestApplyAIScoreApplyModeSetsEarnedAndGradedBy(t *testing.T) {
	qid := uuid.New()
	ans := &domain.SubmissionAnswer{QuestionID: qid}
	applyAIScore(ans, aiScore{Score: 3.5, Rationale: "r"}, domain.AIGradingModeApply, false)

	if ans.EarnedScore != 3.5 || ans.GradedBy != domain.GradedByAI {
		t.Fatalf("apply mode must set earned_score + graded_by=ai, got %v %q", ans.EarnedScore, ans.GradedBy)
	}
	if ans.SuggestedScore == nil || *ans.SuggestedScore != 3.5 {
		t.Fatal("suggestion stored even in apply mode")
	}
}

func TestApplyAIScoreNeverOverwritesManual(t *testing.T) {
	qid := uuid.New()
	ans := &domain.SubmissionAnswer{QuestionID: qid, EarnedScore: 5, GradedBy: domain.GradedByManual}
	applyAIScore(ans, aiScore{Score: 1, Rationale: "r"}, domain.AIGradingModeApply, true)

	if ans.EarnedScore != 5 || ans.GradedBy != domain.GradedByManual {
		t.Fatal("manual grade must be sacred — never overwritten by AI")
	}
	// Suggestion is still recorded so a teacher can compare.
	if ans.SuggestedScore == nil || *ans.SuggestedScore != 1 {
		t.Fatal("suggestion should still be recorded against a manual answer")
	}
}

func TestShouldGradeSkipsAlreadyAIUnlessForce(t *testing.T) {
	qid := uuid.New()
	aiGraded := domain.SubmissionAnswer{QuestionID: qid, GradedBy: domain.GradedByAI}
	if shouldGrade(aiGraded, false) {
		t.Fatal("already AI-graded should be skipped without force")
	}
	if !shouldGrade(aiGraded, true) {
		t.Fatal("force should re-grade AI-graded answers")
	}
	manual := domain.SubmissionAnswer{QuestionID: qid, GradedBy: domain.GradedByManual}
	if shouldGrade(manual, true) {
		t.Fatal("manual must never be re-graded, even with force")
	}
	fresh := domain.SubmissionAnswer{QuestionID: qid}
	if !shouldGrade(fresh, false) {
		t.Fatal("ungraded answer should be graded")
	}
}
