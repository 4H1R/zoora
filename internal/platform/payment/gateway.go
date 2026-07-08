// Package payment defines a provider-agnostic payment gateway boundary. It
// imports no domain/business types — callers (the billing service) adapt
// between domain models and these plain structs.
package payment

import "context"

type RequestInput struct {
	Amount      int64  // minor units (Rial for IRR)
	Currency    string // ISO 4217
	CallbackURL string
	Description string
	Mobile      string
	Email       string
}

type RequestOutput struct {
	Authority   string // provider token (Zarinpal authority)
	RedirectURL string // where to send the payer's browser
}

type VerifyInput struct {
	Authority string
	Amount    int64 // expected amount (Rial) — provider must match
}

type VerifyStatus string

const (
	VerifyStatusSucceeded       VerifyStatus = "succeeded"
	VerifyStatusAlreadyVerified VerifyStatus = "already_verified" // Zarinpal code 101
	VerifyStatusFailed          VerifyStatus = "failed"
)

type VerifyOutput struct {
	Status VerifyStatus
	RefID  string
	Raw    []byte // provider response JSON, stored on the payment row
}

// Gateway is one payment provider. Name() is the stable key (e.g. "zarinpal").
type Gateway interface {
	Name() string
	Request(ctx context.Context, in RequestInput) (RequestOutput, error)
	Verify(ctx context.Context, in VerifyInput) (VerifyOutput, error)
}
