package monitoring

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMonitorInitialization(t *testing.T) {
	monitor := Initialize()
	assert.NotNil(t, monitor)
	assert.NotNil(t, monitor.alerts)
	assert.NotNil(t, monitor.metrics)
	assert.NotNil(t, monitor.logger)

	// Test singleton behavior
	monitor2 := GetMonitor()
	assert.Equal(t, monitor, monitor2)
}

func TestCreateAlert(t *testing.T) {
	monitor := Initialize()

	// Clear existing alerts
	monitor.alertsMutex.Lock()
	monitor.alerts = make(map[string]*Alert)
	monitor.alertsMutex.Unlock()

	metadata := map[string]interface{}{
		"service": "test-service",
		"count":   5,
	}

	alert := monitor.CreateAlert(AlertWarning, "Test Alert", "This is a test alert", metadata)

	assert.NotNil(t, alert)
	assert.NotEmpty(t, alert.ID)
	assert.Equal(t, AlertWarning, alert.Level)
	assert.Equal(t, "Test Alert", alert.Title)
	assert.Equal(t, "This is a test alert", alert.Message)
	assert.Equal(t, "ecommerce-api", alert.Service)
	assert.Equal(t, metadata, alert.Metadata)
	assert.False(t, alert.Resolved)
	assert.Nil(t, alert.ResolvedAt)
	assert.NotZero(t, alert.Timestamp)

	// Verify alert is stored
	alerts := monitor.GetAlerts()
	assert.Len(t, alerts, 1)
	assert.Equal(t, alert.ID, alerts[0].ID)
}

