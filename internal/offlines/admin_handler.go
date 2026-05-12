package offlines

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

// adminRoomsListConfig is the white-list for GET /admin/offlines. Anything
// outside these slices is silently ignored and falls back to defaults.
var adminRoomsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"title", "description"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "published_at", "title", "view_count"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

// AdminHandler registers under /api/v1/admin. The admin group is already
// guarded by auth middleware + RequireAdmin, so this handler only binds
// input, forwards to the service, and attaches errors.
type AdminHandler struct {
	svc domain.OfflineService
}

func NewAdminHandler(svc domain.OfflineService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	idParam := httpx.RequireUUIDParam("id")

	group.GET("/offlines", h.List)
	group.DELETE("/offlines/:id", idParam, h.HardDelete)
}

// List returns offline rooms across all organizations.
// @Summary [Admin] List offline rooms
// @Description Cross-org list. Search matches: title, description. Orderable: created_at, updated_at, published_at, title, view_count. Filters: class_id, class_session_id, creator_id, include_deleted.
// @Tags Admin/Offlines
// @Produce json
// @Security BearerAuth
// @Param class_id query string false "Filter by class UUID"
// @Param class_session_id query string false "Filter by class session UUID"
// @Param creator_id query string false "Filter by creator UUID"
// @Param include_deleted query bool false "Include soft-deleted rooms"
// @Param search query string false "Substring match on title/description"
// @Param order_by query string false "One of: created_at, updated_at, published_at, title, view_count"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.OfflineRoom}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/offlines [get]
func (h *AdminHandler) List(c *gin.Context) {
	var q domain.AdminListOfflineRoomsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
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

// HardDelete permanently deletes an offline room.
// @Summary [Admin] Hard-delete offline room
// @Tags Admin/Offlines
// @Produce json
// @Security BearerAuth
// @Param id path string true "Offline Room UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/offlines/{id} [delete]
func (h *AdminHandler) HardDelete(c *gin.Context) {
	if err := h.svc.AdminHardDelete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
