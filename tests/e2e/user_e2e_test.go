package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"dashtrack/internal/handlers"
	"dashtrack/internal/models"
	"dashtrack/internal/repository"
	"dashtrack/internal/services"
	"dashtrack/tests/testutils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
)

type UserE2ETestSuite struct {
	suite.Suite
	testDB       *testutils.TestDB
	router       *gin.Engine
	tokenService *services.TokenService
	userService  *services.UserService
	masterToken  string
	adminToken   string
	testCompany  *models.Company
	masterRole   *models.Role
	adminRole    *models.Role
	userRole     *models.Role
}

func TestUserE2ESuite(t *testing.T) {
	suite.Run(t, new(UserE2ETestSuite))
}

func (s *UserE2ETestSuite) SetupSuite() {
	var err error
	s.testDB, err = testutils.SetupTestDB("user_e2e_test")
	s.Require().NoError(err)

	// Initialize repositories
	userRepo := repository.NewUserRepository(s.testDB.DB)
	roleRepo := repository.NewRoleRepository(s.testDB.DB)
	companyRepo := repository.NewCompanyRepository(s.testDB.DB)
	authLogRepo := repository.NewAuthLogRepository(s.testDB.DB)

	// Initialize services
	s.tokenService = services.NewTokenService(
		s.testDB.SqlxDB,
		"test-secret-key",
		15*time.Minute,
		7*24*time.Hour,
	)
	s.userService = services.NewUserService(userRepo, roleRepo, companyRepo, 10)

	// Create test roles
	ctx := context.Background()
	s.masterRole = &models.Role{ID: uuid.New(), Name: "master"}
	s.adminRole = &models.Role{ID: uuid.New(), Name: "admin"}
	s.userRole = &models.Role{ID: uuid.New(), Name: "user"}

	s.Require().NoError(s.testDB.DB.Create(s.masterRole).Error)
	s.Require().NoError(s.testDB.DB.Create(s.adminRole).Error)
	s.Require().NoError(s.testDB.DB.Create(s.userRole).Error)

	// Create test company
	s.testCompany = &models.Company{
		ID:               uuid.New(),
		Name:             "E2E Test Company",
		Slug:             "e2e-test-company",
		Email:            "contact@e2etest.com",
		Phone:            "1234567890",
		SubscriptionPlan: "premium",
	}
	s.Require().NoError(s.testDB.DB.Create(s.testCompany).Error)

	// Create master and admin users
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Test123!"), bcrypt.DefaultCost)

	masterUser := &models.User{
		ID:       uuid.New(),
		Email:    "master@e2e.com",
		Password: string(hashedPassword),
		Name:     "Master User",
		Active:   true,
		RoleID:   s.masterRole.ID,
	}
	s.Require().NoError(s.testDB.DB.Create(masterUser).Error)

	adminUser := &models.User{
		ID:        uuid.New(),
		Email:     "admin@e2e.com",
		Password:  string(hashedPassword),
		Name:      "Admin User",
		Active:    true,
		RoleID:    s.adminRole.ID,
		CompanyID: &s.testCompany.ID,
	}
	s.Require().NoError(s.testDB.DB.Create(adminUser).Error)

	// Generate tokens
	accessToken, _, err := s.tokenService.GenerateTokenPair(ctx, masterUser, "127.0.0.1", "test-agent")
	s.Require().NoError(err)
	s.masterToken = accessToken

	accessToken, _, err = s.tokenService.GenerateTokenPair(ctx, adminUser, "127.0.0.1", "test-agent")
	s.Require().NoError(err)
	s.adminToken = accessToken

	// Setup router
	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	authHandler := handlers.NewAuthHandler(userRepo, authLogRepo, roleRepo, s.tokenService, nil, bcrypt.DefaultCost)
	userHandler := handlers.NewUserHandler(s.userService)
	authMiddleware := handlers.NewAuthMiddleware(s.tokenService, userRepo)

	api := s.router.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.LoginGin)
			auth.POST("/logout", authMiddleware.RequireAuth(), authHandler.LogoutGin)
		}

		api.Use(authMiddleware.RequireAuth())
		{
			users := api.Group("/users")
			{
				users.GET("", userHandler.GetUsers)
				users.POST("", authMiddleware.RequireRole("admin", "master"), userHandler.CreateUser)
				users.GET("/:id", userHandler.GetUserByID)
				users.PUT("/:id", authMiddleware.RequireRole("admin", "master"), userHandler.UpdateUser)
				users.DELETE("/:id", authMiddleware.RequireRole("admin", "master"), userHandler.DeleteUser)
			}
		}
	}
}

func (s *UserE2ETestSuite) TearDownSuite() {
	s.testDB.Close()
}

