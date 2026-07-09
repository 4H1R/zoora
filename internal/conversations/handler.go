package conversations

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

// conversationsListConfig is the handler-owned white-list for GET
// /conversations. Only columns in these slices can be searched/ordered by
// the client; anything else is silently ignored in favour of the defaults.
var conversationsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"name"},
	AllowedOrderFields:  []string{"updated_at", "created_at", "name"},
	DefaultOrderBy:      "updated_at",
	DefaultOrderDir:     "desc",
}

type Handler struct{ svc domain.ConversationService }

func NewHandler(svc domain.ConversationService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	id := httpx.RequireUUIDParam("id")
	msgID := httpx.RequireUUIDParam("messageId")
	userID := httpx.RequireUUIDParam("userId")

	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/conversations", perm(domain.PermConversationsView), h.List)
		authed.POST("/conversations", perm(domain.PermConversationsView), h.Create) // group/channel (service enforces manage)
		authed.POST("/conversations/direct", perm(domain.PermConversationsView), h.CreateDirect)
		authed.GET("/conversations/:id", perm(domain.PermConversationsView), id, h.Get)
		authed.PATCH("/conversations/:id", perm(domain.PermConversationsView), id, h.Update)
		authed.DELETE("/conversations/:id", perm(domain.PermConversationsView), id, h.Delete)

		authed.GET("/conversations/:id/members", perm(domain.PermConversationsView), id, h.ListMembers)
		authed.POST("/conversations/:id/members", perm(domain.PermConversationsView), id, h.AddMember)
		authed.DELETE("/conversations/:id/members/:userId", perm(domain.PermConversationsView), id, userID, h.RemoveMember)
		authed.POST("/conversations/:id/leave", perm(domain.PermConversationsView), id, h.Leave)
		authed.POST("/conversations/:id/read", perm(domain.PermConversationsView), id, h.MarkRead)
		authed.POST("/conversations/:id/mute", perm(domain.PermConversationsView), id, h.SetMuted)

		authed.GET("/conversations/:id/messages", perm(domain.PermConversationsView), id, h.ListMessages)
		authed.POST("/conversations/:id/messages", perm(domain.PermConversationsView), id, h.SendMessage)
		authed.GET("/conversations/:id/pins", perm(domain.PermConversationsView), id, h.ListPinned)

		authed.PATCH("/conversations/messages/:messageId", perm(domain.PermConversationsView), msgID, h.EditMessage)
		authed.DELETE("/conversations/messages/:messageId", perm(domain.PermConversationsView), msgID, h.DeleteMessage)
		authed.POST("/conversations/messages/:messageId/reactions", perm(domain.PermConversationsView), msgID, h.ToggleReaction)
		authed.POST("/conversations/messages/:messageId/pin", perm(domain.PermConversationsView), msgID, h.Pin)
		authed.POST("/conversations/messages/:messageId/unpin", perm(domain.PermConversationsView), msgID, h.Unpin)
	}
}

// parsePositiveInt parses s as a strictly positive integer.
func parsePositiveInt(s string) (int, error) {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, strconv.ErrRange
	}
	return n, nil
}

// parseCursor builds a domain.MessageCursor from the request's query string
// (limit/before/after/around). Unparseable values are silently ignored in
// favour of defaults, matching listparams.Bind's tolerant behaviour.
func parseCursor(c *gin.Context) domain.MessageCursor {
	var cur domain.MessageCursor
	cur.Limit = 50
	if v := c.Query("limit"); v != "" {
		if n, err := parsePositiveInt(v); err == nil {
			cur.Limit = n
		}
	}
	if v := c.Query("before"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			cur.Before = &id
		}
	}
	if v := c.Query("after"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			cur.After = &id
		}
	}
	if v := c.Query("around"); v != "" {
		if id, err := uuid.Parse(v); err == nil {
			cur.Around = &id
		}
	}
	return cur
}

// List returns conversations the caller is a member of.
func (h *Handler) List(c *gin.Context) {
	p := listparams.Bind(c, conversationsListConfig)
	q := domain.ListConversationsQuery{ListParams: p}
	if t := domain.ConversationType(c.Query("type")); t.Valid() {
		q.Type = &t
	}
	convs, total, err := h.svc.ListForCaller(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(convs, total, p))
}

