package domain

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorBody  `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type PaginatedData struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Offset   int         `json:"offset,omitempty"`
	Limit    int         `json:"limit,omitempty"`
	Page     int         `json:"page,omitempty"`
	PageSize int         `json:"page_size,omitempty"`
}

// NewPaginatedFromParams builds a PaginatedData using the page-based shape
// (Page/PageSize) from a ListParams. Use this for endpoints adopting the
// standardized list pattern; endpoints still on offset/limit can keep
// setting those fields directly.
func NewPaginatedFromParams(items interface{}, total int64, p ListParams) PaginatedData {
	return PaginatedData{
		Items:    items,
		Total:    total,
		Page:     p.Page,
		PageSize: p.Limit(),
	}
}

func SuccessResponse(c *gin.Context, status int, data interface{}) {
	c.JSON(status, Response{
		Success: true,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, err error) {
	status, code := mapError(err)
	body := &ErrorBody{Code: code, Message: err.Error()}

	var ve *ValidationError
	if errors.As(err, &ve) && len(ve.Fields) > 0 {
		body.Fields = ve.Fields
	}
	c.JSON(status, Response{Success: false, Error: body})
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, ErrUserDisabled):
		return http.StatusForbidden, "USER_DISABLED"
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN"
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, "UNAUTHORIZED"
	case errors.Is(err, ErrConflict):
		return http.StatusConflict, "CONFLICT"
	case errors.Is(err, ErrValidation):
		return http.StatusBadRequest, "VALIDATION_ERROR"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR"
	}
}
