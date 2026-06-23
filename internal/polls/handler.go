package polls

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var pollsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"name"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "name"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

var answersListConfig = domain.ListConfig{
	AllowedOrderFields: []string{"created_at"},
	DefaultOrderBy:     "created_at",
	DefaultOrderDir:    "desc",
}

type Handler struct {
	svc domain.PollService
}

func NewHandler(svc domain.PollService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")

	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/polls", perm(domain.PermPollsView), h.List)
		authed.POST("/polls", perm(domain.PermPollsCreate), h.Create)
		authed.GET("/polls/:id", perm(domain.PermPollsView), idParam, h.Get)
		authed.PUT("/polls/:id", perm(domain.PermPollsUpdate), idParam, h.Update)
		authed.DELETE("/polls/:id", perm(domain.PermPollsDelete), idParam, h.Delete)

		authed.POST("/polls/:id/answer", perm(domain.PermPollsView), idParam, h.Answer)
		authed.GET("/polls/:id/answers", perm(domain.PermPollsView), idParam, h.ListAnswers)
	}
}

// List returns polls visible to the caller.
// @Summary List polls (scoped by RBAC)
// @Description Returns polls filtered by caller role. Search matches substrings of: name. Orderable fields: created_at, updated_at, name.
// @Tags Polls
// @Produce json
// @Security BearerAuth
// @Param model_type query string false "Filter by model type"
// @Param model_id query string false "Filter by model UUID"
// @Param search query string false "Substring match on name"
// @Param order_by query string false "One of: created_at, updated_at, name"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Poll}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /polls [get]
func (h *Handler) List(c *gin.Context) {
	var q domain.ListPollsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{"model_id": &q.ModelID}); err != nil {
		_ = c.Error(err)
		return
	}
	q.ListParams = listparams.Bind(c, pollsListConfig)
	polls, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(polls, total, q.ListParams))
}

// @Summary Create poll
// @Tags Polls
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreatePollDTO true "Poll data"
// @Success 201 {object} domain.Response{data=domain.Poll}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /polls [post]
func (h *Handler) Create(c *gin.Context) {
	var dto domain.CreatePollDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	poll, err := h.svc.Create(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, poll)
}

// @Summary Get poll
// @Tags Polls
// @Produce json
// @Security BearerAuth
// @Param id path string true "Poll UUID"
// @Success 200 {object} domain.Response{data=domain.Poll}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /polls/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	poll, err := h.svc.GetByID(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, poll)
}

// @Summary Update poll
// @Tags Polls
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Poll UUID"
// @Param body body domain.UpdatePollDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.Poll}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /polls/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	var dto domain.UpdatePollDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	poll, err := h.svc.Update(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, poll)
}

// Delete soft-deletes a poll.
// @Summary Delete poll
// @Tags Polls
// @Produce json
// @Security BearerAuth
// @Param id path string true "Poll UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /polls/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// Answer casts votes on a poll.
// @Summary Answer poll
// @Description Submit answers (options) for a poll. Replaces any previous answers by this user.
// @Tags Polls
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Poll UUID"
// @Param body body domain.AnswerPollDTO true "Selected options"
// @Success 201 {object} domain.Response{data=[]domain.PollAnswer}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /polls/{id}/answer [post]
func (h *Handler) Answer(c *gin.Context) {
	var dto domain.AnswerPollDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	answers, err := h.svc.Answer(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, answers)
}

// @Summary List poll answers
// @Tags Polls
// @Produce json
// @Security BearerAuth
// @Param id path string true "Poll UUID"
// @Param user_id query string false "Filter by user UUID"
// @Param order_by query string false "One of: created_at"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.PollAnswer}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /polls/{id}/answers [get]
func (h *Handler) ListAnswers(c *gin.Context) {
	var q domain.ListPollAnswersQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{"user_id": &q.UserID}); err != nil {
		_ = c.Error(err)
		return
	}
	q.ListParams = listparams.Bind(c, answersListConfig)
	answers, total, err := h.svc.ListAnswers(c.Request.Context(), httpx.UUIDParam(c, "id"), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(answers, total, q.ListParams))
}
