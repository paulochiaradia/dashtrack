package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/models"
)

// AuthLogRepositoryInterface defines the contract for auth log repository
type AuthLogRepositoryInterface interface {
	Create(log *models.AuthLog) error
	GetRecentFailedAttempts(email string, since time.Time) (int, error)
	GetByUserID(userID uuid.UUID, limit int) ([]*models.AuthLog, error)
}

// AuthLogRepository handles authentication log database operations
type AuthLogRepository struct {
	db *sql.DB
}

// NewAuthLogRepository creates a new auth log repository
func NewAuthLogRepository(db *sql.DB) *AuthLogRepository {
	return &AuthLogRepository{db: db}
}

// Create inserts a new authentication log
func (r *AuthLogRepository) Create(log *models.AuthLog) error {
	query := `
		INSERT INTO auth_logs (id, user_id, email_attempt, success, ip_address, user_agent, failure_reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at`

	return r.db.QueryRow(
		query,
		log.ID,
		log.UserID,
		log.EmailAttempt,
		log.Success,
		log.IPAddress,
		log.UserAgent,
		log.FailureReason,
	).Scan(&log.CreatedAt)
}

// GetRecentFailedAttempts retrieves recent failed login attempts for an email
func (r *AuthLogRepository) GetRecentFailedAttempts(email string, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM auth_logs 
		WHERE email_attempt = $1 AND success = false AND created_at >= $2`

	var count int
	err := r.db.QueryRow(query, email, since).Scan(&count)
	return count, err
}

// GetByUserID retrieves auth logs for a specific user
func (r *AuthLogRepository) GetByUserID(userID uuid.UUID, limit int) ([]*models.AuthLog, error) {
	query := `
		SELECT id, user_id, email_attempt, success, ip_address, user_agent, failure_reason, created_at
		FROM auth_logs 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2`

	rows, err := r.db.Query(query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.AuthLog
	for rows.Next() {
		log := &models.AuthLog{}
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.EmailAttempt,
			&log.Success,
			&log.IPAddress,
			&log.UserAgent,
			&log.FailureReason,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}

// GetLoginHistory retrieves login history for a user with pagination
func (r *AuthLogRepository) GetLoginHistory(userID uuid.UUID, limit, offset int) ([]*models.AuthLog, error) {
	query := `
		SELECT id, user_id, email_attempt, success, ip_address, user_agent, failure_reason, created_at
		FROM auth_logs 
		WHERE user_id = $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.AuthLog
	for rows.Next() {
		log := &models.AuthLog{}
		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.EmailAttempt,
			&log.Success,
			&log.IPAddress,
			&log.UserAgent,
			&log.FailureReason,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}

	return logs, rows.Err()
}
