package integration_test

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
	"github.com/paulochiaradia/dashtrack/internal/handlers"
	"github.com/paulochiaradia/dashtrack/internal/middleware"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
	"github.com/paulochiaradia/dashtrack/internal/services"
)

// UserIntegrationTestSuite defines the integration test suite
type UserIntegrationTestSuite struct {
	suite.Suite
	db             *sqlx.DB
	router         *gin.Engine
	userRepository *repository.UserRepository
	roleRepository *repository.RoleRepository
	userService    *services.UserService
	authHandler    *handlers.AuthHandler
	userHandler    *handlers.UserHandler
	jwtManager     *auth.JWTManager

	// Test data
	masterUser    *models.User
	companyUser   *models.User
	adminUser     *models.User
	driverUser    *models.User
	testCompanyID uuid.UUID
	authTokens    map[string]string
}

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

func (suite *UserIntegrationTestSuite) SetupSuite() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Setup test database
	suite.setupTestDatabase()

	// Initialize repositories
	suite.userRepository = repository.NewUserRepository(suite.db)
	suite.roleRepository = repository.NewRoleRepository(suite.db.DB)

	// Initialize services
	suite.userService = services.NewUserService(suite.userRepository, suite.roleRepository, 10)

	// Initialize JWT manager
	suite.jwtManager = auth.NewJWTManager("test-secret-key", time.Hour, 24*time.Hour, "dashtrack-test")

	// Initialize handlers - creating mock auth log repository
	authLogRepo := &mockAuthLogRepository{}
	tokenService := services.NewTokenService(suite.db, "test-secret-key", time.Hour, time.Hour*24)
	suite.authHandler = handlers.NewAuthHandler(suite.userRepository, authLogRepo, suite.jwtManager, tokenService, 10)
	suite.userHandler = handlers.NewUserHandler(suite.userService)

	// Setup router
	suite.setupRouter()

	// Initialize test data
	suite.initializeTestData()
}

func (suite *UserIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		suite.cleanupTestData()
		suite.db.Close()
	}
}

func (suite *UserIntegrationTestSuite) setupTestDatabase() {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://dashtrack_user:dashtrack_password@localhost:5433/dashtrack_test_integration?sslmode=disable"
	}

	var err error
	suite.db, err = sqlx.Connect("postgres", dbURL)
	if err != nil {
		suite.T().Skipf("Cannot connect to test database: %v", err)
		return
	}

	// Run migrations
	driver, err := postgres.WithInstance(suite.db.DB, &postgres.Config{})
	if err != nil {
		suite.T().Fatalf("Failed to create migrate driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"postgres",
		driver,
	)
	if err != nil {
		suite.T().Fatalf("Failed to create migrate instance: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		suite.T().Fatalf("Failed to run migrations: %v", err)
	}
}

func (suite *UserIntegrationTestSuite) setupRouter() {
	suite.router = gin.New()

	// Add middleware
	authMiddleware := middleware.NewGinAuthMiddleware(suite.jwtManager)

	// Auth routes
	auth := suite.router.Group("/auth")
	{
		auth.POST("/login", gin.WrapF(suite.authHandler.Login))
	}

	// Protected routes
	api := suite.router.Group("/api")
	api.Use(authMiddleware.RequireAuth())
	{
		users := api.Group("/users")
		{
			users.GET("", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), suite.userHandler.GetUsers)
			users.POST("", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), suite.userHandler.CreateUser)
			users.PUT("/:id", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), suite.userHandler.UpdateUser)
			users.DELETE("/:id", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), suite.userHandler.DeleteUser)
		}
	}
}

