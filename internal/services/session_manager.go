package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/paulochiaradia/dashtrack/internal/logger"
	"go.uber.org/zap"
)

// SessionManager manages user sessions with advanced tracking
type SessionManager struct {
	db *sqlx.DB
}

// NewSessionManager creates a new session manager
func NewSessionManager(db *sqlx.DB) *SessionManager {
	return &SessionManager{
		db: db,
	}
}

// SessionMetrics represents session usage metrics
type SessionMetrics struct {
	UserID             uuid.UUID `json:"user_id" db:"user_id"`
	TotalSessions      int       `json:"total_sessions" db:"total_sessions"`
	ActiveSessions     int       `json:"active_sessions" db:"active_sessions"`
	TotalTimeSpent     float64   `json:"total_time_spent_minutes" db:"total_time_spent"` // in minutes
	LastLogin          time.Time `json:"last_login" db:"last_login"`
	LastIP             string    `json:"last_ip" db:"last_ip"`
	LastUserAgent      string    `json:"last_user_agent" db:"last_user_agent"`
	UniqueDevices      int       `json:"unique_devices" db:"unique_devices"`
	DifferentLocations int       `json:"different_locations" db:"different_locations"`
	TotalLogins        int       `json:"total_logins" db:"total_logins"`
	FailedAttempts     int       `json:"failed_attempts" db:"failed_attempts"`
	AverageSessionTime float64   `json:"average_session_minutes" db:"average_session_time"`
}

