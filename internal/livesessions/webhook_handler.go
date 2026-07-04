package livesessions

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	lk "github.com/4H1R/zoora/internal/platform/livekit"
)

// WebhookHandler receives LiveKit server webhooks. It is mounted outside the
// authenticated/tenant-scoped API group: LiveKit authenticates via a signed
// Authorization header (verified by the LiveKit client), and rooms are looked
// up by their globally-unique LiveKit room name, so no tenant context applies.
type WebhookHandler struct {
	livekit *lk.Client
	svc     domain.LiveSessionService
	logger  *slog.Logger
}

func NewWebhookHandler(livekit *lk.Client, svc domain.LiveSessionService, logger *slog.Logger) *WebhookHandler {
	return &WebhookHandler{livekit: livekit, svc: svc, logger: logger}
}

// RegisterRoutes mounts the webhook receiver on the given router group. Pass a
// root, unauthenticated group (e.g. router.Group("/webhooks")).
func (h *WebhookHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/livekit", h.Handle)
}

// Handle verifies the webhook signature, decodes the event, and dispatches it
// to the service. It always answers 200 once the signature is valid so LiveKit
// does not retry on transient handling errors (the periodic sweep is the
// backstop); only a bad signature yields 401.
func (h *WebhookHandler) Handle(c *gin.Context) {
	event, err := h.livekit.ParseWebhook(c.Request)
	if err != nil {
		h.logger.Warn("rejected livekit webhook", "error", err)
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if event.Room == nil || event.Room.Name == "" {
		c.Status(http.StatusOK)
		return
	}
	if err := h.svc.OnLiveKitEvent(c.Request.Context(), event.Event, event.Room.Name); err != nil {
		h.logger.Error("handling livekit webhook", "event", event.Event, "room", event.Room.Name, "error", err)
	}
	c.Status(http.StatusOK)
}
