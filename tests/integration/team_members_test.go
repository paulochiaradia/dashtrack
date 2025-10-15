package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/handlers"
	"github.com/paulochiaradia/dashtrack/internal/middleware"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/stretchr/testify/suite"
)

// TeamMembersIntegrationTestSuite tests the complete Team Members Management API
type TeamMembersIntegrationTestSuite struct {
	suite.Suite
	router         *gin.Engine
	teamHandler    *handlers.TeamHandler
	authMiddleware *middleware.GinAuthMiddleware
	token          string
	companyID      uuid.UUID
	team1ID        uuid.UUID
	team2ID        uuid.UUID
	driverID       uuid.UUID
	helperID       uuid.UUID
}

func TestTeamMembersIntegrationSuite(t *testing.T) {
	suite.Run(t, new(TeamMembersIntegrationTestSuite))
}

func (s *TeamMembersIntegrationTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	// TODO: Setup test database connection
	// TODO: Setup repositories
	// TODO: Setup handlers
	// TODO: Setup router with routes
	// TODO: Create test data (company, teams, users)

	s.companyID = uuid.New()
	s.team1ID = uuid.New()
	s.team2ID = uuid.New()
	s.driverID = uuid.New()
	s.helperID = uuid.New()
}

func (s *TeamMembersIntegrationTestSuite) TearDownSuite() {
	// TODO: Cleanup test data
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
