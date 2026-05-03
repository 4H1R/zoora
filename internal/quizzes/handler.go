package quizzes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var quizzesListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"title", "description"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "title", "duration_minutes"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

var rulesListConfig = domain.ListConfig{
	AllowedOrderFields: []string{"created_at", "type"},
	DefaultOrderBy:     "created_at",
	DefaultOrderDir:    "asc",
}

var roomsListConfig = domain.ListConfig{
	AllowedOrderFields: []string{"created_at", "started_at", "ended_at"},
	DefaultOrderBy:     "created_at",
	DefaultOrderDir:    "desc",
}

var submissionsListConfig = domain.ListConfig{
	AllowedOrderFields: []string{"created_at", "started_at", "submitted_at", "total_score"},
	DefaultOrderBy:     "created_at",
	DefaultOrderDir:    "desc",
}

type Handler struct {
	svc domain.QuizService
}

func NewHandler(svc domain.QuizService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")
	ruleIDParam := httpx.RequireUUIDParam("ruleId")
	roomIDParam := httpx.RequireUUIDParam("roomId")
	submissionIDParam := httpx.RequireUUIDParam("submissionId")

	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/quizzes", h.List)
		authed.POST("/quizzes", perm(domain.PermQuizzesCreate), h.Create)
		authed.GET("/quizzes/:id", idParam, h.Get)
		authed.PUT("/quizzes/:id", perm(domain.PermQuizzesUpdate), idParam, h.Update)
		authed.DELETE("/quizzes/:id", perm(domain.PermQuizzesDelete), idParam, h.Delete)

		authed.GET("/quizzes/:id/rules", idParam, h.ListRules)
		authed.POST("/quizzes/:id/rules", perm(domain.PermQuizzesUpdate), idParam, h.CreateRule)
		authed.GET("/quizzes/rules/:ruleId", ruleIDParam, h.GetRule)
		authed.PUT("/quizzes/rules/:ruleId", perm(domain.PermQuizzesUpdate), ruleIDParam, h.UpdateRule)
		authed.DELETE("/quizzes/rules/:ruleId", perm(domain.PermQuizzesDelete), ruleIDParam, h.DeleteRule)

		authed.GET("/quizzes/:id/rooms", idParam, h.ListRooms)
		authed.POST("/quizzes/:id/rooms", perm(domain.PermQuizzesUpdate), idParam, h.CreateRoom)
		authed.GET("/quizzes/rooms/:roomId", roomIDParam, h.GetRoom)
		authed.POST("/quizzes/rooms/:roomId/start", perm(domain.PermQuizzesUpdate), roomIDParam, h.StartRoom)
		authed.POST("/quizzes/rooms/:roomId/end", perm(domain.PermQuizzesUpdate), roomIDParam, h.EndRoom)

		authed.POST("/quizzes/:id/submissions", idParam, h.StartSubmission)
		authed.GET("/quizzes/:id/submissions", idParam, h.ListSubmissions)
		authed.POST("/quizzes/submissions/:submissionId/submit", submissionIDParam, h.SubmitQuiz)
		authed.GET("/quizzes/submissions/:submissionId", submissionIDParam, h.GetSubmission)
		authed.POST("/quizzes/submissions/:submissionId/grade", perm(domain.PermQuizzesUpdate), submissionIDParam, h.GradeSubmission)
	}
}

// List returns quizzes visible to the caller.
// @Summary List quizzes (scoped by RBAC)
// @Description Returns quizzes filtered by caller role: admins see all, org-admins see their org, teachers see owned quizzes, students see quizzes for enrolled classes. Search matches substrings of: title, description. Orderable fields: created_at, updated_at, title, duration_minutes.
// @Tags Quizzes
// @Produce json
// @Security BearerAuth
// @Param class_id query string false "Filter by class UUID"
// @Param search query string false "Substring match on title/description"
// @Param order_by query string false "One of: created_at, updated_at, title, duration_minutes"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Quiz}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes [get]
func (h *Handler) List(c *gin.Context) {
	var q domain.ListQuizzesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	q.ListParams = listparams.Bind(c, quizzesListConfig)
	quizzes, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(quizzes, total, q.ListParams))
}

// Create creates a quiz for a class.
// @Summary Create quiz
// @Tags Quizzes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateQuizDTO true "Quiz data"
// @Success 201 {object} domain.Response{data=domain.Quiz}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes [post]
func (h *Handler) Create(c *gin.Context) {
	var dto domain.CreateQuizDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	quiz, err := h.svc.Create(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, quiz)
}

