package roles

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

type Handler struct {
	svc      domain.RoleService
	permRepo domain.PermissionRepository
}

func NewHandler(svc domain.RoleService, permRepo domain.PermissionRepository) *Handler {
	return &Handler{svc: svc, permRepo: permRepo}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")

	authed := rg.Group("", authMiddleware)
	{
		authed.POST("/roles", perm(domain.PermRolesCreate), h.CreateRole)
		authed.GET("/roles", h.ListRoles)
		authed.GET("/roles/stats", perm(domain.PermRolesView), h.RoleStats)
		authed.GET("/roles/:id", perm(domain.PermRolesView), idParam, h.GetRoleByID)
		authed.PUT("/roles/:id", perm(domain.PermRolesUpdate), idParam, h.UpdateRole)
		authed.DELETE("/roles/:id", perm(domain.PermRolesDelete), idParam, h.DeleteRole)
		authed.GET("/permissions", h.ListPermissions)
	}
}

// CreateRole creates a new role in the caller's organization.
// @Summary Create role
// @Tags Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateRoleDTO true "Role data"
// @Success 201 {object} domain.Response{data=domain.Role}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /roles [post]
func (h *Handler) CreateRole(c *gin.Context) {
	var dto domain.CreateRoleDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}

	role, err := h.svc.Create(c.Request.Context(), dto)
	if err != nil {
		domain.ErrorResponse(c, err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, role)
}

// GetRoleByID returns a role by ID.
// @Summary Get role by ID
// @Tags Roles
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role UUID"
// @Success 200 {object} domain.Response{data=domain.Role}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /roles/{id} [get]
func (h *Handler) GetRoleByID(c *gin.Context) {
	role, err := h.svc.GetByID(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		domain.ErrorResponse(c, err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, role)
}

// UpdateRole updates a role by ID.
// @Summary Update role
// @Tags Roles
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
// @Router /roles/{id} [put]
func (h *Handler) UpdateRole(c *gin.Context) {
	var dto domain.UpdateRoleDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	role, err := h.svc.Update(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		domain.ErrorResponse(c, err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, role)
}

// DeleteRole soft-deletes a role by ID.
// @Summary Delete role
// @Tags Roles
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /roles/{id} [delete]
func (h *Handler) DeleteRole(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		domain.ErrorResponse(c, err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// ListRoles returns all roles for the caller's organization.
// @Summary List roles
// @Tags Roles
// @Produce json
// @Security BearerAuth
// @Param organization_id query string false "Filter by organization UUID"
// @Success 200 {object} domain.Response{data=[]domain.Role}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /roles [get]
func (h *Handler) ListRoles(c *gin.Context) {
	var filter domain.RoleFilter
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		if id, err := uuid.Parse(orgIDStr); err == nil {
			filter.OrganizationID = &id
		}
	}
	roleList, err := h.svc.List(c.Request.Context(), filter)
	if err != nil {
		domain.ErrorResponse(c, err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, roleList)
}

// RoleStats returns aggregate stats for roles.
// @Summary Get role stats
// @Tags Roles
// @Produce json
// @Security BearerAuth
// @Param organization_id query string false "Filter by organization UUID"
// @Success 200 {object} domain.Response{data=domain.RoleStats}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /roles/stats [get]
func (h *Handler) RoleStats(c *gin.Context) {
	var orgID *uuid.UUID
	if orgIDStr := c.Query("organization_id"); orgIDStr != "" {
		if id, err := uuid.Parse(orgIDStr); err == nil {
			orgID = &id
		}
	}
	stats, err := h.svc.Stats(c.Request.Context(), orgID)
	if err != nil {
		domain.ErrorResponse(c, err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, stats)
}

// ListPermissions returns all available permissions.
// @Summary List permissions
// @Tags Roles
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=[]domain.Permission}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /permissions [get]
func (h *Handler) ListPermissions(c *gin.Context) {
	perms, err := h.permRepo.List(c.Request.Context())
	if err != nil {
		domain.ErrorResponse(c, err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, perms)
}
