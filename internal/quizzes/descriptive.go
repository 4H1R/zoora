package quizzes

import (
	"math"
	"strings"
	"unicode/utf8"

	"github.com/4H1R/zoora/internal/domain"
)

// suffixTolerance is the max number of extra trailing runes a student token
// may have beyond a concept token and still match ("فتوسنتزی" for "فتوسنتز",
// "books" for "book"). Applied only to concept tokens of at least
// minTokenLenForTolerance runes so short words stay exact ("in" ≠ "into").
const (
	suffixTolerance         = 3
	minTokenLenForTolerance = 3
)

// suggestDescriptive computes the advisory grading signals for a descriptive
// answer: the suggested score (sum of matched rubric concept weights, nil when
// the question has no rubric), the canonical values of matched concepts, and
// the char-trigram cosine similarity to the model answer in percent (nil when
// there is no model answer or no student text). Never used as a real grade.
func suggestDescriptive(q *domain.Question, value string) (*float64, []string, *float64) {
	answerTokens := strings.Fields(normalizeText(value))

	var suggested *float64
	var matched []string
	hasRubric := false
	var sum float64
	for _, o := range q.Options {
		if o.Score <= 0 || normalizeText(o.Value) == "" {
			continue
		}
		hasRubric = true
		for _, candidate := range append([]string{o.Value}, o.Synonyms...) {
			if containsPhrase(answerTokens, strings.Fields(normalizeText(candidate))) {
				sum += o.Score
				matched = append(matched, o.Value)
				break
			}
		}
	}
	if hasRubric {
		suggested = &sum
	}

	var similarity *float64
	if strings.TrimSpace(q.ModelAnswer) != "" && len(answerTokens) > 0 {
		pct := trigramCosine(normalizeText(value), normalizeText(q.ModelAnswer)) * 100
		pct = math.Round(pct*10) / 10
		similarity = &pct
	}
	return suggested, matched, similarity
}

// containsPhrase reports whether the concept tokens appear consecutively in
// the answer tokens, each token matching exactly or via suffix tolerance.
func containsPhrase(answer, concept []string) bool {
	if len(concept) == 0 || len(concept) > len(answer) {
		return false
	}
	for i := 0; i+len(concept) <= len(answer); i++ {
		ok := true
		for j := range concept {
			if !tokenMatches(answer[i+j], concept[j]) {
				ok = false
				break
			}
		}
		if ok {
			return true
		}
	}
	return false
}

func tokenMatches(answerTok, conceptTok string) bool {
	if answerTok == conceptTok {
		return true
	}
	if utf8.RuneCountInString(conceptTok) < minTokenLenForTolerance {
		return false
	}
	return strings.HasPrefix(answerTok, conceptTok) &&
		utf8.RuneCountInString(answerTok)-utf8.RuneCountInString(conceptTok) <= suffixTolerance
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

// stripSuggestions removes the advisory grading signals from every answer.
// Called on student-facing reads so the rubric-driven suggestion never leaks
// to the person being graded.
func stripSuggestions(sub *domain.QuizSubmission) {
	for i := range sub.Answers {
		sub.Answers[i].SuggestedScore = nil
		sub.Answers[i].MatchedConcepts = nil
		sub.Answers[i].SimilarityPct = nil
	}
}
