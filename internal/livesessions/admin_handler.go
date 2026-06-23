package livesessions

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

// adminLiveRoomsListConfig white-lists search/order for GET /admin/live-rooms.
var adminLiveRoomsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"livekit_room_name"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "status", "actual_start_time", "actual_end_time"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
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

// @Summary [Admin] List live rooms
// @Description Cross-org list. Search matches substrings of: livekit_room_name. Orderable fields: created_at, updated_at, status, actual_start_time, actual_end_time. Filters: status, user_id (teacher), class_id, class_session_id, include_deleted.
// @Tags Admin/LiveSessions
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status: created|active|finished"
// @Param user_id query string false "Filter by teacher UUID"
// @Param class_id query string false "Filter by class UUID"
// @Param class_session_id query string false "Filter by class session UUID"
// @Param include_deleted query bool false "Include soft-deleted rooms"
// @Param search query string false "Substring match on livekit_room_name"
// @Param order_by query string false "One of: created_at, updated_at, status, actual_start_time, actual_end_time"
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
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{
		"user_id":          &q.UserID,
		"class_id":         &q.ClassID,
		"class_session_id": &q.ClassSessionID,
	}); err != nil {
		_ = c.Error(err)
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
