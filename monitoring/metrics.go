package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	HTTPRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000, 10000000},
		},
		[]string{"method", "endpoint"},
	)

	HTTPResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000, 10000000},
		},
		[]string{"method", "endpoint"},
	)

	// Email metrics
	EmailsSentTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "emails_sent_total",
			Help: "Total number of emails sent",
		},
		[]string{"provider", "status"},
	)

	EmailSendDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "email_send_duration_seconds",
			Help:    "Email send operation duration in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"provider"},
	)

	EmailQueueLength = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "email_queue_length",
			Help: "Current number of emails in the send queue",
		},
	)

	// Database metrics
	DBConnectionsInUse = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_in_use",
			Help: "Number of database connections currently in use",
		},
	)

	DBConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	DBConnectionsWaitCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "db_connections_wait_count_total",
			Help: "Total number of times waited for a database connection",
		},
	)

	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1, 5},
		},
		[]string{"operation"},
	)

	// Authentication metrics
	AuthAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"method", "status"},
	)

	AuthFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_failures_total",
			Help: "Total number of authentication failures",
		},
		[]string{"method", "reason"},
	)

	ActiveSessionsCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_sessions_count",
			Help: "Number of currently active user sessions",
		},
	)

	LockedAccountsCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "locked_accounts_count",
			Help: "Number of currently locked accounts",
		},
	)

	// Rate limiting metrics
	RateLimitHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"tier", "endpoint"},
	)

	RateLimitExceededTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_exceeded_total",
			Help: "Total number of requests exceeding rate limits",
		},
		[]string{"tier", "endpoint"},
	)

	// API Key metrics
	APIKeyRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_key_requests_total",
			Help: "Total number of API key authenticated requests",
		},
		[]string{"key_id", "status"},
	)

	// Template metrics
	TemplateRenderDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "template_render_duration_seconds",
			Help:    "Template rendering duration in seconds",
			Buckets: []float64{0.001, 0.01, 0.1, 0.5, 1},
		},
		[]string{"template_id"},
	)

	// Error metrics
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "component"},
	)

	// Business metrics
	ActiveUsersCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users_count",
			Help: "Number of active users",
		},
	)

	ContactsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "contacts_total",
			Help: "Total number of contacts in the system",
		},
	)

	TemplatesTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "templates_total",
			Help: "Total number of email templates",
		},
	)

	// System metrics
	AppInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "app_info",
			Help: "Application information",
		},
		[]string{"version", "env"},
	)

	AppUptimeSeconds = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_uptime_seconds",
			Help: "Application uptime in seconds",
		},
	)
)

// InitMetrics initializes application metrics
func InitMetrics(version, env string) {
	AppInfo.WithLabelValues(version, env).Set(1)
}
