package integration_test

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

type UserManagementTestSuite struct {
	suite.Suite
	testDB       *testutils.TestDB
	router       *gin.Engine
	tokenService *services.TokenService
	userService  *services.UserService
	userRepo     *repository.UserRepository
	roleRepo     *repository.RoleRepository
	companyRepo  *repository.CompanyRepository
	masterUser   *models.User
	masterToken  string
	adminUser    *models.User
	adminToken   string
	regularUser  *models.User
	regularToken string
	testCompany  *models.Company
	masterRole   *models.Role
	adminRole    *models.Role
	userRole     *models.Role
}

func TestUserManagementSuite(t *testing.T) {
	suite.Run(t, new(UserManagementTestSuite))
}

func (s *UserManagementTestSuite) SetupSuite() {
	var err error
	s.testDB, err = testutils.SetupTestDB("user_management_test")
	s.Require().NoError(err)

	// Initialize repositories
	s.userRepo = repository.NewUserRepository(s.testDB.DB)
	s.roleRepo = repository.NewRoleRepository(s.testDB.DB)
	s.companyRepo = repository.NewCompanyRepository(s.testDB.DB)

	// Initialize services
	s.tokenService = services.NewTokenService(
		s.testDB.SqlxDB,
		"test-secret-key",
		15*time.Minute,
		7*24*time.Hour,
	)
	s.userService = services.NewUserService(s.userRepo, s.roleRepo, s.companyRepo, 10)

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
		ID:   uuid.New(),
		Name: "Test Company",
	}
	s.Require().NoError(s.testDB.DB.Create(s.testCompany).Error)

	// Create master user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Master123!"), bcrypt.DefaultCost)
	s.masterUser = &models.User{
		ID:       uuid.New(),
		Email:    "master@test.com",
		Password: string(hashedPassword),
		Name:     "Master User",
		Active:   true,
		RoleID:   s.masterRole.ID,
	}
	s.Require().NoError(s.testDB.DB.Create(s.masterUser).Error)

	// Create admin user
	s.adminUser = &models.User{
		ID:        uuid.New(),
		Email:     "admin@test.com",
		Password:  string(hashedPassword),
		Name:      "Admin User",
		Active:    true,
		RoleID:    s.adminRole.ID,
		CompanyID: &s.testCompany.ID,
	}
	s.Require().NoError(s.testDB.DB.Create(s.adminUser).Error)

	// Create regular user
	s.regularUser = &models.User{
		ID:        uuid.New(),
		Email:     "user@test.com",
		Password:  string(hashedPassword),
		Name:      "Regular User",
		Active:    true,
		RoleID:    s.userRole.ID,
		CompanyID: &s.testCompany.ID,
	}
	s.Require().NoError(s.testDB.DB.Create(s.regularUser).Error)

	// Generate tokens for each user
	accessToken, _, err := s.tokenService.GenerateTokenPair(ctx, s.masterUser, "127.0.0.1", "test-agent")
	s.Require().NoError(err)
	s.masterToken = accessToken

	accessToken, _, err = s.tokenService.GenerateTokenPair(ctx, s.adminUser, "127.0.0.1", "test-agent")
	s.Require().NoError(err)
	s.adminToken = accessToken

	accessToken, _, err = s.tokenService.GenerateTokenPair(ctx, s.regularUser, "127.0.0.1", "test-agent")
	s.Require().NoError(err)
	s.regularToken = accessToken

	// Setup router
	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	userHandler := handlers.NewUserHandler(s.userService)
	authMiddleware := handlers.NewAuthMiddleware(s.tokenService, s.userRepo)

	api := s.router.Group("/api/v1")
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

func (s *UserManagementTestSuite) TearDownSuite() {
	s.testDB.Close()
}

// TestGetUsers_Master tests that master can see all users
func (s *UserManagementTestSuite) TestGetUsers_Master() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response services.UserListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	s.GreaterOrEqual(response.Total, 3) // At least 3 users (master, admin, regular)
}

