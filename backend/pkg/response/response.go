// Package response writes the uniform JSON envelope used by every endpoint.
package response

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/posumkm/backend/pkg/apperror"
	"github.com/posumkm/backend/pkg/pagination"
)

type Body struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    any              `json:"data,omitempty"`
	Meta    *pagination.Meta `json:"meta,omitempty"`
}

type ErrorBody struct {
	Success bool                  `json:"success"`
	Message string                `json:"message"`
	Errors  []apperror.FieldError `json:"errors"`
}

func OK(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Body{Success: true, Message: message, Data: data})
}

func Created(c *gin.Context, message string, data any) {
	c.JSON(http.StatusCreated, Body{Success: true, Message: message, Data: data})
}

func Paginated(c *gin.Context, message string, data any, meta pagination.Meta) {
	c.JSON(http.StatusOK, Body{Success: true, Message: message, Data: data, Meta: &meta})
}

// Error converts err into the error envelope. 5xx causes are logged, never exposed.
func Error(c *gin.Context, err error) {
	ae := apperror.From(err)
	if ae.Status >= http.StatusInternalServerError {
		slog.ErrorContext(c.Request.Context(), "internal error",
			"error", err, "path", c.FullPath(), "method", c.Request.Method)
	}
	errs := ae.Errors
	if errs == nil {
		errs = []apperror.FieldError{}
	}
	c.AbortWithStatusJSON(ae.Status, ErrorBody{Success: false, Message: ae.Message, Errors: errs})
}
