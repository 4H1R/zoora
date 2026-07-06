// Package bots is a minimal Telegram-Bot-API client. Bale (tapi.bale.ai)
// speaks the same protocol, so one client serves both — construct one
// instance per platform with its own base URL, token, and optional proxy.
package bots

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

type Config struct {
	BaseURL  string // e.g. https://api.telegram.org or https://tapi.bale.ai
	Token    string
	ProxyURL string // optional http/socks5 proxy (Telegram is blocked in Iran)
}

type Client struct {
	baseURL string
	token   string
	http    *http.Client
	logger  *slog.Logger
}

func NewClient(cfg Config, logger *slog.Logger) (*Client, error) {
	if logger == nil {
		logger = slog.Default()
	}
	transport := &http.Transport{}
	if cfg.ProxyURL != "" {
		proxy, err := url.Parse(cfg.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("bots: parsing proxy URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxy)
	}
	return &Client{
		baseURL: cfg.BaseURL,
		token:   cfg.Token,
		// Long-poll GetUpdates needs headroom beyond the poll timeout.
		http:   &http.Client{Timeout: 50 * time.Second, Transport: transport},
		logger: logger,
	}, nil
}

type apiResponse struct {
	OK          bool            `json:"ok"`
	Description string          `json:"description"`
	Result      json.RawMessage `json:"result"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type Message struct {
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type Update struct {
	UpdateID int64    `json:"update_id"`
	Message  *Message `json:"message"`
}

func (c *Client) call(ctx context.Context, method string, params any) (json.RawMessage, error) {
	body, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("bots.%s: marshaling params: %w", method, err)
	}
	endpoint := fmt.Sprintf("%s/bot%s/%s", c.baseURL, c.token, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("bots.%s: building request: %w", method, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("bots.%s: %w", method, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("bots.%s: reading response: %w", method, err)
	}
	var api apiResponse
	if err := json.Unmarshal(raw, &api); err != nil {
		return nil, fmt.Errorf("bots.%s: decoding response (status %d): %w", method, resp.StatusCode, err)
	}
	if !api.OK {
		return nil, fmt.Errorf("bots.%s: api error: %s", method, api.Description)
	}
	return api.Result, nil
}

// SendMessage implements domain.BotSender.
func (c *Client) SendMessage(ctx context.Context, chatID string, text string) error {
	_, err := c.call(ctx, "sendMessage", map[string]any{
		"chat_id": chatID,
		"text":    text,
	})
	return err
}

// GetUpdates long-polls for new updates starting after offset.
func (c *Client) GetUpdates(ctx context.Context, offset int64, timeoutSec int) ([]Update, error) {
	result, err := c.call(ctx, "getUpdates", map[string]any{
		"offset":          offset,
		"timeout":         timeoutSec,
		"allowed_updates": []string{"message"},
	})
	if err != nil {
		return nil, err
	}
	var ups []Update
	if err := json.Unmarshal(result, &ups); err != nil {
		return nil, fmt.Errorf("bots.getUpdates: decoding updates: %w", err)
	}
	return ups, nil
}
