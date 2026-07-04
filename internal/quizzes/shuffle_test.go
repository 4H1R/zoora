package quizzes

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestShuffleDeterministic(t *testing.T) {
	seed := uuid.MustParse("018f9c2a-0000-7000-8000-000000000001")
	in := []string{"a", "b", "c", "d", "e"}

	got1 := shuffleStrings(seed, "opts:q1", in)
	got2 := shuffleStrings(seed, "opts:q1", in)

	assert.Equal(t, got1, got2, "same seed+salt must produce same order")
	assert.ElementsMatch(t, in, got1, "must be a permutation, no loss")
}

func TestShuffleDiffersBySeed(t *testing.T) {
	in := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	a := shuffleStrings(uuid.MustParse("018f9c2a-0000-7000-8000-000000000001"), "opts:q1", in)
	b := shuffleStrings(uuid.MustParse("018f9c2a-0000-7000-8000-000000000002"), "opts:q1", in)
	assert.NotEqual(t, a, b, "different seeds should (very likely) differ")
}

func TestShuffleDiffersBySalt(t *testing.T) {
	// Same submission, two questions: permutations must be independent, otherwise
	// every same-option-count question shares one pattern.
	seed := uuid.MustParse("018f9c2a-0000-7000-8000-000000000001")
	in := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	a := shuffleStrings(seed, "opts:q1", in)
	b := shuffleStrings(seed, "opts:q2", in)
	assert.NotEqual(t, a, b, "different salts should (very likely) differ")
}
