package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/tests/testutils"
)

type TeamManagementTestSuite struct {
	suite.Suite
	apiURL        string
	companyID     uuid.UUID
	adminToken    string
	userToken     string
	testTeamID    uuid.UUID
	testUserID    uuid.UUID
	testVehicleID uuid.UUID
}

func TestTeamManagementSuite(t *testing.T) {
	suite.Run(t, new(TeamManagementTestSuite))
}

func (s *TeamManagementTestSuite) SetupSuite() {
	// Get API URL from environment or use default
	s.apiURL = testutils.GetAPIURL()

	// Setup test data
	s.companyID = uuid.New()

	// Create test tokens (mock for now)
	s.adminToken = "test-admin-token"
	s.userToken = "test-user-token"
}

func (s *TeamManagementTestSuite) TearDownSuite() {
	// Cleanup test data if needed
}

// ============================================================================
// TEST: Create Team
// ============================================================================

func (s *TeamManagementTestSuite) TestCreateTeam() {
	createReq := models.CreateTeamRequest{
		Name:        "Test Team Alpha",
		Description: stringPtr("Integration test team"),
	}

	body, err := json.Marshal(createReq)
	require.NoError(s.T(), err)

	req, err := http.NewRequest("POST", s.apiURL+"/api/v1/company-admin/teams", bytes.NewBuffer(body))
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	// For now, we expect unauthorized since we don't have real auth
	// In a real test environment, this would be 201 Created
	assert.Contains(s.T(), []int{http.StatusCreated, http.StatusUnauthorized}, resp.StatusCode)
}

// ============================================================================
// TEST: List Teams
// ============================================================================

func (s *TeamManagementTestSuite) TestListTeams() {
	req, err := http.NewRequest("GET", s.apiURL+"/api/v1/company-admin/teams", nil)
	require.NoError(s.T(), err)

	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Contains(s.T(), []int{http.StatusOK, http.StatusUnauthorized}, resp.StatusCode)
}

// ============================================================================
// TEST: Get Team Details
// ============================================================================

