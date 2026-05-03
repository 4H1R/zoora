package livesessions

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var liveRoomsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{},
	AllowedOrderFields:  []string{"created_at", "updated_at", "status"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

var participantsListConfig = domain.ListConfig{
	AllowedOrderFields: []string{"joined_at", "created_at"},
	DefaultOrderBy:     "joined_at",
	DefaultOrderDir:    "desc",
}

type Handler struct {
	svc domain.LiveSessionService
}

func NewHandler(svc domain.LiveSessionService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")
	recordingIDParam := httpx.RequireUUIDParam("recordingId")

	authed := rg.Group("", authMiddleware)
	{
		authed.POST("/live-rooms", perm(domain.PermLiveSessionsCreate), h.CreateRoom)
		authed.GET("/live-rooms", perm(domain.PermLiveSessionsView), h.List)
		authed.GET("/live-rooms/:id", perm(domain.PermLiveSessionsView), idParam, h.GetRoom)
		authed.POST("/live-rooms/:id/join", perm(domain.PermLiveSessionsJoin), idParam, h.JoinRoom)
		authed.POST("/live-rooms/:id/leave", perm(domain.PermLiveSessionsJoin), idParam, h.LeaveRoom)
		authed.POST("/live-rooms/:id/start", perm(domain.PermLiveSessionsManage), idParam, h.StartRoom)
		authed.POST("/live-rooms/:id/end", perm(domain.PermLiveSessionsManage), idParam, h.EndRoom)
		authed.PUT("/live-rooms/:id/config", perm(domain.PermLiveSessionsUpdate), idParam, h.UpdateRoomConfig)
		authed.POST("/live-rooms/:id/heartbeat", perm(domain.PermLiveSessionsManage), idParam, h.Heartbeat)
		authed.GET("/live-rooms/:id/participants", perm(domain.PermLiveSessionsView), idParam, h.ListParticipants)
		authed.POST("/live-rooms/:id/recordings", perm(domain.PermLiveSessionsManage), idParam, h.StartRecording)
		authed.DELETE("/live-rooms/:id/recordings/:recordingId", perm(domain.PermLiveSessionsManage), idParam, recordingIDParam, h.StopRecording)
		authed.GET("/live-rooms/:id/recordings", perm(domain.PermLiveSessionsView), idParam, h.ListRecordings)
	}
}

// CreateRoom creates a live room for a class session.
// @Summary Create live room
// @Tags LiveSessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateLiveRoomDTO true "Room data"
// @Success 201 {object} domain.Response{data=domain.LiveRoom}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms [post]
func (h *Handler) CreateRoom(c *gin.Context) {
	var dto domain.CreateLiveRoomDTO
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

// List returns live rooms visible to the caller.
// @Summary List live rooms (scoped by RBAC)
// @Tags LiveSessions
// @Produce json
// @Security BearerAuth
// @Param order_by query string false "One of: created_at, updated_at, status"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.LiveRoom}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms [get]
func (h *Handler) List(c *gin.Context) {
	p := listparams.Bind(c, liveRoomsListConfig)
	rooms, total, err := h.svc.List(c.Request.Context(), p)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(rooms, total, p))
}

// GetRoom returns a live room by ID.
// @Summary Get live room
// @Tags LiveSessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Success 200 {object} domain.Response{data=domain.LiveRoom}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id} [get]
func (h *Handler) GetRoom(c *gin.Context) {
	room, err := h.svc.GetRoom(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, room)
}

// JoinRoom generates a LiveKit token and records participation.
// @Summary Join live room
// @Tags LiveSessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Success 200 {object} domain.Response{data=domain.JoinLiveRoomResponse}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/join [post]
func (h *Handler) JoinRoom(c *gin.Context) {
	resp, err := h.svc.JoinRoom(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, resp)
}

// LeaveRoom records the caller leaving.
// @Summary Leave live room
// @Tags LiveSessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/leave [post]
func (h *Handler) LeaveRoom(c *gin.Context) {
	if err := h.svc.LeaveRoom(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// StartRoom transitions room to active and creates the LiveKit room.
// @Summary Start live room
// @Tags LiveSessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Success 200 {object} domain.Response{data=domain.LiveRoom}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/start [post]
func (h *Handler) StartRoom(c *gin.Context) {
	room, err := h.svc.StartRoom(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, room)
}

// EndRoom transitions room to finished and cleans up LiveKit resources.
// @Summary End live room
// @Tags LiveSessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Success 200 {object} domain.Response{data=domain.LiveRoom}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/end [post]
func (h *Handler) EndRoom(c *gin.Context) {
	room, err := h.svc.EndRoom(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, room)
}

// UpdateRoomConfig updates the room configuration.
// @Summary Update live room config
// @Tags LiveSessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Param body body domain.UpdateLiveRoomConfigDTO true "Config data"
// @Success 200 {object} domain.Response{data=domain.LiveRoom}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/config [put]
func (h *Handler) UpdateRoomConfig(c *gin.Context) {
	var dto domain.UpdateLiveRoomConfigDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	room, err := h.svc.UpdateRoomConfig(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, room)
}

// Heartbeat updates the host's last-seen timestamp.
// @Summary Heartbeat for live room
// @Tags LiveSessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/heartbeat [post]
func (h *Handler) Heartbeat(c *gin.Context) {
	if err := h.svc.Heartbeat(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// ListParticipants lists participants in a live room.
// @Summary List live room participants
// @Tags LiveSessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Param order_by query string false "One of: joined_at, created_at"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.LiveParticipant}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/participants [get]
func (h *Handler) ListParticipants(c *gin.Context) {
	p := listparams.Bind(c, participantsListConfig)
	participants, total, err := h.svc.ListParticipants(c.Request.Context(), httpx.UUIDParam(c, "id"), p)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(participants, total, p))
}

// StartRecording starts recording a live room.
// @Summary Start recording
// @Tags LiveSessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Success 201 {object} domain.Response{data=domain.LiveRecording}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/recordings [post]
func (h *Handler) StartRecording(c *gin.Context) {
	rec, err := h.svc.StartRecording(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, rec)
}

// StopRecording stops an active recording.
// @Summary Stop recording
// @Tags LiveSessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Param recordingId path string true "Recording UUID"
// @Success 200 {object} domain.Response{data=domain.LiveRecording}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/recordings/{recordingId} [delete]
func (h *Handler) StopRecording(c *gin.Context) {
	rec, err := h.svc.StopRecording(c.Request.Context(), httpx.UUIDParam(c, "recordingId"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, rec)
}

// ListRecordings returns all recordings for a live room.
// @Summary List recordings
// @Tags LiveSessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Success 200 {object} domain.Response{data=[]domain.LiveRecording}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/recordings [get]
func (h *Handler) ListRecordings(c *gin.Context) {
	recs, err := h.svc.ListRecordings(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, recs)
}
