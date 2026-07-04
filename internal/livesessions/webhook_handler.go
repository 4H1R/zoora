package livesessions

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	lkproto "github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/webhook"

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

	// Egress lifecycle events carry EgressInfo instead of (or alongside) Room.
	if event.Event == webhook.EventEgressEnded && event.EgressInfo != nil {
		if err := h.svc.OnEgressEnded(c.Request.Context(), egressResult(event.EgressInfo)); err != nil {
			h.logger.Error("handling egress webhook", "egress_id", event.EgressInfo.EgressId, "error", err)
		}
		c.Status(http.StatusOK)
		return
	}

	if event.Room == nil || event.Room.Name == "" {
		c.Status(http.StatusOK)
		return
	}
	identity := ""
	if event.Participant != nil {
		identity = event.Participant.Identity
	}
	if err := h.svc.OnLiveKitEvent(c.Request.Context(), event.Event, event.Room.Name, identity); err != nil {
		h.logger.Error("handling livekit webhook", "event", event.Event, "room", event.Room.Name, "error", err)
	}
	c.Status(http.StatusOK)
}

// egressResult translates LiveKit's EgressInfo into the domain result: failure
// status plus size/duration from the first file result when present.
func egressResult(info *lkproto.EgressInfo) domain.EgressResult {
	res := domain.EgressResult{
		EgressID: info.EgressId,
		Failed: info.Status == lkproto.EgressStatus_EGRESS_FAILED ||
			info.Status == lkproto.EgressStatus_EGRESS_ABORTED,
	}
	if files := info.GetFileResults(); len(files) > 0 && files[0] != nil {
		res.SizeBytes = files[0].Size
		res.Duration = time.Duration(files[0].Duration)
	}
	return res
}
