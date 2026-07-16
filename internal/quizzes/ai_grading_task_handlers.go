package quizzes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hibiken/asynq"

	"github.com/4H1R/zoora/internal/domain"
)

// NewAIGradeSubmissionHandler returns the Asynq handler for grading one
// submission's descriptive answers. Registered in cmd/worker/main.go. It takes
// the concrete *service (built via NewAIGradingWorker) because it reaches the
// unexported gradeSubmissionAI worker path.
func NewAIGradeSubmissionHandler(svc *service) asynq.HandlerFunc {
	return func(ctx context.Context, t *asynq.Task) error {
		var p domain.QuizAIGradeSubmissionPayload
		if err := json.Unmarshal(t.Payload(), &p); err != nil {
			// Malformed payload will never succeed — do not retry.
			return fmt.Errorf("unmarshal ai-grade payload: %w: %w", err, asynq.SkipRetry)
		}
		if err := svc.gradeSubmissionAI(ctx, p); err != nil {
			return fmt.Errorf("ai grade submission %s: %w", p.SubmissionID, err)
		}
		return nil
	}
}
