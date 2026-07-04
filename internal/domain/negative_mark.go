package domain

import "math"

// NegativeMarkMode controls how wrong choice selections are penalized.
type NegativeMarkMode string

const (
	NegativeMarkNone         NegativeMarkMode = "none"
	NegativeMarkPerWrong     NegativeMarkMode = "per_wrong"
	NegativeMarkAccumulative NegativeMarkMode = "accumulative"
)

func (m NegativeMarkMode) Valid() bool {
	switch m {
	case NegativeMarkNone, NegativeMarkPerWrong, NegativeMarkAccumulative:
		return true
	}
	return false
}

// NegativeMarkConfig is the effective negative-marking setting for one
// question in one attempt. Fraction is display-only for the frontend.
type NegativeMarkConfig struct {
	Mode           NegativeMarkMode `json:"mode"`
	NegativeValue  float64          `json:"negative_value"`
	WrongsPerPoint int              `json:"wrongs_per_point"`
	Fraction       float64          `json:"fraction"`
}

// Penalty returns the points deducted for wrongCount wrong selections.
// accumulative discards the remainder (single grading pass).
func (c NegativeMarkConfig) Penalty(wrongCount int) float64 {
	if wrongCount <= 0 {
		return 0
	}
	switch c.Mode {
	case NegativeMarkPerWrong:
		return float64(wrongCount) * c.NegativeValue
	case NegativeMarkAccumulative:
		if c.WrongsPerPoint <= 0 {
			return 0
		}
		return math.Floor(float64(wrongCount)/float64(c.WrongsPerPoint)) * 1.0
	default:
		return 0
	}
}

// FractionFor returns the suggested per-wrong fraction for an option count.
// Matches the spec table {2:0.5,3:0.33,4:0.25,5:0.2}, else 1/n.
func FractionFor(optionCount int) float64 {
	switch optionCount {
	case 2:
		return 0.5
	case 3:
		return 0.33
	case 4:
		return 0.25
	case 5:
		return 0.2
	}
	if optionCount <= 0 {
		return 0
	}
	return 1.0 / float64(optionCount)
}

// ValidateNegativeMark checks a negative-marking triple. none ignores the
// other fields (callers normalize via NormalizeNegativeMark).
func ValidateNegativeMark(mode NegativeMarkMode, value float64, wrongsPerPoint int) error {
	if !mode.Valid() {
		return NewValidationError(map[string]string{"negative_mark_mode": "invalid mode"})
	}
	switch mode {
	case NegativeMarkPerWrong:
		if value <= 0 {
			return NewValidationError(map[string]string{"negative_value": "must be greater than 0 for per_wrong"})
		}
	case NegativeMarkAccumulative:
		if wrongsPerPoint < 2 || wrongsPerPoint > 5 {
			return NewValidationError(map[string]string{"wrongs_per_point": "must be between 2 and 5 for accumulative"})
		}
	}
	return nil
}

// NormalizeNegativeMark zeroes irrelevant fields for the given mode so stored
// rows stay clean (none => all zero; per_wrong => wpp zero; accumulative => value zero).
func NormalizeNegativeMark(mode NegativeMarkMode, value float64, wrongsPerPoint int) (NegativeMarkMode, float64, int) {
	switch mode {
	case NegativeMarkPerWrong:
		return mode, value, 0
	case NegativeMarkAccumulative:
		return mode, 0, wrongsPerPoint
	default:
		return NegativeMarkNone, 0, 0
	}
}

// ResolveNegativeMark returns the effective config for a question in an
// attempt, applying the priority: per-Q override > rule default > question
// default > quiz-wide > none. The Fraction field is set for display.
//
// ruleDefault is the rule-wide default (Layer 2-bank). Being a nullable mode
// with no stored numbers, its per_wrong/accumulative variants derive their
// numbers from the question's option count; "none" forces no penalty even when
// the question carries its own default. nil falls through to the question.
func ResolveNegativeMark(q Question, override *QuizQuestionNegativeOverride, ruleDefault *NegativeMarkMode, quizWide NegativeMarkConfig) NegativeMarkConfig {
	var cfg NegativeMarkConfig
	switch {
	case override != nil && override.Mode != NegativeMarkNone && override.Mode.Valid():
		cfg = NegativeMarkConfig{Mode: override.Mode, NegativeValue: override.NegativeValue, WrongsPerPoint: override.WrongsPerPoint}
	case ruleDefault != nil && ruleDefault.Valid():
		switch *ruleDefault {
		case NegativeMarkNone:
			return NegativeMarkConfig{Mode: NegativeMarkNone}
		case NegativeMarkPerWrong:
			cfg = NegativeMarkConfig{Mode: NegativeMarkPerWrong, NegativeValue: FractionFor(len(q.Options))}
		case NegativeMarkAccumulative:
			cfg = NegativeMarkConfig{Mode: NegativeMarkAccumulative, WrongsPerPoint: clampInt(len(q.Options), 2, 5)}
		}
	case q.NegativeMarkMode != NegativeMarkNone && q.NegativeMarkMode.Valid():
		cfg = NegativeMarkConfig{Mode: q.NegativeMarkMode, NegativeValue: q.NegativeValue, WrongsPerPoint: q.WrongsPerPoint}
	case quizWide.Mode != NegativeMarkNone && quizWide.Mode.Valid():
		cfg = NegativeMarkConfig{Mode: quizWide.Mode, NegativeValue: quizWide.NegativeValue, WrongsPerPoint: quizWide.WrongsPerPoint}
	default:
		return NegativeMarkConfig{Mode: NegativeMarkNone}
	}
	cfg.Fraction = fractionForConfig(cfg)
	return cfg
}

// clampInt returns n bounded to the inclusive [lo, hi] range.
func clampInt(n, lo, hi int) int {
	if n < lo {
		return lo
	}
	if n > hi {
		return hi
	}
	return n
}

func fractionForConfig(c NegativeMarkConfig) float64 {
	switch c.Mode {
	case NegativeMarkPerWrong:
		return c.NegativeValue
	case NegativeMarkAccumulative:
		if c.WrongsPerPoint <= 0 {
			return 0
		}
		return FractionFor(c.WrongsPerPoint)
	default:
		return 0
	}
}