// Get returns a quiz by ID.
// @Summary Get quiz
// @Tags Quizzes
// @Produce json
// @Security BearerAuth
// @Param id path string true "Quiz UUID"
// @Success 200 {object} domain.Response{data=domain.Quiz}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	quiz, err := h.svc.GetByID(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, quiz)
}

// Update updates a quiz.
// @Summary Update quiz
// @Tags Quizzes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Quiz UUID"
// @Param body body domain.UpdateQuizDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.Quiz}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	var dto domain.UpdateQuizDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	quiz, err := h.svc.Update(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, quiz)
}

// Delete soft-deletes a quiz.
// @Summary Delete quiz
// @Tags Quizzes
// @Produce json
// @Security BearerAuth
// @Param id path string true "Quiz UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// --- Rules ---

// ListRules returns rules for a quiz.
// @Summary List quiz rules
// @Tags Quizzes
// @Produce json
// @Security BearerAuth
// @Param id path string true "Quiz UUID"
// @Param order_by query string false "One of: created_at, type"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.QuizRule}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/{id}/rules [get]
func (h *Handler) ListRules(c *gin.Context) {
	q := domain.ListQuizRulesQuery{ListParams: listparams.Bind(c, rulesListConfig)}
	rules, total, err := h.svc.ListRules(c.Request.Context(), httpx.UUIDParam(c, "id"), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(rules, total, q.ListParams))
}

// CreateRule adds a rule to a quiz.
// @Summary Create quiz rule
// @Tags Quizzes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Quiz UUID"
// @Param body body domain.CreateQuizRuleDTO true "Rule data"
// @Success 201 {object} domain.Response{data=domain.QuizRule}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/{id}/rules [post]
func (h *Handler) CreateRule(c *gin.Context) {
	var dto domain.CreateQuizRuleDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	rule, err := h.svc.CreateRule(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, rule)
}

