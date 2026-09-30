package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Pagination `json:"meta,omitempty"`
}

type APIErrorResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	ErrorCode string      `json:"error_code,omitempty"`
	Errors    interface{} `json:"errors,omitempty"`
}

type Pagination struct {
	Page         int   `json:"page"`
	Limit        int   `json:"limit"`
	TotalRecords int64 `json:"total_records"`
	TotalPages   int   `json:"total_pages"`
}

func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessWithPagination(c *gin.Context, statusCode int, message string, data interface{}, meta Pagination) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    &meta,
	})
}

func Error(c *gin.Context, statusCode int, message string, errorCode string, errors interface{}) {
	c.JSON(statusCode, APIErrorResponse{
		Success:   false,
		Message:   message,
		ErrorCode: errorCode,
		Errors:    errors,
	})
}

func BadRequest(c *gin.Context, message string, errors interface{}) {
	Error(c, http.StatusBadRequest, message, "BAD_REQUEST", errors)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message, "UNAUTHORIZED", nil)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message, "FORBIDDEN", nil)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message, "NOT_FOUND", nil)
}

func Conflict(c *gin.Context, message string, errorCode string) {
	if errorCode == "" {
		errorCode = "CONFLICT"
	}
	Error(c, http.StatusConflict, message, errorCode, nil)
}

func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message, "INTERNAL_SERVER_ERROR", nil)
}
