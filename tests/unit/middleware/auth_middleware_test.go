package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/paulochiaradia/dashtrack/internal/auth"
	"github.com/paulochiaradia/dashtrack/internal/middleware"
	"github.com/paulochiaradia/dashtrack/tests/testutils/mocks"
)

// AuthMiddlewareTestSuite defines the test suite for auth middleware
type AuthMiddlewareTestSuite struct {
	suite.Suite
	ctrl           *gomock.Controller
	jwtManager     *mocks.MockJWTManager
	authMiddleware *middleware.GinAuthMiddleware
	router         *gin.Engine
}

func (suite *AuthMiddlewareTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.ctrl = gomock.NewController(suite.T())
	suite.jwtManager = mocks.NewMockJWTManager(suite.ctrl)
	suite.authMiddleware = middleware.NewGinAuthMiddleware(suite.jwtManager)
	suite.router = gin.New()
}

func (suite *AuthMiddlewareTestSuite) TearDownTest() {
	suite.ctrl.Finish()
}

func (suite *AuthMiddlewareTestSuite) TestRequireAuth_ValidToken() {
	// Setup
	userID := uuid.New()
	roleID := uuid.New()
	companyID := uuid.New()

	claims := &auth.JWTClaims{
		UserID:   userID,
		Email:    "test@example.com",
		Name:     "Test User",
		RoleID:   roleID,
		RoleName: "admin",
		TenantID: &companyID,
	}

	suite.jwtManager.EXPECT().
		ValidateToken("valid-token").
		Return(claims, nil)

	// Setup route
	suite.router.Use(suite.authMiddleware.RequireAuth())
	suite.router.GET("/protected", func(c *gin.Context) {
		// Verify context is set correctly
		userIDFromContext, exists := c.Get("user_id")
		assert.True(suite.T(), exists)
		assert.Equal(suite.T(), userID.String(), userIDFromContext)

		roleIDFromContext, exists := c.Get("role_id")
		assert.True(suite.T(), exists)
		assert.Equal(suite.T(), roleID.String(), roleIDFromContext)

		companyIDFromContext, exists := c.Get("company_id")
		assert.True(suite.T(), exists)
		assert.Equal(suite.T(), companyID.String(), companyIDFromContext)

		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *AuthMiddlewareTestSuite) TestRequireAuth_MissingToken() {
	// Setup route
	suite.router.Use(suite.authMiddleware.RequireAuth())
	suite.router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *AuthMiddlewareTestSuite) TestRequireAuth_InvalidTokenFormat() {
	// Setup route
	suite.router.Use(suite.authMiddleware.RequireAuth())
	suite.router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *AuthMiddlewareTestSuite) TestRequireAuth_InvalidToken() {
	// Setup
	suite.jwtManager.EXPECT().
		ValidateToken("invalid-token").
		Return(nil, assert.AnError)

	// Setup route
	suite.router.Use(suite.authMiddleware.RequireAuth())
	suite.router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Test
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *AuthMiddlewareTestSuite) TestRequireRole_HasRequiredRole() {
	// Setup
	userID := uuid.New()
	roleID := uuid.New()
	companyID := uuid.New()

	claims := &auth.JWTClaims{
		UserID:   userID,
		Email:    "test@example.com",
		Name:     "Test User",
		RoleID:   roleID,
		RoleName: "admin",
		TenantID: &companyID,
	}

	suite.jwtManager.EXPECT().
		ValidateToken("valid-token").
		Return(claims, nil)

	// Setup route
	suite.router.Use(suite.authMiddleware.RequireAuth())
	suite.router.Use(suite.authMiddleware.RequireRole("admin"))
	suite.router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	// Test
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *AuthMiddlewareTestSuite) TestRequireRole_MissingRequiredRole() {
	// Setup
	userID := uuid.New()
	roleID := uuid.New()
	companyID := uuid.New()

	claims := &auth.JWTClaims{
		UserID:   userID,
		Email:    "test@example.com",
		Name:     "Test User",
		RoleID:   roleID,
		RoleName: "user", // User has 'user' role but needs 'admin'
		TenantID: &companyID,
	}

	suite.jwtManager.EXPECT().
		ValidateToken("valid-token").
		Return(claims, nil)

	// Setup route
	suite.router.Use(suite.authMiddleware.RequireAuth())
	suite.router.Use(suite.authMiddleware.RequireRole("admin"))
	suite.router.GET("/admin", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	// Test
	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func TestAuthMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareTestSuite))
}
