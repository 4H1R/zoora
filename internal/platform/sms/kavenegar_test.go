package sms

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendBulkPostsForm(t *testing.T) {
	var gotPath, gotReceptor, gotMessage string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = r.ParseForm()
		gotReceptor = r.Form.Get("receptor")
		gotMessage = r.Form.Get("message")
		_, _ = w.Write([]byte(`{"return":{"status":200,"message":"ok"},"entries":[]}`))
	}))
	defer srv.Close()

	k := NewKavenegar(Config{APIKey: "KEY", BaseURL: srv.URL}, nil)
	err := k.SendBulk(context.Background(), []string{"09120000001", "09120000002"}, "hi")
	if err != nil {
		t.Fatalf("SendBulk: %v", err)
	}
	if gotPath != "/v1/KEY/sms/send.json" {
		t.Fatalf("path = %s", gotPath)
	}
	if gotReceptor != "09120000001,09120000002" || gotMessage != "hi" {
		t.Fatalf("receptor=%s message=%s", gotReceptor, gotMessage)
	}
}

func TestSendOTPUsesVerifyLookup(t *testing.T) {
	var gotPath, gotToken, gotTemplate string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = r.ParseForm()
		gotToken = r.Form.Get("token")
		gotTemplate = r.Form.Get("template")
		_, _ = w.Write([]byte(`{"return":{"status":200,"message":"ok"},"entries":[]}`))
	}))
	defer srv.Close()

	k := NewKavenegar(Config{APIKey: "KEY", OTPTemplate: "zoora-otp", BaseURL: srv.URL}, nil)
	if err := k.SendOTP(context.Background(), "09120000001", "123456"); err != nil {
		t.Fatalf("SendOTP: %v", err)
	}
	if gotPath != "/v1/KEY/verify/lookup.json" {
		t.Fatalf("path = %s", gotPath)
	}
	if gotToken != "123456" || gotTemplate != "zoora-otp" {
		t.Fatalf("token=%s template=%s", gotToken, gotTemplate)
	}
}

func TestProviderErrorSurfaces(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"return":{"status":403,"message":"invalid api key"}}`))
	}))
	defer srv.Close()

	k := NewKavenegar(Config{APIKey: "BAD", BaseURL: srv.URL}, nil)
	if err := k.SendBulk(context.Background(), []string{"0912"}, "x"); err == nil {
		t.Fatal("expected provider error")
	}
}
