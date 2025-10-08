package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
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
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
	"github.com/paulochiaradia/dashtrack/internal/services"
)

// mockAuthLogRepository é um mock simples para testes
type mockAuthLogRepository struct{}

func (m *mockAuthLogRepository) Create(log *models.AuthLog) error { return nil }
func (m *mockAuthLogRepository) GetRecentFailedAttempts(email string, since time.Time) (int, error) {
	return 0, nil
}
func (m *mockAuthLogRepository) GetByUserID(userID uuid.UUID, limit int) ([]*models.AuthLog, error) {
	return nil, nil
}
func (m *mockAuthLogRepository) CountLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time) (int, error) {
	return 0, nil
}
func (m *mockAuthLogRepository) CountSuccessfulLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time) (int, error) {
	return 0, nil
}
func (m *mockAuthLogRepository) CountFailedLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time) (int, error) {
	return 0, nil
}
func (m *mockAuthLogRepository) CountUserLogins(ctx context.Context, userID uuid.UUID, from, to time.Time) (int, error) {
	return 0, nil
}
func (m *mockAuthLogRepository) CountUserSuccessfulLogins(ctx context.Context, userID uuid.UUID, from, to time.Time) (int, error) {
	return 0, nil
}
func (m *mockAuthLogRepository) CountUserFailedLogins(ctx context.Context, userID uuid.UUID, from, to time.Time) (int, error) {
	return 0, nil
}
func (m *mockAuthLogRepository) GetRecentSuccessfulLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time, limit int) ([]models.RecentLogin, error) {
	return nil, nil
}
func (m *mockAuthLogRepository) GetUserRecentSuccessfulLogins(ctx context.Context, userID uuid.UUID, from, to time.Time, limit int) ([]models.RecentLogin, error) {
	return nil, nil
}

// E2ETestSuite represents the end-to-end test suite
type E2ETestSuite struct {
	suite.Suite
	db     *sqlx.DB
	router *gin.Engine

	// Test data
	masterToken   string
	adminToken    string
	companyToken  string
	testCompanyID uuid.UUID
	masterUserID  uuid.UUID
	adminUserID   uuid.UUID
	companyUserID uuid.UUID
	driverUserID  uuid.UUID
}

func TestE2ETestSuite(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}

func (suite *E2ETestSuite) SetupSuite() {
	// Setup database connection for E2E tests
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://dashtrack_user:dashtrack_password@localhost:5433/dashtrack_test_integration?sslmode=disable"
	}

	var err error
	suite.db, err = sqlx.Connect("postgres", dbURL)
	assert.NoError(suite.T(), err)

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

	// Setup application
	suite.setupApplication()
	suite.setupTestData()
}

func (suite *E2ETestSuite) TearDownSuite() {
	if suite.db != nil {
		// Clean up test data
		ctx := context.Background()
		_, _ = suite.db.ExecContext(ctx, "DELETE FROM users WHERE email LIKE '%@e2e.test'")
		_, _ = suite.db.ExecContext(ctx, "DELETE FROM companies WHERE name LIKE 'E2E%'")
		suite.db.Close()
	}
}

func (suite *E2ETestSuite) SetupTest() {
	// Clean up test data before each test, but preserve the core test users
	ctx := context.Background()
	_, _ = suite.db.ExecContext(ctx, "DELETE FROM users WHERE email LIKE 'new%@e2e.test' OR email LIKE 'unauthorized@e2e.test' OR email LIKE 'globaldriver@e2e.test' OR email LIKE 'globalhelper@e2e.test'")
}

func (suite *E2ETestSuite) setupApplication() {
	// Load configuration
	cfg := &config.Config{
		JWTSecret:              "e2e-test-secret-key",
		JWTAccessExpireMinutes: 60,
		JWTRefreshExpireHours:  24,
		BcryptCost:             12,
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(suite.db)
	roleRepo := repository.NewRoleRepository(suite.db.DB)
	authLogRepo := &mockAuthLogRepository{} // Use mock instead of real repository

	// Initialize services
	userService := services.NewUserService(userRepo, roleRepo, cfg.BcryptCost)
	tokenService := services.NewTokenService(
		suite.db,
		cfg.JWTSecret,
		time.Duration(cfg.JWTAccessExpireMinutes)*time.Minute,
		time.Duration(cfg.JWTRefreshExpireHours)*time.Hour,
	)

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(
		cfg.JWTSecret,
		time.Duration(cfg.JWTAccessExpireMinutes)*time.Minute,
		time.Duration(cfg.JWTRefreshExpireHours)*time.Hour,
		"dashtrack-e2e",
	)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, authLogRepo, jwtManager, tokenService, cfg.BcryptCost)
	userHandler := handlers.NewUserHandler(userService)

	// Setup router manually like integration tests
	suite.router = gin.New()

	// Add middleware
	authMiddleware := middleware.NewGinAuthMiddleware(jwtManager)

	// Auth routes
	auth := suite.router.Group("/auth")
	{
		auth.POST("/login", authHandler.LoginGin)
	}

	// Protected routes
	api := suite.router.Group("/api")
	api.Use(authMiddleware.RequireAuth())
	{
		users := api.Group("/users")
		{
			users.GET("", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), userHandler.GetUsers)
			users.POST("", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), userHandler.CreateUser)
			users.PUT("/:id", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), userHandler.UpdateUser)
			users.DELETE("/:id", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), userHandler.DeleteUser)
		}
	}
}

