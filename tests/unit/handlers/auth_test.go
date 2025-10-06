package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"github.com/paulochiaradia/dashtrack/internal/auth"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
)

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}

// Mock repositories for testing
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) (*models.User, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) List(ctx context.Context, limit, offset int, active *bool, roleID *uuid.UUID) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset, active, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	args := m.Called(ctx, id, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepository) GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, companyID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateLoginAttempts(ctx context.Context, id uuid.UUID, attempts int, blockedUntil *time.Time) error {
	args := m.Called(ctx, id, attempts, blockedUntil)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) GetUserContext(ctx context.Context, userID uuid.UUID) (*models.UserContext, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserContext), args.Error(1)
}

// Dashboard methods
func (m *MockUserRepository) CountUsers(ctx context.Context, companyID *uuid.UUID) (int, error) {
	args := m.Called(ctx, companyID)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) CountActiveUsers(ctx context.Context, companyID *uuid.UUID) (int, error) {
	args := m.Called(ctx, companyID)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) Search(ctx context.Context, companyID *uuid.UUID, searchTerm string, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, companyID, searchTerm, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

// Multi-tenant methods
func (m *MockUserRepository) ListByCompanyAndRoles(ctx context.Context, companyID *uuid.UUID, roles []string, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, companyID, roles, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepository) CountByCompanyAndRoles(ctx context.Context, companyID *uuid.UUID, roles []string) (int, error) {
	args := m.Called(ctx, companyID, roles)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) ListByRoles(ctx context.Context, roles []string, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, roles, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

type MockAuthLogRepository struct {
	mock.Mock
}

func (m *MockAuthLogRepository) Create(authLog *models.AuthLog) error {
	args := m.Called(authLog)
	return args.Error(0)
}

func (m *MockAuthLogRepository) GetRecentFailedAttempts(email string, since time.Time) (int, error) {
	args := m.Called(email, since)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthLogRepository) GetByUserID(userID uuid.UUID, limit int) ([]*models.AuthLog, error) {
	args := m.Called(userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AuthLog), args.Error(1)
}

// Dashboard methods
func (m *MockAuthLogRepository) CountLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time) (int, error) {
	args := m.Called(ctx, companyID, from, to)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthLogRepository) CountSuccessfulLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time) (int, error) {
	args := m.Called(ctx, companyID, from, to)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthLogRepository) CountFailedLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time) (int, error) {
	args := m.Called(ctx, companyID, from, to)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthLogRepository) CountUserLogins(ctx context.Context, userID uuid.UUID, from, to time.Time) (int, error) {
	args := m.Called(ctx, userID, from, to)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthLogRepository) CountUserSuccessfulLogins(ctx context.Context, userID uuid.UUID, from, to time.Time) (int, error) {
	args := m.Called(ctx, userID, from, to)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthLogRepository) CountUserFailedLogins(ctx context.Context, userID uuid.UUID, from, to time.Time) (int, error) {
	args := m.Called(ctx, userID, from, to)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthLogRepository) GetRecentSuccessfulLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time, limit int) ([]models.RecentLogin, error) {
	args := m.Called(ctx, companyID, from, to, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RecentLogin), args.Error(1)
}

func (m *MockAuthLogRepository) GetUserRecentSuccessfulLogins(ctx context.Context, userID uuid.UUID, from, to time.Time, limit int) ([]models.RecentLogin, error) {
	args := m.Called(ctx, userID, from, to, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RecentLogin), args.Error(1)
}

// JWTManagerInterface defines the contract for JWT manager
type JWTManagerInterface interface {
	GenerateTokens(userContext auth.UserContext) (accessToken, refreshToken string, err error)
	ValidateToken(tokenString string) (*auth.JWTClaims, error)
	ValidateRefreshToken(tokenString string) (uuid.UUID, error)
}

type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) GenerateTokens(userContext auth.UserContext) (accessToken, refreshToken string, err error) {
	args := m.Called(userContext)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockJWTManager) ValidateToken(tokenString string) (*auth.JWTClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.JWTClaims), args.Error(1)
}

func (m *MockJWTManager) ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	args := m.Called(tokenString)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

// TestAuthHandler é uma versão testável do AuthHandler que usa interfaces
type TestAuthHandler struct {
	userRepo    repository.UserRepositoryInterface
	authLogRepo repository.AuthLogRepositoryInterface
	jwtManager  JWTManagerInterface
	bcryptCost  int
}

