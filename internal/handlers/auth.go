package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/paulochiaradia/dashtrack/internal/logger"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
	"github.com/paulochiaradia/dashtrack/internal/services"
	"github.com/paulochiaradia/dashtrack/internal/utils"
	"go.uber.org/zap"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userRepo     repository.UserRepositoryInterface
	authLogRepo  repository.AuthLogRepositoryInterface
	roleRepo     repository.RoleRepositoryInterface
	tokenService *services.TokenService
	emailService *services.EmailService
	bcryptCost   int
}

// LoginRequest represents login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginResponse represents login response payload
type LoginResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	ExpiresIn    int64        `json:"expires_in"` // seconds until access token expires
}

// RefreshTokenRequest represents refresh token request payload
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents refresh token response payload
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// ChangePasswordRequest represents change password request payload
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}

// UserResponse represents user data in responses (no sensitive info)
type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone,omitempty"`
	Role      string    `json:"role"`
	Active    bool      `json:"active"`
	Avatar    string    `json:"avatar,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserHistoryResponse represents complete user activity history
type UserHistoryResponse struct {
	UserID     string             `json:"user_id"`
	Summary    UserHistorySummary `json:"summary"`
	Activities []UserActivityItem `json:"activities"`
}

// UserHistorySummary represents aggregated statistics
type UserHistorySummary struct {
	TotalLogins           int        `json:"total_logins"`
	SuccessfulLogins      int        `json:"successful_logins"`
	FailedLogins          int        `json:"failed_logins"`
	TotalLogouts          int        `json:"total_logouts"`
	AverageSessionMinutes float64    `json:"average_session_minutes"`
	PasswordChanges       int        `json:"password_changes"`
	UniqueIPs             []string   `json:"unique_ips"`
	LastLoginAt           *time.Time `json:"last_login_at"`
	LastPasswordChangeAt  *time.Time `json:"last_password_change_at"`
}

// UserActivityItem represents a single activity event
type UserActivityItem struct {
	Timestamp time.Time              `json:"timestamp"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Success   bool                   `json:"success"`
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo repository.UserRepositoryInterface, authLogRepo repository.AuthLogRepositoryInterface, roleRepo repository.RoleRepositoryInterface, tokenService *services.TokenService, emailService *services.EmailService, bcryptCost int) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		authLogRepo:  authLogRepo,
		roleRepo:     roleRepo,
		tokenService: tokenService,
		emailService: emailService,
		bcryptCost:   bcryptCost,
	}
}

// Helper function to get string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ============================================================================
// GIN FRAMEWORK HANDLERS
// ============================================================================

