package changelog

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var publicListConfig = domain.ListConfig{
	AllowedOrderFields: []string{"published_at"},
	DefaultOrderBy:     "published_at",
	DefaultOrderDir:    "desc",
}

type Handler struct {
	svc domain.ChangelogService
}

func NewHandler(svc domain.ChangelogService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/changelog", h.List)
		authed.GET("/changelog/status", h.Status)
		authed.POST("/changelog/mark-seen", h.MarkSeen)
	}
}

// List returns published changelog entries, newest first, paginated.
// @Summary List changelog
// @Tags Changelog
// @Produce json
// @Security BearerAuth
// @Param page query int false "1-based page number"
// @Param page_size query int false "Items per page"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.ChangelogEntry}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Router /changelog [get]
func (h *Handler) List(c *gin.Context) {
	p := listparams.Bind(c, publicListConfig)
	items, total, err := h.svc.ListPublished(c.Request.Context(), p)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(items, total, p))
}

// Status returns current version + unseen count + latest major (for the badge/modal).
// @Summary Changelog status
// @Tags Changelog
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=domain.ChangelogStatus}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Router /changelog/status [get]
func (h *Handler) Status(c *gin.Context) {
	st, err := h.svc.Status(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, st)
}

// MarkSeen advances the caller's last-seen marker to now.
// @Summary Mark changelog seen
// @Tags Changelog
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Router /changelog/mark-seen [post]
func (h *Handler) MarkSeen(c *gin.Context) {
	if err := h.svc.MarkSeen(c.Request.Context()); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
