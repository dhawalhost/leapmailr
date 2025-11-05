package middleware

import (
	"strconv"
	"time"

	"github.com/dhawalhost/leapmailr/monitoring"
	"github.com/gin-gonic/gin"
)

// PrometheusMetrics middleware collects HTTP metrics for Prometheus
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Request size
		requestSize := float64(c.Request.ContentLength)
		if requestSize > 0 {
			monitoring.HTTPRequestSize.WithLabelValues(
				c.Request.Method,
				path,
			).Observe(requestSize)
		}

		// Process request
		c.Next()

		// Duration
		duration := time.Since(start).Seconds()
		monitoring.HTTPRequestDuration.WithLabelValues(
			c.Request.Method,
			path,
		).Observe(duration)

		// Total requests counter
		status := strconv.Itoa(c.Writer.Status())
		monitoring.HTTPRequestsTotal.WithLabelValues(
			c.Request.Method,
			path,
			status,
		).Inc()

		// Response size
		responseSize := float64(c.Writer.Size())
		if responseSize > 0 {
			monitoring.HTTPResponseSize.WithLabelValues(
				c.Request.Method,
				path,
			).Observe(responseSize)
		}

		// Error tracking
		if c.Writer.Status() >= 400 {
			errorType := "client_error"
			if c.Writer.Status() >= 500 {
				errorType = "server_error"
			}
			monitoring.ErrorsTotal.WithLabelValues(errorType, "http").Inc()
		}
	}
}
