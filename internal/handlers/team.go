package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/paulochiaradia/dashtrack/internal/logger"
	"github.com/paulochiaradia/dashtrack/internal/middleware"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
	"github.com/paulochiaradia/dashtrack/internal/utils"
)

// TeamHandler handles team-related HTTP requests
type TeamHandler struct {
	teamRepo    *repository.TeamRepository
	userRepo    *repository.UserRepository
	vehicleRepo *repository.VehicleRepository
	tracer      trace.Tracer
}

// NewTeamHandler creates a new team handler
func NewTeamHandler(teamRepo *repository.TeamRepository, userRepo *repository.UserRepository, vehicleRepo *repository.VehicleRepository) *TeamHandler {
	return &TeamHandler{
		teamRepo:    teamRepo,
		userRepo:    userRepo,
		vehicleRepo: vehicleRepo,
		tracer:      otel.Tracer("team-handler"),
	}
}

// CreateTeam creates a new team
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "TeamHandler.CreateTeam")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	var req models.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate manager if provided
	if req.ManagerID != nil {
		manager, err := h.userRepo.GetByID(ctx, *req.ManagerID)
		if err != nil || manager == nil {
			utils.BadRequestResponse(c, "Invalid manager ID")
			return
		}

		// Check if manager belongs to the same company
		if manager.CompanyID == nil || *manager.CompanyID != *companyID {
			utils.BadRequestResponse(c, "Manager must belong to the same company")
			return
		}
	}

	team := &models.Team{
		CompanyID:   *companyID,
		Name:        req.Name,
		Description: req.Description,
		ManagerID:   req.ManagerID,
	}

	err = h.teamRepo.Create(ctx, team)
	if err != nil {
		span.RecordError(err)
		logger.Error("Failed to create team in database", zap.Error(err), zap.String("company_id", companyID.String()))
		utils.InternalServerErrorResponse(c, "Failed to create team")
		return
	}

	span.SetAttributes(
		attribute.String("team.id", team.ID.String()),
		attribute.String("team.name", team.Name),
		attribute.String("company.id", companyID.String()),
	)

	utils.SuccessResponse(c, http.StatusCreated, "Team created successfully", team)
}

// GetTeams retrieves teams for a company
func (h *TeamHandler) GetTeams(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "TeamHandler.GetTeams")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	teams, err := h.teamRepo.GetByCompany(ctx, *companyID, limit, offset)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve teams")
		return
	}

	span.SetAttributes(
		attribute.String("company.id", companyID.String()),
		attribute.Int("teams.count", len(teams)),
	)

	utils.SuccessResponse(c, http.StatusOK, "Teams retrieved successfully", gin.H{
		"teams":  teams,
		"limit":  limit,
		"offset": offset,
		"count":  len(teams),
	})
}

// GetTeam retrieves a specific team
func (h *TeamHandler) GetTeam(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "TeamHandler.GetTeam")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	teamIDStr := c.Param("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID")
		return
	}

	team, err := h.teamRepo.GetByID(ctx, teamID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve team")
		return
	}

	if team == nil {
		utils.NotFoundResponse(c, "Team not found")
		return
	}

	// Get team members
	members, err := h.teamRepo.GetMembers(ctx, teamID)
	if err != nil {
		span.RecordError(err)
		// Don't fail the request if members retrieval fails
		members = []models.TeamMember{}
	}

	team.Members = members

	span.SetAttributes(
		attribute.String("team.id", team.ID.String()),
		attribute.String("company.id", companyID.String()),
	)

	utils.SuccessResponse(c, http.StatusOK, "Team retrieved successfully", team)
}

// UpdateTeam updates a team
func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "TeamHandler.UpdateTeam")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	teamIDStr := c.Param("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID")
		return
	}

	var req models.CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Get existing team
	team, err := h.teamRepo.GetByID(ctx, teamID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve team")
		return
	}

	if team == nil {
		utils.NotFoundResponse(c, "Team not found")
		return
	}

	// Validate manager if provided
	if req.ManagerID != nil {
		manager, err := h.userRepo.GetByID(ctx, *req.ManagerID)
		if err != nil || manager == nil {
			utils.BadRequestResponse(c, "Invalid manager ID")
			return
		}

		// Check if manager belongs to the same company
		if manager.CompanyID == nil || *manager.CompanyID != *companyID {
			utils.BadRequestResponse(c, "Manager must belong to the same company")
			return
		}
	}

	// Update team fields
	team.Name = req.Name
	team.Description = req.Description
	team.ManagerID = req.ManagerID

	err = h.teamRepo.Update(ctx, team)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to update team")
		return
	}

	span.SetAttributes(attribute.String("team.id", team.ID.String()))

	utils.SuccessResponse(c, http.StatusOK, "Team updated successfully", team)
}

