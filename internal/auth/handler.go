package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
	"github.com/4H1R/zoora/internal/platform/httpx"
)

type Handler struct {
	svc domain.AuthService
}

func NewHandler(svc domain.AuthService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, mws ...gin.HandlerFunc) {
	auth := rg.Group("/auth", mws...)
	{
		auth.POST("/login", h.Login)
	}
}

// Login authenticates a user with username and password.
// @Summary Login
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body domain.LoginDTO true "Credentials"
// @Success 200 {object} domain.Response{data=object{user=domain.User,token=string}}
// @Failure 400 {object} domain.Response{error=domain.ErrorBody}
// @Failure 401 {object} domain.Response{error=domain.ErrorBody}
// @Failure 500 {object} domain.Response{error=domain.ErrorBody}
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var dto domain.LoginDTO
	if err := httpx.Bind(c, &dto); err != nil {
		_ = c.Error(err)
		return
	}

	user, token, err := h.svc.Login(c.Request.Context(), dto)
	if err != nil {
		_ = c.Error(err)
		return
	}

	domain.SuccessResponse(c, http.StatusOK, gin.H{
		"user":  user,
		"token": token,
	})
}
