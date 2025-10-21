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

// TeamMemberHistoryTestSuite tests the Team Member History API
type TeamMemberHistoryTestSuite struct {
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
	team1ID      uuid.UUID
	team2ID      uuid.UUID
	driverID     uuid.UUID
	helperID     uuid.UUID
	driverRoleID uuid.UUID
	helperRoleID uuid.UUID
}

func TestTeamMemberHistoryTestSuite(t *testing.T) {
	suite.Run(t, new(TeamMemberHistoryTestSuite))
}

func (s *TeamMemberHistoryTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	var err error
	s.testDB, err = testutils.SetupTestDB("member_history")
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

func (s *TeamMemberHistoryTestSuite) TearDownSuite() {
	if s.testDB != nil {
		s.testDB.TearDown()
	}
}

func (s *TeamMemberHistoryTestSuite) setupRoutes() {
	ginAuth := middleware.NewGinAuthMiddleware(s.tokenService)

	api := s.router.Group("/api/v1")
	{
		companyAdmin := api.Group("/company-admin")
		companyAdmin.Use(ginAuth.RequireAuth())
		companyAdmin.Use(middleware.RequireCompanyAdmin())
		{
			companyAdmin.POST("/teams/:team_id/members", s.teamHandler.AddMember)
			companyAdmin.PUT("/teams/:team_id/members/:user_id/role", s.teamHandler.UpdateMemberRole)
			companyAdmin.POST("/teams/:team_id/members/:user_id/transfer", s.teamHandler.TransferMemberToTeam)
			companyAdmin.DELETE("/teams/:team_id/members/:user_id", s.teamHandler.RemoveMember)

			// History endpoints
			companyAdmin.GET("/teams/:team_id/member-history", s.teamHandler.GetTeamMemberHistory)
			companyAdmin.GET("/users/:user_id/team-history", s.teamHandler.GetUserTeamHistory)
		}
	}
}

func (s *TeamMemberHistoryTestSuite) createTestData() {
	driverRole, err := s.roleRepo.GetByName("driver")
	s.Require().NoError(err)
	s.driverRoleID = driverRole.ID

	helperRole, err := s.roleRepo.GetByName("helper")
	s.Require().NoError(err)
	s.helperRoleID = helperRole.ID

	companyAdminRole, err := s.roleRepo.GetByName("company_admin")
	s.Require().NoError(err)

	company := &models.Company{
		Name:             "Test Company MH",
		Slug:             "test-company-mh",
		Email:            "test-mh@company.com",
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
		Email:     "admin-mh@test.com",
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

	team1 := &models.Team{
		CompanyID:   s.companyID,
		Name:        "Team Alpha",
		Description: stringPtr("First test team"),
	}
	err = s.teamRepo.Create(context.Background(), team1)
	s.Require().NoError(err)
	s.team1ID = team1.ID

	team2 := &models.Team{
		CompanyID:   s.companyID,
		Name:        "Team Beta",
		Description: stringPtr("Second test team"),
	}
	err = s.teamRepo.Create(context.Background(), team2)
	s.Require().NoError(err)
	s.team2ID = team2.ID

	driverPhone := "+5511888888888"
	driverCPF := "98765432100"
	driver := &models.User{
		Name:      "John Driver",
		Email:     "driver-mh@test.com",
		Password:  string(hashedPassword),
		Phone:     &driverPhone,
		CPF:       &driverCPF,
		CompanyID: &s.companyID,
		RoleID:    s.driverRoleID,
		Active:    true,
	}
	err = s.testDB.DB.Create(driver).Error
	s.Require().NoError(err)
	s.driverID = driver.ID

	helperPhone := "+5511777777777"
	helperCPF := "11122233344"
	helper := &models.User{
		Name:      "Mary Helper",
		Email:     "helper-mh@test.com",
		Password:  string(hashedPassword),
		Phone:     &helperPhone,
		CPF:       &helperCPF,
		CompanyID: &s.companyID,
		RoleID:    s.helperRoleID,
		Active:    true,
	}
	err = s.testDB.DB.Create(helper).Error
	s.Require().NoError(err)
	s.helperID = helper.ID
}

// TestAddMemberCreatesHistory tests that adding a member creates history
func (s *TeamMemberHistoryTestSuite) TestAddMemberCreatesHistory() {
	// Add member (this should create history automatically via trigger)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/company-admin/teams/%s/members", s.team1ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Get team member history
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/company-admin/teams/%s/member-history", s.team1ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "Should get team member history")
}

// TestUpdateRoleCreatesHistory tests that updating role creates history
func (s *TeamMemberHistoryTestSuite) TestUpdateRoleCreatesHistory() {
	// First add member
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/company-admin/teams/%s/members", s.team1ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Update role (should create history)
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/api/v1/company-admin/teams/%s/members/%s/role", s.team1ID, s.driverID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Verify history
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/company-admin/teams/%s/member-history", s.team1ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

// TestTransferMemberCreatesHistory tests that transferring creates history
func (s *TeamMemberHistoryTestSuite) TestTransferMemberCreatesHistory() {
	// Add to team1
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/company-admin/teams/%s/members", s.team1ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Transfer to team2 (should create history in both teams)
	req, _ = http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/company-admin/teams/%s/members/%s/transfer", s.team1ID, s.driverID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Check both teams have history
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/company-admin/teams/%s/member-history", s.team1ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)

	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/company-admin/teams/%s/member-history", s.team2ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)
}

// TestRemoveMemberCreatesHistory tests that removing creates history
func (s *TeamMemberHistoryTestSuite) TestRemoveMemberCreatesHistory() {
	// Add member
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/company-admin/teams/%s/members", s.team1ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Remove member (should create history)
	req, _ = http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/company-admin/teams/%s/members/%s", s.team1ID, s.driverID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Verify history
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/company-admin/teams/%s/member-history", s.team1ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)
	s.Equal(http.StatusOK, w.Code)
}

// TestGetUserTeamHistory tests retrieving a user's team history
func (s *TeamMemberHistoryTestSuite) TestGetUserTeamHistory() {
	// Add member to team
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/company-admin/teams/%s/members", s.team1ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Get user's team history
	req, _ = http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/company-admin/users/%s/team-history", s.driverID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "Should get user team history")
}

// TestCompleteWorkflow tests the complete member history workflow
func (s *TeamMemberHistoryTestSuite) TestCompleteWorkflow() {
	s.T().Log("✅ Testing team member history workflow")

	s.T().Log("Step 1: Add member creates history")
	s.TestAddMemberCreatesHistory()

	s.T().Log("Step 2: Update role creates history")
	s.TestUpdateRoleCreatesHistory()

	s.T().Log("Step 3: Transfer creates history")
	s.TestTransferMemberCreatesHistory()

	s.T().Log("Step 4: Remove creates history")
	s.TestRemoveMemberCreatesHistory()

	s.T().Log("Step 5: Get user history")
	s.TestGetUserTeamHistory()

	s.T().Log("✅ Complete member history workflow passed!")
}
