package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CorsMiddleware handles CORS settings with strict origin validation (GAP-SEC-011)
func CorsMiddleware(envMode string) gin.HandlerFunc {
	allowedOrigins := getAllowedOrigins(envMode)

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		isAllowed := isOriginInWhitelist(origin, allowedOrigins)

		// Set CORS headers for allowed origins
		if isAllowed {
			setCORSHeaders(c, origin)
		} else if origin != "" {
			rejectOrigin(c)
		}

		// Handle preflight OPTIONS request
		if c.Request.Method == "OPTIONS" {
			handlePreflightRequest(c, isAllowed)
			return
		}

		c.Next()
	}
}

// isOriginInWhitelist checks if origin is in allowed list
func isOriginInWhitelist(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

// setCORSHeaders sets all CORS headers for allowed origins
func setCORSHeaders(c *gin.Context, origin string) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-API-Key, X-Auth-Token, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
	c.Writer.Header().Set("Access-Control-Max-Age", "86400")
}

// rejectOrigin handles rejected origins
func rejectOrigin(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "")
	// Optionally log: fmt.Printf("CORS: Rejected origin: %s\n", c.Request.Header.Get("Origin"))
}

// handlePreflightRequest handles OPTIONS preflight requests
func handlePreflightRequest(c *gin.Context, isAllowed bool) {
	if isAllowed {
		c.AbortWithStatus(204)
	} else {
		c.AbortWithStatus(403) // Forbidden for non-whitelisted origins
	}
}

// getAllowedOrigins returns the list of allowed origins based on environment
func getAllowedOrigins(envMode string) []string {
	// Development allows localhost and common dev ports
	if envMode == "development" || envMode == "dev" {
		return []string{
			"http://localhost:3000",
			"http://localhost:3001",
			"http://localhost:8080",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:3001",
			"http://127.0.0.1:8080",
		}
	}

	// Production - strict whitelist
	// TODO: Update these with your actual production domains
	return []string{
		"https://app.leapmailr.com",
		"https://leapmailr.com",
		"https://www.leapmailr.com",
		// Add staging environment if needed
		// "https://staging.leapmailr.com",
	}
}

// AddAllowedOrigin adds a custom allowed origin at runtime
// Use this for dynamic origin management (e.g., white-label domains)
func AddAllowedOrigin(envMode string, origin string) {
	// This would require a thread-safe map for runtime updates
	// For now, origins are statically defined above
	// Future enhancement: Store in database or config service
}

// IsOriginAllowed checks if an origin is in the whitelist
func IsOriginAllowed(origin string, envMode string) bool {
	if origin == "" {
		return false
	}

	allowedOrigins := getAllowedOrigins(envMode)
	for _, allowed := range allowedOrigins {
		if origin == allowed {
			return true
		}
		// Support wildcard subdomains in production (e.g., *.leapmailr.com)
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}

	return false
}
