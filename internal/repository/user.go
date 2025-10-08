package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/paulochiaradia/dashtrack/internal/models"
)

// UserRepositoryInterface defines the contract for user repository
type UserRepositoryInterface interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]*models.User, error)
	Update(ctx context.Context, id uuid.UUID, updateReq models.UpdateUserRequest) (*models.User, error)
	UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error
	UpdateCompany(ctx context.Context, userID, companyID uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int, active *bool, roleID *uuid.UUID) ([]*models.User, error)
	ListByCompanyAndRoles(ctx context.Context, companyID *uuid.UUID, roles []string, limit, offset int) ([]*models.User, error)
	ListByRoles(ctx context.Context, roles []string, limit, offset int) ([]*models.User, error)
	CountByCompanyAndRoles(ctx context.Context, companyID *uuid.UUID, roles []string) (int, error)
	UpdateLoginAttempts(ctx context.Context, id uuid.UUID, attempts int, blockedUntil *time.Time) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	GetUserContext(ctx context.Context, userID uuid.UUID) (*models.UserContext, error)

	// Dashboard methods
	CountUsers(ctx context.Context, companyID *uuid.UUID) (int, error)
	CountActiveUsers(ctx context.Context, companyID *uuid.UUID) (int, error)
}

// UserRepository handles user database operations
type UserRepository struct {
	db     *sqlx.DB
	tracer trace.Tracer
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{
		db:     db,
		tracer: otel.Tracer("user-repository"),
	}
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	ctx, span := r.tracer.Start(ctx, "UserRepository.Create",
		trace.WithAttributes(
			attribute.String("user.email", user.Email),
			attribute.String("user.name", user.Name),
		))
	defer span.End()

	if user.CompanyID != nil {
		span.SetAttributes(attribute.String("company.id", user.CompanyID.String()))
	}

	// Set defaults
	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.PasswordChangedAt = time.Now()
	if user.Active == false {
		user.Active = true
	}

	query := `
		INSERT INTO users (
			id, name, email, password, phone, cpf, avatar, role_id, company_id, 
			active, dashboard_config, api_token, created_at, updated_at, password_changed_at
		) VALUES (
			:id, :name, :email, :password, :phone, :cpf, :avatar, :role_id, :company_id,
			:active, :dashboard_config, :api_token, :created_at, :updated_at, :password_changed_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, user)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create user: %w", err)
	}

	span.SetAttributes(attribute.String("user.id", user.ID.String()))
	return nil
}

// GetByID retrieves a user by ID with role information
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.GetByID",
		trace.WithAttributes(attribute.String("user.id", id.String())))
	defer span.End()

	query := `
		SELECT u.id, u.name, u.email, u.password, u.phone, u.cpf, u.avatar, u.role_id, u.company_id,
		       u.active, u.last_login, u.dashboard_config, u.api_token, u.login_attempts,
		       u.blocked_until, u.password_changed_at, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1 AND u.deleted_at IS NULL`

	user := &models.User{Role: &models.Role{}}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Phone,
		&user.CPF,
		&user.Avatar,
		&user.RoleID,
		&user.CompanyID,
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
		if err == sql.ErrNoRows {
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetByEmail retrieves a user by email with role information
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.GetByEmail",
		trace.WithAttributes(attribute.String("user.email", email)))
	defer span.End()

	query := `
		SELECT u.id, u.name, u.email, u.password, u.phone, u.cpf, u.avatar, u.role_id, u.company_id,
		       u.active, u.last_login, u.dashboard_config, u.api_token, u.login_attempts,
		       u.blocked_until, u.password_changed_at, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.email = $1 AND u.deleted_at IS NULL`

	user := &models.User{Role: &models.Role{}}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.Phone,
		&user.CPF,
		&user.Avatar,
		&user.RoleID,
		&user.CompanyID,
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
		if err == sql.ErrNoRows {
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetByCompany retrieves all users for a specific company
func (r *UserRepository) GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.GetByCompany",
		trace.WithAttributes(
			attribute.String("company.id", companyID.String()),
			attribute.Int("limit", limit),
			attribute.Int("offset", offset),
		))
	defer span.End()

	query := `
		SELECT u.id, u.name, u.email, u.phone, u.cpf, u.avatar, u.role_id, u.company_id,
		       u.active, u.last_login, u.dashboard_config, u.login_attempts,
		       u.blocked_until, u.password_changed_at, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.company_id = $1 AND u.active = true
		ORDER BY u.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, companyID, limit, offset)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get users by company: %w", err)
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
			&user.CompanyID,
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
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	span.SetAttributes(attribute.Int("users.count", len(users)))
	return users, nil
}

