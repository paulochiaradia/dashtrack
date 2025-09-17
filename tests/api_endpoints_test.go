package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/auth"
	"github.com/paulochiaradia/dashtrack/internal/config"
	"github.com/paulochiaradia/dashtrack/internal/routes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// APITestSuite contains the test suite for API endpoints
type APITestSuite struct {
	suite.Suite
	router     *gin.Engine
	db         *sql.DB
	mock       sqlmock.Sqlmock
	userID     uuid.UUID
	roleID     uuid.UUID
	jwtManager *auth.JWTManager
	token      string
}

// SetupTest runs before each test
func (suite *APITestSuite) SetupTest() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(suite.T(), err)

	suite.db = db
	suite.mock = mock

	// Create test config
	cfg := &config.Config{
		ServerEnv:              "test",
		JWTSecret:              "test-secret-key-for-testing",
		JWTAccessExpireMinutes: 15,
		JWTRefreshExpireHours:  24,
		AppName:                "Dashtrack Test",
		AppVersion:             "1.0.0",
		BcryptCost:             4, // Lower cost for faster tests
	}

	// Initialize router
	suite.router = routes.NewRouter(db, cfg).GetEngine()

	// Create JWT manager for testing
	accessExpiry := time.Duration(cfg.JWTAccessExpireMinutes) * time.Minute
	refreshExpiry := time.Duration(cfg.JWTRefreshExpireHours) * time.Hour
	suite.jwtManager = auth.NewJWTManager(cfg.JWTSecret, accessExpiry, refreshExpiry, cfg.AppName)

	// Generate test IDs
	suite.userID = uuid.New()
	suite.roleID = uuid.New()

	// Generate test token
	userContext := auth.UserContext{
		UserID:   suite.userID,
		Email:    "admin@test.com",
		Name:     "Test Admin",
		RoleID:   suite.roleID,
		RoleName: "admin",
	}

	accessToken, _, err := suite.jwtManager.GenerateTokens(userContext)
	require.NoError(suite.T(), err)
	suite.token = accessToken
}

// TearDownTest runs after each test
func (suite *APITestSuite) TearDownTest() {
	suite.db.Close()
}

// TestHealthEndpoint tests the health check endpoint
func (suite *APITestSuite) TestHealthEndpoint() {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "ok", response["status"])
	assert.Contains(suite.T(), response["message"], "GIN")
	assert.Equal(suite.T(), "1.0.0", response["version"])
}

// TestMetricsEndpoint tests the metrics endpoint
func (suite *APITestSuite) TestMetricsEndpoint() {
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Contains(suite.T(), w.Header().Get("Content-Type"), "text/plain")
}