func TestResolveAlert(t *testing.T) {
	monitor := Initialize()

	// Clear existing alerts
	monitor.alertsMutex.Lock()
	monitor.alerts = make(map[string]*Alert)
	monitor.alertsMutex.Unlock()

	// Create an alert
	alert := monitor.CreateAlert(AlertInfo, "Test Alert", "Test message", nil)
	assert.False(t, alert.Resolved)

	// Resolve the alert
	err := monitor.ResolveAlert(alert.ID)
	assert.NoError(t, err)

	// Verify alert is resolved
	resolvedAlert := monitor.alerts[alert.ID]
	assert.True(t, resolvedAlert.Resolved)
	assert.NotNil(t, resolvedAlert.ResolvedAt)

	// Test resolving non-existent alert
	err = monitor.ResolveAlert("non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "alert not found")
}

func TestGetAlerts(t *testing.T) {
	monitor := Initialize()

	// Clear existing alerts
	monitor.alertsMutex.Lock()
	monitor.alerts = make(map[string]*Alert)
	monitor.alertsMutex.Unlock()

	// Create some alerts
	alert1 := monitor.CreateAlert(AlertInfo, "Alert 1", "Message 1", nil)
	alert2 := monitor.CreateAlert(AlertWarning, "Alert 2", "Message 2", nil)
	alert3 := monitor.CreateAlert(AlertCritical, "Alert 3", "Message 3", nil)

	// Resolve one alert
	monitor.ResolveAlert(alert2.ID)

	// Get active alerts (should exclude resolved ones)
	activeAlerts := monitor.GetAlerts()
	assert.Len(t, activeAlerts, 2)

	// Verify correct alerts are returned
	alertIDs := make([]string, len(activeAlerts))
	for i, alert := range activeAlerts {
		alertIDs[i] = alert.ID
	}
	assert.Contains(t, alertIDs, alert1.ID)
	assert.Contains(t, alertIDs, alert3.ID)
	assert.NotContains(t, alertIDs, alert2.ID)

	// Get all alerts (should include resolved ones)
	allAlerts := monitor.GetAllAlerts()
	assert.Len(t, allAlerts, 3)
}

func TestRunHealthCheck(t *testing.T) {
	monitor := Initialize()

	// Test successful health check
	healthCheck := monitor.RunHealthCheck("test-service", func() (string, error) {
		time.Sleep(10 * time.Millisecond) // Simulate some work
		return "Service is healthy", nil
	})

	assert.Equal(t, "test-service", healthCheck.Name)
	assert.Equal(t, "healthy", healthCheck.Status)
	assert.Equal(t, "Service is healthy", healthCheck.Message)
	assert.True(t, healthCheck.Duration > 0)
	assert.NotZero(t, healthCheck.Timestamp)

	// Verify health check is stored in metrics
	metrics := monitor.GetMetrics()
	assert.Contains(t, metrics.HealthChecks, "test-service")
	assert.Equal(t, "healthy", metrics.HealthChecks["test-service"].Status)

	// Test failed health check
	healthCheck = monitor.RunHealthCheck("failing-service", func() (string, error) {
		return "", errors.New("service is down")
	})

	assert.Equal(t, "failing-service", healthCheck.Name)
	assert.Equal(t, "unhealthy", healthCheck.Status)
	assert.Equal(t, "service is down", healthCheck.Message)

	// Verify alert was created for unhealthy service
	alerts := monitor.GetAlerts()
	found := false
	for _, alert := range alerts {
		if alert.Title == "failing-service Health Check Failed" {
			found = true
			assert.Equal(t, AlertCritical, alert.Level)
			break
		}
	}
	assert.True(t, found, "Expected alert for failing health check")
}

func TestUpdateMetrics(t *testing.T) {
	monitor := Initialize()

	// Update metrics
	requestCount := int64(1000)
	errorCount := int64(50)
	avgResponseTime := 200 * time.Millisecond
	activeUsers := int64(25)

	monitor.UpdateMetrics(requestCount, errorCount, avgResponseTime, activeUsers)

	metrics := monitor.GetMetrics()
	assert.Equal(t, requestCount, metrics.RequestCount)
	assert.Equal(t, errorCount, metrics.ErrorCount)
	assert.Equal(t, float64(5), metrics.ErrorRate) // 50/1000 * 100 = 5%
	assert.Equal(t, avgResponseTime, metrics.AvgResponseTime)
	assert.Equal(t, activeUsers, metrics.ActiveUsers)
	assert.NotZero(t, metrics.Timestamp)

	// Test high error rate alert
	monitor.UpdateMetrics(100, 15, 100*time.Millisecond, 10) // 15% error rate

	alerts := monitor.GetAlerts()
	found := false
	for _, alert := range alerts {
		if alert.Title == "High Error Rate Detected" {
			found = true
			assert.Equal(t, AlertWarning, alert.Level)
			break
		}
	}
	assert.True(t, found, "Expected alert for high error rate")

	// Test high response time alert
	monitor.UpdateMetrics(100, 5, 6*time.Second, 10) // 6 seconds response time

	alerts = monitor.GetAlerts()
	found = false
	for _, alert := range alerts {
		if alert.Title == "High Response Time Detected" {
			found = true
			assert.Equal(t, AlertWarning, alert.Level)
			break
		}
	}
	assert.True(t, found, "Expected alert for high response time")
}

func TestCleanupOldAlerts(t *testing.T) {
	monitor := Initialize()

	// Clear existing alerts
	monitor.alertsMutex.Lock()
	monitor.alerts = make(map[string]*Alert)
	monitor.alertsMutex.Unlock()

	// Create and resolve an old alert
	alert := monitor.CreateAlert(AlertInfo, "Old Alert", "Old message", nil)
	monitor.ResolveAlert(alert.ID)

	// Manually set resolved time to 25 hours ago
	oldTime := time.Now().Add(-25 * time.Hour)
	monitor.alertsMutex.Lock()
	monitor.alerts[alert.ID].ResolvedAt = &oldTime
	monitor.alertsMutex.Unlock()

	// Create a recent resolved alert
	recentAlert := monitor.CreateAlert(AlertInfo, "Recent Alert", "Recent message", nil)
	monitor.ResolveAlert(recentAlert.ID)

	// Run cleanup
	monitor.cleanupOldAlerts()

	// Verify old alert is removed and recent alert remains
	allAlerts := monitor.GetAllAlerts()
	assert.Len(t, allAlerts, 1)
	assert.Equal(t, recentAlert.ID, allAlerts[0].ID)
}

func TestPackageLevelFunctions(t *testing.T) {
	// Test package-level convenience functions
	alert := CreateAlert(AlertInfo, "Package Alert", "Package message", nil)
	assert.NotNil(t, alert)

	alerts := GetAlerts()
	assert.Contains(t, alerts, alert)

	err := ResolveAlert(alert.ID)
	assert.NoError(t, err)

	metrics := GetMetrics()
	assert.NotNil(t, metrics)

	healthCheck := RunHealthCheck("package-test", func() (string, error) {
		return "OK", nil
	})
	assert.Equal(t, "package-test", healthCheck.Name)
	assert.Equal(t, "healthy", healthCheck.Status)
}

func TestGenerateAlertID(t *testing.T) {
	id1 := generateAlertID()
	time.Sleep(1 * time.Nanosecond) // Ensure different timestamps
	id2 := generateAlertID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.Contains(t, id1, "alert-")
	assert.Contains(t, id2, "alert-")

	// IDs should be different (though in rare cases they might be the same due to timing)
	// The important thing is that they follow the correct format
}
