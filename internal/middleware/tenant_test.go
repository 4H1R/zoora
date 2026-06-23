package middleware

import "testing"

func TestParseHostLabel(t *testing.T) {
	cases := []struct {
		host, base, want string
	}{
		{"acme.localhost:3000", "localhost", "acme"},
		{"acme.zoora.ir", "zoora.ir", "acme"},
		{"admin.zoora.ir", "zoora.ir", "admin"},
		{"zoora.ir", "zoora.ir", ""},        // apex
		{"localhost:8080", "localhost", ""}, // apex dev
		{"acme.zoora.ir:443", "zoora.ir", "acme"},
	}
	for _, c := range cases {
		if got := parseHostLabel(c.host, c.base); got != c.want {
			t.Fatalf("parseHostLabel(%q,%q)=%q want %q", c.host, c.base, got, c.want)
		}
	}
}
