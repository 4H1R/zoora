package changelog

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var adminListConfig = domain.ListConfig{
	AllowedOrderFields: []string{"created_at", "published_at"},
	DefaultOrderBy:     "created_at",
	DefaultOrderDir:    "desc",
}

type AdminHandler struct {
	svc domain.ChangelogService
}

func NewAdminHandler(svc domain.ChangelogService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	idParam := httpx.RequireUUIDParam("id")
	group.GET("/changelog", h.List)
	group.POST("/changelog", h.Create)
	group.POST("/changelog/media/presign", h.PresignMedia)
	group.GET("/changelog/:id", idParam, h.Get)
	group.PUT("/changelog/:id", idParam, h.Update)
	group.POST("/changelog/:id/publish", idParam, h.Publish)
	group.POST("/changelog/:id/unpublish", idParam, h.Unpublish)
	group.DELETE("/changelog/:id", idParam, h.Delete)
}

// @Summary [Admin] List changelog (incl. drafts)
// @Tags Admin/Changelog
// @Produce json
// @Security BearerAuth
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.ChangelogEntry}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/changelog [get]
func (h *AdminHandler) List(c *gin.Context) {
	p := listparams.Bind(c, adminListConfig)
	items, total, err := h.svc.AdminList(c.Request.Context(), p)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(items, total, p))
}

// @Summary [Admin] Create changelog entry (draft)
// @Tags Admin/Changelog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateChangelogDTO true "Entry"
// @Success 201 {object} domain.Response{data=domain.ChangelogEntry}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/changelog [post]
func (h *AdminHandler) Create(c *gin.Context) {
	var dto domain.CreateChangelogDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	e, err := h.svc.AdminCreate(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, e)
}

// @Summary [Admin] Get changelog entry
// @Tags Admin/Changelog
// @Produce json
// @Security BearerAuth
// @Param id path string true "Entry UUID"
// @Success 200 {object} domain.Response{data=domain.ChangelogEntry}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/changelog/{id} [get]
func (h *AdminHandler) Get(c *gin.Context) {
	e, err := h.svc.AdminGet(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, e)
}

// @Summary [Admin] Update changelog entry
// @Tags Admin/Changelog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Entry UUID"
// @Param body body domain.UpdateChangelogDTO true "Fields"
// @Success 200 {object} domain.Response{data=domain.ChangelogEntry}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/changelog/{id} [put]
func (h *AdminHandler) Update(c *gin.Context) {
	var dto domain.UpdateChangelogDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	e, err := h.svc.AdminUpdate(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, e)
}

// @Summary [Admin] Publish changelog entry
// @Tags Admin/Changelog
// @Produce json
// @Security BearerAuth
// @Param id path string true "Entry UUID"
// @Success 200 {object} domain.Response{data=domain.ChangelogEntry}
// @Router /admin/changelog/{id}/publish [post]
func (h *AdminHandler) Publish(c *gin.Context) {
	e, err := h.svc.AdminPublish(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, e)
}

// @Summary [Admin] Unpublish changelog entry
// @Tags Admin/Changelog
// @Produce json
// @Security BearerAuth
// @Param id path string true "Entry UUID"
// @Success 200 {object} domain.Response{data=domain.ChangelogEntry}
// @Router /admin/changelog/{id}/unpublish [post]
func (h *AdminHandler) Unpublish(c *gin.Context) {
	e, err := h.svc.AdminUnpublish(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, e)
}

// @Summary [Admin] Delete changelog entry
// @Tags Admin/Changelog
// @Produce json
// @Security BearerAuth
// @Param id path string true "Entry UUID"
// @Success 200 {object} domain.Response
// @Router /admin/changelog/{id} [delete]
func (h *AdminHandler) Delete(c *gin.Context) {
	if err := h.svc.AdminDelete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// @Summary [Admin] Presign changelog media upload (public bucket)
// @Tags Admin/Changelog
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.ChangelogMediaPresignDTO true "Media"
// @Success 200 {object} domain.Response{data=domain.ChangelogMediaPresignResponse}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/changelog/media/presign [post]
func (h *AdminHandler) PresignMedia(c *gin.Context) {
	var dto domain.ChangelogMediaPresignDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	resp, err := h.svc.AdminPresignMedia(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, resp)
}
