package classes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

// adminClassesListConfig is the white-list for GET /admin/classes. Anything
// outside these slices is silently ignored and falls back to defaults.
var adminClassesListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"name", "description"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "name"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

// adminSessionsListConfig white-lists search/order for GET /admin/sessions.
var adminSessionsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"name", "description"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "name", "start_time"},
	DefaultOrderBy:      "start_time",
	DefaultOrderDir:     "desc",
}

// AdminHandler registers under /api/v1/admin. The admin group is already
// guarded by auth middleware + RequireAdmin, so this handler only binds
// input, forwards to the service, and attaches errors.
type AdminHandler struct {
	svc domain.ClassService
}

func NewAdminHandler(svc domain.ClassService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	idParam := httpx.RequireUUIDParam("id")
	sessionIDParam := httpx.RequireUUIDParam("sessionId")

	group.GET("/classes", h.List)
	group.POST("/classes", h.Create)
	group.PUT("/classes/:id", idParam, h.Update)
	group.DELETE("/classes/:id", idParam, h.HardDelete)
	group.GET("/sessions", h.ListSessions)
	group.DELETE("/classes/sessions/:sessionId", sessionIDParam, h.HardDeleteSession)
}

// @Summary [Admin] Create class
// @Tags Admin/Classes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.AdminCreateClassDTO true "Class data"
// @Success 201 {object} domain.Response{data=domain.Class}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/classes [post]
func (h *AdminHandler) Create(c *gin.Context) {
	var dto domain.AdminCreateClassDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	class, err := h.svc.AdminCreate(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, class)
}

// @Summary [Admin] Update class
// @Tags Admin/Classes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Param body body domain.AdminUpdateClassDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.Class}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/classes/{id} [put]
func (h *AdminHandler) Update(c *gin.Context) {
	var dto domain.AdminUpdateClassDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	class, err := h.svc.AdminUpdate(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, class)
}

// @Summary [Admin] List classes
// @Description Cross-org list. Search matches substrings of: name, description. Orderable fields: created_at, updated_at, name. Filters: user_id (teacher), include_deleted.
// @Tags Admin/Classes
// @Produce json
// @Security BearerAuth
// @Param user_id query string false "Filter by teacher UUID"
// @Param include_deleted query bool false "Include soft-deleted classes"
// @Param search query string false "Substring match on name/description"
// @Param order_by query string false "One of: created_at, updated_at, name"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Class}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/classes [get]
func (h *AdminHandler) List(c *gin.Context) {
	var q domain.AdminListClassesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{"user_id": &q.UserID}); err != nil {
		_ = c.Error(err)
		return
	}
	q.ListParams = listparams.Bind(c, adminClassesListConfig)
	classes, total, err := h.svc.AdminList(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(classes, total, q.ListParams))
}

// HardDelete permanently deletes a class, bypassing soft-delete.
// @Summary [Admin] Hard-delete class
// @Tags Admin/Classes
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/classes/{id} [delete]
func (h *AdminHandler) HardDelete(c *gin.Context) {
	if err := h.svc.AdminHardDelete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// @Summary [Admin] List class sessions
// @Description Cross-org list. Search matches substrings of: name, description. Orderable fields: created_at, updated_at, name, start_time. Filters: class_id, include_deleted.
// @Tags Admin/Classes
// @Produce json
// @Security BearerAuth
// @Param class_id query string false "Filter by class UUID"
// @Param include_deleted query bool false "Include soft-deleted sessions"
// @Param search query string false "Substring match on name/description"
// @Param order_by query string false "One of: created_at, updated_at, name, start_time"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.ClassSession}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/sessions [get]
func (h *AdminHandler) ListSessions(c *gin.Context) {
	var q domain.AdminListClassSessionsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{"class_id": &q.ClassID}); err != nil {
		_ = c.Error(err)
		return
	}
	q.ListParams = listparams.Bind(c, adminSessionsListConfig)
	sessions, total, err := h.svc.AdminListSessions(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(sessions, total, q.ListParams))
}

// HardDeleteSession permanently deletes a class session, bypassing soft-delete.
// @Summary [Admin] Hard-delete class session
// @Tags Admin/Classes
// @Produce json
// @Security BearerAuth
// @Param sessionId path string true "Session UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/classes/sessions/{sessionId} [delete]
func (h *AdminHandler) HardDeleteSession(c *gin.Context) {
	if err := h.svc.AdminHardDeleteSession(c.Request.Context(), httpx.UUIDParam(c, "sessionId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
