package roles

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var adminRolesListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"name"},
	AllowedOrderFields:  []string{"name", "created_at", "updated_at"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

type AdminHandler struct {
	svc domain.RoleService
}

func NewAdminHandler(svc domain.RoleService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	idParam := httpx.RequireUUIDParam("id")

	group.GET("/roles", h.List)
	group.GET("/roles/stats", h.Stats)
	group.GET("/roles/:id", idParam, h.Get)
	group.POST("/roles", h.Create)
	group.PUT("/roles/:id", idParam, h.Update)
	group.DELETE("/roles/:id", idParam, h.Delete)
}

// List returns a filtered, paginated list of roles.
// @Summary [Admin] List roles
// @Tags Admin/Roles
// @Produce json
// @Security BearerAuth
// @Param organization_id query string false "Filter by organization UUID"
// @Param search query string false "Search by name"
// @Param order_by query string false "One of: name, created_at, updated_at"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Param page_size query int false "Items per page (default 20)"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Role}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/roles [get]
func (h *AdminHandler) List(c *gin.Context) {
	p := listparams.Bind(c, adminRolesListConfig)
	filter := domain.AdminRoleFilter{ListParams: p}
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		if id, err := uuid.Parse(orgIDStr); err == nil {
			filter.OrganizationID = &id
		}
	}
	roleList, total, err := h.svc.AdminList(c.Request.Context(), filter)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(roleList, total, p))
}

// Stats returns aggregate stats for roles.
// @Summary [Admin] Get role stats
// @Tags Admin/Roles
// @Produce json
// @Security BearerAuth
// @Param organization_id query string false "Filter by organization UUID"
// @Success 200 {object} domain.Response{data=domain.RoleStats}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/roles/stats [get]
func (h *AdminHandler) Stats(c *gin.Context) {
	var orgID *uuid.UUID
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		if id, err := uuid.Parse(orgIDStr); err == nil {
			orgID = &id
		}
	}
	stats, err := h.svc.Stats(c.Request.Context(), orgID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, stats)
}

// Get returns a role by ID.
// @Summary [Admin] Get role by ID
// @Tags Admin/Roles
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role UUID"
// @Success 200 {object} domain.Response{data=domain.Role}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/roles/{id} [get]
func (h *AdminHandler) Get(c *gin.Context) {
	role, err := h.svc.GetByID(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, role)
}

// Create creates a new role.
// @Summary [Admin] Create role
// @Tags Admin/Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateRoleDTO true "Role data"
// @Success 201 {object} domain.Response{data=domain.Role}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/roles [post]
func (h *AdminHandler) Create(c *gin.Context) {
	var dto domain.CreateRoleDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	role, err := h.svc.Create(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, role)
}

// Update updates a role by ID.
// @Summary [Admin] Update role
// @Tags Admin/Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role UUID"
// @Param body body domain.UpdateRoleDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.Role}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/roles/{id} [put]
func (h *AdminHandler) Update(c *gin.Context) {
	var dto domain.UpdateRoleDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	role, err := h.svc.Update(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, role)
}

// Delete soft-deletes a role by ID.
// @Summary [Admin] Delete role
// @Tags Admin/Roles
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/roles/{id} [delete]
func (h *AdminHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
