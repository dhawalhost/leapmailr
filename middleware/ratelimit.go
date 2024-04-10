package middleware

import (
	"net/http"

	"github.com/dhawalhost/leapmailr/ratelimit"

	"github.com/gin-gonic/gin"
)

func LimitMiddleware() gin.HandlerFunc {
	limiter := ratelimit.NewIPRateLimiter(1, 2)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := limiter.GetLimiter(ip)
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
		}
		c.Next()
	}
}
