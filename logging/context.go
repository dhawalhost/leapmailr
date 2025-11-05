package logging

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Context keys for logging
type contextKey string

const (
	CorrelationIDKey contextKey = "correlation_id"
	UserIDKey        contextKey = "user_id"
	RequestIDKey     contextKey = "request_id"
)

// WithCorrelationID adds a correlation ID to the context
func WithCorrelationID(ctx context.Context) context.Context {
	correlationID := uuid.New().String()
	return context.WithValue(ctx, CorrelationIDKey, correlationID)
}

// WithUserID adds a user ID to the context
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, UserIDKey, userID)
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetCorrelationID retrieves the correlation ID from context
func GetCorrelationID(ctx context.Context) string {
	if id, ok := ctx.Value(CorrelationIDKey).(string); ok {
		return id
	}
	return ""
}

// GetUserID retrieves the user ID from context
func GetUserID(ctx context.Context) string {
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}

// LoggerFromContext creates a logger with context fields (GAP-SEC-009)
func LoggerFromContext(ctx context.Context) *zap.Logger {
	logger := zap.L()

	// Add correlation ID if present
	if correlationID := GetCorrelationID(ctx); correlationID != "" {
		logger = logger.With(zap.String("correlation_id", correlationID))
	}

	// Add user ID if present
	if userID := GetUserID(ctx); userID != "" {
		logger = logger.With(zap.String("user_id", userID))
	}

	// Add request ID if present
	if requestID := GetRequestID(ctx); requestID != "" {
		logger = logger.With(zap.String("request_id", requestID))
	}

	return logger
}

// LogWithFields logs a message with structured fields
func LogWithFields(logger *zap.Logger, level zapcore.Level, msg string, fields map[string]interface{}) {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}

	switch level {
	case zapcore.DebugLevel:
		logger.Debug(msg, zapFields...)
	case zapcore.InfoLevel:
		logger.Info(msg, zapFields...)
	case zapcore.WarnLevel:
		logger.Warn(msg, zapFields...)
	case zapcore.ErrorLevel:
		logger.Error(msg, zapFields...)
	}
}

// RedactSensitiveData masks sensitive information in logs (GAP-SEC-009)
func RedactSensitiveData(data string) string {
	if len(data) <= 4 {
		return "****"
	}
	return data[:2] + "****" + data[len(data)-2:]
}

// RedactEmail masks email addresses for logging
func RedactEmail(email string) string {
	if email == "" {
		return ""
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "****"
	}
	username := parts[0]
	domain := parts[1]

	if len(username) <= 2 {
		return "**@" + domain
	}
	return username[:1] + "****" + username[len(username)-1:] + "@" + domain
}

// SecurityLog logs security-related events with enhanced context
func SecurityLog(ctx context.Context, event string, fields map[string]interface{}) {
	logger := LoggerFromContext(ctx).Named("security")

	// Add security event type
	fields["event_type"] = "security"
	fields["event"] = event

	LogWithFields(logger, zapcore.WarnLevel, event, fields)
}

// AuditLog logs audit trail events
func AuditLog(ctx context.Context, action string, resource string, fields map[string]interface{}) {
	logger := LoggerFromContext(ctx).Named("audit")

	fields["action"] = action
	fields["resource"] = resource
	fields["event_type"] = "audit"

	LogWithFields(logger, zapcore.InfoLevel, action, fields)
}

// ErrorLog logs errors with structured context
func ErrorLog(ctx context.Context, err error, msg string, fields map[string]interface{}) {
	logger := LoggerFromContext(ctx)

	if fields == nil {
		fields = make(map[string]interface{})
	}
	fields["error"] = err.Error()

	LogWithFields(logger, zapcore.ErrorLevel, msg, fields)
}
