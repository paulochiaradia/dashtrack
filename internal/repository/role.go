package repository

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/models"
)

// RoleRepository handles role database operations
type RoleRepository struct {
	db *sql.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *sql.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

// GetAll retrieves all roles
func (r *RoleRepository) GetAll() ([]*models.Role, error) {
	query := "SELECT id, name, description, created_at, updated_at FROM roles ORDER BY name"

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}
		err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	return roles, rows.Err()
}

// GetByID retrieves a role by ID
func (r *RoleRepository) GetByID(id uuid.UUID) (*models.Role, error) {
	query := "SELECT id, name, description, created_at, updated_at FROM roles WHERE id = $1"

	role := &models.Role{}
	err := r.db.QueryRow(query, id).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return role, nil
}

// GetByName retrieves a role by name
func (r *RoleRepository) GetByName(name string) (*models.Role, error) {
	query := "SELECT id, name, description, created_at, updated_at FROM roles WHERE name = $1"

	role := &models.Role{}
	err := r.db.QueryRow(query, name).Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return role, nil
}
