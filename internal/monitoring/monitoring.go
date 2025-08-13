package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"ecommerce-website/internal/database"
	"ecommerce-website/internal/logger"
	"ecommerce-website/internal/middleware"
)

// AlertLevel represents the severity of an alert
type AlertLevel string

const (
	AlertInfo     AlertLevel = "INFO"
	AlertWarning  AlertLevel = "WARNING"
	AlertCritical AlertLevel = "CRITICAL"
)

// Alert represents a monitoring alert
type Alert struct {
	ID         string                 `json:"id"`
	Level      AlertLevel             `json:"level"`
	Title      string                 `json:"title"`
	Message    string                 `json:"message"`
	Service    string                 `json:"service"`
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Resolved   bool                   `json:"resolved"`
	ResolvedAt *time.Time             `json:"resolved_at,omitempty"`
}

// HealthCheck represents a system health check
type HealthCheck struct {
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SystemMetrics represents system-wide metrics
type SystemMetrics struct {
	Timestamp       time.Time                `json:"timestamp"`
	RequestCount    int64                    `json:"request_count"`
	ErrorCount      int64                    `json:"error_count"`
	ErrorRate       float64                  `json:"error_rate"`
	AvgResponseTime time.Duration            `json:"avg_response_time"`
	ActiveUsers     int64                    `json:"active_users"`
	DatabaseHealth  string                   `json:"database_health"`
	RedisHealth     string                   `json:"redis_health"`
	ErrorBreakdown  *middleware.ErrorMetrics `json:"error_breakdown"`
	HealthChecks    map[string]*HealthCheck  `json:"health_checks"`
}

// Monitor provides monitoring and alerting functionality
type Monitor struct {
	alerts       map[string]*Alert
	alertsMutex  sync.RWMutex
	metrics      *SystemMetrics
	metricsMutex sync.RWMutex
	logger       *logger.Logger
}

var defaultMonitor *Monitor
var once sync.Once

// Initialize sets up the default monitor
func Initialize() *Monitor {
	once.Do(func() {
		defaultMonitor = &Monitor{
			alerts: make(map[string]*Alert),
			metrics: &SystemMetrics{
				HealthChecks: make(map[string]*HealthCheck),
			},
			logger: logger.GetLogger(),
		}

		// Start background monitoring
		go defaultMonitor.startBackgroundMonitoring()
	})
	return defaultMonitor
}

// GetMonitor returns the default monitor instance
func GetMonitor() *Monitor {
	if defaultMonitor == nil {
		return Initialize()
	}
	return defaultMonitor
}

// CreateAlert creates a new alert
func (m *Monitor) CreateAlert(level AlertLevel, title, message string, metadata map[string]interface{}) *Alert {
	alert := &Alert{
		ID:        generateAlertID(),
		Level:     level,
		Title:     title,
		Message:   message,
		Service:   "ecommerce-api",
		Timestamp: time.Now(),
		Metadata:  metadata,
		Resolved:  false,
	}

	m.alertsMutex.Lock()
	m.alerts[alert.ID] = alert
	m.alertsMutex.Unlock()

	// Log the alert
	switch level {
	case AlertInfo:
		m.logger.Info(fmt.Sprintf("Alert: %s", title), map[string]interface{}{
			"alert_id": alert.ID,
			"message":  message,
			"metadata": metadata,
		})
	case AlertWarning:
		m.logger.Warn(fmt.Sprintf("Alert: %s", title), map[string]interface{}{
			"alert_id": alert.ID,
			"message":  message,
			"metadata": metadata,
		})
	case AlertCritical:
		m.logger.Error(fmt.Sprintf("Critical Alert: %s", title), nil, map[string]interface{}{
			"alert_id": alert.ID,
			"message":  message,
			"metadata": metadata,
		})
	}

	// Store alert in Redis for persistence
	m.storeAlert(alert)

	return alert
}

// ResolveAlert marks an alert as resolved
func (m *Monitor) ResolveAlert(alertID string) error {
	m.alertsMutex.Lock()
	defer m.alertsMutex.Unlock()

	alert, exists := m.alerts[alertID]
	if !exists {
		return fmt.Errorf("alert not found: %s", alertID)
	}

	now := time.Now()
	alert.Resolved = true
	alert.ResolvedAt = &now

	m.logger.Info("Alert resolved", map[string]interface{}{
		"alert_id": alertID,
		"title":    alert.Title,
	})

	// Update alert in Redis
	m.storeAlert(alert)

	return nil
}

// GetAlerts returns all active alerts
func (m *Monitor) GetAlerts() []*Alert {
	m.alertsMutex.RLock()
	defer m.alertsMutex.RUnlock()

	alerts := make([]*Alert, 0, len(m.alerts))
	for _, alert := range m.alerts {
		if !alert.Resolved {
			alerts = append(alerts, alert)
		}
	}

	return alerts
}

// GetAllAlerts returns all alerts (including resolved ones)
func (m *Monitor) GetAllAlerts() []*Alert {
	m.alertsMutex.RLock()
	defer m.alertsMutex.RUnlock()

	alerts := make([]*Alert, 0, len(m.alerts))
	for _, alert := range m.alerts {
		alerts = append(alerts, alert)
	}

	return alerts
}

// RunHealthCheck performs a health check
func (m *Monitor) RunHealthCheck(name string, checkFunc func() (string, error)) *HealthCheck {
	start := time.Now()
	status := "healthy"
	var message string

	if result, err := checkFunc(); err != nil {
		status = "unhealthy"
		message = err.Error()
	} else {
		message = result
	}

	healthCheck := &HealthCheck{
		Name:      name,
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		Duration:  time.Since(start),
	}

	m.metricsMutex.Lock()
	m.metrics.HealthChecks[name] = healthCheck
	m.metricsMutex.Unlock()

	// Create alert for unhealthy services
	if status == "unhealthy" {
		m.CreateAlert(AlertCritical, fmt.Sprintf("%s Health Check Failed", name), message, map[string]interface{}{
			"service":  name,
			"duration": healthCheck.Duration.String(),
		})
	}

	return healthCheck
}

// UpdateMetrics updates system metrics
func (m *Monitor) UpdateMetrics(requestCount, errorCount int64, avgResponseTime time.Duration, activeUsers int64) {
	m.metricsMutex.Lock()
	defer m.metricsMutex.Unlock()

	var errorRate float64
	if requestCount > 0 {
		errorRate = float64(errorCount) / float64(requestCount) * 100
	}

	m.metrics.Timestamp = time.Now()
	m.metrics.RequestCount = requestCount
	m.metrics.ErrorCount = errorCount
	m.metrics.ErrorRate = errorRate
	m.metrics.AvgResponseTime = avgResponseTime
	m.metrics.ActiveUsers = activeUsers
	m.metrics.ErrorBreakdown = middleware.GetErrorMetrics()

	// Check for high error rate
	if errorRate > 10.0 { // Alert if error rate > 10%
		m.CreateAlert(AlertWarning, "High Error Rate Detected",
			fmt.Sprintf("Error rate is %.2f%% (%d errors out of %d requests)", errorRate, errorCount, requestCount),
			map[string]interface{}{
				"error_rate":    errorRate,
				"error_count":   errorCount,
				"request_count": requestCount,
			})
	}

	// Check for high response time
	if avgResponseTime > 5*time.Second {
		m.CreateAlert(AlertWarning, "High Response Time Detected",
			fmt.Sprintf("Average response time is %v", avgResponseTime),
			map[string]interface{}{
				"avg_response_time": avgResponseTime.String(),
			})
	}
}

// GetMetrics returns current system metrics
func (m *Monitor) GetMetrics() *SystemMetrics {
	m.metricsMutex.RLock()
	defer m.metricsMutex.RUnlock()

	// Create a copy to avoid race conditions
	metrics := *m.metrics
	return &metrics
}

// startBackgroundMonitoring runs periodic health checks and monitoring tasks
func (m *Monitor) startBackgroundMonitoring() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		// Run database health check
		m.RunHealthCheck("database", func() (string, error) {
			db := database.GetDB()
			if db == nil {
				return "", fmt.Errorf("database connection is nil")
			}

			sqlDB, err := db.DB()
			if err != nil {
				return "", fmt.Errorf("failed to get underlying sql.DB: %v", err)
			}

			if err := sqlDB.Ping(); err != nil {
				return "", fmt.Errorf("database ping failed: %v", err)
			}

			return "Database connection is healthy", nil
		})

		// Run Redis health check
		m.RunHealthCheck("redis", func() (string, error) {
			rdb := database.GetRedisClient()
			if rdb == nil {
				return "", fmt.Errorf("redis connection is nil")
			}

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := rdb.Ping(ctx).Err(); err != nil {
				return "", fmt.Errorf("redis ping failed: %v", err)
			}

			return "Redis connection is healthy", nil
		})

		// Update database and Redis health in metrics
		m.metricsMutex.Lock()
		if dbHealth, exists := m.metrics.HealthChecks["database"]; exists {
			m.metrics.DatabaseHealth = dbHealth.Status
		}
		if redisHealth, exists := m.metrics.HealthChecks["redis"]; exists {
			m.metrics.RedisHealth = redisHealth.Status
		}
		m.metricsMutex.Unlock()

		// Clean up old resolved alerts (older than 24 hours)
		m.cleanupOldAlerts()
	}
}