// TestLoginEndpoint tests the login endpoint
func (suite *APITestSuite) TestLoginEndpoint() {
	// Mock the database query for login
	rows := sqlmock.NewRows([]string{
		"id", "name", "email", "password", "phone", "cpf", "avatar", "role_id",
		"active", "last_login", "dashboard_config", "api_token", "login_attempts",
		"blocked_until", "password_changed_at", "created_at", "updated_at",
		"role_id", "role_name", "role_description", "role_created_at", "role_updated_at",
	}).AddRow(
		suite.userID, "Test User", "test@example.com", "$2a$04$gCDFWgNjnmks4kHYwjUvSOAijOtDPY2ML7NZ3kiay/Uyv0OMA8Jre", // bcrypt hash of "password"
		nil, nil, nil, suite.roleID,
		true, nil, nil, nil, 0,
		nil, time.Now(), time.Now(), time.Now(),
		suite.roleID, "admin", "Administrator", time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT u.id, u.name, u.email, u.password.*").
		WithArgs("test@example.com").
		WillReturnRows(rows)

	loginData := map[string]string{
		"email":    "test@example.com",
		"password": "password",
	}

	body, err := json.Marshal(loginData)
	require.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.NotNil(suite.T(), response["access_token"])
	assert.NotNil(suite.T(), response["refresh_token"])
	assert.NotNil(suite.T(), response["user"])
}

// TestLoginEndpointInvalidCredentials tests login with invalid credentials
func (suite *APITestSuite) TestLoginEndpointInvalidCredentials() {
	suite.mock.ExpectQuery("SELECT u.id, u.name, u.email, u.password.*").
		WithArgs("invalid@example.com").
		WillReturnError(sql.ErrNoRows)

	loginData := map[string]string{
		"email":    "invalid@example.com",
		"password": "wrongpassword",
	}

	body, err := json.Marshal(loginData)
	require.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Invalid credentials", response["error"])
}

// TestProfileEndpointAuthenticated tests the profile endpoint with valid token
func (suite *APITestSuite) TestProfileEndpointAuthenticated() {
	// Mock the database query for getting user by ID
	rows := sqlmock.NewRows([]string{
		"id", "name", "email", "password", "phone", "cpf", "avatar", "role_id",
		"active", "last_login", "dashboard_config", "api_token", "login_attempts",
		"blocked_until", "password_changed_at", "created_at", "updated_at",
		"role_id", "role_name", "role_description", "role_created_at", "role_updated_at",
	}).AddRow(
		suite.userID, "Test User", "admin@test.com", "hashed_password",
		nil, nil, nil, suite.roleID,
		true, nil, nil, nil, 0,
		nil, time.Now(), time.Now(), time.Now(),
		suite.roleID, "admin", "Administrator", time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT u.id, u.name, u.email, u.password.*").
		WithArgs(suite.userID).
		WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Test User", response["name"])
	assert.Equal(suite.T(), "admin@test.com", response["email"])
	assert.Equal(suite.T(), "admin", response["role"])
	assert.Equal(suite.T(), true, response["active"])
}

// TestProfileEndpointUnauthenticated tests the profile endpoint without token
func (suite *APITestSuite) TestProfileEndpointUnauthenticated() {
	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Authorization header required", response["error"])
}

// TestProfileEndpointInvalidToken tests the profile endpoint with invalid token
func (suite *APITestSuite) TestProfileEndpointInvalidToken() {
	req := httptest.NewRequest(http.MethodGet, "/profile", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Invalid token", response["error"])
}

// TestChangePasswordEndpoint tests the change password endpoint
func (suite *APITestSuite) TestChangePasswordEndpoint() {
	// Mock the database query for getting user by ID
	rows := sqlmock.NewRows([]string{
		"id", "name", "email", "password", "phone", "cpf", "avatar", "role_id",
		"active", "last_login", "dashboard_config", "api_token", "login_attempts",
		"blocked_until", "password_changed_at", "created_at", "updated_at",
		"role_id", "role_name", "role_description", "role_created_at", "role_updated_at",
	}).AddRow(
		suite.userID, "Test User", "admin@test.com", "$2a$04$gCDFWgNjnmks4kHYwjUvSOAijOtDPY2ML7NZ3kiay/Uyv0OMA8Jre", // bcrypt hash of "password"
		nil, nil, nil, suite.roleID,
		true, nil, nil, nil, 0,
		nil, time.Now(), time.Now(), time.Now(),
		suite.roleID, "admin", "Administrator", time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT u.id, u.name, u.email, u.password.*").
		WithArgs(suite.userID).
		WillReturnRows(rows)

	// Mock the password update
	suite.mock.ExpectExec("UPDATE users SET password.*").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), suite.userID).
		WillReturnResult(sqlmock.NewResult(1, 1))

	changePasswordData := map[string]string{
		"current_password": "password",
		"new_password":     "newpassword",
	}

	body, err := json.Marshal(changePasswordData)
	require.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/profile/change-password", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Password updated successfully", response["message"])
}

// TestRolesEndpoint tests the roles endpoint
func (suite *APITestSuite) TestRolesEndpoint() {
	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	roles, ok := response["roles"].([]interface{})
	require.True(suite.T(), ok)
	assert.Contains(suite.T(), roles, "admin")
	assert.Contains(suite.T(), roles, "manager")
	assert.Contains(suite.T(), roles, "driver")
	assert.Contains(suite.T(), roles, "helper")
}

// TestLogoutEndpoint tests the logout endpoint
func (suite *APITestSuite) TestLogoutEndpoint() {
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+suite.token)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Successfully logged out", response["message"])
}

// TestRefreshTokenEndpoint tests the refresh token endpoint
func (suite *APITestSuite) TestRefreshTokenEndpoint() {
	// Generate a refresh token
	userContext := auth.UserContext{
		UserID:   suite.userID,
		Email:    "admin@test.com",
		Name:     "Test Admin",
		RoleID:   suite.roleID,
		RoleName: "admin",
	}

	_, refreshToken, err := suite.jwtManager.GenerateTokens(userContext)
	require.NoError(suite.T(), err)

	// Mock the database query for getting user by ID
	rows := sqlmock.NewRows([]string{
		"id", "name", "email", "password", "phone", "cpf", "avatar", "role_id",
		"active", "last_login", "dashboard_config", "api_token", "login_attempts",
		"blocked_until", "password_changed_at", "created_at", "updated_at",
		"role_id", "role_name", "role_description", "role_created_at", "role_updated_at",
	}).AddRow(
		suite.userID, "Test User", "admin@test.com", "hashed_password",
		nil, nil, nil, suite.roleID,
		true, nil, nil, nil, 0,
		nil, time.Now(), time.Now(), time.Now(),
		suite.roleID, "admin", "Administrator", time.Now(), time.Now(),
	)

	suite.mock.ExpectQuery("SELECT u.id, u.name, u.email, u.password.*").
		WithArgs(suite.userID).
		WillReturnRows(rows)

	refreshData := map[string]string{
		"refresh_token": refreshToken,
	}

	body, err := json.Marshal(refreshData)
	require.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.NotNil(suite.T(), response["access_token"])
	assert.NotNil(suite.T(), response["refresh_token"])
}

