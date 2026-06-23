package questionbanks

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var adminBanksListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"name", "description"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "name"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

var adminQuestionsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"text"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "type"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

type AdminHandler struct {
	svc domain.QuestionBankService
}

func NewAdminHandler(svc domain.QuestionBankService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func (h *AdminHandler) RegisterAdminRoutes(group *gin.RouterGroup) {
	idParam := httpx.RequireUUIDParam("id")
	questionIDParam := httpx.RequireUUIDParam("questionId")

	group.GET("/question-banks", h.List)
	group.POST("/question-banks", h.Create)
	group.PUT("/question-banks/:id", idParam, h.Update)
	group.DELETE("/question-banks/:id", idParam, h.HardDelete)
	group.GET("/questions", h.ListQuestions)
	group.DELETE("/question-banks/questions/:questionId", questionIDParam, h.HardDeleteQuestion)
}

// @Summary [Admin] List questions
// @Description Cross-bank list. Search matches: text. Orderable: created_at, updated_at, type. Filters: bank_id, organization_id, type, include_deleted.
// @Tags Admin/QuestionBanks
// @Produce json
// @Security BearerAuth
// @Param bank_id query string false "Filter by question bank UUID"
// @Param organization_id query string false "Filter by organization UUID"
// @Param type query string false "Filter by question type (descriptive, short_answer, choice)"
// @Param include_deleted query bool false "Include soft-deleted rows"
// @Param search query string false "Substring match on text"
// @Param order_by query string false "One of: created_at, updated_at, type"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Question}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/questions [get]
func (h *AdminHandler) ListQuestions(c *gin.Context) {
	var q domain.AdminListQuestionsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{
		"bank_id":         &q.BankID,
		"organization_id": &q.OrganizationID,
	}); err != nil {
		_ = c.Error(err)
		return
	}
	q.ListParams = listparams.Bind(c, adminQuestionsListConfig)
	questions, total, err := h.svc.AdminListQuestions(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(questions, total, q.ListParams))
}

// @Summary [Admin] Create question bank
// @Tags Admin/QuestionBanks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.AdminCreateQuestionBankDTO true "Question bank data"
// @Success 201 {object} domain.Response{data=domain.QuestionBank}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/question-banks [post]
func (h *AdminHandler) Create(c *gin.Context) {
	var dto domain.AdminCreateQuestionBankDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	bank, err := h.svc.AdminCreate(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, bank)
}

// @Summary [Admin] Update question bank
// @Tags Admin/QuestionBanks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question bank UUID"
// @Param body body domain.AdminUpdateQuestionBankDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.QuestionBank}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 409 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/question-banks/{id} [put]
func (h *AdminHandler) Update(c *gin.Context) {
	var dto domain.AdminUpdateQuestionBankDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	bank, err := h.svc.AdminUpdate(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, bank)
}

// @Summary [Admin] List question banks
// @Tags Admin/QuestionBanks
// @Produce json
// @Security BearerAuth
// @Param include_deleted query bool false "Include soft-deleted rows"
// @Param search query string false "Substring match on name/description"
// @Param order_by query string false "One of: created_at, updated_at, name"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.QuestionBank}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/question-banks [get]
func (h *AdminHandler) List(c *gin.Context) {
	var q domain.AdminListQuestionBanksQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	if err := httpx.BindUUIDQueries(c, map[string]**uuid.UUID{
		"organization_id": &q.OrganizationID,
	}); err != nil {
		_ = c.Error(err)
		return
	}
	q.ListParams = listparams.Bind(c, adminBanksListConfig)
	banks, total, err := h.svc.AdminList(c.Request.Context(), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(banks, total, q.ListParams))
}

// @Summary [Admin] Hard-delete question bank
// @Tags Admin/QuestionBanks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question bank UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/question-banks/{id} [delete]
func (h *AdminHandler) HardDelete(c *gin.Context) {
	if err := h.svc.AdminHardDelete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// @Summary [Admin] Hard-delete question
// @Tags Admin/QuestionBanks
// @Produce json
// @Security BearerAuth
// @Param questionId path string true "Question UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /admin/question-banks/questions/{questionId} [delete]
func (h *AdminHandler) HardDeleteQuestion(c *gin.Context) {
	if err := h.svc.AdminHardDeleteQuestion(c.Request.Context(), httpx.UUIDParam(c, "questionId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
