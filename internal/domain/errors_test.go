package domain

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestMapError_ParticipantSentinels(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		wantCode int
		wantTag  string
	}{
		{"participant not found", ErrParticipantNotFound, http.StatusNotFound, "PARTICIPANT_NOT_FOUND"},
		{"invalid role", ErrInvalidParticipantRole, http.StatusUnprocessableEntity, "INVALID_PARTICIPANT_ROLE"},
		{"cannot change host", ErrCannotChangeHostRole, http.StatusConflict, "CANNOT_CHANGE_HOST_ROLE"},
		// Wrapped errors must still map via errors.Is.
		{"wrapped not found", fmt.Errorf("loading: %w", ErrParticipantNotFound), http.StatusNotFound, "PARTICIPANT_NOT_FOUND"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			code, tag := MapError(tc.err)
			if code != tc.wantCode || tag != tc.wantTag {
				t.Fatalf("MapError(%v) = (%d, %q), want (%d, %q)", tc.err, code, tag, tc.wantCode, tc.wantTag)
			}
		})
	}
}

func TestPlanErrorsMapTo402(t *testing.T) {
	for _, err := range []error{
		ErrFeatureNotInPlan,
		ErrPlanLimitReached,
		&PlanError{Sentinel: ErrFeatureNotInPlan, Feature: string(FeatureRecording), Plan: string(PlanFree)},
		&PlanError{Sentinel: ErrPlanLimitReached, Limit: string(LimitMaxUsers), Current: 10, Ceiling: 10, Plan: string(PlanFree)},
	} {
		code, _ := MapError(err)
		if code != http.StatusPaymentRequired {
			t.Fatalf("%v mapped to %d, want 402", err, code)
		}
	}
}

func TestPlanErrorUnwraps(t *testing.T) {
	pe := &PlanError{Sentinel: ErrPlanLimitReached}
	if !errors.Is(pe, ErrPlanLimitReached) {
		t.Fatal("PlanError must unwrap to its sentinel")
	}
}

func TestNewFeatureAndLimitErrorCodes(t *testing.T) {
	if _, code := MapError(NewFeatureError(PlanFree, FeatureRecording)); code != "FEATURE_NOT_IN_PLAN" {
		t.Fatalf("feature error code = %q", code)
	}
	if _, code := MapError(NewLimitError(PlanFree, LimitMaxUsers, 10, 10)); code != "PLAN_LIMIT_REACHED" {
		t.Fatalf("limit error code = %q", code)
	}
}
