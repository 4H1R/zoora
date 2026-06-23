package classes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

// classesListConfig is the handler-owned white-list for GET /classes. Only
// columns in these slices can be searched/ordered by the client; anything
// else is silently ignored in favour of the defaults.
var classesListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"name", "description"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "name"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

// sessionsListConfig gates GET /classes/:id/sessions.
var sessionsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"name", "description"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "name", "start_time"},
	DefaultOrderBy:      "start_time",
	DefaultOrderDir:     "asc",
}

// membersListConfig gates GET /classes/:id/members. Search/order target the
// member's user (name/username); memberRepository.ListByClass resolves them
// against the users table via a subquery, so tokens stay friendly here.
var membersListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"name", "username"},
	AllowedOrderFields:  []string{"created_at", "name"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

type Handler struct {
	svc domain.ClassService
}

func NewHandler(svc domain.ClassService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")
	sessionIDParam := httpx.RequireUUIDParam("sessionId")
	userIDParam := httpx.RequireUUIDParam("userId")

	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/classes", perm(domain.PermClassesView), h.List)
		authed.POST("/classes", perm(domain.PermClassesCreate), h.Create)
		authed.GET("/classes/:id", perm(domain.PermClassesView), idParam, h.Get)
		authed.PUT("/classes/:id", perm(domain.PermClassesUpdate), idParam, h.Update)
		authed.DELETE("/classes/:id", perm(domain.PermClassesDelete), idParam, h.Delete)

		authed.GET("/classes/:id/sessions", perm(domain.PermClassesView), idParam, h.ListSessions)
		authed.POST("/classes/:id/sessions", perm(domain.PermClassesUpdate), idParam, h.CreateSession)
		authed.GET("/classes/sessions/:sessionId", perm(domain.PermClassesView), sessionIDParam, h.GetSession)
		authed.PUT("/classes/sessions/:sessionId", perm(domain.PermClassesUpdate), sessionIDParam, h.UpdateSession)
		authed.DELETE("/classes/sessions/:sessionId", perm(domain.PermClassesDelete), sessionIDParam, h.DeleteSession)

		authed.GET("/classes/:id/members", perm(domain.PermClassesView), idParam, h.ListMembers)
		authed.POST("/classes/:id/members", perm(domain.PermClassesJoin), idParam, h.Enroll)
		authed.DELETE("/classes/:id/members/:userId", perm(domain.PermClassesJoin), idParam, userIDParam, h.Leave)
	}
}

// @Summary List classes (scoped by RBAC)
// @Description Returns classes filtered by caller role: super-admins see all, org-admins see their organization, teachers see their own classes, students see classes they are enrolled in. Search matches substrings of: name, description. Orderable fields: created_at, updated_at, name.
// @Tags Classes
// @Produce json
// @Security BearerAuth
// @Param search query string false "Substring match on name/description"
// @Param order_by query string false "One of: created_at, updated_at, name"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Class}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes [get]
func (h *Handler) List(c *gin.Context) {
	p := listparams.Bind(c, classesListConfig)
	classes, total, err := h.svc.List(c.Request.Context(), p)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(classes, total, p))
}

// Create creates a class owned by the caller inside their organization.
// @Summary Create class
// @Tags Classes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateClassDTO true "Class data"
// @Success 201 {object} domain.Response{data=domain.Class}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes [post]
func (h *Handler) Create(c *gin.Context) {
	var dto domain.CreateClassDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	class, err := h.svc.Create(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, class)
}

// @Summary Get class
// @Tags Classes
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Success 200 {object} domain.Response{data=domain.Class}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	class, err := h.svc.GetByID(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, class)
}

// @Summary Update class
// @Tags Classes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Param body body domain.UpdateClassDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.Class}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	var dto domain.UpdateClassDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	class, err := h.svc.Update(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, class)
}