// Update updates user information
func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, updateReq models.UpdateUserRequest) (*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.Update",
		trace.WithAttributes(attribute.String("user.id", id.String())))
	defer span.End()

	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if updateReq.Name != "" {
		updates = append(updates, fmt.Sprintf("name = $%d", argIndex))
		args = append(args, updateReq.Name)
		argIndex++
	}

	if updateReq.Email != "" {
		updates = append(updates, fmt.Sprintf("email = $%d", argIndex))
		args = append(args, updateReq.Email)
		argIndex++
	}

	if updateReq.Phone != "" {
		updates = append(updates, fmt.Sprintf("phone = $%d", argIndex))
		args = append(args, updateReq.Phone)
		argIndex++
	}

	if updateReq.CPF != "" {
		updates = append(updates, fmt.Sprintf("cpf = $%d", argIndex))
		args = append(args, updateReq.CPF)
		argIndex++
	}

	if updateReq.Avatar != "" {
		updates = append(updates, fmt.Sprintf("avatar = $%d", argIndex))
		args = append(args, updateReq.Avatar)
		argIndex++
	}

	if updateReq.Active != nil {
		updates = append(updates, fmt.Sprintf("active = $%d", argIndex))
		args = append(args, *updateReq.Active)
		argIndex++
	}

	if updateReq.DashboardConfig != "" {
		updates = append(updates, fmt.Sprintf("dashboard_config = $%d", argIndex))
		args = append(args, updateReq.DashboardConfig)
		argIndex++
	}

	if len(updates) == 0 {
		return r.GetByID(ctx, id)
	}

	// Add updated_at timestamp
	updates = append(updates, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add the ID for WHERE clause
	args = append(args, id)

	query := fmt.Sprintf("UPDATE users SET %s WHERE id = $%d", strings.Join(updates, ", "), argIndex)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Return updated user
	return r.GetByID(ctx, id)
}

// UpdatePassword updates only the user's password and password_changed_at timestamp
func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	ctx, span := r.tracer.Start(ctx, "UserRepository.UpdatePassword",
		trace.WithAttributes(attribute.String("user.id", id.String())))
	defer span.End()

	query := `
		UPDATE users 
		SET password = $1, password_changed_at = $2, updated_at = $3 
		WHERE id = $4`

	_, err := r.db.ExecContext(ctx, query, hashedPassword, time.Now(), time.Now(), id)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// UpdateCompany updates a user's company (Master only operation)
