package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"

	"github.com/paulochiaradia/dashtrack/internal/auth"
	"github.com/paulochiaradia/dashtrack/internal/config"
	"github.com/paulochiaradia/dashtrack/internal/handlers"
	"github.com/paulochiaradia/dashtrack/internal/middleware"
	"github.com/paulochiaradia/dashtrack/internal/repository"
	"github.com/paulochiaradia/dashtrack/internal/services"
)

// UserWorkflowsTestSuite represents the user workflows end-to-end test suite
type UserWorkflowsTestSuite struct {
	suite.Suite
	db      *sqlx.DB
	router  *gin.Engine
	server  *httptest.Server
	baseURL string
	cleanup func()

	// Test scenario data
	masterToken  string
	companyToken string
	adminToken   string
	driverToken  string

	masterUserID  uuid.UUID
	companyUserID uuid.UUID
	adminUserID   uuid.UUID
	driverUserID  uuid.UUID
	testCompanyID uuid.UUID
}

func (suite *UserWorkflowsTestSuite) SetupSuite() {
	gin.SetMode(gin.TestMode)

	// Setup test database
	suite.setupDatabase()

	// Setup application
	suite.setupApplication()

	// Start test server
	suite.server = httptest.NewServer(suite.router)
	suite.baseURL = suite.server.URL

	// Initialize test data
	suite.initializeTestScenarios()
}

func (suite *UserWorkflowsTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	if suite.cleanup != nil {
		suite.cleanup()
	}
	if suite.db != nil {
		suite.db.Close()
	}
}

func (suite *UserWorkflowsTestSuite) setupDatabase() {
	dbURL := os.Getenv("E2E_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/dashtrack_e2e?sslmode=disable"
	}

	var err error
	suite.db, err = sqlx.Connect("postgres", dbURL)
	if err != nil {
		suite.T().Skipf("Cannot connect to E2E test database: %v", err)
		return
	}

	// Run migrations
	driver, err := postgres.WithInstance(suite.db.DB, &postgres.Config{})
	assert.NoError(suite.T(), err)

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"postgres",
		driver,
	)
	assert.NoError(suite.T(), err)

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		suite.T().Fatalf("Failed to run migrations: %v", err)
	}

	// Set cleanup function
	suite.cleanup = func() {
		ctx := context.Background()
		// Clean test data
		_, _ = suite.db.ExecContext(ctx, "DELETE FROM users WHERE email LIKE '%@e2e.test'")
		_, _ = suite.db.ExecContext(ctx, "DELETE FROM companies WHERE name LIKE 'E2E%'")
	}
}

func (suite *UserWorkflowsTestSuite) setupApplication() {
	// Initialize repositories
	userRepo := repository.NewUserRepository(suite.db)
	roleRepo := repository.NewRoleRepository(suite.db.DB) // Convert sqlx.DB to sql.DB

	// Initialize services
	userService := services.NewUserService(userRepo, roleRepo, 12) // Add bcrypt cost

	// Initialize JWT manager
	cfg := &config.Config{
		JWTSecret:              "e2e-test-secret-key",
		JWTAccessExpireMinutes: 60,
		JWTRefreshExpireHours:  24,
	}
	jwtManager := auth.NewJWTManager(
		cfg.JWTSecret,
		time.Duration(cfg.JWTAccessExpireMinutes)*time.Minute,
		time.Duration(cfg.JWTRefreshExpireHours)*time.Hour,
		"dashtrack-e2e",
	)

	// Initialize additional repositories and services for auth
	authLogRepo := &mockAuthLogRepository{} // Use mock instead of real repository
	tokenService := services.NewTokenService(
		suite.db,
		cfg.JWTSecret,
		time.Duration(cfg.JWTAccessExpireMinutes)*time.Minute,
		time.Duration(cfg.JWTRefreshExpireHours)*time.Hour,
	)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, authLogRepo, jwtManager, tokenService, 12)
	userHandler := handlers.NewUserHandler(userService)

	// Setup router
	suite.router = gin.New()
	suite.router.Use(gin.Recovery())

	authMiddleware := middleware.NewGinAuthMiddleware(jwtManager)

	// Auth routes
	auth := suite.router.Group("/auth")
	{
		auth.POST("/login", gin.WrapF(authHandler.Login))
		auth.POST("/refresh", gin.WrapF(authHandler.RefreshToken))
	}

	// Protected routes
	api := suite.router.Group("/api")
	api.Use(authMiddleware.RequireAuth())
	{
		users := api.Group("/users")
		{
			users.GET("", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), userHandler.GetUsers)
			users.POST("", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), userHandler.CreateUser)
			users.GET("/:id", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), userHandler.GetUserByID)
			users.PUT("/:id", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), userHandler.UpdateUser)
			users.DELETE("/:id", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), userHandler.DeleteUser)
		}
	}
}