// ActiveSession represents a currently active session
type ActiveSession struct {
	ID                uuid.UUID `json:"id" db:"id"`
	UserID            uuid.UUID `json:"user_id" db:"user_id"`
	IPAddress         string    `json:"ip_address" db:"ip_address"`
	UserAgent         string    `json:"user_agent" db:"user_agent"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	LastActivity      time.Time `json:"last_activity" db:"last_activity"`
	ExpiresAt         time.Time `json:"expires_at" db:"expires_at"`
	Location          *string   `json:"location" db:"location"` // Estimated location from IP
	DeviceFingerprint *string   `json:"device_fingerprint" db:"device_fingerprint"`
	SessionDuration   float64   `json:"session_duration_minutes" db:"session_duration_minutes"` // Calculated field
}

// SecurityAlert represents a security concern
type SecurityAlert struct {
	ID          uuid.UUID `json:"id" db:"id"`
	UserID      uuid.UUID `json:"user_id" db:"user_id"`
	AlertType   string    `json:"alert_type" db:"alert_type"` // multiple_locations, too_many_devices, etc.
	Severity    string    `json:"severity" db:"severity"`     // low, medium, high, critical
	Description string    `json:"description" db:"description"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	UserAgent   string    `json:"user_agent" db:"user_agent"`
	Resolved    bool      `json:"resolved" db:"resolved"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// GetUserSessionMetrics retrieves comprehensive session metrics for a user
func (sm *SessionManager) GetUserSessionMetrics(ctx context.Context, userID uuid.UUID) (*SessionMetrics, error) {
	// Ultra simplified query with explicit casting to avoid type deduction issues
	query := `
		SELECT 
			$1::uuid as user_id,
			COUNT(*)::int as total_sessions,
			COUNT(CASE WHEN revoked = false AND refresh_expires_at > NOW() THEN 1 END)::int as active_sessions,
			COALESCE(SUM(EXTRACT(EPOCH FROM COALESCE(revoked_at, NOW()) - created_at) / 60), 0)::float8 as total_time_spent,
			COALESCE(MAX(created_at), NOW()) as last_login,
			'127.0.0.1'::text as last_ip,
			''::text as last_user_agent,
			COUNT(DISTINCT ip_address)::int as unique_devices,
			COUNT(DISTINCT ip_address)::int as different_locations,
			0::int as total_logins,
			0::int as failed_attempts,
			COALESCE(AVG(EXTRACT(EPOCH FROM COALESCE(revoked_at, NOW()) - created_at) / 60), 0)::float8 as average_session_time
		FROM session_tokens 
		WHERE user_id = $1::uuid
	`

	var metrics SessionMetrics
	err := sm.db.GetContext(ctx, &metrics, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session metrics: %w", err)
	}

	return &metrics, nil
}

// GetActiveSessionsForUser retrieves all active sessions for a user
func (sm *SessionManager) GetActiveSessionsForUser(ctx context.Context, userID uuid.UUID) ([]ActiveSession, error) {
	query := `
		SELECT 
			id,
			user_id,
			COALESCE(ip_address, '127.0.0.1') as ip_address,
			COALESCE(user_agent, '') as user_agent,
			created_at,
			updated_at as last_activity,
			refresh_expires_at as expires_at,
			COALESCE(EXTRACT(EPOCH FROM NOW() - created_at) / 60, 0) as session_duration_minutes
		FROM session_tokens
		WHERE user_id = $1 AND revoked = false AND refresh_expires_at > NOW()
		ORDER BY created_at DESC
	`

	var sessions []ActiveSession
	err := sm.db.SelectContext(ctx, &sessions, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}

	return sessions, nil
}

// CheckSessionLimits verifies if user has too many active sessions
func (sm *SessionManager) CheckSessionLimits(ctx context.Context, userID uuid.UUID, maxSessions int) (bool, []uuid.UUID, error) {
	sessions, err := sm.GetActiveSessionsForUser(ctx, userID)
	if err != nil {
		return false, nil, err
	}

	if len(sessions) < maxSessions {
		return true, nil, nil
	}

	// Return IDs of oldest sessions to revoke
	var sessionsToRevoke []uuid.UUID
	for i := maxSessions - 1; i < len(sessions); i++ {
		sessionsToRevoke = append(sessionsToRevoke, sessions[i].ID)
	}

	return false, sessionsToRevoke, nil
}

// RevokeOldestSessions revokes the oldest sessions for a user
func (sm *SessionManager) RevokeOldestSessions(ctx context.Context, sessionIDs []uuid.UUID, reason string) error {
	if len(sessionIDs) == 0 {
		return nil
	}

	// Begin transaction to update both tables
	tx, err := sm.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update session_tokens
	query1 := `
		UPDATE session_tokens 
		SET revoked = true, revoked_at = NOW(), updated_at = NOW()
		WHERE id = ANY($1)
	`

	_, err = tx.ExecContext(ctx, query1, sessionIDs)
	if err != nil {
		return fmt.Errorf("failed to revoke sessions in session_tokens: %w", err)
	}

	// Update user_sessions (mark as inactive)
	// Convertendo UUIDs para strings para usar com VARCHAR(128)
	sessionIDStrings := make([]string, len(sessionIDs))
	for i, id := range sessionIDs {
		sessionIDStrings[i] = id.String()
	}

	query2 := `
		UPDATE user_sessions 
		SET active = false
		WHERE id = ANY($1)
	`

	_, err = tx.ExecContext(ctx, query2, sessionIDStrings)
	if err != nil {
		return fmt.Errorf("failed to revoke sessions in user_sessions: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info("Revoked old sessions in both tables",
		zap.Int("count", len(sessionIDs)),
		zap.String("reason", reason),
		zap.Any("session_ids", sessionIDs))

	return nil
}

// DetectSuspiciousActivity analyzes sessions for security concerns
func (sm *SessionManager) DetectSuspiciousActivity(ctx context.Context, userID uuid.UUID) ([]SecurityAlert, error) {
	var alerts []SecurityAlert

	// Check for multiple simultaneous locations (different IP addresses - simplified)
	locationQuery := `
		SELECT COUNT(DISTINCT ip_address) as ip_count
		FROM session_tokens
		WHERE user_id = $1 AND revoked = false AND refresh_expires_at > NOW() AND ip_address IS NOT NULL
	`

	var ipCount int
	err := sm.db.GetContext(ctx, &ipCount, locationQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check IP count: %w", err)
	}

	if ipCount > 2 {
		alerts = append(alerts, SecurityAlert{
			ID:          uuid.New(),
			UserID:      userID,
			AlertType:   "multiple_locations",
			Severity:    "medium",
			Description: fmt.Sprintf("User has active sessions from %d different IP addresses", ipCount),
			CreatedAt:   time.Now(),
		})
	}

	// Check for too many active devices
	sessions, err := sm.GetActiveSessionsForUser(ctx, userID)
	if err != nil {
		return alerts, err
	}

	if len(sessions) > 5 {
		alerts = append(alerts, SecurityAlert{
			ID:          uuid.New(),
			UserID:      userID,
			AlertType:   "too_many_devices",
			Severity:    "high",
			Description: fmt.Sprintf("User has %d active sessions (limit: 5)", len(sessions)),
			CreatedAt:   time.Now(),
		})
	}

	return alerts, nil
}

// GetUserSessionDashboard returns comprehensive session information for dashboard
func (sm *SessionManager) GetUserSessionDashboard(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	// Get metrics
	metrics, err := sm.GetUserSessionMetrics(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get active sessions
	activeSessions, err := sm.GetActiveSessionsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get security alerts
	alerts, err := sm.DetectSuspiciousActivity(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Recent login history (last 10 sessions)
	recentQuery := `
		SELECT ip_address, user_agent, created_at, 
			   CASE WHEN revoked THEN 'ended' ELSE 'active' END as status,
			   EXTRACT(EPOCH FROM COALESCE(revoked_at, NOW()) - created_at) / 60 as duration_minutes
		FROM session_tokens
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 10
	`

	var recentSessions []map[string]interface{}
	rows, err := sm.db.QueryContext(ctx, recentQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ip, userAgent, status string
		var createdAt time.Time
		var duration float64

		err := rows.Scan(&ip, &userAgent, &createdAt, &status, &duration)
		if err != nil {
			continue
		}

		recentSessions = append(recentSessions, map[string]interface{}{
			"ip_address":       ip,
			"user_agent":       userAgent,
			"created_at":       createdAt,
			"status":           status,
			"duration_minutes": duration,
		})
	}

	return map[string]interface{}{
		"metrics":         metrics,
		"active_sessions": activeSessions,
		"security_alerts": alerts,
		"recent_sessions": recentSessions,
		"session_limit":   3, // Configurable limit
		"warnings": map[string]interface{}{
			"approaching_limit": len(activeSessions) >= 2,
			"security_concerns": len(alerts) > 0,
		},
	}, nil
}
