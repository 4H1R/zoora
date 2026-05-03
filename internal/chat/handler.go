package chat

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var chatsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"name"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "name"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

var messagesListConfig = domain.ListConfig{
	AllowedOrderFields: []string{"created_at"},
	DefaultOrderBy:     "created_at",
	DefaultOrderDir:    "desc",
}

type Handler struct {
	svc domain.ChatService
}

func NewHandler(svc domain.ChatService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	chatIDParam := httpx.RequireUUIDParam("chatId")
	msgIDParam := httpx.RequireUUIDParam("messageId")
	userIDParam := httpx.RequireUUIDParam("userId")

	authed := rg.Group("", authMiddleware)
	{
		authed.POST("/chats", perm(domain.PermChatsCreate), h.CreateChat)
		authed.GET("/chats", perm(domain.PermChatsView), h.ListChats)
		authed.GET("/chats/:chatId", perm(domain.PermChatsView), chatIDParam, h.GetChat)
		authed.PUT("/chats/:chatId", perm(domain.PermChatsUpdate), chatIDParam, h.UpdateChat)
		authed.DELETE("/chats/:chatId", perm(domain.PermChatsDelete), chatIDParam, h.DeleteChat)

		authed.POST("/chats/:chatId/members", perm(domain.PermChatsManage), chatIDParam, h.AddMember)
		authed.DELETE("/chats/:chatId/members/:userId", perm(domain.PermChatsManage), chatIDParam, userIDParam, h.RemoveMember)
		authed.GET("/chats/:chatId/members", perm(domain.PermChatsView), chatIDParam, h.ListMembers)

		authed.POST("/chats/:chatId/messages", perm(domain.PermChatsWrite), chatIDParam, h.SendMessage)
		authed.GET("/chats/:chatId/messages", perm(domain.PermChatsView), chatIDParam, h.ListMessages)
		authed.GET("/chats/:chatId/messages/:messageId", perm(domain.PermChatsView), chatIDParam, msgIDParam, h.GetMessage)
		authed.PUT("/chats/:chatId/messages/:messageId", perm(domain.PermChatsWrite), chatIDParam, msgIDParam, h.UpdateMessage)
		authed.DELETE("/chats/:chatId/messages/:messageId", perm(domain.PermChatsWrite), chatIDParam, msgIDParam, h.DeleteMessage)

		authed.POST("/chats/:chatId/messages/:messageId/reactions", perm(domain.PermChatsWrite), chatIDParam, msgIDParam, h.ToggleReaction)
	}
}

// CreateChat creates a new chat attached to a polymorphic entity.
// @Summary Create chat
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateChatDTO true "Chat data"
// @Success 201 {object} domain.Response{data=domain.Chat}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /chats [post]
func (h *Handler) CreateChat(c *gin.Context) {
	var dto domain.CreateChatDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	chat, err := h.svc.CreateChat(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, chat)
}

