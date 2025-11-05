package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RateLimitConfig defines rate limiting rules
type RateLimitConfig struct {
	PerIP       int            // Requests per IP per window
	PerUser     int            // Requests per authenticated user per window
	PerEndpoint map[string]int // Requests per specific endpoint per window
	Window      time.Duration  // Time window for rate limiting
}

// DefaultRateLimitConfig returns sensible default rate limits
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		PerIP:   100,             // 100 requests per minute per IP
		PerUser: 1000,            // 1000 requests per hour per user
		Window:  1 * time.Minute, // 1 minute window
		PerEndpoint: map[string]int{
			"/api/v1/auth/login":      5,  // 5 login attempts per minute
			"/api/v1/auth/register":   3,  // 3 registration attempts per minute
			"/api/v1/email/send":      50, // 50 emails per minute
			"/api/v1/email/send-bulk": 10, // 10 bulk requests per minute
		},
	}
}

// rateLimiter tracks request counts
type rateLimiter struct {
	mu       sync.RWMutex
	requests map[string]*rateLimitEntry
	config   *RateLimitConfig
}

// rateLimitEntry stores request count and window start time
type rateLimitEntry struct {
	count       int
	windowStart time.Time
}

var (
	limiter *rateLimiter
	once    sync.Once
)

// initRateLimiter initializes the global rate limiter
func initRateLimiter(config *RateLimitConfig) {
	once.Do(func() {
		limiter = &rateLimiter{
			requests: make(map[string]*rateLimitEntry),
			config:   config,
		}

		// Start cleanup goroutine
		go limiter.cleanupOldEntries()
	})
}

// EnhancedRateLimiter provides multi-tier rate limiting (GAP-SEC-010)
// Implements per-IP, per-user, and per-endpoint rate limits
func EnhancedRateLimiter(config *RateLimitConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	initRateLimiter(config)

	return func(c *gin.Context) {
		// Get identifiers
		ip := c.ClientIP()
		endpoint := c.Request.URL.Path

		// Check per-IP rate limit
		ipKey := fmt.Sprintf("ip:%s", ip)
		if !limiter.allow(ipKey, config.PerIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests from your IP address. Please try again later.",
				"retry_after": int(config.Window.Seconds()),
			})
			c.Abort()
			return
		}

		// Check per-endpoint rate limit
		if limit, exists := config.PerEndpoint[endpoint]; exists {
			endpointKey := fmt.Sprintf("endpoint:%s:%s", ip, endpoint)
			if !limiter.allow(endpointKey, limit) {
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":       fmt.Sprintf("Too many requests to this endpoint. Limit: %d per %s", limit, config.Window),
					"retry_after": int(config.Window.Seconds()),
				})
				c.Abort()
				return
			}
		}

		// Check per-user rate limit (if authenticated)
		if userID, exists := c.Get("userID"); exists {
			if uid, ok := userID.(uuid.UUID); ok {
				userKey := fmt.Sprintf("user:%s", uid.String())
				if !limiter.allow(userKey, config.PerUser) {
					c.JSON(http.StatusTooManyRequests, gin.H{
						"error":       "You have exceeded your request quota. Please try again later.",
						"retry_after": int(config.Window.Seconds()),
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// allow checks if a request should be allowed
func (rl *rateLimiter) allow(key string, limit int) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.requests[key]

	if !exists {
		// First request for this key
		rl.requests[key] = &rateLimitEntry{
			count:       1,
			windowStart: now,
		}
		return true
	}

	// Check if we're still in the same window
	if now.Sub(entry.windowStart) < rl.config.Window {
		// Same window - check if limit exceeded
		if entry.count >= limit {
			return false
		}
		entry.count++
		return true
	}

	// New window - reset counter
	entry.count = 1
	entry.windowStart = now
	return true
}

// getRemainingRequests returns the number of requests remaining for a key
func (rl *rateLimiter) getRemainingRequests(key string, limit int) int {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	entry, exists := rl.requests[key]
	if !exists {
		return limit
	}

	now := time.Now()
	if now.Sub(entry.windowStart) >= rl.config.Window {
		return limit
	}

	remaining := limit - entry.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// cleanupOldEntries periodically removes expired entries to prevent memory leaks
func (rl *rateLimiter) cleanupOldEntries() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, entry := range rl.requests {
			// Remove entries older than 2x the window duration
			if now.Sub(entry.windowStart) > 2*rl.config.Window {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware wraps the simple rate limiter for backward compatibility
func RateLimitMiddleware() gin.HandlerFunc {
	// Use default configuration
	return EnhancedRateLimiter(nil)
}

// APIKeyRateLimiter implements plan-based rate limiting for API keys
// Different plans get different quotas
func APIKeyRateLimiter() gin.HandlerFunc {
	limits := map[string]int{
		"free":       100,    // 100 requests/hour
		"basic":      1000,   // 1000 requests/hour
		"pro":        10000,  // 10000 requests/hour
		"enterprise": 100000, // 100000 requests/hour
	}

	return func(c *gin.Context) {
		// Get user plan from context (set by auth middleware)
		plan, exists := c.Get("userPlan")
		if !exists {
			plan = "free" // Default to free plan
		}

		planStr, ok := plan.(string)
		if !ok {
			planStr = "free"
		}

		// Get limit for this plan
		limit, exists := limits[planStr]
		if !exists {
			limit = limits["free"]
		}

		// Check rate limit using user ID
		if userID, exists := c.Get("userID"); exists {
			if uid, ok := userID.(uuid.UUID); ok {
				key := fmt.Sprintf("apikey:%s", uid.String())
				if !limiter.allow(key, limit) {
					remaining := limiter.getRemainingRequests(key, limit)
					c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
					c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
					c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(limiter.config.Window).Unix()))

					c.JSON(http.StatusTooManyRequests, gin.H{
						"error":       fmt.Sprintf("API rate limit exceeded for %s plan. Limit: %d requests per hour", planStr, limit),
						"upgrade_url": "https://leapmailr.com/pricing",
					})
					c.Abort()
					return
				}

				// Add rate limit headers to successful responses
				remaining := limiter.getRemainingRequests(key, limit)
				c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
				c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
				c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(limiter.config.Window).Unix()))
			}
		}

		c.Next()
	}
}

// GetRateLimitInfo returns current rate limit status for a key
func GetRateLimitInfo(key string, limit int) map[string]interface{} {
	if limiter == nil {
		return map[string]interface{}{
			"limit":     limit,
			"remaining": limit,
			"reset":     time.Now().Add(1 * time.Minute).Unix(),
		}
	}

	remaining := limiter.getRemainingRequests(key, limit)
	return map[string]interface{}{
		"limit":     limit,
		"remaining": remaining,
		"reset":     time.Now().Add(limiter.config.Window).Unix(),
	}
}
