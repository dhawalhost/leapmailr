package monitoring

import (
	"sync"
	"time"
)

// HealthStatus represents the health status of a component
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Name      string       `json:"name"`
	Status    HealthStatus `json:"status"`
	Message   string       `json:"message,omitempty"`
	LastCheck time.Time    `json:"last_check"`
	Duration  int64        `json:"duration_ms"`
}

// HealthChecker interface for components that can report health
type HealthChecker interface {
	CheckHealth() ComponentHealth
}

// HealthMonitor manages health checks for all components
type HealthMonitor struct {
	checkers map[string]HealthChecker
	cache    map[string]ComponentHealth
	mu       sync.RWMutex
	cacheTTL time.Duration
}

// NewHealthMonitor creates a new health monitor
func NewHealthMonitor(cacheTTL time.Duration) *HealthMonitor {
	return &HealthMonitor{
		checkers: make(map[string]HealthChecker),
		cache:    make(map[string]ComponentHealth),
		cacheTTL: cacheTTL,
	}
}

// RegisterChecker registers a health checker for a component
func (hm *HealthMonitor) RegisterChecker(name string, checker HealthChecker) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.checkers[name] = checker
}

// CheckAll performs health checks on all registered components
func (hm *HealthMonitor) CheckAll() map[string]ComponentHealth {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	results := make(map[string]ComponentHealth)
	now := time.Now()

	for name, checker := range hm.checkers {
		// Check if cached result is still valid
		if cached, ok := hm.cache[name]; ok {
			if now.Sub(cached.LastCheck) < hm.cacheTTL {
				results[name] = cached
				continue
			}
		}

		// Perform health check
		health := checker.CheckHealth()
		hm.cache[name] = health
		results[name] = health
	}

	return results
}

// GetOverallStatus returns the overall health status
func (hm *HealthMonitor) GetOverallStatus() HealthStatus {
	results := hm.CheckAll()

	hasUnhealthy := false
	hasDegraded := false

	for _, health := range results {
		switch health.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	}
	if hasDegraded {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

// HealthResponse represents the complete health check response
type HealthResponse struct {
	Status     HealthStatus               `json:"status"`
	Timestamp  time.Time                  `json:"timestamp"`
	Version    string                     `json:"version,omitempty"`
	Uptime     int64                      `json:"uptime_seconds"`
	Components map[string]ComponentHealth `json:"components"`
}

// ReadinessResponse represents the readiness probe response
type ReadinessResponse struct {
	Ready      bool                       `json:"ready"`
	Timestamp  time.Time                  `json:"timestamp"`
	Components map[string]ComponentHealth `json:"components,omitempty"`
}

// LivenessResponse represents the liveness probe response
type LivenessResponse struct {
	Alive     bool      `json:"alive"`
	Timestamp time.Time `json:"timestamp"`
}
