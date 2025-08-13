package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"ecommerce-website/internal/logger"
	"ecommerce-website/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func TestErrorHandlingMiddleware(t *testing.T) {
	// Initialize logger for testing
	logger.Initialize(logger.Config{
		Level:       logger.DEBUG,
		ServiceName: "test-service",
		Output:      &bytes.Buffer{}, // Capture logs
	})

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupHandler   func(*gin.Context)
		expectedStatus int
		expectedCode   string
	}{
		{
			name: "handles application error",
			setupHandler: func(c *gin.Context) {
				appErr := NewValidationError("INVALID_INPUT", "Invalid input provided", map[string]interface{}{
					"field": "email",
				})
				c.Error(appErr)
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "INVALID_INPUT",
		},
		{
			name: "handles GORM record not found",
			setupHandler: func(c *gin.Context) {
				c.Error(gorm.ErrRecordNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedCode:   "RECORD_NOT_FOUND",
		},
		{
			name: "handles validation error",
			setupHandler: func(c *gin.Context) {
				c.Error(errors.New("validation failed: email is required"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "VALIDATION_ERROR",
		},
		{
			name: "handles generic error",
			setupHandler: func(c *gin.Context) {
				c.Error(errors.New("something went wrong"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(ErrorHandlingMiddleware())

			r.GET("/test", func(c *gin.Context) {
				tt.setupHandler(c)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response utils.ApiResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.False(t, response.Success)
			assert.Equal(t, tt.expectedCode, response.Error.Code)
		})
	}
}

func TestPanicRecovery(t *testing.T) {
	// Initialize logger for testing
	logger.Initialize(logger.Config{
		Level:       logger.DEBUG,
		ServiceName: "test-service",
		Output:      &bytes.Buffer{},
	})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ErrorHandlingMiddleware())

	r.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
}

func TestAppErrorConstructors(t *testing.T) {
	tests := []struct {
		name           string
		errorFunc      func() *AppError
		expectedType   ErrorType
		expectedStatus int
	}{
		{
			name: "validation error",
			errorFunc: func() *AppError {
				return NewValidationError("INVALID_EMAIL", "Invalid email format", nil)
			},
			expectedType:   ValidationError,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "authentication error",
			errorFunc: func() *AppError {
				return NewAuthenticationError("Invalid credentials")
			},
			expectedType:   AuthenticationError,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "authorization error",
			errorFunc: func() *AppError {
				return NewAuthorizationError("Access denied")
			},
			expectedType:   AuthorizationError,
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "not found error",
			errorFunc: func() *AppError {
				return NewNotFoundError("user", "User not found")
			},
			expectedType:   NotFoundError,
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "conflict error",
			errorFunc: func() *AppError {
				return NewConflictError("EMAIL_EXISTS", "Email already exists", nil)
			},
			expectedType:   ConflictError,
			expectedStatus: http.StatusConflict,
		},
		{
			name: "database error",
			errorFunc: func() *AppError {
				return NewDatabaseError("Connection failed", errors.New("db error"))
			},
			expectedType:   DatabaseError,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name: "external service error",
			errorFunc: func() *AppError {
				return NewExternalServiceError("payment", "Payment service unavailable", errors.New("service error"))
			},
			expectedType:   ExternalServiceError,
			expectedStatus: http.StatusServiceUnavailable,
		},
		{
			name: "internal error",
			errorFunc: func() *AppError {
				return NewInternalError("Unexpected error", errors.New("internal error"))
			},
			expectedType:   InternalError,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.errorFunc()
			assert.Equal(t, tt.expectedType, err.Type)
			assert.Equal(t, tt.expectedStatus, err.StatusCode)
			assert.NotEmpty(t, err.Code)
			assert.NotEmpty(t, err.Message)
		})
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestIDMiddleware())

	r.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		assert.True(t, exists)
		assert.NotEmpty(t, requestID)
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

func TestLoggingMiddleware(t *testing.T) {
	// Capture logs
	logBuffer := &bytes.Buffer{}
	logger.Initialize(logger.Config{
		Level:       logger.DEBUG,
		ServiceName: "test-service",
		Output:      logBuffer,
	})

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RequestIDMiddleware())
	r.Use(LoggingMiddleware())

	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check that logs were written
	logs := logBuffer.String()
	assert.Contains(t, logs, "Request started")
	assert.Contains(t, logs, "Request completed")
}

func TestShouldSkipLogging(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/health", true},
		{"/metrics", true},
		{"/favicon.ico", true},
		{"/robots.txt", true},
		{"/static/css/style.css", true},
		{"/assets/js/app.js", true},
		{"/api/products", false},
		{"/api/users", false},
		{"/", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := shouldSkipLogging(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetLogLevelForStatus(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   logger.LogLevel
	}{
		{200, logger.INFO},
		{201, logger.INFO},
		{400, logger.WARN},
		{401, logger.WARN},
		{404, logger.WARN},
		{500, logger.ERROR},
		{502, logger.ERROR},
		{503, logger.ERROR},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.statusCode)), func(t *testing.T) {
			result := getLogLevelForStatus(tt.statusCode)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestErrorMetrics(t *testing.T) {
	// Reset metrics before test
	ResetErrorMetrics()

	// Record some errors
	RecordError(ValidationError, "INVALID_INPUT")
	RecordError(ValidationError, "INVALID_EMAIL")
	RecordError(DatabaseError, "CONNECTION_FAILED")

	metrics := GetErrorMetrics()
	assert.Equal(t, int64(3), metrics.TotalErrors)
	assert.Equal(t, int64(2), metrics.ErrorsByType[ValidationError])
	assert.Equal(t, int64(1), metrics.ErrorsByType[DatabaseError])
	assert.Equal(t, int64(1), metrics.ErrorsByCode["INVALID_INPUT"])
	assert.Equal(t, int64(1), metrics.ErrorsByCode["INVALID_EMAIL"])
	assert.Equal(t, int64(1), metrics.ErrorsByCode["CONNECTION_FAILED"])
}
