package gradebook

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

// columnsListConfig is the handler-owned white-list for GET
// /classes/:id/gradebook/columns. Only columns in these slices can be searched
// or ordered by the client; anything else falls back to defaults.
var columnsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"title"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "title", "order_index"},
	DefaultOrderBy:      "order_index",
	DefaultOrderDir:     "asc",
}

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
		authed.GET("/gradebook/me", perm(domain.PermGradebookViewOwn), h.GetMine)
		authed.GET("/classes/:id/gradebook", perm(domain.PermGradebookView), idParam, h.GetMatrix)
		authed.GET("/classes/:id/gradebook/columns", perm(domain.PermGradebookView), idParam, h.ListColumns)
		authed.POST("/classes/:id/gradebook/columns", perm(domain.PermGradebookCreate), idParam, h.CreateColumn)
		authed.PUT("/classes/:id/gradebook/columns/:columnId", perm(domain.PermGradebookUpdate), idParam, columnIDParam, h.UpdateColumn)
		authed.DELETE("/classes/:id/gradebook/columns/:columnId", perm(domain.PermGradebookDelete), idParam, columnIDParam, h.DeleteColumn)
		authed.POST("/classes/:id/gradebook/columns/:columnId/cells", perm(domain.PermGradebookUpdate), idParam, columnIDParam, h.UpsertCell)
	}
}

// GetMine returns the caller's own report card across their classes.
// @Summary Get my grades
// @Tags Gradebook
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=domain.MyGradebook}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Router /gradebook/me [get]
func (h *Handler) GetMine(c *gin.Context) {
	rc, err := h.svc.GetMine(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, rc)
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

// ListColumns returns paginated gradebook columns for a class.
// @Summary List gradebook columns
// @Description Search matches substrings of: title. Orderable fields: created_at, updated_at, title, order_index. Filters: type.
// @Tags Gradebook
// @Produce json
// @Security BearerAuth
// @Param id path string true "Class UUID"
// @Param type query string false "Filter by column type" Enums(auto_attendance,auto_practice,auto_quiz,manual_grade,manual_attendance,manual_text)
// @Param search query string false "Substring match on title"
// @Param order_by query string false "One of: created_at, updated_at, title, order_index"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.GradebookColumn}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /classes/{id}/gradebook/columns [get]
func (h *Handler) ListColumns(c *gin.Context) {
	var q domain.ListGradebookColumnsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	q.ListParams = listparams.Bind(c, columnsListConfig)
	cols, total, err := h.svc.ListColumns(c.Request.Context(), httpx.UUIDParam(c, "id"), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(cols, total, q.ListParams))
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