func (r *UserRepository) UpdateCompany(ctx context.Context, userID, companyID uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "UserRepository.UpdateCompany",
		trace.WithAttributes(
			attribute.String("user.id", userID.String()),
			attribute.String("company.id", companyID.String()),
		))
	defer span.End()

	query := `
		UPDATE users 
		SET company_id = $1, updated_at = $2 
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, companyID, time.Now(), userID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update user company: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Delete soft deletes a user (sets active = false)
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "UserRepository.Delete",
		trace.WithAttributes(attribute.String("user.id", id.String())))
	defer span.End()

	query := `UPDATE users SET deleted_at = $1, updated_at = $1 WHERE id = $2`

	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// List retrieves users with optional filters
func (r *UserRepository) List(ctx context.Context, limit, offset int, active *bool, roleID *uuid.UUID) ([]*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.List",
		trace.WithAttributes(
			attribute.Int("limit", limit),
			attribute.Int("offset", offset),
		))
	defer span.End()

	whereConditions := []string{"u.deleted_at IS NULL"} // Always exclude soft-deleted users
	args := []interface{}{}
	argIndex := 1

	if active != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.active = $%d", argIndex))
		args = append(args, *active)
		argIndex++
		span.SetAttributes(attribute.Bool("filter.active", *active))
	}

	if roleID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.role_id = $%d", argIndex))
		args = append(args, *roleID)
		argIndex++
		span.SetAttributes(attribute.String("filter.role_id", roleID.String()))
	}

	whereClause := ""
	if len(whereConditions) > 0 {
		whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
	}

	query := fmt.Sprintf(`
		SELECT u.id, u.name, u.email, u.phone, u.cpf, u.avatar, u.role_id, u.company_id,
		       u.active, u.last_login, u.dashboard_config, u.login_attempts,
		       u.blocked_until, u.password_changed_at, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		%s
		ORDER BY u.created_at DESC
		LIMIT $%d OFFSET $%d`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list users: %w", err)
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
			&user.CompanyID,
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
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	span.SetAttributes(attribute.Int("users.count", len(users)))
	return users, nil
}

// UpdateLoginAttempts updates the login attempts and optionally blocked_until for a user
func (r *UserRepository) UpdateLoginAttempts(ctx context.Context, id uuid.UUID, attempts int, blockedUntil *time.Time) error {
	ctx, span := r.tracer.Start(ctx, "UserRepository.UpdateLoginAttempts",
		trace.WithAttributes(
			attribute.String("user.id", id.String()),
			attribute.Int("attempts", attempts),
		))
	defer span.End()

	query := `UPDATE users SET login_attempts = $1, blocked_until = $2, updated_at = $3 WHERE id = $4`

	_, err := r.db.ExecContext(ctx, query, attempts, blockedUntil, time.Now(), id)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update login attempts: %w", err)
	}

	return nil
}

// UpdateLastLogin updates the last_login timestamp for a user
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "UserRepository.UpdateLastLogin",
		trace.WithAttributes(attribute.String("user.id", id.String())))
	defer span.End()

	now := time.Now()
	query := `UPDATE users SET last_login = $1, updated_at = $2 WHERE id = $3`

	_, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// GetUserContext retrieves user context for permissions and multi-tenancy
func (r *UserRepository) GetUserContext(ctx context.Context, userID uuid.UUID) (*models.UserContext, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.GetUserContext",
		trace.WithAttributes(attribute.String("user.id", userID.String())))
	defer span.End()

	query := `
		SELECT u.id, u.company_id, r.name as role_name
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1 AND u.active = true`

	var userContext models.UserContext
	var roleName string

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&userContext.UserID,
		&userContext.CompanyID,
		&roleName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get user context: %w", err)
	}

	userContext.Role = roleName
	userContext.IsMaster = roleName == "master"

	return &userContext, nil
}

// Search searches users by name or email
func (r *UserRepository) Search(ctx context.Context, companyID *uuid.UUID, searchTerm string, limit, offset int) ([]*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.Search",
		trace.WithAttributes(
			attribute.String("search_term", searchTerm),
			attribute.Int("limit", limit),
			attribute.Int("offset", offset),
		))
	defer span.End()

	searchPattern := "%" + strings.ToLower(searchTerm) + "%"
	whereConditions := []string{"(LOWER(u.name) LIKE $1 OR LOWER(u.email) LIKE $1)", "u.active = true"}
	args := []interface{}{searchPattern}
	argIndex := 2

	if companyID != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("u.company_id = $%d", argIndex))
		args = append(args, *companyID)
		argIndex++
		span.SetAttributes(attribute.String("company.id", companyID.String()))
	}

	query := fmt.Sprintf(`
		SELECT u.id, u.name, u.email, u.phone, u.cpf, u.avatar, u.role_id, u.company_id,
		       u.active, u.last_login, u.dashboard_config, u.login_attempts,
		       u.blocked_until, u.password_changed_at, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE %s
		ORDER BY u.name ASC
		LIMIT $%d OFFSET $%d`, strings.Join(whereConditions, " AND "), argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to search users: %w", err)
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
			&user.CompanyID,
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
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	span.SetAttributes(attribute.Int("users.count", len(users)))
	return users, nil
}

// CountUsers counts total users, optionally filtered by company
func (r *UserRepository) CountUsers(ctx context.Context, companyID *uuid.UUID) (int, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.CountUsers")
	defer span.End()

	query := "SELECT COUNT(*) FROM users WHERE 1=1"
	args := []interface{}{}

	if companyID != nil {
		query += " AND company_id = $1"
		args = append(args, *companyID)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	return count, nil
}

// CountActiveUsers counts active users, optionally filtered by company
func (r *UserRepository) CountActiveUsers(ctx context.Context, companyID *uuid.UUID) (int, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.CountActiveUsers")
	defer span.End()

	query := "SELECT COUNT(*) FROM users WHERE active = true"
	args := []interface{}{}

	if companyID != nil {
		query += " AND company_id = $1"
		args = append(args, *companyID)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to count active users: %w", err)
	}

	return count, nil
}

// ListByCompanyAndRoles retrieves users by company and specific roles
func (r *UserRepository) ListByCompanyAndRoles(ctx context.Context, companyID *uuid.UUID, roles []string, limit, offset int) ([]*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.ListByCompanyAndRoles")
	defer span.End()

	if len(roles) == 0 {
		return []*models.User{}, nil
	}

	// Create placeholders for roles
	rolePlaceholders := make([]string, len(roles))
	args := []interface{}{}
	paramCount := 1

	for i, role := range roles {
		rolePlaceholders[i] = fmt.Sprintf("$%d", paramCount)
		args = append(args, role)
		paramCount++
	}

	query := `
		SELECT u.id, u.name, u.email, u.password, u.phone, u.cpf, u.avatar, u.role_id, u.company_id,
		       u.active, u.last_login, u.dashboard_config, u.api_token, u.login_attempts,
		       u.blocked_until, u.password_changed_at, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.deleted_at IS NULL AND r.name IN (` + strings.Join(rolePlaceholders, ",") + `)`

	if companyID != nil {
		query += fmt.Sprintf(" AND u.company_id = $%d", paramCount)
		args = append(args, *companyID)
		paramCount++
	}

	query += fmt.Sprintf(" ORDER BY u.created_at DESC LIMIT $%d OFFSET $%d", paramCount, paramCount+1)
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list users by company and roles: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{Role: &models.Role{}}
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Password,
			&user.Phone,
			&user.CPF,
			&user.Avatar,
			&user.RoleID,
			&user.CompanyID,
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
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// ListByRoles retrieves users by specific roles (for master and admin users)
func (r *UserRepository) ListByRoles(ctx context.Context, roles []string, limit, offset int) ([]*models.User, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.ListByRoles")
	defer span.End()

	if len(roles) == 0 {
		return []*models.User{}, nil
	}

	// Create placeholders for roles
	rolePlaceholders := make([]string, len(roles))
	args := []interface{}{}

	for i, role := range roles {
		rolePlaceholders[i] = fmt.Sprintf("$%d", i+1)
		args = append(args, role)
	}

	query := `
		SELECT u.id, u.name, u.email, u.password, u.phone, u.cpf, u.avatar, u.role_id, u.company_id,
		       u.active, u.last_login, u.dashboard_config, u.api_token, u.login_attempts,
		       u.blocked_until, u.password_changed_at, u.created_at, u.updated_at,
		       r.id, r.name, r.description, r.created_at, r.updated_at
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.deleted_at IS NULL AND r.name IN (` + strings.Join(rolePlaceholders, ",") + `)
		ORDER BY u.created_at DESC LIMIT $` + fmt.Sprintf("%d", len(roles)+1) + ` OFFSET $` + fmt.Sprintf("%d", len(roles)+2)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list users by roles: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{Role: &models.Role{}}
		err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.Password,
			&user.Phone,
			&user.CPF,
			&user.Avatar,
			&user.RoleID,
			&user.CompanyID,
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
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// CountByCompanyAndRoles counts users by company and specific roles
func (r *UserRepository) CountByCompanyAndRoles(ctx context.Context, companyID *uuid.UUID, roles []string) (int, error) {
	ctx, span := r.tracer.Start(ctx, "UserRepository.CountByCompanyAndRoles")
	defer span.End()

	if len(roles) == 0 {
		return 0, nil
	}

	// Create placeholders for roles
	rolePlaceholders := make([]string, len(roles))
	args := []interface{}{}
	paramCount := 1

	for i, role := range roles {
		rolePlaceholders[i] = fmt.Sprintf("$%d", paramCount)
		args = append(args, role)
		paramCount++
	}

	query := `
		SELECT COUNT(*)
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE r.name IN (` + strings.Join(rolePlaceholders, ",") + `)`

	if companyID != nil {
		query += fmt.Sprintf(" AND u.company_id = $%d", paramCount)
		args = append(args, *companyID)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to count users by company and roles: %w", err)
	}

	return count, nil
}
