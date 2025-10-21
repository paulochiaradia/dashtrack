package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/paulochiaradia/dashtrack/internal/handlers"
	"github.com/paulochiaradia/dashtrack/internal/models"
)

// ============================================================================
// Mock Repositories
// ============================================================================

type MockTeamRepository struct {
	mock.Mock
}

func (m *MockTeamRepository) Create(ctx context.Context, team *models.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *MockTeamRepository) GetByID(ctx context.Context, id uuid.UUID, companyID uuid.UUID) (*models.Team, error) {
	args := m.Called(ctx, id, companyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Team), args.Error(1)
}

func (m *MockTeamRepository) GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]models.Team, error) {
	args := m.Called(ctx, companyID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Team), args.Error(1)
}

func (m *MockTeamRepository) Update(ctx context.Context, team *models.Team) error {
	args := m.Called(ctx, team)
	return args.Error(0)
}

func (m *MockTeamRepository) Delete(ctx context.Context, id uuid.UUID, companyID uuid.UUID) error {
	args := m.Called(ctx, id, companyID)
	return args.Error(0)
}

func (m *MockTeamRepository) AddMember(ctx context.Context, member *models.TeamMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockTeamRepository) RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error {
	args := m.Called(ctx, teamID, userID)
	return args.Error(0)
}

func (m *MockTeamRepository) GetMembers(ctx context.Context, teamID uuid.UUID) ([]models.TeamMember, error) {
	args := m.Called(ctx, teamID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamMember), args.Error(1)
}

func (m *MockTeamRepository) UpdateMemberRole(ctx context.Context, teamID, userID uuid.UUID, roleInTeam string) error {
	args := m.Called(ctx, teamID, userID, roleInTeam)
	return args.Error(0)
}

func (m *MockTeamRepository) GetTeamsByUser(ctx context.Context, userID uuid.UUID) ([]models.Team, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Team), args.Error(1)
}

func (m *MockTeamRepository) CheckMemberExists(ctx context.Context, teamID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, teamID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockTeamRepository) LogMemberChange(ctx context.Context, history *models.TeamMemberHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockTeamRepository) GetMemberHistory(ctx context.Context, teamID, companyID uuid.UUID, limit int) ([]models.TeamMemberHistory, error) {
	args := m.Called(ctx, teamID, companyID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamMemberHistory), args.Error(1)
}

func (m *MockTeamRepository) GetUserTeamHistory(ctx context.Context, userID, companyID uuid.UUID, limit int) ([]models.TeamMemberHistory, error) {
	args := m.Called(ctx, userID, companyID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamMemberHistory), args.Error(1)
}

func (m *MockTeamRepository) GetMemberHistoryWithDetails(ctx context.Context, teamID, companyID uuid.UUID, limit int) ([]models.TeamMemberHistory, error) {
	args := m.Called(ctx, teamID, companyID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamMemberHistory), args.Error(1)
}

func (m *MockTeamRepository) GetUserTeamHistoryWithDetails(ctx context.Context, userID, companyID uuid.UUID, limit int) ([]models.TeamMemberHistory, error) {
	args := m.Called(ctx, userID, companyID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.TeamMemberHistory), args.Error(1)
}

type MockVehicleRepository struct {
	mock.Mock
}

func (m *MockVehicleRepository) GetByID(ctx context.Context, id uuid.UUID, companyID uuid.UUID) (*models.Vehicle, error) {
	args := m.Called(ctx, id, companyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Vehicle), args.Error(1)
}

func (m *MockVehicleRepository) GetByTeam(ctx context.Context, teamID uuid.UUID, companyID uuid.UUID) ([]models.Vehicle, error) {
	args := m.Called(ctx, teamID, companyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Vehicle), args.Error(1)
}

func (m *MockVehicleRepository) UpdateAssignment(ctx context.Context, vehicleID, companyID uuid.UUID, driverID, helperID, teamID *uuid.UUID) error {
	args := m.Called(ctx, vehicleID, companyID, driverID, helperID, teamID)
	return args.Error(0)
}

func (m *MockVehicleRepository) Create(ctx context.Context, vehicle *models.Vehicle) error {
	args := m.Called(ctx, vehicle)
	return args.Error(0)
}

func (m *MockVehicleRepository) GetByLicensePlate(ctx context.Context, licensePlate string, companyID uuid.UUID) (*models.Vehicle, error) {
	args := m.Called(ctx, licensePlate, companyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Vehicle), args.Error(1)
}

