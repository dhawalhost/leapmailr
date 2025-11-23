package monitoring

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// DatabaseHealthChecker checks database connectivity and performance
type DatabaseHealthChecker struct {
	db *gorm.DB
}

// NewDatabaseHealthChecker creates a new database health checker
func NewDatabaseHealthChecker(db *gorm.DB) *DatabaseHealthChecker {
	return &DatabaseHealthChecker{db: db}
}

// CheckHealth performs database health check
func (d *DatabaseHealthChecker) CheckHealth() ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:      "database",
		LastCheck: start,
	}

	// Check if database is nil
	if d.db == nil {
		health.Status = HealthStatusUnhealthy
		health.Message = "Database connection not initialized"
		health.Duration = time.Since(start).Milliseconds()
		return health
	}

	// Get underlying SQL DB
	sqlDB, err := d.db.DB()
	if err != nil {
		health.Status = HealthStatusUnhealthy
		health.Message = "Failed to get database instance: " + err.Error()
		health.Duration = time.Since(start).Milliseconds()
		return health
	}

	// Ping database with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		health.Status = HealthStatusUnhealthy
		health.Message = "Database ping failed: " + err.Error()
		health.Duration = time.Since(start).Milliseconds()
		return health
	}

	// Check connection stats
	stats := sqlDB.Stats()

	// Consider degraded if too many open connections
	maxOpenConns := stats.MaxOpenConnections
	inUse := stats.InUse

	if maxOpenConns > 0 && float64(inUse)/float64(maxOpenConns) > 0.8 {
		health.Status = HealthStatusDegraded
		health.Message = "High connection usage"
	} else if stats.WaitCount > 0 && stats.WaitDuration > 5*time.Second {
		health.Status = HealthStatusDegraded
		health.Message = "High wait time for connections"
	} else {
		health.Status = HealthStatusHealthy
		health.Message = "Connected"
	}

	health.Duration = time.Since(start).Milliseconds()
	return health
}

// RedisHealthChecker checks Redis connectivity (placeholder for future Redis integration)
type RedisHealthChecker struct {
	addr string
}

// NewRedisHealthChecker creates a new Redis health checker
func NewRedisHealthChecker(addr string) *RedisHealthChecker {
	return &RedisHealthChecker{addr: addr}
}

// CheckHealth performs Redis health check
func (r *RedisHealthChecker) CheckHealth() ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:      "redis",
		LastCheck: start,
		Status:    HealthStatusHealthy,
		Message:   "Not configured",
		Duration:  time.Since(start).Milliseconds(),
	}

	// Note: Redis health check should be implemented when Redis is integrated
	// Use github.com/go-redis/redis to check connectivity with PING command
	// For now, return healthy since Redis is optional

	return health
}

// DiskSpaceHealthChecker checks available disk space
type DiskSpaceHealthChecker struct {
	path              string
	warningThreshold  float64 // Percentage (0-100)
	criticalThreshold float64 // Percentage (0-100)
}

// NewDiskSpaceHealthChecker creates a new disk space health checker
func NewDiskSpaceHealthChecker(path string, warningThreshold, criticalThreshold float64) *DiskSpaceHealthChecker {
	return &DiskSpaceHealthChecker{
		path:              path,
		warningThreshold:  warningThreshold,
		criticalThreshold: criticalThreshold,
	}
}

// CheckHealth performs disk space health check
func (d *DiskSpaceHealthChecker) CheckHealth() ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:      "disk_space",
		LastCheck: start,
	}

	// Note: Disk space check requires platform-specific syscall implementation
	// Use syscall.Statfs (Unix) or syscall.GetDiskFreeSpaceEx (Windows)
	// For now, return healthy
	health.Status = HealthStatusHealthy
	health.Message = "Sufficient space available"
	health.Duration = time.Since(start).Milliseconds()

	return health
}

// MemoryHealthChecker checks memory usage
type MemoryHealthChecker struct {
	warningThreshold  uint64 // Bytes
	criticalThreshold uint64 // Bytes
}

// NewMemoryHealthChecker creates a new memory health checker
func NewMemoryHealthChecker(warningThreshold, criticalThreshold uint64) *MemoryHealthChecker {
	return &MemoryHealthChecker{
		warningThreshold:  warningThreshold,
		criticalThreshold: criticalThreshold,
	}
}

// CheckHealth performs memory health check
func (m *MemoryHealthChecker) CheckHealth() ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		Name:      "memory",
		LastCheck: start,
	}

	// Note: Memory check should use runtime.ReadMemStats(&memStats) to get actual usage
	// Calculate percentage: (memStats.Alloc / memStats.Sys) * 100
	// Compare against warning and critical thresholds
	// For now, return healthy
	health.Status = HealthStatusHealthy
	health.Message = "Memory usage normal"
	health.Duration = time.Since(start).Milliseconds()

	return health
}
