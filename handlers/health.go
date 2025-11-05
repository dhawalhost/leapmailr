package handlers

import (
	"net/http"
	"time"

	"github.com/dhawalhost/leapmailr/database"
	"github.com/dhawalhost/leapmailr/monitoring"
	"github.com/gin-gonic/gin"
)

var (
	healthMonitor *monitoring.HealthMonitor
	startTime     time.Time
	appVersion    = "1.0.0" // TODO: Load from build info
)

func init() {
	startTime = time.Now()
	healthMonitor = monitoring.NewHealthMonitor(5 * time.Second) // 5 second cache TTL
}

// InitializeHealthChecks registers all health checkers
func InitializeHealthChecks() {
	// Register database health checker
	if db := database.GetDB(); db != nil {
		healthMonitor.RegisterChecker("database", monitoring.NewDatabaseHealthChecker(db))
	}

	// Register other health checkers
	healthMonitor.RegisterChecker("redis", monitoring.NewRedisHealthChecker(""))
	healthMonitor.RegisterChecker("disk_space", monitoring.NewDiskSpaceHealthChecker("/", 80, 90))
	healthMonitor.RegisterChecker("memory", monitoring.NewMemoryHealthChecker(1<<30, 2<<30)) // 1GB warning, 2GB critical
}

// HandleHealthCheck handles the health check endpoint (detailed)
func HandleHealthCheck(c *gin.Context) {
	components := healthMonitor.CheckAll()
	overallStatus := healthMonitor.GetOverallStatus()

	uptime := int64(time.Since(startTime).Seconds())

	response := monitoring.HealthResponse{
		Status:     overallStatus,
		Timestamp:  time.Now(),
		Version:    appVersion,
		Uptime:     uptime,
		Components: components,
	}

	statusCode := http.StatusOK
	if overallStatus == monitoring.HealthStatusDegraded {
		statusCode = http.StatusOK // Still considered OK for load balancers
	} else if overallStatus == monitoring.HealthStatusUnhealthy {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// HandleReadinessCheck handles the readiness probe endpoint
// Used by Kubernetes/orchestrators to determine if service can receive traffic
func HandleReadinessCheck(c *gin.Context) {
	components := healthMonitor.CheckAll()

	// Service is ready if no components are unhealthy
	ready := true
	for _, health := range components {
		if health.Status == monitoring.HealthStatusUnhealthy {
			ready = false
			break
		}
	}

	response := monitoring.ReadinessResponse{
		Ready:      ready,
		Timestamp:  time.Now(),
		Components: components,
	}

	statusCode := http.StatusOK
	if !ready {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// HandleLivenessCheck handles the liveness probe endpoint
// Used by Kubernetes/orchestrators to determine if service should be restarted
func HandleLivenessCheck(c *gin.Context) {
	response := monitoring.LivenessResponse{
		Alive:     true,
		Timestamp: time.Now(),
	}

	c.JSON(http.StatusOK, response)
}
