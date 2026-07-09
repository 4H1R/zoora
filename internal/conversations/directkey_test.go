package conversations

import (
	"testing"

	"github.com/google/uuid"
)

func TestDirectKey_OrderIndependent(t *testing.T) {
	a := uuid.MustParse("00000000-0000-0000-0000-0000000000aa")
	b := uuid.MustParse("00000000-0000-0000-0000-0000000000bb")
	if directKey(a, b) != directKey(b, a) {
		t.Fatal("directKey must be order-independent")
	}
	if directKey(a, b) != a.String()+":"+b.String() {
		t.Fatalf("unexpected key: %s", directKey(a, b))
	}
}
