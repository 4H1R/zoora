package billing

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var invoiceListConfig = domain.ListConfig{
	AllowedOrderFields: []string{"created_at", "issued_at"},
	DefaultOrderBy:     "created_at",
	DefaultOrderDir:    "desc",
}

type Handler struct {
	svc domain.BillingService
	// appURLTemplate is the tenant-facing URL template containing "{slug}"; the
	// callback substitutes the org slug so the browser returns to its own host.
	appURLTemplate string
}

func NewHandler(svc domain.BillingService, appURLTemplate string) *Handler {
	return &Handler{svc: svc, appURLTemplate: appURLTemplate}
}

// RegisterRoutes wires org billing routes (auth + billing:manage) plus the
// PUBLIC gateway callback (no auth — the gateway redirects the browser here).
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")

	// Public callback: GET /billing/callback/:gateway?Authority=&Status=
	rg.GET("/billing/callback/:gateway", h.Callback)

	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/billing/plans", perm(domain.PermBillingManage), h.ListPlans)
		authed.POST("/billing/checkout", perm(domain.PermBillingManage), h.Checkout)
		authed.GET("/billing/invoices", perm(domain.PermBillingManage), h.ListInvoices)
		authed.GET("/billing/invoices/:id", perm(domain.PermBillingManage), idParam, h.GetInvoice)
		authed.GET("/billing/invoices/:id/receipt", perm(domain.PermBillingManage), idParam, h.Receipt)
	}
}

// ListPlans returns active plan prices for the checkout picker.
// @Summary List plan prices
// @Tags Billing
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=[]domain.PlanPrice}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /billing/plans [get]
func (h *Handler) ListPlans(c *gin.Context) {
	prices, err := h.svc.ListPlanPrices(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, prices)
}

// Checkout creates a pending invoice and returns a gateway redirect URL.
// @Summary Start checkout
// @Tags Billing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CheckoutDTO true "Checkout"
// @Success 201 {object} domain.Response{data=domain.CheckoutResult}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Router /billing/checkout [post]
func (h *Handler) Checkout(c *gin.Context) {
	var dto domain.CheckoutDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	res, err := h.svc.Checkout(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, res)
}

// ListInvoices returns the caller org's invoices (payment history).
// @Summary List invoices
// @Tags Billing
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter: draft,pending,paid,canceled,expired,refunded"
// @Param page query int false "1-based page"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Invoice}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /billing/invoices [get]
func (h *Handler) ListInvoices(c *gin.Context) {
	var q domain.ListInvoicesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	q.ListParams = listparams.Bind(c, invoiceListConfig)
	items, total, err := h.svc.ListInvoices(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(items, total, q.ListParams))
}

// GetInvoice returns one invoice (tenant-scoped).
// @Summary Get invoice
// @Tags Billing
// @Produce json
// @Security BearerAuth
// @Param id path string true "Invoice ID"
// @Success 200 {object} domain.Response{data=domain.Invoice}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /billing/invoices/{id} [get]
func (h *Handler) GetInvoice(c *gin.Context) {
	inv, err := h.svc.GetInvoice(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, inv)
}

// Receipt returns a short-lived presigned URL to the invoice PDF.
// @Summary Get invoice receipt URL
// @Tags Billing
// @Produce json
// @Security BearerAuth
// @Param id path string true "Invoice ID"
// @Success 200 {object} domain.Response{data=map[string]string}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /billing/invoices/{id}/receipt [get]
func (h *Handler) Receipt(c *gin.Context) {
	url, err := h.svc.InvoicePDFURL(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, gin.H{"url": url})
}

// Callback is the public gateway return endpoint. It settles the payment and
// redirects the browser to the frontend result page.
func (h *Handler) Callback(c *gin.Context) {
	gateway := domain.GatewayName(c.Param("gateway"))
	authority := c.Query("Authority")
	status := c.Query("Status") // Zarinpal: "OK" | "NOK"
	// slug identifies the org whose subdomain we redirect back to; set on the
	// callback URL at checkout so it survives even when the invoice lookup fails.
	// Validate it against the strict slug charset before it can reach the redirect
	// host, otherwise a crafted value would make this public endpoint an open redirect.
	slug := c.Query("org")
	if domain.ValidateSlug(slug) != nil {
		c.String(http.StatusBadRequest, "invalid organization")
		return
	}
	inv, err := h.svc.HandleCallback(c.Request.Context(), gateway, authority, status == "OK")
	// Redirect regardless — the result page reads the invoice status.
	if err != nil {
		c.Redirect(http.StatusFound, h.resultURL(slug, "error", nil))
		return
	}
	outcome := "failed"
	if inv.Status == domain.InvoiceStatusPaid {
		outcome = "success"
	}
	c.Redirect(http.StatusFound, h.resultURL(slug, outcome, inv))
}

func (h *Handler) resultURL(slug, outcome string, inv *domain.Invoice) string {
	base := strings.ReplaceAll(h.appURLTemplate, "{slug}", slug) + "/org/billing/result?status=" + outcome
	if inv != nil {
		base += "&invoice=" + inv.ID.String()
	}
	return base
}