// DeleteTeam deletes a team
func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "TeamHandler.DeleteTeam")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	teamIDStr := c.Param("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID")
		return
	}

	err = h.teamRepo.Delete(ctx, teamID, *companyID)
	if err != nil {
		span.RecordError(err)
		logger.Error("Failed to delete team", zap.Error(err), zap.String("team_id", teamID.String()), zap.String("company_id", companyID.String()))
		utils.InternalServerErrorResponse(c, "Failed to delete team")
		return
	}

	span.SetAttributes(attribute.String("team.id", teamID.String()))

	utils.SuccessResponse(c, http.StatusOK, "Team deleted successfully", nil)
}

// AddMember adds a user to a team
func (h *TeamHandler) AddMember(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "TeamHandler.AddMember")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	teamIDStr := c.Param("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID")
		return
	}

	var req models.AssignTeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Verify team exists and belongs to company
	team, err := h.teamRepo.GetByID(ctx, teamID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve team")
		return
	}

	if team == nil {
		utils.NotFoundResponse(c, "Team not found")
		return
	}

	// Verify user exists and belongs to the same company
	user, err := h.userRepo.GetByID(ctx, req.UserID)
	if err != nil || user == nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}

	if user.CompanyID == nil || *user.CompanyID != *companyID {
		utils.BadRequestResponse(c, "User must belong to the same company")
		return
	}

	// Check if user is already a member
	exists, err := h.teamRepo.CheckMemberExists(ctx, teamID, req.UserID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to check member existence")
		return
	}

	if exists {
		utils.ConflictResponse(c, "User is already a member of this team")
		return
	}

	teamMember := &models.TeamMember{
		TeamID:     teamID,
		UserID:     req.UserID,
		RoleInTeam: req.RoleInTeam,
	}

	err = h.teamRepo.AddMember(ctx, teamMember)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to add team member")
		return
	}

	span.SetAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.String("user.id", req.UserID.String()),
		attribute.String("role", req.RoleInTeam),
	)

	utils.SuccessResponse(c, http.StatusCreated, "Team member added successfully", teamMember)
}

// RemoveMember removes a user from a team
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "TeamHandler.RemoveMember")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	teamIDStr := c.Param("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID")
		return
	}

	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}

	// Verify team exists and belongs to company
	team, err := h.teamRepo.GetByID(ctx, teamID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve team")
		return
	}

	if team == nil {
		utils.NotFoundResponse(c, "Team not found")
		return
	}

	err = h.teamRepo.RemoveMember(ctx, teamID, userID)
	if err != nil {
		span.RecordError(err)
		logger.Error("Failed to remove team member", zap.Error(err), zap.String("team_id", teamID.String()), zap.String("user_id", userID.String()))
		utils.InternalServerErrorResponse(c, "Failed to remove team member")
		return
	}

	span.SetAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.String("user.id", userID.String()),
	)

	utils.SuccessResponse(c, http.StatusOK, "Team member removed successfully", nil)
}

// GetMembers retrieves all members of a team
func (h *TeamHandler) GetMembers(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "TeamHandler.GetMembers")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	teamIDStr := c.Param("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID")
		return
	}

	// Verify team exists and belongs to company
	team, err := h.teamRepo.GetByID(ctx, teamID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve team")
		return
	}

	if team == nil {
		utils.NotFoundResponse(c, "Team not found")
		return
	}

	members, err := h.teamRepo.GetMembers(ctx, teamID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve team members")
		return
	}

	span.SetAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.Int("members.count", len(members)),
	)

	utils.SuccessResponse(c, http.StatusOK, "Team members retrieved successfully", gin.H{
		"team":    team,
		"members": members,
		"count":   len(members),
	})
}

