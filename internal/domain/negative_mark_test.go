package domain

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestNegativeMarkMode_Valid(t *testing.T) {
	valid := []NegativeMarkMode{NegativeMarkNone, NegativeMarkPerWrong, NegativeMarkAccumulative}
	for _, m := range valid {
		if !m.Valid() {
			t.Fatalf("expected %q valid", m)
		}
	}
	if NegativeMarkMode("bogus").Valid() {
		t.Fatal("expected bogus invalid")
	}
}

func TestNegativeMarkConfig_Penalty(t *testing.T) {
	cases := []struct {
		name  string
		cfg   NegativeMarkConfig
		wrong int
		want  float64
	}{
		{"none", NegativeMarkConfig{Mode: NegativeMarkNone}, 3, 0},
		{"per_wrong 2x0.5", NegativeMarkConfig{Mode: NegativeMarkPerWrong, NegativeValue: 0.5}, 2, 1.0},
		{"per_wrong zero wrong", NegativeMarkConfig{Mode: NegativeMarkPerWrong, NegativeValue: 0.5}, 0, 0},
		{"accumulative 3/3", NegativeMarkConfig{Mode: NegativeMarkAccumulative, WrongsPerPoint: 3}, 3, 1.0},
		{"accumulative 5/3 floor", NegativeMarkConfig{Mode: NegativeMarkAccumulative, WrongsPerPoint: 3}, 5, 1.0},
		{"accumulative 6/3", NegativeMarkConfig{Mode: NegativeMarkAccumulative, WrongsPerPoint: 3}, 6, 2.0},
		{"accumulative wpp 0 safe", NegativeMarkConfig{Mode: NegativeMarkAccumulative, WrongsPerPoint: 0}, 5, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.cfg.Penalty(c.wrong); got != c.want {
				t.Fatalf("Penalty(%d) = %v, want %v", c.wrong, got, c.want)
			}
		})
	}
}

func TestFractionFor(t *testing.T) {
	cases := map[int]float64{2: 0.5, 3: 0.33, 4: 0.25, 5: 0.2}
	for n, want := range cases {
		if got := FractionFor(n); got != want {
			t.Fatalf("FractionFor(%d) = %v, want %v", n, got, want)
		}
	}
	if got := FractionFor(8); got != 1.0/8.0 {
		t.Fatalf("FractionFor(8) = %v, want %v", got, 1.0/8.0)
	}
	if got := FractionFor(0); got != 0 {
		t.Fatalf("FractionFor(0) = %v, want 0", got)
	}
}

func TestValidateNegativeMark(t *testing.T) {
	ok := []struct {
		mode NegativeMarkMode
		val  float64
		wpp  int
	}{
		{NegativeMarkNone, 0, 0},
		{NegativeMarkPerWrong, 0.5, 0},
		{NegativeMarkAccumulative, 0, 3},
		{NegativeMarkAccumulative, 0, 2},
		{NegativeMarkAccumulative, 0, 5},
	}
	for _, c := range ok {
		if err := ValidateNegativeMark(c.mode, c.val, c.wpp); err != nil {
			t.Fatalf("expected valid (%v,%v,%v), got %v", c.mode, c.val, c.wpp, err)
		}
	}
	bad := []struct {
		mode NegativeMarkMode
		val  float64
		wpp  int
	}{
		{"bogus", 0, 0},
		{NegativeMarkPerWrong, 0, 0},
		{NegativeMarkPerWrong, -1, 0},
		{NegativeMarkAccumulative, 0, 1},
		{NegativeMarkAccumulative, 0, 6},
	}
	for _, c := range bad {
		err := ValidateNegativeMark(c.mode, c.val, c.wpp)
		if err == nil {
			t.Fatalf("expected invalid (%v,%v,%v)", c.mode, c.val, c.wpp)
		}
		if !errors.Is(err, ErrValidation) {
			t.Fatalf("expected ErrValidation, got %v", err)
		}
	}
}

func TestNormalizeNegativeMark(t *testing.T) {
	m, v, w := NormalizeNegativeMark(NegativeMarkAccumulative, 9, 3)
	if m != NegativeMarkAccumulative || v != 0 || w != 3 {
		t.Fatalf("accumulative normalize = (%v,%v,%v)", m, v, w)
	}
	m, v, w = NormalizeNegativeMark(NegativeMarkPerWrong, 0.5, 9)
	if m != NegativeMarkPerWrong || v != 0.5 || w != 0 {
		t.Fatalf("per_wrong normalize = (%v,%v,%v)", m, v, w)
	}
	m, v, w = NormalizeNegativeMark("bogus", 5, 5)
	if m != NegativeMarkNone || v != 0 || w != 0 {
		t.Fatalf("none normalize = (%v,%v,%v)", m, v, w)
	}
}

func TestQuestionOptionPhotosCollection(t *testing.T) {
	if QuestionOptionPhotosCollection != "option-photos" {
		t.Fatalf("got %q", QuestionOptionPhotosCollection)
	}
}

func TestValidateQuestionOptions_OptionImageOnlyOnChoice(t *testing.T) {
	id := uuid.New()
	choice := []QuestionOption{
		{ID: "a", Value: "x", Score: 1, ImageMediaID: &id},
		{ID: "b", Value: "y", Score: 0},
	}
	if err := ValidateQuestionOptions(QuestionTypeChoice, choice); err != nil {
		t.Fatalf("choice with image should be valid: %v", err)
	}
	sa := []QuestionOption{{ID: "a", Value: "x", Score: 1, ImageMediaID: &id}}
	if err := ValidateQuestionOptions(QuestionTypeShortAnswer, sa); err == nil {
		t.Fatal("expected error: image on non-choice option")
	}
}

