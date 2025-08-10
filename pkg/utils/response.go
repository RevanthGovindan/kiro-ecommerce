package utils

import (
	"time"

	"github.com/gin-gonic/gin"
)

type ApiResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorDetail `json:"error,omitempty"`
	Timestamp string      `json:"timestamp"`
}

type ErrorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, ApiResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}

func ErrorResponse(c *gin.Context, statusCode int, code, message string, details interface{}) {
	c.JSON(statusCode, ApiResponse{
		Success: false,
		Error: &ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	})
}