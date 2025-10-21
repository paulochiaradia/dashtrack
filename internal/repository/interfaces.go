package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/models"
)

// TeamRepositoryInterface defines the interface for team repository operations
type TeamRepositoryInterface interface {
	Create(ctx context.Context, team *models.Team) error
	GetByID(ctx context.Context, id uuid.UUID, companyID uuid.UUID) (*models.Team, error)
	GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]models.Team, error)
	Update(ctx context.Context, team *models.Team) error
	Delete(ctx context.Context, id uuid.UUID, companyID uuid.UUID) error
	AddMember(ctx context.Context, teamMember *models.TeamMember) error
	RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error
	GetMembers(ctx context.Context, teamID uuid.UUID) ([]models.TeamMember, error)
	UpdateMemberRole(ctx context.Context, teamID, userID uuid.UUID, newRole string) error
	GetTeamsByUser(ctx context.Context, userID uuid.UUID) ([]models.Team, error)
	CheckMemberExists(ctx context.Context, teamID, userID uuid.UUID) (bool, error)
	LogMemberChange(ctx context.Context, history *models.TeamMemberHistory) error
	GetMemberHistory(ctx context.Context, teamID, companyID uuid.UUID, limit int) ([]models.TeamMemberHistory, error)
	GetUserTeamHistory(ctx context.Context, userID, companyID uuid.UUID, limit int) ([]models.TeamMemberHistory, error)
	GetMemberHistoryWithDetails(ctx context.Context, teamID, companyID uuid.UUID, limit int) ([]models.TeamMemberHistory, error)
	GetUserTeamHistoryWithDetails(ctx context.Context, userID, companyID uuid.UUID, limit int) ([]models.TeamMemberHistory, error)
}

// VehicleRepositoryInterface defines the interface for vehicle repository operations
type VehicleRepositoryInterface interface {
	Create(ctx context.Context, vehicle *models.Vehicle) error
	GetByID(ctx context.Context, id uuid.UUID, companyID uuid.UUID) (*models.Vehicle, error)
	GetByLicensePlate(ctx context.Context, licensePlate string, companyID uuid.UUID) (*models.Vehicle, error)
	GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]models.Vehicle, error)
	GetByTeam(ctx context.Context, teamID uuid.UUID, companyID uuid.UUID) ([]models.Vehicle, error)
	GetByDriver(ctx context.Context, driverID uuid.UUID, companyID uuid.UUID) ([]models.Vehicle, error)
	Update(ctx context.Context, vehicle *models.Vehicle) error
	UpdateAssignment(ctx context.Context, vehicleID, companyID uuid.UUID, driverID, helperID, teamID *uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID, companyID uuid.UUID) error
	GetVehicleDashboardData(ctx context.Context, vehicleID, companyID uuid.UUID) (*models.VehicleDashboardData, error)
	GetActiveTrip(ctx context.Context, vehicleID uuid.UUID) (*models.VehicleTrip, error)
	Search(ctx context.Context, companyID uuid.UUID, searchTerm string, limit, offset int) ([]models.Vehicle, error)
	CheckLicensePlateExists(ctx context.Context, licensePlate string, companyID uuid.UUID, excludeID *uuid.UUID) (bool, error)
	LogAssignmentChange(ctx context.Context, history *models.VehicleAssignmentHistory) error
	GetAssignmentHistory(ctx context.Context, vehicleID, companyID uuid.UUID, limit int) ([]models.VehicleAssignmentHistory, error)
	GetAssignmentHistoryWithDetails(ctx context.Context, vehicleID, companyID uuid.UUID, limit int) ([]models.VehicleAssignmentHistory, error)
}
