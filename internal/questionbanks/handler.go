package questionbanks

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
	"github.com/4H1R/zoora/internal/platform/listparams"
)

var banksListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"name", "description"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "name"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

var questionsListConfig = domain.ListConfig{
	AllowedSearchFields: []string{"text"},
	AllowedOrderFields:  []string{"created_at", "updated_at", "type"},
	DefaultOrderBy:      "created_at",
	DefaultOrderDir:     "desc",
}

type Handler struct {
	svc domain.QuestionBankService
}

func NewHandler(svc domain.QuestionBankService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc, perm func(domain.PermissionName) gin.HandlerFunc) {
	idParam := httpx.RequireUUIDParam("id")
	questionIDParam := httpx.RequireUUIDParam("questionId")

	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/question-banks", perm(domain.PermQuestionBanksView), h.List)
		authed.POST("/question-banks", perm(domain.PermQuestionBanksCreate), h.Create)
		authed.GET("/question-banks/:id", perm(domain.PermQuestionBanksView), idParam, h.Get)
		authed.PUT("/question-banks/:id", perm(domain.PermQuestionBanksUpdate), idParam, h.Update)
		authed.DELETE("/question-banks/:id", perm(domain.PermQuestionBanksDelete), idParam, h.Delete)

		authed.POST("/question-banks/redeem", perm(domain.PermQuestionBanksCreate), h.RedeemShareCode)
		authed.GET("/question-banks/share-codes/:code", perm(domain.PermQuestionBanksCreate), h.PreviewShareCode)
		authed.POST("/question-banks/:id/share-code", perm(domain.PermQuestionBanksUpdate), idParam, h.GenerateShareCode)
		authed.GET("/question-banks/:id/share-code", perm(domain.PermQuestionBanksUpdate), idParam, h.GetShareCode)
		authed.DELETE("/question-banks/:id/share-code", perm(domain.PermQuestionBanksUpdate), idParam, h.RevokeShareCode)

		authed.GET("/question-banks/:id/questions", perm(domain.PermQuestionBanksView), idParam, h.ListQuestions)
		authed.POST("/question-banks/:id/questions", perm(domain.PermQuestionBanksUpdate), idParam, h.CreateQuestion)
		authed.GET("/question-banks/questions/:questionId", perm(domain.PermQuestionBanksView), questionIDParam, h.GetQuestion)
		authed.PUT("/question-banks/questions/:questionId", perm(domain.PermQuestionBanksUpdate), questionIDParam, h.UpdateQuestion)
		authed.DELETE("/question-banks/questions/:questionId", perm(domain.PermQuestionBanksDelete), questionIDParam, h.DeleteQuestion)
	}
}

// @Summary List question banks (scoped by org)
// @Description Returns question banks filtered by caller's organization. Admins see all. Search matches substrings of: name, description. Orderable fields: created_at, updated_at, name.
// @Tags QuestionBanks
// @Produce json
// @Security BearerAuth
// @Param search query string false "Substring match on name/description"
// @Param order_by query string false "One of: created_at, updated_at, name"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.QuestionBank}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /question-banks [get]
func (h *Handler) List(c *gin.Context) {
	p := listparams.Bind(c, banksListConfig)
	banks, total, err := h.svc.List(c.Request.Context(), p)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(banks, total, p))
}

// @Summary Create question bank
// @Tags QuestionBanks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.CreateQuestionBankDTO true "Question bank data"
// @Success 201 {object} domain.Response{data=domain.QuestionBank}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /question-banks [post]
func (h *Handler) Create(c *gin.Context) {
	var dto domain.CreateQuestionBankDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	bank, err := h.svc.Create(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, bank)
}

// @Summary Get question bank
// @Tags QuestionBanks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question bank UUID"
// @Success 200 {object} domain.Response{data=domain.QuestionBank}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /question-banks/{id} [get]
func (h *Handler) Get(c *gin.Context) {
	bank, err := h.svc.GetByID(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, bank)
}

// @Summary Update question bank
// @Tags QuestionBanks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question bank UUID"
// @Param body body domain.UpdateQuestionBankDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.QuestionBank}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /question-banks/{id} [put]
func (h *Handler) Update(c *gin.Context) {
	var dto domain.UpdateQuestionBankDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	bank, err := h.svc.Update(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, bank)
}

