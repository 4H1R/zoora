package billing

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

// pdfStorage is the S3 subset the renderer needs (upload bytes under a key).
type pdfStorage interface {
	PutObject(ctx context.Context, key string, body []byte, contentType string) error
}

// orgNamer resolves an org's display name for the receipt header. Satisfied by
// domain.OrganizationRepository.
type orgNamer interface {
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error)
}

type chromePDFRenderer struct {
	storage pdfStorage
	orgRepo orgNamer
	issuer  IssuerConfig
	// execAllocatorURL is a running Chromium's remote-debugging URL, or empty to
	// launch a local headless Chromium (binary must be on PATH in the image).
	execAllocatorURL string
}

// NewPDFRenderer builds the chromedp-backed receipt renderer. chromeURL is an
// optional remote Chromium debugging URL; empty launches a local headless
// Chromium via the exec allocator.
func NewPDFRenderer(storage pdfStorage, orgRepo orgNamer, issuer IssuerConfig, chromeURL string) pdfRenderer {
	return &chromePDFRenderer{storage: storage, orgRepo: orgRepo, issuer: issuer, execAllocatorURL: chromeURL}
}

func (r *chromePDFRenderer) RenderAndStore(ctx context.Context, inv *domain.Invoice) (string, error) {
	org, err := r.orgRepo.FindByID(ctx, inv.OrganizationID)
	if err != nil {
		return "", err
	}
	html, err := renderReceiptHTML(buildReceiptVM(inv, org.Name, r.issuer))
	if err != nil {
		return "", fmt.Errorf("billing.pdf.renderHTML: %w", err)
	}
	pdf, err := r.htmlToPDF(ctx, html)
	if err != nil {
		return "", err
	}
	key := invoicePDFKey(inv)
	if err := r.storage.PutObject(ctx, key, pdf, "application/pdf"); err != nil {
		return "", fmt.Errorf("billing.pdf.put: %w", err)
	}
	return key, nil
}

func invoicePDFKey(inv *domain.Invoice) string {
	return "orgs/" + inv.OrganizationID.String() + "/invoices/" + inv.ID.String() + ".pdf"
}

func (r *chromePDFRenderer) htmlToPDF(ctx context.Context, html []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	var allocCtx context.Context
	var allocCancel context.CancelFunc
	if r.execAllocatorURL != "" {
		allocCtx, allocCancel = chromedp.NewRemoteAllocator(ctx, r.execAllocatorURL)
	} else {
		// --no-sandbox is required inside most containers.
		opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.NoSandbox)
		allocCtx, allocCancel = chromedp.NewExecAllocator(ctx, opts...)
	}
	defer allocCancel()

	taskCtx, taskCancel := chromedp.NewContext(allocCtx)
	defer taskCancel()

	// Load the HTML via a data URL so no web server is needed.
	dataURL := "data:text/html;charset=utf-8;base64," + base64.StdEncoding.EncodeToString(html)
	var pdf []byte
	err := chromedp.Run(taskCtx,
		chromedp.Navigate(dataURL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			pdf, _, err = page.PrintToPDF().
				WithPrintBackground(true).
				WithPreferCSSPageSize(true).
				Do(ctx)
			return err
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("billing.pdf.chromedp: %w", err)
	}
	return pdf, nil
}

var _ pdfRenderer = (*chromePDFRenderer)(nil)
