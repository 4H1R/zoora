package payment

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestZarinpalRequestAndVerify(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/pg/v4/payment/request.json":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":   map[string]any{"code": 100, "authority": "A0000000000000000000000000000012345"},
				"errors": []any{},
			})
		case "/pg/v4/payment/verify.json":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data":   map[string]any{"code": 100, "ref_id": 2000000201},
				"errors": []any{},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	z := NewZarinpal(ZarinpalConfig{MerchantID: "test-merchant", BaseURL: srv.URL, StartPayURL: srv.URL})
	ctx := context.Background()

	out, err := z.Request(ctx, RequestInput{Amount: 15_000_000, Currency: "IRR", CallbackURL: "https://x/cb", Description: "pro"})
	if err != nil {
		t.Fatalf("Request: %v", err)
	}
	if out.Authority == "" {
		t.Fatal("empty authority")
	}
	if out.RedirectURL != srv.URL+"/pg/StartPay/"+out.Authority {
		t.Errorf("redirect = %s", out.RedirectURL)
	}

	v, err := z.Verify(ctx, VerifyInput{Authority: out.Authority, Amount: 15_000_000})
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if v.Status != VerifyStatusSucceeded {
		t.Errorf("status = %s, want succeeded", v.Status)
	}
	if v.RefID != "2000000201" {
		t.Errorf("ref_id = %s", v.RefID)
	}
}
