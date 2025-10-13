package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Audit metrics for monitoring audit log activity
var (
	// AuditActionsTotal counts the total number of audit actions
	AuditActionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "audit_actions_total",
			Help: "Total number of audit actions performed",
		},
		[]string{"action", "resource", "role", "success"},
	)

	// AuditActionDuration tracks the duration of actions
	AuditActionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "audit_action_duration_seconds",
			Help:    "Duration of audit actions in seconds",
			Buckets: prometheus.DefBuckets, // Default: 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10
		},
		[]string{"action", "resource", "method"},
	)

	// AuditErrorsTotal counts the total number of failed audit actions
	AuditErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "audit_errors_total",
			Help: "Total number of failed audit actions",
		},
		[]string{"action", "resource", "error_type", "status_code"},
	)

	// AuditUserActionsTotal counts actions per user
	AuditUserActionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "audit_user_actions_total",
			Help: "Total number of actions per user",
		},
		[]string{"user_email", "action", "resource"},
	)

	// AuditCompanyActionsTotal counts actions per company (multi-tenancy)
	AuditCompanyActionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "audit_company_actions_total",
			Help: "Total number of actions per company",
		},
		[]string{"company_id", "action", "resource"},
	)

	// AuditResourceAccessTotal counts access to specific resource types
	AuditResourceAccessTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "audit_resource_access_total",
			Help: "Total number of accesses per resource type",
		},
		[]string{"resource", "action", "method"},
	)

	// AuditAuthenticationTotal tracks authentication events
	AuditAuthenticationTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "audit_authentication_total",
			Help: "Total number of authentication events",
		},
		[]string{"action", "success", "ip_address"},
	)

	// AuditSuspiciousActivityTotal tracks suspicious activities
	AuditSuspiciousActivityTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "audit_suspicious_activity_total",
			Help: "Total number of suspicious activities detected",
		},
		[]string{"activity_type", "user_email", "resource"},
	)

	// AuditDatabaseWritesTotal tracks writes to audit_logs table
	AuditDatabaseWritesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "audit_database_writes_total",
			Help: "Total number of writes to audit_logs table",
		},
	)

	// AuditDatabaseWriteErrors tracks errors writing to database
	AuditDatabaseWriteErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "audit_database_write_errors_total",
			Help: "Total number of errors writing audit logs to database",
		},
	)

	// AuditMiddlewareProcessingDuration tracks middleware overhead
	AuditMiddlewareProcessingDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "audit_middleware_processing_duration_seconds",
			Help:    "Time spent processing audit middleware",
			Buckets: []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.05, 0.1}, // Smaller buckets for middleware
		},
	)

	// AuditQueueSize tracks the size of async audit queue
	AuditQueueSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "audit_queue_size",
			Help: "Current size of async audit log queue",
		},
	)

	// AuditRequestBodySize tracks the size of request bodies captured
	AuditRequestBodySize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "audit_request_body_size_bytes",
			Help:    "Size of request bodies captured in audit logs",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000}, // Bytes
		},
		[]string{"method", "resource"},
	)

	// AuditResponseSize tracks the size of responses
	AuditResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "audit_response_size_bytes",
			Help:    "Size of HTTP responses",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000, 10000000}, // Bytes
		},
		[]string{"method", "resource", "status_code"},
	)

	// AuditHTTPStatusCodes tracks HTTP status codes
	AuditHTTPStatusCodes = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "audit_http_status_codes_total",
			Help: "Total number of HTTP status codes observed",
		},
		[]string{"method", "path", "status_code"},
	)

	// AuditSlowRequests tracks requests that take longer than threshold
	AuditSlowRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "audit_slow_requests_total",
			Help: "Total number of requests that exceeded performance threshold",
		},
		[]string{"method", "path", "threshold"},
	)
)

// IncrementAuditAction increments the audit action counter
func IncrementAuditAction(action, resource, role string, success bool) {
	successStr := "false"
	if success {
		successStr = "true"
	}
	AuditActionsTotal.WithLabelValues(action, resource, role, successStr).Inc()
}

