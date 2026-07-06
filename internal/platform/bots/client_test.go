package bots

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendMessagePostsToBotAPI(t *testing.T) {
	var gotPath string
	var gotBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		_, _ = w.Write([]byte(`{"ok":true,"result":{}}`))
	}))
	defer srv.Close()

	c, err := NewClient(Config{BaseURL: srv.URL, Token: "TOKEN"}, nil)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if err := c.SendMessage(context.Background(), "12345", "hello"); err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	if gotPath != "/botTOKEN/sendMessage" {
		t.Fatalf("path = %s, want /botTOKEN/sendMessage", gotPath)
	}
	if gotBody["chat_id"] != "12345" || gotBody["text"] != "hello" {
		t.Fatalf("body = %v", gotBody)
	}
}

func TestSendMessageAPIErrorSurfaces(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"ok":false,"description":"chat not found"}`))
	}))
	defer srv.Close()

	c, _ := NewClient(Config{BaseURL: srv.URL, Token: "TOKEN"}, nil)
	if err := c.SendMessage(context.Background(), "1", "x"); err == nil {
		t.Fatal("expected error from ok=false response")
	}
}

func TestGetUpdatesParsesMessages(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true,"result":[{"update_id":7,"message":{"chat":{"id":42},"text":"/start abc"}}]}`))
	}))
	defer srv.Close()

	c, _ := NewClient(Config{BaseURL: srv.URL, Token: "TOKEN"}, nil)
	ups, err := c.GetUpdates(context.Background(), 0, 1)
	if err != nil {
		t.Fatalf("GetUpdates: %v", err)
	}
	if len(ups) != 1 || ups[0].UpdateID != 7 || ups[0].Message.Chat.ID != 42 || ups[0].Message.Text != "/start abc" {
		t.Fatalf("updates = %+v", ups)
	}
}