// Create creates a group or channel conversation. The service enforces that
// only org admins / PermConversationsManage holders may do so.
func (h *Handler) Create(c *gin.Context) {
	var dto domain.CreateConversationDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	conv, err := h.svc.CreateGroupOrChannel(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, conv)
}

// CreateDirect creates (or returns the existing) direct conversation between
// the caller and another user.
func (h *Handler) CreateDirect(c *gin.Context) {
	var dto domain.CreateDirectDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	conv, err := h.svc.CreateOrGetDirect(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, conv)
}

// Get returns a single conversation the caller is a member of.
func (h *Handler) Get(c *gin.Context) {
	conv, err := h.svc.Get(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, conv)
}

// Update updates a conversation's metadata.
func (h *Handler) Update(c *gin.Context) {
	var dto domain.UpdateConversationDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	conv, err := h.svc.Update(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, conv)
}

// Delete hard-deletes a conversation and cascades its members/messages.
func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// ListMembers lists a conversation's roster.
func (h *Handler) ListMembers(c *gin.Context) {
	members, err := h.svc.ListMembers(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, members)
}

// AddMember adds a user to a group/channel conversation.
func (h *Handler) AddMember(c *gin.Context) {
	var dto domain.AddConversationMemberDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	m, err := h.svc.AddMember(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, m)
}

// RemoveMember removes a user from a conversation.
func (h *Handler) RemoveMember(c *gin.Context) {
	if err := h.svc.RemoveMember(c.Request.Context(), httpx.UUIDParam(c, "id"), httpx.UUIDParam(c, "userId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// Leave removes the caller from a conversation.
func (h *Handler) Leave(c *gin.Context) {
	if err := h.svc.Leave(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// MarkRead advances the caller's read cursor for a conversation.
func (h *Handler) MarkRead(c *gin.Context) {
	var dto domain.MarkReadDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.svc.MarkRead(c.Request.Context(), httpx.UUIDParam(c, "id"), dto); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// setMutedDTO is the request body for SetMuted: nil/absent = unmute.
type setMutedDTO struct {
	MutedUntil *time.Time `json:"muted_until"`
}

// SetMuted mutes/unmutes the caller's notifications for a conversation until
// the given time (nil clears the mute).
func (h *Handler) SetMuted(c *gin.Context) {
	var dto setMutedDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.svc.SetMuted(c.Request.Context(), httpx.UUIDParam(c, "id"), dto.MutedUntil); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// ListMessages returns a keyset window of messages in a conversation.
func (h *Handler) ListMessages(c *gin.Context) {
	cur := parseCursor(c)
	msgs, err := h.svc.ListMessages(c.Request.Context(), httpx.UUIDParam(c, "id"), cur)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, msgs)
}

// SendMessage posts a new message to a conversation.
func (h *Handler) SendMessage(c *gin.Context) {
	var dto domain.SendConversationMessageDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	msg, err := h.svc.SendMessage(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, msg)
}

// ListPinned returns the pinned messages of a conversation.
func (h *Handler) ListPinned(c *gin.Context) {
	msgs, err := h.svc.ListPinned(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, msgs)
}

// EditMessage updates a message's content. Sender or a conversation admin.
func (h *Handler) EditMessage(c *gin.Context) {
	var dto domain.UpdateConversationMessageDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	msg, err := h.svc.EditMessage(c.Request.Context(), httpx.UUIDParam(c, "messageId"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, msg)
}

// DeleteMessage hard-deletes a message. Sender or a conversation admin.
func (h *Handler) DeleteMessage(c *gin.Context) {
	if err := h.svc.DeleteMessage(c.Request.Context(), httpx.UUIDParam(c, "messageId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// ToggleReaction adds/removes the caller's reaction emoji on a message.
func (h *Handler) ToggleReaction(c *gin.Context) {
	var dto domain.ToggleConversationReactionDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	msg, err := h.svc.ToggleReaction(c.Request.Context(), httpx.UUIDParam(c, "messageId"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, msg)
}

// Pin pins a message. Conversation-manage authz.
func (h *Handler) Pin(c *gin.Context) {
	if err := h.svc.PinMessage(c.Request.Context(), httpx.UUIDParam(c, "messageId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// Unpin unpins a message. Conversation-manage authz.
func (h *Handler) Unpin(c *gin.Context) {
	if err := h.svc.UnpinMessage(c.Request.Context(), httpx.UUIDParam(c, "messageId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