// LoginGin handles login requests using Gin framework
func (h *AuthHandler) LoginGin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Get user by email
	user, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)

	// If user not found, log failed attempt and return
	if err != nil {
		if err == sql.ErrNoRows {
			// Log failed attempt for non-existent user
			_ = h.logAuthAttempt(nil, req.Email, false, clientIP, userAgent, "User not found")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		_ = h.logAuthAttempt(nil, req.Email, false, clientIP, userAgent, "Database error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if user is blocked (after 3 failed attempts)
	if user.BlockedUntil != nil && user.BlockedUntil.After(time.Now()) {
		remainingTime := time.Until(*user.BlockedUntil)
		_ = h.logAuthAttempt(&user.ID, req.Email, false, clientIP, userAgent, "Account temporarily blocked")
		c.JSON(http.StatusForbidden, gin.H{
			"error":            "Account temporarily blocked due to multiple failed login attempts",
			"blocked_until":    user.BlockedUntil.Format(time.RFC3339),
			"retry_in_seconds": int(remainingTime.Seconds()),
		})
		return
	}

	// Check if user is active
	if !user.Active {
		_ = h.logAuthAttempt(&user.ID, req.Email, false, clientIP, userAgent, "Account is inactive")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is inactive"})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		// Password incorrect - increment login attempts
		newAttempts := user.LoginAttempts + 1

		// Block user if 3 or more failed attempts
		var blockedUntil *time.Time
		var failureReason string

		if newAttempts >= 3 {
			blockTime := time.Now().Add(15 * time.Minute) // Block for 15 minutes
			blockedUntil = &blockTime
			failureReason = fmt.Sprintf("Account blocked after %d failed attempts", newAttempts)

			// Send password reset email asynchronously
			go h.sendBlockedAccountEmail(user.Email, user.Name, blockTime)
		} else {
			failureReason = fmt.Sprintf("Invalid password (attempt %d/3)", newAttempts)
		}

		// Update login attempts and blocked_until
		_ = h.userRepo.UpdateLoginAttempts(c.Request.Context(), user.ID, newAttempts, blockedUntil)

		// Log failed attempt
		_ = h.logAuthAttempt(&user.ID, req.Email, false, clientIP, userAgent, failureReason)

		if blockedUntil != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Account temporarily blocked due to multiple failed login attempts. Check your email for password reset instructions.",
				"blocked_until": blockedUntil.Format(time.RFC3339),
			})
			return
		}

		c.JSON(http.StatusUnauthorized, gin.H{
			"error":              "Invalid credentials",
			"attempts_remaining": 3 - newAttempts,
		})
		return
	}

	// Password correct - Reset login attempts if any
	if user.LoginAttempts > 0 || user.BlockedUntil != nil {
		_ = h.userRepo.UpdateLoginAttempts(c.Request.Context(), user.ID, 0, nil)
	}

	// Update last login
	_ = h.userRepo.UpdateLastLogin(c.Request.Context(), user.ID)

	// Generate token pair using tokenService (with session management)
	tokenPair, err := h.tokenService.GenerateTokenPair(c.Request.Context(), user, clientIP, userAgent)
	if err != nil {
		_ = h.logAuthAttempt(&user.ID, req.Email, false, clientIP, userAgent, "Failed to generate tokens")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Log successful login
	_ = h.logAuthAttempt(&user.ID, req.Email, true, clientIP, userAgent, "")

	response := LoginResponse{
		User: UserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			Role:      user.Role.Name,
			Active:    user.Active,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int64(tokenPair.ExpiresIn),
	}

	c.JSON(http.StatusOK, response)
}

// RefreshTokenGin handles refresh token requests using Gin framework
func (h *AuthHandler) RefreshTokenGin(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Get client info
	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Refresh token pair using tokenService
	tokenPair, err := h.tokenService.RefreshTokenPair(c.Request.Context(), req.RefreshToken, clientIP, userAgent)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	response := RefreshTokenResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    int64(tokenPair.ExpiresIn),
	}

	c.JSON(http.StatusOK, response)
}

