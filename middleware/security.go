package middleware

import (
	"io"
	"net/http"

	"github.com/dhawalhost/leapmailr/utils"
	"github.com/gin-gonic/gin"
)

// SecurityHeaders adds comprehensive security headers to all responses (GAP-SEC-012)
// This middleware implements security best practices to protect against common web vulnerabilities
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// HSTS - Force HTTPS for 1 year including subdomains
		// Prevents SSL stripping attacks and ensures all connections use HTTPS
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// CSP - Content Security Policy
		// Prevents XSS attacks by controlling which resources can be loaded
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net https://unpkg.com; " +
			"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
			"font-src 'self' https://fonts.gstatic.com; " +
			"img-src 'self' data: https:; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"
		c.Header("Content-Security-Policy", csp)

		// X-Frame-Options - Prevents clickjacking attacks
		// DENY prevents the page from being embedded in any frame
		c.Header("X-Frame-Options", "DENY")

		// X-Content-Type-Options - Prevents MIME type sniffing
		// Forces browsers to respect the declared Content-Type
		c.Header("X-Content-Type-Options", "nosniff")

		// X-XSS-Protection - Enable browser's XSS filter (legacy but doesn't hurt)
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy - Control information sent in Referer header
		// no-referrer-when-downgrade: Send full URL to HTTPS, nothing to HTTP
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy - Control browser features (replaces Feature-Policy)
		// Disable potentially dangerous features
		permissions := "geolocation=(), " +
			"microphone=(), " +
			"camera=(), " +
			"payment=(), " +
			"usb=(), " +
			"magnetometer=(), " +
			"gyroscope=(), " +
			"accelerometer=()"
		c.Header("Permissions-Policy", permissions)

		// X-Permitted-Cross-Domain-Policies - Adobe cross-domain policy
		c.Header("X-Permitted-Cross-Domain-Policies", "none")

		// X-Download-Options - Prevent IE from executing downloads in site's context
		c.Header("X-Download-Options", "noopen")

		// Cache-Control for sensitive endpoints
		// Prevent caching of authenticated responses
		if c.Request.Header.Get("Authorization") != "" || c.Request.Header.Get("X-API-Key") != "" {
			c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")
		}

		c.Next()
	}
}

// RedirectToHTTPS middleware redirects HTTP requests to HTTPS
// Only active when not in local development mode
func RedirectToHTTPS(envMode string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip redirect in development mode
		if envMode == "development" || envMode == "dev" {
			c.Next()
			return
		}

		// Check if request is already HTTPS
		if c.Request.TLS != nil {
			c.Next()
			return
		}

		// Check X-Forwarded-Proto header (set by load balancers/proxies)
		proto := c.Request.Header.Get("X-Forwarded-Proto")
		if proto == "https" {
			c.Next()
			return
		}

		// SECURITY: Safely construct HTTPS URL with validation
		// Prevents open redirect attacks via Host header injection
		httpsURL, err := utils.BuildSecureHTTPSURL(c.Request.Host, c.Request.RequestURI)
		if err != nil {
			// If URL construction fails, reject the request
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid request headers",
			})
			c.Abort()
			return
		}

		// Redirect to validated HTTPS URL
		c.Redirect(http.StatusMovedPermanently, httpsURL) // 301 Permanent Redirect
		c.Abort()
	}
}

// TrustedProxyHeaders sets up trusted proxy configuration
// This is important when running behind load balancers or reverse proxies
func TrustedProxyHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get real client IP from proxy headers
		// Priority: X-Real-IP > X-Forwarded-For > RemoteAddr
		if realIP := c.Request.Header.Get("X-Real-IP"); realIP != "" {
			c.Set("ClientIP", realIP)
		} else if forwardedFor := c.Request.Header.Get("X-Forwarded-For"); forwardedFor != "" {
			// X-Forwarded-For can contain multiple IPs: "client, proxy1, proxy2"
			// Take the first one (the original client)
			c.Set("ClientIP", forwardedFor)
		} else {
			c.Set("ClientIP", c.ClientIP())
		}

		c.Next()
	}
}

// RequestSizeLimit limits the size of incoming requests
// Prevents DOS attacks via large payloads
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

// MaxBytesReader wraps http.MaxBytesReader for Gin
type maxBytesReader struct {
	r http.ResponseWriter
	b io.ReadCloser
	n int64
}

func MaxBytesReader(w http.ResponseWriter, r io.ReadCloser, n int64) io.ReadCloser {
	return &maxBytesReader{w, r, n}
}

func (mbr *maxBytesReader) Read(p []byte) (n int, err error) {
	if mbr.n <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > mbr.n {
		p = p[0:mbr.n]
	}
	n, err = mbr.b.Read(p)
	mbr.n -= int64(n)
	return
}

func (mbr *maxBytesReader) Close() error {
	return mbr.b.Close()
}
