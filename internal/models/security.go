package models

import (
	"time"

	"github.com/google/uuid"
)

// RateLimitRule represents a rate limiting rule
type RateLimitRule struct {
	ID          uuid.UUID     `json:"id" db:"id"`
	Name        string        `json:"name" db:"name"`
	Path        string        `json:"path" db:"path"`     // API path pattern
	Method      string        `json:"method" db:"method"` // HTTP method
	MaxRequests int           `json:"max_requests" db:"max_requests"`
	WindowSize  time.Duration `json:"window_size" db:"window_size"` // Time window
	UserBased   bool          `json:"user_based" db:"user_based"`   // Per user or per IP
	Active      bool          `json:"active" db:"active"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
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

// AuditLog represents an audit log entry (extends existing AuthLog)
type AuditLog struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	UserID       *uuid.UUID             `json:"user_id" db:"user_id"`
	Action       string                 `json:"action" db:"action"`           // LOGIN, LOGOUT, CREATE_USER, etc.
	Resource     string                 `json:"resource" db:"resource"`       // users, vehicles, companies
	ResourceID   *string                `json:"resource_id" db:"resource_id"` // ID of affected resource
	IPAddress    string                 `json:"ip_address" db:"ip_address"`
	UserAgent    string                 `json:"user_agent" db:"user_agent"`
	Details      map[string]interface{} `json:"details" db:"details"` // JSON for additional info
	Success      bool                   `json:"success" db:"success"`
	ErrorMessage *string                `json:"error_message" db:"error_message"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}
