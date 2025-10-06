package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/paulochiaradia/dashtrack/internal/logger"
	"github.com/paulochiaradia/dashtrack/internal/models"
)

// TokenService handles JWT token operations with session management
type TokenService struct {
	db              *sqlx.DB
	jwtSecret       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	sessionManager  *SessionManager
}

// NewTokenService creates a new token service
func NewTokenService(db *sqlx.DB, jwtSecret string, accessTokenTTL, refreshTokenTTL time.Duration) *TokenService {
	return &TokenService{
		db:              db,
		jwtSecret:       []byte(jwtSecret),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		sessionManager:  NewSessionManager(db),
	}
}

// TokenPair represents an access and refresh token pair
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// GenerateTokenPair generates a new access and refresh token pair
func (ts *TokenService) GenerateTokenPair(ctx context.Context, user *models.User, clientIP, userAgent string) (*TokenPair, error) {
	now := time.Now()
	accessTokenExp := now.Add(ts.accessTokenTTL)
	refreshTokenExp := now.Add(ts.refreshTokenTTL)

	// Generate access token
	accessToken, err := ts.generateAccessToken(user, accessTokenExp)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := ts.generateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Check session limits and revoke old sessions if necessary
	const maxSessions = 3 // Maximum allowed concurrent sessions
	allowed, sessionsToRevoke, err := ts.sessionManager.CheckSessionLimits(ctx, user.ID, maxSessions)
	if err != nil {
		logger.Error("Failed to check session limits", zap.Error(err))
	} else if !allowed && len(sessionsToRevoke) > 0 {
		// Revoke oldest sessions to make room
		err = ts.sessionManager.RevokeOldestSessions(ctx, sessionsToRevoke, "session_limit_exceeded")
		if err != nil {
			logger.Error("Failed to revoke old sessions", zap.Error(err))
		} else {
			logger.Info("Revoked old sessions due to limit",
				zap.String("user_id", user.ID.String()),
				zap.Int("revoked_count", len(sessionsToRevoke)))
		}
	}

	// Store session in database
	session := &models.SessionToken{
		ID:               uuid.New(),
		UserID:           user.ID,
		AccessToken:      ts.hashToken(accessToken),
		RefreshToken:     ts.hashToken(refreshToken),
		IPAddress:        clientIP,
		UserAgent:        userAgent,
		ExpiresAt:        accessTokenExp,
		RefreshExpiresAt: refreshTokenExp,
		Revoked:          false,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err = ts.storeSession(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(ts.accessTokenTTL.Seconds()),
		ExpiresAt:    accessTokenExp,
	}, nil
}

// RefreshTokenPair generates a new token pair using a refresh token
func (ts *TokenService) RefreshTokenPair(ctx context.Context, refreshToken, clientIP, userAgent string) (*TokenPair, error) {
	logger.Info("Starting refresh token validation",
		zap.String("client_ip", clientIP),
		zap.String("refresh_token_prefix", refreshToken[:20]+"..."))

	// Validate refresh token
	session, err := ts.validateRefreshToken(ctx, refreshToken)
	if err != nil {
		logger.Error("Refresh token validation failed", zap.Error(err))
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	logger.Info("Refresh token validated successfully",
		zap.String("user_id", session.UserID.String()),
		zap.String("session_id", session.ID.String()))

	// Get user information
	user, err := ts.getUserByID(ctx, session.UserID)
	if err != nil {
		logger.Error("Failed to get user by ID", zap.Error(err))
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Revoke old session
	err = ts.revokeSession(ctx, session.ID)
	if err != nil {
		logger.Error("Failed to revoke old session", zap.Error(err))
	}

	// Generate new token pair
	return ts.GenerateTokenPair(ctx, user, clientIP, userAgent)
}

// ValidateAccessToken validates an access token
func (ts *TokenService) ValidateAccessToken(ctx context.Context, tokenString string) (*models.User, error) {
	// Parse JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return ts.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user_id in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id format: %w", err)
	}

	// Check if session exists and is not revoked
	sessionExists, err := ts.isSessionValid(ctx, ts.hashToken(tokenString))
	if err != nil {
		return nil, fmt.Errorf("failed to check session: %w", err)
	}

	if !sessionExists {
		return nil, fmt.Errorf("session not found or revoked")
	}

	// Get user information
	return ts.getUserByID(ctx, userID)
}

// RevokeAllUserSessions revokes all sessions for a user
func (ts *TokenService) RevokeAllUserSessions(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE session_tokens 
		SET revoked = true, revoked_at = NOW(), updated_at = NOW()
		WHERE user_id = $1 AND revoked = false
	`

	_, err := ts.db.ExecContext(ctx, query, userID)
	return err
}

// generateAccessToken generates a JWT access token
func (ts *TokenService) generateAccessToken(user *models.User, expiresAt time.Time) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    user.ID.String(),
		"email":      user.Email,
		"role":       user.Role.Name,
		"company_id": user.CompanyID,
		"exp":        expiresAt.Unix(),
		"iat":        time.Now().Unix(),
		"iss":        "dashtrack-api",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(ts.jwtSecret)
}

// generateRefreshToken generates a JWT refresh token for compatibility
func (ts *TokenService) generateRefreshToken(userID uuid.UUID) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"iss": "Dashtrack API",
		"sub": userID.String(),
		"exp": now.Add(ts.refreshTokenTTL).Unix(),
		"nbf": now.Unix(),
		"iat": now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(ts.jwtSecret)
}

// hashToken creates a hash of the token for storage
func (ts *TokenService) hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", hash)
}

// storeSession stores a session in the database
func (ts *TokenService) storeSession(ctx context.Context, session *models.SessionToken) error {
	query := `
		INSERT INTO session_tokens (
			id, user_id, access_token_hash, refresh_token_hash, ip_address, user_agent,
			expires_at, refresh_expires_at, revoked, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err := ts.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.AccessToken, session.RefreshToken,
		session.IPAddress, session.UserAgent, session.ExpiresAt, session.RefreshExpiresAt,
		session.Revoked, session.CreatedAt, session.UpdatedAt,
	)

	return err
}

// validateRefreshToken validates a JWT refresh token and returns the session
func (ts *TokenService) validateRefreshToken(ctx context.Context, refreshToken string) (*models.SessionToken, error) {
	// First validate JWT format and get user ID
	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return ts.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	userIDStr, ok := claims["sub"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID format")
	}

	// Then find session by hashed token
	hashedToken := ts.hashToken(refreshToken)

	query := `
		SELECT id, user_id, access_token_hash, refresh_token_hash, ip_address, user_agent,
			   expires_at, refresh_expires_at, revoked, revoked_at, created_at, updated_at
		FROM session_tokens
		WHERE refresh_token_hash = $1 AND user_id = $2 AND revoked = false AND refresh_expires_at > NOW()
	`

	var session models.SessionToken
	err = ts.db.GetContext(ctx, &session, query, hashedToken, userID)
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// isSessionValid checks if a session is valid (not revoked)
func (ts *TokenService) isSessionValid(ctx context.Context, accessTokenHash string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM session_tokens
			WHERE access_token_hash = $1 AND revoked = false AND expires_at > NOW()
		)
	`

	var exists bool
	err := ts.db.GetContext(ctx, &exists, query, accessTokenHash)
	return exists, err
}

// revokeSession revokes a specific session
func (ts *TokenService) revokeSession(ctx context.Context, sessionID uuid.UUID) error {
	query := `
		UPDATE session_tokens
		SET revoked = true, revoked_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`

	_, err := ts.db.ExecContext(ctx, query, sessionID)
	return err
}

// getUserByID gets user information by ID
func (ts *TokenService) getUserByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	// Use a simpler query that matches the User model structure
	query := `
		SELECT id, name, email, phone, cpf, avatar, role_id, company_id,
			   active, last_login, created_at, updated_at
		FROM users
		WHERE id = $1 AND active = true
	`

	user := &models.User{}
	err := ts.db.GetContext(ctx, user, query, userID)
	if err != nil {
		return nil, err
	}

	// Separately get role information
	roleQuery := `
		SELECT id, name, description, created_at, updated_at
		FROM roles
		WHERE id = $1
	`

	role := &models.Role{}
	err = ts.db.GetContext(ctx, role, roleQuery, user.RoleID)
	if err != nil {
		return nil, err
	}

	user.Role = role
	return user, nil
}

// CleanupExpiredSessions removes expired sessions from database
func (ts *TokenService) CleanupExpiredSessions(ctx context.Context) error {
	query := `
		DELETE FROM session_tokens
		WHERE refresh_expires_at < NOW() - INTERVAL '7 days'
	`

	result, err := ts.db.ExecContext(ctx, query)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	logger.Info("Cleaned up expired sessions", zap.Int64("count", rowsAffected))

	return nil
}
