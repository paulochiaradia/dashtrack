package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/paulochiaradia/dashtrack/internal/models"
)

// TeamRepository handles database operations for teams
type TeamRepository struct {
	db     *sqlx.DB
	tracer trace.Tracer
}

// NewTeamRepository creates a new team repository
func NewTeamRepository(db *sqlx.DB) *TeamRepository {
	return &TeamRepository{
		db:     db,
		tracer: otel.Tracer("team-repository"),
	}
}

// Create creates a new team
func (r *TeamRepository) Create(ctx context.Context, team *models.Team) error {
	ctx, span := r.tracer.Start(ctx, "TeamRepository.Create",
		trace.WithAttributes(
			attribute.String("team.name", team.Name),
			attribute.String("company.id", team.CompanyID.String()),
		))
	defer span.End()

	team.ID = uuid.New()
	team.CreatedAt = time.Now()
	team.UpdatedAt = time.Now()
	if team.Status == "" {
		team.Status = "active"
	}

	query := `
		INSERT INTO teams (
			id, company_id, name, description, manager_id, status, created_at, updated_at
		) VALUES (
			:id, :company_id, :name, :description, :manager_id, :status, :created_at, :updated_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, team)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create team: %w", err)
	}

	span.SetAttributes(attribute.String("team.id", team.ID.String()))
	return nil
}

// GetByID retrieves a team by ID with company context
func (r *TeamRepository) GetByID(ctx context.Context, id uuid.UUID, companyID uuid.UUID) (*models.Team, error) {
	ctx, span := r.tracer.Start(ctx, "TeamRepository.GetByID",
		trace.WithAttributes(
			attribute.String("team.id", id.String()),
			attribute.String("company.id", companyID.String()),
		))
	defer span.End()

	var team models.Team
	query := `
		SELECT id, company_id, name, description, manager_id, status, created_at, updated_at
		FROM teams 
		WHERE id = $1 AND company_id = $2
	`

	err := r.db.GetContext(ctx, &team, query, id, companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get team by ID: %w", err)
	}

	return &team, nil
}

// GetByCompany retrieves all teams for a company
func (r *TeamRepository) GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]models.Team, error) {
	ctx, span := r.tracer.Start(ctx, "TeamRepository.GetByCompany",
		trace.WithAttributes(
			attribute.String("company.id", companyID.String()),
			attribute.Int("limit", limit),
			attribute.Int("offset", offset),
		))
	defer span.End()

	var teams []models.Team
	query := `
		SELECT id, company_id, name, description, manager_id, status, created_at, updated_at
		FROM teams 
		WHERE company_id = $1 AND status != 'deleted'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	err := r.db.SelectContext(ctx, &teams, query, companyID, limit, offset)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get teams by company: %w", err)
	}

	span.SetAttributes(attribute.Int("teams.count", len(teams)))
	return teams, nil
}

// Update updates a team
func (r *TeamRepository) Update(ctx context.Context, team *models.Team) error {
	ctx, span := r.tracer.Start(ctx, "TeamRepository.Update",
		trace.WithAttributes(attribute.String("team.id", team.ID.String())))
	defer span.End()

	team.UpdatedAt = time.Now()

	query := `
		UPDATE teams SET
			name = :name,
			description = :description,
			manager_id = :manager_id,
			status = :status,
			updated_at = :updated_at
		WHERE id = :id AND company_id = :company_id
	`

	result, err := r.db.NamedExecContext(ctx, query, team)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update team: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("team not found or not authorized")
	}

	return nil
}