// TestUserCRUD_E2E tests complete CRUD operations on users
func (s *UserE2ETestSuite) TestUserCRUD_E2E() {
	// CREATE: Master creates a new user
	createReq := models.CreateUserRequest{
		Email:     "crud@e2e.com",
		Password:  "CrudTest123!",
		Name:      "CRUD Test User",
		Phone:     "9999999999",
		CPF:       "11122233344",
		RoleID:    s.userRole.ID,
		CompanyID: &s.testCompany.ID,
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var createdUser models.User
	err := json.Unmarshal(w.Body.Bytes(), &createdUser)
	s.NoError(err)
	s.Equal("crud@e2e.com", createdUser.Email)
	userID := createdUser.ID

	// READ: Get the created user
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s", userID), nil)
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var fetchedUser models.User
	err = json.Unmarshal(w.Body.Bytes(), &fetchedUser)
	s.NoError(err)
	s.Equal(userID, fetchedUser.ID)
	s.Equal("CRUD Test User", fetchedUser.Name)

	// UPDATE: Modify the user
	updateReq := models.UpdateUserRequest{
		Name:  stringPointer("CRUD Updated User"),
		Phone: stringPointer("8888888888"),
	}

	body, _ = json.Marshal(updateReq)
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/users/%s", userID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var updatedUser models.User
	err = json.Unmarshal(w.Body.Bytes(), &updatedUser)
	s.NoError(err)
	s.Equal("CRUD Updated User", updatedUser.Name)
	s.Equal("8888888888", updatedUser.Phone)

	// DELETE: Remove the user
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%s", userID), nil)
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Verify soft delete
	var deletedUser models.User
	err = s.testDB.DB.Unscoped().First(&deletedUser, userID).Error
	s.NoError(err)
	s.NotNil(deletedUser.DeletedAt)
}

// TestUserPermissions_E2E tests permission boundaries
func (s *UserE2ETestSuite) TestUserPermissions_E2E() {
	// Admin tries to create user in their company (should succeed)
	createReq := models.CreateUserRequest{
		Email:     "adminuser@e2e.com",
		Password:  "Admin123!",
		Name:      "Admin Created User",
		Phone:     "7777777777",
		CPF:       "55566677788",
		RoleID:    s.userRole.ID,
		CompanyID: &s.testCompany.ID,
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.adminToken)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var createdUser models.User
	err := json.Unmarshal(w.Body.Bytes(), &createdUser)
	s.NoError(err)

	// Admin tries to list all users (should only see their company)
	req = httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+s.adminToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var listResp services.UserListResponse
	err = json.Unmarshal(w.Body.Bytes(), &listResp)
	s.NoError(err)

	// Verify all users belong to admin's company
	for _, user := range listResp.Users {
		if user.CompanyID != nil {
			s.Equal(s.testCompany.ID, *user.CompanyID)
		}
	}
}

// TestUserSearch_E2E tests user search functionality
func (s *UserE2ETestSuite) TestUserSearch_E2E() {
	// Create multiple test users
	for i := 1; i <= 3; i++ {
		user := &models.User{
			ID:        uuid.New(),
			Email:     fmt.Sprintf("search%d@e2e.com", i),
			Password:  "password",
			Name:      fmt.Sprintf("Search User %d", i),
			Phone:     fmt.Sprintf("666666666%d", i),
			CPF:       fmt.Sprintf("9999999999%d", i),
			Active:    true,
			RoleID:    s.userRole.ID,
			CompanyID: &s.testCompany.ID,
		}
		s.Require().NoError(s.testDB.DB.Create(user).Error)
	}

	// Master searches for all users
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users?page=1&limit=50", nil)
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var listResp services.UserListResponse
	err := json.Unmarshal(w.Body.Bytes(), &listResp)
	s.NoError(err)
	s.GreaterOrEqual(listResp.Total, 5) // At least master, admin, and 3 search users
}

// TestUserValidation_E2E tests input validation
func (s *UserE2ETestSuite) TestUserValidation_E2E() {
	// Try to create user with invalid email
	createReq := models.CreateUserRequest{
		Email:     "invalid-email",
		Password:  "ValidPass123!",
		Name:      "Invalid User",
		Phone:     "1111111111",
		CPF:       "11111111111",
		RoleID:    s.userRole.ID,
		CompanyID: &s.testCompany.ID,
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)

	// Try to create user with weak password
	createReq.Email = "valid@e2e.com"
	createReq.Password = "weak"

	body, _ = json.Marshal(createReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusBadRequest, w.Code)
}

// TestConcurrentUserOperations_E2E tests concurrent access patterns
func (s *UserE2ETestSuite) TestConcurrentUserOperations_E2E() {
	// Create a test user
	testUser := &models.User{
		ID:        uuid.New(),
		Email:     "concurrent@e2e.com",
		Password:  "password",
		Name:      "Concurrent Test User",
		Phone:     "3333333333",
		CPF:       "33333333333",
		Active:    true,
		RoleID:    s.userRole.ID,
		CompanyID: &s.testCompany.ID,
	}
	s.Require().NoError(s.testDB.DB.Create(testUser).Error)

	// Master and Admin both try to update the same user
	updateReq1 := models.UpdateUserRequest{
		Name: stringPointer("Updated by Master"),
	}

	updateReq2 := models.UpdateUserRequest{
		Phone: stringPointer("4444444444"),
	}

	// First update by master
	body1, _ := json.Marshal(updateReq1)
	req1 := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/users/%s", testUser.ID), bytes.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", "Bearer "+s.masterToken)
	w1 := httptest.NewRecorder()
	s.router.ServeHTTP(w1, req1)

	s.Equal(http.StatusOK, w1.Code)

	// Second update by admin
	body2, _ := json.Marshal(updateReq2)
	req2 := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/users/%s", testUser.ID), bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+s.adminToken)
	w2 := httptest.NewRecorder()
	s.router.ServeHTTP(w2, req2)

	s.Equal(http.StatusOK, w2.Code)

	// Verify final state
	var finalUser models.User
	err := s.testDB.DB.First(&finalUser, testUser.ID).Error
	s.NoError(err)
	s.Equal("Updated by Master", finalUser.Name)
	s.Equal("4444444444", finalUser.Phone)
}

// Helper function
func stringPointer(s string) *string {
	return &s
}
