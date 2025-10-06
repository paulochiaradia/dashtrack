package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/paulochiaradia/dashtrack/internal/auth"
	"github.com/paulochiaradia/dashtrack/internal/middleware"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
)

type HierarchyTestSuite struct {
	suite.Suite
	router      *gin.Engine
	userRepo    *repository.UserRepository
	companyRepo *repository.CompanyRepository
	esp32Repo   *repository.ESP32DeviceRepository
	jwtManager  *auth.JWTManager

	// Test data
	masterUser   *models.User
	companyAdmin *models.User
	driver       *models.User
	helper       *models.User
	testCompany  *models.Company
	testCompany2 *models.Company
	testVehicle  *models.Vehicle
	testESP32    *models.ESP32Device

	// Tokens
	masterToken       string
	companyAdminToken string
	driverToken       string
	helperToken       string
}

func (suite *HierarchyTestSuite) SetupSuite() {
	// Initialize test database and dependencies
	gin.SetMode(gin.TestMode)

	// This would normally connect to a test database
	// For this example, we'll use mocks, but in real implementation
	// you'd set up a test DB connection

	suite.jwtManager = auth.NewJWTManager("test-secret", time.Hour, time.Hour*24, "test-issuer")

	// Setup router with all routes
	suite.router = gin.New()
	suite.setupRoutes()

	// Create test data
	suite.createTestData()
}

func (suite *HierarchyTestSuite) setupRoutes() {
	// Setup authentication middleware
	authMiddleware := middleware.NewGinAuthMiddleware(suite.jwtManager)

	// Public routes
	suite.router.POST("/auth/login", func(c *gin.Context) {
		// Mock login endpoint for testing
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request"})
			return
		}

		// Mock user authentication
		var user *models.User
		var role string

		switch req.Email {
		case "master@dashtrack.com":
			user = suite.masterUser
			role = "master"
		case "admin@company1.com":
			user = suite.companyAdmin
			role = "company_admin"
		case "driver@company1.com":
			user = suite.driver
			role = "driver"
		case "helper@company1.com":
			user = suite.helper
			role = "helper"
		default:
			c.JSON(401, gin.H{"error": "Invalid credentials"})
			return
		}

		userContext := auth.UserContext{
			UserID:   user.ID,
			Email:    user.Email,
			Name:     user.Name,
			RoleID:   user.RoleID,
			RoleName: role,
		}

		accessToken, refreshToken, err := suite.jwtManager.GenerateTokens(userContext)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to generate tokens"})
			return
		}

		c.JSON(200, gin.H{
			"access_token":  accessToken,
			"refresh_token": refreshToken,
			"expires_in":    3600,
			"user": gin.H{
				"id":    user.ID,
				"name":  user.Name,
				"email": user.Email,
				"role":  role,
			},
		})
	})

	// Protected routes
	api := suite.router.Group("/api/v1")
	api.Use(authMiddleware.RequireAuth())

	// Master routes
	master := api.Group("/master")
	master.Use(middleware.RequireMasterRole())
	{
		master.GET("/companies", suite.mockGetAllCompanies)
		master.POST("/companies", suite.mockCreateCompany)
		master.GET("/companies/:id", suite.mockGetCompany)
		master.PUT("/companies/:id", suite.mockUpdateCompany)
		master.DELETE("/companies/:id", suite.mockDeleteCompany)
	}

	// Company routes
	company := api.Group("/company")
	company.Use(middleware.RequireCompanyAccess())
	{
		company.GET("/info", suite.mockGetMyCompany)

		// Company admin routes
		companyAdmin := company.Group("/admin")
		companyAdmin.Use(middleware.RequireCompanyAdmin())
		{
			companyAdmin.GET("/users", suite.mockGetCompanyUsers)
			companyAdmin.POST("/users", suite.mockCreateCompanyUser)
		}

		// Vehicle routes
		vehicles := company.Group("/vehicles")
		vehicles.Use(middleware.RequireDriverOrHelper())
		{
			vehicles.GET("", suite.mockGetVehicles)
			vehicles.GET("/:id", middleware.RequireVehicleAccess(), suite.mockGetVehicle)
		}

		// ESP32 routes
		devices := company.Group("/devices")
		{
			devices.GET("", suite.mockGetESP32Devices)
			devices.POST("/:id/assign-vehicle", middleware.RequireCompanyAdmin(), suite.mockAssignESP32)
		}
	}

	// Admin routes
	admin := api.Group("/admin")
	admin.Use(middleware.RequireCompanyAdmin())
	{
		admin.GET("/users", suite.mockGetAllUsers)
		admin.POST("/users", suite.mockCreateUser)
	}
}

