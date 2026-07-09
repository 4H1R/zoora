package conversations

import "github.com/google/uuid"

// directKey builds a deterministic, order-independent key for a DM pair,
// used by the uq_conversations_direct_key unique index to guarantee one
// direct conversation per (org, user-pair).
func directKey(a, b uuid.UUID) string {
	if a.String() > b.String() {
		a, b = b, a
	}
	return a.String() + ":" + b.String()
}
