package orgsettings

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

type Handler struct {
	svc domain.OrganizationSettingsService
}

func NewHandler(svc domain.OrganizationSettingsService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")
	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/organizations/:id/settings", perm(domain.PermOrganizationsUpdate), idParam, h.Get)
		authed.PUT("/organizations/:id/settings", perm(domain.PermOrganizationsUpdate), idParam, h.Update)
	}
}

// @Summary Get organization settings
// @Tags Organizations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization UUID"
// @Success 200 {object} domain.Response{data=domain.OrganizationSettings}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /organizations/{id}/settings [get]
func (h *Handler) Get(c *gin.Context) {
	settings, err := h.svc.Get(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, settings)
}

// @Summary Update organization settings
// @Tags Organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization UUID"
// @Param body body domain.UpdateOrganizationSettingsDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.OrganizationSettings}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /organizations/{id}/settings [put]
func (h *Handler) Update(c *gin.Context) {
	var dto domain.UpdateOrganizationSettingsDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	settings, err := h.svc.Update(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, settings)
}