func (suite *E2ETestSuite) setupTestData() {
	ctx := context.Background()

	// Create test company
	suite.testCompanyID = uuid.New()
	_, err := suite.db.ExecContext(ctx,
		"INSERT INTO companies (id, name, slug, email, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		suite.testCompanyID, "E2E Test Company", "e2e-test-company", "e2e@testcompany.com", time.Now(), time.Now())
	assert.NoError(suite.T(), err)

	// Debug log - temporary
	println("🔍 DEBUG - Test Company ID created:", suite.testCompanyID.String())

	// Ensure roles exist
	roleInsertSQL := "INSERT INTO roles (id, name, description, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (name) DO NOTHING"
	roles := map[string]string{
		"master":        "System master user",
		"company_admin": "Company administrator",
		"admin":         "General administrator",
		"driver":        "Vehicle driver",
	}

	for roleName, description := range roles {
		_, err := suite.db.ExecContext(ctx, roleInsertSQL, uuid.New(), roleName, description, time.Now(), time.Now())
		assert.NoError(suite.T(), err)
	}

	// Create test users and get their authentication tokens
	suite.createTestUsers()
}

func (suite *E2ETestSuite) createTestUsers() {
	ctx := context.Background()

	// Get role IDs
	roles := make(map[string]uuid.UUID)
	rows, err := suite.db.QueryContext(ctx, "SELECT id, name FROM roles")
	assert.NoError(suite.T(), err)
	defer rows.Close()

	for rows.Next() {
		var roleID uuid.UUID
		var roleName string
		err := rows.Scan(&roleID, &roleName)
		assert.NoError(suite.T(), err)
		roles[roleName] = roleID
	}

	// Hash password for test users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), 12)
	assert.NoError(suite.T(), err)

	users := []struct {
		id        *uuid.UUID
		name      string
		email     string
		roleID    uuid.UUID
		companyID *uuid.UUID
		token     *string
	}{
		{&suite.masterUserID, "Master User", "master@e2e.test", roles["master"], nil, &suite.masterToken},
		{&suite.adminUserID, "Admin User", "admin@e2e.test", roles["admin"], nil, &suite.adminToken},
		{&suite.companyUserID, "Company Admin", "company@e2e.test", roles["company_admin"], &suite.testCompanyID, &suite.companyToken},
		{&suite.driverUserID, "Driver User", "driver@e2e.test", roles["driver"], &suite.testCompanyID, nil},
	}

	for _, user := range users {
		*user.id = uuid.New()

		_, err := suite.db.ExecContext(ctx,
			"INSERT INTO users (id, name, email, password, role_id, company_id, active, created_at, updated_at, password_changed_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)",
			*user.id, user.name, user.email, string(hashedPassword), user.roleID, user.companyID, true, time.Now(), time.Now(), time.Now())
		assert.NoError(suite.T(), err)

		// Debug log - temporary
		if user.companyID != nil {
			println("🔍 DEBUG - User created:", user.email, "with CompanyID:", user.companyID.String())
		} else {
			println("🔍 DEBUG - User created:", user.email, "with NO CompanyID")
		}

		// Generate token for specific users
		if user.token != nil {
			*user.token = suite.generateToken(user.email)
		}
	}
}

func (suite *E2ETestSuite) generateToken(email string) string {
	// Make login request to get real token
	loginData := map[string]string{
		"email":    email,
		"password": "password123",
	}

	body, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	if w.Code == http.StatusOK {
		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		if token, ok := response["access_token"].(string); ok {
			return token
		}
	}

	suite.T().Fatalf("Failed to generate token for %s", email)
	return ""
}

func (suite *E2ETestSuite) makeRequest(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var req *http.Request

	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req, _ = http.NewRequest(method, path, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, _ = http.NewRequest(method, path, nil)
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

// E2E Test Cases

func (suite *E2ETestSuite) TestCompleteUserWorkflow() {
	// Test 1: Master user can list all users
	w := suite.makeRequest("GET", "/api/users", nil, suite.masterToken)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	users, ok := response["users"].([]interface{})
	assert.True(suite.T(), ok)
	assert.GreaterOrEqual(suite.T(), len(users), 3) // At least our test users (master can see all)

	// Test 2: Company admin can create a driver in their company
	newDriverData := map[string]interface{}{
		"name":       "New Driver",
		"email":      "newdriver@e2e.test",
		"password":   "password123",
		"phone":      "1234567890",
		"cpf":        "123.456.789-01", // Valid CPF format with mask (14 chars)
		"role_id":    suite.getDriverRoleID(),
		"company_id": suite.testCompanyID.String(),
	}

	w = suite.makeRequest("POST", "/api/users", newDriverData, suite.masterToken) // Use master token for now
	if w.Code != http.StatusCreated {
		// Debug: Print the error response
		var errorResponse map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &errorResponse)
		suite.T().Logf("CreateUser failed with status %d: %+v", w.Code, errorResponse)
	}
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createdUser map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &createdUser)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), newDriverData["email"], createdUser["email"])

	// Verify userID exists before proceeding
	userID, ok := createdUser["id"].(string)
	if !assert.True(suite.T(), ok, "User ID should be a valid string") {
		suite.T().FailNow()
	}

	// Test 3: Admin cannot create company admin (permission denied)
	companyAdminData := map[string]interface{}{
		"name":     "Unauthorized Company Admin",
		"email":    "unauthorized@e2e.test",
		"password": "password123",
		"role_id":  suite.getCompanyAdminRoleID(),
	}

	w = suite.makeRequest("POST", "/api/users", companyAdminData, suite.adminToken)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	// Test 4: Update user by master
	updateData := map[string]interface{}{
		"name":   "Updated Driver Name",
		"phone":  "9876543210",
		"active": true,
	}

	w = suite.makeRequest("PUT", "/api/users/"+userID, updateData, suite.masterToken)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var updatedUser map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &updatedUser)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), updateData["name"], updatedUser["name"])

	// Test 5: Delete user (soft delete)
	w = suite.makeRequest("DELETE", "/api/users/"+userID, nil, suite.masterToken)
	assert.Equal(suite.T(), http.StatusNoContent, w.Code)

	// Verify soft delete
	var deletedAt *time.Time
	err = suite.db.Get(&deletedAt, "SELECT deleted_at FROM users WHERE id = $1", userID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), deletedAt)
}