// Delete soft deletes a team
func (r *TeamRepository) Delete(ctx context.Context, id uuid.UUID, companyID uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "TeamRepository.Delete",
		trace.WithAttributes(
			attribute.String("team.id", id.String()),
			attribute.String("company.id", companyID.String()),
		))
	defer span.End()

	query := `
		UPDATE teams 
		SET deleted_at = NOW(), updated_at = NOW() 
		WHERE id = $1 AND company_id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id, companyID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete team: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("team not found or not authorized")
	}

	return nil
}

// AddMember adds a user to a team
func (r *TeamRepository) AddMember(ctx context.Context, teamMember *models.TeamMember) error {
	ctx, span := r.tracer.Start(ctx, "TeamRepository.AddMember",
		trace.WithAttributes(
			attribute.String("team.id", teamMember.TeamID.String()),
			attribute.String("user.id", teamMember.UserID.String()),
			attribute.String("role", teamMember.RoleInTeam),
		))
	defer span.End()

	teamMember.ID = uuid.New()
	teamMember.JoinedAt = time.Now()

	query := `
		INSERT INTO team_members (id, team_id, user_id, role_in_team, joined_at)
		VALUES (:id, :team_id, :user_id, :role_in_team, :joined_at)
	`

	_, err := r.db.NamedExecContext(ctx, query, teamMember)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to add team member: %w", err)
	}

	span.SetAttributes(attribute.String("team_member.id", teamMember.ID.String()))

	// Log member addition to history
	// Get team to retrieve company_id
	var companyID uuid.UUID
	err = r.db.GetContext(ctx, &companyID, `SELECT company_id FROM teams WHERE id = $1`, teamMember.TeamID)
	if err == nil {
		newRole := teamMember.RoleInTeam
		history := &models.TeamMemberHistory{
			TeamID:        teamMember.TeamID,
			UserID:        teamMember.UserID,
			CompanyID:     companyID,
			NewRoleInTeam: &newRole,
			ChangeType:    "added",
		}

		// Log history (non-critical, don't fail if logging fails)
		if err := r.LogMemberChange(ctx, history); err != nil {
			span.RecordError(fmt.Errorf("failed to log member addition: %w", err))
		}
	}

	return nil
}

// RemoveMember removes a user from a team
func (r *TeamRepository) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "TeamRepository.RemoveMember",
		trace.WithAttributes(
			attribute.String("team.id", teamID.String()),
			attribute.String("user.id", userID.String()),
		))
	defer span.End()

	// Get current member state before removal (for history)
	var currentMember models.TeamMember
	var companyID uuid.UUID
	err := r.db.GetContext(ctx, &currentMember,
		`SELECT tm.*, t.company_id 
		 FROM team_members tm 
		 JOIN teams t ON tm.team_id = t.id 
		 WHERE tm.team_id = $1 AND tm.user_id = $2`,
		teamID, userID)

	if err == nil {
		err = r.db.GetContext(ctx, &companyID, `SELECT company_id FROM teams WHERE id = $1`, teamID)
	}

	query := `DELETE FROM team_members WHERE team_id = $1 AND user_id = $2`

	result, err2 := r.db.ExecContext(ctx, query, teamID, userID)
	if err2 != nil {
		span.RecordError(err2)
		return fmt.Errorf("failed to remove team member: %w", err2)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("team member not found")
	}

	// Log member removal to history
	if err == nil {
		prevRole := currentMember.RoleInTeam
		history := &models.TeamMemberHistory{
			TeamID:             teamID,
			UserID:             userID,
			CompanyID:          companyID,
			PreviousRoleInTeam: &prevRole,
			ChangeType:         "removed",
		}

		// Log history (non-critical)
		if err := r.LogMemberChange(ctx, history); err != nil {
			span.RecordError(fmt.Errorf("failed to log member removal: %w", err))
		}
	}

	return nil
}

// GetMembers retrieves all members of a team
func (r *TeamRepository) GetMembers(ctx context.Context, teamID uuid.UUID) ([]models.TeamMember, error) {
	ctx, span := r.tracer.Start(ctx, "TeamRepository.GetMembers",
		trace.WithAttributes(attribute.String("team.id", teamID.String())))
	defer span.End()

	var members []models.TeamMember
	query := `
		SELECT tm.id, tm.team_id, tm.user_id, tm.role_in_team, tm.joined_at,
			   u.name, u.email, u.phone, u.active
		FROM team_members tm
		JOIN users u ON tm.user_id = u.id
		WHERE tm.team_id = $1
		ORDER BY tm.joined_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, teamID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get team members: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var member models.TeamMember
		var user models.User

		err := rows.Scan(
			&member.ID, &member.TeamID, &member.UserID, &member.RoleInTeam, &member.JoinedAt,
			&user.Name, &user.Email, &user.Phone, &user.Active,
		)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}

		user.ID = member.UserID
		member.User = &user
		members = append(members, member)
	}

	span.SetAttributes(attribute.Int("members.count", len(members)))
	return members, nil
}

