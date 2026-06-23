package attendance

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

// adminAttendanceListConfig white-lists search/order for GET /admin/attendance.
var adminAttendanceListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"remarks"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "status"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

type AdminHandler struct {
	svc domain.AttendanceService
}

func NewAdminHandler(svc domain.AttendanceService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	idParam := httpx.RequireUUIDParam("id")

	group.GET("/attendance", h.List)
	group.PUT("/attendance/:id", idParam, h.Update)
	group.DELETE("/attendance/:id", idParam, h.HardDelete)
}

// @Summary [Admin] List attendance
// @Description Cross-org list. Search matches substrings of: remarks. Orderable fields: created_at, updated_at, status. Filters: status, is_auto_marked, user_id, class_id, class_session_id, organization_id.
// @Tags Admin/Attendance
// @Produce json
// @Security BearerAuth
// @Param status query string false "Filter by status" Enums(present,absent,late,excused)
// @Param is_auto_marked query bool false "Filter auto-marked vs manual"
// @Param user_id query string false "Filter by user UUID"
// @Param class_id query string false "Filter by class UUID"
// @Param class_session_id query string false "Filter by class session UUID"
// @Param organization_id query string false "Filter by organization UUID"
// @Param search query string false "Substring match on remarks"
// @Param order_by query string false "One of: created_at, updated_at, status"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Attendance}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/attendance [get]
func (h *AdminHandler) List(c *gin.Context) {
	var q domain.AdminListAttendanceQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{
		"user_id":          &q.UserID,
		"class_id":         &q.ClassID,
		"class_session_id": &q.ClassSessionID,
		"organization_id":  &q.OrganizationID,
	}); err != nil {
		_ = c.Error(err)
		return
	}
	q.ListParams = listparams.Bind(c, adminAttendanceListConfig)
	items, total, err := h.svc.AdminList(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(items, total, q.ListParams))
}

// @Summary [Admin] Update attendance
// @Tags Admin/Attendance
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Attendance UUID"
// @Param body body domain.AdminUpdateAttendanceDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.Attendance}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/attendance/{id} [put]
func (h *AdminHandler) Update(c *gin.Context) {
	var dto domain.AdminUpdateAttendanceDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	a, err := h.svc.AdminUpdate(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, a)
}

// @Summary [Admin] Delete attendance
// @Tags Admin/Attendance
// @Produce json
// @Security BearerAuth
// @Param id path string true "Attendance UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/attendance/{id} [delete]
func (h *AdminHandler) HardDelete(c *gin.Context) {
	if err := h.svc.AdminHardDelete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
