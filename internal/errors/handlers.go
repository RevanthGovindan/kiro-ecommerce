package errors

import (
	"net/http"
	"strings"
	"time"

	"ecommerce-website/internal/logger"
	"ecommerce-website/internal/middleware"
	"ecommerce-website/internal/monitoring"
	"ecommerce-website/pkg/utils"

	"github.com/gin-gonic/gin"
)

// ClientErrorRequest represents a client-side error report
type ClientErrorRequest struct {
	Message   string                 `json:"message" binding:"required"`
	Stack     string                 `json:"stack,omitempty"`
	Digest    string                 `json:"digest,omitempty"`
	Timestamp string                 `json:"timestamp" binding:"required"`
	UserAgent string                 `json:"userAgent,omitempty"`
	URL       string                 `json:"url,omitempty"`
	UserID    string                 `json:"userId,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// Handler handles error-related endpoints
type Handler struct {
	logger *logger.Logger
}

// NewHandler creates a new error handler
func NewHandler() *Handler {
	return &Handler{
		logger: logger.GetLogger(),
	}
}

// LogClientError logs client-side errors
func (h *Handler) LogClientError(c *gin.Context) {
	var req ClientErrorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request data", err.Error())
		return
	}

	// Get user ID from context if available
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(string)
	}

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		timestamp = time.Now()
	}

	// Log the client error
	log := logger.WithRequest(c)
	log.Error("Client-side error reported", nil, map[string]interface{}{
		"client_message": req.Message,
		"client_stack":   req.Stack,
		"client_digest":  req.Digest,
		"client_url":     req.URL,
		"user_agent":     req.UserAgent,
		"user_id":        req.UserID,
		"metadata":       req.Metadata,
		"reported_at":    timestamp,
	})

	// Create monitoring alert for critical client errors
	if containsCriticalKeywords(req.Message) {
		monitoring.CreateAlert(
			monitoring.AlertWarning,
			"Critical Client Error",
			req.Message,
			map[string]interface{}{
				"url":        req.URL,
				"user_agent": req.UserAgent,
				"user_id":    req.UserID,
				"stack":      req.Stack,
			},
		)
	}

	// Record error metrics
	middleware.RecordError(middleware.InternalError, "CLIENT_ERROR")

	utils.SuccessResponse(c, http.StatusOK, "Error logged successfully", nil)
}

// GetErrorMetrics returns error metrics
func (h *Handler) GetErrorMetrics(c *gin.Context) {
	metrics := middleware.GetErrorMetrics()
	utils.SuccessResponse(c, http.StatusOK, "Error metrics retrieved", metrics)
}

// GetAlerts returns monitoring alerts
func (h *Handler) GetAlerts(c *gin.Context) {
	alerts := monitoring.GetAlerts()
	utils.SuccessResponse(c, http.StatusOK, "Alerts retrieved", alerts)
}

// GetSystemMetrics returns system metrics
func (h *Handler) GetSystemMetrics(c *gin.Context) {
	metrics := monitoring.GetMetrics()
	utils.SuccessResponse(c, http.StatusOK, "System metrics retrieved", metrics)
}

// ResolveAlert resolves a monitoring alert
func (h *Handler) ResolveAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "VALIDATION_ERROR", "Alert ID is required", nil)
		return
	}

	if err := monitoring.ResolveAlert(alertID); err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "ALERT_NOT_FOUND", err.Error(), nil)
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Alert resolved successfully", nil)
}

// HealthCheck returns system health status
func (h *Handler) HealthCheck(c *gin.Context) {
	monitor := monitoring.GetMonitor()

	// Run immediate health checks
	dbHealth := monitor.RunHealthCheck("database", func() (string, error) {
		// Database health check logic would go here
		return "Database is healthy", nil
	})

	redisHealth := monitor.RunHealthCheck("redis", func() (string, error) {
		// Redis health check logic would go here
		return "Redis is healthy", nil
	})

	healthStatus := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks": map[string]interface{}{
			"database": map[string]interface{}{
				"status":   dbHealth.Status,
				"message":  dbHealth.Message,
				"duration": dbHealth.Duration.String(),
			},
			"redis": map[string]interface{}{
				"status":   redisHealth.Status,
				"message":  redisHealth.Message,
				"duration": redisHealth.Duration.String(),
			},
		},
		"metrics": monitoring.GetMetrics(),
	}

	// Determine overall status
	if dbHealth.Status == "unhealthy" || redisHealth.Status == "unhealthy" {
		healthStatus["status"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, healthStatus)
		return
	}

	c.JSON(http.StatusOK, healthStatus)
}

// containsCriticalKeywords checks if error message contains critical keywords
func containsCriticalKeywords(message string) bool {
	criticalKeywords := []string{
		"payment",
		"checkout",
		"order",
		"security",
		"authentication",
		"authorization",
		"database",
		"network",
		"timeout",
	}

	messageLower := strings.ToLower(message)
	for _, keyword := range criticalKeywords {
		if strings.Contains(messageLower, keyword) {
			return true
		}
	}
	return false
}
