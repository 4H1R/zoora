package domain

import (
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
