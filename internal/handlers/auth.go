package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/paulochiaradia/dashtrack/internal/repository"
	"github.com/paulochiaradia/dashtrack/internal/services"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userRepo     repository.UserRepositoryInterface
	authLogRepo  repository.AuthLogRepositoryInterface
	roleRepo     repository.RoleRepositoryInterface
	tokenService *services.TokenService
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

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo repository.UserRepositoryInterface, authLogRepo repository.AuthLogRepositoryInterface, roleRepo repository.RoleRepositoryInterface, tokenService *services.TokenService, bcryptCost int) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		authLogRepo:  authLogRepo,
		roleRepo:     roleRepo,
		tokenService: tokenService,
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

	// Get user by email
	user, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	// Check if user is active
	if !user.Active {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is inactive"})
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Update last login
	_ = h.userRepo.UpdateLastLogin(c.Request.Context(), user.ID)

	// Generate token pair using tokenService (with session management)
	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	tokenPair, err := h.tokenService.GenerateTokenPair(c.Request.Context(), user, clientIP, userAgent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

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
	// In a JWT stateless system, logout is typically handled client-side
	// by simply discarding the token. However, for better security,
	// you might want to blacklist the token in a cache/database

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
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
