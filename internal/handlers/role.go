package handlers

import (
	"encoding/json"
	"net/http"

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

// ListRoles handles GET /roles
func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	roles, err := h.roleRepo.GetAll(r.Context())
	if err != nil {
		http.Error(w, "Failed to list roles", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"roles": roles,
	})
}
