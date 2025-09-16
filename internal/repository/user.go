package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/models"
)

// UserRepositoryInterface defines the contract for user repository
type UserRepositoryInterface interface {
	Create(user *models.User) error
	GetByID(id uuid.UUID) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(id uuid.UUID, updateReq models.UpdateUserRequest) (*models.User, error)
	UpdatePassword(id uuid.UUID, hashedPassword string) error
	Delete(id uuid.UUID) error
	List(limit, offset int, active *bool, roleID *uuid.UUID) ([]*models.User, error)
	UpdateLoginAttempts(id uuid.UUID, attempts int, blockedUntil *time.Time) error
	UpdateLastLogin(id uuid.UUID) error
}

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
func (r *UserRepository) Update(id uuid.UUID, updateReq models.UpdateUserRequest) (*models.User, error) {
	// Build dynamic query based on provided fields
	var setParts []string
	var args []interface{}
	argIndex := 1

	if updateReq.Name != "" {
		setParts = append(setParts, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, updateReq.Name)
		argIndex++
	}
	if updateReq.Email != "" {
		setParts = append(setParts, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, updateReq.Email)
		argIndex++
	}
	if updateReq.Phone != "" {
		setParts = append(setParts, fmt.Sprintf("phone = $%d", argIndex))
		args = append(args, &updateReq.Phone)
		argIndex++
	}
	if updateReq.CPF != "" {
		setParts = append(setParts, fmt.Sprintf("cpf = $%d", argIndex))
		args = append(args, &updateReq.CPF)
		argIndex++
	}
	if updateReq.Avatar != "" {
		setParts = append(setParts, fmt.Sprintf("avatar = $%d", argIndex))
		args = append(args, &updateReq.Avatar)
		argIndex++
	}
	if updateReq.DashboardConfig != "" {
		setParts = append(setParts, fmt.Sprintf("dashboard_config = $%d", argIndex))
		args = append(args, &updateReq.DashboardConfig)
		argIndex++
	}
	if updateReq.Active != nil {
		setParts = append(setParts, fmt.Sprintf("active = $%d", argIndex))
		args = append(args, *updateReq.Active)
		argIndex++
	}

	if len(setParts) == 0 {
		return nil, fmt.Errorf("no fields to update")
	}

	// Always update the updated_at field
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Build and execute update query
	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d",
		strings.Join(setParts, ", "), argIndex)
	args = append(args, id)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	// Return updated user
	return r.GetByID(id)
}

// UpdatePassword updates only the user's password and password_changed_at timestamp
func (r *UserRepository) UpdatePassword(id uuid.UUID, hashedPassword string) error {
	query := `
		UPDATE users 
		SET password = $1, password_changed_at = $2, updated_at = $3 
		WHERE id = $4`

	_, err := r.db.Exec(query, hashedPassword, time.Now(), time.Now(), id)
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
