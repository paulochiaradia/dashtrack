package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/paulochiaradia/dashtrack/internal/auth"
	"github.com/paulochiaradia/dashtrack/internal/repository"
	"github.com/paulochiaradia/dashtrack/internal/services"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userRepo     repository.UserRepositoryInterface
	authLogRepo  repository.AuthLogRepositoryInterface
	jwtManager   *auth.JWTManager
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
func NewAuthHandler(userRepo repository.UserRepositoryInterface, authLogRepo repository.AuthLogRepositoryInterface, jwtManager *auth.JWTManager, tokenService *services.TokenService, bcryptCost int) *AuthHandler {
	return &AuthHandler{
		userRepo:     userRepo,
		authLogRepo:  authLogRepo,
		jwtManager:   jwtManager,
		tokenService: tokenService,
		bcryptCost:   bcryptCost,
	}
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user by email with role information
	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if user is active
	if !user.Active {
		http.Error(w, "Account is disabled", http.StatusUnauthorized)
		return
	}

	// Verify role was loaded
	if user.Role == nil {
		http.Error(w, "User role not found", http.StatusInternalServerError)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Create user context for JWT
	userContext := auth.UserContext{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		RoleID:   user.RoleID,
		RoleName: user.Role.Name, // Assuming Role is populated
		TenantID: user.CompanyID, // Company ID as TenantID for multi-tenancy
	}

	// Generate tokens
	accessToken, refreshToken, err := h.jwtManager.GenerateTokens(userContext)
	if err != nil {
		http.Error(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := LoginResponse{
		User: UserResponse{
			ID:     user.ID.String(),
			Name:   user.Name,
			Email:  user.Email,
			Phone:  getStringValue(user.Phone),
			Role:   user.Role.Name,
			Active: user.Active,
			Avatar: getStringValue(user.Avatar),
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    15 * 60, // 15 minutes in seconds
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate refresh token
	userID, err := h.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Get fresh user data
	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if user is still active
	if !user.Active {
		http.Error(w, "Account is disabled", http.StatusUnauthorized)
		return
	}

	// Create fresh user context
	userContext := auth.UserContext{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		RoleID:   user.RoleID,
		RoleName: user.Role.Name,
		TenantID: nil,
	}
	// Generate new access token
	accessToken, refreshToken, err := h.jwtManager.GenerateTokens(userContext)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    15 * 60, // 15 minutes in seconds
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Me returns current user information
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userContext, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User context not found", http.StatusInternalServerError)
		return
	}

	// Get fresh user data
	user, err := h.userRepo.GetByID(r.Context(), userContext.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	response := UserResponse{
		ID:     user.ID.String(),
		Name:   user.Name,
		Email:  user.Email,
		Phone:  getStringValue(user.Phone),
		Role:   user.Role.Name,
		Active: user.Active,
		Avatar: getStringValue(user.Avatar),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Logout handles user logout (in a stateless JWT system, this is mainly for client-side cleanup)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// In a stateless JWT system, logout is mainly handled client-side
	// Here we could implement token blacklisting if needed in the future

	response := map[string]string{
		"message": "Successfully logged out",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ChangePassword handles password change
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userContext, ok := auth.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "User context not found", http.StatusInternalServerError)
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password" validate:"required"`
		NewPassword     string `json:"new_password" validate:"required,min=6"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get current user
	user, err := h.userRepo.GetByID(r.Context(), userContext.UserID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		http.Error(w, "Current password is incorrect", http.StatusBadRequest)
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Update password using the dedicated repository method
	err = h.userRepo.UpdatePassword(r.Context(), userContext.UserID, string(hashedPassword))
	if err != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Password updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function to get string value from pointer
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Gin Framework Methods

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

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate tokens
	userContext := auth.UserContext{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		RoleID:   user.RoleID,
		RoleName: user.Role.Name,
		TenantID: user.CompanyID,
	}

	accessToken, refreshToken, err := h.jwtManager.GenerateTokens(userContext)
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
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 hour
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

	// Validate refresh token and get user ID
	userID, err := h.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Get user details
	user, err := h.userRepo.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
		return
	}

	// Generate new access token
	userContext := auth.UserContext{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		RoleID:   user.RoleID,
		RoleName: user.Role.Name,
		TenantID: user.CompanyID,
	}
	accessToken, refreshToken, err := h.jwtManager.GenerateTokens(userContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := RefreshTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    3600, // 1 hour
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
	// TODO: Implement role repository and handler
	// For now, return basic roles
	roles := []gin.H{
		{"id": "1", "name": "master", "description": "Master administrator"},
		{"id": "2", "name": "company_admin", "description": "Company administrator"},
		{"id": "3", "name": "admin", "description": "Administrator"},
		{"id": "4", "name": "driver", "description": "Driver"},
		{"id": "5", "name": "helper", "description": "Helper"},
	}

	c.JSON(http.StatusOK, gin.H{"roles": roles})
}

// ForgotPasswordGin handles forgot password requests using Gin framework
func (h *AuthHandler) ForgotPasswordGin(c *gin.Context) {
	// TODO: Implement forgot password functionality
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Forgot password functionality not implemented yet"})
}

// ResetPasswordGin handles password reset using Gin framework
func (h *AuthHandler) ResetPasswordGin(c *gin.Context) {
	// TODO: Implement password reset functionality
	c.JSON(http.StatusNotImplemented, gin.H{"message": "Password reset functionality not implemented yet"})
}
