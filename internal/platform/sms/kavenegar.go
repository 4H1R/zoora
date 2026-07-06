// Package sms implements domain.SMSSender against Kavenegar
// (https://kavenegar.com — Iranian SMS provider).
package sms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Config struct {
	APIKey      string
	Sender      string // optional dedicated line number
	OTPTemplate string // Kavenegar verify-lookup template name
	BaseURL     string // default https://api.kavenegar.com; overridable in tests
}

type Kavenegar struct {
	cfg    Config
	http   *http.Client
	logger *slog.Logger
}

func NewKavenegar(cfg Config, logger *slog.Logger) *Kavenegar {
	if logger == nil {
		logger = slog.Default()
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.kavenegar.com"
	}
	return &Kavenegar{cfg: cfg, http: &http.Client{Timeout: 30 * time.Second}, logger: logger}
}

type kavenegarReturn struct {
	Return struct {
		Status  int    `json:"status"`
		Message string `json:"message"`
	} `json:"return"`
}

func (k *Kavenegar) post(ctx context.Context, path string, form url.Values) error {
	endpoint := fmt.Sprintf("%s/v1/%s/%s", k.cfg.BaseURL, k.cfg.APIKey, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("sms.kavenegar: building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.http.Do(req)
	if err != nil {
		return fmt.Errorf("sms.kavenegar: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("sms.kavenegar: reading response: %w", err)
	}
	var kr kavenegarReturn
	if err := json.Unmarshal(raw, &kr); err != nil {
		return fmt.Errorf("sms.kavenegar: decoding response (status %d): %w", resp.StatusCode, err)
	}
	if kr.Return.Status != 200 {
		return fmt.Errorf("sms.kavenegar: provider error %d: %s", kr.Return.Status, kr.Return.Message)
	}
	return nil
}

// SendBulk implements domain.SMSSender. Kavenegar accepts comma-joined
// receptors in one call (cap batches at ~100 upstream of this client).
func (k *Kavenegar) SendBulk(ctx context.Context, phones []string, message string) error {
	form := url.Values{
		"receptor": {strings.Join(phones, ",")},
		"message":  {message},
	}
	if k.cfg.Sender != "" {
		form.Set("sender", k.cfg.Sender)
	}
	return k.post(ctx, "sms/send.json", form)
}

// SendOTP implements domain.SMSSender via Kavenegar verify/lookup (dedicated
// OTP route — faster and cheaper than a regular line).
func (k *Kavenegar) SendOTP(ctx context.Context, phone, code string) error {
	form := url.Values{
		"receptor": {phone},
		"token":    {code},
		"template": {k.cfg.OTPTemplate},
	}
	return k.post(ctx, "verify/lookup.json", form)
}
