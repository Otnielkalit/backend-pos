package apperror

import (
	"errors"
	"net/http"
)

// AppError is the standard error type returned by usecase and repository layers.
// Handler layer translates this into the appropriate HTTP response using HTTPStatus.
type AppError struct {
	Code       string // Machine-readable error code, e.g. "PRODUCT_NOT_FOUND"
	Message    string // Human-readable message for API consumer
	HTTPStatus int    // HTTP status code for the handler to use
}

func (e *AppError) Error() string {
	return e.Message
}

// --- Constructors ---

func NewBadRequest(message string) *AppError {
	return &AppError{Code: "BAD_REQUEST", Message: message, HTTPStatus: http.StatusBadRequest}
}

func NewUnauthorized(message string) *AppError {
	return &AppError{Code: "UNAUTHORIZED", Message: message, HTTPStatus: http.StatusUnauthorized}
}

func NewForbidden(message string) *AppError {
	return &AppError{Code: "FORBIDDEN", Message: message, HTTPStatus: http.StatusForbidden}
}

func NewNotFound(message string) *AppError {
	return &AppError{Code: "NOT_FOUND", Message: message, HTTPStatus: http.StatusNotFound}
}

func NewConflict(message string) *AppError {
	return &AppError{Code: "CONFLICT", Message: message, HTTPStatus: http.StatusConflict}
}

func NewUnprocessable(message string) *AppError {
	return &AppError{Code: "UNPROCESSABLE_ENTITY", Message: message, HTTPStatus: http.StatusUnprocessableEntity}
}

func NewInternal(message string) *AppError {
	return &AppError{Code: "INTERNAL_SERVER_ERROR", Message: message, HTTPStatus: http.StatusInternalServerError}
}

// New creates an AppError with a custom code.
func New(code, message string, httpStatus int) *AppError {
	return &AppError{Code: code, Message: message, HTTPStatus: httpStatus}
}

// --- Helper ---

// As extracts an *AppError from err chain (works with fmt.Errorf wrapping).
// Returns nil if err is not (or does not wrap) an *AppError.
func As(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}