func TestQuizRule_NegativeOverridesShape(t *testing.T) {
	qid := uuid.New()
	r := QuizRule{NegativeOverrides: []QuizQuestionNegativeOverride{
		{QuestionID: qid, Mode: NegativeMarkPerWrong, NegativeValue: 0.5},
	}}
	if len(r.NegativeOverrides) != 1 || r.NegativeOverrides[0].Mode != NegativeMarkPerWrong {
		t.Fatal("override not stored")
	}
}

func TestResolveNegativeMark_Priority(t *testing.T) {
	qid := uuid.New()
	q := Question{NegativeMarkMode: NegativeMarkAccumulative, WrongsPerPoint: 3}
	override := &QuizQuestionNegativeOverride{QuestionID: qid, Mode: NegativeMarkPerWrong, NegativeValue: 0.5}
	quizWide := NegativeMarkConfig{Mode: NegativeMarkAccumulative, WrongsPerPoint: 4}

	got := ResolveNegativeMark(q, override, nil, quizWide)
	if got.Mode != NegativeMarkPerWrong || got.NegativeValue != 0.5 {
		t.Fatalf("override should win, got %+v", got)
	}
	if got.Fraction != 0.5 {
		t.Fatalf("per_wrong fraction = value, got %v", got.Fraction)
	}

	got = ResolveNegativeMark(q, nil, nil, quizWide)
	if got.Mode != NegativeMarkAccumulative || got.WrongsPerPoint != 3 {
		t.Fatalf("question default should win, got %+v", got)
	}
	if got.Fraction != FractionFor(3) {
		t.Fatalf("accumulative fraction = 1/wpp table, got %v", got.Fraction)
	}

	none := Question{NegativeMarkMode: NegativeMarkNone}
	got = ResolveNegativeMark(none, nil, nil, quizWide)
	if got.Mode != NegativeMarkAccumulative || got.WrongsPerPoint != 4 {
		t.Fatalf("quiz-wide should fill gap, got %+v", got)
	}

	noneOverride := &QuizQuestionNegativeOverride{QuestionID: qid, Mode: NegativeMarkNone}
	got = ResolveNegativeMark(q, noneOverride, nil, quizWide)
	if got.Mode != NegativeMarkAccumulative || got.WrongsPerPoint != 3 {
		t.Fatalf("none override ignored => question default, got %+v", got)
	}

	got = ResolveNegativeMark(none, nil, nil, NegativeMarkConfig{Mode: NegativeMarkNone})
	if got.Mode != NegativeMarkNone {
		t.Fatalf("expected none, got %+v", got)
	}
}

func TestResolveNegativeMark_RuleDefault(t *testing.T) {
	qid := uuid.New()
	// A choice question with its own L1 default and 4 options.
	q := Question{
		NegativeMarkMode: NegativeMarkAccumulative,
		WrongsPerPoint:   3,
		Options:          []QuestionOption{{ID: "a"}, {ID: "b"}, {ID: "c"}, {ID: "d"}},
	}
	quizWide := NegativeMarkConfig{Mode: NegativeMarkPerWrong, NegativeValue: 0.9}

	// none rule default forces no penalty even over the question's own L1 default.
	forceNone := NegativeMarkNone
	got := ResolveNegativeMark(q, nil, &forceNone, quizWide)
	if got.Mode != NegativeMarkNone {
		t.Fatalf("rule default none must force none, got %+v", got)
	}

	// per_wrong derives negative_value from the option count (4 => 0.25).
	perWrong := NegativeMarkPerWrong
	got = ResolveNegativeMark(q, nil, &perWrong, quizWide)
	if got.Mode != NegativeMarkPerWrong || got.NegativeValue != FractionFor(4) {
		t.Fatalf("rule default per_wrong should auto-fraction from options, got %+v", got)
	}
	if got.Fraction != FractionFor(4) {
		t.Fatalf("per_wrong fraction = value, got %v", got.Fraction)
	}

	// accumulative derives wrongs_per_point = clamp(optionCount, 2, 5).
	accumulative := NegativeMarkAccumulative
	got = ResolveNegativeMark(q, nil, &accumulative, quizWide)
	if got.Mode != NegativeMarkAccumulative || got.WrongsPerPoint != 4 {
		t.Fatalf("rule default accumulative should clamp option count, got %+v", got)
	}

	// Two-option question clamps up to the minimum of 2.
	twoOpt := Question{Options: []QuestionOption{{ID: "a"}, {ID: "b"}}}
	got = ResolveNegativeMark(twoOpt, nil, &accumulative, quizWide)
	if got.WrongsPerPoint != 2 {
		t.Fatalf("accumulative should clamp to 2 min, got %+v", got)
	}

	// Per-question override beats the rule default.
	override := &QuizQuestionNegativeOverride{QuestionID: qid, Mode: NegativeMarkPerWrong, NegativeValue: 0.5}
	got = ResolveNegativeMark(q, override, &forceNone, quizWide)
	if got.Mode != NegativeMarkPerWrong || got.NegativeValue != 0.5 {
		t.Fatalf("override should beat rule default, got %+v", got)
	}

	// nil rule default falls through to the question's own L1 default.
	got = ResolveNegativeMark(q, nil, nil, quizWide)
	if got.Mode != NegativeMarkAccumulative || got.WrongsPerPoint != 3 {
		t.Fatalf("nil rule default should fall through to question, got %+v", got)
	}
}
