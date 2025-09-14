package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/models"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user into the database
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (id, name, email, password, phone, cpf, avatar, role_id, active, dashboard_config, api_token)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at, password_changed_at`
	
	err := r.db.QueryRow(
		query,
		user.ID,
		user.Name,
		user.Email,
		user.Password,
		user.Phone,
		user.CPF,
		user.Avatar,
		user.RoleID,
		user.Active,
		user.DashboardConfig,
		user.APIToken,
	).Scan(&user.CreatedAt, &user.UpdatedAt, &user.PasswordChangedAt)
	
	return err
}

// GetByID retrieves a user by ID with role information
func (r *UserRepository) GetByID(id uuid.UUID) (*models.User, error) {
	query := `
		SELECT u.id, u.name, u.email, u.password, u.phone, u.cpf, u.avatar, u.role_id,
		       u.active, u.last_login, u.dashboard_config, u.api_token, u.login_attempts,
		       u.blocked_until, u.password_changed_at, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1`
	
	user := &models.User{Role: &models.Role{}}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Phone,
		&user.CPF,
		&user.Avatar,
		&user.RoleID,
		&user.Active,
		&user.LastLogin,
		&user.DashboardConfig,
		&user.APIToken,
		&user.LoginAttempts,
		&user.BlockedUntil,
		&user.PasswordChangedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Description,
		&user.Role.CreatedAt,
		&user.Role.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

// GetByEmail retrieves a user by email with role information
func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	query := `
		SELECT u.id, u.name, u.email, u.password, u.phone, u.cpf, u.avatar, u.role_id,
		       u.active, u.last_login, u.dashboard_config, u.api_token, u.login_attempts,
		       u.blocked_until, u.password_changed_at, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1`
	
	user := &models.User{Role: &models.Role{}}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Phone,
		&user.CPF,
		&user.Avatar,
		&user.RoleID,
		&user.Active,
		&user.LastLogin,
		&user.DashboardConfig,
		&user.APIToken,
		&user.LoginAttempts,
		&user.BlockedUntil,
		&user.PasswordChangedAt,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Role.ID,
		&user.Role.Name,
		&user.Role.Description,
		&user.Role.CreatedAt,
		&user.Role.UpdatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

// Update updates user information
func (r *UserRepository) Update(id uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return fmt.Errorf("no fields to update")
	}
	
	// Build dynamic query
	setParts := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates)+1)
	argIndex := 1
	
	for field, value := range updates {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
		args = append(args, value)
		argIndex++
	}
	
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", 
		fmt.Sprintf("%s", setParts), argIndex)
	args = append(args, id)
	
	_, err := r.db.Exec(query, args...)
	return err
}

// Delete soft deletes a user (sets active = false)
func (r *UserRepository) Delete(id uuid.UUID) error {
	query := "UPDATE users SET active = false WHERE id = $1"
	_, err := r.db.Exec(query, id)
	return err
}

// List retrieves users with pagination and filtering
func (r *UserRepository) List(limit, offset int, active *bool, roleID *uuid.UUID) ([]*models.User, error) {
	query := `
		SELECT u.id, u.name, u.email, u.phone, u.cpf, u.avatar, u.role_id,
		       u.active, u.last_login, u.dashboard_config, u.login_attempts,
		       u.blocked_until, u.password_changed_at, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id`
	
	var whereClauses []string
	var args []interface{}
	argIndex := 1
	
	if active != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("u.active = $%d", argIndex))
		args = append(args, *active)
		argIndex++
	}
	
	if roleID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("u.role_id = $%d", argIndex))
		args = append(args, *roleID)
		argIndex++
	}
	
	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}
	
	query += fmt.Sprintf(" ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)
	
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var users []*models.User
	for rows.Next() {
		user := &models.User{Role: &models.Role{}}
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Phone,
			&user.CPF,
			&user.Avatar,
			&user.RoleID,
			&user.Active,
			&user.LastLogin,
			&user.DashboardConfig,
			&user.LoginAttempts,
			&user.BlockedUntil,
			&user.PasswordChangedAt,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.Role.ID,
			&user.Role.Name,
			&user.Role.Description,
			&user.Role.CreatedAt,
			&user.Role.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	
	return users, rows.Err()
}

// UpdateLoginAttempts updates login attempts and blocked_until fields
func (r *UserRepository) UpdateLoginAttempts(id uuid.UUID, attempts int, blockedUntil *time.Time) error {
	query := "UPDATE users SET login_attempts = $1, blocked_until = $2 WHERE id = $3"
	_, err := r.db.Exec(query, attempts, blockedUntil, id)
	return err
}

// UpdateLastLogin updates the last_login field
func (r *UserRepository) UpdateLastLogin(id uuid.UUID) error {
	query := "UPDATE users SET last_login = NOW() WHERE id = $1"
	_, err := r.db.Exec(query, id)
	return err
}
