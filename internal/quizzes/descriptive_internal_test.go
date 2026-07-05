package quizzes

import (
	"testing"

	"github.com/4H1R/zoora/internal/domain"
)

func fptr(f float64) *float64 { return &f }

func TestSuggestDescriptiveConceptMatching(t *testing.T) {
	q := &domain.Question{
		Type: domain.QuestionTypeDescriptive,
		Options: []domain.QuestionOption{
			{ID: "c1", Value: "فتوسنتز", Score: 2, Synonyms: []string{"photosynthesis"}},
			{ID: "c2", Value: "چرخه آب", Score: 1.5},
			{ID: "c3", Value: "energy", Score: 1},
		},
	}
	cases := []struct {
		name        string
		answer      string
		wantScore   *float64
		wantMatched []string
	}{
		{
			"all concepts matched",
			"فتوسنتز و چرخه آب باعث تولید energy می‌شوند",
			fptr(4.5),
			[]string{"فتوسنتز", "چرخه آب", "energy"},
		},
		{
			"partial match",
			"در این فرایند فتوسنتز رخ می‌دهد",
			fptr(2),
			[]string{"فتوسنتز"},
		},
		{
			"synonym counts as canonical concept",
			"the process of photosynthesis",
			fptr(2),
			[]string{"فتوسنتز"},
		},
		{
			"suffix tolerance matches suffixed form",
			"فرایند فتوسنتزی گیاهان",
			fptr(2),
			[]string{"فتوسنتز"},
		},
		{
			"phrase must be consecutive",
			"چرخه بزرگ آب",
			fptr(0),
			nil,
		},
		{
			"arabic characters fold",
			"فتوسنتز با كلروفيل", // Arabic kaf/yeh in unrelated word, concept still matches
			fptr(2),
			[]string{"فتوسنتز"},
		},
		{
			"no match",
			"پاسخ نامربوط",
			fptr(0),
			nil,
		},
		{
			"empty answer",
			"",
			fptr(0),
			nil,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			score, matched, _ := suggestDescriptive(q, c.answer)
			if (score == nil) != (c.wantScore == nil) || (score != nil && *score != *c.wantScore) {
				t.Fatalf("score = %v, want %v", score, c.wantScore)
			}
			if len(matched) != len(c.wantMatched) {
				t.Fatalf("matched = %v, want %v", matched, c.wantMatched)
			}
			for i := range matched {
				if matched[i] != c.wantMatched[i] {
					t.Fatalf("matched = %v, want %v", matched, c.wantMatched)
				}
			}
		})
	}
}

func TestSuggestDescriptiveSuffixToleranceLimits(t *testing.T) {
	q := &domain.Question{
		Type: domain.QuestionTypeDescriptive,
		Options: []domain.QuestionOption{
			{ID: "c1", Value: "کار", Score: 1},
			{ID: "c2", Value: "in", Score: 1},
		},
	}
	// "کارخانه" adds 4 runes to "کار" — beyond tolerance; "into" adds 2 to "in"
	// but tolerance requires concept token >= 3 runes.
	score, matched, _ := suggestDescriptive(q, "کارخانه into")
	if *score != 0 || len(matched) != 0 {
		t.Fatalf("score = %v matched = %v, want 0 and none", *score, matched)
	}
	// short suffix within tolerance still matches
	score, matched, _ = suggestDescriptive(q, "کاری سخت")
	if *score != 1 || len(matched) != 1 {
		t.Fatalf("score = %v matched = %v, want 1 and [کار]", *score, matched)
	}
}

func TestSuggestDescriptiveLegacyScoreHolderOptions(t *testing.T) {
	// Legacy descriptive questions store a single option with empty Value as
	// the score holder — no rubric, so no suggested score.
	q := &domain.Question{
		Type:    domain.QuestionTypeDescriptive,
		Options: []domain.QuestionOption{{ID: "a", Score: 4}},
	}
	score, matched, sim := suggestDescriptive(q, "هر پاسخی")
	if score != nil || matched != nil || sim != nil {
		t.Fatalf("want all nil for legacy question, got %v %v %v", score, matched, sim)
	}
}

func TestSuggestDescriptiveSimilarity(t *testing.T) {
	q := &domain.Question{
		Type:        domain.QuestionTypeDescriptive,
		ModelAnswer: "گیاهان با استفاده از نور خورشید غذا می‌سازند",
		Options:     []domain.QuestionOption{{ID: "a", Score: 4}},
	}
	_, _, sim := suggestDescriptive(q, "گیاهان با استفاده از نور خورشید غذا می‌سازند")
	if sim == nil || *sim != 100 {
		t.Fatalf("identical answer similarity = %v, want 100", sim)
	}
	_, _, sim = suggestDescriptive(q, "qwerty zzz")
	if sim == nil || *sim != 0 {
		t.Fatalf("disjoint answer similarity = %v, want 0", sim)
	}
	_, _, sim = suggestDescriptive(q, "گیاهان با نور خورشید غذا درست میکنند")
	if sim == nil || *sim <= 0 || *sim >= 100 {
		t.Fatalf("paraphrase similarity = %v, want in (0,100)", sim)
	}
	// no model answer -> no similarity
	qNo := &domain.Question{Type: domain.QuestionTypeDescriptive, Options: q.Options}
	_, _, sim = suggestDescriptive(qNo, "گیاهان")
	if sim != nil {
		t.Fatalf("similarity without model answer = %v, want nil", sim)
	}
	// empty student answer -> no similarity
	_, _, sim = suggestDescriptive(q, "")
	if sim != nil {
		t.Fatalf("similarity for empty answer = %v, want nil", sim)
	}
}

func TestStripSuggestions(t *testing.T) {
	sub := &domain.QuizSubmission{
		Answers: []domain.SubmissionAnswer{
			{SuggestedScore: fptr(2), MatchedConcepts: []string{"x"}, SimilarityPct: fptr(50), EarnedScore: 1},
			{EarnedScore: 3},
		},
	}
	stripSuggestions(sub)
	for i, a := range sub.Answers {
		if a.SuggestedScore != nil || a.MatchedConcepts != nil || a.SimilarityPct != nil {
			t.Fatalf("answer %d still has suggestion fields: %+v", i, a)
		}
	}
	if sub.Answers[0].EarnedScore != 1 || sub.Answers[1].EarnedScore != 3 {
		t.Fatal("earned scores must be untouched")
	}
}
