package qa

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var qaListConfig = domain.ListConfig{
	AllowedOrderFields: []string{"created_at"},
	DefaultOrderBy:     "created_at",
	DefaultOrderDir:    "desc",
}

type Handler struct {
	svc domain.QAService
}

func NewHandler(svc domain.QAService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")
	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/qa", perm(domain.PermQAView), h.List)
		authed.POST("/qa", perm(domain.PermQACreate), h.Ask)
		authed.PATCH("/qa/:id", perm(domain.PermQACreate), idParam, h.Update)
		authed.DELETE("/qa/:id", perm(domain.PermQADelete), idParam, h.Delete)
		authed.POST("/qa/:id/vote", perm(domain.PermQAView), idParam, h.Vote)
		authed.POST("/qa/:id/resolve", perm(domain.PermQAModerate), idParam, h.Resolve)
		authed.POST("/qa/:id/dismiss", perm(domain.PermQAModerate), idParam, h.Dismiss)
		authed.POST("/qa/:id/reopen", perm(domain.PermQAModerate), idParam, h.Reopen)
	}
}

// List returns questions for a live model, ordered open-first by vote count.
// @Summary List QA questions
// @Description Returns audience questions for a model (model_type + model_id required), ordered open-first, then by vote count desc, then oldest-first. vote_count and voted_by_me are computed for the caller.
// @Tags QA
// @Produce json
// @Security BearerAuth
// @Param model_type query string true "Model type (live_session)"
// @Param model_id query string true "Model UUID (live room id)"
// @Param status query string false "Filter by status: open, resolved, dismissed"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.QAQuestionView}}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /qa [get]
func (h *Handler) List(c *gin.Context) {
	var q domain.ListQAQuestionsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{"model_id": &q.ModelID}); err != nil {
		_ = c.Error(err)
		return
	}
	q.ListParams = listparams.Bind(c, qaListConfig)
	items, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(items, total, q.ListParams))
}

// Ask creates a new question.
// @Summary Ask a QA question
// @Tags QA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateQAQuestionDTO true "Question"
// @Success 201 {object} domain.Response{data=domain.QAQuestion}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /qa [post]
func (h *Handler) Ask(c *gin.Context) {
	var dto domain.CreateQAQuestionDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	q, err := h.svc.Ask(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, q)
}

// Update edits the caller's own open question text.
// @Summary Edit a QA question
// @Tags QA
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question UUID"
// @Param body body domain.UpdateQAQuestionDTO true "New text"
// @Success 200 {object} domain.Response{data=domain.QAQuestion}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /qa/{id} [patch]
func (h *Handler) Update(c *gin.Context) {
	var dto domain.UpdateQAQuestionDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	q, err := h.svc.UpdateText(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, q)
}

// Delete soft-deletes a question (author or moderator).
// @Summary Delete a QA question
// @Tags QA
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question UUID"
// @Success 200 {object} domain.Response
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /qa/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// Vote toggles the caller's upvote on a question.
// @Summary Toggle upvote on a QA question
// @Tags QA
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question UUID"
// @Success 200 {object} domain.Response{data=domain.QAVoteResult}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /qa/{id}/vote [post]
func (h *Handler) Vote(c *gin.Context) {
	voted, count, err := h.svc.ToggleVote(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.QAVoteResult{Voted: voted, VoteCount: count})
}

// Resolve marks a question resolved (moderator only).
// @Summary Resolve a QA question
// @Tags QA
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question UUID"
// @Success 200 {object} domain.Response{data=domain.QAQuestion}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /qa/{id}/resolve [post]
func (h *Handler) Resolve(c *gin.Context) { h.transition(c, h.svc.Resolve) }

// Dismiss marks a question dismissed (moderator only).
// @Summary Dismiss a QA question
// @Tags QA
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question UUID"
// @Success 200 {object} domain.Response{data=domain.QAQuestion}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /qa/{id}/dismiss [post]
func (h *Handler) Dismiss(c *gin.Context) { h.transition(c, h.svc.Dismiss) }

// Reopen moves a closed question back to open (moderator only).
// @Summary Reopen a QA question
// @Tags QA
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question UUID"
// @Success 200 {object} domain.Response{data=domain.QAQuestion}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /qa/{id}/reopen [post]
func (h *Handler) Reopen(c *gin.Context) { h.transition(c, h.svc.Reopen) }

func (h *Handler) transition(c *gin.Context, fn func(context.Context, uuid.UUID) (*domain.QAQuestion, error)) {
	q, err := fn(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, q)
}