// GetChat returns a chat by ID.
// @Summary Get chat
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param chatId path string true "Chat UUID"
// @Success 200 {object} domain.Response{data=domain.Chat}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /chats/{chatId} [get]
func (h *Handler) GetChat(c *gin.Context) {
	chat, err := h.svc.GetChat(c.Request.Context(), httpx.UUIDParam(c, "chatId"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, chat)
}

// UpdateChat updates a chat.
// @Summary Update chat
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param chatId path string true "Chat UUID"
// @Param body body domain.UpdateChatDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.Chat}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /chats/{chatId} [put]
func (h *Handler) UpdateChat(c *gin.Context) {
	var dto domain.UpdateChatDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	chat, err := h.svc.UpdateChat(c.Request.Context(), httpx.UUIDParam(c, "chatId"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, chat)
}

// DeleteChat soft-deletes a chat.
// @Summary Delete chat
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param chatId path string true "Chat UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /chats/{chatId} [delete]
func (h *Handler) DeleteChat(c *gin.Context) {
	if err := h.svc.DeleteChat(c.Request.Context(), httpx.UUIDParam(c, "chatId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// ListChats returns paginated chats filtered by model_type and model_id.
// @Summary List chats
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param model_type query string false "Model type filter"
// @Param model_id query string false "Model UUID filter"
// @Param status query string false "Status filter"
// @Param page query int false "Page number"
// @Param search query string false "Search term"
// @Param order_by query string false "Order field"
// @Param order_dir query string false "Order direction"
// @Success 200 {object} domain.Response{data=domain.PaginatedData}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Router /chats [get]
func (h *Handler) ListChats(c *gin.Context) {
	var q domain.ListChatsQuery
	q.ModelType = c.Query("model_type")
	if idStr := c.Query("model_id"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err != nil {
			_ = c.Error(domain.NewValidationError(map[string]string{"model_id": "must be a valid UUID"}))
			return
		}
		q.ModelID = &id
	}
	if status := c.Query("status"); status != "" {
		s := domain.ChatStatus(status)
		q.Status = &s
	}
	q.ListParams = listparams.Bind(c, chatsListConfig)

	chats, total, err := h.svc.ListChats(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(chats, total, q.ListParams))
}

// AddMember adds a user to a chat.
// @Summary Add chat member
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param chatId path string true "Chat UUID"
// @Param body body domain.AddChatMemberDTO true "Member data"
// @Success 201 {object} domain.Response{data=domain.ChatMember}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Router /chats/{chatId}/members [post]
func (h *Handler) AddMember(c *gin.Context) {
	var dto domain.AddChatMemberDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	member, err := h.svc.AddMember(c.Request.Context(), httpx.UUIDParam(c, "chatId"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, member)
}

// RemoveMember removes a user from a chat.
// @Summary Remove chat member
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param chatId path string true "Chat UUID"
// @Param userId path string true "User UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /chats/{chatId}/members/{userId} [delete]
func (h *Handler) RemoveMember(c *gin.Context) {
	if err := h.svc.RemoveMember(
		c.Request.Context(),
		httpx.UUIDParam(c, "chatId"),
		httpx.UUIDParam(c, "userId"),
	); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// ListMembers returns all members of a chat.
// @Summary List chat members
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param chatId path string true "Chat UUID"
// @Success 200 {object} domain.Response{data=[]domain.ChatMember}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /chats/{chatId}/members [get]
func (h *Handler) ListMembers(c *gin.Context) {
	members, err := h.svc.ListMembers(c.Request.Context(), httpx.UUIDParam(c, "chatId"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, members)
}

// SendMessage sends a message to a chat.
// @Summary Send message
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param chatId path string true "Chat UUID"
// @Param body body domain.SendMessageDTO true "Message data"
// @Success 201 {object} domain.Response{data=domain.Message}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /chats/{chatId}/messages [post]
func (h *Handler) SendMessage(c *gin.Context) {
	var dto domain.SendMessageDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	msg, err := h.svc.SendMessage(c.Request.Context(), httpx.UUIDParam(c, "chatId"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, msg)
}

// ListMessages returns paginated messages for a chat.
// @Summary List messages
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param chatId path string true "Chat UUID"
// @Param parent_message_id query string false "Parent message UUID for threads"
// @Param page query int false "Page number"
// @Param order_by query string false "Order field"
// @Param order_dir query string false "Order direction"
// @Success 200 {object} domain.Response{data=domain.PaginatedData}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /chats/{chatId}/messages [get]
func (h *Handler) ListMessages(c *gin.Context) {
	var q domain.ListMessagesQuery
	if idStr := c.Query("parent_message_id"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err != nil {
			_ = c.Error(domain.NewValidationError(map[string]string{"parent_message_id": "must be a valid UUID"}))
			return
		}
		q.ParentMessageID = &id
	}
	q.ListParams = listparams.Bind(c, messagesListConfig)

	msgs, total, err := h.svc.ListMessages(c.Request.Context(), httpx.UUIDParam(c, "chatId"), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(msgs, total, q.ListParams))
}

// GetMessage returns a message by ID.
// @Summary Get message
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param chatId path string true "Chat UUID"
// @Param messageId path string true "Message UUID"
// @Success 200 {object} domain.Response{data=domain.Message}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /chats/{chatId}/messages/{messageId} [get]
func (h *Handler) GetMessage(c *gin.Context) {
	msg, err := h.svc.GetMessage(c.Request.Context(), httpx.UUIDParam(c, "messageId"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, msg)
}

// UpdateMessage edits a message's content.
// @Summary Update message
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param chatId path string true "Chat UUID"
// @Param messageId path string true "Message UUID"
// @Param body body domain.UpdateMessageDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.Message}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /chats/{chatId}/messages/{messageId} [put]
func (h *Handler) UpdateMessage(c *gin.Context) {
	var dto domain.UpdateMessageDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	msg, err := h.svc.UpdateMessage(c.Request.Context(), httpx.UUIDParam(c, "messageId"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, msg)
}

// DeleteMessage soft-deletes a message.
// @Summary Delete message
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param chatId path string true "Chat UUID"
// @Param messageId path string true "Message UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /chats/{chatId}/messages/{messageId} [delete]
func (h *Handler) DeleteMessage(c *gin.Context) {
	if err := h.svc.DeleteMessage(c.Request.Context(), httpx.UUIDParam(c, "messageId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// ToggleReaction adds or removes a reaction on a message.
// @Summary Toggle reaction
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param chatId path string true "Chat UUID"
// @Param messageId path string true "Message UUID"
// @Param body body domain.ToggleReactionDTO true "Reaction data"
// @Success 200 {object} domain.Response{data=domain.Message}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /chats/{chatId}/messages/{messageId}/reactions [post]
func (h *Handler) ToggleReaction(c *gin.Context) {
	var dto domain.ToggleReactionDTO
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
