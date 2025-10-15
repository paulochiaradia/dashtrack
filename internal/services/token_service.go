package services

import (
	"context"
	"crypto/sha256"
	"encoding/json"
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
	emailService    *EmailService
}

// NewTokenService creates a new token service
func NewTokenService(db *sqlx.DB, jwtSecret string, accessTokenTTL, refreshTokenTTL time.Duration) *TokenService {
	return &TokenService{
		db:              db,
		jwtSecret:       []byte(jwtSecret),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		sessionManager:  NewSessionManager(db),
		emailService:    nil, // Will be set later via SetEmailService
	}
}

// SetEmailService sets the email service for sending notifications
func (ts *TokenService) SetEmailService(emailService *EmailService) {
	ts.emailService = emailService
}

// GetDB returns the database connection
func (ts *TokenService) GetDB() *sqlx.DB {
	return ts.db
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
	}

	// Variável para controlar se deve enviar email depois
	shouldSendEmail := false
	revokedCount := 0

	if !allowed && len(sessionsToRevoke) > 0 {
		// Revoke oldest sessions to make room
		err = ts.sessionManager.RevokeOldestSessions(ctx, sessionsToRevoke, "session_limit_exceeded")
		if err != nil {
			logger.Error("Failed to revoke old sessions", zap.Error(err))
		} else {
			logger.Info("Revoked old sessions due to limit",
				zap.String("user_id", user.ID.String()),
				zap.Int("revoked_count", len(sessionsToRevoke)))

			// Marcar para enviar email DEPOIS de criar a nova sessão
			shouldSendEmail = true
			revokedCount = len(sessionsToRevoke)
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

	// AGORA envia email DEPOIS de criar a nova sessão
	if shouldSendEmail && ts.emailService != nil {
		err = ts.sendSessionLimitEmail(user, clientIP, userAgent, revokedCount)
		if err != nil {
			logger.Error("Failed to send session limit email",
				zap.Error(err),
				zap.String("user_id", user.ID.String()))
		}
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

// ValidateAccessTokenWithSession validates a token and returns both user and session_id
func (ts *TokenService) ValidateAccessTokenWithSession(ctx context.Context, tokenString string) (*models.User, uuid.UUID, error) {
	// Parse JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return ts.jwtSecret, nil
	})

	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, uuid.Nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, uuid.Nil, fmt.Errorf("invalid token claims")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, uuid.Nil, fmt.Errorf("invalid user_id in token")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("invalid user_id format: %w", err)
	}

	// Get session ID from token hash
	tokenHash := ts.hashToken(tokenString)
	var sessionID uuid.UUID
	query := `SELECT id FROM session_tokens WHERE access_token_hash = $1 AND revoked = false`
	err = ts.db.GetContext(ctx, &sessionID, query, tokenHash)
	if err != nil {
		return nil, uuid.Nil, fmt.Errorf("session not found or revoked: %w", err)
	}

	// Get user information
	user, err := ts.getUserByID(ctx, userID)
	if err != nil {
		return nil, uuid.Nil, err
	}

	return user, sessionID, nil
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

// storeSession stores a session in the database (both session_tokens and user_sessions)
func (ts *TokenService) storeSession(ctx context.Context, session *models.SessionToken) error {
	// Begin transaction to ensure both inserts succeed or fail together
	tx, err := ts.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert into session_tokens (existing table)
	query1 := `
		INSERT INTO session_tokens (
			id, user_id, access_token_hash, refresh_token_hash, ip_address, user_agent,
			expires_at, refresh_expires_at, revoked, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = tx.ExecContext(ctx, query1,
		session.ID, session.UserID, session.AccessToken, session.RefreshToken,
		session.IPAddress, session.UserAgent, session.ExpiresAt, session.RefreshExpiresAt,
		session.Revoked, session.CreatedAt, session.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert into session_tokens: %w", err)
	}

	// Insert into user_sessions (new tracking table)
	// session_data JSON will contain useful metadata
	sessionData := map[string]interface{}{
		"session_id":        session.ID.String(),
		"access_token_exp":  session.ExpiresAt,
		"refresh_token_exp": session.RefreshExpiresAt,
		"device_info": map[string]interface{}{
			"user_agent": session.UserAgent,
			"ip_address": session.IPAddress,
		},
	}

	sessionDataJSON, err := json.Marshal(sessionData)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	query2 := `
		INSERT INTO user_sessions (
			id, user_id, ip_address, user_agent, session_data, expires_at, active, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = tx.ExecContext(ctx, query2,
		session.ID.String(), // usando session_id como id
		session.UserID,
		session.IPAddress,
		session.UserAgent,
		sessionDataJSON,
		session.RefreshExpiresAt, // expira junto com refresh token
		true,                     // active = true
		session.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert into user_sessions: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info("Session stored in both tables",
		zap.String("session_id", session.ID.String()),
		zap.String("user_id", session.UserID.String()))

	return nil
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

// sendSessionLimitEmail sends an email notification when sessions are revoked due to limit
func (ts *TokenService) sendSessionLimitEmail(user *models.User, newIP, newUserAgent string, revokedCount int) error {
	subject := "🔒 Nova sessão ativada - Sessões antigas revogadas"

	// Configurar timezone de Brasília
	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		location = time.UTC
	}

	// Buscar sessões ativas do usuário
	ctx := context.Background()
	activeSessions, err := ts.sessionManager.GetActiveSessionsForUser(ctx, user.ID)
	if err != nil {
		logger.Error("Failed to get active sessions for email", zap.Error(err))
		activeSessions = []ActiveSession{} // Continue mesmo se falhar
	}

	// Construir lista de sessões ativas em HTML
	sessionsListHTML := ""
	for i, session := range activeSessions {
		sessionTime := session.CreatedAt.In(location).Format("02/01/2006 às 15:04:05")
		isCurrent := (i == 0) // A primeira é a mais recente (sessão atual)
		currentBadge := ""
		if isCurrent {
			currentBadge = " <span style='background: #4caf50; color: white; padding: 2px 8px; border-radius: 3px; font-size: 11px;'>ATUAL</span>"
		}

		sessionsListHTML += fmt.Sprintf(`
                <div style="background: white; padding: 12px; margin: 10px 0; border-radius: 5px; border-left: 3px solid %s;">
                    <div style="font-weight: bold; margin-bottom: 5px;">Sessão %d%s</div>
                    <div style="font-size: 13px; color: #666;">
                        <div>📍 IP: %s</div>
                        <div>💻 Dispositivo: %s</div>
                        <div>🕐 Início: %s</div>
                    </div>
                </div>`,
			map[bool]string{true: "#4caf50", false: "#667eea"}[isCurrent],
			i+1,
			currentBadge,
			session.IPAddress,
			truncateUserAgent(session.UserAgent),
			sessionTime,
		)
	}

	currentTime := time.Now().In(location).Format("02/01/2006 às 15:04:05")

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%); color: white; padding: 30px; text-align: center; border-radius: 10px 10px 0 0; }
        .content { background: #f9f9f9; padding: 30px; border-radius: 0 0 10px 10px; }
        .info-box { background: white; border-left: 4px solid #667eea; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .warning-box { background: #fff3cd; border-left: 4px solid #ffc107; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .sessions-box { background: white; border-left: 4px solid #2196f3; padding: 15px; margin: 20px 0; border-radius: 4px; }
        .footer { text-align: center; margin-top: 30px; font-size: 12px; color: #666; }
        ul { padding-left: 20px; }
        li { margin: 8px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔒 Alerta de Segurança</h1>
            <p>Nova sessão ativada</p>
        </div>
        <div class="content">
            <p>Olá <strong>%s</strong>,</p>
            
            <p>Detectamos um novo login na sua conta DashTrack. Como você atingiu o limite de <strong>3 sessões simultâneas</strong>, revogamos automaticamente %d sessão(ões) antiga(s) para manter sua conta segura.</p>
            
            <div class="info-box">
                <h3>📱 Detalhes da Nova Sessão</h3>
                <ul>
                    <li><strong>Endereço IP:</strong> %s</li>
                    <li><strong>Dispositivo:</strong> %s</li>
                    <li><strong>Data/Hora:</strong> %s</li>
                </ul>
            </div>
            
            <div class="warning-box">
                <h3>⚠️ Sessões Revogadas</h3>
                <p><strong>%d sessão(ões) antiga(s)</strong> foi(foram) automaticamente revogada(s) para liberar espaço para esta nova sessão.</p>
                <p>As sessões mais antigas são sempre revogadas primeiro quando você atinge o limite de 3 sessões ativas.</p>
            </div>
            
            <h3>🔐 Não foi você?</h3>
            <p>Se você não reconhece esta atividade, recomendamos que você:</p>
            <ol>
                <li>Altere sua senha imediatamente</li>
                <li>Revogue todas as sessões ativas</li>
                <li>Verifique as configurações de segurança da sua conta</li>
            </ol>
            
            <div class="sessions-box">
                <h3>🖥️ Suas Sessões Ativas Atuais (%d)</h3>
                <p style="font-size: 13px; color: #666; margin-bottom: 15px;">Abaixo estão todas as suas sessões ativas no momento:</p>
                %s
            </div>
            
            <p style="margin-top: 30px; font-size: 14px; color: #666;">
                <strong>Dica de Segurança:</strong> Você pode gerenciar todas as suas sessões ativas e revogar dispositivos não reconhecidos a qualquer momento através do painel de segurança.
            </p>
        </div>
        <div class="footer">
            <p>Este é um email automático de segurança do DashTrack</p>
            <p>Se você tem dúvidas, entre em contato com nosso suporte</p>
        </div>
    </div>
</body>
</html>
`, user.Name, revokedCount, newIP, truncateUserAgent(newUserAgent), currentTime, revokedCount, len(activeSessions), sessionsListHTML)

	emailData := EmailData{
		To:      user.Email,
		Subject: subject,
		Body:    htmlBody,
		IsHTML:  true,
	}

	return ts.emailService.SendEmail(emailData)
}

// truncateUserAgent encurta o user-agent para exibição
func truncateUserAgent(ua string) string {
	if len(ua) > 80 {
		return ua[:77] + "..."
	}
	return ua
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
