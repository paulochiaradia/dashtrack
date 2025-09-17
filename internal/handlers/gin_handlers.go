package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/auth"
	"github.com/paulochiaradia/dashtrack/internal/logger"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// Helper function to convert string to string pointer
func stringPtr(s string) *string {
	return &s
}

// LoginGin handles login requests using Gin
func (h *AuthHandler) LoginGin(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Get user by email
	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		// Log failed attempt
		authLog := &models.AuthLog{
			ID:            uuid.New(),
			UserID:        nil, // No user found
			EmailAttempt:  req.Email,
			Success:       false,
			IPAddress:     stringPtr(clientIP),
			UserAgent:     stringPtr(userAgent),
			FailureReason: stringPtr("invalid_email"),
		}
		h.authLogRepo.Create(authLog)

		logger.Warn("Login attempt with invalid email",
			zap.String("email", req.Email),
			zap.String("client_ip", clientIP),
		)

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// Log failed attempt - wrong password
		authLog := &models.AuthLog{
			ID:            uuid.New(),
			UserID:        &user.ID,
			EmailAttempt:  req.Email,
			Success:       false,
			IPAddress:     stringPtr(clientIP),
			UserAgent:     stringPtr(userAgent),
			FailureReason: stringPtr("invalid_password"),
		}
		h.authLogRepo.Create(authLog)

		logger.Warn("Login attempt with invalid password",
			zap.String("email", req.Email),
			zap.String("user_id", user.ID.String()),
			zap.String("client_ip", clientIP),
		)

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check if user is active
	if !user.Active {
		// Log failed attempt - account disabled
		authLog := &models.AuthLog{
			ID:            uuid.New(),
			UserID:        &user.ID,
			EmailAttempt:  req.Email,
			Success:       false,
			IPAddress:     stringPtr(clientIP),
			UserAgent:     stringPtr(userAgent),
			FailureReason: stringPtr("account_disabled"),
		}
		h.authLogRepo.Create(authLog)

		logger.Warn("Login attempt with disabled account",
			zap.String("email", req.Email),
			zap.String("user_id", user.ID.String()),
			zap.String("client_ip", clientIP),
		)

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is disabled"})
		return
	}

	// Generate tokens
	userContext := auth.UserContext{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		RoleID:   user.RoleID,
		RoleName: user.Role.Name,
	}

	accessToken, refreshToken, err := h.jwtManager.GenerateTokens(userContext)
	if err != nil {
		logger.Error("Failed to generate tokens",
			zap.Error(err),
			zap.String("user_id", user.ID.String()),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Log successful attempt
	authLog := &models.AuthLog{
		ID:           uuid.New(),
		UserID:       &user.ID,
		EmailAttempt: req.Email,
		Success:      true,
		IPAddress:    stringPtr(clientIP),
		UserAgent:    stringPtr(userAgent),
	}
	h.authLogRepo.Create(authLog)

	logger.Info("Successful login",
		zap.String("email", req.Email),
		zap.String("user_id", user.ID.String()),
		zap.String("role", user.Role.Name),
		zap.String("client_ip", clientIP),
	)

	response := LoginResponse{
		User: UserResponse{
			ID:     user.ID.String(),
			Name:   user.Name,
			Email:  user.Email,
			Role:   user.Role.Name,
			Active: user.Active,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes in seconds
	}

	c.JSON(http.StatusOK, response)
}

// RefreshTokenGin handles refresh token requests using Gin
func (h *AuthHandler) RefreshTokenGin(c *gin.Context) {
	var req RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate refresh token and get user ID
	userID, err := h.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Get user to ensure they still exist and are active
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if !user.Active {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is disabled"})
		return
	}

	// Generate new tokens
	userContext := auth.UserContext{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		RoleID:   user.RoleID,
		RoleName: user.Role.Name,
	}

	accessToken, refreshToken, err := h.jwtManager.GenerateTokens(userContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	response := LoginResponse{
		User: UserResponse{
			ID:     user.ID.String(),
			Name:   user.Name,
			Email:  user.Email,
			Role:   user.Role.Name,
			Active: user.Active,
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    900, // 15 minutes in seconds
	}

	c.JSON(http.StatusOK, response)
}

// LogoutGin handles logout requests using Gin
func (h *AuthHandler) LogoutGin(c *gin.Context) {
	// For now, just return success
	// In a real application, you might want to blacklist the token
	c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
}

// MeGin handles profile requests using Gin
func (h *AuthHandler) MeGin(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	response := UserResponse{
		ID:     user.ID.String(),
		Name:   user.Name,
		Email:  user.Email,
		Role:   user.Role.Name,
		Active: user.Active,
	}

	c.JSON(http.StatusOK, response)
}

// ChangePasswordGin handles password change requests using Gin
func (h *AuthHandler) ChangePasswordGin(c *gin.Context) {
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), h.bcryptCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password
	if err := h.userRepo.UpdatePassword(userID, string(hashedPassword)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password updated successfully"})
}

// GetRolesGin handles role listing using Gin
func (h *AuthHandler) GetRolesGin(c *gin.Context) {
	roles := []string{"admin", "manager", "driver", "helper"}
	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// Placeholder methods for admin functionality
func (h *AuthHandler) GetUsersGin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *AuthHandler) CreateUserGin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *AuthHandler) GetUserByIDGin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *AuthHandler) UpdateUserGin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *AuthHandler) DeleteUserGin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

func (h *AuthHandler) GetStoreUsersGin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

// Password reset methods (placeholders for now)
func (h *AuthHandler) ForgotPasswordGin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Password reset not implemented yet"})
}

func (h *AuthHandler) ResetPasswordGin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Password reset not implemented yet"})
}
