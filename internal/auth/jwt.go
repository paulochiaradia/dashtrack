package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims represents the JWT claims structure
type JWTClaims struct {
	UserID   uuid.UUID  `json:"user_id"`
	Email    string     `json:"email"`
	Name     string     `json:"name"`
	RoleID   uuid.UUID  `json:"role_id"`
	RoleName string     `json:"role_name"`
	TenantID *uuid.UUID `json:"tenant_id,omitempty"` // For future multi-tenancy
	jwt.RegisteredClaims
}

// ToUserContext converts JWTClaims to UserContext
func (j *JWTClaims) ToUserContext() UserContext {
	return UserContext{
		UserID:   j.UserID,
		Email:    j.Email,
		Name:     j.Name,
		RoleID:   j.RoleID,
		RoleName: j.RoleName,
		TenantID: j.TenantID,
	}
}

// UserContext represents authenticated user context
type UserContext struct {
	UserID   uuid.UUID  `json:"user_id"`
	Email    string     `json:"email"`
	Name     string     `json:"name"`
	RoleID   uuid.UUID  `json:"role_id"`
	RoleName string     `json:"role_name"`
	TenantID *uuid.UUID `json:"tenant_id,omitempty"`
}

// JWTManager handles JWT operations
type JWTManager struct {
	secretKey     []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secretKey string, accessExpiry, refreshExpiry time.Duration, issuer string) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secretKey),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
		issuer:        issuer,
	}
}

// GenerateTokens generates both access and refresh tokens
func (j *JWTManager) GenerateTokens(userContext UserContext) (accessToken, refreshToken string, err error) {
	// Access token
	accessClaims := JWTClaims{
		UserID:   userContext.UserID,
		Email:    userContext.Email,
		Name:     userContext.Name,
		RoleID:   userContext.RoleID,
		RoleName: userContext.RoleName,
		TenantID: userContext.TenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
			Subject:   userContext.UserID.String(),
		},
	}

	accessTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessTokenObj.SignedString(j.secretKey)
	if err != nil {
		return "", "", err
	}

	// Refresh token (longer expiry, minimal claims)
	refreshClaims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.refreshExpiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		NotBefore: jwt.NewNumericDate(time.Now()),
		Issuer:    j.issuer,
		Subject:   userContext.UserID.String(),
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshTokenObj.SignedString(j.secretKey)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// ValidateToken validates and parses a JWT token (access token)
func (j *JWTManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns user ID
func (j *JWTManager) ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, errors.New("invalid refresh token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.Nil, errors.New("invalid refresh token claims")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, errors.New("invalid user ID in refresh token")
	}

	return userID, nil
}

// RefreshToken creates new tokens using a valid refresh token
func (j *JWTManager) RefreshToken(refreshTokenString string, userContext UserContext) (string, string, error) {
	// Validate the refresh token first
	_, err := j.ValidateRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", err
	}

	// Generate new tokens
	return j.GenerateTokens(userContext)
}

// PasswordResetClaims represents password reset token claims
type PasswordResetClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	jwt.RegisteredClaims
}

// GeneratePasswordResetToken generates a password reset token (valid for 1 hour)
func (j *JWTManager) GeneratePasswordResetToken(userID uuid.UUID, email string) (string, error) {
	claims := PasswordResetClaims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)), // 1 hour expiry
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    j.issuer,
			Subject:   "password_reset",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidatePasswordResetToken validates a password reset token and returns user info
func (j *JWTManager) ValidatePasswordResetToken(tokenString string) (uuid.UUID, string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &PasswordResetClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return uuid.Nil, "", err
	}

	if !token.Valid {
		return uuid.Nil, "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(*PasswordResetClaims)
	if !ok {
		return uuid.Nil, "", errors.New("invalid token claims")
	}

	return claims.UserID, claims.Email, nil
}
