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

type UserWorkflowsTestSuite struct {
	suite.Suite
	testDB       *testutils.TestDB
	router       *gin.Engine
	tokenService *services.TokenService
	userService  *services.UserService
	userRepo     *repository.UserRepository
	roleRepo     *repository.RoleRepository
	companyRepo  *repository.CompanyRepository
	authLogRepo  *repository.AuthLogRepository
	testCompany  *models.Company
	masterRole   *models.Role
	adminRole    *models.Role
	userRole     *models.Role
}

func TestUserWorkflowsSuite(t *testing.T) {
	suite.Run(t, new(UserWorkflowsTestSuite))
}

func (s *UserWorkflowsTestSuite) SetupSuite() {
	var err error
	s.testDB, err = testutils.SetupTestDB("user_workflows_test")
	s.Require().NoError(err)

	// Initialize repositories
	s.userRepo = repository.NewUserRepository(s.testDB.DB)
	s.roleRepo = repository.NewRoleRepository(s.testDB.DB)
	s.companyRepo = repository.NewCompanyRepository(s.testDB.DB)
	s.authLogRepo = repository.NewAuthLogRepository(s.testDB.DB)

	// Initialize services
	s.tokenService = services.NewTokenService(
		s.testDB.SqlxDB,
		"test-secret-key",
		15*time.Minute,
		7*24*time.Hour,
	)
	s.userService = services.NewUserService(s.userRepo, s.roleRepo, s.companyRepo, 10)

	// Create test roles
	s.masterRole = &models.Role{ID: uuid.New(), Name: "master"}
	s.adminRole = &models.Role{ID: uuid.New(), Name: "admin"}
	s.userRole = &models.Role{ID: uuid.New(), Name: "user"}

	s.Require().NoError(s.testDB.DB.Create(s.masterRole).Error)
	s.Require().NoError(s.testDB.DB.Create(s.adminRole).Error)
	s.Require().NoError(s.testDB.DB.Create(s.userRole).Error)

	// Create test company
	s.testCompany = &models.Company{
		ID:               uuid.New(),
		Name:             "Test Company",
		Slug:             "test-company",
		Email:            "contact@testcompany.com",
		Phone:            "1234567890",
		SubscriptionPlan: "premium",
	}
	s.Require().NoError(s.testDB.DB.Create(s.testCompany).Error)

	// Setup router with all endpoints
	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	authHandler := handlers.NewAuthHandler(s.userRepo, s.authLogRepo, s.roleRepo, s.tokenService, nil, bcrypt.DefaultCost)
	userHandler := handlers.NewUserHandler(s.userService)
	authMiddleware := handlers.NewAuthMiddleware(s.tokenService, s.userRepo)

	api := s.router.Group("/api/v1")
	{
		// Public auth endpoints
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.LoginGin)
			auth.POST("/logout", authMiddleware.RequireAuth(), authHandler.LogoutGin)
			auth.POST("/refresh", authHandler.RefreshTokenGin)
		}

		// Protected user endpoints
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

func (s *UserWorkflowsTestSuite) TearDownSuite() {
	s.testDB.Close()
}

