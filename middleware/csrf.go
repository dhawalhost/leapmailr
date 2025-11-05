package middleware

import (
	"net/http"
	"strings"

	"github.com/dhawalhost/leapmailr/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CSRFProtection validates CSRF tokens for state-changing operations (GAP-SEC-014)
func CSRFProtection() gin.HandlerFunc {
	csrfService := utils.GetCSRFService()

	return func(c *gin.Context) {
		// Skip CSRF validation for safe methods
		if c.Request.Method == "GET" || c.Request.Method == "HEAD" || c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Skip CSRF for public endpoints (login, register)
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/v1/auth/login") ||
			strings.HasPrefix(path, "/api/v1/auth/register") ||
			strings.HasPrefix(path, "/api/v1/auth/refresh") ||
			strings.HasPrefix(path, "/health") ||
			strings.HasPrefix(path, "/metrics") {
			c.Next()
			return
		}

		// Get authentication method from context (set by auth middleware)
		authMethod, authMethodExists := c.Get("auth_method")

		// Skip CSRF validation for non-browser authentication methods
		// CSRF protection is only needed when using cookies (browser-based auth)
		if authMethodExists {
			method := authMethod.(string)
			// Skip CSRF for: API keys, SDK usage (public/private key pairs)
			// These don't use cookies, so CSRF attacks don't apply
			if method == "api_key" || method == "api_key_pair_full" || method == "public_key_only" {
				c.Next()
				return
			}

			// For JWT authentication, check if using Authorization header (not cookies)
			// If Authorization header is present, it's from Postman/SDK, not browser cookies
			if method == "jwt" && c.GetHeader("Authorization") != "" {
				// JWT via Authorization header = no CSRF needed (not relying on cookies)
				c.Next()
				return
			}
		}

		// Get CSRF token from header
		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "CSRF token missing",
			})
			c.Abort()
			return
		}

		// Get user ID from context (set by auth middleware)
		userIDValue, exists := c.Get("userID")
		if !exists {
			// If no userID in context, skip CSRF validation
			// This happens for public endpoints or when auth middleware hasn't run yet
			c.Next()
			return
		}

		userID, ok := userIDValue.(string)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user ID format",
			})
			c.Abort()
			return
		}

		// Parse user ID
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Invalid user ID",
			})
			c.Abort()
			return
		}

		// Validate CSRF token
		if err := csrfService.ValidateToken(csrfToken, parsedUserID); err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Invalid or expired CSRF token",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
