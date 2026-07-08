package billing

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

type AdminHandler struct {
	svc domain.BillingAdminService
}

func NewAdminHandler(svc domain.BillingAdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// RegisterAdminRoutes mounts under the admin group (admin auth + RequireAdmin
// already applied by the caller).
func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	idParam := httpx.RequireUUIDParam("id")

	group.GET("/billing/prices", h.ListPrices)
	group.PUT("/billing/prices", h.UpsertPrice)
	group.DELETE("/billing/prices/:id", idParam, h.DeactivatePrice)

	group.GET("/billing/invoices", h.ListInvoices)
	group.POST("/billing/invoices", h.CreateInvoice)
	group.GET("/billing/invoices/:id", idParam, h.GetInvoice)
	group.POST("/billing/invoices/:id/issue", idParam, h.IssueInvoice)
	group.POST("/billing/invoices/:id/mark-paid", idParam, h.MarkPaid)
	group.POST("/billing/invoices/:id/cancel", idParam, h.CancelInvoice)
	group.POST("/billing/invoices/:id/refund", idParam, h.RefundInvoice)
}

// ListPrices returns all plan prices (including inactive).
// @Summary [Admin] List plan prices
// @Tags Admin/Billing
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=[]domain.PlanPrice}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/billing/prices [get]
func (h *AdminHandler) ListPrices(c *gin.Context) {
	prices, err := h.svc.ListPrices(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, prices)
}

// UpsertPrice creates or updates a plan price.
// @Summary [Admin] Upsert plan price
// @Tags Admin/Billing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.UpsertPlanPriceDTO true "Price"
// @Success 200 {object} domain.Response{data=domain.PlanPrice}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/billing/prices [put]
func (h *AdminHandler) UpsertPrice(c *gin.Context) {
	var dto domain.UpsertPlanPriceDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	p, err := h.svc.UpsertPrice(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, p)
}

// DeactivatePrice marks a plan price inactive.
// @Summary [Admin] Deactivate plan price
// @Tags Admin/Billing
// @Produce json
// @Security BearerAuth
// @Param id path string true "Price ID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/billing/prices/{id} [delete]
func (h *AdminHandler) DeactivatePrice(c *gin.Context) {
	if err := h.svc.DeactivatePrice(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, gin.H{"deactivated": true})
}

// ListInvoices returns invoices across orgs, optionally filtered by organization.
// @Summary [Admin] List invoices
// @Tags Admin/Billing
// @Produce json
// @Security BearerAuth
// @Param organization_id query string false "Filter by organization UUID"
// @Param status query string false "Filter: draft,pending,paid,canceled,expired,refunded"
// @Param page query int false "1-based page"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Invoice}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/billing/invoices [get]
func (h *AdminHandler) ListInvoices(c *gin.Context) {
	var q domain.AdminListInvoicesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{"organization_id": &q.OrganizationID}); err != nil {
		_ = c.Error(err)
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

// CreateInvoice creates a custom draft invoice for an organization.
// @Summary [Admin] Create custom invoice (draft)
// @Tags Admin/Billing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.AdminCreateInvoiceDTO true "Invoice"
// @Success 201 {object} domain.Response{data=domain.Invoice}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/billing/invoices [post]
func (h *AdminHandler) CreateInvoice(c *gin.Context) {
	var dto domain.AdminCreateInvoiceDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	inv, err := h.svc.CreateInvoice(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, inv)
}

// GetInvoice returns one invoice by ID (cross-org).
// @Summary [Admin] Get invoice
// @Tags Admin/Billing
// @Produce json
// @Security BearerAuth
// @Param id path string true "Invoice ID"
// @Success 200 {object} domain.Response{data=domain.Invoice}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/billing/invoices/{id} [get]
func (h *AdminHandler) GetInvoice(c *gin.Context) {
	inv, err := h.svc.GetInvoice(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, inv)
}

// IssueInvoice transitions a draft invoice to pending (issued).
// @Summary [Admin] Issue invoice
// @Tags Admin/Billing
// @Produce json
// @Security BearerAuth
// @Param id path string true "Invoice ID"
// @Success 200 {object} domain.Response{data=domain.Invoice}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/billing/invoices/{id}/issue [post]
func (h *AdminHandler) IssueInvoice(c *gin.Context) {
	inv, err := h.svc.IssueInvoice(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, inv)
}

// MarkPaid records an offline/manual payment and activates the plan.
// @Summary [Admin] Mark invoice paid (offline)
// @Tags Admin/Billing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Invoice ID"
// @Param body body domain.AdminMarkPaidDTO true "Manual payment"
// @Success 200 {object} domain.Response{data=domain.Invoice}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/billing/invoices/{id}/mark-paid [post]
func (h *AdminHandler) MarkPaid(c *gin.Context) {
	var dto domain.AdminMarkPaidDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	inv, err := h.svc.MarkPaid(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, inv)
}

// CancelInvoice cancels a pending/draft invoice.
// @Summary [Admin] Cancel invoice
// @Tags Admin/Billing
// @Produce json
// @Security BearerAuth
// @Param id path string true "Invoice ID"
// @Success 200 {object} domain.Response{data=domain.Invoice}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/billing/invoices/{id}/cancel [post]
func (h *AdminHandler) CancelInvoice(c *gin.Context) {
	inv, err := h.svc.CancelInvoice(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, inv)
}

// RefundInvoice refunds a paid invoice.
// @Summary [Admin] Refund invoice
// @Tags Admin/Billing
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Invoice ID"
// @Param body body domain.AdminRefundDTO true "Refund"
// @Success 200 {object} domain.Response{data=domain.Invoice}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/billing/invoices/{id}/refund [post]
func (h *AdminHandler) RefundInvoice(c *gin.Context) {
	var dto domain.AdminRefundDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	inv, err := h.svc.RefundInvoice(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, inv)
}
