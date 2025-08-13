package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
	FATAL LogLevel = "FATAL"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  string                 `json:"timestamp"`
	Level      LogLevel               `json:"level"`
	Message    string                 `json:"message"`
	Service    string                 `json:"service"`
	RequestID  string                 `json:"request_id,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Duration   string                 `json:"duration,omitempty"`
	Error      string                 `json:"error,omitempty"`
	StackTrace string                 `json:"stack_trace,omitempty"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	File       string                 `json:"file,omitempty"`
	Function   string                 `json:"function,omitempty"`
}

// Logger provides structured logging functionality
type Logger struct {
	output      io.Writer
	level       LogLevel
	serviceName string
}

// Config holds logger configuration
type Config struct {
	Level       LogLevel
	ServiceName string
	Output      io.Writer
}

var defaultLogger *Logger

// Initialize sets up the default logger
func Initialize(config Config) {
	if config.Output == nil {
		config.Output = os.Stdout
	}
	if config.ServiceName == "" {
		config.ServiceName = "ecommerce-api"
	}
	if config.Level == "" {
		config.Level = INFO
	}

	defaultLogger = &Logger{
		output:      config.Output,
		level:       config.Level,
		serviceName: config.ServiceName,
	}
}

// GetLogger returns the default logger instance
func GetLogger() *Logger {
	if defaultLogger == nil {
		Initialize(Config{})
	}
	return defaultLogger
}

// shouldLog determines if a message should be logged based on level
func (l *Logger) shouldLog(level LogLevel) bool {
	levels := map[LogLevel]int{
		DEBUG: 0,
		INFO:  1,
		WARN:  2,
		ERROR: 3,
		FATAL: 4,
	}
	return levels[level] >= levels[l.level]
}

// log writes a log entry
func (l *Logger) log(level LogLevel, message string, fields map[string]interface{}, err error) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     level,
		Message:   message,
		Service:   l.serviceName,
		Fields:    fields,
	}

	// Add error information if provided
	if err != nil {
		entry.Error = err.Error()
		if level == ERROR || level == FATAL {
			entry.StackTrace = getStackTrace()
		}
	}

	// Add caller information for errors
	if level == ERROR || level == FATAL {
		if pc, file, line, ok := runtime.Caller(2); ok {
			entry.File = fmt.Sprintf("%s:%d", file, line)
			if fn := runtime.FuncForPC(pc); fn != nil {
				entry.Function = fn.Name()
			}
		}
	}

	// Marshal and write the log entry
	data, _ := json.Marshal(entry)
	fmt.Fprintln(l.output, string(data))
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(DEBUG, message, f, nil)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(INFO, message, f, nil)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(WARN, message, f, nil)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(ERROR, message, f, err)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, err error, fields ...map[string]interface{}) {
	var f map[string]interface{}
	if len(fields) > 0 {
		f = fields[0]
	}
	l.log(FATAL, message, f, err)
	os.Exit(1)
}

// WithContext creates a logger with context information
func (l *Logger) WithContext(ctx context.Context) *ContextLogger {
	return &ContextLogger{
		logger: l,
		ctx:    ctx,
	}
}

// WithRequest creates a logger with request information
func (l *Logger) WithRequest(c *gin.Context) *RequestLogger {
	return &RequestLogger{
		logger: l,
		ctx:    c,
	}
}

// ContextLogger provides logging with context
type ContextLogger struct {
	logger *Logger
	ctx    context.Context
}

// Debug logs a debug message with context
func (cl *ContextLogger) Debug(message string, fields ...map[string]interface{}) {
	f := cl.enrichFields(fields...)
	cl.logger.log(DEBUG, message, f, nil)
}

// Info logs an info message with context
func (cl *ContextLogger) Info(message string, fields ...map[string]interface{}) {
	f := cl.enrichFields(fields...)
	cl.logger.log(INFO, message, f, nil)
}

// Warn logs a warning message with context
func (cl *ContextLogger) Warn(message string, fields ...map[string]interface{}) {
	f := cl.enrichFields(fields...)
	cl.logger.log(WARN, message, f, nil)
}

// Error logs an error message with context
func (cl *ContextLogger) Error(message string, err error, fields ...map[string]interface{}) {
	f := cl.enrichFields(fields...)
	cl.logger.log(ERROR, message, f, err)
}

func (cl *ContextLogger) enrichFields(fields ...map[string]interface{}) map[string]interface{} {
	f := make(map[string]interface{})
	if len(fields) > 0 && fields[0] != nil {
		for k, v := range fields[0] {
			f[k] = v
		}
	}

	// Add context values if available
	if requestID := cl.ctx.Value("request_id"); requestID != nil {
		f["request_id"] = requestID
	}
	if userID := cl.ctx.Value("user_id"); userID != nil {
		f["user_id"] = userID
	}

	return f
}

// RequestLogger provides logging with request information
type RequestLogger struct {
	logger *Logger
	ctx    *gin.Context
}

// Debug logs a debug message with request context
func (rl *RequestLogger) Debug(message string, fields ...map[string]interface{}) {
	f := rl.enrichFields(fields...)
	rl.logger.log(DEBUG, message, f, nil)
}

// Info logs an info message with request context
func (rl *RequestLogger) Info(message string, fields ...map[string]interface{}) {
	f := rl.enrichFields(fields...)
	rl.logger.log(INFO, message, f, nil)
}

// Warn logs a warning message with request context
func (rl *RequestLogger) Warn(message string, fields ...map[string]interface{}) {
	f := rl.enrichFields(fields...)
	rl.logger.log(WARN, message, f, nil)
}

// Error logs an error message with request context
func (rl *RequestLogger) Error(message string, err error, fields ...map[string]interface{}) {
	f := rl.enrichFields(fields...)
	rl.logger.log(ERROR, message, f, err)
}

func (rl *RequestLogger) enrichFields(fields ...map[string]interface{}) map[string]interface{} {
	f := make(map[string]interface{})
	if len(fields) > 0 && fields[0] != nil {
		for k, v := range fields[0] {
			f[k] = v
		}
	}

	// Add request context
	if requestID, exists := rl.ctx.Get("request_id"); exists {
		f["request_id"] = requestID
	}
	if userID, exists := rl.ctx.Get("userID"); exists {
		f["user_id"] = userID
	}
	f["method"] = rl.ctx.Request.Method
	f["path"] = rl.ctx.Request.URL.Path
	f["ip"] = rl.ctx.ClientIP()
	f["user_agent"] = rl.ctx.GetHeader("User-Agent")

	return f
}

// getStackTrace returns the current stack trace
func getStackTrace() string {
	buf := make([]byte, 4096)
	n := runtime.Stack(buf, false)
	return string(buf[:n])
}

// Package-level convenience functions
func Debug(message string, fields ...map[string]interface{}) {
	GetLogger().Debug(message, fields...)
}

func Info(message string, fields ...map[string]interface{}) {
	GetLogger().Info(message, fields...)
}

func Warn(message string, fields ...map[string]interface{}) {
	GetLogger().Warn(message, fields...)
}

func Error(message string, err error, fields ...map[string]interface{}) {
	GetLogger().Error(message, err, fields...)
}

func Fatal(message string, err error, fields ...map[string]interface{}) {
	GetLogger().Fatal(message, err, fields...)
}

func WithContext(ctx context.Context) *ContextLogger {
	return GetLogger().WithContext(ctx)
}

func WithRequest(c *gin.Context) *RequestLogger {
	return GetLogger().WithRequest(c)
}
