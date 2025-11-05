package middleware

import (
	"time"

	"github.com/dhawalhost/leapmailr/logging"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// StructuredLogger middleware adds correlation IDs and structured logging (GAP-SEC-009)
func StructuredLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Generate correlation ID
		correlationID := uuid.New().String()
		c.Set("correlation_id", correlationID)
		c.Header("X-Correlation-ID", correlationID)

		// Generate request ID
		requestID := uuid.New().String()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		// Add to context
		ctx := c.Request.Context()
		ctx = logging.WithCorrelationID(ctx)
		ctx = logging.WithRequestID(ctx, requestID)
		c.Request = c.Request.WithContext(ctx)

		// Get user ID if authenticated
		if userID, exists := c.Get("userID"); exists {
			ctx = logging.WithUserID(ctx, userID.(string))
			c.Request = c.Request.WithContext(ctx)
		}

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		// Get logger with context
		logger := logging.LoggerFromContext(c.Request.Context()).Named("http")

		// Get user ID from context if available
		userID := ""
		if uid, exists := c.Get("userID"); exists {
			userID = uid.(string)
		}

		// Structured log fields
		fields := []zap.Field{
			zap.String("correlation_id", correlationID),
			zap.String("request_id", requestID),
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
		}

		if userID != "" {
			fields = append(fields, zap.String("user_id", userID))
		}

		// Add error if present
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.String()))
		}

		// Log based on status code
		if statusCode >= 500 {
			logger.Error("Server error", fields...)
		} else if statusCode >= 400 {
			logger.Warn("Client error", fields...)
		} else {
			logger.Info("Request completed", fields...)
		}
	}
}