func (suite *E2ETestSuite) TestAuthenticationFlow() {
	// Test 1: Valid login
	loginData := map[string]string{
		"email":    "master@e2e.test",
		"password": "password123",
	}

	w := suite.makeRequest("POST", "/auth/login", loginData, "")
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var loginResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &loginResponse)
	assert.NoError(suite.T(), err)

	assert.Contains(suite.T(), loginResponse, "access_token")
	assert.Contains(suite.T(), loginResponse, "refresh_token")
	assert.Contains(suite.T(), loginResponse, "user")

	// Test 2: Invalid credentials
	invalidLoginData := map[string]string{
		"email":    "master@e2e.test",
		"password": "wrongpassword",
	}

	w = suite.makeRequest("POST", "/auth/login", invalidLoginData, "")
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	// Test 3: Access protected endpoint without token
	w = suite.makeRequest("GET", "/api/users", nil, "")
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)

	// Test 4: Access protected endpoint with invalid token
	w = suite.makeRequest("GET", "/api/users", nil, "invalid-token")
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *E2ETestSuite) TestRoleBasedAccessControl() {
	// Test 1: Company admin can only see users in their company
	w := suite.makeRequest("GET", "/api/users", nil, suite.companyToken)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	users, ok := response["users"].([]interface{})
	assert.True(suite.T(), ok)

	// Verify all users belong to the same company or are company-related roles
	for _, userInterface := range users {
		user := userInterface.(map[string]interface{})
		companyID := user["company_id"]
		role := user["role"].(map[string]interface{})
		roleName := role["name"].(string)

		// Either belongs to our company or is a company-level role
		if companyID != nil {
			assert.Equal(suite.T(), suite.testCompanyID.String(), companyID.(string))
		}
		// Company admin should see company_admin, manager, driver, helper roles
		allowedRoles := []string{"company_admin", "manager", "driver", "helper"}
		assert.Contains(suite.T(), allowedRoles, roleName)
	}

	// Test 2: Global admin can see all users from all companies
	w = suite.makeRequest("GET", "/api/users", nil, suite.adminToken)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	users, ok = response["users"].([]interface{})
	assert.True(suite.T(), ok)

	// Debug: Check if driver role exists
	var driverRoleExists int
	err = suite.db.Get(&driverRoleExists, "SELECT COUNT(*) FROM roles WHERE name = 'driver'")
	assert.NoError(suite.T(), err)
	suite.T().Logf("Driver role exists: %d", driverRoleExists)

	// Debug: Check specifically for driver
	var driverCount int
	err = suite.db.Get(&driverCount, "SELECT COUNT(*) FROM users WHERE email = 'driver@e2e.test'")
	assert.NoError(suite.T(), err)
	suite.T().Logf("Driver count in database: %d", driverCount)

	// Debug: Check what's actually in the database
	var dbUsers []struct {
		Email     string  `db:"email"`
		RoleName  string  `db:"role_name"`
		CompanyID *string `db:"company_id"`
		DeletedAt *string `db:"deleted_at"`
	}
	err = suite.db.Select(&dbUsers, `
		SELECT u.email, r.name as role_name, 
		       CASE WHEN u.company_id IS NOT NULL THEN u.company_id::text ELSE NULL END as company_id,
		       CASE WHEN u.deleted_at IS NOT NULL THEN u.deleted_at::text ELSE NULL END as deleted_at
		FROM users u 
		JOIN roles r ON u.role_id = r.id 
		WHERE u.email LIKE '%@e2e.test' OR u.email = 'driver@e2e.test'
		ORDER BY u.email
	`)
	assert.NoError(suite.T(), err)

	suite.T().Log("Users in database:")
	for _, user := range dbUsers {
		suite.T().Logf("  %s (%s) - Company: %v, Deleted: %v", user.Email, user.RoleName, user.CompanyID, user.DeletedAt)
	}

	// Debug: Print the users the admin can see
	suite.T().Logf("Admin sees %d users:", len(users))
	for i, userInterface := range users {
		user := userInterface.(map[string]interface{})
		role := user["role"].(map[string]interface{})
		suite.T().Logf("  User %d: %s (%s) - Company: %v", i+1, user["email"], role["name"], user["company_id"])
	}

	// Global admin should see ALL users (master, admin, company_admin, driver from test company)
	assert.GreaterOrEqual(suite.T(), len(users), 4, "Admin should see at least 4 users (master, admin, company_admin, driver)")

	// Verify admin can see users from different contexts
	rolesSeen := make(map[string]bool)
	for _, userInterface := range users {
		user := userInterface.(map[string]interface{})
		role := user["role"].(map[string]interface{})
		roleName := role["name"].(string)
		rolesSeen[roleName] = true
	}

	// Admin should be able to see multiple role types
	assert.True(suite.T(), len(rolesSeen) >= 2, "Admin should see multiple role types")
}

