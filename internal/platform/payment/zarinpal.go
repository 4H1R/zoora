package payment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type ZarinpalConfig struct {
	MerchantID  string
	BaseURL     string // e.g. https://payment.zarinpal.com  (sandbox: https://sandbox.zarinpal.com)
	StartPayURL string // usually same host as BaseURL
}

type zarinpal struct {
	cfg    ZarinpalConfig
	client *http.Client
}

func NewZarinpal(cfg ZarinpalConfig) Gateway {
	if cfg.StartPayURL == "" {
		cfg.StartPayURL = cfg.BaseURL
	}
	return &zarinpal{cfg: cfg, client: &http.Client{Timeout: 15 * time.Second}}
}

func (z *zarinpal) Name() string { return "zarinpal" }

type zpRequestBody struct {
	MerchantID  string            `json:"merchant_id"`
	Amount      int64             `json:"amount"`
	CallbackURL string            `json:"callback_url"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type zpResponse struct {
	Data struct {
		Code      int    `json:"code"`
		Authority string `json:"authority"`
		RefID     any    `json:"ref_id"`
		Message   string `json:"message"`
	} `json:"data"`
	// errors is an object on failure, [] on success — decode leniently.
	Errors json.RawMessage `json:"errors"`
}

func (z *zarinpal) Request(ctx context.Context, in RequestInput) (RequestOutput, error) {
	body := zpRequestBody{
		MerchantID:  z.cfg.MerchantID,
		Amount:      in.Amount, // Rial
		CallbackURL: in.CallbackURL,
		Description: in.Description,
	}
	var resp zpResponse
	if err := z.post(ctx, "/pg/v4/payment/request.json", body, &resp); err != nil {
		return RequestOutput{}, err
	}
	if resp.Data.Code != 100 || resp.Data.Authority == "" {
		return RequestOutput{}, fmt.Errorf("zarinpal request failed: code=%d %s", resp.Data.Code, resp.Data.Message)
	}
	return RequestOutput{
		Authority:   resp.Data.Authority,
		RedirectURL: z.cfg.StartPayURL + "/pg/StartPay/" + resp.Data.Authority,
	}, nil
}

type zpVerifyBody struct {
	MerchantID string `json:"merchant_id"`
	Amount     int64  `json:"amount"`
	Authority  string `json:"authority"`
}

func (z *zarinpal) Verify(ctx context.Context, in VerifyInput) (VerifyOutput, error) {
	body := zpVerifyBody{MerchantID: z.cfg.MerchantID, Amount: in.Amount, Authority: in.Authority}
	var resp zpResponse
	raw, err := z.postRaw(ctx, "/pg/v4/payment/verify.json", body, &resp)
	if err != nil {
		return VerifyOutput{}, err
	}
	out := VerifyOutput{Raw: raw, RefID: refIDToString(resp.Data.RefID)}
	switch resp.Data.Code {
	case 100:
		out.Status = VerifyStatusSucceeded
	case 101:
		out.Status = VerifyStatusAlreadyVerified
	default:
		out.Status = VerifyStatusFailed
	}
	return out, nil
}

func (z *zarinpal) post(ctx context.Context, path string, body any, out *zpResponse) error {
	_, err := z.postRaw(ctx, path, body, out)
	return err
}

func (z *zarinpal) postRaw(ctx context.Context, path string, body any, out *zpResponse) ([]byte, error) {
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("zarinpal marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, z.cfg.BaseURL+path, bytes.NewReader(buf))
	if err != nil {
		return nil, fmt.Errorf("zarinpal new request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	res, err := z.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("zarinpal http: %w", err)
	}
	defer res.Body.Close()
	raw := new(bytes.Buffer)
	if _, err := raw.ReadFrom(res.Body); err != nil {
		return nil, fmt.Errorf("zarinpal read body: %w", err)
	}
	if err := json.Unmarshal(raw.Bytes(), out); err != nil {
		return nil, fmt.Errorf("zarinpal decode (%s): %w", raw.String(), err)
	}
	return raw.Bytes(), nil
}

func refIDToString(v any) string {
	switch n := v.(type) {
	case float64:
		return strconv.FormatInt(int64(n), 10)
	case string:
		return n
	default:
		return ""
	}
}
