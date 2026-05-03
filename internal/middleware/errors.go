package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/4H1R/zoora/internal/domain"
)

// ErrorHandler is terminal middleware. Handlers attach errors via c.Error(err)
// and return; this middleware maps the last attached error to an HTTP status
// and a standardized JSON body. If a response was already written (legacy
// handlers using domain.ErrorResponse directly), it is left untouched.
func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}
		if c.Writer.Written() {
			return
		}

		err := c.Errors.Last().Err
		status, code := mapError(err)
		body := &domain.ErrorBody{Code: code, Message: err.Error()}

		var ve *domain.ValidationError
		if errors.As(err, &ve) && len(ve.Fields) > 0 {
			body.Fields = ve.Fields
		}

		if status >= http.StatusInternalServerError {
			logger.Error("request error",
				"err", err,
				"method", c.Request.Method,
				"path", c.Request.URL.Path,
			)
			body.Message = "internal server error"
		}

		c.AbortWithStatusJSON(status, domain.Response{Success: false, Error: body})
	}
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN"
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, "UNAUTHORIZED"
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, "CONFLICT"
	case errors.Is(err, domain.ErrValidation):
		return http.StatusBadRequest, "VALIDATION_ERROR"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR"
	}
}