func (m *MockVehicleRepository) GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]models.Vehicle, error) {
	args := m.Called(ctx, companyID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Vehicle), args.Error(1)
}

func (m *MockVehicleRepository) GetByDriver(ctx context.Context, driverID uuid.UUID, companyID uuid.UUID) ([]models.Vehicle, error) {
	args := m.Called(ctx, driverID, companyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Vehicle), args.Error(1)
}

func (m *MockVehicleRepository) Update(ctx context.Context, vehicle *models.Vehicle) error {
	args := m.Called(ctx, vehicle)
	return args.Error(0)
}

func (m *MockVehicleRepository) Delete(ctx context.Context, id uuid.UUID, companyID uuid.UUID) error {
	args := m.Called(ctx, id, companyID)
	return args.Error(0)
}

func (m *MockVehicleRepository) GetVehicleDashboardData(ctx context.Context, vehicleID, companyID uuid.UUID) (*models.VehicleDashboardData, error) {
	args := m.Called(ctx, vehicleID, companyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.VehicleDashboardData), args.Error(1)
}

func (m *MockVehicleRepository) GetActiveTrip(ctx context.Context, vehicleID uuid.UUID) (*models.VehicleTrip, error) {
	args := m.Called(ctx, vehicleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.VehicleTrip), args.Error(1)
}

func (m *MockVehicleRepository) Search(ctx context.Context, companyID uuid.UUID, searchTerm string, limit, offset int) ([]models.Vehicle, error) {
	args := m.Called(ctx, companyID, searchTerm, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Vehicle), args.Error(1)
}

func (m *MockVehicleRepository) CheckLicensePlateExists(ctx context.Context, licensePlate string, companyID uuid.UUID, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, licensePlate, companyID, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockVehicleRepository) LogAssignmentChange(ctx context.Context, history *models.VehicleAssignmentHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockVehicleRepository) GetAssignmentHistory(ctx context.Context, vehicleID, companyID uuid.UUID, limit int) ([]models.VehicleAssignmentHistory, error) {
	args := m.Called(ctx, vehicleID, companyID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.VehicleAssignmentHistory), args.Error(1)
}

func (m *MockVehicleRepository) GetAssignmentHistoryWithDetails(ctx context.Context, vehicleID, companyID uuid.UUID, limit int) ([]models.VehicleAssignmentHistory, error) {
	args := m.Called(ctx, vehicleID, companyID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.VehicleAssignmentHistory), args.Error(1)
}

type MockUserRepositoryForTeam struct {
	mock.Mock
}

func (m *MockUserRepositoryForTeam) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryForTeam) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepositoryForTeam) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryForTeam) GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, companyID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepositoryForTeam) Update(ctx context.Context, id uuid.UUID, updateReq models.UpdateUserRequest) (*models.User, error) {
	args := m.Called(ctx, id, updateReq)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryForTeam) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	args := m.Called(ctx, id, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepositoryForTeam) UpdateCompany(ctx context.Context, userID, companyID uuid.UUID) error {
	args := m.Called(ctx, userID, companyID)
	return args.Error(0)
}

func (m *MockUserRepositoryForTeam) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepositoryForTeam) List(ctx context.Context, limit, offset int, active *bool, roleID *uuid.UUID) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset, active, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepositoryForTeam) ListByCompanyAndRoles(ctx context.Context, companyID *uuid.UUID, roles []string, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, companyID, roles, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepositoryForTeam) ListByRoles(ctx context.Context, roles []string, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, roles, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepositoryForTeam) CountByCompanyAndRoles(ctx context.Context, companyID *uuid.UUID, roles []string) (int, error) {
	args := m.Called(ctx, companyID, roles)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepositoryForTeam) UpdateLoginAttempts(ctx context.Context, id uuid.UUID, attempts int, blockedUntil *time.Time) error {
	args := m.Called(ctx, id, attempts, blockedUntil)
	return args.Error(0)
}

func (m *MockUserRepositoryForTeam) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepositoryForTeam) GetUserContext(ctx context.Context, userID uuid.UUID) (*models.UserContext, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserContext), args.Error(1)
}

func (m *MockUserRepositoryForTeam) Search(ctx context.Context, companyID *uuid.UUID, searchTerm string, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, companyID, searchTerm, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepositoryForTeam) CountUsers(ctx context.Context, companyID *uuid.UUID) (int, error) {
	args := m.Called(ctx, companyID)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepositoryForTeam) CountActiveUsers(ctx context.Context, companyID *uuid.UUID) (int, error) {
	args := m.Called(ctx, companyID)
	return args.Int(0), args.Error(1)
}

