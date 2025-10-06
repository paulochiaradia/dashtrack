package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/models"
)

// AuthLogRepositoryInterface defines the contract for auth log repository
type AuthLogRepositoryInterface interface {
	Create(log *models.AuthLog) error
	GetRecentFailedAttempts(email string, since time.Time) (int, error)
	GetByUserID(userID uuid.UUID, limit int) ([]*models.AuthLog, error)

	// Dashboard methods
	CountLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time) (int, error)
	CountSuccessfulLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time) (int, error)
	CountFailedLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time) (int, error)
	CountUserLogins(ctx context.Context, userID uuid.UUID, from, to time.Time) (int, error)
	CountUserSuccessfulLogins(ctx context.Context, userID uuid.UUID, from, to time.Time) (int, error)
	CountUserFailedLogins(ctx context.Context, userID uuid.UUID, from, to time.Time) (int, error)

	// Recent login methods
	GetRecentSuccessfulLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time, limit int) ([]models.RecentLogin, error)
	GetUserRecentSuccessfulLogins(ctx context.Context, userID uuid.UUID, from, to time.Time, limit int) ([]models.RecentLogin, error)
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

// CountLogins counts total login attempts for a company or all companies in a time range
func (r *AuthLogRepository) CountLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time) (int, error) {
	var query string
	var args []interface{}

	if companyID != nil {
		query = `
		SELECT COUNT(*) FROM auth_logs al
		JOIN users u ON al.user_id = u.id
		WHERE u.company_id = $1 AND al.created_at BETWEEN $2 AND $3
		`
		args = []interface{}{*companyID, from, to}
	} else {
		query = "SELECT COUNT(*) FROM auth_logs WHERE created_at BETWEEN $1 AND $2"
		args = []interface{}{from, to}
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// CountSuccessfulLogins counts successful login attempts
func (r *AuthLogRepository) CountSuccessfulLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time) (int, error) {
	var query string
	var args []interface{}

	if companyID != nil {
		query = `
		SELECT COUNT(*) FROM auth_logs al
		JOIN users u ON al.user_id = u.id
		WHERE u.company_id = $1 AND al.success = true AND al.created_at BETWEEN $2 AND $3
		`
		args = []interface{}{*companyID, from, to}
	} else {
		query = "SELECT COUNT(*) FROM auth_logs WHERE success = true AND created_at BETWEEN $1 AND $2"
		args = []interface{}{from, to}
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// CountFailedLogins counts failed login attempts
func (r *AuthLogRepository) CountFailedLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time) (int, error) {
	var query string
	var args []interface{}

	if companyID != nil {
		query = `
		SELECT COUNT(*) FROM auth_logs al
		JOIN users u ON al.user_id = u.id
		WHERE u.company_id = $1 AND al.success = false AND al.created_at BETWEEN $2 AND $3
		`
		args = []interface{}{*companyID, from, to}
	} else {
		query = "SELECT COUNT(*) FROM auth_logs WHERE success = false AND created_at BETWEEN $1 AND $2"
		args = []interface{}{from, to}
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// CountUserLogins counts login attempts for a specific user
func (r *AuthLogRepository) CountUserLogins(ctx context.Context, userID uuid.UUID, from, to time.Time) (int, error) {
	query := "SELECT COUNT(*) FROM auth_logs WHERE user_id = $1 AND created_at BETWEEN $2 AND $3"
	var count int
	err := r.db.QueryRowContext(ctx, query, userID, from, to).Scan(&count)
	return count, err
}

// CountUserSuccessfulLogins counts successful login attempts for a specific user
func (r *AuthLogRepository) CountUserSuccessfulLogins(ctx context.Context, userID uuid.UUID, from, to time.Time) (int, error) {
	query := "SELECT COUNT(*) FROM auth_logs WHERE user_id = $1 AND success = true AND created_at BETWEEN $2 AND $3"
	var count int
	err := r.db.QueryRowContext(ctx, query, userID, from, to).Scan(&count)
	return count, err
}

// CountUserFailedLogins counts failed login attempts for a specific user
func (r *AuthLogRepository) CountUserFailedLogins(ctx context.Context, userID uuid.UUID, from, to time.Time) (int, error) {
	query := "SELECT COUNT(*) FROM auth_logs WHERE user_id = $1 AND success = false AND created_at BETWEEN $2 AND $3"
	var count int
	err := r.db.QueryRowContext(ctx, query, userID, from, to).Scan(&count)
	return count, err
}

// GetRecentSuccessfulLogins gets recent successful logins with user information
func (r *AuthLogRepository) GetRecentSuccessfulLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time, limit int) ([]models.RecentLogin, error) {
	query := `
		SELECT 
			al.user_id, 
			COALESCE(u.name, '') as user_name,
			al.email_attempt as user_email,
			al.success,
			al.ip_address,
			al.user_agent,
			al.created_at as login_time,
			u.company_id,
			COALESCE(c.name, '') as company_name
		FROM auth_logs al
		LEFT JOIN users u ON al.user_id = u.id
		LEFT JOIN companies c ON u.company_id = c.id
		WHERE al.success = true 
		AND al.created_at BETWEEN $1 AND $2`

	args := []interface{}{from, to}

	if companyID != nil {
		query += " AND u.company_id = $3"
		args = append(args, *companyID)
	}

	query += " ORDER BY al.created_at DESC LIMIT $" + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logins []models.RecentLogin
	for rows.Next() {
		var login models.RecentLogin
		err := rows.Scan(
			&login.UserID,
			&login.UserName,
			&login.UserEmail,
			&login.Success,
			&login.IPAddress,
			&login.UserAgent,
			&login.LoginTime,
			&login.CompanyID,
			&login.CompanyName,
		)
		if err != nil {
			return nil, err
		}
		logins = append(logins, login)
	}

	return logins, rows.Err()
}

// GetUserRecentSuccessfulLogins gets recent successful logins for a specific user
func (r *AuthLogRepository) GetUserRecentSuccessfulLogins(ctx context.Context, userID uuid.UUID, from, to time.Time, limit int) ([]models.RecentLogin, error) {
	query := `
		SELECT 
			al.user_id, 
			COALESCE(u.name, '') as user_name,
			al.email_attempt as user_email,
			al.success,
			al.ip_address,
			al.user_agent,
			al.created_at as login_time,
			u.company_id,
			COALESCE(c.name, '') as company_name
		FROM auth_logs al
		LEFT JOIN users u ON al.user_id = u.id
		LEFT JOIN companies c ON u.company_id = c.id
		WHERE al.user_id = $1 
		AND al.success = true 
		AND al.created_at BETWEEN $2 AND $3
		ORDER BY al.created_at DESC 
		LIMIT $4`

	rows, err := r.db.QueryContext(ctx, query, userID, from, to, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logins []models.RecentLogin
	for rows.Next() {
		var login models.RecentLogin
		err := rows.Scan(
			&login.UserID,
			&login.UserName,
			&login.UserEmail,
			&login.Success,
			&login.IPAddress,
			&login.UserAgent,
			&login.LoginTime,
			&login.CompanyID,
			&login.CompanyName,
		)
		if err != nil {
			return nil, err
		}
		logins = append(logins, login)
	}

	return logins, rows.Err()
}
