package customfields

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/auth"
	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

type Handler struct {
	svc domain.CustomFieldService
}

func NewHandler(svc domain.CustomFieldService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")

	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/custom-field-definitions", perm(domain.PermCustomFieldsManage), h.ListDefinitions)
		authed.POST("/custom-field-definitions", perm(domain.PermCustomFieldsManage), h.CreateDefinition)
		authed.PATCH("/custom-field-definitions/:id", perm(domain.PermCustomFieldsManage), idParam, h.UpdateDefinition)
		authed.DELETE("/custom-field-definitions/:id", perm(domain.PermCustomFieldsManage), idParam, h.ArchiveDefinition)

		authed.PATCH("/users/:id/custom-fields", perm(domain.PermUsersUpdate), idParam, h.SetUserValues)
		authed.GET("/users/:id/custom-fields",
			auth.RequireSelfOrPermission(domain.PermUsersView, domain.PermUsersViewAny, "id"), idParam, h.GetUserValues)
	}
}

// @Summary List custom field definitions
// @Tags CustomFields
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=[]domain.UserCustomFieldDefinition}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /custom-field-definitions [get]
func (h *Handler) ListDefinitions(c *gin.Context) {
	defs, err := h.svc.ListDefinitions(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, defs)
}

// @Summary Create custom field definition
// @Tags CustomFields
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateCustomFieldDefinitionDTO true "Definition"
// @Success 201 {object} domain.Response{data=domain.UserCustomFieldDefinition}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 422 {object} domain.Response{error=domain.ErrorBody}
// @Router /custom-field-definitions [post]
func (h *Handler) CreateDefinition(c *gin.Context) {
	var dto domain.CreateCustomFieldDefinitionDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	def, err := h.svc.CreateDefinition(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, def)
}

// @Summary Update custom field definition
// @Tags CustomFields
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Definition UUID"
// @Param body body domain.UpdateCustomFieldDefinitionDTO true "Patch"
// @Success 200 {object} domain.Response{data=domain.UserCustomFieldDefinition}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Router /custom-field-definitions/{id} [patch]
func (h *Handler) UpdateDefinition(c *gin.Context) {
	var dto domain.UpdateCustomFieldDefinitionDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	def, err := h.svc.UpdateDefinition(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, def)
}

// @Summary Archive custom field definition
// @Tags CustomFields
// @Produce json
// @Security BearerAuth
// @Param id path string true "Definition UUID"
// @Success 204 "No Content"
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /custom-field-definitions/{id} [delete]
func (h *Handler) ArchiveDefinition(c *gin.Context) {
	if err := h.svc.ArchiveDefinition(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Set a user's custom field values (partial merge)
// @Tags CustomFields
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Param body body domain.SetUserCustomFieldsDTO true "Values"
// @Success 200 {object} domain.Response{data=object}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Router /users/{id}/custom-fields [patch]
func (h *Handler) SetUserValues(c *gin.Context) {
	var dto domain.SetUserCustomFieldsDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	values, err := h.svc.SetUserValues(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, values)
}

// @Summary Get a user's visible custom field values
// @Tags CustomFields
// @Produce json
// @Security BearerAuth
// @Param id path string true "User UUID"
// @Success 200 {object} domain.Response{data=[]domain.VisibleCustomField}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /users/{id}/custom-fields [get]
func (h *Handler) GetUserValues(c *gin.Context) {
	fields, err := h.svc.GetVisibleUserValues(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, fields)
}
