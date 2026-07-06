package connectors

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/4H1R/zoora/internal/domain"
)

type Handler struct {
	svc domain.ConnectorService
}

func NewHandler(svc domain.ConnectorService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, authMiddleware gin.HandlerFunc) {
	authed := rg.Group("", authMiddleware)
	{
		authed.GET("/connectors", h.List)
		authed.POST("/connectors/telegram/link", h.LinkTelegram)
		authed.POST("/connectors/bale/link", h.LinkBale)
		authed.POST("/connectors/sms/request-otp", h.RequestSMSOTP)
		authed.POST("/connectors/sms/verify-otp", h.VerifySMSOTP)
		authed.POST("/connectors/push", h.RegisterPush)
		authed.PATCH("/connectors/:id", h.Update)
		authed.DELETE("/connectors/:id", h.Unlink)
	}
}

// List returns the caller's linked connectors.
// @Summary List my connectors
// @Tags Connectors
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=[]domain.UserConnector}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Router /connectors [get]
func (h *Handler) List(c *gin.Context) {
	items, err := h.svc.List(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, items)
}

// LinkTelegram issues a one-time Telegram bot-linking deep link.
// @Summary Start Telegram link
// @Tags Connectors
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=domain.LinkTokenResponse}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Router /connectors/telegram/link [post]
func (h *Handler) LinkTelegram(c *gin.Context) {
	resp, err := h.svc.CreateLinkToken(c.Request.Context(), domain.ConnectorTelegram)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, resp)
}

// LinkBale issues a one-time Bale bot-linking deep link.
// @Summary Start Bale link
// @Tags Connectors
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.Response{data=domain.LinkTokenResponse}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Router /connectors/bale/link [post]
func (h *Handler) LinkBale(c *gin.Context) {
	resp, err := h.svc.CreateLinkToken(c.Request.Context(), domain.ConnectorBale)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, resp)
}

// RequestSMSOTP sends a verification code to the given phone.
// @Summary Request SMS OTP
// @Tags Connectors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.RequestSMSOTPDTO true "Phone"
// @Success 200 {object} domain.Response
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 429 {object} domain.Response{error=domain.ErrorBody}
// @Router /connectors/sms/request-otp [post]
func (h *Handler) RequestSMSOTP(c *gin.Context) {
	var dto domain.RequestSMSOTPDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.svc.RequestSMSOTP(c.Request.Context(), dto); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// VerifySMSOTP verifies the code and links the phone as an SMS connector.
// @Summary Verify SMS OTP
// @Tags Connectors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.VerifySMSOTPDTO true "Code"
// @Success 200 {object} domain.Response
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Router /connectors/sms/verify-otp [post]
func (h *Handler) VerifySMSOTP(c *gin.Context) {
	var dto domain.VerifySMSOTPDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		_ = c.Error(err)
		return
	}
	if err := h.svc.VerifySMSOTP(c.Request.Context(), dto); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}

// RegisterPush registers an FCM device token as a push connector.
// @Summary Register push token
// @Tags Connectors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body domain.RegisterPushTokenDTO true "FCM token"
// @Success 201 {object} domain.Response{data=domain.UserConnector}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Router /connectors/push [post]
func (h *Handler) RegisterPush(c *gin.Context) {
	var dto domain.RegisterPushTokenDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		_ = c.Error(err)
		return
	}
	conn, err := h.svc.RegisterPushToken(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusCreated, conn)
}

// Update enables or disables one of the caller's connectors.
// @Summary Update connector
// @Tags Connectors
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Connector ID"
// @Param request body domain.UpdateConnectorDTO true "Enabled flag"
// @Success 200 {object} domain.Response{data=domain.UserConnector}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /connectors/{id} [patch]
func (h *Handler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"id": "must be a valid UUID"}))
		return
	}
	var dto domain.UpdateConnectorDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		_ = c.Error(err)
		return
	}
	conn, err := h.svc.Update(c.Request.Context(), id, dto)
	if err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, conn)
}

// Unlink deletes one of the caller's connectors.
// @Summary Unlink connector
// @Tags Connectors
// @Produce json
// @Security BearerAuth
// @Param id path string true "Connector ID"
// @Success 200 {object} domain.Response
// @Failure 403 {object} domain.Response{error=domain.ErrorBody}
// @Failure 404 {object} domain.Response{error=domain.ErrorBody}
// @Router /connectors/{id} [delete]
func (h *Handler) Unlink(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		_ = c.Error(domain.NewValidationError(map[string]string{"id": "must be a valid UUID"}))
		return
	}
	if err := h.svc.Unlink(c.Request.Context(), id); err != nil {
		_ = c.Error(err)
		return
	}
	domain.SuccessResponse(c, http.StatusOK, nil)
}
