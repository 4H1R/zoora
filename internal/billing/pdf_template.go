package billing

import (
	"bytes"
	_ "embed"
	"html/template"

	"github.com/4H1R/zoora/internal/domain"
)

//go:embed templates/receipt.html.tmpl
var receiptTmplSrc string

var receiptTmpl = template.Must(template.New("receipt").Parse(receiptTmplSrc))

type receiptItemVM struct {
	Description  string
	QuantityFa   string
	UnitAmountFa string
	AmountFa     string
}

type receiptVM struct {
	IssuerName, IssuerEconomicID, IssuerAddress, IssuerPhone string
	Number, OrgName                                          string
	IssuedAtFa, PaidAtFa                                     string
	StatusFa, StatusClass                                    string
	Items                                                    []receiptItemVM
	SubtotalFa, TaxAmountFa, TotalFa, TaxPercentFa           string
	HasTax                                                   bool
	RefID, RefIDFa, GatewayFa                                string
}

func buildReceiptVM(inv *domain.Invoice, orgName string, issuer IssuerConfig) receiptVM {
	vm := receiptVM{
		IssuerName:       issuer.Name,
		IssuerEconomicID: issuer.EconomicID,
		IssuerAddress:    issuer.Address,
		IssuerPhone:      issuer.Phone,
		OrgName:          orgName,
		SubtotalFa:       formatTomanFa(inv.Subtotal),
		TotalFa:          formatTomanFa(inv.Total),
		HasTax:           inv.TaxPercent > 0,
		TaxPercentFa:     toPersianDigits(itoaInt(inv.TaxPercent)),
		TaxAmountFa:      formatTomanFa(inv.TaxAmount),
		StatusFa:         invoiceStatusFa(inv.Status),
		StatusClass:      string(inv.Status),
	}
	if inv.Number != nil {
		vm.Number = toPersianDigits(*inv.Number)
	}
	if inv.IssuedAt != nil {
		vm.IssuedAtFa = formatJalaliFa(*inv.IssuedAt)
	}
	if inv.PaidAt != nil {
		vm.PaidAtFa = formatJalaliFa(*inv.PaidAt)
	}
	for _, it := range inv.Items {
		vm.Items = append(vm.Items, receiptItemVM{
			Description:  it.Description,
			QuantityFa:   toPersianDigits(itoaInt(it.Quantity)),
			UnitAmountFa: formatTomanFa(it.UnitAmount),
			AmountFa:     formatTomanFa(it.Amount),
		})
	}
	// Reflect a succeeded payment's ref/gateway if present.
	for _, p := range inv.Payments {
		if p.Status == domain.PaymentStatusSucceeded {
			if p.RefID != nil {
				vm.RefID = *p.RefID
				vm.RefIDFa = toPersianDigits(*p.RefID)
			}
			vm.GatewayFa = gatewayFa(p.Gateway)
			break
		}
	}
	return vm
}

func renderReceiptHTML(vm receiptVM) ([]byte, error) {
	var buf bytes.Buffer
	if err := receiptTmpl.Execute(&buf, vm); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func itoaInt(n int) string { return toASCIIInt(n) }

func invoiceStatusFa(s domain.InvoiceStatus) string {
	switch s {
	case domain.InvoiceStatusPaid:
		return "پرداخت شده"
	case domain.InvoiceStatusPending:
		return "در انتظار پرداخت"
	case domain.InvoiceStatusRefunded:
		return "مسترد شده"
	case domain.InvoiceStatusCanceled:
		return "لغو شده"
	case domain.InvoiceStatusExpired:
		return "منقضی شده"
	default:
		return "پیش‌نویس"
	}
}

func gatewayFa(g domain.GatewayName) string {
	if g == domain.GatewayManual {
		return "پرداخت دستی"
	}
	return "زرین‌پال"
}