// ============================================================================
// Helper Functions
// ============================================================================

func setupTeamTestContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set company context
	companyID := uuid.New()
	c.Set("company_id", companyID)

	return c, w
}

func setupTeamTestContextWithUser() (*gin.Context, *httptest.ResponseRecorder, uuid.UUID) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Set company and user context
	companyID := uuid.New()
	userID := uuid.New()
	c.Set("company_id", companyID)
	c.Set("userContext", &models.UserContext{
		UserID:    userID,
		CompanyID: &companyID,
		Role:      "company_admin",
	})

	return c, w, userID
}

// ============================================================================
// TEST: Get Team Statistics
// ============================================================================

func TestGetTeamStats(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepositoryForTeam)
	mockVehicleRepo := new(MockVehicleRepository)

	handler := handlers.NewTeamHandler(mockTeamRepo, mockUserRepo, mockVehicleRepo)

	teamID := uuid.New()
	companyID := uuid.New()

	team := &models.Team{
		ID:        teamID,
		CompanyID: companyID,
		Name:      "Test Team",
		Status:    "active",
	}

	members := []models.TeamMember{
		{ID: uuid.New(), TeamID: teamID, UserID: uuid.New(), RoleInTeam: "driver"},
		{ID: uuid.New(), TeamID: teamID, UserID: uuid.New(), RoleInTeam: "helper"},
	}

	vehicles := []models.Vehicle{
		{ID: uuid.New(), CompanyID: companyID, TeamID: &teamID, Status: "active"},
		{ID: uuid.New(), CompanyID: companyID, TeamID: &teamID, Status: "active"},
		{ID: uuid.New(), CompanyID: companyID, TeamID: &teamID, Status: "inactive"},
	}

	mockTeamRepo.On("GetByID", mock.Anything, teamID, companyID).Return(team, nil)
	mockTeamRepo.On("GetMembers", mock.Anything, teamID).Return(members, nil)
	mockVehicleRepo.On("GetByTeam", mock.Anything, teamID, companyID).Return(vehicles, nil)

	c, w := setupTeamTestContext()
	c.Set("company_id", companyID)
	c.Params = gin.Params{{Key: "id", Value: teamID.String()}}
	c.Request = httptest.NewRequest("GET", "/teams/"+teamID.String()+"/stats", nil)

	handler.GetTeamStats(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, float64(2), data["member_count"])
	assert.Equal(t, float64(3), data["vehicle_count"])
	assert.Equal(t, float64(2), data["active_vehicles"])

	mockTeamRepo.AssertExpectations(t)
	mockVehicleRepo.AssertExpectations(t)
}

// ============================================================================
// TEST: Get Team Vehicles
// ============================================================================

func TestGetTeamVehicles(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepositoryForTeam)
	mockVehicleRepo := new(MockVehicleRepository)

	handler := handlers.NewTeamHandler(mockTeamRepo, mockUserRepo, mockVehicleRepo)

	teamID := uuid.New()
	companyID := uuid.New()

	team := &models.Team{
		ID:        teamID,
		CompanyID: companyID,
		Name:      "Test Team",
		Status:    "active",
	}

	vehicles := []models.Vehicle{
		{
			ID:           uuid.New(),
			CompanyID:    companyID,
			TeamID:       &teamID,
			LicensePlate: "ABC-1234",
			Status:       "active",
		},
		{
			ID:           uuid.New(),
			CompanyID:    companyID,
			TeamID:       &teamID,
			LicensePlate: "XYZ-5678",
			Status:       "active",
		},
	}

	mockTeamRepo.On("GetByID", mock.Anything, teamID, companyID).Return(team, nil)
	mockVehicleRepo.On("GetByTeam", mock.Anything, teamID, companyID).Return(vehicles, nil)

	c, w := setupTeamTestContext()
	c.Set("company_id", companyID)
	c.Params = gin.Params{{Key: "id", Value: teamID.String()}}
	c.Request = httptest.NewRequest("GET", "/teams/"+teamID.String()+"/vehicles", nil)

	handler.GetTeamVehicles(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	vehiclesList := data["vehicles"].([]interface{})
	assert.Equal(t, 2, len(vehiclesList))
	assert.Equal(t, float64(2), data["count"])

	mockTeamRepo.AssertExpectations(t)
	mockVehicleRepo.AssertExpectations(t)
}