// LoginGin implementa o mesmo método do AuthHandler para testes
func (h *TestAuthHandler) LoginGin(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Get user by email
	user, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		// Log failed attempt
		authLog := &models.AuthLog{
			ID:            uuid.New(),
			UserID:        nil,
			EmailAttempt:  req.Email,
			Success:       false,
			IPAddress:     &clientIP,
			UserAgent:     &userAgent,
			FailureReason: stringPtr("invalid_email"),
		}
		h.authLogRepo.Create(authLog)

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// Log failed attempt
		authLog := &models.AuthLog{
			ID:            uuid.New(),
			UserID:        &user.ID,
			EmailAttempt:  req.Email,
			Success:       false,
			IPAddress:     &clientIP,
			UserAgent:     &userAgent,
			FailureReason: stringPtr("invalid_password"),
		}
		h.authLogRepo.Create(authLog)

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Generate tokens
	userContext := auth.UserContext{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		RoleID:   user.RoleID,
		RoleName: user.Role.Name,
	}

	accessToken, refreshToken, err := h.jwtManager.GenerateTokens(userContext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
		return
	}

	// Log successful attempt
	authLog := &models.AuthLog{
		ID:        uuid.New(),
		UserID:    &user.ID,
		Success:   true,
		IPAddress: &clientIP,
		UserAgent: &userAgent,
	}
	h.authLogRepo.Create(authLog)

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role.Name,
		},
	})
}

// Helper function to create test user
func createTestUser(role string, companyID *uuid.UUID) *models.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), 4)
	return &models.User{
		ID:        uuid.New(),
		Name:      "Test User",
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		RoleID:    uuid.New(),
		CompanyID: companyID,
		Role: &models.Role{
			ID:   uuid.New(),
			Name: role,
		},
		Active: true,
	}
}

// Test cases
func TestLoginGin_Success(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	mockUserRepo := new(MockUserRepository)
	mockAuthLogRepo := new(MockAuthLogRepository)
	mockJWTManager := new(MockJWTManager)

	handler := &TestAuthHandler{
		userRepo:    mockUserRepo,
		authLogRepo: mockAuthLogRepo,
		jwtManager:  mockJWTManager,
		bcryptCost:  4, // Low cost for testing
	}

	// Create test user
	testUser := createTestUser("admin", nil)

	// Setup mocks
	mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(testUser, nil)
	mockAuthLogRepo.On("Create", mock.Anything).Return(nil)

	expectedUserContext := auth.UserContext{
		UserID:   testUser.ID,
		Email:    testUser.Email,
		Name:     testUser.Name,
		RoleID:   testUser.RoleID,
		RoleName: testUser.Role.Name,
	}
	mockJWTManager.On("GenerateTokens", expectedUserContext).Return("access_token", "refresh_token", nil)

	// Setup request
	loginReq := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/auth/login", handler.LoginGin)

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "access_token", response["access_token"])

	mockUserRepo.AssertExpectations(t)
	mockAuthLogRepo.AssertExpectations(t)
	mockJWTManager.AssertExpectations(t)
}

func TestLoginGin_InvalidEmail(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	mockUserRepo := new(MockUserRepository)
	mockAuthLogRepo := new(MockAuthLogRepository)
	mockJWTManager := new(MockJWTManager)

	handler := &TestAuthHandler{
		userRepo:    mockUserRepo,
		authLogRepo: mockAuthLogRepo,
		jwtManager:  mockJWTManager,
		bcryptCost:  4,
	}

	// Setup mocks - user not found
	mockUserRepo.On("GetByEmail", mock.Anything, "invalid@example.com").Return(nil, assert.AnError)
	mockAuthLogRepo.On("Create", mock.Anything).Return(nil)

	// Setup request
	loginReq := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    "invalid@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/auth/login", handler.LoginGin)

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid credentials", response["error"])

	mockUserRepo.AssertExpectations(t)
	mockAuthLogRepo.AssertExpectations(t)
	// JWT manager should not be called for failed authentication
	mockJWTManager.AssertNotCalled(t, "GenerateTokens")
}

func TestLoginGin_InvalidPassword(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)

	mockUserRepo := new(MockUserRepository)
	mockAuthLogRepo := new(MockAuthLogRepository)
	mockJWTManager := new(MockJWTManager)

	handler := &TestAuthHandler{
		userRepo:    mockUserRepo,
		authLogRepo: mockAuthLogRepo,
		jwtManager:  mockJWTManager,
		bcryptCost:  4,
	}

	// Create test user
	testUser := createTestUser("admin", nil)

	// Setup mocks
	mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(testUser, nil)
	mockAuthLogRepo.On("Create", mock.Anything).Return(nil)

	// Setup request with wrong password
	loginReq := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{
		Email:    "test@example.com",
		Password: "wrongpassword",
	}
	jsonBody, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router := gin.New()
	router.POST("/auth/login", handler.LoginGin)

	// Execute
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid credentials", response["error"])

	mockUserRepo.AssertExpectations(t)
	mockAuthLogRepo.AssertExpectations(t)
	// JWT manager should not be called for failed authentication
	mockJWTManager.AssertNotCalled(t, "GenerateTokens")
}
