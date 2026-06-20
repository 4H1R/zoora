package attendance

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

// attendanceListConfig is the handler-owned white-list for GET
// /classes/:id/sessions/:sessionId/attendance. Only columns in these slices
// can be searched/ordered by the client; anything else silently falls back
// to defaults.
var attendanceListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"remarks"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "status"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

type Handler struct {
	svc domain.AttendanceService
}

func NewHandler(svc domain.AttendanceService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")
	sessionIDParam := httpx.RequireUUIDParam("sessionId")
	attendanceIDParam := httpx.RequireUUIDParam("attendanceId")

	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/attendance/me", perm(domain.PermAttendanceViewOwn), h.ListMine)
		authed.GET("/classes/:id/sessions/:sessionId/attendance", perm(domain.PermAttendanceView), idParam, sessionIDParam, h.List)
		authed.POST("/classes/:id/sessions/:sessionId/attendance", perm(domain.PermAttendanceCreate), idParam, sessionIDParam, h.Mark)
		authed.POST("/classes/:id/sessions/:sessionId/attendance/bulk", perm(domain.PermAttendanceCreate), idParam, sessionIDParam, h.BulkMark)
		authed.POST("/classes/:id/sessions/:sessionId/attendance/auto-mark", perm(domain.PermAttendanceCreate), idParam, sessionIDParam, h.AutoMark)
		authed.GET("/attendance/:id", perm(domain.PermAttendanceView), idParam, h.Get)
		authed.PUT("/attendance/:attendanceId", perm(domain.PermAttendanceUpdate), attendanceIDParam, h.Update)
		authed.DELETE("/attendance/:attendanceId", perm(domain.PermAttendanceDelete), attendanceIDParam, h.Delete)
	}
}

// ListMine returns the caller's own attendance history + summary across classes.
// @Summary List my attendance
// @Tags Attendance
// @Produce json
// @Security BearerAuth
// @Param order_by query string false "One of: created_at, updated_at, status"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.MyAttendance}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Router /attendance/me [get]
func (h *Handler) ListMine(c *gin.Context) {
	p := listparams.Bind(c, attendanceListConfig)
	res, err := h.svc.ListMine(c.Request.Context(), p)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, res)
}

// List returns attendance records for a session.
// @Summary List attendance by session
// @Description Returns attendance records for a class session. Teachers/admins see all; students see only their own record. Search matches substrings of: remarks. Filters: status, user_id, is_auto_marked. Orderable fields: created_at, updated_at, status.
// @Tags Attendance
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Param sessionId path string true "Session UUID"
// @Param status query string false "Filter by status" Enums(present,absent,late,excused)
// @Param user_id query string false "Filter by user UUID"
// @Param is_auto_marked query bool false "Filter auto-marked vs manual"
// @Param search query string false "Substring match on remarks"
// @Param order_by query string false "One of: created_at, updated_at, status"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Attendance}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/{id}/sessions/{sessionId}/attendance [get]
func (h *Handler) List(c *gin.Context) {
	var q domain.ListAttendanceQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{"user_id": &q.UserID}); err != nil {
		_ = c.Error(err)
		return
	}
	q.ListParams = listparams.Bind(c, attendanceListConfig)
	items, total, err := h.svc.ListBySession(
		c.Request.Context(),
		httpx.UUIDParam(c, "id"),
		httpx.UUIDParam(c, "sessionId"),
		q,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(items, total, q.ListParams))
}

// Mark creates a single attendance record for a student in a session.
// @Summary Mark attendance
// @Tags Attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Param sessionId path string true "Session UUID"
// @Param body body domain.CreateAttendanceDTO true "Attendance data"
// @Success 201 {object} domain.Response{data=domain.Attendance}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody} "Already marked"
// @Router /classes/{id}/sessions/{sessionId}/attendance [post]
func (h *Handler) Mark(c *gin.Context) {
	var dto domain.CreateAttendanceDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	a, err := h.svc.Mark(
		c.Request.Context(),
		httpx.UUIDParam(c, "id"),
		httpx.UUIDParam(c, "sessionId"),
		dto,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, a)
}

// BulkMark creates attendance records for multiple students in a session.
// @Summary Bulk mark attendance
// @Tags Attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Param sessionId path string true "Session UUID"
// @Param body body domain.BulkCreateAttendanceDTO true "Bulk attendance data"
// @Success 201 {object} domain.Response{data=[]domain.Attendance}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody} "Duplicate entry"
// @Router /classes/{id}/sessions/{sessionId}/attendance/bulk [post]
func (h *Handler) BulkMark(c *gin.Context) {
	var dto domain.BulkCreateAttendanceDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	results, err := h.svc.BulkMark(
		c.Request.Context(),
		httpx.UUIDParam(c, "id"),
		httpx.UUIDParam(c, "sessionId"),
		dto,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, results)
}

// AutoMark triggers automated attendance marking based on live or offline room activity.
// @Summary Auto-mark attendance from room activity
// @Description Teacher triggers auto-marking. For live rooms: students with total_duration >= min_duration_seconds are marked present. For offline rooms: students who viewed the content are marked present. Students already marked are skipped.
// @Tags Attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Param sessionId path string true "Session UUID"
// @Param body body domain.AutoMarkAttendanceDTO true "Auto-mark config"
// @Success 200 {object} domain.Response{data=domain.AutoMarkResult}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/{id}/sessions/{sessionId}/attendance/auto-mark [post]
func (h *Handler) AutoMark(c *gin.Context) {
	var dto domain.AutoMarkAttendanceDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	result, err := h.svc.AutoMark(
		c.Request.Context(),
		httpx.UUIDParam(c, "id"),
		httpx.UUIDParam(c, "sessionId"),
		dto,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, result)
}

// Get returns a single attendance record.
// @Summary Get attendance record
// @Tags Attendance
// @Produce json
// @Security BearerAuth
// @Param id path string true "Attendance UUID"
// @Success 200 {object} domain.Response{data=domain.Attendance}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /attendance/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	a, err := h.svc.GetByID(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, a)
}

// Update modifies an attendance record.
// @Summary Update attendance
// @Tags Attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param attendanceId path string true "Attendance UUID"
// @Param body body domain.UpdateAttendanceDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.Attendance}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /attendance/{attendanceId} [put]
func (h *Handler) Update(c *gin.Context) {
	var dto domain.UpdateAttendanceDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	a, err := h.svc.Update(c.Request.Context(), httpx.UUIDParam(c, "attendanceId"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, a)
}

// Delete removes an attendance record.
// @Summary Delete attendance
// @Tags Attendance
// @Produce json
// @Security BearerAuth
// @Param attendanceId path string true "Attendance UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /attendance/{attendanceId} [delete]
func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), httpx.UUIDParam(c, "attendanceId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
