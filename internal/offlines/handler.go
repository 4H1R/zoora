package offlines

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var roomsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"title", "description"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "published_at", "title", "view_count"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

type Handler struct {
	svc domain.OfflineService
}

func NewHandler(svc domain.OfflineService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")

	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/offlines", h.ListRooms)
		authed.POST("/offlines", perm(domain.PermOfflinesCreate), h.CreateRoom)
		authed.GET("/offlines/:id", idParam, h.GetRoom)
		authed.PUT("/offlines/:id", perm(domain.PermOfflinesUpdate), idParam, h.UpdateRoom)
		authed.DELETE("/offlines/:id", perm(domain.PermOfflinesDelete), idParam, h.DeleteRoom)
	}
}

// ListRooms returns offline rooms visible to the caller.
// @Summary List offline rooms
// @Description Returns offline rooms filtered by caller role. Filter by class_id or class_session_id. Search matches: title, description. Orderable: created_at, updated_at, published_at, title, view_count.
// @Tags Offlines
// @Produce json
// @Security BearerAuth
// @Param class_id query string false "Filter by class UUID"
// @Param class_session_id query string false "Filter by class session UUID"
// @Param search query string false "Substring match on title/description"
// @Param order_by query string false "One of: created_at, updated_at, published_at, title, view_count"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.OfflineRoom}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /offlines [get]
func (h *Handler) ListRooms(c *gin.Context) {
	var q domain.ListOfflineRoomsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	q.ListParams = listparams.Bind(c, roomsListConfig)
	rooms, total, err := h.svc.ListRooms(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(rooms, total, q.ListParams))
}

// CreateRoom creates an offline room inside a class session.
// @Summary Create offline room
// @Tags Offlines
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateOfflineRoomDTO true "Offline room data"
// @Success 201 {object} domain.Response{data=domain.OfflineRoom}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /offlines [post]
func (h *Handler) CreateRoom(c *gin.Context) {
	var dto domain.CreateOfflineRoomDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	room, err := h.svc.CreateRoom(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, room)
}

// GetRoom returns an offline room by ID and increments view count.
// @Summary Get offline room
// @Tags Offlines
// @Produce json
// @Security BearerAuth
// @Param id path string true "Offline Room UUID"
// @Success 200 {object} domain.Response{data=domain.OfflineRoom}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /offlines/{id} [get]
func (h *Handler) GetRoom(c *gin.Context) {
	room, err := h.svc.GetRoom(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, room)
}

// UpdateRoom updates an offline room.
// @Summary Update offline room
// @Tags Offlines
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Offline Room UUID"
// @Param body body domain.UpdateOfflineRoomDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.OfflineRoom}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /offlines/{id} [put]
func (h *Handler) UpdateRoom(c *gin.Context) {
	var dto domain.UpdateOfflineRoomDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	room, err := h.svc.UpdateRoom(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, room)
}

// DeleteRoom soft-deletes an offline room.
// @Summary Delete offline room
// @Tags Offlines
// @Produce json
// @Security BearerAuth
// @Param id path string true "Offline Room UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /offlines/{id} [delete]
func (h *Handler) DeleteRoom(c *gin.Context) {
	if err := h.svc.DeleteRoom(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