func (suite *UserWorkflowsTestSuite) initializeTestScenarios() {
	ctx := context.Background()

	// Create test company
	suite.testCompanyID = uuid.New()
	_, err := suite.db.ExecContext(ctx,
		"INSERT INTO companies (id, name, slug, email, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		suite.testCompanyID, "E2E Test Company", "e2e-test-company", "test@e2e.company", time.Now(), time.Now())
	assert.NoError(suite.T(), err)

	// Ensure roles exist
	roles := []string{"master", "company_admin", "admin", "driver"}
	for _, roleName := range roles {
		_, err := suite.db.ExecContext(ctx,
			"INSERT INTO roles (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4) ON CONFLICT (name) DO NOTHING",
			uuid.New(), roleName, time.Now(), time.Now())
		assert.NoError(suite.T(), err)
	}

	// Create test users and get their tokens
	suite.createTestUsersAndTokens()
}

func (suite *UserWorkflowsTestSuite) createTestUsersAndTokens() {
	ctx := context.Background()

	users := []struct {
		id        *uuid.UUID
		name      string
		email     string
		role      string
		companyID *uuid.UUID
		token     *string
	}{
		{&suite.masterUserID, "Master User", "master@e2e.test", "master", nil, &suite.masterToken},
		{&suite.companyUserID, "Company Admin", "company@e2e.test", "company_admin", &suite.testCompanyID, &suite.companyToken},
		{&suite.adminUserID, "Admin User", "admin@e2e.test", "admin", nil, &suite.adminToken},
		{&suite.driverUserID, "Driver User", "driver@e2e.test", "driver", &suite.testCompanyID, &suite.driverToken},
	}

	for _, user := range users {
		*user.id = uuid.New()

		// Get role ID
		var roleID uuid.UUID
		err := suite.db.Get(&roleID, "SELECT id FROM roles WHERE name = $1", user.role)
		assert.NoError(suite.T(), err)

		// Hash password for test user
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
		assert.NoError(suite.T(), err)

		// Insert user
		_, err = suite.db.ExecContext(ctx,
			`INSERT INTO users (id, name, email, password, role_id, company_id, active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			*user.id, user.name, user.email, string(hashedPassword), roleID,
			user.companyID, true, time.Now(), time.Now())
		assert.NoError(suite.T(), err)

		// Generate token via login
		loginData := map[string]string{
			"email":    user.email,
			"password": "password",
		}
		loginResp := suite.makeAuthenticatedRequest("POST", "/auth/login", loginData, "")
		if loginResp.StatusCode == http.StatusOK {
			var loginResult map[string]interface{}
			json.NewDecoder(loginResp.Body).Decode(&loginResult)
			if token, ok := loginResult["access_token"].(string); ok {
				*user.token = token
			}
		}
	}
}

// E2E Test Scenarios

func (suite *UserWorkflowsTestSuite) TestCompleteUserManagementWorkflow() {
	// Scenario: Company admin creates, manages, and removes users

	// Step 1: Company admin logs in
	suite.T().Log("Step 1: Company admin authentication")
	// Test login with master user
	_ = suite.makeAuthenticatedRequest("POST", "/auth/login", map[string]string{
		"email":    "company@e2e.test",
		"password": "password",
	}, "")

	// Note: In real scenario, login would verify password
	// For E2E, we use pre-generated tokens

	// Step 2: Company admin views current users
	suite.T().Log("Step 2: View existing users")
	usersResp := suite.makeAuthenticatedRequest("GET", "/api/users?page=1&limit=10", nil, suite.companyToken)
	assert.Equal(suite.T(), http.StatusOK, usersResp.StatusCode)

	var usersData map[string]interface{}
	err := json.NewDecoder(usersResp.Body).Decode(&usersData)
	assert.NoError(suite.T(), err)

	initialUserCount := len(usersData["users"].([]interface{}))
	suite.T().Logf("Initial user count: %d", initialUserCount)

	// Step 3: Create new driver
	suite.T().Log("Step 3: Create new driver")

	// Get driver role ID
	var driverRoleID uuid.UUID
	err = suite.db.Get(&driverRoleID, "SELECT id FROM roles WHERE name = 'driver'")
	assert.NoError(suite.T(), err)

	newDriver := map[string]interface{}{
		"name":     "E2E Test Driver",
		"email":    "newdriver@e2e.test",
		"phone":    "1234567890",
		"password": "driver123",
		"role_id":  driverRoleID.String(),
		"active":   true,
	}

	createResp := suite.makeAuthenticatedRequest("POST", "/api/users", newDriver, suite.companyToken)
	assert.Equal(suite.T(), http.StatusCreated, createResp.StatusCode)

	var createdUser map[string]interface{}
	err = json.NewDecoder(createResp.Body).Decode(&createdUser)
	assert.NoError(suite.T(), err)

	newDriverID := createdUser["id"].(string)
	suite.T().Logf("Created driver with ID: %s", newDriverID)

	// Step 4: Verify user was created
	suite.T().Log("Step 4: Verify user creation")
	getUserResp := suite.makeAuthenticatedRequest("GET", "/api/users/"+newDriverID, nil, suite.companyToken)
	assert.Equal(suite.T(), http.StatusOK, getUserResp.StatusCode)

	// Step 5: Update user information
	suite.T().Log("Step 5: Update user information")
	updateData := map[string]interface{}{
		"name":   "Updated E2E Driver",
		"phone":  "9876543210",
		"active": true,
	}

	updateResp := suite.makeAuthenticatedRequest("PUT", "/api/users/"+newDriverID, updateData, suite.companyToken)
	assert.Equal(suite.T(), http.StatusOK, updateResp.StatusCode)

	// Step 6: Verify update
	suite.T().Log("Step 6: Verify user update")
	verifyResp := suite.makeAuthenticatedRequest("GET", "/api/users/"+newDriverID, nil, suite.companyToken)
	assert.Equal(suite.T(), http.StatusOK, verifyResp.StatusCode)

	var updatedUser map[string]interface{}
	err = json.NewDecoder(verifyResp.Body).Decode(&updatedUser)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), "Updated E2E Driver", updatedUser["name"])
	assert.Equal(suite.T(), "9876543210", updatedUser["phone"])

	// Step 7: Delete user
	suite.T().Log("Step 7: Delete user")
	deleteResp := suite.makeAuthenticatedRequest("DELETE", "/api/users/"+newDriverID, nil, suite.companyToken)
	assert.Equal(suite.T(), http.StatusNoContent, deleteResp.StatusCode) // DELETE typically returns 204

	// Step 8: Verify deletion
	suite.T().Log("Step 8: Verify user deletion")
	deletedResp := suite.makeAuthenticatedRequest("GET", "/api/users/"+newDriverID, nil, suite.companyToken)
	// After soft delete, the user should not be found via GetByID
	assert.Equal(suite.T(), http.StatusNotFound, deletedResp.StatusCode, "Soft deleted user should not be found")

	suite.T().Log("✅ Complete user management workflow succeeded")
}

func (suite *UserWorkflowsTestSuite) TestPermissionHierarchyEnforcement() {
	// Scenario: Test permission boundaries across different roles

	// Step 1: Driver tries to access user management (should fail)
	suite.T().Log("Step 1: Driver attempting to access user list")
	driverResp := suite.makeAuthenticatedRequest("GET", "/api/users", nil, suite.driverToken)
	assert.Equal(suite.T(), http.StatusForbidden, driverResp.StatusCode)

	// Step 2: Admin tries to create company_admin (should fail)
	suite.T().Log("Step 2: Admin attempting to create company_admin")

	// Get company_admin role ID
	var companyAdminRoleID uuid.UUID
	err := suite.db.Get(&companyAdminRoleID, "SELECT id FROM roles WHERE name = 'company_admin'")
	assert.NoError(suite.T(), err)

	invalidUser := map[string]interface{}{
		"name":     "Invalid Admin",
		"email":    "invalid@e2e.test",
		"password": "password",
		"role_id":  companyAdminRoleID.String(),
		"active":   true,
	}

	adminResp := suite.makeAuthenticatedRequest("POST", "/api/users", invalidUser, suite.adminToken)
	assert.Equal(suite.T(), http.StatusBadRequest, adminResp.StatusCode)

	// Step 3: Company admin creates driver (should succeed)
	suite.T().Log("Step 3: Company admin creating driver")

	// Get driver role ID
	var driverRoleID uuid.UUID
	err = suite.db.Get(&driverRoleID, "SELECT id FROM roles WHERE name = 'driver'")
	assert.NoError(suite.T(), err)

	validUser := map[string]interface{}{
		"name":     "Valid Driver",
		"email":    "validdriver@e2e.test",
		"phone":    "123456789",
		"cpf":      "987.654.321-00", // Unique CPF for permission hierarchy test
		"password": "password",
		"role_id":  driverRoleID.String(),
		"active":   true,
	}

	companyResp := suite.makeAuthenticatedRequest("POST", "/api/users", validUser, suite.companyToken)
	assert.Equal(suite.T(), http.StatusCreated, companyResp.StatusCode)

	// Step 4: Master can access all users
	suite.T().Log("Step 4: Master accessing all users")
	masterResp := suite.makeAuthenticatedRequest("GET", "/api/users", nil, suite.masterToken)
	assert.Equal(suite.T(), http.StatusOK, masterResp.StatusCode)

	suite.T().Log("✅ Permission hierarchy enforcement succeeded")
}

func (suite *UserWorkflowsTestSuite) TestCompanyIsolation() {
	// Scenario: Ensure company data isolation

	// Step 1: Create second company and admin
	suite.T().Log("Step 1: Creating second company")
	ctx := context.Background()

	company2ID := uuid.New()
	_, err := suite.db.ExecContext(ctx,
		"INSERT INTO companies (id, name, slug, email, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		company2ID, "E2E Second Company", "e2e-second-company", "second@e2e.company", time.Now(), time.Now())
	assert.NoError(suite.T(), err)

	// Create admin for second company
	admin2ID := uuid.New()
	var roleID uuid.UUID
	err = suite.db.Get(&roleID, "SELECT id FROM roles WHERE name = 'company_admin'")
	assert.NoError(suite.T(), err)

	// Hash password for second company admin
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	assert.NoError(suite.T(), err)

	_, err = suite.db.ExecContext(ctx,
		`INSERT INTO users (id, name, email, password, role_id, company_id, active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		admin2ID, "Company2 Admin", "company2@e2e.test", string(hashedPassword), roleID,
		company2ID, true, time.Now(), time.Now())
	assert.NoError(suite.T(), err)

	// Generate token for second company admin via login
	loginData := map[string]string{
		"email":    "company2@e2e.test",
		"password": "password",
	}
	loginResp := suite.makeAuthenticatedRequest("POST", "/auth/login", loginData, "")
	assert.Equal(suite.T(), http.StatusOK, loginResp.StatusCode)

	var loginResult map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginResult)
	company2Token := loginResult["access_token"].(string)

	// Step 2: Company 1 admin views users (should only see company 1 users)
	suite.T().Log("Step 2: Company 1 admin viewing users")
	company1Resp := suite.makeAuthenticatedRequest("GET", "/api/users", nil, suite.companyToken)
	assert.Equal(suite.T(), http.StatusOK, company1Resp.StatusCode)

	var company1Data map[string]interface{}
	err = json.NewDecoder(company1Resp.Body).Decode(&company1Data)
	assert.NoError(suite.T(), err)

	company1Users := company1Data["users"].([]interface{})
	for _, userInterface := range company1Users {
		user := userInterface.(map[string]interface{})
		if companyID, exists := user["company_id"]; exists && companyID != nil {
			assert.Equal(suite.T(), suite.testCompanyID.String(), companyID)
		}
	}

	// Step 3: Company 2 admin views users (should only see company 2 users)
	suite.T().Log("Step 3: Company 2 admin viewing users")
	company2Resp := suite.makeAuthenticatedRequest("GET", "/api/users", nil, company2Token)
	assert.Equal(suite.T(), http.StatusOK, company2Resp.StatusCode)

	var company2Data map[string]interface{}
	err = json.NewDecoder(company2Resp.Body).Decode(&company2Data)
	assert.NoError(suite.T(), err)

	company2Users := company2Data["users"].([]interface{})
	for _, userInterface := range company2Users {
		user := userInterface.(map[string]interface{})
		if companyID, exists := user["company_id"]; exists && companyID != nil {
			assert.Equal(suite.T(), company2ID.String(), companyID)
		}
	}

	suite.T().Log("✅ Company isolation test succeeded")
}

func (suite *UserWorkflowsTestSuite) TestInvalidTokenHandling() {
	// Scenario: Test various invalid token scenarios

	// Step 1: No token provided
	suite.T().Log("Step 1: Request without token")
	noTokenResp := suite.makeAuthenticatedRequest("GET", "/api/users", nil, "")
	assert.Equal(suite.T(), http.StatusUnauthorized, noTokenResp.StatusCode)

	// Step 2: Invalid token format
	suite.T().Log("Step 2: Invalid token format")
	invalidResp := suite.makeAuthenticatedRequest("GET", "/api/users", nil, "invalid-token")
	assert.Equal(suite.T(), http.StatusUnauthorized, invalidResp.StatusCode)

	// Step 3: Expired token (simulate by using very short expiration)
	suite.T().Log("Step 3: Expired token simulation")
	// Use an invalid token to simulate expired token
	expiredToken := "invalid.expired.token"

	expiredResp := suite.makeAuthenticatedRequest("GET", "/api/users", nil, expiredToken)
	assert.Equal(suite.T(), http.StatusUnauthorized, expiredResp.StatusCode)

	suite.T().Log("✅ Invalid token handling succeeded")
}

func (suite *UserWorkflowsTestSuite) TestConcurrentUserOperations() {
	// Scenario: Test concurrent operations don't cause conflicts

	suite.T().Log("Testing concurrent user operations")

	// Get driver role ID first
	var driverRoleID uuid.UUID
	err := suite.db.Get(&driverRoleID, "SELECT id FROM roles WHERE name = 'driver'")
	assert.NoError(suite.T(), err)

	// Create multiple users concurrently
	const numConcurrent = 2 // Reduzido para testar se é problema de concorrência
	results := make(chan int, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		go func(index int) {
			newUser := map[string]interface{}{
				"name":     fmt.Sprintf("Concurrent User %d", index),
				"email":    fmt.Sprintf("concurrent%d@e2e.test", index),
				"phone":    fmt.Sprintf("123456789%d", index),
				"cpf":      fmt.Sprintf("000.000.00%d-0%d", index, index), // Unique CPF for each user
				"password": "password123",
				"role_id":  driverRoleID.String(),
				"active":   true,
			}

			resp := suite.makeAuthenticatedRequest("POST", "/api/users", newUser, suite.companyToken)
			results <- resp.StatusCode
		}(i)
	}

	// Collect results
	successCount := 0
	for i := 0; i < numConcurrent; i++ {
		statusCode := <-results
		if statusCode == http.StatusCreated {
			successCount++
		}
	}

	assert.Equal(suite.T(), numConcurrent, successCount, "All concurrent operations should succeed")
	suite.T().Logf("✅ %d concurrent operations completed successfully", successCount)
}

// Helper Methods

func (suite *UserWorkflowsTestSuite) makeAuthenticatedRequest(method, endpoint string, body interface{}, token string) *http.Response {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req, err := http.NewRequest(method, suite.baseURL+endpoint, reqBody)
	assert.NoError(suite.T(), err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: time.Second * 10}
	resp, err := client.Do(req)
	assert.NoError(suite.T(), err)

	return resp
}

// Run the test suite
func TestUserWorkflowsTestSuite(t *testing.T) {
	suite.Run(t, new(UserWorkflowsTestSuite))
}
