package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLogger(t *testing.T) {
	// Create a buffer to capture log output
	logBuffer := &bytes.Buffer{}

	// Initialize logger with test configuration
	Initialize(Config{
		Level:       DEBUG,
		ServiceName: "test-service",
		Output:      logBuffer,
	})

	logger := GetLogger()

	tests := []struct {
		name     string
		logFunc  func()
		level    LogLevel
		message  string
		hasError bool
	}{
		{
			name: "debug log",
			logFunc: func() {
				logger.Debug("debug message", map[string]interface{}{"key": "value"})
			},
			level:   DEBUG,
			message: "debug message",
		},
		{
			name: "info log",
			logFunc: func() {
				logger.Info("info message")
			},
			level:   INFO,
			message: "info message",
		},
		{
			name: "warn log",
			logFunc: func() {
				logger.Warn("warning message")
			},
			level:   WARN,
			message: "warning message",
		},
		{
			name: "error log",
			logFunc: func() {
				logger.Error("error message", errors.New("test error"))
			},
			level:    ERROR,
			message:  "error message",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear buffer
			logBuffer.Reset()

			// Execute log function
			tt.logFunc()

			// Parse log output
			logOutput := logBuffer.String()
			assert.NotEmpty(t, logOutput)

			var logEntry LogEntry
			err := json.Unmarshal([]byte(strings.TrimSpace(logOutput)), &logEntry)
			assert.NoError(t, err)

			// Verify log entry
			assert.Equal(t, tt.level, logEntry.Level)
			assert.Equal(t, tt.message, logEntry.Message)
			assert.Equal(t, "test-service", logEntry.Service)
			assert.NotEmpty(t, logEntry.Timestamp)

			if tt.hasError {
				assert.NotEmpty(t, logEntry.Error)
				assert.NotEmpty(t, logEntry.StackTrace)
				assert.NotEmpty(t, logEntry.File)
				assert.NotEmpty(t, logEntry.Function)
			}
		})
	}
}

func TestLogLevels(t *testing.T) {
	logBuffer := &bytes.Buffer{}

	// Test with WARN level - should not log DEBUG and INFO
	Initialize(Config{
		Level:       WARN,
		ServiceName: "test-service",
		Output:      logBuffer,
	})

	logger := GetLogger()

	// These should not be logged
	logger.Debug("debug message")
	logger.Info("info message")

	// These should be logged
	logger.Warn("warn message")
	logger.Error("error message", errors.New("test error"))

	logOutput := logBuffer.String()
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")

	// Should have 2 log entries (warn and error)
	assert.Len(t, lines, 2)

	// Verify warn log
	var warnEntry LogEntry
	err := json.Unmarshal([]byte(lines[0]), &warnEntry)
	assert.NoError(t, err)
	assert.Equal(t, WARN, warnEntry.Level)
	assert.Equal(t, "warn message", warnEntry.Message)

	// Verify error log
	var errorEntry LogEntry
	err = json.Unmarshal([]byte(lines[1]), &errorEntry)
	assert.NoError(t, err)
	assert.Equal(t, ERROR, errorEntry.Level)
	assert.Equal(t, "error message", errorEntry.Message)
}

func TestContextLogger(t *testing.T) {
	logBuffer := &bytes.Buffer{}
	Initialize(Config{
		Level:       DEBUG,
		ServiceName: "test-service",
		Output:      logBuffer,
	})

	logger := GetLogger()

	// Create context with values
	ctx := context.WithValue(context.Background(), "request_id", "test-request-123")
	ctx = context.WithValue(ctx, "user_id", "user-456")

	contextLogger := logger.WithContext(ctx)
	contextLogger.Info("context message")

	logOutput := logBuffer.String()
	var logEntry LogEntry
	err := json.Unmarshal([]byte(strings.TrimSpace(logOutput)), &logEntry)
	assert.NoError(t, err)

	assert.Equal(t, "context message", logEntry.Message)
	assert.Equal(t, "test-request-123", logEntry.Fields["request_id"])
	assert.Equal(t, "user-456", logEntry.Fields["user_id"])
}

func TestRequestLogger(t *testing.T) {
	logBuffer := &bytes.Buffer{}
	Initialize(Config{
		Level:       DEBUG,
		ServiceName: "test-service",
		Output:      logBuffer,
	})

	logger := GetLogger()

	// Create Gin context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/test?param=value", nil)
	c.Request.Header.Set("User-Agent", "test-agent")
	c.Set("request_id", "req-123")
	c.Set("userID", "user-456")

	requestLogger := logger.WithRequest(c)
	requestLogger.Info("request message")

	logOutput := logBuffer.String()
	var logEntry LogEntry
	err := json.Unmarshal([]byte(strings.TrimSpace(logOutput)), &logEntry)
	assert.NoError(t, err)

	assert.Equal(t, "request message", logEntry.Message)
	assert.Equal(t, "req-123", logEntry.Fields["request_id"])
	assert.Equal(t, "user-456", logEntry.Fields["user_id"])
	assert.Equal(t, "GET", logEntry.Fields["method"])
	assert.Equal(t, "/test", logEntry.Fields["path"])
	assert.Equal(t, "test-agent", logEntry.Fields["user_agent"])
}

func TestPackageLevelFunctions(t *testing.T) {
	logBuffer := &bytes.Buffer{}
	Initialize(Config{
		Level:       DEBUG,
		ServiceName: "test-service",
		Output:      logBuffer,
	})

	// Test package-level functions
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message", errors.New("test error"))

	logOutput := logBuffer.String()
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")
	assert.Len(t, lines, 4)

	// Verify each log level
	levels := []LogLevel{DEBUG, INFO, WARN, ERROR}
	messages := []string{"debug message", "info message", "warn message", "error message"}

	for i, line := range lines {
		var logEntry LogEntry
		err := json.Unmarshal([]byte(line), &logEntry)
		assert.NoError(t, err)
		assert.Equal(t, levels[i], logEntry.Level)
		assert.Equal(t, messages[i], logEntry.Message)
	}
}

func TestLoggerInitialization(t *testing.T) {
	// Test default initialization
	Initialize(Config{})
	logger := GetLogger()
	assert.NotNil(t, logger)
	assert.Equal(t, INFO, logger.level)
	assert.Equal(t, "ecommerce-api", logger.serviceName)

	// Test custom initialization
	logBuffer := &bytes.Buffer{}
	Initialize(Config{
		Level:       ERROR,
		ServiceName: "custom-service",
		Output:      logBuffer,
	})
	logger = GetLogger()
	assert.Equal(t, ERROR, logger.level)
	assert.Equal(t, "custom-service", logger.serviceName)
	assert.Equal(t, logBuffer, logger.output)
}

func TestShouldLog(t *testing.T) {
	logger := &Logger{level: WARN}

	tests := []struct {
		level    LogLevel
		expected bool
	}{
		{DEBUG, false},
		{INFO, false},
		{WARN, true},
		{ERROR, true},
		{FATAL, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			result := logger.shouldLog(tt.level)
			assert.Equal(t, tt.expected, result)
		})
	}
}
