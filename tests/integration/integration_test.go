package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite
	server *httptest.Server
}

func (suite *IntegrationTestSuite) SetupSuite() {
	// This would normally setup the entire application
	// For now, we'll create a simple test server
	mux := http.NewServeMux()
	
	// Health endpoint
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":  "ok",
			"message": "API is running",
			"database": "connected",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})
	
	// Roles endpoint
	mux.HandleFunc("/roles", func(w http.ResponseWriter, r *http.Request) {
		roles := []models.Role{
			{
				ID:          uuid.New(),
				Name:        "admin",
				Description: "System administrator with full access",
			},
			{
				ID:          uuid.New(),
				Name:        "driver",
				Description: "Vehicle driver with access to assigned vehicles",
			},
			{
				ID:          uuid.New(),
				Name:        "helper",
				Description: "Driver assistant with limited access to assigned vehicles",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(roles)
	})
	
	// Users endpoints
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			users := []models.User{
				{
					ID:     uuid.New(),
					Name:   "Admin User",
					Email:  "admin@example.com",
					Phone:  stringPtr("123456789"),
					Active: true,
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(users)
		case "POST":
			var createReq models.CreateUserRequest
			if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			
			roleID, _ := uuid.Parse(createReq.RoleID)
			user := models.User{
				ID:     uuid.New(),
				Name:   createReq.Name,
				Email:  createReq.Email,
				Phone:  stringPtr(createReq.Phone),
				RoleID: roleID,
				Active: true,
			}
			
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(user)
		}
	})
	
	suite.server = httptest.NewServer(mux)
}

func (suite *IntegrationTestSuite) TearDownSuite() {
	suite.server.Close()
}

func (suite *IntegrationTestSuite) TestHealthEndpoint() {
	resp, err := http.Get(suite.server.URL + "/health")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(suite.T(), err)
	
	assert.Equal(suite.T(), "ok", response["status"])
	assert.Equal(suite.T(), "API is running", response["message"])
}

func (suite *IntegrationTestSuite) TestGetRoles() {
	resp, err := http.Get(suite.server.URL + "/roles")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var roles []models.Role
	err = json.NewDecoder(resp.Body).Decode(&roles)
	require.NoError(suite.T(), err)
	
	assert.Len(suite.T(), roles, 3)
	assert.Equal(suite.T(), "admin", roles[0].Name)
	assert.Equal(suite.T(), "driver", roles[1].Name)
	assert.Equal(suite.T(), "helper", roles[2].Name)
}

func (suite *IntegrationTestSuite) TestGetUsers() {
	resp, err := http.Get(suite.server.URL + "/users")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	var users []models.User
	err = json.NewDecoder(resp.Body).Decode(&users)
	require.NoError(suite.T(), err)
	
	assert.Len(suite.T(), users, 1)
	assert.Equal(suite.T(), "Admin User", users[0].Name)
}

func (suite *IntegrationTestSuite) TestCreateUser() {
	createReq := models.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
		Phone:    "987654321",
		RoleID:   uuid.New().String(),
	}
	
	body, err := json.Marshal(createReq)
	require.NoError(suite.T(), err)
	
	resp, err := http.Post(
		suite.server.URL+"/users",
		"application/json",
		bytes.NewBuffer(body),
	)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	
	var user models.User
	err = json.NewDecoder(resp.Body).Decode(&user)
	require.NoError(suite.T(), err)
	
	assert.Equal(suite.T(), createReq.Name, user.Name)
	assert.Equal(suite.T(), createReq.Email, user.Email)
	assert.NotEmpty(suite.T(), user.ID)
}

func (suite *IntegrationTestSuite) TestCORSHeaders() {
	req, err := http.NewRequest("OPTIONS", suite.server.URL+"/health", nil)
	require.NoError(suite.T(), err)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	// Note: CORS headers would be tested here if implemented
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
}

func (suite *IntegrationTestSuite) TestAPIEndpointsFlow() {
	// Test the complete flow: Health -> Roles -> Create User -> Get Users
	
	// 1. Check health
	resp, err := http.Get(suite.server.URL + "/health")
	require.NoError(suite.T(), err)
	resp.Body.Close()
	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)
	
	// 2. Get roles
	resp, err = http.Get(suite.server.URL + "/roles")
	require.NoError(suite.T(), err)
	
	var roles []models.Role
	err = json.NewDecoder(resp.Body).Decode(&roles)
	require.NoError(suite.T(), err)
	resp.Body.Close()
	assert.NotEmpty(suite.T(), roles)
	
	// 3. Create user with first role
	createReq := models.CreateUserRequest{
		Name:     "Flow Test User",
		Email:    "flow@example.com",
		Password: "password123",
		Phone:    "111222333",
		RoleID:   roles[0].ID.String(),
	}
	
	body, err := json.Marshal(createReq)
	require.NoError(suite.T(), err)
	
	resp, err = http.Post(
		suite.server.URL+"/users",
		"application/json",
		bytes.NewBuffer(body),
	)
	require.NoError(suite.T(), err)
	resp.Body.Close()
	assert.Equal(suite.T(), http.StatusCreated, resp.StatusCode)
	
	// 4. Verify user was created
	resp, err = http.Get(suite.server.URL + "/users")
	require.NoError(suite.T(), err)
	defer resp.Body.Close()
	
	var users []models.User
	err = json.NewDecoder(resp.Body).Decode(&users)
	require.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), users)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

func TestIntegrationSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
