package quizzes

import "github.com/4H1R/zoora/internal/domain"

// shouldGrade decides whether the AI should score this answer.
// Manual grades are sacred (never). AI grades are re-graded only with force.
// Unset/empty answers are always eligible.
func shouldGrade(a domain.SubmissionAnswer, force bool) bool {
	switch a.GradedBy {
	case domain.GradedByManual:
		return false
	case domain.GradedByAI:
		return force
	default:
		return true
	}
}

// applyAIScore writes the AI result onto an answer per mode & precedence rules.
// The suggestion + rationale are ALWAYS recorded (for audit/compare). In Apply
// mode the earned score is set and graded_by flipped to ai — but a manual grade
// is never overwritten.
func applyAIScore(a *domain.SubmissionAnswer, s aiScore, mode domain.AIGradingMode, force bool) {
	score := s.Score
	a.SuggestedScore = &score
	a.AIRationale = s.Rationale
	a.AIStatus = domain.AIAnswerStatusScored

	if mode != domain.AIGradingModeApply {
		return
	}
	if a.GradedBy == domain.GradedByManual {
		return // sacred
	}
	a.EarnedScore = score
	a.GradedBy = domain.GradedByAI
}