// Delete soft-deletes a class.
// @Summary Delete class
// @Tags Classes
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// @Summary List class sessions
// @Description Search matches substrings of: name, description. Orderable fields: created_at, updated_at, name, start_time.
// @Tags Classes
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Param search query string false "Substring match on name/description"
// @Param order_by query string false "One of: created_at, updated_at, name, start_time"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Param page_size query int false "Items per page"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.ClassSession}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/{id}/sessions [get]
func (h *Handler) ListSessions(c *gin.Context) {
	var q domain.ListClassSessionsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	q.ListParams = listparams.Bind(c, sessionsListConfig)
	sessions, total, err := h.svc.ListSessions(c.Request.Context(), httpx.UUIDParam(c, "id"), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(sessions, total, q.ListParams))
}

// @Summary Create class session
// @Tags Classes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Param body body domain.CreateClassSessionDTO true "Session data"
// @Success 201 {object} domain.Response{data=domain.ClassSession}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/{id}/sessions [post]
func (h *Handler) CreateSession(c *gin.Context) {
	var dto domain.CreateClassSessionDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	session, err := h.svc.CreateSession(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, session)
}

// @Summary Get class session
// @Tags Classes
// @Produce json
// @Security BearerAuth
// @Param sessionId path string true "Session UUID"
// @Success 200 {object} domain.Response{data=domain.ClassSession}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/sessions/{sessionId} [get]
func (h *Handler) GetSession(c *gin.Context) {
	session, err := h.svc.GetSession(c.Request.Context(), httpx.UUIDParam(c, "sessionId"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, session)
}

// @Summary Update class session
// @Tags Classes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sessionId path string true "Session UUID"
// @Param body body domain.UpdateClassSessionDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.ClassSession}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/sessions/{sessionId} [put]
func (h *Handler) UpdateSession(c *gin.Context) {
	var dto domain.UpdateClassSessionDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	session, err := h.svc.UpdateSession(c.Request.Context(), httpx.UUIDParam(c, "sessionId"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, session)
}

// DeleteSession soft-deletes a session.
// @Summary Delete class session
// @Tags Classes
// @Produce json
// @Security BearerAuth
// @Param sessionId path string true "Session UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/sessions/{sessionId} [delete]
func (h *Handler) DeleteSession(c *gin.Context) {
	if err := h.svc.DeleteSession(c.Request.Context(), httpx.UUIDParam(c, "sessionId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// ListMembers lists the roster of a class (managers only).
// @Summary List class members
// @Description Search matches substrings of the member's name/username. Orderable fields: created_at, name. Only class managers (teacher, org-admin, super-admin) may view the roster.
// @Tags Classes
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Param search query string false "Substring match on member name/username"
// @Param order_by query string false "One of: created_at, name"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Param page_size query int false "Items per page"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.ClassMember}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/{id}/members [get]
func (h *Handler) ListMembers(c *gin.Context) {
	q := domain.ListClassMembersQuery{ListParams: listparams.Bind(c, membersListConfig)}
	members, total, err := h.svc.ListMembers(c.Request.Context(), httpx.UUIDParam(c, "id"), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(members, total, q.ListParams))
}

// Enroll enrolls a user into a class. Students may only self-enroll.
// @Summary Enroll user in class
// @Tags Classes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Param body body domain.EnrollClassMemberDTO true "User to enroll"
// @Success 201 {object} domain.Response{data=domain.ClassMember}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody} "Capacity reached or user already enrolled"
// @Router /classes/{id}/members [post]
func (h *Handler) Enroll(c *gin.Context) {
	var dto domain.EnrollClassMemberDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	m, err := h.svc.Enroll(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, m)
}

// Leave removes a user from a class. Users may self-leave; managers may remove anyone.
// @Summary Remove user from class
// @Tags Classes
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Param userId path string true "User UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/{id}/members/{userId} [delete]
func (h *Handler) Leave(c *gin.Context) {
	if err := h.svc.Leave(c.Request.Context(), httpx.UUIDParam(c, "id"), httpx.UUIDParam(c, "userId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
