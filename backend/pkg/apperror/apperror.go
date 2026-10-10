// Package apperror defines errors that carry an HTTP status and a user-facing message.
package apperror

import (
	"errors"
	"net/http"
)

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type AppError struct {
	Status  int
	Message string
	Errors  []FieldError
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Unwrap() error { return e.Err }

func New(status int, message string) *AppError {
	return &AppError{Status: status, Message: message}
}

func BadRequest(message string) *AppError   { return New(http.StatusBadRequest, message) }
func Unauthorized(message string) *AppError { return New(http.StatusUnauthorized, message) }
func Forbidden(message string) *AppError    { return New(http.StatusForbidden, message) }
func NotFound(message string) *AppError     { return New(http.StatusNotFound, message) }
func Conflict(message string) *AppError     { return New(http.StatusConflict, message) }

func Unprocessable(message string) *AppError {
	return New(http.StatusUnprocessableEntity, message)
}

func Validation(errs []FieldError) *AppError {
	return &AppError{Status: http.StatusUnprocessableEntity, Message: "Data yang dikirim tidak valid", Errors: errs}
}

func Internal(err error) *AppError {
	return &AppError{Status: http.StatusInternalServerError, Message: "Terjadi kesalahan pada server", Err: err}
}

// From converts any error into an *AppError, defaulting to 500.
func From(err error) *AppError {
	var ae *AppError
	if errors.As(err, &ae) {
		return ae
	}
	return Internal(err)
}
