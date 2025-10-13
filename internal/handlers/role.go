package handlers

import (
	"github.com/paulochiaradia/dashtrack/internal/repository"
)

// RoleHandler handles HTTP requests for role operations
type RoleHandler struct {
	roleRepo *repository.RoleRepository
}

// NewRoleHandler creates a new role handler
func NewRoleHandler(roleRepo *repository.RoleRepository) *RoleHandler {
	return &RoleHandler{
		roleRepo: roleRepo,
	}
}
