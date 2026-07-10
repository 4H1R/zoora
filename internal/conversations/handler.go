package conversations

import (
	"net/http"
	"strconv"
	"strings"
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
		authed.GET("/conversations/search", perm(domain.PermConversationsView), h.Search)                        // global ?q= (registered before /:id so gin's tree matches the static segment first)
		authed.GET("/conversations/presence", perm(domain.PermConversationsView), h.GetPresence)                 // batch online/last-seen (static segment, before /:id)
		authed.GET("/conversations/directory", perm(domain.PermConversationsView), h.SearchDirectory)            // people search (static, before /:id)
		authed.GET("/conversations/directory/:username", perm(domain.PermConversationsView), h.GetDirectoryUser) // mention resolve by username
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
		authed.GET("/conversations/:id/search", perm(domain.PermConversationsView), id, h.SearchInConv) // nav ?q=

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
// @Summary List conversations
// @Description Returns conversations the caller is a member of, newest-activity first. Search matches substrings of: name. Orderable fields: updated_at, created_at, name.
// @Tags Conversations
// @Produce json
// @Security BearerAuth
// @Param type query string false "Filter by type: direct, group, channel"
// @Param search query string false "Substring match on name"
// @Param order_by query string false "One of: updated_at, created_at, name"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Param page_size query int false "Items per page"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Conversation}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations [get]
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
// @Summary Create group or channel conversation
// @Description Only org admins or PermConversationsManage holders may create a group/channel; the service enforces this beyond the route's PermConversationsView gate.
// @Tags Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateConversationDTO true "Conversation data"
// @Success 201 {object} domain.Response{data=domain.Conversation}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations [post]
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
// @Summary Create or get direct conversation
// @Description Idempotent: repeated calls with the same user_id return the same DM. Rejects self-DMs and, when the other user resolves to a different org, cross-org DMs.
// @Tags Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateDirectDTO true "Other user"
// @Success 201 {object} domain.Response{data=domain.Conversation}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/direct [post]
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
// @Summary Get conversation
// @Tags Conversations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Success 200 {object} domain.Response{data=domain.Conversation}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	conv, err := h.svc.Get(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, conv)
}

// Update updates a conversation's metadata.
// @Summary Update conversation
// @Description Platform admins, same-org PermConversationsManage holders, and conversation-admin members may update; plain members may not.
// @Tags Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Param body body domain.UpdateConversationDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.Conversation}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/{id} [patch]
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
// @Summary Delete conversation
// @Tags Conversations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// ListMembers lists a conversation's roster.
// @Summary List conversation members
// @Tags Conversations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Success 200 {object} domain.Response{data=[]domain.ConversationMember}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/{id}/members [get]
func (h *Handler) ListMembers(c *gin.Context) {
	members, err := h.svc.ListMembers(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, members)
}

