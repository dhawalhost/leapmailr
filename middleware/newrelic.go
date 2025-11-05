package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/newrelic/go-agent/v3/integrations/nrgin"
	"github.com/newrelic/go-agent/v3/newrelic"
)

// NewRelicMiddleware creates a Gin middleware for New Relic APM transaction tracking
// If nrApp is nil, returns a no-op middleware (allows app to run without New Relic)
func NewRelicMiddleware(nrApp *newrelic.Application) gin.HandlerFunc {
	// If New Relic is not configured, return a no-op middleware
	if nrApp == nil {
		return func(c *gin.Context) {
			c.Next()
		}
	}

	// Use the official New Relic Gin integration
	return nrgin.Middleware(nrApp)
}