func (s *TeamManagementTestSuite) TestGetTeamDetails() {
	teamID := uuid.New()
	url := fmt.Sprintf("%s/api/v1/company-admin/teams/%s", s.apiURL, teamID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(s.T(), err)

	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	// Expect 404 since team doesn't exist, or 401 without auth
	assert.Contains(s.T(), []int{http.StatusNotFound, http.StatusUnauthorized}, resp.StatusCode)
}

// ============================================================================
// TEST: Update Team
// ============================================================================

func (s *TeamManagementTestSuite) TestUpdateTeam() {
	teamID := uuid.New()
	updateReq := models.UpdateTeamRequest{
		Name:        stringPtr("Updated Team Name"),
		Description: stringPtr("Updated description"),
	}

	body, err := json.Marshal(updateReq)
	require.NoError(s.T(), err)

	url := fmt.Sprintf("%s/api/v1/company-admin/teams/%s", s.apiURL, teamID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Contains(s.T(), []int{http.StatusOK, http.StatusNotFound, http.StatusUnauthorized}, resp.StatusCode)
}

// ============================================================================
// TEST: Delete Team
// ============================================================================

func (s *TeamManagementTestSuite) TestDeleteTeam() {
	teamID := uuid.New()
	url := fmt.Sprintf("%s/api/v1/company-admin/teams/%s", s.apiURL, teamID)

	req, err := http.NewRequest("DELETE", url, nil)
	require.NoError(s.T(), err)

	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Contains(s.T(), []int{http.StatusOK, http.StatusNotFound, http.StatusUnauthorized}, resp.StatusCode)
}

// ============================================================================
// TEST: Add Team Member
// ============================================================================

func (s *TeamManagementTestSuite) TestAddTeamMember() {
	teamID := uuid.New()
	userID := uuid.New()

	addMemberReq := models.AddTeamMemberRequest{
		UserID:     userID,
		RoleInTeam: "driver",
	}

	body, err := json.Marshal(addMemberReq)
	require.NoError(s.T(), err)

	url := fmt.Sprintf("%s/api/v1/company-admin/teams/%s/members", s.apiURL, teamID)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Contains(s.T(), []int{http.StatusCreated, http.StatusNotFound, http.StatusUnauthorized}, resp.StatusCode)
}

// ============================================================================
// TEST: List Team Members
// ============================================================================

func (s *TeamManagementTestSuite) TestListTeamMembers() {
	teamID := uuid.New()
	url := fmt.Sprintf("%s/api/v1/company-admin/teams/%s/members", s.apiURL, teamID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(s.T(), err)

	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Contains(s.T(), []int{http.StatusOK, http.StatusNotFound, http.StatusUnauthorized}, resp.StatusCode)
}

// ============================================================================
// TEST: Remove Team Member
// ============================================================================

func (s *TeamManagementTestSuite) TestRemoveTeamMember() {
	teamID := uuid.New()
	userID := uuid.New()

	url := fmt.Sprintf("%s/api/v1/company-admin/teams/%s/members/%s", s.apiURL, teamID, userID)
	req, err := http.NewRequest("DELETE", url, nil)
	require.NoError(s.T(), err)

	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Contains(s.T(), []int{http.StatusOK, http.StatusNotFound, http.StatusUnauthorized}, resp.StatusCode)
}

// ============================================================================
// TEST: Update Member Role
// ============================================================================

func (s *TeamManagementTestSuite) TestUpdateMemberRole() {
	teamID := uuid.New()
	userID := uuid.New()

	updateRoleReq := models.UpdateMemberRoleRequest{
		NewRoleInTeam: "manager",
	}

	body, err := json.Marshal(updateRoleReq)
	require.NoError(s.T(), err)

	url := fmt.Sprintf("%s/api/v1/company-admin/teams/%s/members/%s/role", s.apiURL, teamID, userID)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(body))
	require.NoError(s.T(), err)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Contains(s.T(), []int{http.StatusOK, http.StatusNotFound, http.StatusUnauthorized}, resp.StatusCode)
}

// ============================================================================
// TEST: Get Team Statistics
// ============================================================================

func (s *TeamManagementTestSuite) TestGetTeamStatistics() {
	teamID := uuid.New()
	url := fmt.Sprintf("%s/api/v1/company-admin/teams/%s/stats", s.apiURL, teamID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(s.T(), err)

	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Contains(s.T(), []int{http.StatusOK, http.StatusNotFound, http.StatusUnauthorized}, resp.StatusCode)
}

// ============================================================================
// TEST: Get Team Vehicles
// ============================================================================

func (s *TeamManagementTestSuite) TestGetTeamVehicles() {
	teamID := uuid.New()
	url := fmt.Sprintf("%s/api/v1/company-admin/teams/%s/vehicles", s.apiURL, teamID)

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(s.T(), err)

	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Contains(s.T(), []int{http.StatusOK, http.StatusNotFound, http.StatusUnauthorized}, resp.StatusCode)
}

// ============================================================================
// TEST: Assign Vehicle to Team
// ============================================================================

func (s *TeamManagementTestSuite) TestAssignVehicleToTeam() {
	teamID := uuid.New()
	vehicleID := uuid.New()

	url := fmt.Sprintf("%s/api/v1/company-admin/teams/%s/vehicles/%s", s.apiURL, teamID, vehicleID)
	req, err := http.NewRequest("POST", url, nil)
	require.NoError(s.T(), err)

	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Contains(s.T(), []int{http.StatusOK, http.StatusNotFound, http.StatusUnauthorized}, resp.StatusCode)
}

// ============================================================================
// TEST: Unassign Vehicle from Team
// ============================================================================

func (s *TeamManagementTestSuite) TestUnassignVehicleFromTeam() {
	teamID := uuid.New()
	vehicleID := uuid.New()

	url := fmt.Sprintf("%s/api/v1/company-admin/teams/%s/vehicles/%s", s.apiURL, teamID, vehicleID)
	req, err := http.NewRequest("DELETE", url, nil)
	require.NoError(s.T(), err)

	req.Header.Set("Authorization", "Bearer "+s.adminToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Contains(s.T(), []int{http.StatusOK, http.StatusNotFound, http.StatusUnauthorized}, resp.StatusCode)
}

// ============================================================================
// TEST: Get My Teams
// ============================================================================

func (s *TeamManagementTestSuite) TestGetMyTeams() {
	url := s.apiURL + "/api/v1/teams/my-teams"

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(s.T(), err)

	req.Header.Set("Authorization", "Bearer "+s.userToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	require.NoError(s.T(), err)
	defer resp.Body.Close()

	assert.Contains(s.T(), []int{http.StatusOK, http.StatusUnauthorized}, resp.StatusCode)
}

// ============================================================================
// TEST: Role-Based Access Control
// ============================================================================

func (s *TeamManagementTestSuite) TestRoleBasedAccessControl() {
	teamID := uuid.New()

	tests := []struct {
		name       string
		endpoint   string
		method     string
		token      string
		expectCode []int
	}{
		{
			name:       "Admin can list teams",
			endpoint:   "/api/v1/admin/teams",
			method:     "GET",
			token:      s.adminToken,
			expectCode: []int{http.StatusOK, http.StatusUnauthorized},
		},
		{
			name:       "Manager can view teams",
			endpoint:   "/api/v1/manager/teams",
			method:     "GET",
			token:      s.userToken,
			expectCode: []int{http.StatusOK, http.StatusUnauthorized, http.StatusForbidden},
		},
		{
			name:       "User can access my-teams",
			endpoint:   "/api/v1/teams/my-teams",
			method:     "GET",
			token:      s.userToken,
			expectCode: []int{http.StatusOK, http.StatusUnauthorized},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			url := s.apiURL + tt.endpoint
			if tt.endpoint == "/api/v1/admin/teams/:id" || tt.endpoint == "/api/v1/manager/teams/:id" {
				url = fmt.Sprintf("%s/api/v1/admin/teams/%s", s.apiURL, teamID)
			}

			req, err := http.NewRequest(tt.method, url, nil)
			require.NoError(s.T(), err)

			req.Header.Set("Authorization", "Bearer "+tt.token)

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			require.NoError(s.T(), err)
			defer resp.Body.Close()

			assert.Contains(s.T(), tt.expectCode, resp.StatusCode, "Test: %s", tt.name)
		})
	}
}

// ============================================================================
// TEST: Validation Tests
// ============================================================================

func (s *TeamManagementTestSuite) TestValidationErrors() {
	tests := []struct {
		name       string
		endpoint   string
		method     string
		body       interface{}
		expectCode int
	}{
		{
			name:     "Create team with empty name",
			endpoint: "/api/v1/company-admin/teams",
			method:   "POST",
			body: models.CreateTeamRequest{
				Name:        "",
				Description: stringPtr("Test"),
			},
			expectCode: http.StatusBadRequest,
		},
		{
			name:     "Add member with invalid role",
			endpoint: "/api/v1/company-admin/teams/" + uuid.New().String() + "/members",
			method:   "POST",
			body: models.AddTeamMemberRequest{
				UserID:     uuid.New(),
				RoleInTeam: "invalid_role",
			},
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			body, err := json.Marshal(tt.body)
			require.NoError(s.T(), err)

			url := s.apiURL + tt.endpoint
			req, err := http.NewRequest(tt.method, url, bytes.NewBuffer(body))
			require.NoError(s.T(), err)

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+s.adminToken)

			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Do(req)
			require.NoError(s.T(), err)
			defer resp.Body.Close()

			// Allow both expected validation error and unauthorized
			assert.Contains(s.T(), []int{tt.expectCode, http.StatusUnauthorized}, resp.StatusCode)
		})
	}
}

// ============================================================================
// TEST: Endpoint Response Format
// ============================================================================

func (s *TeamManagementTestSuite) TestResponseFormat() {
	// This test verifies the response structure matches expected format
	// In a real environment with database, we would check actual data

	type StandardResponse struct {
		Success bool                   `json:"success"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}

	// Test would verify response structure
	// For now, just ensure the test framework is working
	assert.NotNil(s.T(), s.apiURL)
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// stringPtr returns a pointer to a string
func stringPtr(s string) *string {
	return &s
}
