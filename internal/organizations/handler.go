package organizations

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

type Handler struct {
	svc domain.OrganizationService
}

func NewHandler(svc domain.OrganizationService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")
	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/organizations/:id", idParam, h.Get)
		authed.PUT("/organizations/:id", perm(domain.PermOrganizationsUpdate), idParam, h.Update)
	}
}

// Get returns an organization by ID.
// @Summary Get organization by ID
// @Tags Organizations
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization UUID"
// @Success 200 {object} domain.Response{data=domain.Organization}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /organizations/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	org, err := h.svc.GetByID(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, org)
}

// Update updates an organization by ID.
// @Summary Update organization
// @Tags Organizations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization UUID"
// @Param body body domain.UpdateOrganizationDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.Organization}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /organizations/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	var dto domain.UpdateOrganizationDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	org, err := h.svc.Update(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, org)
}