// AddMember adds a user to a group/channel conversation.
// @Summary Add conversation member
// @Description Platform admins, same-org PermConversationsManage holders, and conversation-admin members may add; the user must resolve to the caller's org.
// @Tags Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Param body body domain.AddConversationMemberDTO true "User to add"
// @Success 201 {object} domain.Response{data=domain.ConversationMember}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/{id}/members [post]
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
// @Summary Remove conversation member
// @Tags Conversations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Param userId path string true "User UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/{id}/members/{userId} [delete]
func (h *Handler) RemoveMember(c *gin.Context) {
	if err := h.svc.RemoveMember(c.Request.Context(), httpx.UUIDParam(c, "id"), httpx.UUIDParam(c, "userId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// Leave removes the caller from a conversation.
// @Summary Leave conversation
// @Tags Conversations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/{id}/leave [post]
func (h *Handler) Leave(c *gin.Context) {
	if err := h.svc.Leave(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// MarkRead advances the caller's read cursor for a conversation.
// @Summary Mark conversation read
// @Tags Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Param body body domain.MarkReadDTO true "Last-read message"
// @Success 200 {object} domain.Response
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/{id}/read [post]
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
// @Summary Mute/unmute conversation
// @Description Sets the caller's muted_until for this conversation; omit or set null muted_until to unmute. Muted members are skipped by notification fan-out.
// @Tags Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Param body body setMutedDTO true "Mute settings"
// @Success 200 {object} domain.Response
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/{id}/mute [post]
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
// @Summary List conversation messages
// @Description Keyset-paginated message window. Exactly one of before/after/around may be set; none returns the latest page. Each message's reactions map is populated (Step 11 serialization); media_ids are returned raw for the client to resolve.
// @Tags Conversations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Param limit query int false "Page size (default 50, max 100)"
// @Param before query string false "Message UUID: return messages before this one"
// @Param after query string false "Message UUID: return messages after this one"
// @Param around query string false "Message UUID: return messages around this one"
// @Success 200 {object} domain.Response{data=[]domain.ConversationMessage}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/{id}/messages [get]
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
// @Summary Send conversation message
// @Description Channels require the sender to be a conversation admin (or platform admin / PermConversationsManage). media_ids must reference media rows already presigned via POST /media/presign with model_type=conversation, model_id=<this conversation>; otherwise the send is rejected. mentions must reference conversation members (non-members and self are silently dropped).
// @Tags Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Param body body domain.SendConversationMessageDTO true "Message data"
// @Success 201 {object} domain.Response{data=domain.ConversationMessage}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/{id}/messages [post]
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
// @Summary List pinned messages
// @Tags Conversations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Success 200 {object} domain.Response{data=[]domain.ConversationMessage}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/{id}/pins [get]
func (h *Handler) ListPinned(c *gin.Context) {
	msgs, err := h.svc.ListPinned(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, msgs)
}

// EditMessage updates a message's content. Sender or a conversation admin.
// @Summary Edit message
// @Tags Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param messageId path string true "Message UUID"
// @Param body body domain.UpdateConversationMessageDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.ConversationMessage}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/messages/{messageId} [patch]
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
// @Summary Delete message
// @Tags Conversations
// @Produce json
// @Security BearerAuth
// @Param messageId path string true "Message UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/messages/{messageId} [delete]
func (h *Handler) DeleteMessage(c *gin.Context) {
	if err := h.svc.DeleteMessage(c.Request.Context(), httpx.UUIDParam(c, "messageId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// ToggleReaction adds/removes the caller's reaction emoji on a message.
// @Summary Toggle message reaction
// @Description Adds the caller's emoji reaction if not present, removes it if already present. Returns the message with a fresh reaction count map.
// @Tags Conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param messageId path string true "Message UUID"
// @Param body body domain.ToggleConversationReactionDTO true "Emoji"
// @Success 200 {object} domain.Response{data=domain.ConversationMessage}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/messages/{messageId}/reactions [post]
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
// @Summary Pin message
// @Tags Conversations
// @Produce json
// @Security BearerAuth
// @Param messageId path string true "Message UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/messages/{messageId}/pin [post]
func (h *Handler) Pin(c *gin.Context) {
	if err := h.svc.PinMessage(c.Request.Context(), httpx.UUIDParam(c, "messageId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// Unpin unpins a message. Conversation-manage authz.
// @Summary Unpin message
// @Tags Conversations
// @Produce json
// @Security BearerAuth
// @Param messageId path string true "Message UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/messages/{messageId}/unpin [post]
func (h *Handler) Unpin(c *gin.Context) {
	if err := h.svc.UnpinMessage(c.Request.Context(), httpx.UUIDParam(c, "messageId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// searchLimit parses the optional ?limit= query param, returning 0 (service
// default) if absent/invalid.
func searchLimit(c *gin.Context) int {
	if v := c.Query("limit"); v != "" {
		if n, err := parsePositiveInt(v); err == nil {
			return n
		}
	}
	return 0
}

// Search performs a global ranked full-text search across every conversation
// the caller is a member of.
// @Summary Search messages (global)
// @Description Postgres full-text search, ranked, across every conversation the caller is a member of (org-scoped). Requires q of at least 3 characters.
// @Tags Conversations
// @Produce json
// @Security BearerAuth
// @Param q query string true "Search query (min 3 characters)"
// @Param limit query int false "Max results (service default applies if omitted)"
// @Success 200 {object} domain.Response{data=[]domain.ConversationMessage}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/search [get]
func (h *Handler) Search(c *gin.Context) {
	q := c.Query("q")
	msgs, err := h.svc.Search(c.Request.Context(), q, searchLimit(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, msgs)
}

// SearchDirectory lists active org users (minus the caller) for chat discovery.
// @Summary Search user directory (chat discovery)
// @Description Member-safe people search for starting DMs / resolving mentions. Returns id, name, username only.
// @Tags Conversations
// @Produce json
// @Security BearerAuth
// @Param search query string false "Substring match on username/name"
// @Success 200 {object} domain.Response{data=[]domain.DirectoryUser}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/directory [get]
func (h *Handler) SearchDirectory(c *gin.Context) {
	q := strings.TrimSpace(c.Query("search"))
	users, err := h.svc.SearchDirectory(c.Request.Context(), q, searchLimit(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, users)
}

// GetDirectoryUser resolves one active org user by exact username.
// @Summary Resolve directory user by username
// @Tags Conversations
// @Produce json
// @Security BearerAuth
// @Param username path string true "Username (without leading @)"
// @Success 200 {object} domain.Response{data=domain.DirectoryUser}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/directory/{username} [get]
func (h *Handler) GetDirectoryUser(c *gin.Context) {
	username := strings.TrimSpace(c.Param("username"))
	u, err := h.svc.GetDirectoryUser(c.Request.Context(), username)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, u)
}

// maxPresenceIDs caps how many user ids a single presence lookup may request,
// bounding the per-id org-membership checks the service performs.
const maxPresenceIDs = 100

// parsePresenceIDs parses the comma-separated ?user_ids= list into deduped
// UUIDs, silently skipping blank/unparseable entries and truncating at
// maxPresenceIDs (matching the tolerant query parsing used elsewhere).
func parsePresenceIDs(raw string) []uuid.UUID {
	if raw == "" {
		return nil
	}
	seen := make(map[uuid.UUID]bool)
	ids := make([]uuid.UUID, 0)
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		id, err := uuid.Parse(part)
		if err != nil || seen[id] {
			continue
		}
		seen[id] = true
		ids = append(ids, id)
		if len(ids) == maxPresenceIDs {
			break
		}
	}
	return ids
}

// GetPresence returns the online/last-seen status for a batch of users.
// @Summary Batch user presence
// @Description Returns online/last-seen status for the requested users, filtered to the caller's organization. Pass a comma-separated ?user_ids= list (capped at 100; unknown or cross-org ids are omitted from the response).
// @Tags Conversations
// @Produce json
// @Security BearerAuth
// @Param user_ids query string true "Comma-separated user UUIDs (max 100)"
// @Success 200 {object} domain.Response{data=map[string]domain.PresenceStatus}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/presence [get]
func (h *Handler) GetPresence(c *gin.Context) {
	ids := parsePresenceIDs(c.Query("user_ids"))
	statuses, err := h.svc.Presence(c.Request.Context(), ids)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, statuses)
}

// SearchInConv performs an in-conversation nav search.
// @Summary Search messages (in conversation)
// @Description ILIKE substring nav search within a single conversation, gated on membership.
// @Tags Conversations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation UUID"
// @Param q query string true "Search query"
// @Param limit query int false "Max results (service default applies if omitted)"
// @Success 200 {object} domain.Response{data=[]domain.ConversationMessage}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /conversations/{id}/search [get]
func (h *Handler) SearchInConv(c *gin.Context) {
	q := c.Query("q")
	msgs, err := h.svc.SearchInConversation(c.Request.Context(), httpx.UUIDParam(c, "id"), q, searchLimit(c))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, msgs)
}