// GetRule returns a quiz rule by ID.
// @Summary Get quiz rule
// @Tags Quizzes
// @Produce json
// @Security BearerAuth
// @Param ruleId path string true "Rule UUID"
// @Success 200 {object} domain.Response{data=domain.QuizRule}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/rules/{ruleId} [get]
func (h *Handler) GetRule(c *gin.Context) {
	rule, err := h.svc.GetRule(c.Request.Context(), httpx.UUIDParam(c, "ruleId"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, rule)
}

// UpdateRule updates a quiz rule.
// @Summary Update quiz rule
// @Tags Quizzes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param ruleId path string true "Rule UUID"
// @Param body body domain.UpdateQuizRuleDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.QuizRule}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/rules/{ruleId} [put]
func (h *Handler) UpdateRule(c *gin.Context) {
	var dto domain.UpdateQuizRuleDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	rule, err := h.svc.UpdateRule(c.Request.Context(), httpx.UUIDParam(c, "ruleId"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, rule)
}

// DeleteRule deletes a quiz rule.
// @Summary Delete quiz rule
// @Tags Quizzes
// @Produce json
// @Security BearerAuth
// @Param ruleId path string true "Rule UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/rules/{ruleId} [delete]
func (h *Handler) DeleteRule(c *gin.Context) {
	if err := h.svc.DeleteRule(c.Request.Context(), httpx.UUIDParam(c, "ruleId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// --- Rooms ---

// ListRooms returns rooms for a quiz.
// @Summary List quiz rooms
// @Tags Quizzes
// @Produce json
// @Security BearerAuth
// @Param id path string true "Quiz UUID"
// @Param order_by query string false "One of: created_at, started_at, ended_at"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.QuizRoom}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/{id}/rooms [get]
func (h *Handler) ListRooms(c *gin.Context) {
	q := domain.ListQuizRoomsQuery{ListParams: listparams.Bind(c, roomsListConfig)}
	rooms, total, err := h.svc.ListRooms(c.Request.Context(), httpx.UUIDParam(c, "id"), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(rooms, total, q.ListParams))
}

// CreateRoom creates a room for a quiz.
// @Summary Create quiz room
// @Tags Quizzes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Quiz UUID"
// @Param body body domain.CreateQuizRoomDTO true "Room data"
// @Success 201 {object} domain.Response{data=domain.QuizRoom}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/{id}/rooms [post]
func (h *Handler) CreateRoom(c *gin.Context) {
	var dto domain.CreateQuizRoomDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	room, err := h.svc.CreateRoom(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, room)
}

// GetRoom returns a quiz room by ID.
// @Summary Get quiz room
// @Tags Quizzes
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "Room UUID"
// @Success 200 {object} domain.Response{data=domain.QuizRoom}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/rooms/{roomId} [get]
func (h *Handler) GetRoom(c *gin.Context) {
	room, err := h.svc.GetRoom(c.Request.Context(), httpx.UUIDParam(c, "roomId"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, room)
}

// StartRoom starts a quiz room, allowing students to begin submissions.
// @Summary Start quiz room
// @Tags Quizzes
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "Room UUID"
// @Success 200 {object} domain.Response{data=domain.QuizRoom}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/rooms/{roomId}/start [post]
func (h *Handler) StartRoom(c *gin.Context) {
	room, err := h.svc.StartRoom(c.Request.Context(), httpx.UUIDParam(c, "roomId"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, room)
}

// EndRoom ends a quiz room, closing it for new submissions.
// @Summary End quiz room
// @Tags Quizzes
// @Produce json
// @Security BearerAuth
// @Param roomId path string true "Room UUID"
// @Success 200 {object} domain.Response{data=domain.QuizRoom}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/rooms/{roomId}/end [post]
func (h *Handler) EndRoom(c *gin.Context) {
	room, err := h.svc.EndRoom(c.Request.Context(), httpx.UUIDParam(c, "roomId"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, room)
}

// --- Submissions ---

// StartSubmission begins a quiz attempt for the caller.
// @Summary Start quiz submission
// @Description Starts a submission for the authenticated student. Requires enrollment in the quiz's class and an open quiz room. Only one submission per user per quiz is allowed.
// @Tags Quizzes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Quiz UUID"
// @Param body body domain.StartQuizSubmissionDTO true "Submission start data"
// @Success 201 {object} domain.Response{data=domain.QuizSubmission}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/{id}/submissions [post]
func (h *Handler) StartSubmission(c *gin.Context) {
	var dto domain.StartQuizSubmissionDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	sub, err := h.svc.StartSubmission(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, sub)
}

// SubmitQuiz finalizes a submission with answers and triggers auto-grading.
// @Summary Submit quiz answers
// @Description Submits answers for an in-progress submission. Auto-grades choice and short_answer questions. Enforces duration limit with 30s grace period.
// @Tags Quizzes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param submissionId path string true "Submission UUID"
// @Param body body domain.SubmitQuizDTO true "Answers"
// @Success 200 {object} domain.Response{data=domain.QuizSubmission}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/submissions/{submissionId}/submit [post]
func (h *Handler) SubmitQuiz(c *gin.Context) {
	var dto domain.SubmitQuizDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	sub, err := h.svc.SubmitQuiz(c.Request.Context(), httpx.UUIDParam(c, "submissionId"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, sub)
}

// GetSubmission returns a submission by ID.
// @Summary Get quiz submission
// @Tags Quizzes
// @Produce json
// @Security BearerAuth
// @Param submissionId path string true "Submission UUID"
// @Success 200 {object} domain.Response{data=domain.QuizSubmission}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/submissions/{submissionId} [get]
func (h *Handler) GetSubmission(c *gin.Context) {
	sub, err := h.svc.GetSubmission(c.Request.Context(), httpx.UUIDParam(c, "submissionId"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, sub)
}

// ListSubmissions returns submissions for a quiz.
// @Summary List quiz submissions
// @Description Teachers/managers see all submissions; students see only their own.
// @Tags Quizzes
// @Produce json
// @Security BearerAuth
// @Param id path string true "Quiz UUID"
// @Param user_id query string false "Filter by user UUID"
// @Param status query string false "Filter by status (in_progress, submitted, graded)"
// @Param order_by query string false "One of: created_at, started_at, submitted_at, total_score"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.QuizSubmission}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/{id}/submissions [get]
func (h *Handler) ListSubmissions(c *gin.Context) {
	var q domain.ListSubmissionsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	q.ListParams = listparams.Bind(c, submissionsListConfig)
	subs, total, err := h.svc.ListSubmissions(c.Request.Context(), httpx.UUIDParam(c, "id"), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(subs, total, q.ListParams))
}

// GradeSubmission manually grades a submitted quiz (e.g. descriptive questions).
// @Summary Grade quiz submission
// @Description Allows the quiz owner to manually set scores for individual answers and recalculates the total.
// @Tags Quizzes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param submissionId path string true "Submission UUID"
// @Param body body domain.GradeSubmissionDTO true "Grade data"
// @Success 200 {object} domain.Response{data=domain.QuizSubmission}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /quizzes/submissions/{submissionId}/grade [post]
func (h *Handler) GradeSubmission(c *gin.Context) {
	var dto domain.GradeSubmissionDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	sub, err := h.svc.GradeSubmission(c.Request.Context(), httpx.UUIDParam(c, "submissionId"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, sub)
}
