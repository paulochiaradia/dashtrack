package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// SessionRepositoryInterface defines the contract for session repository
type SessionRepositoryInterface interface {
	CountActiveSessions(ctx context.Context, companyID *uuid.UUID) (int, error)
	GetAverageSessionDuration(ctx context.Context, companyID *uuid.UUID, from, to time.Time) (float64, error)
	CountUserActiveSessions(ctx context.Context, userID uuid.UUID) (int, error)
	GetUserAverageSessionDuration(ctx context.Context, userID uuid.UUID, from, to time.Time) (float64, error)
}

// SessionRepository handles session database operations
type SessionRepository struct {
	db     *sqlx.DB
	tracer trace.Tracer
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *sqlx.DB) *SessionRepository {
	return &SessionRepository{
		db:     db,
		tracer: otel.Tracer("session-repository"),
	}
}

// CountActiveSessions counts active sessions, optionally filtered by company
func (r *SessionRepository) CountActiveSessions(ctx context.Context, companyID *uuid.UUID) (int, error) {
	ctx, span := r.tracer.Start(ctx, "SessionRepository.CountActiveSessions")
	defer span.End()

	// Count sessions from user_sessions table that are still active
	query := `
		SELECT COUNT(DISTINCT us.user_id)
		FROM user_sessions us
		JOIN users u ON us.user_id = u.id
		WHERE us.expires_at > NOW() AND us.revoked = false`

	args := []interface{}{}

	if companyID != nil {
		query += " AND u.company_id = $1"
		args = append(args, *companyID)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		// If table doesn't exist, return 0 (sessions not implemented yet)
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return count, nil
}

// GetAverageSessionDuration gets average session duration for a time range
func (r *SessionRepository) GetAverageSessionDuration(ctx context.Context, companyID *uuid.UUID, from, to time.Time) (float64, error) {
	ctx, span := r.tracer.Start(ctx, "SessionRepository.GetAverageSessionDuration")
	defer span.End()

	// Calculate average session duration from user_sessions table
	query := `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (COALESCE(revoked_at, expires_at) - created_at))/60), 0)
		FROM user_sessions us
		JOIN users u ON us.user_id = u.id
		WHERE us.created_at BETWEEN $1 AND $2`

	args := []interface{}{from, to}

	if companyID != nil {
		query += " AND u.company_id = $3"
		args = append(args, *companyID)
	}

	var avgDuration float64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&avgDuration)
	if err != nil {
		// If table doesn't exist, return 0 (sessions not implemented yet)
		if err == sql.ErrNoRows {
			return 0.0, nil
		}
		return 0.0, err
	}

	return avgDuration, nil
}

// CountUserActiveSessions counts active sessions for a specific user
func (r *SessionRepository) CountUserActiveSessions(ctx context.Context, userID uuid.UUID) (int, error) {
	ctx, span := r.tracer.Start(ctx, "SessionRepository.CountUserActiveSessions")
	defer span.End()

	query := `
		SELECT COUNT(*)
		FROM user_sessions
		WHERE user_id = $1 AND expires_at > NOW() AND revoked = false`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		// If table doesn't exist, return 0 (sessions not implemented yet)
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	return count, nil
}

// GetUserAverageSessionDuration gets average session duration for a specific user
func (r *SessionRepository) GetUserAverageSessionDuration(ctx context.Context, userID uuid.UUID, from, to time.Time) (float64, error) {
	ctx, span := r.tracer.Start(ctx, "SessionRepository.GetUserAverageSessionDuration")
	defer span.End()

	query := `
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (COALESCE(revoked_at, expires_at) - created_at))/60), 0)
		FROM user_sessions
		WHERE user_id = $1 AND created_at BETWEEN $2 AND $3`

	var avgDuration float64
	err := r.db.QueryRowContext(ctx, query, userID, from, to).Scan(&avgDuration)
	if err != nil {
		// If table doesn't exist, return 0 (sessions not implemented yet)
		if err == sql.ErrNoRows {
			return 0.0, nil
		}
		return 0.0, err
	}

	return avgDuration, nil
}
