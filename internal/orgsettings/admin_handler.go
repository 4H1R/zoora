package orgsettings

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

type AdminHandler struct {
	svc domain.OrganizationSettingsService
}

func NewAdminHandler(svc domain.OrganizationSettingsService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	group.PATCH("/organizations/:id/settings", httpx.RequireUUIDParam("id"), h.Update)
}

// Update mutates superAdmin-only org settings (the SMS delivery gate).
// @Summary [Admin] Update org settings (SMS gate)
// @Tags Admin/Organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization UUID"
// @Param body body domain.AdminUpdateOrgSettingsDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.OrganizationSettings}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/organizations/{id}/settings [patch]
func (h *AdminHandler) Update(c *gin.Context) {
	var dto domain.AdminUpdateOrgSettingsDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	settings, err := h.svc.AdminUpdate(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, settings)
}
