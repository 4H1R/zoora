package domain_test

import (
	"testing"

	"github.com/4H1R/zoora/internal/domain"
)

func TestValidateSlug(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"acme", false},
		{"acme-2", false},
		{"a1", false},
		{"AB", true},    // uppercase
		{"a", true},     // too short (<2)
		{"-acme", true}, // leading dash
		{"acme_", true}, // underscore
		{"api", true},   // reserved
		{"admin", true}, // reserved
		{"with space", true},
	}
	for _, c := range cases {
		err := domain.ValidateSlug(c.in)
		if (err != nil) != c.wantErr {
			t.Fatalf("ValidateSlug(%q) err=%v, wantErr=%v", c.in, err, c.wantErr)
		}
	}
}
