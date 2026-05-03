package polls

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var adminPollsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"name"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "name"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

type AdminHandler struct {
	svc domain.PollService
}

func NewAdminHandler(svc domain.PollService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	idParam := httpx.RequireUUIDParam("id")

	group.GET("/polls", h.List)
	group.DELETE("/polls/:id", idParam, h.HardDelete)
}

// List returns polls with optional filters.
// @Summary [Admin] List polls
// @Tags Admin/Polls
// @Produce json
// @Security BearerAuth
// @Param user_id query string false "Filter by owner UUID"
// @Param model_type query string false "Filter by model type"
// @Param model_id query string false "Filter by model UUID"
// @Param include_deleted query bool false "Include soft-deleted rows"
// @Param search query string false "Substring match on name"
// @Param order_by query string false "One of: created_at, updated_at, name"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Poll}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/polls [get]
func (h *AdminHandler) List(c *gin.Context) {
	var q domain.AdminListPollsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	q.ListParams = listparams.Bind(c, adminPollsListConfig)
	polls, total, err := h.svc.AdminList(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(polls, total, q.ListParams))
}

// HardDelete permanently removes a poll.
// @Summary [Admin] Hard-delete poll
// @Tags Admin/Polls
// @Produce json
// @Security BearerAuth
// @Param id path string true "Poll UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/polls/{id} [delete]
func (h *AdminHandler) HardDelete(c *gin.Context) {
	if err := h.svc.AdminHardDelete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
