package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Authentication metrics
	AuthSuccessTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_success_total",
			Help: "Total number of successful authentications",
		},
		[]string{"method", "user_role"},
	)

	AuthFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_failures_total",
			Help: "Total number of authentication failures",
		},
		[]string{"method", "reason"},
	)

	PasswordResetRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "password_reset_requests_total",
			Help: "Total number of password reset requests",
		},
		[]string{"status"},
	)

	// Database metrics
	DatabaseConnections = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
		[]string{"database"},
	)

	DatabaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	DatabaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "status"},
	)

	// Application metrics
	ActiveUsersTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Total number of active users in the system",
		},
	)

	UserSessionsTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "user_sessions_total",
			Help: "Total number of user sessions",
		},
	)

	// Business metrics
	DashboardViewsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dashboard_views_total",
			Help: "Total number of dashboard views",
		},
		[]string{"user_role", "dashboard_type"},
	)

	CompaniesTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "companies_total",
			Help: "Total number of companies in the system",
		},
	)
)