// ============================================================================
// TEST: Assign Vehicle to Team
// ============================================================================

func TestAssignVehicleToTeam(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepositoryForTeam)
	mockVehicleRepo := new(MockVehicleRepository)

	handler := handlers.NewTeamHandler(mockTeamRepo, mockUserRepo, mockVehicleRepo)

	teamID := uuid.New()
	vehicleID := uuid.New()
	companyID := uuid.New()

	team := &models.Team{
		ID:        teamID,
		CompanyID: companyID,
		Name:      "Test Team",
		Status:    "active",
	}

	vehicle := &models.Vehicle{
		ID:           vehicleID,
		CompanyID:    companyID,
		LicensePlate: "ABC-1234",
		Status:       "active",
	}

	mockTeamRepo.On("GetByID", mock.Anything, teamID, companyID).Return(team, nil)
	mockVehicleRepo.On("GetByID", mock.Anything, vehicleID, companyID).Return(vehicle, nil)
	mockVehicleRepo.On("UpdateAssignment", mock.Anything, vehicleID, companyID, vehicle.DriverID, vehicle.HelperID, &teamID).Return(nil)

	c, w := setupTeamTestContext()
	c.Set("company_id", companyID)
	c.Params = gin.Params{
		{Key: "id", Value: teamID.String()},
		{Key: "vehicleId", Value: vehicleID.String()},
	}
	c.Request = httptest.NewRequest("POST", "/teams/"+teamID.String()+"/vehicles/"+vehicleID.String(), nil)

	handler.AssignVehicleToTeam(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, teamID.String(), data["team_id"])
	assert.Equal(t, vehicleID.String(), data["vehicle_id"])

	mockTeamRepo.AssertExpectations(t)
	mockVehicleRepo.AssertExpectations(t)
}

// ============================================================================
// TEST: Unassign Vehicle from Team
// ============================================================================

func TestUnassignVehicleFromTeam(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepositoryForTeam)
	mockVehicleRepo := new(MockVehicleRepository)

	handler := handlers.NewTeamHandler(mockTeamRepo, mockUserRepo, mockVehicleRepo)

	teamID := uuid.New()
	vehicleID := uuid.New()
	companyID := uuid.New()

	team := &models.Team{
		ID:        teamID,
		CompanyID: companyID,
		Name:      "Test Team",
		Status:    "active",
	}

	vehicle := &models.Vehicle{
		ID:           vehicleID,
		CompanyID:    companyID,
		TeamID:       &teamID,
		LicensePlate: "ABC-1234",
		Status:       "active",
	}

	mockTeamRepo.On("GetByID", mock.Anything, teamID, companyID).Return(team, nil)
	mockVehicleRepo.On("GetByID", mock.Anything, vehicleID, companyID).Return(vehicle, nil)
	mockVehicleRepo.On("UpdateAssignment", mock.Anything, vehicleID, companyID, vehicle.DriverID, vehicle.HelperID, (*uuid.UUID)(nil)).Return(nil)

	c, w := setupTeamTestContext()
	c.Set("company_id", companyID)
	c.Params = gin.Params{
		{Key: "id", Value: teamID.String()},
		{Key: "vehicleId", Value: vehicleID.String()},
	}
	c.Request = httptest.NewRequest("DELETE", "/teams/"+teamID.String()+"/vehicles/"+vehicleID.String(), nil)

	handler.UnassignVehicleFromTeam(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	assert.Equal(t, teamID.String(), data["team_id"])
	assert.Equal(t, vehicleID.String(), data["vehicle_id"])

	mockTeamRepo.AssertExpectations(t)
	mockVehicleRepo.AssertExpectations(t)
}

// ============================================================================
// TEST: Get My Teams
// ============================================================================

func TestGetMyTeams(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepositoryForTeam)
	mockVehicleRepo := new(MockVehicleRepository)

	handler := handlers.NewTeamHandler(mockTeamRepo, mockUserRepo, mockVehicleRepo)

	userID := uuid.New()
	teams := []models.Team{
		{ID: uuid.New(), Name: "Team 1", Status: "active"},
		{ID: uuid.New(), Name: "Team 2", Status: "active"},
	}

	mockTeamRepo.On("GetTeamsByUser", mock.Anything, userID).Return(teams, nil)

	c, w, _ := setupTeamTestContextWithUser()
	userContext := &models.UserContext{
		UserID: userID,
		Role:   "user",
	}
	c.Set("userContext", userContext)
	c.Request = httptest.NewRequest("GET", "/teams/my-teams", nil)

	handler.GetMyTeams(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	data := response["data"].(map[string]interface{})
	teamsList := data["teams"].([]interface{})
	assert.Equal(t, 2, len(teamsList))
	assert.Equal(t, float64(2), data["count"])

	mockTeamRepo.AssertExpectations(t)
}

