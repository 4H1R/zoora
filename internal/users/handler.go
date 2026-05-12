package users

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/auth"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var usersListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"username", "name"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "username", "name"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

type Handler struct {
	svc domain.UserService
}

func NewHandler(svc domain.UserService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")

	authed := rg.Group("", authMiddleware)
	{
		// Self-access (no permission needed, just authenticated).
		authed.GET("/users/me", h.GetProfile)
		authed.PUT("/users/me", h.UpdateProfile)
		authed.POST("/users/me/password", h.ChangePassword)

		// CRUD with self-vs-any permission pattern.
		authed.POST("/users", perm(domain.PermUsersCreate), h.CreateUser)
		authed.GET("/users", auth.RequireAnyPermission(domain.PermUsersView, domain.PermUsersViewAny), h.ListUsers)
		authed.GET("/users/:id", auth.RequireSelfOrPermission(domain.PermUsersView, domain.PermUsersViewAny, "id"), idParam, h.GetUserByID)
		authed.PUT("/users/:id", auth.RequireSelfOrPermission(domain.PermUsersUpdate, domain.PermUsersUpdateAny, "id"), idParam, h.UpdateUser)
		authed.DELETE("/users/:id", auth.RequireSelfOrPermission(domain.PermUsersDelete, domain.PermUsersDeleteAny, "id"), idParam, h.DeleteUser)

		// Role assignment on a user.
		authed.PUT("/users/:id/role", perm(domain.PermRolesUpdate), idParam, h.AssignRole)
		authed.DELETE("/users/:id/role", perm(domain.PermRolesUpdate), idParam, h.RemoveRole)
	}
}

// CreateUser creates a new user.
// @Summary Create user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateUserDTO true "User data"
// @Success 201 {object} domain.Response{data=domain.User}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /users [post]
func (h *Handler) CreateUser(c *gin.Context) {
	var dto domain.CreateUserDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	user, err := h.svc.Create(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, user)
}

// GetUserByID returns a user by ID.
// @Summary Get user by ID
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Success 200 {object} domain.Response{data=domain.User}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /users/{id} [get]
func (h *Handler) GetUserByID(c *gin.Context) {
	user, err := h.svc.GetByID(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, user)
}

// ListUsers returns a paginated list of users.
// @Summary List users
// @Description Returns users scoped by caller role. Search matches substrings of: username, name. Orderable fields: created_at, updated_at, username, name.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param search query string false "Substring match on username/name"
// @Param order_by query string false "One of: created_at, updated_at, username, name"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Param page_size query int false "Items per page (default 20)"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.User}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	p := listparams.Bind(c, usersListConfig)
	users, total, err := h.svc.List(c.Request.Context(), p)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(users, total, p))
}

// UpdateUser updates a user by ID.
// @Summary Update user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Param body body domain.UpdateUserDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.User}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /users/{id} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	var dto domain.UpdateUserDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	user, err := h.svc.Update(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, user)
}

// DeleteUser soft-deletes a user by ID.
// @Summary Delete user
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /users/{id} [delete]
func (h *Handler) DeleteUser(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// GetProfile returns the authenticated user's profile.
// @Summary Get my profile
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=domain.User}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /users/me [get]
func (h *Handler) GetProfile(c *gin.Context) {
	user, err := h.svc.GetProfile(c.Request.Context(), auth.GetUserID(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, user)
}

// UpdateProfile updates the authenticated user's profile.
// @Summary Update my profile
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.UpdateProfileDTO true "Profile data"
// @Success 200 {object} domain.Response{data=domain.User}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /users/me [put]
func (h *Handler) UpdateProfile(c *gin.Context) {
	var dto domain.UpdateProfileDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	user, err := h.svc.UpdateProfile(c.Request.Context(), auth.GetUserID(c), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, user)
}

// ChangePassword changes the authenticated user's password.
// @Summary Change my password
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.ChangePasswordDTO true "Password data"
// @Success 200 {object} domain.Response
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /users/me/password [post]
func (h *Handler) ChangePassword(c *gin.Context) {
	var dto domain.ChangePasswordDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), auth.GetUserID(c), dto); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// AssignRole assigns a role to a user.
// @Summary Assign role to user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Param body body domain.AssignRoleDTO true "Role data"
// @Success 200 {object} domain.Response{data=domain.User}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /users/{id}/role [put]
func (h *Handler) AssignRole(c *gin.Context) {
	var dto domain.AssignRoleDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	user, err := h.svc.AssignRole(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, user)
}

// RemoveRole removes the role from a user.
// @Summary Remove role from user
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Success 200 {object} domain.Response{data=domain.User}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /users/{id}/role [delete]
func (h *Handler) RemoveRole(c *gin.Context) {
	user, err := h.svc.RemoveRole(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, user)
}
