package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"ecommerce-website/internal/logger"
	"ecommerce-website/pkg/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ErrorType represents different types of application errors
type ErrorType string

const (
	ValidationError      ErrorType = "VALIDATION_ERROR"
	AuthenticationError  ErrorType = "AUTHENTICATION_ERROR"
	AuthorizationError   ErrorType = "AUTHORIZATION_ERROR"
	NotFoundError        ErrorType = "NOT_FOUND_ERROR"
	ConflictError        ErrorType = "CONFLICT_ERROR"
	DatabaseError        ErrorType = "DATABASE_ERROR"
	ExternalServiceError ErrorType = "EXTERNAL_SERVICE_ERROR"
	InternalError        ErrorType = "INTERNAL_ERROR"
	RateLimitError       ErrorType = "RATE_LIMIT_ERROR"
)

// AppError represents an application-specific error
type AppError struct {
	Type       ErrorType              `json:"type"`
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	StatusCode int                    `json:"status_code"`
	Cause      error                  `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewAppError creates a new application error
func NewAppError(errorType ErrorType, code, message string, statusCode int, cause error) *AppError {
	return &AppError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Cause:      cause,
	}
}

// Common error constructors
func NewValidationError(code, message string, details map[string]interface{}) *AppError {
	return &AppError{
		Type:       ValidationError,
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: http.StatusBadRequest,
	}
}

func NewAuthenticationError(message string) *AppError {
	return &AppError{
		Type:       AuthenticationError,
		Code:       "AUTHENTICATION_FAILED",
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

func NewAuthorizationError(message string) *AppError {
	return &AppError{
		Type:       AuthorizationError,
		Code:       "AUTHORIZATION_FAILED",
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

func NewNotFoundError(resource, message string) *AppError {
	return &AppError{
		Type:       NotFoundError,
		Code:       fmt.Sprintf("%s_NOT_FOUND", strings.ToUpper(resource)),
		Message:    message,
		StatusCode: http.StatusNotFound,
	}
}

func NewConflictError(code, message string, details map[string]interface{}) *AppError {
	return &AppError{
		Type:       ConflictError,
		Code:       code,
		Message:    message,
		Details:    details,
		StatusCode: http.StatusConflict,
	}
}

func NewDatabaseError(message string, cause error) *AppError {
	return &AppError{
		Type:       DatabaseError,
		Code:       "DATABASE_ERROR",
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Cause:      cause,
	}
}

func NewExternalServiceError(service, message string, cause error) *AppError {
	return &AppError{
		Type:       ExternalServiceError,
		Code:       fmt.Sprintf("%s_SERVICE_ERROR", strings.ToUpper(service)),
		Message:    message,
		StatusCode: http.StatusServiceUnavailable,
		Cause:      cause,
	}
}

func NewInternalError(message string, cause error) *AppError {
	return &AppError{
		Type:       InternalError,
		Code:       "INTERNAL_ERROR",
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Cause:      cause,
	}
}

// ErrorHandlingMiddleware provides global error handling
func ErrorHandlingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log panic with stack trace
				logger.WithRequest(c).Error("Panic recovered", fmt.Errorf("panic: %v", err), map[string]interface{}{
					"stack_trace": string(debug.Stack()),
				})

				// Return internal server error
				utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR",
					"An unexpected error occurred", nil)
				c.Abort()
			}
		}()

		c.Next()

		// Handle errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleError(c, err)
		}
	}
}

// handleError processes different types of errors and returns appropriate responses
func handleError(c *gin.Context, err error) {
	log := logger.WithRequest(c)

	// Handle application-specific errors
	var appErr *AppError
	if errors.As(err, &appErr) {
		// Log based on error type
		switch appErr.Type {
		case ValidationError, AuthenticationError, AuthorizationError, NotFoundError:
			log.Warn("Client error", map[string]interface{}{
				"error_type": appErr.Type,
				"error_code": appErr.Code,
			})
		case DatabaseError, ExternalServiceError, InternalError:
			log.Error("Server error", appErr.Cause, map[string]interface{}{
				"error_type": appErr.Type,
				"error_code": appErr.Code,
			})
		default:
			log.Error("Unknown error type", err, map[string]interface{}{
				"error_type": appErr.Type,
				"error_code": appErr.Code,
			})
		}

		utils.ErrorResponse(c, appErr.StatusCode, appErr.Code, appErr.Message, appErr.Details)
		return
	}

	// Handle GORM errors
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Warn("Record not found", map[string]interface{}{
			"error": err.Error(),
		})
		utils.ErrorResponse(c, http.StatusNotFound, "RECORD_NOT_FOUND", "The requested resource was not found", nil)
		return
	}

	// Handle other GORM errors
	if isGormError(err) {
		log.Error("Database error", err, nil)
		utils.ErrorResponse(c, http.StatusInternalServerError, "DATABASE_ERROR", "A database error occurred", nil)
		return
	}

	// Handle validation errors from Gin binding
	if strings.Contains(err.Error(), "validation") || strings.Contains(err.Error(), "binding") {
		log.Warn("Validation error", map[string]interface{}{
			"error": err.Error(),
		})
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// Handle generic errors
	log.Error("Unhandled error", err, nil)
	utils.ErrorResponse(c, http.StatusInternalServerError, "INTERNAL_ERROR", "An unexpected error occurred", nil)
}

// isGormError checks if an error is a GORM-related error
func isGormError(err error) bool {
	gormErrors := []string{
		"duplicate key",
		"foreign key constraint",
		"check constraint",
		"not null constraint",
		"unique constraint",
		"connection refused",
		"connection timeout",
	}

	errStr := strings.ToLower(err.Error())
	for _, gormErr := range gormErrors {
		if strings.Contains(errStr, gormErr) {
			return true
		}
	}
	return false
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// generateRequestID creates a unique request identifier
func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Nanosecond())
}

// LoggingMiddleware logs HTTP requests and responses
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Skip logging for health checks and static assets
		if shouldSkipLogging(path) {
			c.Next()
			return
		}

		log := logger.WithRequest(c)

		// Log request
		log.Info("Request started", map[string]interface{}{
			"query": raw,
		})

		c.Next()

		// Log response
		duration := time.Since(start)
		statusCode := c.Writer.Status()

		logLevel := getLogLevelForStatus(statusCode)
		fields := map[string]interface{}{
			"status_code": statusCode,
			"duration":    duration.String(),
			"size":        c.Writer.Size(),
		}

		switch logLevel {
		case logger.INFO:
			log.Info("Request completed", fields)
		case logger.WARN:
			log.Warn("Request completed with warning", fields)
		case logger.ERROR:
			log.Error("Request completed with error", nil, fields)
		}
	}
}

// shouldSkipLogging determines if a request should be skipped from logging
func shouldSkipLogging(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/favicon.ico",
		"/robots.txt",
	}

	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}

	// Skip static assets
	if strings.HasPrefix(path, "/static/") || strings.HasPrefix(path, "/assets/") {
		return true
	}

	return false
}

// getLogLevelForStatus returns appropriate log level based on HTTP status code
func getLogLevelForStatus(statusCode int) logger.LogLevel {
	switch {
	case statusCode >= 500:
		return logger.ERROR
	case statusCode >= 400:
		return logger.WARN
	default:
		return logger.INFO
	}
}

// ErrorMetrics tracks error metrics for monitoring
type ErrorMetrics struct {
	TotalErrors  int64               `json:"total_errors"`
	ErrorsByType map[ErrorType]int64 `json:"errors_by_type"`
	ErrorsByCode map[string]int64    `json:"errors_by_code"`
	LastUpdated  time.Time           `json:"last_updated"`
}

var errorMetrics = &ErrorMetrics{
	ErrorsByType: make(map[ErrorType]int64),
	ErrorsByCode: make(map[string]int64),
}

// RecordError records error metrics
func RecordError(errorType ErrorType, code string) {
	errorMetrics.TotalErrors++
	errorMetrics.ErrorsByType[errorType]++
	errorMetrics.ErrorsByCode[code]++
	errorMetrics.LastUpdated = time.Now()
}

// GetErrorMetrics returns current error metrics
func GetErrorMetrics() *ErrorMetrics {
	return errorMetrics
}

// ResetErrorMetrics resets error metrics (useful for testing)
func ResetErrorMetrics() {
	errorMetrics = &ErrorMetrics{
		ErrorsByType: make(map[ErrorType]int64),
		ErrorsByCode: make(map[string]int64),
	}
}
