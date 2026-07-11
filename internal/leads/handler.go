package leads

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

// Handler serves the public, unauthenticated lead-capture endpoint. It is
// host-agnostic (submitted from the apex marketing site, which carries no
// tenant context) and gated only by rate-limiting + a honeypot in the service.
type Handler struct {
	svc domain.LeadService
}

func NewHandler(svc domain.LeadService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes mounts the public submit endpoint. Callers pass a rate-limit
// middleware to bound abuse.
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, mws ...gin.HandlerFunc) {
	leads := rg.Group("/leads", mws...)
	{
		leads.POST("", h.Submit)
	}
}

// Submit records a public "Get started" lead.
// @Summary Submit a sales lead
// @Description Public, unauthenticated. Captures a marketing contact from the pricing page. Includes a honeypot field ("website") that must be left empty. Rate-limited.
// @Tags Leads
// @Accept json
// @Produce json
// @Param body body domain.CreateLeadDTO true "Lead details"
// @Success 201 {object} domain.Response
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 429 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /leads [post]
func (h *Handler) Submit(c *gin.Context) {
	var dto domain.CreateLeadDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	if _, err := h.svc.Submit(c.Request.Context(), dto); err != nil {
		_ = c.Error(err)
		return
	}
	// Always 201, even for a dropped honeypot hit — bots learn nothing.
	domain.SuccessResponse(c, http.StatusCreated, nil)
}
