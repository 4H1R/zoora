package notifications

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/auth"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var inboxListConfig = domain.ListConfig{
	AllowedOrderFields: []string{"created_at"},
	DefaultOrderBy:     "created_at",
	DefaultOrderDir:    "desc",
}

type Handler struct {
	svc domain.NotificationService
}

func NewHandler(svc domain.NotificationService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/notifications", h.ListInbox)
		authed.GET("/notifications/status", h.Status)
		authed.POST("/notifications/:id/read", h.MarkRead)
		authed.POST("/notifications/mark-all-read", h.MarkAllRead)

		sender := authed.Group("", auth.RequireAnyPermission(
			domain.PermNotificationsSend, domain.PermNotificationsSendAny))
		{
			sender.POST("/notifications", h.Send)
			sender.GET("/notifications/sent", h.ListSent)
			sender.GET("/notifications/:id/report", h.Report)
		}
	}
}

// Send creates a notification and enqueues audience fan-out.
// @Summary Send a notification
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.SendNotificationDTO true "Notification"
// @Success 201 {object} domain.Response{data=domain.Notification}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 429 {object} domain.Response{error=domain.ErrorBody}
// @Router /notifications [post]
func (h *Handler) Send(c *gin.Context) {
	var dto domain.SendNotificationDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		_ = c.Error(err)
		return
	}
	n, err := h.svc.Send(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, n)
}

// ListInbox returns the caller's notifications, newest first.
// @Summary List my notifications
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Param page query int false "1-based page number"
// @Param page_size query int false "Items per page"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.NotificationInboxItem}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Router /notifications [get]
func (h *Handler) ListInbox(c *gin.Context) {
	p := listparams.Bind(c, inboxListConfig)
	items, total, err := h.svc.ListInbox(c.Request.Context(), p)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(items, total, p))
}

// Status returns the caller's unread count (bell badge).
// @Summary Notification status
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=domain.NotificationStatus}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Router /notifications/status [get]
func (h *Handler) Status(c *gin.Context) {
	st, err := h.svc.Status(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, st)
}

// MarkRead marks one notification read for the caller.
// @Summary Mark notification read
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} domain.Response
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /notifications/{id}/read [post]
func (h *Handler) MarkRead(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"id": "must be a valid UUID"}))
		return
	}
	if err := h.svc.MarkRead(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// MarkAllRead marks all the caller's notifications read.
// @Summary Mark all notifications read
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response
// @Router /notifications/mark-all-read [post]
func (h *Handler) MarkAllRead(c *gin.Context) {
	if err := h.svc.MarkAllRead(c.Request.Context()); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// Report returns the per-channel delivery report for a sent notification.
// @Summary Notification delivery report
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Param id path string true "Notification ID"
// @Success 200 {object} domain.Response{data=domain.NotificationDeliveryReport}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /notifications/{id}/report [get]
func (h *Handler) Report(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"id": "must be a valid UUID"}))
		return
	}
	report, err := h.svc.Report(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, report)
}

// ListSent returns notifications the caller has sent, newest first.
// @Summary List sent notifications
// @Tags Notifications
// @Produce json
// @Security BearerAuth
// @Param page query int false "1-based page number"
// @Param page_size query int false "Items per page"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Notification}}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /notifications/sent [get]
func (h *Handler) ListSent(c *gin.Context) {
	p := listparams.Bind(c, inboxListConfig)
	items, total, err := h.svc.ListSent(c.Request.Context(), p)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(items, total, p))
}