// UpdateMemberRole updates a team member's role
func (r *TeamRepository) UpdateMemberRole(ctx context.Context, teamID, userID uuid.UUID, newRole string) error {
	ctx, span := r.tracer.Start(ctx, "TeamRepository.UpdateMemberRole",
		trace.WithAttributes(
			attribute.String("team.id", teamID.String()),
			attribute.String("user.id", userID.String()),
			attribute.String("new_role", newRole),
		))
	defer span.End()

	// Get current role before update (for history)
	var currentRole string
	var companyID uuid.UUID
	err := r.db.GetContext(ctx, &currentRole,
		`SELECT role_in_team FROM team_members WHERE team_id = $1 AND user_id = $2`,
		teamID, userID)

	if err == nil {
		err = r.db.GetContext(ctx, &companyID, `SELECT company_id FROM teams WHERE id = $1`, teamID)
	}

	query := `
		UPDATE team_members 
		SET role_in_team = $1 
		WHERE team_id = $2 AND user_id = $3
	`

	result, err2 := r.db.ExecContext(ctx, query, newRole, teamID, userID)
	if err2 != nil {
		span.RecordError(err2)
		return fmt.Errorf("failed to update member role: %w", err2)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("team member not found")
	}

	// Log role change to history (only if role actually changed)
	if err == nil && currentRole != newRole {
		history := &models.TeamMemberHistory{
			TeamID:             teamID,
			UserID:             userID,
			CompanyID:          companyID,
			PreviousRoleInTeam: &currentRole,
			NewRoleInTeam:      &newRole,
			ChangeType:         "role_changed",
		}

		// Log history (non-critical)
		if err := r.LogMemberChange(ctx, history); err != nil {
			span.RecordError(fmt.Errorf("failed to log role change: %w", err))
		}
	}

	return nil
}

// GetTeamsByUser retrieves all teams a user belongs to
func (r *TeamRepository) GetTeamsByUser(ctx context.Context, userID uuid.UUID) ([]models.Team, error) {
	ctx, span := r.tracer.Start(ctx, "TeamRepository.GetTeamsByUser",
		trace.WithAttributes(attribute.String("user.id", userID.String())))
	defer span.End()

	var teams []models.Team
	query := `
		SELECT t.id, t.company_id, t.name, t.description, t.manager_id, t.status, t.created_at, t.updated_at
		FROM teams t
		JOIN team_members tm ON t.id = tm.team_id
		WHERE tm.user_id = $1 AND t.status = 'active'
		ORDER BY t.name ASC
	`

	err := r.db.SelectContext(ctx, &teams, query, userID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get teams by user: %w", err)
	}

	span.SetAttributes(attribute.Int("teams.count", len(teams)))
	return teams, nil
}

// CheckMemberExists checks if a user is already a member of a team
func (r *TeamRepository) CheckMemberExists(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "TeamRepository.CheckMemberExists",
		trace.WithAttributes(
			attribute.String("team.id", teamID.String()),
			attribute.String("user.id", userID.String()),
		))
	defer span.End()

	var count int
	query := `SELECT COUNT(*) FROM team_members WHERE team_id = $1 AND user_id = $2`

	err := r.db.GetContext(ctx, &count, query, teamID, userID)
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("failed to check member existence: %w", err)
	}

	return count > 0, nil
}

// ============================================================================
// TEAM MEMBER HISTORY METHODS
// ============================================================================

