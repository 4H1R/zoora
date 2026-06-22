package practices

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var roomsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"title", "content"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "start_time", "end_time", "title"},
	DefaultOrderBy:      "start_time",
	DefaultOrderDir:     "desc",
}

var submissionsListConfig = domain.ListConfig{
	AllowedOrderFields: []string{"submitted_at", "created_at", "score"},
	DefaultOrderBy:     "submitted_at",
	DefaultOrderDir:    "desc",
}

type Handler struct {
	svc domain.PracticeService
}

func NewHandler(svc domain.PracticeService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")
	submissionIDParam := httpx.RequireUUIDParam("submissionId")

	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/practices", perm(domain.PermPracticesView), h.ListRooms)
		authed.POST("/practices", perm(domain.PermPracticesCreate), h.CreateRoom)
		authed.GET("/practices/:id", perm(domain.PermPracticesView), idParam, h.GetRoom)
		authed.PUT("/practices/:id", perm(domain.PermPracticesUpdate), idParam, h.UpdateRoom)
		authed.DELETE("/practices/:id", perm(domain.PermPracticesDelete), idParam, h.DeleteRoom)

		authed.POST("/practices/:id/submissions", perm(domain.PermPracticesSubmit), idParam, h.Submit)
		authed.GET("/practices/:id/submissions", perm(domain.PermPracticesView), idParam, h.ListSubmissions)
		authed.GET("/practices/submissions/:submissionId", perm(domain.PermPracticesView), submissionIDParam, h.GetSubmission)
		authed.PUT("/practices/submissions/:submissionId/grade", perm(domain.PermPracticesGrade), submissionIDParam, h.Grade)
	}
}

// ListRooms returns practice rooms visible to the caller.
// @Summary List practice rooms
// @Description Returns practices visible to the caller, enriched per-viewer (status, my_submission) and with manager stats when permitted. Filter by class_id, class_session_id, status (upcoming|to_submit|submitted|graded|missed), window (upcoming|open|ended), needs_grading. Search matches: title, content. Orderable: created_at, updated_at, start_time, end_time, title.
// @Tags Practices
// @Produce json
// @Security BearerAuth
// @Param class_id query string false "Filter by class UUID"
// @Param class_session_id query string false "Filter by class session UUID"
// @Param status query string false "Student status: upcoming, to_submit, submitted, graded, missed"
// @Param window query string false "Manager window state: upcoming, open, ended"
// @Param needs_grading query bool false "Manager: only rooms with ungraded submissions"
// @Param search query string false "Substring match on title/content"
// @Param order_by query string false "One of: created_at, updated_at, start_time, end_time, title"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.PracticeRoomView}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /practices [get]
func (h *Handler) ListRooms(c *gin.Context) {
	var q domain.ListPracticeRoomsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{
		"class_id":         &q.ClassID,
		"class_session_id": &q.ClassSessionID,
	}); err != nil {
		_ = c.Error(err)
		return
	}
	if v := c.Query("status"); v != "" {
		q.StudentStatus = &v
	}
	if v := c.Query("window"); v != "" {
		q.WindowState = &v
	}
	if c.Query("needs_grading") == "true" {
		needs := true
		q.NeedsGrading = &needs
	}
	q.ListParams = listparams.Bind(c, roomsListConfig)
	rooms, total, err := h.svc.ListRooms(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(rooms, total, q.ListParams))
}

// CreateRoom creates a practice room inside a class session.
// @Summary Create practice room
// @Tags Practices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreatePracticeRoomDTO true "Practice room data"
// @Success 201 {object} domain.Response{data=domain.PracticeRoom}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /practices [post]
func (h *Handler) CreateRoom(c *gin.Context) {
	var dto domain.CreatePracticeRoomDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	room, err := h.svc.CreateRoom(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, room)
}

// GetRoom returns a practice room by ID.
// @Summary Get practice room
// @Tags Practices
// @Produce json
// @Security BearerAuth
// @Param id path string true "Practice Room UUID"
// @Success 200 {object} domain.Response{data=domain.PracticeRoom}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /practices/{id} [get]
func (h *Handler) GetRoom(c *gin.Context) {
	room, err := h.svc.GetRoom(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, room)
}

// UpdateRoom updates a practice room.
// @Summary Update practice room
// @Tags Practices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Practice Room UUID"
// @Param body body domain.UpdatePracticeRoomDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.PracticeRoom}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /practices/{id} [put]
func (h *Handler) UpdateRoom(c *gin.Context) {
	var dto domain.UpdatePracticeRoomDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	room, err := h.svc.UpdateRoom(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, room)
}

// DeleteRoom soft-deletes a practice room.
// @Summary Delete practice room
// @Tags Practices
// @Produce json
// @Security BearerAuth
// @Param id path string true "Practice Room UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /practices/{id} [delete]
func (h *Handler) DeleteRoom(c *gin.Context) {
	if err := h.svc.DeleteRoom(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// Submit creates a submission for a practice room.
// @Summary Submit to practice room
// @Description Students can only submit between start_time and end_time.
// @Tags Practices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Practice Room UUID"
// @Param body body domain.CreatePracticeSubmissionDTO true "Submission data"
// @Success 201 {object} domain.Response{data=domain.PracticeSubmission}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody} "Already submitted"
// @Router /practices/{id}/submissions [post]
func (h *Handler) Submit(c *gin.Context) {
	var dto domain.CreatePracticeSubmissionDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	sub, err := h.svc.Submit(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, sub)
}

// ListSubmissions lists all submissions for a practice room (managers only).
// @Summary List submissions for practice room
// @Description Only room owner, staff, and admins can view all submissions. Orderable: submitted_at, created_at, score.
// @Tags Practices
// @Produce json
// @Security BearerAuth
// @Param id path string true "Practice Room UUID"
// @Param order_by query string false "One of: submitted_at, created_at, score"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.PracticeSubmission}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /practices/{id}/submissions [get]
func (h *Handler) ListSubmissions(c *gin.Context) {
	q := domain.ListPracticeSubmissionsQuery{
		ListParams: listparams.Bind(c, submissionsListConfig),
	}
	subs, total, err := h.svc.ListSubmissions(c.Request.Context(), httpx.UUIDParam(c, "id"), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(subs, total, q.ListParams))
}

// GetSubmission returns a single submission.
// @Summary Get submission
// @Tags Practices
// @Produce json
// @Security BearerAuth
// @Param submissionId path string true "Submission UUID"
// @Success 200 {object} domain.Response{data=domain.PracticeSubmission}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /practices/submissions/{submissionId} [get]
func (h *Handler) GetSubmission(c *gin.Context) {
	sub, err := h.svc.GetSubmission(c.Request.Context(), httpx.UUIDParam(c, "submissionId"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, sub)
}

// Grade scores a submission.
// @Summary Grade submission
// @Tags Practices
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param submissionId path string true "Submission UUID"
// @Param body body domain.GradePracticeSubmissionDTO true "Grade data"
// @Success 200 {object} domain.Response{data=domain.PracticeSubmission}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /practices/submissions/{submissionId}/grade [put]
func (h *Handler) Grade(c *gin.Context) {
	var dto domain.GradePracticeSubmissionDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	sub, err := h.svc.Grade(c.Request.Context(), httpx.UUIDParam(c, "submissionId"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, sub)
}
