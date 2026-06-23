package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

// adminUsersListConfig is the white-list that gates what the admin-users list
// endpoint allows for search and ordering. Anything outside these slices is
// silently ignored and falls back to the defaults.
var adminUsersListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"username", "name"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "username", "name"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

// AdminHandler registers under /api/v1/admin. The admin group is already
// guarded by auth middleware + RequireAdmin, so handlers here only do I/O:
// bind input, call service, attach error. All authorization, scoping, and
// business logic lives in the service layer.
type AdminHandler struct {
	svc     domain.UserService
	authSvc domain.AuthService
}

func NewAdminHandler(svc domain.UserService, authSvc domain.AuthService) *AdminHandler {
	return &AdminHandler{svc: svc, authSvc: authSvc}
}

func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	idParam := httpx.RequireUUIDParam("id")

	group.GET("/users", h.List)
	group.POST("/users", h.Create)
	group.GET("/users/:id", idParam, h.Get)
	group.PUT("/users/:id", idParam, h.Update)
	group.DELETE("/users/:id", idParam, h.HardDelete)
	group.POST("/users/:id/password", idParam, h.ForceResetPassword)
	group.POST("/users/:id/revoke-sessions", idParam, h.RevokeSessions)
	group.POST("/users/:id/disable", idParam, h.Disable)
	group.POST("/users/:id/enable", idParam, h.Enable)
}

// List returns a filtered, paginated list of users across all organizations.
// @Summary [Admin] List users
// @Tags Admin/Users
// @Produce json
// @Security BearerAuth
// @Param organization_id query string false "Filter by organization UUID"
// @Param role_id query string false "Filter by role UUID"
// @Param is_admin query bool false "Filter by admin flag"
// @Param disabled query bool false "Filter by disabled state"
// @Param include_deleted query bool false "Include soft-deleted users"
// @Param search query string false "Substring match on username/name"
// @Param order_by query string false "One of: created_at, updated_at, username, name"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Param page_size query int false "Items per page (default 20)"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.User}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/users [get]
func (h *AdminHandler) List(c *gin.Context) {
	var q domain.AdminListUsersQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{
		"organization_id": &q.OrganizationID,
		"role_id":         &q.RoleID,
	}); err != nil {
		_ = c.Error(err)
		return
	}
	q.ListParams = listparams.Bind(c, adminUsersListConfig)
	users, total, err := h.svc.AdminList(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(users, total, q.ListParams))
}

// Create creates a user in any organization; may set admin/staff flags.
// @Summary [Admin] Create user
// @Tags Admin/Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.AdminCreateUserDTO true "User data"
// @Success 201 {object} domain.Response{data=domain.User}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/users [post]
func (h *AdminHandler) Create(c *gin.Context) {
	var dto domain.AdminCreateUserDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	user, err := h.svc.AdminCreate(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, user)
}

// Get returns a user by ID, bypassing organization scoping.
// @Summary [Admin] Get user
// @Tags Admin/Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Success 200 {object} domain.Response{data=domain.User}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/users/{id} [get]
func (h *AdminHandler) Get(c *gin.Context) {
	user, err := h.svc.AdminGetByID(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, user)
}

// Update updates any user field including admin/staff flags.
// @Summary [Admin] Update user
// @Tags Admin/Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Param body body domain.AdminUpdateUserDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.User}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/users/{id} [put]
func (h *AdminHandler) Update(c *gin.Context) {
	var dto domain.AdminUpdateUserDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	user, err := h.svc.AdminUpdate(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, user)
}

// @Summary [Admin] Hard-delete user
// @Tags Admin/Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/users/{id} [delete]
func (h *AdminHandler) HardDelete(c *gin.Context) {
	if err := h.svc.AdminHardDelete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// ForceResetPassword sets a new password without requiring the current one.
// @Summary [Admin] Force-reset user password
// @Tags Admin/Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Param body body domain.AdminForceResetPasswordDTO true "New password"
// @Success 200 {object} domain.Response
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/users/{id}/password [post]
func (h *AdminHandler) ForceResetPassword(c *gin.Context) {
	var dto domain.AdminForceResetPasswordDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.svc.AdminForceResetPassword(c.Request.Context(), httpx.UUIDParam(c, "id"), dto); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// @Summary [Admin] Disable user
// @Tags Admin/Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Param body body domain.DisableUserDTO true "Disable reason"
// @Success 200 {object} domain.Response{data=domain.User}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/users/{id}/disable [post]
func (h *AdminHandler) Disable(c *gin.Context) {
	var dto domain.DisableUserDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	user, err := h.svc.Disable(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, user)
}

// @Summary [Admin] Enable user
// @Tags Admin/Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Success 200 {object} domain.Response{data=domain.User}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/users/{id}/enable [post]
func (h *AdminHandler) Enable(c *gin.Context) {
	user, err := h.svc.Enable(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, user)
}

// RevokeSessions invalidates all outstanding JWTs for the target user.
// @Summary [Admin] Revoke all sessions for a user
// @Tags Admin/Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/users/{id}/revoke-sessions [post]
func (h *AdminHandler) RevokeSessions(c *gin.Context) {
	if err := h.authSvc.AdminRevokeSessions(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
