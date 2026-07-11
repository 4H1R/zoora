package tutorials

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
)

type Handler struct {
	svc domain.TutorialService
}

func NewHandler(svc domain.TutorialService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/tutorials", h.List)
	}
}

// List returns all published tutorials in curated order (no pagination — the
// client filters/searches the full set in-browser).
// @Summary List tutorials
// @Tags Tutorials
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=[]domain.Tutorial}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Router /tutorials [get]
func (h *Handler) List(c *gin.Context) {
	items, err := h.svc.ListPublished(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, items)
}