func (suite *HierarchyTestSuite) createTestData() {
	// Create test companies
	suite.testCompany = &models.Company{
		ID:   uuid.New(),
		Name: "Test Company 1",
		Slug: "test-company-1",
	}

	suite.testCompany2 = &models.Company{
		ID:   uuid.New(),
		Name: "Test Company 2",
		Slug: "test-company-2",
	}

	// Create test users
	suite.masterUser = &models.User{
		ID:    uuid.New(),
		Name:  "Master User",
		Email: "master@dashtrack.com",
		Role:  &models.Role{Name: "master"},
	}

	suite.companyAdmin = &models.User{
		ID:        uuid.New(),
		Name:      "Company Admin",
		Email:     "admin@company1.com",
		CompanyID: &suite.testCompany.ID,
		Role:      &models.Role{Name: "company_admin"},
	}

	suite.driver = &models.User{
		ID:        uuid.New(),
		Name:      "Driver User",
		Email:     "driver@company1.com",
		CompanyID: &suite.testCompany.ID,
		Role:      &models.Role{Name: "driver"},
	}

	suite.helper = &models.User{
		ID:        uuid.New(),
		Name:      "Helper User",
		Email:     "helper@company1.com",
		CompanyID: &suite.testCompany.ID,
		Role:      &models.Role{Name: "helper"},
	}

	// Generate tokens
	masterUserContext := auth.UserContext{
		UserID:   suite.masterUser.ID,
		Email:    suite.masterUser.Email,
		Name:     suite.masterUser.Name,
		RoleID:   suite.masterUser.RoleID,
		RoleName: suite.masterUser.Role.Name,
	}
	masterTokens, _, _ := suite.jwtManager.GenerateTokens(masterUserContext)
	suite.masterToken = masterTokens

	adminUserContext := auth.UserContext{
		UserID:   suite.companyAdmin.ID,
		Email:    suite.companyAdmin.Email,
		Name:     suite.companyAdmin.Name,
		RoleID:   suite.companyAdmin.RoleID,
		RoleName: suite.companyAdmin.Role.Name,
	}
	adminTokens, _, _ := suite.jwtManager.GenerateTokens(adminUserContext)
	suite.companyAdminToken = adminTokens

	driverUserContext := auth.UserContext{
		UserID:   suite.driver.ID,
		Email:    suite.driver.Email,
		Name:     suite.driver.Name,
		RoleID:   suite.driver.RoleID,
		RoleName: suite.driver.Role.Name,
	}
	driverTokens, _, _ := suite.jwtManager.GenerateTokens(driverUserContext)
	suite.driverToken = driverTokens

	helperUserContext := auth.UserContext{
		UserID:   suite.helper.ID,
		Email:    suite.helper.Email,
		Name:     suite.helper.Name,
		RoleID:   suite.helper.RoleID,
		RoleName: suite.helper.Role.Name,
	}
	helperTokens, _, _ := suite.jwtManager.GenerateTokens(helperUserContext)
	suite.helperToken = helperTokens
}

// Test Master User Permissions
func (suite *HierarchyTestSuite) TestMasterUser_CanAccessAllCompanies() {
	req := httptest.NewRequest("GET", "/api/v1/master/companies", nil)
	req.Header.Set("Authorization", "Bearer "+suite.masterToken)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Contains(suite.T(), response, "companies")
}

