package organizations

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var adminOrgsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"name", "description"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "name"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

type AdminHandler struct {
	svc domain.OrganizationService
}

func NewAdminHandler(svc domain.OrganizationService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	idParam := httpx.RequireUUIDParam("id")

	group.GET("/organizations", h.List)
	group.GET("/organizations/stats", h.Stats)
	group.GET("/plans", h.Plans)
	group.POST("/organizations", h.Create)
	group.GET("/organizations/:id", idParam, h.Get)
	group.PUT("/organizations/:id", idParam, h.Update)
	group.PUT("/organizations/:id/plan", idParam, h.SetPlan)
	group.DELETE("/organizations/:id", idParam, h.HardDelete)
	group.POST("/organizations/:id/restore", idParam, h.Restore)
}

// Stats returns organization aggregate statistics.
// @Summary [Admin] Organization stats
// @Tags Admin/Organizations
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=domain.OrganizationStats}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/organizations/stats [get]
func (h *AdminHandler) Stats(c *gin.Context) {
	stats, err := h.svc.GetStats(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, stats)
}

// @Summary [Admin] List organizations
// @Description Cross-org list. Search matches substrings of: name, description. Orderable fields: created_at, updated_at, name. Filters: status, include_deleted.
// @Tags Admin/Organizations
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status (active, trial, suspended, archived)"
// @Param include_deleted query bool false "Include soft-deleted organizations"
// @Param search query string false "Substring match on name/description"
// @Param order_by query string false "One of: created_at, updated_at, name"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Param page_size query int false "Items per page (default 20)"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Organization}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/organizations [get]
func (h *AdminHandler) List(c *gin.Context) {
	var q domain.AdminListOrganizationsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	q.ListParams = listparams.Bind(c, adminOrgsListConfig)
	orgs, total, err := h.svc.AdminList(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(orgs, total, q.ListParams))
}

// @Summary [Admin] Create organization
// @Tags Admin/Organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.AdminCreateOrganizationDTO true "Organization data"
// @Success 201 {object} domain.Response{data=domain.Organization}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/organizations [post]
func (h *AdminHandler) Create(c *gin.Context) {
	var dto domain.AdminCreateOrganizationDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	org, err := h.svc.AdminCreate(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, org)
}

// @Summary [Admin] Get organization
// @Tags Admin/Organizations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization UUID"
// @Success 200 {object} domain.Response{data=domain.Organization}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/organizations/{id} [get]
func (h *AdminHandler) Get(c *gin.Context) {
	org, err := h.svc.GetByID(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, org)
}

// @Summary [Admin] Update organization
// @Tags Admin/Organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization UUID"
// @Param body body domain.AdminUpdateOrganizationDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.Organization}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/organizations/{id} [put]
func (h *AdminHandler) Update(c *gin.Context) {
	var dto domain.AdminUpdateOrganizationDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	org, err := h.svc.AdminUpdate(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, org)
}

// Plans returns the static plan catalog (tiers, features, limits) for the admin
// plan picker.
// @Summary [Admin] Plan catalog
// @Tags Admin/Organizations
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=[]domain.PlanInfo}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/plans [get]
func (h *AdminHandler) Plans(c *gin.Context) {
	domain.SuccessResponse(c, http.StatusOK, domain.PublicCatalog())
}

// SetPlan assigns a subscription plan and optional expiry to an organization.
// @Summary [Admin] Set organization plan
// @Description Assign a subscription plan (free/pro/enterprise) and optional expiry. Omit expires_at for a perpetual plan; an expired plan downgrades to free.
// @Tags Admin/Organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID"
// @Param request body domain.SetPlanDTO true "Plan assignment"
// @Success 200 {object} domain.Response{data=domain.Organization}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/organizations/{id}/plan [put]
func (h *AdminHandler) SetPlan(c *gin.Context) {
	var dto domain.SetPlanDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	org, err := h.svc.SetPlan(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, org)
}

// HardDelete permanently deletes an organization, bypassing soft-delete.
// @Summary [Admin] Hard-delete organization
// @Tags Admin/Organizations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/organizations/{id} [delete]
func (h *AdminHandler) HardDelete(c *gin.Context) {
	if err := h.svc.AdminHardDelete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// Restore re-activates a soft-deleted organization.
// @Summary [Admin] Restore organization
// @Tags Admin/Organizations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/organizations/{id}/restore [post]
func (h *AdminHandler) Restore(c *gin.Context) {
	if err := h.svc.AdminRestore(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
