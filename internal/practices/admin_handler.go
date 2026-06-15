package practices

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var adminRoomsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"title", "content"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "start_time", "end_time", "title"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

type AdminHandler struct {
	svc domain.PracticeService
}

func NewAdminHandler(svc domain.PracticeService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	idParam := httpx.RequireUUIDParam("id")

	group.GET("/practices", h.List)
	group.DELETE("/practices/:id", idParam, h.HardDelete)
}

// List returns practice rooms across all organizations.
// @Summary [Admin] List practice rooms
// @Description Cross-org list. Search matches: title, content. Orderable: created_at, updated_at, start_time, end_time, title. Filters: class_id, class_session_id, user_id, include_deleted.
// @Tags Admin/Practices
// @Produce json
// @Security BearerAuth
// @Param class_id query string false "Filter by class UUID"
// @Param class_session_id query string false "Filter by class session UUID"
// @Param user_id query string false "Filter by creator UUID"
// @Param include_deleted query bool false "Include soft-deleted rooms"
// @Param search query string false "Substring match on title/content"
// @Param order_by query string false "One of: created_at, updated_at, start_time, end_time, title"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.PracticeRoom}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/practices [get]
func (h *AdminHandler) List(c *gin.Context) {
	var q domain.AdminListPracticeRoomsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{
		"class_id":         &q.ClassID,
		"class_session_id": &q.ClassSessionID,
		"user_id":          &q.UserID,
	}); err != nil {
		_ = c.Error(err)
		return
	}
	q.ListParams = listparams.Bind(c, adminRoomsListConfig)
	rooms, total, err := h.svc.AdminList(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(rooms, total, q.ListParams))
}

// HardDelete permanently deletes a practice room.
// @Summary [Admin] Hard-delete practice room
// @Tags Admin/Practices
// @Produce json
// @Security BearerAuth
// @Param id path string true "Practice Room UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/practices/{id} [delete]
func (h *AdminHandler) HardDelete(c *gin.Context) {
	if err := h.svc.AdminHardDelete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