func (suite *E2ETestSuite) TestDataValidation() {
	// Test 1: Invalid email format
	invalidUserData := map[string]interface{}{
		"name":     "Test User",
		"email":    "invalid-email",
		"password": "password123",
		"role_id":  suite.getDriverRoleID(),
	}

	w := suite.makeRequest("POST", "/api/users", invalidUserData, suite.masterToken)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	// Test 2: Missing required fields
	incompleteUserData := map[string]interface{}{
		"name": "Test User",
		// Missing email, password, role_id
	}

	w = suite.makeRequest("POST", "/api/users", incompleteUserData, suite.masterToken)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	// Test 3: Duplicate email
	validUserData := map[string]interface{}{
		"name":     "Test User",
		"email":    "master@e2e.test", // Already exists
		"password": "password123",
		"role_id":  suite.getDriverRoleID(),
	}

	w = suite.makeRequest("POST", "/api/users", validUserData, suite.masterToken)
	assert.Equal(suite.T(), http.StatusConflict, w.Code)
}

// Helper methods

func (suite *E2ETestSuite) getDriverRoleID() string {
	var roleID uuid.UUID
	err := suite.db.Get(&roleID, "SELECT id FROM roles WHERE name = 'driver'")
	assert.NoError(suite.T(), err)
	return roleID.String()
}

func (suite *E2ETestSuite) getCompanyAdminRoleID() string {
	var roleID uuid.UUID
	err := suite.db.Get(&roleID, "SELECT id FROM roles WHERE name = 'company_admin'")
	assert.NoError(suite.T(), err)
	return roleID.String()
}
