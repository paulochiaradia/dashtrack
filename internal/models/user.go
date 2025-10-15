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
	ID        uuid.UUID  `json:"id" db:"id"`
	Name      string     `json:"name" db:"name"`
	Email     string     `json:"email" db:"email"`
	Password  string     `json:"-" db:"password"` // Never include password in JSON responses
	Phone     *string    `json:"phone" db:"phone"`
	CPF       *string    `json:"cpf" db:"cpf"`
	Avatar    *string    `json:"avatar" db:"avatar"`
	RoleID    uuid.UUID  `json:"role_id" db:"role_id"`
	CompanyID *uuid.UUID `json:"company_id" db:"company_id"` // Multi-tenant support
	Role      *Role      `json:"role,omitempty"`             // For joined queries
	// Company will be populated separately to avoid circular dependency
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

// RecentLogin represents recent login activity for dashboard
type RecentLogin struct {
	UserID      uuid.UUID  `json:"user_id"`
	UserName    string     `json:"user_name"`
	UserEmail   string     `json:"user_email"`
	Success     bool       `json:"success"`
	IPAddress   *string    `json:"ip_address"`
	UserAgent   *string    `json:"user_agent"`
	LoginTime   time.Time  `json:"login_time"`
	CompanyID   *uuid.UUID `json:"company_id,omitempty"`
	CompanyName *string    `json:"company_name,omitempty"`
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	Name      string  `json:"name" binding:"required,min=2,max=100"`
	Email     string  `json:"email" binding:"required,email,max=100"`
	Password  string  `json:"password" binding:"required,min=8,max=255"`
	Phone     string  `json:"phone" binding:"required,min=10,max=20"` // Obrigatório: telefone
	CPF       string  `json:"cpf" binding:"required,len=14"`          // Obrigatório: CPF no formato XXX.XXX.XXX-XX
	RoleID    string  `json:"role_id" binding:"required,uuid"`
	CompanyID *string `json:"company_id,omitempty" binding:"omitempty,uuid"` // For company users
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
	RoleID          string `json:"role_id,omitempty" binding:"omitempty,uuid"`
}

// TransferUserRequest represents the request to transfer a user to another company (Master only)
type TransferUserRequest struct {
	CompanyID string `json:"company_id" binding:"required,uuid"`
	Reason    string `json:"reason,omitempty" binding:"omitempty,max=255"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	User         User   `json:"user"`
}

// RefreshTokenRequest represents a refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// TwoFactorAuth represents 2FA settings for a user
type TwoFactorAuth struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	Secret      string     `json:"-" db:"secret"`       // TOTP secret, never expose
	BackupCodes []string   `json:"-" db:"backup_codes"` // JSON array of backup codes
	Enabled     bool       `json:"enabled" db:"enabled"`
	LastUsed    *time.Time `json:"last_used" db:"last_used"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}

// TwoFactorSetupRequest represents 2FA setup request
type TwoFactorSetupRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

// TwoFactorVerifyRequest represents 2FA verification request
type TwoFactorVerifyRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

// UserContext represents the context of a user for permissions and multi-tenancy
type UserContext struct {
	UserID    uuid.UUID  `json:"user_id"`
	CompanyID *uuid.UUID `json:"company_id,omitempty"`
	Role      string     `json:"role"`
	IsMaster  bool       `json:"is_master"`
}

// HasCompanyAccess checks if user has access to a specific company
func (uc *UserContext) HasCompanyAccess(companyID uuid.UUID) bool {
	if uc.IsMaster {
		return true
	}
	return uc.CompanyID != nil && *uc.CompanyID == companyID
}

// CanManageCompany checks if user can manage company settings
func (uc *UserContext) CanManageCompany() bool {
	return uc.IsMaster || uc.Role == "company_admin"
}

// CanManageUsers checks if user can manage other users
func (uc *UserContext) CanManageUsers() bool {
	return uc.IsMaster || uc.Role == "company_admin"
}

// CanManageVehicles checks if user can manage vehicles
func (uc *UserContext) CanManageVehicles() bool {
	return uc.IsMaster || uc.Role == "company_admin"
}

// CanViewAllData checks if user can view all company data
func (uc *UserContext) CanViewAllData() bool {
	return uc.IsMaster || uc.Role == "company_admin"
}

// CanAccessVehicle checks if user can access a specific vehicle
// Master and company_admin have access to all vehicles in company
// Drivers and helpers only have access to their assigned vehicles
func (uc *UserContext) CanAccessVehicle(vehicleID uuid.UUID) bool {
	if uc.IsMaster || uc.Role == "company_admin" {
		return true
	}
	// For drivers and helpers, access is checked at repository level
	return uc.Role == "driver" || uc.Role == "helper"
}

// IsDriver checks if user is a driver
func (uc *UserContext) IsDriver() bool {
	return uc.Role == "driver"
}

// IsHelper checks if user is a helper
func (uc *UserContext) IsHelper() bool {
	return uc.Role == "helper"
}

// CanManageTeam checks if user can manage team operations
func (uc *UserContext) CanManageTeam() bool {
	return uc.IsMaster || uc.Role == "company_admin"
}

// AssignTeamMemberRequest represents request to assign user to team
type AssignTeamMemberRequest struct {
	UserID     uuid.UUID `json:"user_id" binding:"required"`
	RoleInTeam string    `json:"role_in_team" binding:"required,oneof=manager driver assistant supervisor"`
}

// CreateCompanyUserRequest represents request to create a company user
type CreateCompanyUserRequest struct {
	Name     string `json:"name" binding:"required,min=2,max=100"`
	Email    string `json:"email" binding:"required,email,max=100"`
	Password string `json:"password" binding:"required,min=8,max=255"`
	Phone    string `json:"phone,omitempty" binding:"omitempty,max=20"`
	CPF      string `json:"cpf,omitempty" binding:"omitempty,len=14"`
	Role     string `json:"role" binding:"required,oneof=driver helper supervisor company_admin"`
}