// LogMemberChange logs a change to team membership (add, remove, role change, transfer)
func (r *TeamRepository) LogMemberChange(ctx context.Context, history *models.TeamMemberHistory) error {
	ctx, span := r.tracer.Start(ctx, "TeamRepository.LogMemberChange",
		trace.WithAttributes(
			attribute.String("team.id", history.TeamID.String()),
			attribute.String("user.id", history.UserID.String()),
			attribute.String("change.type", history.ChangeType),
		))
	defer span.End()

	history.ID = uuid.New()
	history.ChangedAt = time.Now()
	history.CreatedAt = time.Now()

	query := `
		INSERT INTO team_member_history (
			id, team_id, user_id, company_id,
			previous_role_in_team, new_role_in_team,
			change_type, previous_team_id, new_team_id,
			changed_by_user_id, change_reason,
			changed_at, created_at
		) VALUES (
			:id, :team_id, :user_id, :company_id,
			:previous_role_in_team, :new_role_in_team,
			:change_type, :previous_team_id, :new_team_id,
			:changed_by_user_id, :change_reason,
			:changed_at, :created_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, history)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to log member change: %w", err)
	}

	return nil
}

// GetMemberHistory retrieves membership history for a team
func (r *TeamRepository) GetMemberHistory(ctx context.Context, teamID, companyID uuid.UUID, limit int) ([]models.TeamMemberHistory, error) {
	ctx, span := r.tracer.Start(ctx, "TeamRepository.GetMemberHistory",
		trace.WithAttributes(
			attribute.String("team.id", teamID.String()),
			attribute.Int("limit", limit),
		))
	defer span.End()

	if limit == 0 {
		limit = 50 // Default limit
	}

	query := `
		SELECT 
			h.id, h.team_id, h.user_id, h.company_id,
			h.previous_role_in_team, h.new_role_in_team,
			h.change_type, h.previous_team_id, h.new_team_id,
			h.changed_by_user_id, h.change_reason,
			h.changed_at, h.created_at
		FROM team_member_history h
		WHERE h.team_id = $1 AND h.company_id = $2
		ORDER BY h.changed_at DESC
		LIMIT $3
	`

	var history []models.TeamMemberHistory
	err := r.db.SelectContext(ctx, &history, query, teamID, companyID, limit)
	if err != nil && err != sql.ErrNoRows {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get member history: %w", err)
	}

	span.SetAttributes(attribute.Int("history.count", len(history)))

	return history, nil
}

// GetUserTeamHistory retrieves team membership history for a specific user
func (r *TeamRepository) GetUserTeamHistory(ctx context.Context, userID, companyID uuid.UUID, limit int) ([]models.TeamMemberHistory, error) {
	ctx, span := r.tracer.Start(ctx, "TeamRepository.GetUserTeamHistory",
		trace.WithAttributes(
			attribute.String("user.id", userID.String()),
			attribute.Int("limit", limit),
		))
	defer span.End()

	if limit == 0 {
		limit = 50
	}

	query := `
		SELECT 
			h.id, h.team_id, h.user_id, h.company_id,
			h.previous_role_in_team, h.new_role_in_team,
			h.change_type, h.previous_team_id, h.new_team_id,
			h.changed_by_user_id, h.change_reason,
			h.changed_at, h.created_at
		FROM team_member_history h
		WHERE h.user_id = $1 AND h.company_id = $2
		ORDER BY h.changed_at DESC
		LIMIT $3
	`

	var history []models.TeamMemberHistory
	err := r.db.SelectContext(ctx, &history, query, userID, companyID, limit)
	if err != nil && err != sql.ErrNoRows {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get user team history: %w", err)
	}

	span.SetAttributes(attribute.Int("history.count", len(history)))

	return history, nil
}

// GetMemberHistoryWithDetails retrieves membership history with populated user/team details
func (r *TeamRepository) GetMemberHistoryWithDetails(ctx context.Context, teamID, companyID uuid.UUID, limit int) ([]models.TeamMemberHistory, error) {
	ctx, span := r.tracer.Start(ctx, "TeamRepository.GetMemberHistoryWithDetails",
		trace.WithAttributes(
			attribute.String("team.id", teamID.String()),
			attribute.Int("limit", limit),
		))
	defer span.End()

	// Get history first
	history, err := r.GetMemberHistory(ctx, teamID, companyID, limit)
	if err != nil {
		return nil, err
	}

	// Populate user and team details for each history entry
	for i := range history {
		entry := &history[i]

		// Load user
		var user models.User
		err := r.db.GetContext(ctx, &user, `SELECT id, name, email, role FROM users WHERE id = $1`, entry.UserID)
		if err == nil {
			entry.User = &user
		}

		// Load team
		var team models.Team
		err = r.db.GetContext(ctx, &team, `SELECT id, name, company_id FROM teams WHERE id = $1`, entry.TeamID)
		if err == nil {
			entry.Team = &team
		}

		// Load previous team (for transfers)
		if entry.PreviousTeamID != nil {
			var prevTeam models.Team
			err = r.db.GetContext(ctx, &prevTeam, `SELECT id, name, company_id FROM teams WHERE id = $1`, entry.PreviousTeamID)
			if err == nil {
				entry.PreviousTeam = &prevTeam
			}
		}

		// Load new team (for transfers)
		if entry.NewTeamID != nil {
			var newTeam models.Team
			err = r.db.GetContext(ctx, &newTeam, `SELECT id, name, company_id FROM teams WHERE id = $1`, entry.NewTeamID)
			if err == nil {
				entry.NewTeam = &newTeam
			}
		}

		// Load changed by user
		if entry.ChangedByUserID != nil {
			var changedBy models.User
			err = r.db.GetContext(ctx, &changedBy, `SELECT id, name, email, role FROM users WHERE id = $1`, entry.ChangedByUserID)
			if err == nil {
				entry.ChangedByUser = &changedBy
			}
		}
	}

	return history, nil
}

// GetUserTeamHistoryWithDetails retrieves user's team history with populated details
func (r *TeamRepository) GetUserTeamHistoryWithDetails(ctx context.Context, userID, companyID uuid.UUID, limit int) ([]models.TeamMemberHistory, error) {
	ctx, span := r.tracer.Start(ctx, "TeamRepository.GetUserTeamHistoryWithDetails",
		trace.WithAttributes(
			attribute.String("user.id", userID.String()),
			attribute.Int("limit", limit),
		))
	defer span.End()

	// Get history first
	history, err := r.GetUserTeamHistory(ctx, userID, companyID, limit)
	if err != nil {
		return nil, err
	}

	// Populate details (same as GetMemberHistoryWithDetails)
	for i := range history {
		entry := &history[i]

		// Load user
		var user models.User
		err := r.db.GetContext(ctx, &user, `SELECT id, name, email, role FROM users WHERE id = $1`, entry.UserID)
		if err == nil {
			entry.User = &user
		}

		// Load team
		var team models.Team
		err = r.db.GetContext(ctx, &team, `SELECT id, name, company_id FROM teams WHERE id = $1`, entry.TeamID)
		if err == nil {
			entry.Team = &team
		}

		// Load previous team
		if entry.PreviousTeamID != nil {
			var prevTeam models.Team
			err = r.db.GetContext(ctx, &prevTeam, `SELECT id, name, company_id FROM teams WHERE id = $1`, entry.PreviousTeamID)
			if err == nil {
				entry.PreviousTeam = &prevTeam
			}
		}

		// Load new team
		if entry.NewTeamID != nil {
			var newTeam models.Team
			err = r.db.GetContext(ctx, &newTeam, `SELECT id, name, company_id FROM teams WHERE id = $1`, entry.NewTeamID)
			if err == nil {
				entry.NewTeam = &newTeam
			}
		}

		// Load changed by user
		if entry.ChangedByUserID != nil {
			var changedBy models.User
			err = r.db.GetContext(ctx, &changedBy, `SELECT id, name, email, role FROM users WHERE id = $1`, entry.ChangedByUserID)
			if err == nil {
				entry.ChangedByUser = &changedBy
			}
		}
	}

	return history, nil
}