// ObserveAuditActionDuration records the duration of an action
func ObserveAuditActionDuration(action, resource, method string, durationMs int64) {
	durationSec := float64(durationMs) / 1000.0
	AuditActionDuration.WithLabelValues(action, resource, method).Observe(durationSec)

	// Track slow requests (> 1 second)
	if durationSec > 1.0 {
		AuditSlowRequests.WithLabelValues(method, resource, "1s").Inc()
	}
	// Track very slow requests (> 5 seconds)
	if durationSec > 5.0 {
		AuditSlowRequests.WithLabelValues(method, resource, "5s").Inc()
	}
}

// IncrementAuditError increments the error counter
func IncrementAuditError(action, resource, errorType string, statusCode int) {
	statusCodeStr := fmt.Sprintf("%d", statusCode)
	AuditErrorsTotal.WithLabelValues(action, resource, errorType, statusCodeStr).Inc()
}

// IncrementUserAction increments the user action counter
func IncrementUserAction(userEmail, action, resource string) {
	if userEmail == "" {
		userEmail = "anonymous"
	}
	AuditUserActionsTotal.WithLabelValues(userEmail, action, resource).Inc()
}

// IncrementCompanyAction increments the company action counter
func IncrementCompanyAction(companyID, action, resource string) {
	if companyID == "" {
		companyID = "no_company"
	}
	AuditCompanyActionsTotal.WithLabelValues(companyID, action, resource).Inc()
}

// IncrementResourceAccess increments the resource access counter
func IncrementResourceAccess(resource, action, method string) {
	AuditResourceAccessTotal.WithLabelValues(resource, action, method).Inc()
}

// IncrementAuthenticationEvent increments the authentication counter
func IncrementAuthenticationEvent(action string, success bool, ipAddress string) {
	successStr := "false"
	if success {
		successStr = "true"
	}
	AuditAuthenticationTotal.WithLabelValues(action, successStr, ipAddress).Inc()
}

// IncrementSuspiciousActivity increments the suspicious activity counter
func IncrementSuspiciousActivity(activityType, userEmail, resource string) {
	if userEmail == "" {
		userEmail = "anonymous"
	}
	AuditSuspiciousActivityTotal.WithLabelValues(activityType, userEmail, resource).Inc()
}

// IncrementDatabaseWrite increments the database write counter
func IncrementDatabaseWrite() {
	AuditDatabaseWritesTotal.Inc()
}

// IncrementDatabaseWriteError increments the database write error counter
func IncrementDatabaseWriteError() {
	AuditDatabaseWriteErrors.Inc()
}

// ObserveMiddlewareProcessing records middleware processing time
func ObserveMiddlewareProcessing(durationMs float64) {
	durationSec := durationMs / 1000.0
	AuditMiddlewareProcessingDuration.Observe(durationSec)
}

// SetQueueSize sets the current queue size
func SetQueueSize(size int) {
	AuditQueueSize.Set(float64(size))
}

// ObserveRequestBodySize records request body size
func ObserveRequestBodySize(method, resource string, sizeBytes int) {
	AuditRequestBodySize.WithLabelValues(method, resource).Observe(float64(sizeBytes))
}

// ObserveResponseSize records response size
func ObserveResponseSize(method, resource string, statusCode, sizeBytes int) {
	statusCodeStr := fmt.Sprintf("%d", statusCode)
	AuditResponseSize.WithLabelValues(method, resource, statusCodeStr).Observe(float64(sizeBytes))
}

// IncrementHTTPStatusCode increments the HTTP status code counter
func IncrementHTTPStatusCode(method, path string, statusCode int) {
	statusCodeStr := fmt.Sprintf("%d", statusCode)
	AuditHTTPStatusCodes.WithLabelValues(method, path, statusCodeStr).Inc()
}

// DetectSuspiciousActivity analyzes audit log and detects suspicious patterns
func DetectSuspiciousActivity(action, resource string, statusCode int, userEmail string) {
	// Detect multiple failed authentications
	if action == "LOGIN" && statusCode >= 400 {
		IncrementSuspiciousActivity("failed_login", userEmail, "authentication")
	}

	// Detect DELETE operations
	if action == "DELETE" {
		IncrementSuspiciousActivity("delete_operation", userEmail, resource)
	}

	// Detect permission denied
	if statusCode == 403 {
		IncrementSuspiciousActivity("permission_denied", userEmail, resource)
	}

	// Detect rate limit triggers
	if statusCode == 429 {
		IncrementSuspiciousActivity("rate_limit_exceeded", userEmail, resource)
	}
}