// TestCompleteUserWorkflow tests a complete user lifecycle
func (s *UserWorkflowsTestSuite) TestCompleteUserWorkflow() {
	ctx := context.Background()

	// Step 1: Create master user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Master123!"), bcrypt.DefaultCost)
	masterUser := &models.User{
		ID:       uuid.New(),
		Email:    "master@workflow.com",
		Password: string(hashedPassword),
		Name:     "Master User",
		Active:   true,
		RoleID:   s.masterRole.ID,
	}
	s.Require().NoError(s.testDB.DB.Create(masterUser).Error)

	// Step 2: Master logs in
	loginReq := handlers.LoginRequest{
		Email:    "master@workflow.com",
		Password: "Master123!",
	}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var loginResp handlers.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &loginResp)
	s.NoError(err)
	s.NotEmpty(loginResp.AccessToken)
	masterToken := loginResp.AccessToken

	// Step 3: Master creates a new user
	createReq := models.CreateUserRequest{
		Email:     "newuser@workflow.com",
		Password:  "User123!",
		Name:      "New User",
		Phone:     "9876543210",
		CPF:       "12345678901",
		RoleID:    s.userRole.ID,
		CompanyID: &s.testCompany.ID,
	}
	body, _ = json.Marshal(createReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+masterToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var createdUser models.User
	err = json.Unmarshal(w.Body.Bytes(), &createdUser)
	s.NoError(err)
	s.Equal("newuser@workflow.com", createdUser.Email)
	newUserID := createdUser.ID

	// Step 4: New user logs in
	loginReq = handlers.LoginRequest{
		Email:    "newuser@workflow.com",
		Password: "User123!",
	}
	body, _ = json.Marshal(loginReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &loginResp)
	s.NoError(err)
	s.NotEmpty(loginResp.AccessToken)
	userToken := loginResp.AccessToken

	// Step 5: New user views their own profile
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s", newUserID), nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var fetchedUser models.User
	err = json.Unmarshal(w.Body.Bytes(), &fetchedUser)
	s.NoError(err)
	s.Equal(newUserID, fetchedUser.ID)

	// Step 6: Master updates the user
	updateReq := models.UpdateUserRequest{
		Name: ptrString("Updated User Name"),
	}
	body, _ = json.Marshal(updateReq)
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/users/%s", newUserID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+masterToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &fetchedUser)
	s.NoError(err)
	s.Equal("Updated User Name", fetchedUser.Name)

	// Step 7: New user logs out
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Step 8: Verify token is invalidated (logout should fail with same token)
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s", newUserID), nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	// Token should be invalid after logout
	s.NotEqual(http.StatusOK, w.Code)

	// Step 9: Master deletes the user
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/users/%s", newUserID), nil)
	req.Header.Set("Authorization", "Bearer "+masterToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Step 10: Verify user is soft-deleted
	var deletedUser models.User
	err = s.testDB.DB.Unscoped().First(&deletedUser, newUserID).Error
	s.NoError(err)
	s.NotNil(deletedUser.DeletedAt)

	// Cleanup
	_ = ctx
}

// TestLoginRefreshLogoutWorkflow tests token refresh workflow
func (s *UserWorkflowsTestSuite) TestLoginRefreshLogoutWorkflow() {
	// Create test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Test123!"), bcrypt.DefaultCost)
	testUser := &models.User{
		ID:        uuid.New(),
		Email:     "refresh@workflow.com",
		Password:  string(hashedPassword),
		Name:      "Refresh Test User",
		Active:    true,
		RoleID:    s.userRole.ID,
		CompanyID: &s.testCompany.ID,
	}
	s.Require().NoError(s.testDB.DB.Create(testUser).Error)

	// Step 1: User logs in
	loginReq := handlers.LoginRequest{
		Email:    "refresh@workflow.com",
		Password: "Test123!",
	}
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var loginResp handlers.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &loginResp)
	s.NoError(err)
	s.NotEmpty(loginResp.AccessToken)
	s.NotEmpty(loginResp.RefreshToken)

	accessToken := loginResp.AccessToken
	refreshToken := loginResp.RefreshToken

	// Step 2: Use access token to access protected endpoint
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s", testUser.ID), nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Step 3: Refresh token
	refreshReq := handlers.RefreshTokenRequest{
		RefreshToken: refreshToken,
	}
	body, _ = json.Marshal(refreshReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var refreshResp handlers.LoginResponse
	err = json.Unmarshal(w.Body.Bytes(), &refreshResp)
	s.NoError(err)
	s.NotEmpty(refreshResp.AccessToken)
	s.NotEqual(accessToken, refreshResp.AccessToken) // New token should be different

	// Step 4: Use new access token
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%s", testUser.ID), nil)
	req.Header.Set("Authorization", "Bearer "+refreshResp.AccessToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Step 5: Logout
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+refreshResp.AccessToken)
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)
}

// TestUnauthorizedAccessWorkflow tests unauthorized access attempts
func (s *UserWorkflowsTestSuite) TestUnauthorizedAccessWorkflow() {
	// Step 1: Try to access protected endpoint without token
	req := httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	w := httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusUnauthorized, w.Code)

	// Step 2: Try with invalid token
	req = httptest.NewRequest(http.MethodGet, "/api/v1/users", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusUnauthorized, w.Code)

	// Step 3: Try login with wrong credentials
	loginReq := handlers.LoginRequest{
		Email:    "nonexistent@workflow.com",
		Password: "WrongPassword!",
	}
	body, _ := json.Marshal(loginReq)
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusUnauthorized, w.Code)
}

// Helper function
func ptrString(s string) *string {
	return &s
}
