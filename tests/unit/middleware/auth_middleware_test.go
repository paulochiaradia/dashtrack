package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paulochiaradia/dashtrack/internal/middleware"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/services"
	"github.com/paulochiaradia/dashtrack/tests/testutils"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	testDB, err := testutils.SetupTestDB("auth_middleware_test")
	require.NoError(t, err)
	defer testDB.TearDown()

	// Create token service with test configuration
	tokenService := services.NewTokenService(
		testDB.SqlxDB,
		"test-secret-key-for-jwt-tokens",
		15*time.Minute, // access token TTL
		7*24*time.Hour, // refresh token TTL
	)

	// Create middleware
	authMiddleware := middleware.NewGinAuthMiddleware(tokenService)

	// Create test user
	user := &models.User{
		ID:       uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "hashedpassword",
		Active:   true,
		RoleID:   uuid.New(),
	}

	// Create test role
	role := &models.Role{
		ID:   user.RoleID,
		Name: "company_admin",
	}

	// Insert role first
	err = testDB.DB.Create(role).Error
	require.NoError(t, err)

	// Insert user
	err = testDB.DB.Create(user).Error
	require.NoError(t, err)

	// Generate valid token pair
	tokenPair, err := tokenService.GenerateTokenPair(context.Background(), user, "127.0.0.1", "test-agent")
	require.NoError(t, err)
	token := tokenPair.AccessToken

	t.Run("RequireAuth - Valid Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer "+token)

		handler := authMiddleware.RequireAuth()
		handler(c)

		assert.False(t, c.IsAborted())

		// Check context values
		userID, exists := c.Get("user_id")
		assert.True(t, exists)
		assert.Equal(t, user.ID.String(), userID)

		email, exists := c.Get("email")
		assert.True(t, exists)
		assert.Equal(t, user.Email, email)

		roleName, exists := c.Get("role_name")
		assert.True(t, exists)
		assert.Equal(t, "company_admin", roleName)
	})

	t.Run("RequireAuth - Missing Authorization Header", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		handler := authMiddleware.RequireAuth()
		handler(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("RequireAuth - Invalid Authorization Format", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "InvalidFormat "+token)

		handler := authMiddleware.RequireAuth()
		handler(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("RequireAuth - Invalid Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Request.Header.Set("Authorization", "Bearer invalidtoken123")

		handler := authMiddleware.RequireAuth()
		handler(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("RequireAuth - Expired Token", func(t *testing.T) {
		// This test would require generating an expired token
		// Skip for now as it requires token manipulation
		t.Skip("Requires token expiration testing")
	})

	t.Run("RequireRole - User Has Required Role", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Set("role_name", "company_admin")

		handler := authMiddleware.RequireRole("company_admin")
		handler(c)

		assert.False(t, c.IsAborted())
	})

	t.Run("RequireRole - User Does Not Have Required Role", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Set("role_name", "driver")

		handler := authMiddleware.RequireRole("company_admin")
		handler(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("RequireRole - Master Has Universal Access", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Set("role_name", "master")

		handler := authMiddleware.RequireRole("company_admin")
		handler(c)

		assert.False(t, c.IsAborted())
	})

	t.Run("RequireRole - Missing Role Context", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)

		handler := authMiddleware.RequireRole("company_admin")
		handler(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestRequireCompanyAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Company Admin Has Access", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Set("role_name", "company_admin")

		middleware.RequireCompanyAdmin()(c)

		assert.False(t, c.IsAborted())
	})

	t.Run("Master Has Access", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Set("role_name", "master")

		middleware.RequireCompanyAdmin()(c)

		assert.False(t, c.IsAborted())
	})

	t.Run("Regular User Does Not Have Access", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/test", nil)
		c.Set("role_name", "driver")

		middleware.RequireCompanyAdmin()(c)

		assert.True(t, c.IsAborted())
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
