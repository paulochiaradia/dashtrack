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

// VehicleAssignmentHistoryTestSuite tests the Vehicle Assignment History API
type VehicleAssignmentHistoryTestSuite struct {
	suite.Suite
	router         *gin.Engine
	vehicleHandler *handlers.VehicleHandler
	authMiddleware *middleware.GinAuthMiddleware
	token          string
	companyID      uuid.UUID
	vehicleID      uuid.UUID
	driverID       uuid.UUID
	helperID       uuid.UUID
}

func TestVehicleAssignmentHistorySuite(t *testing.T) {
	suite.Run(t, new(VehicleAssignmentHistoryTestSuite))
}

func (s *VehicleAssignmentHistoryTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	// TODO: Setup test database connection
	// TODO: Setup repositories
	// TODO: Setup handlers
	// TODO: Setup router with routes
	// TODO: Create test data (company, vehicle, users)

	s.companyID = uuid.New()
	s.vehicleID = uuid.New()
	s.driverID = uuid.New()
	s.helperID = uuid.New()
}

func (s *VehicleAssignmentHistoryTestSuite) TearDownSuite() {
	// TODO: Cleanup test data
}

// TestUpdateDriverAssignment tests updating the driver assignment
func (s *VehicleAssignmentHistoryTestSuite) TestUpdateDriverAssignment() {
	reqBody := models.UpdateVehicleAssignmentRequest{
		DriverID: &s.driverID,
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/company-admin/vehicles/%s/assign", s.vehicleID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "Should update driver assignment successfully")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.True(response["success"].(bool))
}

// TestUpdateHelperAssignment tests updating the helper assignment
func (s *VehicleAssignmentHistoryTestSuite) TestUpdateHelperAssignment() {
	reqBody := models.UpdateVehicleAssignmentRequest{
		HelperID: &s.helperID,
	}

	body, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/company-admin/vehicles/%s/assign", s.vehicleID), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "Should update helper assignment successfully")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.True(response["success"].(bool))
}

// TestGetAssignmentHistory tests retrieving vehicle assignment history
func (s *VehicleAssignmentHistoryTestSuite) TestGetAssignmentHistory() {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/company-admin/vehicles/%s/assignment-history?limit=10", s.vehicleID), nil)
	req.Header.Set("Authorization", "Bearer "+s.token)

	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code, "Should retrieve assignment history successfully")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.True(response["success"].(bool))

	data := response["data"].(map[string]interface{})
	history := data["history"].([]interface{})
	s.GreaterOrEqual(len(history), 1, "Should have at least one history record")

	// Verify history record has populated details
	firstRecord := history[0].(map[string]interface{})
	s.NotEmpty(firstRecord["change_type"], "Should have change type")
	s.NotEmpty(firstRecord["changed_at"], "Should have timestamp")
}

// TestAutomaticHistoryCreation tests that history is created automatically
func (s *VehicleAssignmentHistoryTestSuite) TestAutomaticHistoryCreation() {
	// Get initial history count
	req1, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/company-admin/vehicles/%s/assignment-history?limit=100", s.vehicleID), nil)
	req1.Header.Set("Authorization", "Bearer "+s.token)
	w1 := httptest.NewRecorder()
	s.router.ServeHTTP(w1, req1)

	var initialResponse map[string]interface{}
	json.Unmarshal(w1.Body.Bytes(), &initialResponse)
	initialData := initialResponse["data"].(map[string]interface{})
	initialCount := int(initialData["count"].(float64))

	// Make an assignment change
	reqBody := models.UpdateVehicleAssignmentRequest{
		DriverID: &s.driverID,
	}
	body, _ := json.Marshal(reqBody)
	req2, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/company-admin/vehicles/%s/assign", s.vehicleID), bytes.NewBuffer(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+s.token)
	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)

	s.Equal(http.StatusOK, w2.Code)

	// Get updated history count
	req3, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/company-admin/vehicles/%s/assignment-history?limit=100", s.vehicleID), nil)
	req3.Header.Set("Authorization", "Bearer "+s.token)
	w3 := httptest.NewRecorder()
	s.router.ServeHTTP(w3, req3)

	var finalResponse map[string]interface{}
	json.Unmarshal(w3.Body.Bytes(), &finalResponse)
	finalData := finalResponse["data"].(map[string]interface{})
	finalCount := int(finalData["count"].(float64))

	s.Greater(finalCount, initialCount, "History should be created automatically")
}

// TestCompleteWorkflow tests the complete vehicle assignment workflow
func (s *VehicleAssignmentHistoryTestSuite) TestCompleteWorkflow() {
	// 1. Update driver assignment
	s.T().Log("Step 1: Updating driver assignment")
	s.TestUpdateDriverAssignment()

	// 2. Update helper assignment
	s.T().Log("Step 2: Updating helper assignment")
	s.TestUpdateHelperAssignment()

	// 3. Verify history was created
	s.T().Log("Step 3: Verifying assignment history")
	s.TestGetAssignmentHistory()

	// 4. Test automatic history creation
	s.T().Log("Step 4: Testing automatic history creation")
	s.TestAutomaticHistoryCreation()
}
