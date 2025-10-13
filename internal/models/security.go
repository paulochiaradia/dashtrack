package models

import (
	"time"

	"github.com/google/uuid"
)

// RateLimitRule represents a rate limiting rule
type RateLimitRule struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Path        string    `json:"path" db:"path"`     // API path pattern
	Method      string    `json:"method" db:"method"` // HTTP method
	MaxRequests int       `json:"max_requests" db:"max_requests"`
	WindowSize  string    `json:"window_size" db:"window_size"` // Time window as string (interval)
	UserBased   bool      `json:"user_based" db:"user_based"`   // Per user or per IP
	Active      bool      `json:"active" db:"active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// RateLimitEvent represents a rate limit event
type RateLimitEvent struct {
	ID        uuid.UUID  `json:"id" db:"id"`
	UserID    *uuid.UUID `json:"user_id" db:"user_id"` // Null for IP-based
	IPAddress string     `json:"ip_address" db:"ip_address"`
	Path      string     `json:"path" db:"path"`
	Method    string     `json:"method" db:"method"`
	UserAgent string     `json:"user_agent" db:"user_agent"`
	Blocked   bool       `json:"blocked" db:"blocked"`
	RuleID    *uuid.UUID `json:"rule_id" db:"rule_id"`
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
}

// SessionToken represents a user session with refresh token
type SessionToken struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	UserID           uuid.UUID  `json:"user_id" db:"user_id"`
	AccessToken      string     `json:"-" db:"access_token_hash"`  // Hash of access token
	RefreshToken     string     `json:"-" db:"refresh_token_hash"` // Hash of refresh token
	IPAddress        string     `json:"ip_address" db:"ip_address"`
	UserAgent        string     `json:"user_agent" db:"user_agent"`
	ExpiresAt        time.Time  `json:"expires_at" db:"expires_at"`
	RefreshExpiresAt time.Time  `json:"refresh_expires_at" db:"refresh_expires_at"`
	Revoked          bool       `json:"revoked" db:"revoked"`
	RevokedAt        *time.Time `json:"revoked_at" db:"revoked_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// AuditLog represents a comprehensive audit log entry for system actions
type AuditLog struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	UserID       *uuid.UUID             `json:"user_id" db:"user_id"`
	UserEmail    *string                `json:"user_email" db:"user_email"`
	CompanyID    *uuid.UUID             `json:"company_id" db:"company_id"`
	
	// Action details
	Action       string                 `json:"action" db:"action"`           // CREATE, UPDATE, DELETE, READ, LOGIN, LOGOUT
	Resource     string                 `json:"resource" db:"resource"`       // user, vehicle, team, company, etc
	ResourceID   *string                `json:"resource_id" db:"resource_id"` // UUID as string
	
	// Request context
	Method       *string                `json:"method" db:"method"`         // GET, POST, PUT, DELETE, PATCH
	Path         *string                `json:"path" db:"path"`             // Request path
	IPAddress    string                 `json:"ip_address" db:"ip_address"`
	UserAgent    string                 `json:"user_agent" db:"user_agent"`
	
	// Data changes (for audit trail)
	Details      map[string]interface{} `json:"details" db:"details"`   // Legacy field (kept for compatibility)
	Changes      map[string]interface{} `json:"changes" db:"changes"`   // Before/after state for updates
	Metadata     map[string]interface{} `json:"metadata" db:"metadata"` // Additional context
	
	// Result
	Success      bool                   `json:"success" db:"success"`
	ErrorMessage *string                `json:"error_message" db:"error_message"`
	StatusCode   *int                   `json:"status_code" db:"status_code"`
	DurationMs   *int64                 `json:"duration_ms" db:"duration_ms"` // Operation duration in milliseconds
	
	// Distributed tracing
	TraceID      *string                `json:"trace_id" db:"trace_id"` // Jaeger trace ID
	SpanID       *string                `json:"span_id" db:"span_id"`   // Jaeger span ID
	
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// AuditLogFilter represents filters for querying audit logs
type AuditLogFilter struct {
	UserID     *uuid.UUID `json:"user_id"`
	CompanyID  *uuid.UUID `json:"company_id"`
	Action     *string    `json:"action"`
	Resource   *string    `json:"resource"`
	ResourceID *string    `json:"resource_id"`
	Success    *bool      `json:"success"`
	From       *time.Time `json:"from"`
	To         *time.Time `json:"to"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}

// AuditLogStats represents aggregated audit log statistics
type AuditLogStats struct {
	TotalActions      int64                      `json:"total_actions"`
	SuccessRate       float64                    `json:"success_rate"`
	ActionsByType     map[string]int64           `json:"actions_by_type"`
	ActionsByResource map[string]int64           `json:"actions_by_resource"`
	TopUsers          []UserActionCount          `json:"top_users"`
	RecentFailures    []AuditLog                 `json:"recent_failures"`
	AvgDurationMs     float64                    `json:"avg_duration_ms"`
}

// UserActionCount represents action count per user
type UserActionCount struct {
	UserID    uuid.UUID `json:"user_id"`
	UserEmail string    `json:"user_email"`
	Count     int64     `json:"count"`
}
