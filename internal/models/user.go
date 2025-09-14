package models

import (
	"time"

	"github.com/google/uuid"
)

// Role represents a user role in the system
type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// User represents a user in the system
type User struct {
	ID                uuid.UUID  `json:"id" db:"id"`
	Name              string     `json:"name" db:"name"`
	Email             string     `json:"email" db:"email"`
	Password          string     `json:"-" db:"password"` // Never include password in JSON responses
	Phone             *string    `json:"phone" db:"phone"`
	CPF               *string    `json:"cpf" db:"cpf"`
	Avatar            *string    `json:"avatar" db:"avatar"`
	RoleID            uuid.UUID  `json:"role_id" db:"role_id"`
	Role              *Role      `json:"role,omitempty"` // For joined queries
	Active            bool       `json:"active" db:"active"`
	LastLogin         *time.Time `json:"last_login" db:"last_login"`
	DashboardConfig   *string    `json:"dashboard_config" db:"dashboard_config"` // JSON stored as string
	APIToken          *string    `json:"api_token,omitempty" db:"api_token"`
	LoginAttempts     int        `json:"login_attempts" db:"login_attempts"`
	BlockedUntil      *time.Time `json:"blocked_until" db:"blocked_until"`
	PasswordChangedAt time.Time  `json:"password_changed_at" db:"password_changed_at"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// UserSession represents a user session
type UserSession struct {
	ID          string    `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	IPAddress   *string   `json:"ip_address" db:"ip_address"`
	UserAgent   *string   `json:"user_agent" db:"user_agent"`
	SessionData *string   `json:"session_data" db:"session_data"` // JSON stored as string
	ExpiresAt   time.Time `json:"expires_at" db:"expires_at"`
	Active      bool      `json:"active" db:"active"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// AuthLog represents an authentication attempt log
type AuthLog struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	UserID        *uuid.UUID `json:"user_id" db:"user_id"`
	EmailAttempt  string     `json:"email_attempt" db:"email_attempt"`
	Success       bool       `json:"success" db:"success"`
	IPAddress     *string    `json:"ip_address" db:"ip_address"`
	UserAgent     *string    `json:"user_agent" db:"user_agent"`
	FailureReason *string    `json:"failure_reason" db:"failure_reason"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=8,max=255"`
	Phone    string `json:"phone,omitempty" binding:"omitempty,max=20"`
	CPF      string `json:"cpf,omitempty" binding:"omitempty,len=14"`
	RoleID   string `json:"role_id" binding:"required,uuid"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Name            string `json:"name,omitempty" binding:"omitempty,min=2,max=100"`
	Email           string `json:"email,omitempty" binding:"omitempty,email,max=100"`
	Phone           string `json:"phone,omitempty" binding:"omitempty,max=20"`
	CPF             string `json:"cpf,omitempty" binding:"omitempty,len=14"`
	Avatar          string `json:"avatar,omitempty" binding:"omitempty,max=255"`
	Active          *bool  `json:"active,omitempty"`
	DashboardConfig string `json:"dashboard_config,omitempty"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	User        User   `json:"user"`
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}
