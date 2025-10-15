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

// TeamMemberHistoryTestSuite tests the Team Member History API
type TeamMemberHistoryTestSuite struct {
	suite.Suite
	router         *gin.Engine
	teamHandler    *handlers.TeamHandler
	authMiddleware *middleware.GinAuthMiddleware
	token          string
	companyID      uuid.UUID
	teamID         uuid.UUID
	driverID       uuid.UUID
	helperID       uuid.UUID
}

func TestTeamMemberHistorySuite(t *testing.T) {
	suite.Run(t, new(TeamMemberHistoryTestSuite))
}

func (s *TeamMemberHistoryTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	// TODO: Setup test database connection
	// TODO: Setup repositories
	// TODO: Setup handlers
	// TODO: Setup router with routes
	// TODO: Create test data (company, team, users)

	s.companyID = uuid.New()
	s.teamID = uuid.New()
	s.driverID = uuid.New()
	s.helperID = uuid.New()
}

func (s *TeamMemberHistoryTestSuite) TearDownSuite() {
	// TODO: Cleanup test data
}

// TestAddMemberCreatesHistory tests that adding a member creates history
func (s *TeamMemberHistoryTestSuite) TestAddMemberCreatesHistory() {
	// Add member
	reqBody := models.AssignTeamMemberRequest{
		UserID:     s.driverID,
		RoleInTeam: "driver",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/company-admin/teams/%s/members", s.teamID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code, "Should add member successfully")

	// Check history was created
	req2, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/company-admin/teams/%s/member-history?limit=10", s.teamID), nil)
	req2.Header.Set("Authorization", "Bearer "+s.token)

	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)

	s.Equal(http.StatusOK, w2.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	s.NoError(err)

	data := response["data"].(map[string]interface{})
	history := data["history"].([]interface{})
	s.GreaterOrEqual(len(history), 1, "Should have history record for member addition")

	// Verify history record details
	firstRecord := history[0].(map[string]interface{})
	s.Equal("added", firstRecord["change_type"], "Should have 'added' change type")
}

// TestUpdateRoleCreatesHistory tests that updating role creates history
func (s *TeamMemberHistoryTestSuite) TestUpdateRoleCreatesHistory() {
	// Update member role
	reqBody := map[string]string{
		"role_in_team": "team_lead",
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/company-admin/teams/%s/members/%s/role", s.teamID, s.driverID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "Should update role successfully")

	// Check history was created
	req2, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/company-admin/teams/%s/member-history?limit=10", s.teamID), nil)
	req2.Header.Set("Authorization", "Bearer "+s.token)

	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)

	s.Equal(http.StatusOK, w2.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	s.NoError(err)

	data := response["data"].(map[string]interface{})
	history := data["history"].([]interface{})

	// Find role_changed record
	var foundRoleChange bool
	for _, record := range history {
		rec := record.(map[string]interface{})
		if rec["change_type"] == "role_changed" {
			foundRoleChange = true
			s.Equal("driver", rec["previous_role_in_team"], "Should have previous role")
			s.Equal("team_lead", rec["new_role_in_team"], "Should have new role")
			break
		}
	}
	s.True(foundRoleChange, "Should have role_changed history record")
}

// TestRemoveMemberCreatesHistory tests that removing member creates history
func (s *TeamMemberHistoryTestSuite) TestRemoveMemberCreatesHistory() {
	// Remove member
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/company-admin/teams/%s/members/%s", s.teamID, s.helperID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "Should remove member successfully")

	// Check history was created
	req2, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/company-admin/teams/%s/member-history?limit=10", s.teamID), nil)
	req2.Header.Set("Authorization", "Bearer "+s.token)

	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)

	s.Equal(http.StatusOK, w2.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w2.Body.Bytes(), &response)
	s.NoError(err)

	data := response["data"].(map[string]interface{})
	history := data["history"].([]interface{})

	// Find removed record
	var foundRemoved bool
	for _, record := range history {
		rec := record.(map[string]interface{})
		if rec["change_type"] == "removed" {
			foundRemoved = true
			break
		}
	}
	s.True(foundRemoved, "Should have removed history record")
}

// TestGetTeamMemberHistory tests retrieving team member history
func (s *TeamMemberHistoryTestSuite) TestGetTeamMemberHistory() {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/company-admin/teams/%s/member-history?limit=10", s.teamID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "Should retrieve team history successfully")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.True(response["success"].(bool))

	data := response["data"].(map[string]interface{})
	history := data["history"].([]interface{})
	s.GreaterOrEqual(len(history), 1, "Should have at least one history record")

	// Verify populated details
	firstRecord := history[0].(map[string]interface{})
	s.NotEmpty(firstRecord["change_type"], "Should have change type")
	s.NotEmpty(firstRecord["changed_at"], "Should have timestamp")
	s.NotNil(firstRecord["user"], "Should have user details populated")
}

// TestGetUserTeamHistory tests retrieving user team history
func (s *TeamMemberHistoryTestSuite) TestGetUserTeamHistory() {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/company-admin/teams/users/%s/team-history?limit=10", s.driverID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "Should retrieve user history successfully")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.True(response["success"].(bool))

	data := response["data"].(map[string]interface{})
	history := data["history"].([]interface{})
	s.GreaterOrEqual(len(history), 1, "Should have at least one history record")

	// Verify all records are for this user
	for _, record := range history {
		rec := record.(map[string]interface{})
		s.Equal(s.driverID.String(), rec["user_id"], "All records should be for this user")
		s.NotNil(rec["team"], "Should have team details populated")
	}
}

// TestCompleteWorkflow tests the complete team member history workflow
func (s *TeamMemberHistoryTestSuite) TestCompleteWorkflow() {
	// 1. Add member and verify history
	s.T().Log("Step 1: Adding member and checking history")
	s.TestAddMemberCreatesHistory()

	// 2. Update role and verify history
	s.T().Log("Step 2: Updating role and checking history")
	s.TestUpdateRoleCreatesHistory()

	// 3. Get team history
	s.T().Log("Step 3: Getting team history")
	s.TestGetTeamMemberHistory()

	// 4. Get user history
	s.T().Log("Step 4: Getting user history")
	s.TestGetUserTeamHistory()

	// 5. Remove member and verify history
	s.T().Log("Step 5: Removing member and checking history")
	s.TestRemoveMemberCreatesHistory()
}
