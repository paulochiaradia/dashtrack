package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"

	"github.com/paulochiaradia/dashtrack/internal/handlers"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
	"github.com/paulochiaradia/dashtrack/internal/services"
	"github.com/paulochiaradia/dashtrack/tests/testutils"
)

// MockAuthLogRepository for testing
type MockAuthLogRepository struct {
	mock.Mock
}

func (m *MockAuthLogRepository) Create(log *models.AuthLog) error {
	args := m.Called(log)
	return args.Error(0)
}

func (m *MockAuthLogRepository) GetByUserID(userID uuid.UUID, limit int) ([]*models.AuthLog, error) {
	args := m.Called(userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.AuthLog), args.Error(1)
}

func (m *MockAuthLogRepository) CountFailedLogins(ctx context.Context, userID *uuid.UUID, startTime time.Time, endTime time.Time) (int, error) {
	args := m.Called(ctx, userID, startTime, endTime)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthLogRepository) CountLogins(ctx context.Context, userID *uuid.UUID, startTime time.Time, endTime time.Time) (int, error) {
	args := m.Called(ctx, userID, startTime, endTime)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthLogRepository) CountSuccessfulLogins(ctx context.Context, userID *uuid.UUID, startTime time.Time, endTime time.Time) (int, error) {
	args := m.Called(ctx, userID, startTime, endTime)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthLogRepository) CountUserFailedLogins(ctx context.Context, userID uuid.UUID, startTime time.Time, endTime time.Time) (int, error) {
	args := m.Called(ctx, userID, startTime, endTime)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthLogRepository) CountUserLogins(ctx context.Context, userID uuid.UUID, startTime time.Time, endTime time.Time) (int, error) {
	args := m.Called(ctx, userID, startTime, endTime)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthLogRepository) CountUserSuccessfulLogins(ctx context.Context, userID uuid.UUID, startTime time.Time, endTime time.Time) (int, error) {
	args := m.Called(ctx, userID, startTime, endTime)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthLogRepository) GetRecentFailedAttempts(email string, since time.Time) (int, error) {
	args := m.Called(email, since)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthLogRepository) GetRecentSuccessfulLogins(ctx context.Context, companyID *uuid.UUID, startTime time.Time, endTime time.Time, limit int) ([]models.RecentLogin, error) {
	args := m.Called(ctx, companyID, startTime, endTime, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RecentLogin), args.Error(1)
}

func (m *MockAuthLogRepository) GetUserRecentSuccessfulLogins(ctx context.Context, userID uuid.UUID, startTime time.Time, endTime time.Time, limit int) ([]models.RecentLogin, error) {
	args := m.Called(ctx, userID, startTime, endTime, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.RecentLogin), args.Error(1)
}

// MockRoleRepository for testing
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleRepository) GetByName(ctx context.Context, name string) (*models.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Role), args.Error(1)
}

func (m *MockRoleRepository) List(ctx context.Context) ([]models.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Role), args.Error(1)
}

func (m *MockRoleRepository) Create(ctx context.Context, role *models.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) Update(ctx context.Context, role *models.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleRepository) GetAll(ctx context.Context) ([]*models.Role, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Role), args.Error(1)
}

// MockUserRepositoryForAuth implements UserRepositoryInterface for auth tests
type MockUserRepositoryForAuth struct {
	mock.Mock
}

func (m *MockUserRepositoryForAuth) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryForAuth) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryForAuth) Create(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepositoryForAuth) Update(ctx context.Context, id uuid.UUID, updateReq models.UpdateUserRequest) (*models.User, error) {
	args := m.Called(ctx, id, updateReq)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepositoryForAuth) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	args := m.Called(ctx, id, hashedPassword)
	return args.Error(0)
}

func (m *MockUserRepositoryForAuth) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepositoryForAuth) List(ctx context.Context, limit, offset int, active *bool, roleID *uuid.UUID) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset, active, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepositoryForAuth) GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, companyID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepositoryForAuth) UpdateCompany(ctx context.Context, userID, companyID uuid.UUID) error {
	args := m.Called(ctx, userID, companyID)
	return args.Error(0)
}

func (m *MockUserRepositoryForAuth) UpdateLoginAttempts(ctx context.Context, id uuid.UUID, attempts int, blockedUntil *time.Time) error {
	args := m.Called(ctx, id, attempts, blockedUntil)
	return args.Error(0)
}

func (m *MockUserRepositoryForAuth) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepositoryForAuth) GetUserContext(ctx context.Context, userID uuid.UUID) (*models.UserContext, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserContext), args.Error(1)
}