// TestGetUsers_Admin tests that admin can only see users from their company
func (s *UserManagementTestSuite) TestGetUsers_Admin() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+s.adminToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var response services.UserListResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	s.NoError(err)
	// Admin should only see users from their company
	for _, user := range response.Users {
		if user.CompanyID != nil {
			s.Equal(s.testCompany.ID, *user.CompanyID)
		}
	}
}

// TestGetUsers_Regular tests that regular users cannot list users
func (s *UserManagementTestSuite) TestGetUsers_Regular() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer "+s.regularToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusForbidden, w.Code)
}

// TestGetUserByID_Success tests getting a specific user
func (s *UserManagementTestSuite) TestGetUserByID_Success() {
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s", s.adminUser.ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var user models.User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	s.NoError(err)
	s.Equal(s.adminUser.Email, user.Email)
}

// TestGetUserByID_NotFound tests getting a non-existent user
func (s *UserManagementTestSuite) TestGetUserByID_NotFound() {
	nonExistentID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s", nonExistentID), nil)
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

// TestCreateUser_Success tests successful user creation
func (s *UserManagementTestSuite) TestCreateUser_Success() {
	createReq := models.CreateUserRequest{
		Email:     "newuser@test.com",
		Password:  "Password123!",
		Name:      "New User",
		Phone:     "1234567890",
		CPF:       "12345678901",
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

	var user models.User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	s.NoError(err)
	s.Equal(createReq.Email, user.Email)
	s.Equal(createReq.Name, user.Name)
}

// TestCreateUser_DuplicateEmail tests creating user with existing email
func (s *UserManagementTestSuite) TestCreateUser_DuplicateEmail() {
	createReq := models.CreateUserRequest{
		Email:     s.adminUser.Email, // Existing email
		Password:  "Password123!",
		Name:      "Duplicate User",
		Phone:     "1234567890",
		CPF:       "12345678901",
		RoleID:    s.userRole.ID,
		CompanyID: &s.testCompany.ID,
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusConflict, w.Code)
}

// TestCreateUser_Forbidden tests that regular users cannot create users
func (s *UserManagementTestSuite) TestCreateUser_Forbidden() {
	createReq := models.CreateUserRequest{
		Email:     "forbidden@test.com",
		Password:  "Password123!",
		Name:      "Forbidden User",
		Phone:     "1234567890",
		CPF:       "12345678901",
		RoleID:    s.userRole.ID,
		CompanyID: &s.testCompany.ID,
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.regularToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusForbidden, w.Code)
}

// TestUpdateUser_Success tests successful user update
func (s *UserManagementTestSuite) TestUpdateUser_Success() {
	updateReq := models.UpdateUserRequest{
		Name:   stringPtr("Updated Name"),
		Active: boolPtr(true),
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/users/%s", s.regularUser.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.adminToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var user models.User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	s.NoError(err)
	s.Equal("Updated Name", user.Name)
}

// TestUpdateUser_NotFound tests updating non-existent user
func (s *UserManagementTestSuite) TestUpdateUser_NotFound() {
	updateReq := models.UpdateUserRequest{
		Name: stringPtr("Updated Name"),
	}

	nonExistentID := uuid.New()
	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/users/%s", nonExistentID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

// TestDeleteUser_Success tests successful user deletion
func (s *UserManagementTestSuite) TestDeleteUser_Success() {
	// Create a user to delete
	userToDelete := &models.User{
		ID:        uuid.New(),
		Email:     "todelete@test.com",
		Password:  "password",
		Name:      "To Delete",
		Active:    true,
		RoleID:    s.userRole.ID,
		CompanyID: &s.testCompany.ID,
	}
	s.Require().NoError(s.testDB.DB.Create(userToDelete).Error)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%s", userToDelete.ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Verify user is soft-deleted
	var deletedUser models.User
	err := s.testDB.DB.Unscoped().First(&deletedUser, userToDelete.ID).Error
	s.NoError(err)
	s.NotNil(deletedUser.DeletedAt)
}

// TestDeleteUser_NotFound tests deleting non-existent user
func (s *UserManagementTestSuite) TestDeleteUser_NotFound() {
	nonExistentID := uuid.New()
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%s", nonExistentID), nil)
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusNotFound, w.Code)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