// LogoutGin handles logout requests using Gin framework
func (h *AuthHandler) LogoutGin(c *gin.Context) {
	// Get user context from middleware
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get session_id from context (set by auth middleware)
	sessionIDStr, exists := c.Get("session_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session not found"})
		return
	}

	sessionID, err := uuid.Parse(sessionIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	// Get user email for audit log
	userEmail, _ := c.Get("email")
	emailStr := ""
	if userEmail != nil {
		emailStr = userEmail.(string)
	}

	// Get IP and User-Agent
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Get session start time to calculate duration
	var sessionStart time.Time
	var sessionDurationMinutes float64

	err = h.tokenService.GetDB().GetContext(c.Request.Context(), &sessionStart,
		"SELECT created_at FROM user_sessions WHERE id = $1", sessionID.String())

	if err == nil {
		sessionDurationMinutes = time.Since(sessionStart).Minutes()
	}

	// Mark session as inactive in both tables
	tx, err := h.tokenService.GetDB().BeginTxx(c.Request.Context(), nil)
	if err != nil {
		logger.Error("Failed to begin transaction for logout", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}
	defer tx.Rollback()

	// Update session_tokens (revoke)
	_, err = tx.ExecContext(c.Request.Context(),
		"UPDATE session_tokens SET revoked = true, revoked_at = NOW(), updated_at = NOW() WHERE id = $1",
		sessionID)
	if err != nil {
		logger.Error("Failed to revoke session in session_tokens", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	// Update user_sessions (mark inactive)
	_, err = tx.ExecContext(c.Request.Context(),
		"UPDATE user_sessions SET active = false WHERE id = $1",
		sessionID.String())
	if err != nil {
		logger.Error("Failed to mark session inactive in user_sessions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	// Create audit log entry
	auditLog := map[string]interface{}{
		"user_id":     userID,
		"user_email":  emailStr,
		"action":      "logout",
		"resource":    "session",
		"resource_id": sessionID,
		"method":      "POST",
		"path":        c.Request.URL.Path,
		"ip_address":  clientIP,
		"user_agent":  userAgent,
		"success":     true,
		"status_code": 200,
		"metadata": map[string]interface{}{
			"session_id":               sessionID.String(),
			"session_duration_minutes": sessionDurationMinutes,
			"logout_time":              utils.Now(),
		},
	}

	metadataJSON, _ := json.Marshal(auditLog["metadata"])

	_, err = tx.ExecContext(c.Request.Context(), `
		INSERT INTO audit_logs (
			user_id, user_email, action, resource, resource_id, method, path,
			ip_address, user_agent, metadata, success, status_code, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`, userID, emailStr, "logout", "session", sessionID, "POST", c.Request.URL.Path,
		clientIP, userAgent, metadataJSON, true, 200, time.Now())

	if err != nil {
		logger.Error("Failed to create audit log for logout", zap.Error(err))
		// Don't fail logout if audit log fails
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		logger.Error("Failed to commit logout transaction", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	logger.Info("User logged out successfully",
		zap.String("user_id", userID.String()),
		zap.String("session_id", sessionID.String()),
		zap.Float64("session_duration_minutes", sessionDurationMinutes))

	c.JSON(http.StatusOK, gin.H{
		"message":                  "Logout successful",
		"session_duration_minutes": sessionDurationMinutes,
	})
}

// ChangePasswordGin handles password change requests using Gin framework
func (h *AuthHandler) ChangePasswordGin(c *gin.Context) {
	// Get user context from middleware
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Get user
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	// Verify current password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password
	err = h.userRepo.UpdatePassword(c.Request.Context(), userID, string(hashedPassword))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Create audit log for password change
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	metadata := map[string]interface{}{
		"change_method": "manual",
		"ip_address":    clientIP,
		"user_agent":    userAgent,
		"changed_at":    utils.Now().Format(time.RFC3339),
	}

	resourceIDStr := userID.String()
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		UserID:     &userID,
		Action:     "password_change",
		Resource:   "user",
		ResourceID: &resourceIDStr,
		IPAddress:  clientIP,
		UserAgent:  userAgent,
		Metadata:   metadata,
		Success:    true,
		CreatedAt:  utils.Now(),
	}

	// Store audit log in database
	query := `
		INSERT INTO audit_logs (
			id, user_id, action, resource, resource_id,
			ip_address, user_agent, metadata, success, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	metadataJSON, _ := json.Marshal(metadata)
	_, err = h.tokenService.GetDB().ExecContext(c.Request.Context(), query,
		auditLog.ID, auditLog.UserID, auditLog.Action, auditLog.Resource,
		auditLog.ResourceID, auditLog.IPAddress, auditLog.UserAgent,
		metadataJSON, auditLog.Success, auditLog.CreatedAt,
	)

	if err != nil {
		logger.Error("Failed to create audit log for password change",
			zap.Error(err),
			zap.String("user_id", userID.String()))
		// Don't fail the request if audit log fails
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// MeGin returns current user information using Gin framework
func (h *AuthHandler) MeGin(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get fresh user data
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	response := UserResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		Phone:     getStringValue(user.Phone),
		Role:      user.Role.Name,
		Active:    user.Active,
		Avatar:    getStringValue(user.Avatar),
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// GetRolesGin returns available roles using Gin framework
func (h *AuthHandler) GetRolesGin(c *gin.Context) {
	roles, err := h.roleRepo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve roles"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// GetUserHistoryGin returns complete user activity history
func (h *AuthHandler) GetUserHistoryGin(c *gin.Context) {
	// Get target user ID from URL parameter
	targetUserIDStr := c.Param("id")
	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get current user context for authorization
	currentUserIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	currentUserID, err := uuid.Parse(currentUserIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid current user ID"})
		return
	}

	// Authorization: user can only view their own history unless they're admin/master
	role, _ := c.Get("role_name")
	roleStr := ""
	if role != nil {
		roleStr = role.(string)
	}

	if currentUserID != targetUserID && roleStr != "admin" && roleStr != "master" {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only view your own history"})
		return
	}

	// Aggregate data from multiple sources
	db := h.tokenService.GetDB()

	// 1. Get login statistics from auth_logs
	var totalLogins, successfulLogins, failedLogins int
	var lastLoginAt sql.NullTime

	err = db.QueryRowContext(c.Request.Context(), `
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE success = true) as successful,
			COUNT(*) FILTER (WHERE success = false) as failed,
			MAX(created_at) FILTER (WHERE success = true) as last_login
		FROM auth_logs
		WHERE user_id = $1
	`, targetUserID).Scan(&totalLogins, &successfulLogins, &failedLogins, &lastLoginAt)

	if err != nil && err != sql.ErrNoRows {
		logger.Error("Failed to get login statistics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve history"})
		return
	}

	// 2. Get logout count and average session duration from audit_logs
	var totalLogouts int
	var avgSessionMinutes sql.NullFloat64

	err = db.QueryRowContext(c.Request.Context(), `
		SELECT 
			COUNT(*) as total_logouts,
			AVG((metadata->>'session_duration_minutes')::float) as avg_duration
		FROM audit_logs
		WHERE user_id = $1 AND action = 'logout'
	`, targetUserID).Scan(&totalLogouts, &avgSessionMinutes)

	if err != nil && err != sql.ErrNoRows {
		logger.Error("Failed to get logout statistics", zap.Error(err))
	}

	// 3. Get password changes from audit_logs
	var passwordChanges int
	var lastPasswordChangeAt sql.NullTime

	err = db.QueryRowContext(c.Request.Context(), `
		SELECT 
			COUNT(*) as total,
			MAX(created_at) as last_change
		FROM audit_logs
		WHERE user_id = $1 AND action = 'password_change'
	`, targetUserID).Scan(&passwordChanges, &lastPasswordChangeAt)

	if err != nil && err != sql.ErrNoRows {
		logger.Error("Failed to get password change statistics", zap.Error(err))
	}

	// 4. Get unique IPs from both auth_logs and audit_logs
	rows, err := db.QueryContext(c.Request.Context(), `
		SELECT DISTINCT ip_address FROM (
			SELECT ip_address FROM auth_logs WHERE user_id = $1 AND ip_address IS NOT NULL
			UNION
			SELECT ip_address FROM audit_logs WHERE user_id = $1 AND ip_address IS NOT NULL
		) AS ips
		ORDER BY ip_address
	`, targetUserID)

	uniqueIPs := []string{}
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var ip sql.NullString
			if err := rows.Scan(&ip); err == nil && ip.Valid {
				uniqueIPs = append(uniqueIPs, ip.String)
			}
		}
	}

	// 5. Get activity timeline (combine auth_logs and audit_logs)
	activityRows, err := db.QueryContext(c.Request.Context(), `
		SELECT 
			'login' as action,
			CASE WHEN success THEN 'auth' ELSE 'auth' END as resource,
			success,
			COALESCE(ip_address, '') as ip,
			COALESCE(user_agent, '') as user_agent,
			NULL::jsonb as details,
			created_at as timestamp
		FROM auth_logs
		WHERE user_id = $1
		
		UNION ALL
		
		SELECT 
			action,
			resource,
			COALESCE(success, true) as success,
			ip_address as ip,
			user_agent,
			metadata as details,
			created_at as timestamp
		FROM audit_logs
		WHERE user_id = $1
		
		ORDER BY timestamp DESC
		LIMIT 100
	`, targetUserID)

	activities := []UserActivityItem{}
	if err == nil {
		defer activityRows.Close()
		for activityRows.Next() {
			var item UserActivityItem
			var detailsJSON []byte
			var ipStr, uaStr sql.NullString

			err := activityRows.Scan(
				&item.Action,
				&item.Resource,
				&item.Success,
				&ipStr,
				&uaStr,
				&detailsJSON,
				&item.Timestamp,
			)

			if err == nil {
				item.IPAddress = ipStr.String
				item.UserAgent = uaStr.String

				if len(detailsJSON) > 0 {
					json.Unmarshal(detailsJSON, &item.Details)
				}

				activities = append(activities, item)
			}
		}
	}

	// Build response
	summary := UserHistorySummary{
		TotalLogins:           totalLogins,
		SuccessfulLogins:      successfulLogins,
		FailedLogins:          failedLogins,
		TotalLogouts:          totalLogouts,
		AverageSessionMinutes: 0,
		PasswordChanges:       passwordChanges,
		UniqueIPs:             uniqueIPs,
		LastLoginAt:           nil,
		LastPasswordChangeAt:  nil,
	}

	if avgSessionMinutes.Valid {
		summary.AverageSessionMinutes = avgSessionMinutes.Float64
	}

	if lastLoginAt.Valid {
		summary.LastLoginAt = &lastLoginAt.Time
	}

	if lastPasswordChangeAt.Valid {
		summary.LastPasswordChangeAt = &lastPasswordChangeAt.Time
	}

	response := UserHistoryResponse{
		UserID:     targetUserID.String(),
		Summary:    summary,
		Activities: activities,
	}

	c.JSON(http.StatusOK, response)
}

// ForgotPasswordRequest represents forgot password request payload
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ResetPasswordRequest represents reset password request payload
type ResetPasswordRequest struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// ForgotPasswordGin handles forgot password requests using Gin framework
func (h *AuthHandler) ForgotPasswordGin(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Check if user exists
	user, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		// For security, always return success even if user doesn't exist
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a password reset link will be sent"})
		return
	}

	if !user.Active {
		// Don't reveal that account is inactive
		c.JSON(http.StatusOK, gin.H{"message": "If the email exists, a password reset link will be sent"})
		return
	}

	// TODO: Generate password reset token and send email
	// For now, we'll return a placeholder response
	// In production, you would:
	// 1. Generate a secure reset token with expiration
	// 2. Store it in database (password_reset_tokens table)
	// 3. Send email with reset link containing the token

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a password reset link will be sent",
		// TODO: Remove this in production
		"note": "Email sending not yet implemented. Password reset token would be sent to: " + req.Email,
	})
}

// ResetPasswordGin handles password reset using Gin framework
func (h *AuthHandler) ResetPasswordGin(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// TODO: Implement password reset functionality
	// In production, you would:
	// 1. Validate the reset token from database
	// 2. Check if token is not expired
	// 3. Get user ID from token
	// 4. Hash new password
	// 5. Update user password
	// 6. Invalidate/delete the reset token
	// 7. Optionally: Invalidate all existing sessions

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset functionality is not fully implemented yet",
		"note":    "Requires password_reset_tokens table and email service",
	})
}

// ============================================================================
// HELPER METHODS
// ============================================================================

// logAuthAttempt logs authentication attempts to auth_logs table
func (h *AuthHandler) logAuthAttempt(userID *uuid.UUID, email string, success bool, ipAddress, userAgent, failureReason string) error {
	authLog := &models.AuthLog{
		ID:           uuid.New(),
		UserID:       userID,
		EmailAttempt: email,
		Success:      success,
		IPAddress:    &ipAddress,
		UserAgent:    &userAgent,
	}

	if !success && failureReason != "" {
		authLog.FailureReason = &failureReason
	}

	err := h.authLogRepo.Create(authLog)
	if err != nil {
		logger.Error("Failed to log auth attempt",
			zap.Error(err),
			zap.String("email", email),
			zap.Bool("success", success))
	}

	return err
}

// sendBlockedAccountEmail sends an email to user when account is blocked
func (h *AuthHandler) sendBlockedAccountEmail(email, name string, blockedUntil time.Time) {
	if h.emailService == nil {
		logger.Warn("Email service not available, skipping blocked account email",
			zap.String("email", email))
		return
	}

	subject := "Conta Temporariamente Bloqueada - DashTrack"

	// Formatar data em português (timezone de Brasília)
	blockedDate := utils.FormatBrasiliaDefault(blockedUntil)

	// Calcular minutos restantes
	minutesRemaining := int(time.Until(blockedUntil).Minutes())

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #f44336; color: white; padding: 20px; text-align: center; }
        .content { background-color: #f9f9f9; padding: 30px; border-radius: 5px; margin-top: 20px; }
        .alert { background-color: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 15px 0; }
        .info-box { background-color: #fff; border: 2px solid #f44336; padding: 15px; text-align: center; margin: 20px 0; border-radius: 5px; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #777; }
        .button { display: inline-block; background-color: #4CAF50; color: white; padding: 12px 24px; text-decoration: none; border-radius: 5px; margin: 15px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔒 Conta Temporariamente Bloqueada</h1>
        </div>
        <div class="content">
            <p>Olá <strong>%s</strong>,</p>
            <p>Sua conta DashTrack foi temporariamente bloqueada devido a <strong>3 tentativas consecutivas de login com senha incorreta</strong>.</p>
            
            <div class="info-box">
                <h3 style="margin: 0; color: #f44336;">⏰ Bloqueio Expira Em:</h3>
                <p style="font-size: 18px; font-weight: bold; margin: 10px 0;">%s</p>
                <p style="color: #666; margin: 5px 0;">Aproximadamente %d minutos</p>
            </div>
            
            <div class="alert">
                <strong>🔐 Recomendação de Segurança:</strong>
                <p style="margin: 10px 0;">
                    Por segurança, recomendamos fortemente que você redefina sua senha.
                </p>
            </div>
            
            <div style="background-color: #e3f2fd; border-left: 4px solid #2196F3; padding: 15px; margin: 20px 0;">
                <h4 style="margin: 0 0 10px 0; color: #1976d2;">📋 Como Redefinir Sua Senha:</h4>
                <ol style="margin: 10px 0; padding-left: 20px;">
                    <li style="margin: 8px 0;"><strong>Acesse a plataforma DashTrack</strong></li>
                    <li style="margin: 8px 0;">Na tela de login, clique em <strong>"Esqueci minha senha"</strong></li>
                    <li style="margin: 8px 0;">Digite seu email e receba um <strong>código de verificação</strong></li>
                    <li style="margin: 8px 0;">Use o código para <strong>criar uma nova senha segura</strong></li>
                </ol>
                <p style="margin: 10px 0 0 0; font-size: 14px; color: #666;">
                    💡 <em>Após redefinir a senha, você poderá fazer login normalmente.</em>
                </p>
            </div>
            
            <div class="alert" style="background-color: #f8d7da; border-left-color: #dc3545; margin-top: 20px;">
                <strong>⚠️ Atenção:</strong>
                <p style="margin: 10px 0;">
                    Se você <strong>não reconhece</strong> estas tentativas de login, sua conta pode estar sob ataque. 
                    Entre em contato com o suporte imediatamente.
                </p>
            </div>
            
            <p style="margin-top: 20px; font-size: 14px; color: #666;">
                <strong>Dica:</strong> Após o desbloqueio, você terá novamente 3 tentativas. 
                Use senhas fortes e únicas para cada serviço.
            </p>
        </div>
        <div class="footer">
            <p>DashTrack - Sistema de Gestão de Entregas</p>
            <p>Este é um email automático, não responda.</p>
        </div>
    </div>
</body>
</html>
`, name, blockedDate, minutesRemaining)

	err := h.emailService.SendEmail(services.EmailData{
		To:      email,
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})

	if err != nil {
		logger.Error("Failed to send blocked account email",
			zap.Error(err),
			zap.String("email", email))
	} else {
		logger.Info("Blocked account email sent",
			zap.String("email", email),
			zap.Time("blocked_until", blockedUntil))
	}
}

// sendNewSessionAlert sends an email when a new session is created and old ones are revoked
func (h *AuthHandler) sendNewSessionAlert(email, name, newIP, newUserAgent string, revokedCount int) {
	if h.emailService == nil {
		logger.Warn("Email service not available, skipping new session alert",
			zap.String("email", email))
		return
	}

	subject := "Nova Sessão Detectada - DashTrack"
	loginTime := utils.FormatBrasiliaDefault(utils.Now())

	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #2196F3; color: white; padding: 20px; text-align: center; }
        .content { background-color: #f9f9f9; padding: 30px; border-radius: 5px; margin-top: 20px; }
        .info-box { background-color: #fff; border: 2px solid #2196F3; padding: 15px; margin: 20px 0; border-radius: 5px; }
        .alert { background-color: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 15px 0; }
        .footer { text-align: center; margin-top: 20px; font-size: 12px; color: #777; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔔 Nova Sessão Detectada</h1>
        </div>
        <div class="content">
            <p>Olá <strong>%s</strong>,</p>
            <p>Detectamos um novo login na sua conta DashTrack.</p>
            
            <div class="info-box">
                <h3 style="margin: 0 0 10px 0; color: #2196F3;">📍 Detalhes da Nova Sessão:</h3>
                <p style="margin: 5px 0;"><strong>Data/Hora:</strong> %s</p>
                <p style="margin: 5px 0;"><strong>Endereço IP:</strong> %s</p>
                <p style="margin: 5px 0;"><strong>Dispositivo:</strong> %s</p>
            </div>
            
            <div class="alert">
                <strong>ℹ️ Limite de Sessões Atingido</strong>
                <p style="margin: 10px 0;">
                    Você tinha <strong>%d sessão(ões)</strong> ativa(s) que foi(ram) encerrada(s) automaticamente 
                    para permitir este novo login, pois o limite máximo é de <strong>3 sessões simultâneas</strong>.
                </p>
            </div>
            
            <div style="background-color: #e3f2fd; border-left: 4px solid #2196F3; padding: 15px; margin: 20px 0;">
                <h4 style="margin: 0 0 10px 0; color: #1976d2;">🛡️ Dicas de Segurança:</h4>
                <ul style="margin: 5px 0; padding-left: 20px;">
                    <li style="margin: 5px 0;">Sempre faça logout ao sair de dispositivos compartilhados</li>
                    <li style="margin: 5px 0;">Verifique suas sessões ativas regularmente no painel de controle</li>
                    <li style="margin: 5px 0;">Use senhas fortes e únicas</li>
                </ul>
            </div>
            
            <div class="alert" style="background-color: #f8d7da; border-left-color: #dc3545;">
                <strong>⚠️ Não foi você?</strong>
                <p style="margin: 10px 0;">
                    Se você não reconhece este login, <strong>altere sua senha imediatamente</strong> 
                    e revogue todas as sessões ativas no painel de controle.
                </p>
            </div>
        </div>
        <div class="footer">
            <p>DashTrack - Sistema de Gestão de Entregas</p>
            <p>Este é um email automático, não responda.</p>
        </div>
    </div>
</body>
</html>
`, name, loginTime, newIP, newUserAgent, revokedCount)

	err := h.emailService.SendEmail(services.EmailData{
		To:      email,
		Subject: subject,
		Body:    body,
		IsHTML:  true,
	})

	if err != nil {
		logger.Error("Failed to send new session alert email",
			zap.Error(err),
			zap.String("email", email))
	} else {
		logger.Info("New session alert email sent",
			zap.String("email", email),
			zap.String("ip", newIP),
			zap.Int("revoked_count", revokedCount))
	}
}
