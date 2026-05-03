package media

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

type Handler struct {
	svc domain.MediaService
}

func NewHandler(svc domain.MediaService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")

	authed := rg.Group("", authMiddleware)
	{
		authed.POST("/media/presign", perm(domain.PermMediaCreate), h.PresignUpload)
		authed.GET("/media/:id", perm(domain.PermMediaView), idParam, h.Get)
		authed.DELETE("/media/:id", perm(domain.PermMediaDelete), idParam, h.Delete)
		authed.GET("/media", perm(domain.PermMediaView), h.ListByModel)
	}
}

// PresignUpload creates a media record and returns a presigned S3 upload URL.
// @Summary Get presigned upload URL
// @Description Creates a media record and returns a presigned PUT URL for direct upload to S3/RustFS.
// @Tags Media
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.PresignUploadDTO true "Upload details"
// @Success 201 {object} domain.Response{data=domain.PresignUploadResponse}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /media/presign [post]
func (h *Handler) PresignUpload(c *gin.Context) {
	var dto domain.PresignUploadDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	resp, err := h.svc.PresignUpload(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, resp)
}

// Get returns a media record by ID.
// @Summary Get media
// @Tags Media
// @Produce json
// @Security BearerAuth
// @Param id path string true "Media UUID"
// @Success 200 {object} domain.Response{data=domain.Media}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /media/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	m, err := h.svc.GetByID(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, m)
}

// Delete removes a media record (admin/staff only).
// @Summary Delete media
// @Tags Media
// @Produce json
// @Security BearerAuth
// @Param id path string true "Media UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /media/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// ListByModel returns media for a given model type + ID.
// @Summary List media by model
// @Tags Media
// @Produce json
// @Security BearerAuth
// @Param model_type query string true "Model type (e.g. users, classes)"
// @Param model_id query string true "Model UUID"
// @Param collection query string false "Collection name filter"
// @Success 200 {object} domain.Response{data=[]domain.Media}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Router /media [get]
func (h *Handler) ListByModel(c *gin.Context) {
	modelType := c.Query("model_type")
	modelIDStr := c.Query("model_id")
	collection := c.Query("collection")

	if modelType == "" || modelIDStr == "" {
		_ = c.Error(domain.NewValidationError(map[string]string{
			"model_type": "required",
			"model_id":   "required",
		}))
		return
	}

	modelID, err := uuid.Parse(modelIDStr)
	if err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{
			"model_id": "must be a valid UUID",
		}))
		return
	}

	items, err := h.svc.ListByModel(c.Request.Context(), modelType, modelID, collection)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, items)
}
