package tutorials

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

type AdminHandler struct {
	svc domain.TutorialService
}

func NewAdminHandler(svc domain.TutorialService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	idParam := httpx.RequireUUIDParam("id")
	group.GET("/tutorials", h.List)
	group.POST("/tutorials", h.Create)
	// Static siblings of /:id — registered alongside it, mirroring
	// /changelog/media/presign.
	group.GET("/tutorials/aparat-oembed", h.AparatOEmbed)
	group.PUT("/tutorials/reorder", h.Reorder)
	group.GET("/tutorials/:id", idParam, h.Get)
	group.PUT("/tutorials/:id", idParam, h.Update)
	group.POST("/tutorials/:id/publish", idParam, h.Publish)
	group.POST("/tutorials/:id/unpublish", idParam, h.Unpublish)
	group.DELETE("/tutorials/:id", idParam, h.Delete)
}

// @Summary [Admin] List tutorials (incl. drafts)
// @Tags Admin/Tutorials
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=[]domain.Tutorial}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/tutorials [get]
func (h *AdminHandler) List(c *gin.Context) {
	items, err := h.svc.AdminList(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, items)
}

// @Summary [Admin] Create tutorial (draft)
// @Tags Admin/Tutorials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateTutorialDTO true "Tutorial"
// @Success 201 {object} domain.Response{data=domain.Tutorial}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/tutorials [post]
func (h *AdminHandler) Create(c *gin.Context) {
	var dto domain.CreateTutorialDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	tu, err := h.svc.AdminCreate(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, tu)
}

// @Summary [Admin] Get tutorial
// @Tags Admin/Tutorials
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tutorial UUID"
// @Success 200 {object} domain.Response{data=domain.Tutorial}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/tutorials/{id} [get]
func (h *AdminHandler) Get(c *gin.Context) {
	tu, err := h.svc.AdminGet(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, tu)
}

// @Summary [Admin] Update tutorial
// @Tags Admin/Tutorials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tutorial UUID"
// @Param body body domain.UpdateTutorialDTO true "Fields"
// @Success 200 {object} domain.Response{data=domain.Tutorial}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/tutorials/{id} [put]
func (h *AdminHandler) Update(c *gin.Context) {
	var dto domain.UpdateTutorialDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	tu, err := h.svc.AdminUpdate(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, tu)
}

// @Summary [Admin] Publish tutorial
// @Tags Admin/Tutorials
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tutorial UUID"
// @Success 200 {object} domain.Response{data=domain.Tutorial}
// @Router /admin/tutorials/{id}/publish [post]
func (h *AdminHandler) Publish(c *gin.Context) {
	tu, err := h.svc.AdminPublish(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, tu)
}

// @Summary [Admin] Unpublish tutorial
// @Tags Admin/Tutorials
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tutorial UUID"
// @Success 200 {object} domain.Response{data=domain.Tutorial}
// @Router /admin/tutorials/{id}/unpublish [post]
func (h *AdminHandler) Unpublish(c *gin.Context) {
	tu, err := h.svc.AdminUnpublish(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, tu)
}

// @Summary [Admin] Delete tutorial
// @Tags Admin/Tutorials
// @Produce json
// @Security BearerAuth
// @Param id path string true "Tutorial UUID"
// @Success 200 {object} domain.Response
// @Router /admin/tutorials/{id} [delete]
func (h *AdminHandler) Delete(c *gin.Context) {
	if err := h.svc.AdminDelete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// @Summary [Admin] Resolve Aparat oEmbed (title + thumbnail)
// @Tags Admin/Tutorials
// @Produce json
// @Security BearerAuth
// @Param hash query string true "Aparat video hash"
// @Success 200 {object} domain.Response{data=domain.AparatOEmbedResponse}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/tutorials/aparat-oembed [get]
func (h *AdminHandler) AparatOEmbed(c *gin.Context) {
	meta, err := h.svc.AdminAparatOEmbed(c.Request.Context(), c.Query("hash"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, meta)
}

// @Summary [Admin] Reorder tutorials
// @Tags Admin/Tutorials
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.ReorderTutorialsDTO true "Ordered ids"
// @Success 200 {object} domain.Response
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/tutorials/reorder [put]
func (h *AdminHandler) Reorder(c *gin.Context) {
	var dto domain.ReorderTutorialsDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.svc.AdminReorder(c.Request.Context(), dto); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
