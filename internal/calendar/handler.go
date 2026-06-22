package calendar

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
)

// maxRangeDays bounds the query window so a single request can't scan an
// unbounded date span across four tables.
const maxRangeDays = 92

type Handler struct {
	svc domain.CalendarService
}

func NewHandler(svc domain.CalendarService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts the calendar endpoint. It is auth-gated only; the
// service scopes results by role, so no permission middleware is needed.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/calendar/events", h.List)
	}
}

// List returns the caller's schedulable events within [from, to].
// @Summary List calendar events (scoped by RBAC)
// @Description Returns live/quiz/practice/offline events the caller can see within a UTC time window. Scope: super-admins and classes:view_any holders see their organization; teachers see their own classes; students see classes they are enrolled in. Times are UTC; bucket into days client-side.
// @Tags Calendar
// @Produce json
// @Security BearerAuth
// @Param from query string true "Range start, RFC3339 (e.g. 2026-06-01T00:00:00Z)"
// @Param to query string true "Range end, RFC3339 (e.g. 2026-06-30T23:59:59Z)"
// @Success 200 {object} domain.Response{data=domain.CalendarEventsData}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /calendar/events [get]
func (h *Handler) List(c *gin.Context) {
	from, err := time.Parse(time.RFC3339, c.Query("from"))
	if err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"from": "must be RFC3339 datetime"}))
		return
	}
	to, err := time.Parse(time.RFC3339, c.Query("to"))
	if err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"to": "must be RFC3339 datetime"}))
		return
	}
	if to.Before(from) {
		_ = c.Error(domain.NewValidationError(map[string]string{"to": "must be after from"}))
		return
	}
	if to.Sub(from) > maxRangeDays*24*time.Hour {
		_ = c.Error(domain.NewValidationError(map[string]string{"to": "range too large"}))
		return
	}

	events, err := h.svc.ListEvents(c.Request.Context(), domain.CalendarRange{From: from, To: to})
	if err != nil {
		_ = c.Error(err)
		return
	}
	if events == nil {
		events = []domain.CalendarEvent{}
	}
	domain.SuccessResponse(c, http.StatusOK, domain.CalendarEventsData{Events: events})
}
