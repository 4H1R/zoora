package imports

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

type Handler struct {
	svc domain.ImportService
}

func NewHandler(svc domain.ImportService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")

	authed := rg.Group("", authMiddleware)
	{
		// permission is type-dependent (users:create vs classes:create_any),
		// enforced in the service
		authed.POST("/imports", h.Create)
		authed.GET("/imports/latest", h.Latest)
		authed.GET("/imports/:id", idParam, h.Get)
		authed.GET("/imports/:id/result", idParam, h.Result)
	}
}

// @Summary Start a bulk import
// @Tags Imports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateImportJobDTO true "Import type and uploaded media id"
// @Success 201 {object} domain.Response{data=domain.ImportJob}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /imports [post]
func (h *Handler) Create(c *gin.Context) {
	var dto domain.CreateImportJobDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	job, err := h.svc.Create(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, job)
}

// @Summary Latest import job of a type (null when none)
// @Tags Imports
// @Produce json
// @Security BearerAuth
// @Param type query string true "Import type" Enums(users, classes)
// @Success 200 {object} domain.Response{data=domain.ImportJob}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /imports/latest [get]
func (h *Handler) Latest(c *gin.Context) {
	t := domain.ImportType(c.Query("type"))
	if t != domain.ImportTypeUsers && t != domain.ImportTypeClasses {
		_ = c.Error(domain.NewValidationError(map[string]string{"type": "must be users or classes"}))
		return
	}
	job, err := h.svc.Latest(c.Request.Context(), t)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, job)
}

// @Summary Get import job status
// @Tags Imports
// @Produce json
// @Security BearerAuth
// @Param id path string true "Import job ID"
// @Success 200 {object} domain.Response{data=domain.ImportJob}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /imports/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	id := httpx.UUIDParam(c, "id")
	job, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, job)
}

// @Summary Download import result file (contains generated passwords; expires 24h after completion)
// @Tags Imports
// @Produce application/vnd.openxmlformats-officedocument.spreadsheetml.sheet
// @Security BearerAuth
// @Param id path string true "Import job ID"
// @Success 200 {file} file
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /imports/{id}/result [get]
func (h *Handler) Result(c *gin.Context) {
	id := httpx.UUIDParam(c, "id")
	data, err := h.svc.Result(c.Request.Context(), id)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.Header("Content-Disposition", `attachment; filename="import-result-`+id.String()+`.xlsx"`)
	c.Data(http.StatusOK, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", data)
}