// ============================================================================
// TEST: Update Member Role
// ============================================================================

func TestUpdateMemberRole(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepositoryForTeam)
	mockVehicleRepo := new(MockVehicleRepository)

	handler := handlers.NewTeamHandler(mockTeamRepo, mockUserRepo, mockVehicleRepo)

	teamID := uuid.New()
	userID := uuid.New()
	companyID := uuid.New()

	team := &models.Team{
		ID:        teamID,
		CompanyID: companyID,
		Name:      "Test Team",
		Status:    "active",
	}

	user := &models.User{
		ID:    userID,
		Name:  "Test User",
		Email: "test@example.com",
	}

	mockTeamRepo.On("GetByID", mock.Anything, teamID, companyID).Return(team, nil)
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	mockTeamRepo.On("CheckMemberExists", mock.Anything, teamID, userID).Return(true, nil)
	mockTeamRepo.On("UpdateMemberRole", mock.Anything, teamID, userID, "manager").Return(nil)

	updateReq := models.UpdateMemberRoleRequest{
		NewRoleInTeam: "manager",
	}
	body, _ := json.Marshal(updateReq)

	c, w := setupTeamTestContext()
	c.Set("company_id", companyID)
	c.Params = gin.Params{
		{Key: "id", Value: teamID.String()},
		{Key: "userId", Value: userID.String()},
	}
	c.Request = httptest.NewRequest("PUT", "/teams/"+teamID.String()+"/members/"+userID.String()+"/role", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.UpdateMemberRole(c)

	assert.Equal(t, http.StatusOK, w.Code)

	mockTeamRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// ============================================================================
// TEST: Error Cases
// ============================================================================

func TestGetTeamStats_TeamNotFound(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepositoryForTeam)
	mockVehicleRepo := new(MockVehicleRepository)

	handler := handlers.NewTeamHandler(mockTeamRepo, mockUserRepo, mockVehicleRepo)

	teamID := uuid.New()
	companyID := uuid.New()

	mockTeamRepo.On("GetByID", mock.Anything, teamID, companyID).Return((*models.Team)(nil), nil)

	c, w := setupTeamTestContext()
	c.Set("company_id", companyID)
	c.Params = gin.Params{{Key: "id", Value: teamID.String()}}
	c.Request = httptest.NewRequest("GET", "/teams/"+teamID.String()+"/stats", nil)

	handler.GetTeamStats(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockTeamRepo.AssertExpectations(t)
}

func TestAssignVehicle_VehicleNotFound(t *testing.T) {
	mockTeamRepo := new(MockTeamRepository)
	mockUserRepo := new(MockUserRepositoryForTeam)
	mockVehicleRepo := new(MockVehicleRepository)

	handler := handlers.NewTeamHandler(mockTeamRepo, mockUserRepo, mockVehicleRepo)

	teamID := uuid.New()
	vehicleID := uuid.New()
	companyID := uuid.New()

	team := &models.Team{
		ID:        teamID,
		CompanyID: companyID,
		Name:      "Test Team",
		Status:    "active",
	}

	mockTeamRepo.On("GetByID", mock.Anything, teamID, companyID).Return(team, nil)
	mockVehicleRepo.On("GetByID", mock.Anything, vehicleID, companyID).Return((*models.Vehicle)(nil), nil)

	c, w := setupTeamTestContext()
	c.Set("company_id", companyID)
	c.Params = gin.Params{
		{Key: "id", Value: teamID.String()},
		{Key: "vehicleId", Value: vehicleID.String()},
	}
	c.Request = httptest.NewRequest("POST", "/teams/"+teamID.String()+"/vehicles/"+vehicleID.String(), nil)

	handler.AssignVehicleToTeam(c)

	assert.Equal(t, http.StatusNotFound, w.Code)

	mockTeamRepo.AssertExpectations(t)
	mockVehicleRepo.AssertExpectations(t)
}