// @Summary Delete question bank
// @Tags QuestionBanks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question bank UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /question-banks/{id} [delete]
func (h *Handler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// @Summary Generate share code for a question bank
// @Description Creates a new multi-use share code for the bank, revoking any previous one. Optional expiry in days; omitted = never expires. Redeeming the code clones the bank into the redeemer's organization.
// @Tags QuestionBanks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question bank UUID"
// @Param body body domain.GenerateShareCodeDTO true "Share code options"
// @Success 201 {object} domain.Response{data=domain.QuestionBankShareCode}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /question-banks/{id}/share-code [post]
func (h *Handler) GenerateShareCode(c *gin.Context) {
	var dto domain.GenerateShareCodeDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	code, err := h.svc.GenerateShareCode(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, code)
}

// @Summary Get active share code of a question bank
// @Tags QuestionBanks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question bank UUID"
// @Success 200 {object} domain.Response{data=domain.QuestionBankShareCode}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /question-banks/{id}/share-code [get]
func (h *Handler) GetShareCode(c *gin.Context) {
	code, err := h.svc.GetShareCode(c.Request.Context(), httpx.UUIDParam(c, "id"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, code)
}

// @Summary Revoke share code of a question bank
// @Tags QuestionBanks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question bank UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /question-banks/{id}/share-code [delete]
func (h *Handler) RevokeShareCode(c *gin.Context) {
	if err := h.svc.RevokeShareCode(c.Request.Context(), httpx.UUIDParam(c, "id")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// @Summary Preview a share code
// @Description Shows the shared bank's name, description and question count so the redeemer can decide before cloning. Invalid, expired, or revoked codes return a generic validation error.
// @Tags QuestionBanks
// @Produce json
// @Security BearerAuth
// @Param code path string true "Share code"
// @Success 200 {object} domain.Response{data=domain.ShareCodePreview}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /question-banks/share-codes/{code} [get]
func (h *Handler) PreviewShareCode(c *gin.Context) {
	preview, err := h.svc.PreviewShareCode(c.Request.Context(), c.Param("code"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, preview)
}

// @Summary Redeem a share code
// @Description Clones the shared bank (questions + media) into the caller's organization as an independent copy. Returns the new bank with status 'copying'; the copy completes in the background and the bank flips to 'ready'.
// @Tags QuestionBanks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body domain.RedeemShareCodeDTO true "Share code"
// @Success 201 {object} domain.Response{data=domain.QuestionBank}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Router /question-banks/redeem [post]
func (h *Handler) RedeemShareCode(c *gin.Context) {
	var dto domain.RedeemShareCodeDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	bank, err := h.svc.RedeemShareCode(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, bank)
}

// @Summary List questions in a bank
// @Description Returns questions filtered by optional type. Search matches substrings of: text. Orderable fields: created_at, updated_at, type.
// @Tags QuestionBanks
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question bank UUID"
// @Param type query string false "Filter by question type (descriptive, short_answer, choice)"
// @Param search query string false "Substring match on text"
// @Param order_by query string false "One of: created_at, updated_at, type"
// @Param order_dir query string false "asc or desc"
// @Param page query int false "1-based page number"
// @Success 200 {object} domain.Response{data=domain.PaginatedData{items=[]domain.Question}}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /question-banks/{id}/questions [get]
func (h *Handler) ListQuestions(c *gin.Context) {
	var q domain.ListQuestionsQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"query": err.Error()}))
		return
	}
	q.ListParams = listparams.Bind(c, questionsListConfig)
	questions, total, err := h.svc.ListQuestions(c.Request.Context(), httpx.UUIDParam(c, "id"), q)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, domain.NewPaginatedFromParams(questions, total, q.ListParams))
}

// @Summary Create question in bank
// @Tags QuestionBanks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Question bank UUID"
// @Param body body domain.CreateQuestionDTO true "Question data"
// @Success 201 {object} domain.Response{data=domain.Question}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /question-banks/{id}/questions [post]
func (h *Handler) CreateQuestion(c *gin.Context) {
	var dto domain.CreateQuestionDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	question, err := h.svc.CreateQuestion(c.Request.Context(), httpx.UUIDParam(c, "id"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, question)
}

// @Summary Get question
// @Tags QuestionBanks
// @Produce json
// @Security BearerAuth
// @Param questionId path string true "Question UUID"
// @Success 200 {object} domain.Response{data=domain.Question}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /question-banks/questions/{questionId} [get]
func (h *Handler) GetQuestion(c *gin.Context) {
	question, err := h.svc.GetQuestion(c.Request.Context(), httpx.UUIDParam(c, "questionId"))
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, question)
}

// @Summary Update question
// @Tags QuestionBanks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param questionId path string true "Question UUID"
// @Param body body domain.UpdateQuestionDTO true "Update data"
// @Success 200 {object} domain.Response{data=domain.Question}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /question-banks/questions/{questionId} [put]
func (h *Handler) UpdateQuestion(c *gin.Context) {
	var dto domain.UpdateQuestionDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}
	question, err := h.svc.UpdateQuestion(c.Request.Context(), httpx.UUIDParam(c, "questionId"), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, question)
}

// @Summary Delete question
// @Tags QuestionBanks
// @Produce json
// @Security BearerAuth
// @Param questionId path string true "Question UUID"
// @Success 200 {object} domain.Response
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /question-banks/questions/{questionId} [delete]
func (h *Handler) DeleteQuestion(c *gin.Context) {
	if err := h.svc.DeleteQuestion(c.Request.Context(), httpx.UUIDParam(c, "questionId")); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