// storeAlert stores an alert in Redis for persistence
func (m *Monitor) storeAlert(alert *Alert) {
	rdb := database.GetRedisClient()
	if rdb == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	data, err := json.Marshal(alert)
	if err != nil {
		m.logger.Error("Failed to marshal alert", err, map[string]interface{}{
			"alert_id": alert.ID,
		})
		return
	}

	key := fmt.Sprintf("alert:%s", alert.ID)
	if err := rdb.Set(ctx, key, data, 24*time.Hour).Err(); err != nil {
		m.logger.Error("Failed to store alert in Redis", err, map[string]interface{}{
			"alert_id": alert.ID,
		})
	}
}

// cleanupOldAlerts removes resolved alerts older than 24 hours
func (m *Monitor) cleanupOldAlerts() {
	m.alertsMutex.Lock()
	defer m.alertsMutex.Unlock()

	cutoff := time.Now().Add(-24 * time.Hour)
	for id, alert := range m.alerts {
		if alert.Resolved && alert.ResolvedAt != nil && alert.ResolvedAt.Before(cutoff) {
			delete(m.alerts, id)
		}
	}
}

// generateAlertID generates a unique alert identifier
func generateAlertID() string {
	return fmt.Sprintf("alert-%d", time.Now().UnixNano())
}

// Package-level convenience functions
func CreateAlert(level AlertLevel, title, message string, metadata map[string]interface{}) *Alert {
	return GetMonitor().CreateAlert(level, title, message, metadata)
}

func ResolveAlert(alertID string) error {
	return GetMonitor().ResolveAlert(alertID)
}

func GetAlerts() []*Alert {
	return GetMonitor().GetAlerts()
}

func GetMetrics() *SystemMetrics {
	return GetMonitor().GetMetrics()
}

func RunHealthCheck(name string, checkFunc func() (string, error)) *HealthCheck {
	return GetMonitor().RunHealthCheck(name, checkFunc)
}