func (suite *UserIntegrationTestSuite) initializeTestData() {
	ctx := context.Background()
	suite.authTokens = make(map[string]string)

	// Create test company
	suite.testCompanyID = uuid.New()
	_, err := suite.db.ExecContext(ctx,
		"INSERT INTO companies (id, name, slug, email, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		suite.testCompanyID, "Test Company", "test-company", "test@testcompany.com", time.Now(), time.Now())
	assert.NoError(suite.T(), err)

	// Create roles if they don't exist
	roles := []struct {
		name string
		id   uuid.UUID
	}{
		{"master", uuid.New()},
		{"company_admin", uuid.New()},
		{"admin", uuid.New()},
		{"driver", uuid.New()},
	}

	for _, role := range roles {
		_, err := suite.db.ExecContext(ctx,
			"INSERT INTO roles (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4) ON CONFLICT (name) DO NOTHING",
			role.id, role.name, time.Now(), time.Now())
		assert.NoError(suite.T(), err)
	}

	// Create test users
	suite.createTestUsers(ctx)

	// Generate auth tokens
	suite.generateAuthTokens()
}

func (suite *UserIntegrationTestSuite) createTestUsers(ctx context.Context) {
	// Get role IDs
	masterRoleID := suite.getRoleID("master")
	companyRoleID := suite.getRoleID("company_admin")
	adminRoleID := suite.getRoleID("admin")
	driverRoleID := suite.getRoleID("driver")

	// Hash password for test users
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), 10)
	assert.NoError(suite.T(), err)

	// Create master user
	suite.masterUser = &models.User{
		ID:       uuid.New(),
		Name:     "Master User",
		Email:    "master@dashtrack.com",
		Password: string(hashedPassword),
		Active:   true,
		Role:     &models.Role{ID: masterRoleID, Name: "master"},
	}

	// Create company admin user
	suite.companyUser = &models.User{
		ID:        uuid.New(),
		Name:      "Company Admin",
		Email:     "company@example.com",
		Password:  string(hashedPassword),
		Active:    true,
		CompanyID: &suite.testCompanyID,
		Role:      &models.Role{ID: companyRoleID, Name: "company_admin"},
	}

	// Create admin user
	suite.adminUser = &models.User{
		ID:        uuid.New(),
		Name:      "Admin User",
		Email:     "admin@example.com",
		Password:  string(hashedPassword),
		Active:    true,
		CompanyID: &suite.testCompanyID,
		Role:      &models.Role{ID: adminRoleID, Name: "admin"},
	}

	// Create driver user
	suite.driverUser = &models.User{
		ID:        uuid.New(),
		Name:      "Driver User",
		Email:     "driver@example.com",
		Password:  string(hashedPassword),
		Active:    true,
		CompanyID: &suite.testCompanyID,
		Role:      &models.Role{ID: driverRoleID, Name: "driver"},
	}

	// Insert users into database
	users := []*models.User{suite.masterUser, suite.companyUser, suite.adminUser, suite.driverUser}
	for _, user := range users {
		_, err := suite.db.ExecContext(ctx,
			`INSERT INTO users (id, name, email, password, role_id, company_id, active, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			user.ID, user.Name, user.Email, user.Password, user.Role.ID,
			user.CompanyID, user.Active, time.Now(), time.Now())
		assert.NoError(suite.T(), err)
	}
}

func (suite *UserIntegrationTestSuite) getRoleID(roleName string) uuid.UUID {
	var roleID uuid.UUID
	err := suite.db.Get(&roleID, "SELECT id FROM roles WHERE name = $1", roleName)
	assert.NoError(suite.T(), err)
	return roleID
}

func (suite *UserIntegrationTestSuite) generateAuthTokens() {
	users := map[string]*models.User{
		"master":        suite.masterUser,
		"company_admin": suite.companyUser,
		"admin":         suite.adminUser,
		"driver":        suite.driverUser,
	}

	for role, user := range users {
		userContext := auth.UserContext{
			UserID:   user.ID,
			Email:    user.Email,
			Name:     user.Name,
			RoleID:   user.RoleID,
			RoleName: user.Role.Name,
			TenantID: user.CompanyID,
		}
		token, _, err := suite.jwtManager.GenerateTokens(userContext)
		assert.NoError(suite.T(), err)
		suite.authTokens[role] = token
	}
}

func (suite *UserIntegrationTestSuite) cleanupTestData() {
	ctx := context.Background()

	// Clean up test data
	_, _ = suite.db.ExecContext(ctx, "DELETE FROM users WHERE email LIKE '%@dashtrack.com' OR email LIKE '%@example.com'")
	_, _ = suite.db.ExecContext(ctx, "DELETE FROM companies WHERE name = 'Test Company'")
}

// Test Cases

func (suite *UserIntegrationTestSuite) TestLoginFlow_Success() {
	loginData := map[string]string{
		"email":    suite.masterUser.Email,
		"password": "password", // Assume this matches the hashed password
	}

	body, _ := json.Marshal(loginData)
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Should return success with token (assuming password verification is mocked or simplified)
	// In real scenario, you'd need to hash the test password properly
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *UserIntegrationTestSuite) TestGetUsers_MasterUser_Success() {
	req, _ := http.NewRequest("GET", "/api/users", nil)
	req.Header.Set("Authorization", "Bearer "+suite.authTokens["master"])

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	users, ok := response["users"].([]interface{})
	assert.True(suite.T(), ok)
	assert.Greater(suite.T(), len(users), 0)
}

func (suite *UserIntegrationTestSuite) TestGetUsers_CompanyAdmin_LimitedAccess() {
	req, _ := http.NewRequest("GET", "/api/users", nil)
	req.Header.Set("Authorization", "Bearer "+suite.authTokens["company_admin"])

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	users, ok := response["users"].([]interface{})
	assert.True(suite.T(), ok)

	// Company admin should only see users from their company
	for _, userInterface := range users {
		user := userInterface.(map[string]interface{})
		if companyID, exists := user["company_id"]; exists && companyID != nil {
			assert.Equal(suite.T(), suite.testCompanyID.String(), companyID)
		}
	}
}

func (suite *UserIntegrationTestSuite) TestCreateUser_CompanyAdmin_Success() {
	driverRoleID := suite.getRoleID("driver")

	newUser := map[string]interface{}{
		"name":     "New Driver",
		"email":    "newdriver@example.com",
		"phone":    "1234567890",
		"password": "password123",
		"role_id":  driverRoleID.String(),
	}

	body, _ := json.Marshal(newUser)
	req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authTokens["company_admin"])

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var user map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &user)
	assert.NoError(suite.T(), err)

	assert.Equal(suite.T(), newUser["name"], user["name"])
	assert.Equal(suite.T(), newUser["email"], user["email"])
	assert.Equal(suite.T(), suite.testCompanyID.String(), user["company_id"])
}

func (suite *UserIntegrationTestSuite) TestCreateUser_Admin_CannotCreateCompanyAdmin() {
	companyAdminRoleID := suite.getRoleID("company_admin")

	newUser := map[string]interface{}{
		"name":     "Unauthorized Admin",
		"email":    "unauthorized@example.com",
		"password": "password123",
		"role_id":  companyAdminRoleID.String(), // Admin trying to create company_admin
	}

	body, _ := json.Marshal(newUser)
	req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authTokens["admin"])

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *UserIntegrationTestSuite) TestUpdateUser_Success() {
	updateData := map[string]interface{}{
		"name":   "Updated Driver Name",
		"phone":  "9876543210",
		"active": true,
	}

	body, _ := json.Marshal(updateData)
	req, _ := http.NewRequest("PUT", "/api/users/"+suite.driverUser.ID.String(), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+suite.authTokens["admin"])

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Response is the user directly, not wrapped in a user object
	assert.Equal(suite.T(), updateData["name"], response["name"])
	assert.Equal(suite.T(), updateData["phone"], response["phone"])
}

func (suite *UserIntegrationTestSuite) TestDeleteUser_Success() {
	req, _ := http.NewRequest("DELETE", "/api/users/"+suite.driverUser.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+suite.authTokens["admin"])

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNoContent, w.Code)

	// Verify user is soft deleted
	var deletedAt *time.Time
	err := suite.db.Get(&deletedAt, "SELECT deleted_at FROM users WHERE id = $1", suite.driverUser.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), deletedAt)
}

func (suite *UserIntegrationTestSuite) TestUnauthorizedAccess_DriverRole() {
	req, _ := http.NewRequest("GET", "/api/users", nil)
	req.Header.Set("Authorization", "Bearer "+suite.authTokens["driver"])

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *UserIntegrationTestSuite) TestInvalidToken_Unauthorized() {
	req, _ := http.NewRequest("GET", "/api/users", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

// Run the test suite
func TestUserIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserIntegrationTestSuite))
}