func (m *MockUserRepositoryForAuth) Search(ctx context.Context, companyID *uuid.UUID, searchTerm string, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, companyID, searchTerm, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepositoryForAuth) CountUsers(ctx context.Context, companyID *uuid.UUID) (int, error) {
	args := m.Called(ctx, companyID)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepositoryForAuth) CountActiveUsers(ctx context.Context, companyID *uuid.UUID) (int, error) {
	args := m.Called(ctx, companyID)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepositoryForAuth) ListByCompanyAndRoles(ctx context.Context, companyID *uuid.UUID, roles []string, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, companyID, roles, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepositoryForAuth) ListByRoles(ctx context.Context, roles []string, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, roles, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserRepositoryForAuth) CountByCompanyAndRoles(ctx context.Context, companyID *uuid.UUID, roles []string) (int, error) {
	args := m.Called(ctx, companyID, roles)
	return args.Int(0), args.Error(1)
}

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testDB, err := testutils.SetupTestDB("auth_handler_test")
	require.NoError(t, err)
	defer testDB.TearDown()

	tokenService := services.NewTokenService(
		testDB.SqlxDB,
		"test-secret-key",
		15*time.Minute,
		7*24*time.Hour,
	)

	role := &models.Role{
		ID:   uuid.New(),
		Name: "company_admin",
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("ValidPassword123"), bcrypt.DefaultCost)
	user := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		Name:      "Test User",
		Password:  string(hashedPassword),
		Active:    true,
		RoleID:    role.ID,
		Role:      role,
		CompanyID: nil,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = testDB.DB.Create(role).Error
	require.NoError(t, err)
	err = testDB.DB.Create(user).Error
	require.NoError(t, err)

	t.Run("Successful Login", func(t *testing.T) {
		mockUserRepo := new(MockUserRepositoryForAuth)
		mockAuthLogRepo := new(MockAuthLogRepository)
		mockRoleRepo := new(MockRoleRepository)

		authHandler := handlers.NewAuthHandler(
			mockUserRepo,
			mockAuthLogRepo,
			mockRoleRepo,
			tokenService,
			nil,
			bcrypt.DefaultCost,
		)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		loginReq := map[string]string{
			"email":    "test@example.com",
			"password": "ValidPassword123",
		}
		body, _ := json.Marshal(loginReq)
		c.Request = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
		mockAuthLogRepo.On("Create", mock.AnythingOfType("*models.AuthLog")).Return(nil)

		authHandler.LoginGin(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")
		assert.Contains(t, response, "user")

		mockUserRepo.AssertExpectations(t)
		mockAuthLogRepo.AssertExpectations(t)
	})

	t.Run("Invalid Email", func(t *testing.T) {
		mockUserRepo := new(MockUserRepositoryForAuth)
		mockAuthLogRepo := new(MockAuthLogRepository)
		mockRoleRepo := new(MockRoleRepository)

		authHandler := handlers.NewAuthHandler(
			mockUserRepo,
			mockAuthLogRepo,
			mockRoleRepo,
			tokenService,
			nil,
			bcrypt.DefaultCost,
		)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		loginReq := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "ValidPassword123",
		}
		body, _ := json.Marshal(loginReq)
		c.Request = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		mockUserRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return((*models.User)(nil), errors.New("user not found"))
		mockUserRepo.On("GetByEmail", mock.Anything, "nonexistent@example.com").Return((*models.User)(nil), repository.ErrUserNotFound)
		mockAuthLogRepo.On("Create", mock.AnythingOfType("*models.AuthLog")).Return(nil)

		authHandler.LoginGin(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Invalid Password", func(t *testing.T) {
		mockUserRepo := new(MockUserRepositoryForAuth)
		mockAuthLogRepo := new(MockAuthLogRepository)
		mockRoleRepo := new(MockRoleRepository)

		authHandler := handlers.NewAuthHandler(
			mockUserRepo,
			mockAuthLogRepo,
			mockRoleRepo,
			tokenService,
			nil,
			bcrypt.DefaultCost,
		)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		loginReq := map[string]string{
			"email":    "test@example.com",
			"password": "WrongPassword",
		}
		body, _ := json.Marshal(loginReq)
		c.Request = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		mockUserRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(user, nil)
		mockAuthLogRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.AuthLog")).Return(nil)

		authHandler.LoginGin(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Inactive User", func(t *testing.T) {
		mockUserRepo := new(MockUserRepositoryForAuth)
		mockAuthLogRepo := new(MockAuthLogRepository)
		mockRoleRepo := new(MockRoleRepository)

		authHandler := handlers.NewAuthHandler(
			mockUserRepo,
			mockAuthLogRepo,
			mockRoleRepo,
			tokenService,
			nil,
			bcrypt.DefaultCost,
		)

		inactiveUser := &models.User{
			ID:        uuid.New(),
			Email:     "inactive@example.com",
			Name:      "Inactive User",
			Password:  string(hashedPassword),
			Active:    false,
			RoleID:    role.ID,
			Role:      role,
			CompanyID: nil,
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		loginReq := map[string]string{
			"email":    "inactive@example.com",
			"password": "ValidPassword123",
		}
		body, _ := json.Marshal(loginReq)
		c.Request = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		mockUserRepo.On("GetByEmail", mock.Anything, "inactive@example.com").Return(inactiveUser, nil)
		mockAuthLogRepo.On("Create", mock.AnythingOfType("*models.AuthLog")).Return(nil)

		authHandler.LoginGin(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	testDB, err := testutils.SetupTestDB("auth_handler_refresh_test")
	require.NoError(t, err)
	defer testDB.TearDown()

	tokenService := services.NewTokenService(
		testDB.SqlxDB,
		"test-secret-key",
		15*time.Minute,
		7*24*time.Hour,
	)

	userID := uuid.New()
	user := &models.User{
		ID:    userID,
		Email: "test@example.com",
		Name:  "Test User",
		Role: &models.Role{
			Name: "user",
		},
		CompanyID: &uuid.UUID{},
	}
	*user.CompanyID = uuid.New()

	userContext := &models.UserContext{
		UserID:      userID,
		Email:       "test@example.com",
		Name:        "Test User",
		Role:        "user",
		Permissions: []string{"read:own_data"},
		CompanyID:   user.CompanyID,
	}

	refreshToken, err := tokenService.GenerateTokenPair(context.Background(), user, "", "")
	require.NoError(t, err)
	require.NotEmpty(t, refreshToken)

	t.Run("Successful Refresh", func(t *testing.T) {
		mockUserRepo := new(MockUserRepositoryForAuth)
		mockAuthLogRepo := new(MockAuthLogRepository)
		mockRoleRepo := new(MockRoleRepository)

		authHandler := handlers.NewAuthHandler(
			mockUserRepo,
			mockAuthLogRepo,
			mockRoleRepo,
			tokenService,
			nil,
			bcrypt.DefaultCost,
		)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]string{"refresh_token": refreshToken}
		body, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		mockUserRepo.On("GetUserContext", mock.Anything, userID).Return(userContext, nil).Once()

		authHandler.RefreshTokenGin(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "access_token")
		assert.Contains(t, response, "refresh_token")
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Invalid Refresh Token", func(t *testing.T) {
		mockUserRepo := new(MockUserRepositoryForAuth)
		mockAuthLogRepo := new(MockAuthLogRepository)
		mockRoleRepo := new(MockRoleRepository)

		authHandler := handlers.NewAuthHandler(
			mockUserRepo,
			mockAuthLogRepo,
			mockRoleRepo,
			tokenService,
			nil,
			bcrypt.DefaultCost,
		)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]string{"refresh_token": "invalid-token"}
		body, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		authHandler.RefreshTokenGin(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Missing Refresh Token", func(t *testing.T) {
		mockUserRepo := new(MockUserRepositoryForAuth)
		mockAuthLogRepo := new(MockAuthLogRepository)
		mockRoleRepo := new(MockRoleRepository)

		authHandler := handlers.NewAuthHandler(
			mockUserRepo,
			mockAuthLogRepo,
			mockRoleRepo,
			tokenService,
			nil,
			bcrypt.DefaultCost,
		)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]string{}
		body, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		authHandler.RefreshTokenGin(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("User Not Found During Refresh", func(t *testing.T) {
		mockUserRepo := new(MockUserRepositoryForAuth)
		mockAuthLogRepo := new(MockAuthLogRepository)
		mockRoleRepo := new(MockRoleRepository)

		authHandler := handlers.NewAuthHandler(
			mockUserRepo,
			mockAuthLogRepo,
			mockRoleRepo,
			tokenService,
			nil,
			bcrypt.DefaultCost,
		)

		otherUserID := uuid.New()
		otherUser := &models.User{
			ID: otherUserID,
			Role: &models.Role{
				Name: "user",
			},
		}
		otherRefreshToken, err := tokenService.GenerateTokenPair(context.Background(), otherUser, "", "")
		require.NoError(t, err)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		reqBody := map[string]string{"refresh_token": otherRefreshToken}
		body, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("POST", "/api/v1/auth/refresh", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")
		mockUserRepo.On("GetUserContext", mock.Anything, otherUserID).Return(nil, errors.New("user not found")).Once()
		mockUserRepo.On("GetUserContext", mock.Anything, otherUserID).Return(nil, repository.ErrUserNotFound).Once()

		authHandler.RefreshTokenGin(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockUserRepo.AssertExpectations(t)
	})
}
