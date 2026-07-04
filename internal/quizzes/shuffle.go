package quizzes

import (
	"encoding/binary"
	"hash/fnv"
	"math/rand"

	"github.com/google/uuid"
)

// seedFrom derives a stable int64 seed from a UUID (submission id) combined
// with a salt discriminator (e.g. "opts:"+questionID, "questions") so option
// and question shuffles use independent but reproducible seeds.
func seedFrom(id uuid.UUID, salt string) int64 {
	h := fnv.New64a()
	b := id
	_, _ = h.Write(b[:])
	_, _ = h.Write([]byte(salt))
	return int64(binary.LittleEndian.Uint64(h.Sum(nil)[:8])) //nolint:gosec // deterministic, non-crypto shuffle
}

// shuffleStrings returns a deterministic permutation of in for the given
// seed+salt. Input slice is not mutated.
func shuffleStrings(seed uuid.UUID, salt string, in []string) []string {
	out := make([]string, len(in))
	copy(out, in)
	r := rand.New(rand.NewSource(seedFrom(seed, salt))) //nolint:gosec // non-crypto
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}
