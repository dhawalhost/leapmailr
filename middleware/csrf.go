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
		// Early exit for requests that don't need CSRF protection
		if shouldSkipCSRF(c) {
			c.Next()
			return
		}

		// Extract and validate CSRF token
		csrfToken := c.GetHeader("X-CSRF-Token")
		if csrfToken == "" {
			abortWithCSRFError(c, "CSRF token missing")
			return
		}

		// Extract and parse user ID
		parsedUserID, err := extractUserID(c)
		if err != nil {
			abortWithCSRFError(c, err.Error())
			return
		}

		// No user ID means public endpoint - skip validation
		if parsedUserID == uuid.Nil {
			c.Next()
			return
		}

		// Validate CSRF token
		if err := csrfService.ValidateToken(csrfToken, parsedUserID); err != nil {
			abortWithCSRFError(c, "Invalid or expired CSRF token")
			return
		}

		c.Next()
	}
}

// shouldSkipCSRF determines if CSRF protection should be skipped for the request
func shouldSkipCSRF(c *gin.Context) bool {
	// Skip for safe HTTP methods
	if isSafeMethod(c.Request.Method) {
		return true
	}

	// Skip for public endpoints
	if isPublicEndpoint(c.Request.URL.Path) {
		return true
	}

	// Skip for non-browser authentication methods
	return isNonBrowserAuth(c)
}

// isSafeMethod checks if the HTTP method is safe (doesn't modify state)
func isSafeMethod(method string) bool {
	return method == "GET" || method == "HEAD" || method == "OPTIONS"
}

// isPublicEndpoint checks if the path is a public endpoint that doesn't need CSRF protection
func isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/refresh",
		"/health",
		"/metrics",
	}

	for _, publicPath := range publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}

// isNonBrowserAuth checks if the request uses non-browser authentication (API keys, SDK)
func isNonBrowserAuth(c *gin.Context) bool {
	authMethod, authMethodExists := c.Get("auth_method")
	if !authMethodExists {
		return false
	}

	method := authMethod.(string)

	// API keys and SDK authentication don't need CSRF protection
	if method == "api_key" || method == "api_key_pair_full" || method == "public_key_only" {
		return true
	}

	// JWT via Authorization header (not cookies) doesn't need CSRF protection
	if method == "jwt" && c.GetHeader("Authorization") != "" {
		return true
	}

	return false
}

// extractUserID extracts and parses the user ID from the request context
func extractUserID(c *gin.Context) (uuid.UUID, error) {
	userIDValue, exists := c.Get("userID")
	if !exists {
		// No user ID means public endpoint - return nil UUID
		return uuid.Nil, nil
	}

	userID, ok := userIDValue.(string)
	if !ok {
		return uuid.Nil, &csrfError{message: "Invalid user ID format"}
	}

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return uuid.Nil, &csrfError{message: "Invalid user ID"}
	}

	return parsedUserID, nil
}

// csrfError represents a CSRF validation error
type csrfError struct {
	message string
}

func (e *csrfError) Error() string {
	return e.message
}

// abortWithCSRFError aborts the request with a CSRF error
func abortWithCSRFError(c *gin.Context, message string) {
	statusCode := http.StatusForbidden
	if strings.Contains(message, "format") || strings.Contains(message, "Invalid user ID") {
		statusCode = http.StatusInternalServerError
	}

	c.JSON(statusCode, gin.H{
		"error": message,
	})
	c.Abort()
}
