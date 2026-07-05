package livesessions

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

// liveRoomsListConfig is the handler-owned white-list for GET /live-rooms.
var liveRoomsListConfig = domain.ListConfig{
	AllowedOrderFields:  []string{"created_at", "updated_at", "status", "scheduled_start_time", "actual_start_time", "actual_end_time"},
	AllowedSearchFields: []string{"name"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

// participantsListConfig gates GET /live-rooms/:id/participants.
var participantsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"identity"},
	AllowedOrderFields:  []string{"joined_at", "left_at", "created_at", "total_duration_seconds"},
	DefaultOrderBy:      "joined_at",
	DefaultOrderDir:     "desc",
}

// recordingsListConfig gates GET /live-rooms/:id/recordings.
var recordingsListConfig = domain.ListConfig{
	AllowedOrderFields: []string{"started_at", "ended_at", "created_at", "duration", "size"},
	DefaultOrderBy:     "started_at",
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
		authed.PUT("/live-rooms/:id/participants/:identity/role", perm(domain.PermLiveSessionsJoin), idParam, h.SetParticipantRole)
		authed.POST("/live-rooms/:id/participants/:identity/mute", perm(domain.PermLiveSessionsJoin), idParam, h.MuteParticipant)
		authed.DELETE("/live-rooms/:id/participants/:identity", perm(domain.PermLiveSessionsJoin), idParam, h.RemoveParticipant)
		authed.POST("/live-rooms/:id/hand", perm(domain.PermLiveSessionsJoin), idParam, h.SetHand)
		authed.PUT("/live-rooms/:id/participants/:identity/hand", perm(domain.PermLiveSessionsJoin), idParam, h.SetParticipantHand)
		authed.POST("/live-rooms/:id/recordings", perm(domain.PermLiveSessionsManage), idParam, h.StartRecording)
		authed.POST("/live-rooms/:id/recordings/:recordingId/stop", perm(domain.PermLiveSessionsManage), idParam, recordingIDParam, h.StopRecording)
		authed.GET("/live-rooms/:id/recordings", perm(domain.PermLiveSessionsView), idParam, h.ListRecordings)
		authed.GET("/live-rooms/:id/whiteboard", perm(domain.PermLiveSessionsJoin), idParam, h.GetWhiteboard)
		authed.PUT("/live-rooms/:id/whiteboard", perm(domain.PermLiveSessionsJoin), idParam, h.SaveWhiteboard)
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
// @Description Returns rooms filtered by caller role: super-admins / livesessions:view_any see all, teachers see their classes' rooms, students see rooms in classes they are enrolled in. Orderable fields: created_at, updated_at, status, scheduled_start_time, actual_start_time, actual_end_time. Filters: status, class_id, class_session_id.
// @Tags LiveSessions
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status: created|active|finished"
// @Param class_id query string false "Filter by class UUID"
// @Param class_session_id query string false "Filter by class session UUID"
// @Param include_deleted query bool false "Include soft-deleted rooms (managers only)"
// @Param order_by query string false "One of: created_at, updated_at, status, scheduled_start_time, actual_start_time, actual_end_time"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.LiveRoom}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms [get]
func (h *Handler) List(c *gin.Context) {
	var q domain.ListLiveRoomsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{
		"class_id":         &q.ClassID,
		"class_session_id": &q.ClassSessionID,
	}); err != nil {
		_ = c.Error(err)
		return
	}
	q.ListParams = listparams.Bind(c, liveRoomsListConfig)
	rooms, total, err := h.svc.List(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(rooms, total, q.ListParams))
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
// @Description Search matches substrings of: identity. Orderable fields: joined_at, left_at, created_at, total_duration_seconds. Filters: active_only, user_id.
// @Tags LiveSessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Param active_only query bool false "true = still in room, false = already left"
// @Param user_id query string false "Filter by user UUID"
// @Param search query string false "Substring match on identity"
// @Param order_by query string false "One of: joined_at, left_at, created_at, total_duration_seconds"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.LiveParticipant}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/participants [get]
func (h *Handler) ListParticipants(c *gin.Context) {
	var q domain.ListLiveParticipantsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{"user_id": &q.UserID}); err != nil {
		_ = c.Error(err)
		return
	}
	q.ListParams = listparams.Bind(c, participantsListConfig)
	participants, total, err := h.svc.ListParticipants(c.Request.Context(), httpx.UUIDParam(c, "id"), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(participants, total, q.ListParams))
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
// @Router /live-rooms/{id}/recordings/{recordingId}/stop [post]
func (h *Handler) StopRecording(c *gin.Context) {
	rec, err := h.svc.StopRecording(c.Request.Context(), httpx.UUIDParam(c, "recordingId"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, rec)
}

// SetParticipantRole sets the role of a participant in a live room.
// @Summary Set participant role
// @Tags LiveSessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Param identity path string true "Participant identity"
// @Param body body domain.SetParticipantRoleDTO true "Role data"
// @Success 200 {object} domain.Response{data=domain.LiveParticipant}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/participants/{identity}/role [put]
func (h *Handler) SetParticipantRole(c *gin.Context) {
	var dto domain.SetParticipantRoleDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	participant, err := h.svc.SetParticipantRole(c.Request.Context(), httpx.UUIDParam(c, "id"), c.Param("identity"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, participant)
}

// MuteParticipant mutes or unmutes a participant track in a live room.
// @Summary Mute participant
// @Tags LiveSessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Param identity path string true "Participant identity"
// @Param body body domain.MuteParticipantDTO true "Mute data"
// @Success 200 {object} domain.Response
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/participants/{identity}/mute [post]
func (h *Handler) MuteParticipant(c *gin.Context) {
	var dto domain.MuteParticipantDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.svc.MuteParticipant(c.Request.Context(), httpx.UUIDParam(c, "id"), c.Param("identity"), dto); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// RemoveParticipant ejects a participant from a live room.
// @Summary Remove participant
// @Description Host-only. Ejects a participant from the room. A host cannot remove another host or themselves.
// @Tags LiveSessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Param identity path string true "Participant identity"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/participants/{identity} [delete]
func (h *Handler) RemoveParticipant(c *gin.Context) {
	if err := h.svc.RemoveParticipant(c.Request.Context(), httpx.UUIDParam(c, "id"), c.Param("identity")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// SetHand raises or lowers the caller's hand in a live room.
// @Summary Set hand raised
// @Tags LiveSessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Param body body domain.SetHandDTO true "Hand data"
// @Success 200 {object} domain.Response{data=domain.LiveParticipant}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/hand [post]
func (h *Handler) SetHand(c *gin.Context) {
	var dto domain.SetHandDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	participant, err := h.svc.SetHand(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, participant)
}

// SetParticipantHand lets a host lower or raise another participant's hand.
// @Summary Set participant hand
// @Tags LiveSessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Param identity path string true "Participant identity"
// @Param body body domain.SetHandDTO true "Hand data"
// @Success 200 {object} domain.Response{data=domain.LiveParticipant}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/participants/{identity}/hand [put]
func (h *Handler) SetParticipantHand(c *gin.Context) {
	var dto domain.SetHandDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	participant, err := h.svc.SetParticipantHand(c.Request.Context(), httpx.UUIDParam(c, "id"), c.Param("identity"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, participant)
}

// ListRecordings returns all recordings for a live room.
// @Summary List recordings
// @Description Orderable fields: started_at, ended_at, created_at, duration, size. Filters: status.
// @Tags LiveSessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Param status query string false "Filter by status: started|completed|failed"
// @Param order_by query string false "One of: started_at, ended_at, created_at, duration, size"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.LiveRecording}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/recordings [get]
func (h *Handler) ListRecordings(c *gin.Context) {
	var q domain.ListLiveRecordingsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	q.ListParams = listparams.Bind(c, recordingsListConfig)
	recs, total, err := h.svc.ListRecordings(c.Request.Context(), httpx.UUIDParam(c, "id"), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(recs, total, q.ListParams))
}

// GetWhiteboard returns the current tldraw snapshot for a live room.
// @Summary Get whiteboard snapshot
// @Description Returns the persisted tldraw snapshot. Any room participant (viewer or above) may read it. Returns an empty board ({}) if no snapshot has been saved yet.
// @Tags LiveSessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Success 200 {object} domain.Response{data=domain.LiveWhiteboard}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/whiteboard [get]
func (h *Handler) GetWhiteboard(c *gin.Context) {
	wb, err := h.svc.GetWhiteboard(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, wb)
}

// SaveWhiteboard persists a tldraw snapshot for a live room.
// @Summary Save whiteboard snapshot
// @Description Upserts the tldraw snapshot. Only hosts and presenters may write.
// @Tags LiveSessions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "LiveRoom UUID"
// @Param body body domain.SaveWhiteboardDTO true "Snapshot data"
// @Success 200 {object} domain.Response{data=domain.LiveWhiteboard}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /live-rooms/{id}/whiteboard [put]
func (h *Handler) SaveWhiteboard(c *gin.Context) {
	var dto domain.SaveWhiteboardDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	wb, err := h.svc.SaveWhiteboard(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, wb)
}
