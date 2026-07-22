package quizzes

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

func descQuestion(id uuid.UUID, text, model string, max float64) domain.Question {
	return domain.Question{
		ID:          id,
		Type:        domain.QuestionTypeDescriptive,
		Text:        text,
		ModelAnswer: model,
		Options:     []domain.QuestionOption{{ID: "o1", Value: "", Score: max}},
	}
}

func TestBuildGradingPromptSeparatesSystemAndUser(t *testing.T) {
	qid := uuid.New()
	items := []gradeItem{{Question: descQuestion(qid, "تعریف کن", "پاسخ درست", 5), Answer: "پاسخ دانش‌آموز"}}

	sys, user := buildGradingPrompt(items)
	if !strings.Contains(sys, "JSON") || !strings.Contains(sys, "question_id") {
		t.Fatal("system prompt must define the JSON contract")
	}
	// Untrusted student content must be in the user channel, not system.
	if strings.Contains(sys, "پاسخ دانش‌آموز") {
		t.Fatal("student answer leaked into system prompt")
	}
	if !strings.Contains(user, "پاسخ دانش‌آموز") || !strings.Contains(user, qid.String()) {
		t.Fatal("user prompt must contain the delimited student answer and question id")
	}
}

func TestBuildGradingPromptTruncatesLongAnswer(t *testing.T) {
	qid := uuid.New()
	long := strings.Repeat("x", 5000)
	items := []gradeItem{{Question: descQuestion(qid, "q", "m", 5), Answer: long}}
	_, user := buildGradingPrompt(items)
	if strings.Count(user, "x") > maxAnswerChars {
		t.Fatalf("answer not truncated to %d chars", maxAnswerChars)
	}
	if !strings.Contains(user, "[truncated]") {
		t.Fatal("truncation marker missing")
	}
}

func TestParseAndValidateHappyPath(t *testing.T) {
	qid := uuid.New()
	items := []gradeItem{{Question: descQuestion(qid, "q", "m", 5), Answer: "a"}}
	raw := `{"scores":[{"question_id":"` + qid.String() + `","score":4,"rationale":"خوب"}]}`

	scored, missing, err := parseAndValidate(raw, items)
	if err != nil {
		t.Fatalf("parseAndValidate: %v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("expected no missing, got %d", len(missing))
	}
	if scored[qid].Score != 4 || scored[qid].Rationale != "خوب" {
		t.Fatalf("bad score: %+v", scored[qid])
	}
}

func TestParseAndValidateClampsOutOfRange(t *testing.T) {
	qid := uuid.New()
	items := []gradeItem{{Question: descQuestion(qid, "q", "m", 5), Answer: "a"}}
	raw := `{"scores":[{"question_id":"` + qid.String() + `","score":99,"rationale":"x"}]}`
	scored, _, err := parseAndValidate(raw, items)
	if err != nil {
		t.Fatalf("parseAndValidate: %v", err)
	}
	if scored[qid].Score != 5 {
		t.Fatalf("score should clamp to max 5, got %v", scored[qid].Score)
	}
}

func TestParseAndValidateReportsMissing(t *testing.T) {
	q1, q2 := uuid.New(), uuid.New()
	items := []gradeItem{
		{Question: descQuestion(q1, "q1", "m", 5), Answer: "a"},
		{Question: descQuestion(q2, "q2", "m", 5), Answer: "b"},
	}
	raw := `{"scores":[{"question_id":"` + q1.String() + `","score":3,"rationale":"x"}]}`
	scored, missing, err := parseAndValidate(raw, items)
	if err != nil {
		t.Fatalf("parseAndValidate: %v", err)
	}
	if len(scored) != 1 || len(missing) != 1 || missing[0] != q2 {
		t.Fatalf("expected q2 missing, got scored=%v missing=%v", scored, missing)
	}
}

func TestParseAndValidateGarbageErrors(t *testing.T) {
	qid := uuid.New()
	items := []gradeItem{{Question: descQuestion(qid, "q", "m", 5), Answer: "a"}}
	if _, _, err := parseAndValidate("totally not json", items); err == nil {
		t.Fatal("expected error on unparseable output")
	}
}
