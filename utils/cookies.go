package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// CookieConfig holds cookie configuration (GAP-SEC-014)
type CookieConfig struct {
	Name     string
	Value    string
	MaxAge   int
	Path     string
	Domain   string
	Secure   bool   // Only send over HTTPS
	HttpOnly bool   // Not accessible via JavaScript
	SameSite string // "Strict", "Lax", or "None"
}

// SetSecureCookie sets an HTTP-only, Secure, SameSite cookie
func SetSecureCookie(c *gin.Context, config CookieConfig) {
	// Map SameSite string to http.SameSite constant
	var sameSite http.SameSite
	switch config.SameSite {
	case "Strict":
		sameSite = http.SameSiteStrictMode
	case "None":
		sameSite = http.SameSiteNoneMode
	default: // "Lax"
		sameSite = http.SameSiteLaxMode
	}

	c.SetSameSite(sameSite)
	c.SetCookie(
		config.Name,
		config.Value,
		config.MaxAge,
		config.Path,
		config.Domain,
		config.Secure,
		config.HttpOnly,
	)
}

// DeleteCookie removes a cookie by setting MaxAge to -1
func DeleteCookie(c *gin.Context, name string, path string, domain string) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		name,
		"",
		-1, // MaxAge -1 deletes the cookie
		path,
		domain,
		true, // Secure
		true, // HttpOnly
	)
}

// GetCookie retrieves a cookie value
func GetCookie(c *gin.Context, name string) (string, error) {
	return c.Cookie(name)
}

// SetAuthCookies sets access token, refresh token, and CSRF token cookies
func SetAuthCookies(c *gin.Context, accessToken, refreshToken, csrfToken string, envMode string) {
	// Determine if we're in production (use Secure flag)
	isProduction := envMode == "release" || envMode == "production"

	// Use Lax for development (localhost cross-origin), Strict for production
	sameSiteMode := "Strict"
	if !isProduction {
		sameSiteMode = "Lax" // Allow cookies in cross-site GET requests (localhost:3000 -> localhost:8080)
	}

	// Access token cookie (15 minutes)
	SetSecureCookie(c, CookieConfig{
		Name:     "access_token",
		Value:    accessToken,
		MaxAge:   15 * 60, // 15 minutes
		Path:     "/",
		Domain:   "",
		Secure:   isProduction,
		HttpOnly: true,
		SameSite: sameSiteMode,
	})

	// Refresh token cookie (7 days)
	SetSecureCookie(c, CookieConfig{
		Name:     "refresh_token",
		Value:    refreshToken,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
		Path:     "/",
		Domain:   "",
		Secure:   isProduction,
		HttpOnly: true,
		SameSite: sameSiteMode,
	})

	// CSRF token cookie (24 hours, NOT HttpOnly - needs to be readable by JS)
	SetSecureCookie(c, CookieConfig{
		Name:     "csrf_token",
		Value:    csrfToken,
		MaxAge:   24 * 60 * 60, // 24 hours
		Path:     "/",
		Domain:   "",
		Secure:   isProduction,
		HttpOnly: false, // JS needs to read this to send in X-CSRF-Token header
		SameSite: sameSiteMode,
	})
}

// DeleteAuthCookies removes all authentication cookies
func DeleteAuthCookies(c *gin.Context) {
	DeleteCookie(c, "access_token", "/", "")
	DeleteCookie(c, "refresh_token", "/", "")
	DeleteCookie(c, "csrf_token", "/", "")
}
