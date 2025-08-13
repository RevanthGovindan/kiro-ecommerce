package errors

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"ecommerce-website/internal/logger"
	"ecommerce-website/internal/middleware"
	"ecommerce-website/internal/monitoring"
	"ecommerce-website/pkg/utils"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestHandler() *Handler {
	// Initialize logger for testing
	logger.Initialize(logger.Config{
		Level:       logger.DEBUG,
		ServiceName: "test-service",
		Output:      &bytes.Buffer{},
	})

	// Initialize monitoring
	monitoring.Initialize()

	// Reset error metrics
	middleware.ResetErrorMetrics()

	return NewHandler()
}

func TestLogClientError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedCode   string
	}{
		{
			name: "valid client error",
			requestBody: ClientErrorRequest{
				Message:   "TypeError: Cannot read property 'foo' of undefined",
				Stack:     "Error: at line 123",
				Timestamp: time.Now().Format(time.RFC3339),
				UserAgent: "Mozilla/5.0",
				URL:       "https://example.com/page",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing required fields",
			requestBody: map[string]interface{}{
				"message": "Error message",
				// Missing timestamp
			},
			expectedStatus: http.StatusBadRequest,
			expectedCode:   "VALIDATION_ERROR",
		},
		{
			name: "invalid timestamp format",
			requestBody: ClientErrorRequest{
				Message:   "Error message",
				Timestamp: "invalid-timestamp",
			},
			expectedStatus: http.StatusOK, // Should still succeed but use current time
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.POST("/api/errors/client", handler.LogClientError)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/api/errors/client", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response utils.ApiResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedStatus == http.StatusOK {
				assert.True(t, response.Success)
			} else {
				assert.False(t, response.Success)
				if tt.expectedCode != "" {
					assert.Equal(t, tt.expectedCode, response.Error.Code)
				}
			}
		})
	}
}

func TestLogClientErrorWithCriticalKeywords(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	r := gin.New()
	r.POST("/api/errors/client", handler.LogClientError)

	// Test with critical keyword
	requestBody := ClientErrorRequest{
		Message:   "Payment processing failed",
		Timestamp: time.Now().Format(time.RFC3339),
	}

	body, _ := json.Marshal(requestBody)
	req := httptest.NewRequest("POST", "/api/errors/client", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify alert was created
	alerts := monitoring.GetAlerts()
	found := false
	for _, alert := range alerts {
		if alert.Title == "Critical Client Error" {
			found = true
			assert.Equal(t, monitoring.AlertWarning, alert.Level)
			break
		}
	}
	assert.True(t, found, "Expected alert for critical client error")
}

func TestGetErrorMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	// Record some test errors
	middleware.RecordError(middleware.ValidationError, "INVALID_INPUT")
	middleware.RecordError(middleware.DatabaseError, "CONNECTION_FAILED")

	r := gin.New()
	r.GET("/api/admin/monitoring/metrics", handler.GetErrorMetrics)

	req := httptest.NewRequest("GET", "/api/admin/monitoring/metrics", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	// Verify metrics data exists
	assert.NotNil(t, response.Data)
}

func TestGetAlerts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	// Create test alerts
	monitoring.CreateAlert(monitoring.AlertWarning, "Test Alert 1", "Message 1", nil)
	monitoring.CreateAlert(monitoring.AlertCritical, "Test Alert 2", "Message 2", nil)

	r := gin.New()
	r.GET("/api/admin/monitoring/alerts", handler.GetAlerts)

	req := httptest.NewRequest("GET", "/api/admin/monitoring/alerts", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	// Verify alerts data exists
	assert.NotNil(t, response.Data)
}

func TestGetSystemMetrics(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	// Update some test metrics
	monitor := monitoring.GetMonitor()
	monitor.UpdateMetrics(1000, 50, 200*time.Millisecond, 25)

	r := gin.New()
	r.GET("/api/admin/monitoring/system", handler.GetSystemMetrics)

	req := httptest.NewRequest("GET", "/api/admin/monitoring/system", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	// Verify metrics data exists
	assert.NotNil(t, response.Data)
}

func TestResolveAlert(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	// Create test alert
	alert := monitoring.CreateAlert(monitoring.AlertInfo, "Test Alert", "Test message", nil)

	r := gin.New()
	r.POST("/api/admin/monitoring/alerts/:id/resolve", handler.ResolveAlert)

	// Test successful resolution
	req := httptest.NewRequest("POST", "/api/admin/monitoring/alerts/"+alert.ID+"/resolve", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response utils.ApiResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	// Test resolving non-existent alert
	req = httptest.NewRequest("POST", "/api/admin/monitoring/alerts/non-existent/resolve", nil)
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "ALERT_NOT_FOUND", response.Error.Code)

	// Test missing alert ID
	req = httptest.NewRequest("POST", "/api/admin/monitoring/alerts//resolve", nil)
	w = httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := setupTestHandler()

	r := gin.New()
	r.GET("/api/health", handler.HealthCheck)

	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var healthStatus map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &healthStatus)
	assert.NoError(t, err)

	assert.Equal(t, "healthy", healthStatus["status"])
	assert.Contains(t, healthStatus, "timestamp")
	assert.Contains(t, healthStatus, "checks")
	assert.Contains(t, healthStatus, "metrics")

	checks, ok := healthStatus["checks"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, checks, "database")
	assert.Contains(t, checks, "redis")
}

func TestContainsCriticalKeywords(t *testing.T) {
	tests := []struct {
		message  string
		expected bool
	}{
		{"Payment processing failed", true},
		{"Checkout error occurred", true},
		{"Order submission failed", true},
		{"Security violation detected", true},
		{"Authentication failed", true},
		{"Database connection lost", true},
		{"Network timeout", true},
		{"Simple UI error", false},
		{"Button click failed", false},
		{"Form validation error", false},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			result := containsCriticalKeywords(tt.message)
			assert.Equal(t, tt.expected, result)
		})
	}
}
