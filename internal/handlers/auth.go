package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/paulochiaradia/dashtrack/internal/auth"
	"github.com/paulochiaradia/dashtrack/internal/repository"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userRepo    repository.UserRepositoryInterface
	authLogRepo repository.AuthLogRepositoryInterface
	jwtManager  *auth.JWTManager
	bcryptCost  int
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
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// UserResponse represents user data in responses (no sensitive info)
type UserResponse struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Phone  string `json:"phone,omitempty"`
	Role   string `json:"role"`
	Active bool   `json:"active"`
	Avatar string `json:"avatar,omitempty"`
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userRepo repository.UserRepositoryInterface, authLogRepo repository.AuthLogRepositoryInterface, jwtManager *auth.JWTManager, bcryptCost int) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		authLogRepo: authLogRepo,
		jwtManager:  jwtManager,
		bcryptCost:  bcryptCost,
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
	user, err := h.userRepo.GetByEmail(req.Email)
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
		TenantID: nil,            // Will be used later for multi-tenancy
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
	user, err := h.userRepo.GetByID(userID)
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
	accessToken, _, err := h.jwtManager.GenerateTokens(userContext)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	response := RefreshTokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   15 * 60, // 15 minutes in seconds
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
	user, err := h.userRepo.GetByID(userContext.UserID)
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
	user, err := h.userRepo.GetByID(userContext.UserID)
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
	hashedPasswordStr := string(hashedPassword)
	err = h.userRepo.UpdatePassword(userContext.UserID, hashedPasswordStr)
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
