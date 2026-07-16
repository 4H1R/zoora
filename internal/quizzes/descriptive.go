package quizzes

import (
	"math"
	"strings"

	"github.com/4H1R/zoora/internal/domain"
)

// computeSimilarity returns the char-trigram cosine similarity between a
// descriptive answer and the question's model answer, in percent. Nil when the
// question has no model answer or the student wrote nothing. Advisory display
// only — shown to the grader as a hint, never used as a real grade.
func computeSimilarity(q *domain.Question, value string) *float64 {
	answer := normalizeText(value)
	if strings.TrimSpace(q.ModelAnswer) == "" || strings.TrimSpace(answer) == "" {
		return nil
	}
	pct := trigramCosine(answer, normalizeText(q.ModelAnswer)) * 100
	pct = math.Round(pct*10) / 10
	return &pct
}

// trigramCosine is the cosine similarity between the character-trigram
// frequency vectors of two strings. Language-agnostic and robust to FA
// morphology/spacing, which is why it is used instead of token overlap.
func trigramCosine(a, b string) float64 {
	va, vb := trigramFreq(a), trigramFreq(b)
	if len(va) == 0 || len(vb) == 0 {
		return 0
	}
	var dot, na, nb float64
	for g, ca := range va {
		na += float64(ca * ca)
		if cb, ok := vb[g]; ok {
			dot += float64(ca * cb)
		}
	}
	for _, cb := range vb {
		nb += float64(cb * cb)
	}
	if dot == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

func trigramFreq(s string) map[string]int {
	runes := []rune(s)
	freq := make(map[string]int)
	if len(runes) == 0 {
		return freq
	}
	if len(runes) < 3 {
		freq[string(runes)]++
		return freq
	}
	for i := 0; i+3 <= len(runes); i++ {
		freq[string(runes[i:i+3])]++
	}
	return freq
}

// stripSimilarity removes the advisory similarity signal from every answer.
// Called on student-facing reads so the grading hint never leaks to the person
// being graded.
func stripSimilarity(sub *domain.QuizSubmission) {
	for i := range sub.Answers {
		sub.Answers[i].SimilarityPct = nil
		sub.Answers[i].SuggestedScore = nil
		sub.Answers[i].AIRationale = ""
		sub.Answers[i].AIStatus = ""
		// GradedBy is safe to expose (it's just "graded by teacher/ai"); keep it.
	}
}