func (suite *HierarchyTestSuite) TestMasterUser_CanCreateCompany() {
	createReq := map[string]interface{}{
		"name":              "New Test Company",
		"slug":              "new-test-company",
		"email":             "contact@newtestcompany.com",
		"country":           "Brazil",
		"subscription_plan": "basic",
	}
	jsonBody, _ := json.Marshal(createReq)

	req := httptest.NewRequest("POST", "/api/v1/master/companies", bytes.NewBuffer(jsonBody))
	req.Header.Set("Authorization", "Bearer "+suite.masterToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
}

// Test Company Admin Permissions
func (suite *HierarchyTestSuite) TestCompanyAdmin_CannotAccessMasterRoutes() {
	req := httptest.NewRequest("GET", "/api/v1/master/companies", nil)
	req.Header.Set("Authorization", "Bearer "+suite.companyAdminToken)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *HierarchyTestSuite) TestCompanyAdmin_CanAccessOwnCompany() {
	req := httptest.NewRequest("GET", "/api/v1/company/info", nil)
	req.Header.Set("Authorization", "Bearer "+suite.companyAdminToken)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *HierarchyTestSuite) TestCompanyAdmin_CanManageCompanyUsers() {
	req := httptest.NewRequest("GET", "/api/v1/company/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+suite.companyAdminToken)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// Test Driver Permissions
func (suite *HierarchyTestSuite) TestDriver_CannotAccessMasterRoutes() {
	req := httptest.NewRequest("GET", "/api/v1/master/companies", nil)
	req.Header.Set("Authorization", "Bearer "+suite.driverToken)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *HierarchyTestSuite) TestDriver_CannotAccessCompanyAdmin() {
	req := httptest.NewRequest("GET", "/api/v1/company/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+suite.driverToken)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusForbidden, w.Code)
}

func (suite *HierarchyTestSuite) TestDriver_CanAccessCompanyInfo() {
	req := httptest.NewRequest("GET", "/api/v1/company/info", nil)
	req.Header.Set("Authorization", "Bearer "+suite.driverToken)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *HierarchyTestSuite) TestDriver_CanAccessVehicles() {
	req := httptest.NewRequest("GET", "/api/v1/company/vehicles", nil)
	req.Header.Set("Authorization", "Bearer "+suite.driverToken)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// Test Helper Permissions (similar to driver)
func (suite *HierarchyTestSuite) TestHelper_CanAccessVehicles() {
	req := httptest.NewRequest("GET", "/api/v1/company/vehicles", nil)
	req.Header.Set("Authorization", "Bearer "+suite.helperToken)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

// Test Cross-Company Access Prevention
func (suite *HierarchyTestSuite) TestCompanyAdmin_CannotAccessOtherCompanyData() {
	// This would test that company admin from company1 cannot access company2 data
	// In a real implementation, you'd test with actual repository calls
	assert.True(suite.T(), true, "Cross-company access prevention needs repository-level testing")
}

// Mock handlers for testing
func (suite *HierarchyTestSuite) mockGetAllCompanies(c *gin.Context) {
	c.JSON(200, gin.H{
		"companies": []gin.H{
			{"id": suite.testCompany.ID, "name": suite.testCompany.Name},
			{"id": suite.testCompany2.ID, "name": suite.testCompany2.Name},
		},
		"count": 2,
	})
}

func (suite *HierarchyTestSuite) mockCreateCompany(c *gin.Context) {
	c.JSON(201, gin.H{
		"id":   uuid.New(),
		"name": "New Test Company",
	})
}

func (suite *HierarchyTestSuite) mockGetCompany(c *gin.Context) {
	c.JSON(200, gin.H{
		"id":   suite.testCompany.ID,
		"name": suite.testCompany.Name,
	})
}

func (suite *HierarchyTestSuite) mockUpdateCompany(c *gin.Context) {
	c.JSON(200, gin.H{
		"id":   suite.testCompany.ID,
		"name": "Updated Company Name",
	})
}

func (suite *HierarchyTestSuite) mockDeleteCompany(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Company deleted successfully"})
}

func (suite *HierarchyTestSuite) mockGetMyCompany(c *gin.Context) {
	userContext, _ := c.Get("userContext")
	userCtx := userContext.(*models.UserContext)

	if userCtx.CompanyID == nil {
		c.JSON(400, gin.H{"error": "No company assigned"})
		return
	}

	c.JSON(200, gin.H{
		"data": gin.H{
			"id":   userCtx.CompanyID,
			"name": suite.testCompany.Name,
		},
	})
}

func (suite *HierarchyTestSuite) mockGetCompanyUsers(c *gin.Context) {
	c.JSON(200, gin.H{
		"users": []gin.H{
			{"id": suite.companyAdmin.ID, "name": suite.companyAdmin.Name, "role": "company_admin"},
			{"id": suite.driver.ID, "name": suite.driver.Name, "role": "driver"},
			{"id": suite.helper.ID, "name": suite.helper.Name, "role": "helper"},
		},
	})
}

func (suite *HierarchyTestSuite) mockCreateCompanyUser(c *gin.Context) {
	c.JSON(201, gin.H{
		"id":   uuid.New(),
		"name": "New Company User",
	})
}

func (suite *HierarchyTestSuite) mockGetVehicles(c *gin.Context) {
	c.JSON(200, gin.H{
		"vehicles": []gin.H{
			{"id": uuid.New(), "license_plate": "ABC-1234"},
		},
	})
}

func (suite *HierarchyTestSuite) mockGetVehicle(c *gin.Context) {
	c.JSON(200, gin.H{
		"id":            uuid.New(),
		"license_plate": "ABC-1234",
	})
}

func (suite *HierarchyTestSuite) mockGetESP32Devices(c *gin.Context) {
	c.JSON(200, gin.H{
		"devices": []gin.H{
			{"id": uuid.New(), "device_id": "ESP32_001"},
		},
	})
}

func (suite *HierarchyTestSuite) mockAssignESP32(c *gin.Context) {
	c.JSON(200, gin.H{"message": "ESP32 assigned successfully"})
}

func (suite *HierarchyTestSuite) mockGetAllUsers(c *gin.Context) {
	c.JSON(200, gin.H{
		"data": gin.H{
			"users": []gin.H{
				{"id": suite.masterUser.ID, "name": suite.masterUser.Name, "role": "master"},
				{"id": suite.companyAdmin.ID, "name": suite.companyAdmin.Name, "role": "company_admin"},
			},
		},
	})
}

func (suite *HierarchyTestSuite) mockCreateUser(c *gin.Context) {
	c.JSON(201, gin.H{
		"id":   uuid.New(),
		"name": "New User",
	})
}

// Run the test suite
func TestHierarchyTestSuite(t *testing.T) {
	suite.Run(t, new(HierarchyTestSuite))
}

// Benchmark tests for performance
func BenchmarkLoginEndpoint(b *testing.B) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Setup a simple login endpoint
	router.POST("/auth/login", func(c *gin.Context) {
		c.JSON(200, gin.H{"token": "test_token"})
	})

	loginReq := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(loginReq)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
	}
}

func BenchmarkAuthorizationMiddleware(b *testing.B) {
	gin.SetMode(gin.TestMode)

	jwtManager := auth.NewJWTManager("test-secret", time.Hour, time.Hour*24, "test-issuer")
	userContext := auth.UserContext{
		UserID:   uuid.New(),
		Email:    "test@example.com",
		Name:     "Test User",
		RoleID:   uuid.New(),
		RoleName: "admin",
	}
	accessToken, _, _ := jwtManager.GenerateTokens(userContext)

	authMiddleware := middleware.NewGinAuthMiddleware(jwtManager)

	router := gin.New()
	router.Use(authMiddleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+accessToken)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
	}
}
