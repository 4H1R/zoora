package qa

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var adminQAListConfig = domain.ListConfig{
	AllowedOrderFields: []string{"created_at", "updated_at"},
	DefaultOrderBy:     "created_at",
	DefaultOrderDir:    "desc",
}

type AdminHandler struct {
	svc domain.QAService
}

func NewAdminHandler(svc domain.QAService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	idParam := httpx.RequireUUIDParam("id")

	group.GET("/qa", h.List)
	group.DELETE("/qa/:id", idParam, h.HardDelete)
}

// List returns QA questions with optional filters.
// @Summary [Admin] List QA questions
// @Tags Admin/QA
// @Produce json
// @Security BearerAuth
// @Param user_id query string false "Filter by author UUID"
// @Param model_type query string false "Filter by model type"
// @Param model_id query string false "Filter by model UUID"
// @Param status query string false "Filter by status: open, resolved, dismissed"
// @Param include_deleted query bool false "Include soft-deleted rows"
// @Param order_by query string false "One of: created_at, updated_at"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.QAQuestion}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/qa [get]
func (h *AdminHandler) List(c *gin.Context) {
	var q domain.AdminListQAQuestionsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{
		"user_id":  &q.UserID,
		"model_id": &q.ModelID,
	}); err != nil {
		_ = c.Error(err)
		return
	}
	q.ListParams = listparams.Bind(c, adminQAListConfig)
	items, total, err := h.svc.AdminList(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(items, total, q.ListParams))
}

// HardDelete permanently removes a QA question.
// @Summary [Admin] Hard-delete QA question
// @Tags Admin/QA
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/qa/{id} [delete]
func (h *AdminHandler) HardDelete(c *gin.Context) {
	if err := h.svc.AdminHardDelete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
