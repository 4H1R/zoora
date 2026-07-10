package chathub

import (
	"net/http"
	"testing"
)

func requestWithOrigin(t *testing.T, origin string) *http.Request {
	t.Helper()
	r, err := http.NewRequest(http.MethodGet, "/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	return r
}

func TestOriginChecker(t *testing.T) {
	cases := []struct {
		name    string
		allowed []string
		origin  string
		want    bool
	}{
		{"wildcard allows anything", []string{"*"}, "https://evil.example", true},
		{"exact match allowed", []string{"https://app.zoora.io"}, "https://app.zoora.io", true},
		{"case-insensitive match", []string{"https://App.Zoora.io"}, "https://app.zoora.io", true},
		{"trailing slash normalized", []string{"https://app.zoora.io/"}, "https://app.zoora.io", true},
		{"foreign origin rejected", []string{"https://app.zoora.io"}, "https://evil.example", false},
		{"subdomain not implicitly allowed", []string{"https://zoora.io"}, "https://evil.zoora.io", false},
		{"no Origin header allowed (non-browser client)", []string{"https://app.zoora.io"}, "", true},
		{"empty allow-list rejects browser origins", nil, "https://app.zoora.io", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			check := originChecker(tc.allowed)
			if got := check(requestWithOrigin(t, tc.origin)); got != tc.want {
				t.Fatalf("originChecker(%v) with Origin %q = %v, want %v", tc.allowed, tc.origin, got, tc.want)
			}
		})
	}
}
