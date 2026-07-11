package leads

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var adminLeadsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"name", "org_name", "phone"},
	AllowedOrderFields:  []string{"created_at", "updated_at"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

type AdminHandler struct {
	svc domain.LeadService
}

func NewAdminHandler(svc domain.LeadService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	idParam := httpx.RequireUUIDParam("id")

	group.GET("/leads", h.List)
	group.PATCH("/leads/:id/status", idParam, h.UpdateStatus)
	group.POST("/leads/:id/convert", idParam, h.Convert)
	group.DELETE("/leads/:id", idParam, h.HardDelete)
}

// List returns leads for the admin pipeline.
// @Summary [Admin] List leads
// @Description Cross-org sales-lead list. Filter by status; newest first. Search matches name/org_name/phone.
// @Tags Admin/Leads
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (new, contacted, converted, rejected)"
// @Param search query string false "Substring match on name/org_name/phone"
// @Param order_by query string false "One of: created_at, updated_at"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Param page_size query int false "Items per page (default 20)"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Lead}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/leads [get]
func (h *AdminHandler) List(c *gin.Context) {
	var q domain.AdminListLeadsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	q.ListParams = listparams.Bind(c, adminLeadsListConfig)
	leads, total, err := h.svc.AdminList(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(leads, total, q.ListParams))
}

// UpdateStatus moves a lead through the pipeline.
// @Summary [Admin] Update lead status
// @Tags Admin/Leads
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lead UUID"
// @Param body body domain.UpdateLeadStatusDTO true "New status"
// @Success 200 {object} domain.Response{data=domain.Lead}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/leads/{id}/status [patch]
func (h *AdminHandler) UpdateStatus(c *gin.Context) {
	var dto domain.UpdateLeadStatusDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	lead, err := h.svc.UpdateStatus(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, lead)
}

// Convert provisions an org + owner account from a lead atomically.
// @Summary [Admin] Convert lead to org
// @Description Creates the organization (+ default settings) and an owner user with the Manager preset role, marks the lead converted, and links the new org. All-or-nothing.
// @Tags Admin/Leads
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lead UUID"
// @Param body body domain.ConvertLeadDTO true "Org + owner details"
// @Success 200 {object} domain.Response{data=domain.Lead}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Failure 422 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/leads/{id}/convert [post]
func (h *AdminHandler) Convert(c *gin.Context) {
	var dto domain.ConvertLeadDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	lead, err := h.svc.Convert(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, lead)
}

// HardDelete permanently removes a lead.
// @Summary [Admin] Delete lead
// @Tags Admin/Leads
// @Produce json
// @Security BearerAuth
// @Param id path string true "Lead UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/leads/{id} [delete]
func (h *AdminHandler) HardDelete(c *gin.Context) {
	if err := h.svc.AdminHardDelete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