// UpdateMemberRole updates a team member's role
func (h *TeamHandler) UpdateMemberRole(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "TeamHandler.UpdateMemberRole")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	teamIDStr := c.Param("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID")
		return
	}

	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid user ID")
		return
	}

	var req struct {
		RoleInTeam string `json:"role_in_team" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Verify team exists and belongs to company
	team, err := h.teamRepo.GetByID(ctx, teamID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve team")
		return
	}

	if team == nil {
		utils.NotFoundResponse(c, "Team not found")
		return
	}

	// Update member role
	err = h.teamRepo.UpdateMemberRole(ctx, teamID, userID, req.RoleInTeam)
	if err != nil {
		span.RecordError(err)
		logger.Error("Failed to update member role", zap.Error(err), zap.String("team_id", teamID.String()), zap.String("user_id", userID.String()), zap.String("new_role", req.RoleInTeam))
		utils.InternalServerErrorResponse(c, "Failed to update member role")
		return
	}

	span.SetAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.String("user.id", userID.String()),
		attribute.String("new_role", req.RoleInTeam),
	)

	utils.SuccessResponse(c, http.StatusOK, "Member role updated successfully", gin.H{
		"team_id":      teamID,
		"user_id":      userID,
		"role_in_team": req.RoleInTeam,
	})
}

// GetTeamStats retrieves team statistics
func (h *TeamHandler) GetTeamStats(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "TeamHandler.GetTeamStats")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	teamIDStr := c.Param("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID")
		return
	}

	// Verify team exists and belongs to company
	team, err := h.teamRepo.GetByID(ctx, teamID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve team")
		return
	}

	if team == nil {
		utils.NotFoundResponse(c, "Team not found")
		return
	}

	// Get team members count
	members, err := h.teamRepo.GetMembers(ctx, teamID)
	if err != nil {
		span.RecordError(err)
		// Don't fail the request
		members = []models.TeamMember{}
	}

	// Get vehicles count
	vehicles, err := h.vehicleRepo.GetByTeam(ctx, teamID, *companyID)
	if err != nil {
		span.RecordError(err)
		// Don't fail the request
		vehicles = []models.Vehicle{}
	}

	// Count active vehicles
	activeVehicles := 0
	for _, v := range vehicles {
		if v.Status == "active" {
			activeVehicles++
		}
	}

	stats := gin.H{
		"team_id":         teamID,
		"team_name":       team.Name,
		"member_count":    len(members),
		"vehicle_count":   len(vehicles),
		"active_vehicles": activeVehicles,
		"status":          team.Status,
		"created_at":      team.CreatedAt,
		"manager_id":      team.ManagerID,
	}

	span.SetAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.Int("members.count", len(members)),
		attribute.Int("vehicles.count", len(vehicles)),
		attribute.Int("vehicles.active", activeVehicles),
	)

	utils.SuccessResponse(c, http.StatusOK, "Team statistics retrieved successfully", stats)
}

// GetTeamVehicles retrieves vehicles assigned to a team
func (h *TeamHandler) GetTeamVehicles(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "TeamHandler.GetTeamVehicles")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	teamIDStr := c.Param("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID")
		return
	}

	// Verify team exists and belongs to company
	team, err := h.teamRepo.GetByID(ctx, teamID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve team")
		return
	}

	if team == nil {
		utils.NotFoundResponse(c, "Team not found")
		return
	}

	// Get vehicles for this team
	vehicles, err := h.vehicleRepo.GetByTeam(ctx, teamID, *companyID)
	if err != nil {
		span.RecordError(err)
		logger.Error("Failed to retrieve team vehicles", zap.Error(err), zap.String("team_id", teamID.String()), zap.String("company_id", companyID.String()))
		utils.InternalServerErrorResponse(c, "Failed to retrieve team vehicles")
		return
	}

	span.SetAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.Int("vehicle.count", len(vehicles)),
	)

	utils.SuccessResponse(c, http.StatusOK, "Team vehicles retrieved successfully", gin.H{
		"team":     team,
		"vehicles": vehicles,
		"count":    len(vehicles),
	})
}

// GetMyTeams retrieves teams for the current user
func (h *TeamHandler) GetMyTeams(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "TeamHandler.GetMyTeams")
	defer span.End()

	// Get user ID from context
	userID, err := middleware.GetUserIDFromContext(c)
	if err != nil || userID == nil {
		utils.BadRequestResponse(c, "User context required")
		return
	}

	teams, err := h.teamRepo.GetTeamsByUser(ctx, *userID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve user teams")
		return
	}

	span.SetAttributes(
		attribute.String("user.id", userID.String()),
		attribute.Int("teams.count", len(teams)),
	)

	utils.SuccessResponse(c, http.StatusOK, "User teams retrieved successfully", gin.H{
		"teams": teams,
		"count": len(teams),
	})
}

// AssignVehicleToTeam assigns a vehicle to a team
func (h *TeamHandler) AssignVehicleToTeam(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "TeamHandler.AssignVehicleToTeam")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	teamIDStr := c.Param("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID")
		return
	}

	vehicleIDStr := c.Param("vehicleId")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid vehicle ID")
		return
	}

	// Verify team exists and belongs to company
	team, err := h.teamRepo.GetByID(ctx, teamID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve team")
		return
	}

	if team == nil {
		utils.NotFoundResponse(c, "Team not found")
		return
	}

	// Verify vehicle exists and belongs to company
	vehicle, err := h.vehicleRepo.GetByID(ctx, vehicleID, *companyID)
	if err != nil {
		span.RecordError(err)
		logger.Error("Failed to retrieve vehicle for assignment", zap.Error(err), zap.String("vehicle_id", vehicleID.String()), zap.String("company_id", companyID.String()))
		utils.InternalServerErrorResponse(c, "Failed to retrieve vehicle")
		return
	}

	if vehicle == nil {
		utils.NotFoundResponse(c, "Vehicle not found")
		return
	}

	// Update vehicle assignment
	err = h.vehicleRepo.UpdateAssignment(ctx, vehicleID, *companyID, vehicle.DriverID, vehicle.HelperID, &teamID)
	if err != nil {
		span.RecordError(err)
		logger.Error("Failed to assign vehicle to team", zap.Error(err), zap.String("vehicle_id", vehicleID.String()), zap.String("team_id", teamID.String()))
		utils.InternalServerErrorResponse(c, "Failed to assign vehicle to team")
		return
	}

	span.SetAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.String("vehicle.id", vehicleID.String()),
	)

	utils.SuccessResponse(c, http.StatusOK, "Vehicle assigned to team successfully", gin.H{
		"team_id":    teamID,
		"vehicle_id": vehicleID,
	})
}

// UnassignVehicleFromTeam removes a vehicle from a team
func (h *TeamHandler) UnassignVehicleFromTeam(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "TeamHandler.UnassignVehicleFromTeam")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	teamIDStr := c.Param("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid team ID")
		return
	}

	vehicleIDStr := c.Param("vehicleId")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid vehicle ID")
		return
	}

	// Verify team exists and belongs to company
	team, err := h.teamRepo.GetByID(ctx, teamID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve team")
		return
	}

	if team == nil {
		utils.NotFoundResponse(c, "Team not found")
		return
	}

	// Verify vehicle exists and belongs to company
	vehicle, err := h.vehicleRepo.GetByID(ctx, vehicleID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve vehicle")
		return
	}

	if vehicle == nil {
		utils.NotFoundResponse(c, "Vehicle not found")
		return
	}

	// Check if vehicle is assigned to this team
	if vehicle.TeamID == nil || *vehicle.TeamID != teamID {
		utils.BadRequestResponse(c, "Vehicle is not assigned to this team")
		return
	}

	// Update vehicle assignment (set team_id to nil)
	err = h.vehicleRepo.UpdateAssignment(ctx, vehicleID, *companyID, vehicle.DriverID, vehicle.HelperID, nil)
	if err != nil {
		span.RecordError(err)
		logger.Error("Failed to unassign vehicle from team", zap.Error(err), zap.String("vehicle_id", vehicleID.String()), zap.String("team_id", teamID.String()))
		utils.InternalServerErrorResponse(c, "Failed to unassign vehicle from team")
		return
	}

	span.SetAttributes(
		attribute.String("team.id", teamID.String()),
		attribute.String("vehicle.id", vehicleID.String()),
	)

	utils.SuccessResponse(c, http.StatusOK, "Vehicle unassigned from team successfully", gin.H{
		"team_id":    teamID,
		"vehicle_id": vehicleID,
	})
}
