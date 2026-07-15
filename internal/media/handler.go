package media

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
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
		authed.GET("/media/:id/download-url", perm(domain.PermMediaView), idParam, h.PresignDownload)
		authed.DELETE("/media/:id", perm(domain.PermMediaDelete), idParam, h.Delete)
		authed.GET("/media", perm(domain.PermMediaView), h.ListByModel)
		authed.GET("/files/folders", perm(domain.PermMediaViewAny), h.ListFolders)
		authed.GET("/files", perm(domain.PermMediaViewAny), h.ListFiles)
		authed.GET("/files/owners", perm(domain.PermMediaViewAny), h.ListOwners)
		authed.GET("/files/owners/:kind/files", perm(domain.PermMediaViewAny), h.ListOwnerFiles)
	}
}

// filesListConfig white-lists search/order for the org files list.
var filesListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"name", "file_name"},
	AllowedOrderFields:  []string{"created_at", "size", "name"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

// ownersListConfig drives pagination for the "by owner" view. Owners are
// sorted by size server-side, so no order white-list is exposed.
var ownersListConfig = domain.ListConfig{DefaultOrderBy: "", DefaultOrderDir: "desc"}

// ownerFilesListConfig white-lists search/order for one owner's file drill-down.
var ownerFilesListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"name"},
	AllowedOrderFields:  []string{"created_at", "size", "name"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

// ownerKinds white-lists the owner-kind path segment.
var ownerKinds = map[string]bool{
	domain.MediaOwnerClass:        true,
	domain.MediaOwnerQuestionBank: true,
	domain.MediaOwnerConversation: true,
	domain.MediaOwnerShared:       true,
	domain.MediaOwnerOther:        true,
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

// shareExpiries whitelists the client-selectable share-link lifetimes.
// 7d is the SigV4 presign maximum.
var shareExpiries = map[string]time.Duration{
	"1h":  time.Hour,
	"24h": 24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
}

// PresignDownload returns a presigned GET URL for a media object.
// @Summary Get presigned download URL
// @Description Returns a presigned URL that grants temporary read access to the underlying S3 object.
// @Tags Media
// @Produce json
// @Security BearerAuth
// @Param id path string true "Media UUID"
// @Param expiry query string false "Link lifetime" Enums(1h, 24h, 7d) default(1h)
// @Success 200 {object} domain.Response{data=domain.PresignDownloadResponse}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /media/{id}/download-url [get]
func (h *Handler) PresignDownload(c *gin.Context) {
	expiryParam := c.DefaultQuery("expiry", "1h")
	expiry, ok := shareExpiries[expiryParam]
	if !ok {
		_ = c.Error(domain.NewValidationError(map[string]string{"expiry": "must be one of 1h, 24h, 7d"}))
		return
	}
	resp, err := h.svc.PresignDownload(c.Request.Context(), httpx.UUIDParam(c, "id"), expiry)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, resp)
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

// ListFolders returns the org's media grouped into folders by model type.
// @Summary List org file folders
// @Description Aggregates the caller's org media by model_type for the files page. Requires media:view_any.
// @Tags Media
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=[]domain.MediaFolder}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /files/folders [get]
func (h *Handler) ListFolders(c *gin.Context) {
	folders, err := h.svc.ListFolders(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, folders)
}

// ListFiles pages one folder (model_type) of the caller's org.
// @Summary List org files in a folder
// @Description Paginated, searchable list of the caller's org media of one model_type. Requires media:view_any.
// @Tags Media
// @Produce json
// @Security BearerAuth
// @Param model_type query string true "Folder model type (e.g. organization, live_room)"
// @Param page query int false "Page (1-based)"
// @Param page_size query int false "Page size"
// @Param search query string false "Search in name/file_name"
// @Param order_by query string false "Order field" Enums(created_at, size, name)
// @Param order_dir query string false "Order direction" Enums(asc, desc)
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Media}}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /files [get]
func (h *Handler) ListFiles(c *gin.Context) {
	modelType := c.Query("model_type")
	if modelType == "" {
		_ = c.Error(domain.NewValidationError(map[string]string{"model_type": "required"}))
		return
	}
	p := listparams.Bind(c, filesListConfig)
	items, total, err := h.svc.ListFiles(c.Request.Context(), modelType, p)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(items, total, p))
}

// ListOwners returns the org's storage grouped by owning entity.
// @Summary List org storage by owner
// @Description Aggregates the caller's org media + recordings by resolved owner (class, question_bank, conversation, shared, other), size-sorted, with a storage quota header. Requires media:view_any.
// @Tags Media
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page (1-based)"
// @Param page_size query int false "Page size"
// @Success 200 {object} domain.Response{data=domain.MediaOwnersResponse}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /files/owners [get]
func (h *Handler) ListOwners(c *gin.Context) {
	p := listparams.Bind(c, ownersListConfig)
	resp, err := h.svc.ListOwners(c.Request.Context(), p)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, resp)
}

// ListOwnerFiles pages the files under one owner (media + recordings).
// @Summary List files under one owner
// @Description Paginated, searchable files for a single owner. Class owners include read-only recordings. ownerID is omitted for shared/other buckets. Requires media:view_any.
// @Tags Media
// @Produce json
// @Security BearerAuth
// @Param kind path string true "Owner kind" Enums(class, question_bank, conversation, shared, other)
// @Param owner_id query string false "Owner UUID (omit for shared/other)"
// @Param page query int false "Page (1-based)"
// @Param page_size query int false "Page size"
// @Param search query string false "Search in file name"
// @Param order_by query string false "Order field" Enums(created_at, size, name)
// @Param order_dir query string false "Order direction" Enums(asc, desc)
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.OwnerFile}}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /files/owners/{kind}/files [get]
func (h *Handler) ListOwnerFiles(c *gin.Context) {
	kind := c.Param("kind")
	if !ownerKinds[kind] {
		_ = c.Error(domain.NewValidationError(map[string]string{"kind": "unknown owner kind"}))
		return
	}
	var ownerID *uuid.UUID
	if raw := c.Query("owner_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			_ = c.Error(domain.NewValidationError(map[string]string{"owner_id": "must be a valid UUID"}))
			return
		}
		ownerID = &id
	}
	p := listparams.Bind(c, ownerFilesListConfig)
	items, total, err := h.svc.ListOwnerFiles(c.Request.Context(), kind, ownerID, p)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(items, total, p))
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
