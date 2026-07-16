package quizzes

import (
	"testing"

	"github.com/4H1R/zoora/internal/domain"
)

func fptr(f float64) *float64 { return &f }

func TestComputeSimilarity(t *testing.T) {
	q := &domain.Question{
		Type:        domain.QuestionTypeDescriptive,
		ModelAnswer: "گیاهان با استفاده از نور خورشید غذا می‌سازند",
		Options:     []domain.QuestionOption{{ID: "a", Score: 4}},
	}
	sim := computeSimilarity(q, "گیاهان با استفاده از نور خورشید غذا می‌سازند")
	if sim == nil || *sim != 100 {
		t.Fatalf("identical answer similarity = %v, want 100", sim)
	}
	sim = computeSimilarity(q, "qwerty zzz")
	if sim == nil || *sim != 0 {
		t.Fatalf("disjoint answer similarity = %v, want 0", sim)
	}
	sim = computeSimilarity(q, "گیاهان با نور خورشید غذا درست میکنند")
	if sim == nil || *sim <= 0 || *sim >= 100 {
		t.Fatalf("paraphrase similarity = %v, want in (0,100)", sim)
	}
	// no model answer -> no similarity
	qNo := &domain.Question{Type: domain.QuestionTypeDescriptive, Options: q.Options}
	if sim = computeSimilarity(qNo, "گیاهان"); sim != nil {
		t.Fatalf("similarity without model answer = %v, want nil", sim)
	}
	// empty student answer -> no similarity
	if sim = computeSimilarity(q, ""); sim != nil {
		t.Fatalf("similarity for empty answer = %v, want nil", sim)
	}
}

func TestStripSimilarity(t *testing.T) {
	sub := &domain.QuizSubmission{
		Answers: []domain.SubmissionAnswer{
			{SimilarityPct: fptr(50), EarnedScore: 1},
			{EarnedScore: 3},
		},
	}
	stripSimilarity(sub)
	for i, a := range sub.Answers {
		if a.SimilarityPct != nil {
			t.Fatalf("answer %d still has similarity: %+v", i, a)
		}
	}
	if sub.Answers[0].EarnedScore != 1 || sub.Answers[1].EarnedScore != 3 {
		t.Fatal("earned scores must be untouched")
	}
}
