package response

import (
	"net/http"

	"github.com/Otnielkalit/backend-pos/internal/shared/apperror"
	"github.com/gin-gonic/gin"
)

// Success is the standard envelope for successful API responses.
//
//	{
//	  "success": true,
//	  "message": "Product created successfully",
//	  "data": { ... },
//	  "meta": { "page": 1, "total": 50 }   // optional, for paginated lists
//	}
type Success struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Meta    interface{} `json:"meta,omitempty"`
}

// Error is the standard envelope for error API responses.
//
//	{
//	  "success": false,
//	  "message": "Product not found",
//	  "error": { "code": "NOT_FOUND", "details": null }
//	}
type Error struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Err     ErrorDetail `json:"error"`
}

// ErrorDetail contains the machine-readable error code and optional details.
type ErrorDetail struct {
	Code    string      `json:"code"`
	Details interface{} `json:"details,omitempty"`
}

// PaginationMeta is used by offset-based paginated list endpoints
// (products, employees, categories).
type PaginationMeta struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

// CursorMeta is used by cursor-based paginated list endpoints
// (transactions, stock_adjustments, audit_logs).
type CursorMeta struct {
	NextCursor string `json:"next_cursor"` // empty string means no more pages
	HasMore    bool   `json:"has_more"`
}

// --- Helpers ---

// OK sends a 200 response with data and no meta.
func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Success{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// OKWithMeta sends a 200 response with data and pagination meta.
func OKWithMeta(c *gin.Context, message string, data interface{}, meta interface{}) {
	c.JSON(http.StatusOK, Success{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// Created sends a 201 response.
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Success{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Err translates an error into the standard error response.
// It checks if err is an *apperror.AppError for the status code and code field.
// Falls back to 500 Internal Server Error for unrecognized errors.
func Err(c *gin.Context, err error) {
	appErr := apperror.As(err)
	if appErr == nil {
		// Unexpected error — do not leak internal details to client
		c.JSON(http.StatusInternalServerError, Error{
			Success: false,
			Message: "an unexpected error occurred",
			Err:     ErrorDetail{Code: "INTERNAL_SERVER_ERROR"},
		})
		return
	}

	c.JSON(appErr.HTTPStatus, Error{
		Success: false,
		Message: appErr.Message,
		Err:     ErrorDetail{Code: appErr.Code},
	})
}

// ErrWithDetails sends an error response with additional details (e.g., validation errors).
func ErrWithDetails(c *gin.Context, err error, details interface{}) {
	appErr := apperror.As(err)
	if appErr == nil {
		c.JSON(http.StatusInternalServerError, Error{
			Success: false,
			Message: "an unexpected error occurred",
			Err:     ErrorDetail{Code: "INTERNAL_SERVER_ERROR", Details: details},
		})
		return
	}

	c.JSON(appErr.HTTPStatus, Error{
		Success: false,
		Message: appErr.Message,
		Err:     ErrorDetail{Code: appErr.Code, Details: details},
	})
}
