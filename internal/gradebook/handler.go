package gradebook

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

type Handler struct {
	svc domain.GradebookService
}

func NewHandler(svc domain.GradebookService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")
	columnIDParam := httpx.RequireUUIDParam("columnId")

	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/classes/:id/gradebook", idParam, h.GetMatrix)
		authed.POST("/classes/:id/gradebook/columns", perm(domain.PermClassesUpdate), idParam, h.CreateColumn)
		authed.PUT("/classes/:id/gradebook/columns/:columnId", perm(domain.PermClassesUpdate), idParam, columnIDParam, h.UpdateColumn)
		authed.DELETE("/classes/:id/gradebook/columns/:columnId", perm(domain.PermClassesUpdate), idParam, columnIDParam, h.DeleteColumn)
		authed.POST("/classes/:id/gradebook/columns/:columnId/cells", perm(domain.PermClassesUpdate), idParam, columnIDParam, h.UpsertCell)
	}
}

// GetMatrix returns the full gradebook grid for a class.
// @Summary Get gradebook matrix
// @Tags Gradebook
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Success 200 {object} domain.Response{data=domain.GradebookMatrix}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/{id}/gradebook [get]
func (h *Handler) GetMatrix(c *gin.Context) {
	matrix, err := h.svc.GetMatrix(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, matrix)
}

// CreateColumn creates a gradebook column for a class.
// @Summary Create gradebook column
// @Tags Gradebook
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Param body body domain.CreateGradebookColumnDTO true "Column data"
// @Success 201 {object} domain.Response{data=domain.GradebookColumn}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/{id}/gradebook/columns [post]
func (h *Handler) CreateColumn(c *gin.Context) {
	var dto domain.CreateGradebookColumnDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	col, err := h.svc.CreateColumn(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, col)
}

// UpdateColumn updates a gradebook column.
// @Summary Update gradebook column
// @Tags Gradebook
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Param columnId path string true "Column UUID"
// @Param body body domain.UpdateGradebookColumnDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.GradebookColumn}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/{id}/gradebook/columns/{columnId} [put]
func (h *Handler) UpdateColumn(c *gin.Context) {
	var dto domain.UpdateGradebookColumnDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	col, err := h.svc.UpdateColumn(c.Request.Context(), httpx.UUIDParam(c, "columnId"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, col)
}

// DeleteColumn deletes a gradebook column and all its cells.
// @Summary Delete gradebook column
// @Tags Gradebook
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Param columnId path string true "Column UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/{id}/gradebook/columns/{columnId} [delete]
func (h *Handler) DeleteColumn(c *gin.Context) {
	if err := h.svc.DeleteColumn(c.Request.Context(), httpx.UUIDParam(c, "columnId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// UpsertCell sets or updates a gradebook cell value for a student (manual columns only).
// @Summary Upsert gradebook cell
// @Tags Gradebook
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Param columnId path string true "Column UUID"
// @Param body body domain.UpsertGradebookCellDTO true "Cell data"
// @Success 200 {object} domain.Response{data=domain.GradebookCell}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/{id}/gradebook/columns/{columnId}/cells [post]
func (h *Handler) UpsertCell(c *gin.Context) {
	var dto domain.UpsertGradebookCellDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	cell, err := h.svc.UpsertCell(
		c.Request.Context(),
		httpx.UUIDParam(c, "id"),
		httpx.UUIDParam(c, "columnId"),
		dto,
	)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, cell)
}
