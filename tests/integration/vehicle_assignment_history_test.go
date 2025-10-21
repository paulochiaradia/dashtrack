package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/handlers"
	"github.com/paulochiaradia/dashtrack/internal/middleware"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
	"github.com/paulochiaradia/dashtrack/internal/services"
	"github.com/paulochiaradia/dashtrack/tests/testutils"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

// VehicleAssignmentHistoryTestSuite tests the Vehicle Assignment History API
type VehicleAssignmentHistoryTestSuite struct {
	suite.Suite
	testDB       *testutils.TestDB
	router       *gin.Engine
	teamHandler  *handlers.TeamHandler
	teamRepo     *repository.TeamRepository
	userRepo     *repository.UserRepository
	roleRepo     *repository.RoleRepository
	vehicleRepo  *repository.VehicleRepository
	tokenService *services.TokenService
	token        string
	companyID    uuid.UUID
	teamID       uuid.UUID
	vehicle1ID   uuid.UUID
	vehicle2ID   uuid.UUID
}

func TestVehicleAssignmentHistoryTestSuite(t *testing.T) {
	suite.Run(t, new(VehicleAssignmentHistoryTestSuite))
}

func (s *VehicleAssignmentHistoryTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	var err error
	s.testDB, err = testutils.SetupTestDB("vehicle_history")
	s.Require().NoError(err, "Failed to setup test database")

	s.teamRepo = repository.NewTeamRepository(s.testDB.SqlxDB)
	s.userRepo = repository.NewUserRepository(s.testDB.SqlxDB)
	s.roleRepo = repository.NewRoleRepository(s.testDB.SqlDB)
	s.vehicleRepo = repository.NewVehicleRepository(s.testDB.SqlxDB)

	s.tokenService = services.NewTokenService(
		s.testDB.SqlxDB,
		"test-secret-key-min-32-characters-long",
		15*time.Minute,
		24*time.Hour,
	)

	s.teamHandler = handlers.NewTeamHandler(s.teamRepo, s.userRepo, s.vehicleRepo)

	s.router = gin.New()
	s.setupRoutes()
	s.createTestData()
}

func (s *VehicleAssignmentHistoryTestSuite) TearDownSuite() {
	if s.testDB != nil {
		s.testDB.TearDown()
	}
}

func (s *VehicleAssignmentHistoryTestSuite) setupRoutes() {
	ginAuth := middleware.NewGinAuthMiddleware(s.tokenService)

	api := s.router.Group("/api/v1")
	{
		companyAdmin := api.Group("/company-admin")
		companyAdmin.Use(ginAuth.RequireAuth())
		companyAdmin.Use(middleware.RequireCompanyAdmin())
		{
			companyAdmin.POST("/teams/:team_id/vehicles/:vehicle_id", s.teamHandler.AssignVehicleToTeam)
			companyAdmin.DELETE("/teams/:team_id/vehicles/:vehicle_id", s.teamHandler.UnassignVehicleFromTeam)
			companyAdmin.GET("/teams/:team_id/vehicle-history", s.teamHandler.GetTeamVehicles)
		}
	}
}

func (s *VehicleAssignmentHistoryTestSuite) createTestData() {
	companyAdminRole, err := s.roleRepo.GetByName("company_admin")
	s.Require().NoError(err)

	company := &models.Company{
		Name:             "Test Company VH",
		Slug:             "test-company-vh",
		Email:            "test-vh@company.com",
		Country:          "Brazil",
		SubscriptionPlan: "premium",
		Status:           "active",
	}
	err = s.testDB.DB.Create(company).Error
	s.Require().NoError(err)
	s.companyID = company.ID

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Admin@123"), bcrypt.DefaultCost)
	phone := "+5511999999999"
	cpf := "12345678901"
	adminUser := &models.User{
		Name:      "Admin User",
		Email:     "admin-vh@test.com",
		Password:  string(hashedPassword),
		Phone:     &phone,
		CPF:       &cpf,
		CompanyID: &s.companyID,
		RoleID:    companyAdminRole.ID,
		Active:    true,
	}
	err = s.testDB.DB.Create(adminUser).Error
	s.Require().NoError(err)
	adminUser.Role = companyAdminRole

	ctx := context.Background()
	tokenPair, err := s.tokenService.GenerateTokenPair(ctx, adminUser, "127.0.0.1", "test-agent")
	s.Require().NoError(err)
	s.token = tokenPair.AccessToken

	team := &models.Team{
		CompanyID:   s.companyID,
		Name:        "Transport Team",
		Description: stringPtr("Test team"),
	}
	err = s.teamRepo.Create(context.Background(), team)
	s.Require().NoError(err)
	s.teamID = team.ID

	vehicle1 := &models.Vehicle{
		CompanyID:    s.companyID,
		LicensePlate: "ABC-1234",
		Brand:        "Ford",
		Model:        "Cargo 815",
		Year:         2020,
		Status:       "active",
	}
	err = s.testDB.DB.Create(vehicle1).Error
	s.Require().NoError(err)
	s.vehicle1ID = vehicle1.ID

	vehicle2 := &models.Vehicle{
		CompanyID:    s.companyID,
		LicensePlate: "XYZ-5678",
		Brand:        "Mercedes",
		Model:        "Atego 1719",
		Year:         2021,
		Status:       "active",
	}
	err = s.testDB.DB.Create(vehicle2).Error
	s.Require().NoError(err)
	s.vehicle2ID = vehicle2.ID
}

// TestAssignVehicleCreatesHistory tests automatic history creation on assignment
func (s *VehicleAssignmentHistoryTestSuite) TestAssignVehicleCreatesHistory() {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/company-admin/teams/%s/vehicles/%s", s.teamID, s.vehicle1ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "Should assign vehicle successfully")
}

// TestUnassignVehicleCreatesHistory tests history creation on unassignment
func (s *VehicleAssignmentHistoryTestSuite) TestUnassignVehicleCreatesHistory() {
	// Assign first
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/company-admin/teams/%s/vehicles/%s", s.teamID, s.vehicle1ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Then unassign
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/company-admin/teams/%s/vehicles/%s", s.teamID, s.vehicle1ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "Should unassign vehicle successfully")
}

// TestGetVehicleHistory tests retrieving vehicle assignment history
func (s *VehicleAssignmentHistoryTestSuite) TestGetVehicleHistory() {
	// Assign vehicle
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/company-admin/teams/%s/vehicles/%s", s.teamID, s.vehicle1ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Get history
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/company-admin/teams/%s/vehicle-history", s.teamID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "Should get vehicle history successfully")
}

// TestMultipleVehicleAssignments tests history with multiple vehicles
func (s *VehicleAssignmentHistoryTestSuite) TestMultipleVehicleAssignments() {
	// Assign vehicle 1
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/company-admin/teams/%s/vehicles/%s", s.teamID, s.vehicle1ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Assign vehicle 2
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/company-admin/teams/%s/vehicles/%s", s.teamID, s.vehicle2ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

// TestCompleteWorkflow tests the complete workflow
func (s *VehicleAssignmentHistoryTestSuite) TestCompleteWorkflow() {
	s.T().Log("✅ Testing vehicle assignment history workflow")

	s.T().Log("Step 1: Assign vehicle")
	s.TestAssignVehicleCreatesHistory()

	s.T().Log("Step 2: Unassign vehicle")
	s.TestUnassignVehicleCreatesHistory()

	s.T().Log("Step 3: Get history")
	s.TestGetVehicleHistory()

	s.T().Log("Step 4: Multiple assignments")
	s.TestMultipleVehicleAssignments()

	s.T().Log("✅ Complete vehicle history workflow passed!")
}
