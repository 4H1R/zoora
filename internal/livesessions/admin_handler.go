package livesessions

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var adminLiveRoomsListConfig = domain.ListConfig{
	AllowedOrderFields: []string{"created_at", "updated_at", "status"},
	DefaultOrderBy:     "created_at",
	DefaultOrderDir:    "desc",
}

type AdminHandler struct {
	svc domain.LiveSessionService
}

func NewAdminHandler(svc domain.LiveSessionService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	idParam := httpx.RequireUUIDParam("id")

	group.GET("/live-rooms", h.List)
	group.POST("/live-rooms/:id/end", idParam, h.EndRoom)
	group.DELETE("/live-rooms/:id", idParam, h.HardDelete)
}

// List returns live rooms.
// @Summary [Admin] List live rooms
// @Tags Admin/LiveSessions
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status: created|active|finished"
// @Param include_deleted query bool false "Include soft-deleted rooms"
// @Param order_by query string false "One of: created_at, updated_at, status"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.LiveRoom}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/live-rooms [get]
func (h *AdminHandler) List(c *gin.Context) {
	var q domain.AdminListLiveRoomsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	q.ListParams = listparams.Bind(c, adminLiveRoomsListConfig)
	rooms, total, err := h.svc.AdminList(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(rooms, total, q.ListParams))
}

// EndRoom force-ends an active live room.
// @Summary [Admin] End live room
// @Tags Admin/LiveSessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Success 200 {object} domain.Response{data=domain.LiveRoom}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/live-rooms/{id}/end [post]
func (h *AdminHandler) EndRoom(c *gin.Context) {
	room, err := h.svc.AdminEndRoom(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, room)
}

// HardDelete permanently deletes a live room.
// @Summary [Admin] Hard-delete live room
// @Tags Admin/LiveSessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/live-rooms/{id} [delete]
func (h *AdminHandler) HardDelete(c *gin.Context) {
	if err := h.svc.AdminHardDelete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
