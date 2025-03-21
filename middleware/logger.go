package middleware

import (
	"time"

	"github.com/dhawalhost/leapmailr/logging"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggerMiddleware is a middleware that logs the server request
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		c.Next()
		end := time.Now()
		latency := end.Sub(start)
		if raw != "" {
			path = path + "?" + raw
		}
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		logging.GetApiLogger().Info("request", zap.String("path", path), zap.String("method", method), zap.Int("status", status), zap.String("clientIP", clientIP), zap.Duration("latency", latency))
	}
}
