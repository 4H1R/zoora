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
	svc domain.LiveRoomChatService
}

func NewHandler(svc domain.LiveRoomChatService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	chatIDParam := httpx.RequireUUIDParam("chatId")
	msgIDParam := httpx.RequireUUIDParam("messageId")

	authed := rg.Group("", authMiddleware)
	{
		authed.POST("/chats", perm(domain.PermChatsCreate), h.CreateChat)
		authed.GET("/chats", perm(domain.PermChatsView), h.ListChats)
		authed.GET("/chats/:chatId", perm(domain.PermChatsView), chatIDParam, h.GetChat)
		authed.PUT("/chats/:chatId", perm(domain.PermChatsUpdate), chatIDParam, h.UpdateChat)
		authed.DELETE("/chats/:chatId", perm(domain.PermChatsDelete), chatIDParam, h.DeleteChat)

		authed.POST("/chats/:chatId/messages", perm(domain.PermChatsWrite), chatIDParam, h.SendMessage)
		authed.GET("/chats/:chatId/messages", perm(domain.PermChatsView), chatIDParam, h.ListMessages)
		authed.GET("/chats/:chatId/messages/:messageId", perm(domain.PermChatsView), chatIDParam, msgIDParam, h.GetMessage)
		authed.PUT("/chats/:chatId/messages/:messageId", perm(domain.PermChatsWrite), chatIDParam, msgIDParam, h.UpdateMessage)
		authed.DELETE("/chats/:chatId/messages/:messageId", perm(domain.PermChatsWrite), chatIDParam, msgIDParam, h.DeleteMessage)
	}
}

// CreateChat creates a new chat backing a live room.
// @Summary Create chat
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateChatDTO true "Chat data"
// @Success 201 {object} domain.Response{data=domain.LiveRoomChat}
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
// @Success 200 {object} domain.Response{data=domain.LiveRoomChat}
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
// @Success 200 {object} domain.Response{data=domain.LiveRoomChat}
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

// ListChats returns paginated chats, optionally filtered by live room.
// @Summary List chats
// @Tags Chat
// @Produce json
// @Security BearerAuth
// @Param live_room_id query string false "Live room UUID filter"
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
	if idStr := c.Query("live_room_id"); idStr != "" {
		id, err := uuid.Parse(idStr)
		if err != nil {
			_ = c.Error(domain.NewValidationError(map[string]string{"live_room_id": "must be a valid UUID"}))
			return
		}
		q.LiveRoomID = &id
	}
	if status := c.Query("status"); status != "" {
		s := domain.LiveRoomChatStatus(status)
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

// SendMessage sends a message to a chat.
// @Summary Send message
// @Tags Chat
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param chatId path string true "Chat UUID"
// @Param body body domain.SendMessageDTO true "Message data"
// @Success 201 {object} domain.Response{data=domain.LiveRoomMessage}
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
// @Success 200 {object} domain.Response{data=domain.LiveRoomMessage}
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
// @Success 200 {object} domain.Response{data=domain.LiveRoomMessage}
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
