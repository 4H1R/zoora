package domain

import (
	"errors"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorBody `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

type PaginatedData struct {
	Items    any   `json:"items"`
	Total    int64 `json:"total"`
	Offset   int   `json:"offset,omitempty"`
	Limit    int   `json:"limit,omitempty"`
	Page     int   `json:"page,omitempty"`
	PageSize int   `json:"page_size,omitempty"`
}

// NewPaginatedFromParams builds a PaginatedData using the page-based shape
// (Page/PageSize) from a ListParams. Use this for endpoints adopting the
// standardized list pattern; endpoints still on offset/limit can keep
// setting those fields directly.
func NewPaginatedFromParams(items any, total int64, p ListParams) PaginatedData {
	return PaginatedData{
		Items:    items,
		Total:    total,
		Page:     p.Page,
		PageSize: p.Limit(),
	}
}

func SuccessResponse(c *gin.Context, status int, data any) {
	c.JSON(status, Response{
		Success: true,
		Data:    data,
	})
}

func ErrorResponse(c *gin.Context, err error) {
	status, code := MapError(err)
	body := &ErrorBody{Code: code, Message: err.Error()}

	var ve *ValidationError
	if errors.As(err, &ve) && len(ve.Fields) > 0 {
		body.Fields = ve.Fields
	}
	c.JSON(status, Response{Success: false, Error: body})
}
