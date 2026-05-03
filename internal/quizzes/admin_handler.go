package quizzes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var adminQuizzesListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"title", "description"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "title"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

type AdminHandler struct {
	svc domain.QuizService
}

func NewAdminHandler(svc domain.QuizService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	idParam := httpx.RequireUUIDParam("id")

	group.GET("/quizzes", h.List)
	group.DELETE("/quizzes/:id", idParam, h.HardDelete)
}

// List returns quizzes with optional filters.
// @Summary [Admin] List quizzes
// @Tags Admin/Quizzes
// @Produce json
// @Security BearerAuth
// @Param class_id query string false "Filter by class UUID"
// @Param user_id query string false "Filter by owner UUID"
// @Param include_deleted query bool false "Include soft-deleted rows"
// @Param search query string false "Substring match on title/description"
// @Param order_by query string false "One of: created_at, updated_at, title"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Quiz}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/quizzes [get]
func (h *AdminHandler) List(c *gin.Context) {
	var q domain.AdminListQuizzesQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	q.ListParams = listparams.Bind(c, adminQuizzesListConfig)
	quizzes, total, err := h.svc.AdminList(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(quizzes, total, q.ListParams))
}

// HardDelete permanently removes a quiz.
// @Summary [Admin] Hard-delete quiz
// @Tags Admin/Quizzes
// @Produce json
// @Security BearerAuth
// @Param id path string true "Quiz UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/quizzes/{id} [delete]
func (h *AdminHandler) HardDelete(c *gin.Context) {
	if err := h.svc.AdminHardDelete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
