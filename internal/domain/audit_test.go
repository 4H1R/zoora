package domain

import "testing"

func TestAuditActionValid(t *testing.T) {
	if !AuditCreated.Valid() {
		t.Fatal("AuditCreated should be valid")
	}
	if AuditAction("frobnicate").Valid() {
		t.Fatal("unknown action should be invalid")
	}
}

func TestAuditTargetTypeValid(t *testing.T) {
	if !AuditTargetClass.Valid() {
		t.Fatal("AuditTargetClass should be valid")
	}
	if AuditTargetType("nonsense").Valid() {
		t.Fatal("unknown target type should be invalid")
	}
}

func TestAuditOutcomeValid(t *testing.T) {
	if !AuditOutcomeSuccess.Valid() || !AuditOutcomeDenied.Valid() {
		t.Fatal("success and denied are valid outcomes")
	}
	if AuditOutcome("maybe").Valid() {
		t.Fatal("unknown outcome should be invalid")
	}
}
