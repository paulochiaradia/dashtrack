package integration

import (
	"bytes"
	"context"
	"encoding/json"
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

// TeamMembersIntegrationTestSuite tests the complete Team Members Management API
type TeamMembersIntegrationTestSuite struct {
	suite.Suite
	testDB       *testutils.TestDB
	router       *gin.Engine
	teamHandler  *handlers.TeamHandler
	authHandler  *handlers.AuthHandler
	teamRepo     *repository.TeamRepository
	userRepo     *repository.UserRepository
	roleRepo     *repository.RoleRepository
	vehicleRepo  *repository.VehicleRepository
	tokenService *services.TokenService
	token        string
	companyID    uuid.UUID
	adminUserID  uuid.UUID
	team1ID      uuid.UUID
	team2ID      uuid.UUID
	driverID     uuid.UUID
	helperID     uuid.UUID
	driverRoleID uuid.UUID
	helperRoleID uuid.UUID
}

func TestTeamMembersIntegrationSuite(t *testing.T) {
	suite.Run(t, new(TeamMembersIntegrationTestSuite))
}

func (s *TeamMembersIntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	// Setup test database
	var err error
	s.testDB, err = testutils.SetupTestDB("team_members")
	s.Require().NoError(err, "Failed to setup test database")

	// Initialize repositories
	s.teamRepo = repository.NewTeamRepository(s.testDB.SqlxDB)
	s.userRepo = repository.NewUserRepository(s.testDB.SqlxDB)
	s.roleRepo = repository.NewRoleRepository(s.testDB.SqlDB)
	s.vehicleRepo = repository.NewVehicleRepository(s.testDB.SqlxDB)

	// Initialize token service
	s.tokenService = services.NewTokenService(
		s.testDB.SqlxDB,
		"test-secret-key-min-32-characters-long",
		15*time.Minute,
		24*time.Hour,
	)

	// Initialize handlers
	s.teamHandler = handlers.NewTeamHandler(s.teamRepo, s.userRepo, s.vehicleRepo)

	// Setup router
	s.router = gin.New()
	s.setupRoutes()

	// Create test data
	s.createTestData()
}

func (s *TeamMembersIntegrationTestSuite) TearDownSuite() {
	if s.testDB != nil {
		s.testDB.TearDown()
	}
}

func (s *TeamMembersIntegrationTestSuite) setupRoutes() {
	ginAuth := middleware.NewGinAuthMiddleware(s.tokenService)

	api := s.router.Group("/api/v1")
	{
		companyAdmin := api.Group("/company-admin")
		companyAdmin.Use(ginAuth.RequireAuth())
		companyAdmin.Use(middleware.RequireCompanyAdmin())
		{
			companyAdmin.POST("/teams/:team_id/members", s.teamHandler.AddMember)
			companyAdmin.GET("/teams/:team_id/members", s.teamHandler.GetMembers)
			companyAdmin.PUT("/teams/:team_id/members/:user_id/role", s.teamHandler.UpdateMemberRole)
			companyAdmin.POST("/teams/:team_id/members/:user_id/transfer", s.teamHandler.TransferMemberToTeam)
			companyAdmin.DELETE("/teams/:team_id/members/:user_id", s.teamHandler.RemoveMember)
		}
	}
}

func (s *TeamMembersIntegrationTestSuite) createTestData() {
	// Get roles
	driverRole, err := s.roleRepo.GetByName("driver")
	s.Require().NoError(err)
	s.driverRoleID = driverRole.ID

	helperRole, err := s.roleRepo.GetByName("helper")
	s.Require().NoError(err)
	s.helperRoleID = helperRole.ID

	companyAdminRole, err := s.roleRepo.GetByName("company_admin")
	s.Require().NoError(err)

	// Create company
	company := &models.Company{
		Name:             "Test Company",
		Slug:             "test-company",
		Email:            "test@company.com",
		Country:          "Brazil",
		SubscriptionPlan: "premium",
		Status:           "active",
	}
	err = s.testDB.DB.Create(company).Error
	s.Require().NoError(err)
	s.companyID = company.ID

	// Create admin user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Admin@123"), bcrypt.DefaultCost)
	phone := "+5511999999999"
	cpf := "12345678901"
	adminUser := &models.User{
		Name:      "Admin User",
		Email:     "admin@test.com",
		Password:  string(hashedPassword),
		Phone:     &phone,
		CPF:       &cpf,
		CompanyID: &s.companyID,
		RoleID:    companyAdminRole.ID,
		Active:    true,
	}
	err = s.testDB.DB.Create(adminUser).Error
	s.Require().NoError(err)
	s.adminUserID = adminUser.ID

	// Reload user with role for token generation
	adminUser.Role = companyAdminRole

	// Generate token for admin
	ctx := context.Background()
	tokenPair, err := s.tokenService.GenerateTokenPair(ctx, adminUser, "127.0.0.1", "test-agent")
	s.Require().NoError(err)
	s.token = tokenPair.AccessToken

	// Create teams
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

	// Create driver user
	driverPhone := "+5511888888888"
	driverCPF := "98765432100"
	driver := &models.User{
		Name:      "John Driver",
		Email:     "driver@test.com",
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

	// Create helper user
	helperPhone := "+5511777777777"
	helperCPF := "11122233344"
	helper := &models.User{
		Name:      "Mary Helper",
		Email:     "helper@test.com",
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

// TestAddTeamMember tests adding a member to a team
func (s *TeamMembersIntegrationTestSuite) TestAddTeamMember() {
	reqBody := models.AssignTeamMemberRequest{
		UserID:     s.driverID,
		RoleInTeam: "driver",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/company-admin/teams/%s/members", s.team1ID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code, "Should create team member successfully")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.True(response["success"].(bool))
}

// TestGetTeamMembers tests retrieving team members
func (s *TeamMembersIntegrationTestSuite) TestGetTeamMembers() {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/company-admin/teams/%s/members", s.team1ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "Should retrieve team members successfully")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.True(response["success"].(bool))

	data := response["data"].(map[string]interface{})
	members := data["members"].([]interface{})
	s.GreaterOrEqual(len(members), 1, "Should have at least one member")
}

// TestUpdateMemberRole tests updating a team member's role
func (s *TeamMembersIntegrationTestSuite) TestUpdateMemberRole() {
	reqBody := map[string]string{
		"role_in_team": "team_lead",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/company-admin/teams/%s/members/%s/role", s.team1ID, s.driverID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "Should update member role successfully")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.True(response["success"].(bool))
}

// TestTransferMember tests transferring a member to another team
func (s *TeamMembersIntegrationTestSuite) TestTransferMember() {
	reqBody := map[string]interface{}{
		"to_team_id":   s.team2ID.String(),
		"role_in_team": "driver",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/company-admin/teams/%s/members/%s/transfer", s.team1ID, s.helperID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "Should transfer member successfully")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.True(response["success"].(bool))
}

// TestRemoveMember tests removing a member from a team
func (s *TeamMembersIntegrationTestSuite) TestRemoveMember() {
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/company-admin/teams/%s/members/%s", s.team1ID, s.driverID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "Should remove member successfully")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.True(response["success"].(bool))
}

// TestCompleteWorkflow tests the complete workflow of member management
func (s *TeamMembersIntegrationTestSuite) TestCompleteWorkflow() {
	// 1. Add driver to team 1
	s.T().Log("Step 1: Adding driver to team 1")
	s.TestAddTeamMember()

	// 2. Verify member is in team
	s.T().Log("Step 2: Verifying member in team")
	s.TestGetTeamMembers()

	// 3. Update member role
	s.T().Log("Step 3: Updating member role")
	s.TestUpdateMemberRole()

	// 4. Transfer member to team 2
	s.T().Log("Step 4: Transferring member")
	s.TestTransferMember()

	// 5. Remove member
	s.T().Log("Step 5: Removing member")
	s.TestRemoveMember()
}