// TestAdminEndpointsWithAdminRole tests admin endpoints with admin role
func (suite *APITestSuite) TestAdminEndpointsWithAdminRole() {
	// Test all admin endpoints return "Not implemented yet" for now
	adminEndpoints := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/admin/users"},
		{http.MethodPost, "/admin/users"},
		{http.MethodGet, "/admin/users/123"},
		{http.MethodPut, "/admin/users/123"},
		{http.MethodDelete, "/admin/users/123"},
	}

	for _, endpoint := range adminEndpoints {
		req := httptest.NewRequest(endpoint.method, endpoint.path, nil)
		req.Header.Set("Authorization", "Bearer "+suite.token)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		assert.Equal(suite.T(), http.StatusNotImplemented, w.Code,
			"Endpoint %s %s should return 501", endpoint.method, endpoint.path)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(suite.T(), err)

		assert.Equal(suite.T(), "Not implemented yet", response["error"])
	}
}

// TestAdminEndpointsWithoutAdminRole tests admin endpoints without admin role
func (suite *APITestSuite) TestAdminEndpointsWithoutAdminRole() {
	// Create token for non-admin user
	userContext := auth.UserContext{
		UserID:   uuid.New(),
		Email:    "manager@test.com",
		Name:     "Test Manager",
		RoleID:   uuid.New(),
		RoleName: "manager",
	}

	managerToken, _, err := suite.jwtManager.GenerateTokens(userContext)
	require.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+managerToken)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Insufficient permissions", response["error"])
}

// TestManagerEndpointsWithManagerRole tests manager endpoints with manager role
func (suite *APITestSuite) TestManagerEndpointsWithManagerRole() {
	// Create token for manager user
	userContext := auth.UserContext{
		UserID:   uuid.New(),
		Email:    "manager@test.com",
		Name:     "Test Manager",
		RoleID:   uuid.New(),
		RoleName: "manager",
	}

	managerToken, _, err := suite.jwtManager.GenerateTokens(userContext)
	require.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodGet, "/manager/users", nil)
	req.Header.Set("Authorization", "Bearer "+managerToken)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotImplemented, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Not implemented yet", response["error"])
}

// TestManagerEndpointsWithoutManagerRole tests manager endpoints without manager/admin role
func (suite *APITestSuite) TestManagerEndpointsWithoutManagerRole() {
	// Create token for driver user
	userContext := auth.UserContext{
		UserID:   uuid.New(),
		Email:    "driver@test.com",
		Name:     "Test Driver",
		RoleID:   uuid.New(),
		RoleName: "driver",
	}

	driverToken, _, err := suite.jwtManager.GenerateTokens(userContext)
	require.NoError(suite.T(), err)

	req := httptest.NewRequest(http.MethodGet, "/manager/users", nil)
	req.Header.Set("Authorization", "Bearer "+driverToken)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Insufficient permissions", response["error"])
}

// TestInvalidJSONRequest tests endpoints with invalid JSON
func (suite *APITestSuite) TestInvalidJSONRequest() {
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("invalid-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Invalid request format", response["error"])
}

// TestCORSHeaders tests CORS headers are properly set
func (suite *APITestSuite) TestCORSHeaders() {
	req := httptest.NewRequest(http.MethodOptions, "/health", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)
	assert.Equal(suite.T(), "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(suite.T(), w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(suite.T(), w.Header().Get("Access-Control-Allow-Methods"), "POST")
	assert.Contains(suite.T(), w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
}

// TestPasswordResetEndpoints tests password reset endpoints (currently not implemented)
func (suite *APITestSuite) TestPasswordResetEndpoints() {
	// Test forgot password
	req := httptest.NewRequest(http.MethodPost, "/auth/forgot-password", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotImplemented, w.Code)

	// Test reset password
	req = httptest.NewRequest(http.MethodPost, "/auth/reset-password", nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotImplemented, w.Code)
}

// TestRunAPITestSuite runs the API test suite
func TestRunAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
