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

type HierarchyTestSuite struct {
	suite.Suite
	testDB       *testutils.TestDB
	router       *gin.Engine
	tokenService *services.TokenService
	companyRepo  *repository.CompanyRepository
	userRepo     *repository.UserRepository
	roleRepo     *repository.RoleRepository
	masterUser   *models.User
	masterToken  string
	adminUser    *models.User
	adminToken   string
	company1     *models.Company
	company2     *models.Company
	masterRole   *models.Role
	adminRole    *models.Role
	userRole     *models.Role
}

func TestHierarchySuite(t *testing.T) {
	suite.Run(t, new(HierarchyTestSuite))
}

func (s *HierarchyTestSuite) SetupSuite() {
	var err error
	s.testDB, err = testutils.SetupTestDB("hierarchy_test")
	s.Require().NoError(err)

	// Initialize repositories
	s.companyRepo = repository.NewCompanyRepository(s.testDB.DB)
	s.userRepo = repository.NewUserRepository(s.testDB.DB)
	s.roleRepo = repository.NewRoleRepository(s.testDB.DB)

	// Initialize token service
	s.tokenService = services.NewTokenService(
		s.testDB.SqlxDB,
		"test-secret-key",
		15*time.Minute,
		7*24*time.Hour,
	)

	// Create test roles
	ctx := context.Background()
	s.masterRole = &models.Role{ID: uuid.New(), Name: "master"}
	s.adminRole = &models.Role{ID: uuid.New(), Name: "admin"}
	s.userRole = &models.Role{ID: uuid.New(), Name: "user"}

	s.Require().NoError(s.testDB.DB.Create(s.masterRole).Error)
	s.Require().NoError(s.testDB.DB.Create(s.adminRole).Error)
	s.Require().NoError(s.testDB.DB.Create(s.userRole).Error)

	// Create test companies
	s.company1 = &models.Company{
		ID:               uuid.New(),
		Name:             "Company One",
		Slug:             "company-one",
		Email:            "contact@company1.com",
		Phone:            "1111111111",
		SubscriptionPlan: "basic",
	}
	s.Require().NoError(s.testDB.DB.Create(s.company1).Error)

	s.company2 = &models.Company{
		ID:               uuid.New(),
		Name:             "Company Two",
		Slug:             "company-two",
		Email:            "contact@company2.com",
		Phone:            "2222222222",
		SubscriptionPlan: "premium",
	}
	s.Require().NoError(s.testDB.DB.Create(s.company2).Error)

	// Create master user (no company)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Master123!"), bcrypt.DefaultCost)
	s.masterUser = &models.User{
		ID:       uuid.New(),
		Email:    "master@system.com",
		Password: string(hashedPassword),
		Name:     "Master User",
		Active:   true,
		RoleID:   s.masterRole.ID,
	}
	s.Require().NoError(s.testDB.DB.Create(s.masterUser).Error)

	// Create admin user for company1
	s.adminUser = &models.User{
		ID:        uuid.New(),
		Email:     "admin@company1.com",
		Password:  string(hashedPassword),
		Name:      "Admin User",
		Active:    true,
		RoleID:    s.adminRole.ID,
		CompanyID: &s.company1.ID,
	}
	s.Require().NoError(s.testDB.DB.Create(s.adminUser).Error)

	// Generate tokens
	accessToken, _, err := s.tokenService.GenerateTokenPair(ctx, s.masterUser, "127.0.0.1", "test-agent")
	s.Require().NoError(err)
	s.masterToken = accessToken

	accessToken, _, err = s.tokenService.GenerateTokenPair(ctx, s.adminUser, "127.0.0.1", "test-agent")
	s.Require().NoError(err)
	s.adminToken = accessToken

	// Setup router
	gin.SetMode(gin.TestMode)
	s.router = gin.New()

	companyHandler := handlers.NewCompanyHandler(s.companyRepo)
	authMiddleware := handlers.NewAuthMiddleware(s.tokenService, s.userRepo)

	api := s.router.Group("/api/v1")
	api.Use(authMiddleware.RequireAuth())
	{
		companies := api.Group("/companies")
		{
			companies.POST("", authMiddleware.RequireRole("master"), companyHandler.CreateCompany)
			companies.GET("", companyHandler.GetCompanies)
			companies.GET("/:id", companyHandler.GetCompanyByID)
			companies.PUT("/:id", companyHandler.UpdateCompany)
			companies.DELETE("/:id", authMiddleware.RequireRole("master"), companyHandler.DeleteCompany)
		}
	}
}

func (s *HierarchyTestSuite) TearDownSuite() {
	s.testDB.Close()
}

// TestGetCompanies_Master tests that master can see all companies
func (s *HierarchyTestSuite) TestGetCompanies_Master() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/companies", nil)
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var companies []models.Company
	err := json.Unmarshal(w.Body.Bytes(), &companies)
	s.NoError(err)
	s.GreaterOrEqual(len(companies), 2) // At least company1 and company2
}

// TestGetCompanies_Admin tests that admin can only see their own company
func (s *HierarchyTestSuite) TestGetCompanies_Admin() {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/companies", nil)
	req.Header.Set("Authorization", "Bearer "+s.adminToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var companies []models.Company
	err := json.Unmarshal(w.Body.Bytes(), &companies)
	s.NoError(err)
	// Admin should only see their company
	s.Equal(1, len(companies))
	s.Equal(s.company1.ID, companies[0].ID)
}

// TestGetCompanyByID_Master tests that master can access any company
func (s *HierarchyTestSuite) TestGetCompanyByID_Master() {
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/companies/%s", s.company2.ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var company models.Company
	err := json.Unmarshal(w.Body.Bytes(), &company)
	s.NoError(err)
	s.Equal(s.company2.ID, company.ID)
	s.Equal("Company Two", company.Name)
}

// TestGetCompanyByID_Admin_OwnCompany tests admin accessing their own company
func (s *HierarchyTestSuite) TestGetCompanyByID_Admin_OwnCompany() {
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/companies/%s", s.company1.ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.adminToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var company models.Company
	err := json.Unmarshal(w.Body.Bytes(), &company)
	s.NoError(err)
	s.Equal(s.company1.ID, company.ID)
}

// TestGetCompanyByID_Admin_OtherCompany tests admin accessing another company (should fail)
func (s *HierarchyTestSuite) TestGetCompanyByID_Admin_OtherCompany() {
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/companies/%s", s.company2.ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.adminToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusForbidden, w.Code)
}

// TestCreateCompany_Master tests that master can create companies
func (s *HierarchyTestSuite) TestCreateCompany_Master() {
	createReq := models.CreateCompanyRequest{
		Name:             "New Company",
		Slug:             "new-company",
		Email:            "contact@newcompany.com",
		Phone:            "3333333333",
		SubscriptionPlan: "basic",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/companies", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusCreated, w.Code)

	var company models.Company
	err := json.Unmarshal(w.Body.Bytes(), &company)
	s.NoError(err)
	s.Equal("New Company", company.Name)
	s.Equal("new-company", company.Slug)
}

// TestCreateCompany_Admin tests that admin cannot create companies
func (s *HierarchyTestSuite) TestCreateCompany_Admin() {
	createReq := models.CreateCompanyRequest{
		Name:             "Forbidden Company",
		Slug:             "forbidden-company",
		Email:            "contact@forbidden.com",
		Phone:            "4444444444",
		SubscriptionPlan: "basic",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/companies", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.adminToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusForbidden, w.Code)
}

// TestCreateCompany_DuplicateSlug tests creating company with existing slug
func (s *HierarchyTestSuite) TestCreateCompany_DuplicateSlug() {
	createReq := models.CreateCompanyRequest{
		Name:             "Duplicate Slug Company",
		Slug:             s.company1.Slug, // Existing slug
		Email:            "contact@duplicate.com",
		Phone:            "5555555555",
		SubscriptionPlan: "basic",
	}

	body, _ := json.Marshal(createReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/companies", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusConflict, w.Code)
}

// TestUpdateCompany_Admin tests that admin can update their own company
func (s *HierarchyTestSuite) TestUpdateCompany_Admin() {
	updateReq := models.UpdateCompanyRequest{
		Name:  strPtr("Company One Updated"),
		Phone: strPtr("9999999999"),
	}

	body, _ := json.Marshal(updateReq)
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/companies/%s", s.company1.ID), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.adminToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	var company models.Company
	err := json.Unmarshal(w.Body.Bytes(), &company)
	s.NoError(err)
	s.Equal("Company One Updated", company.Name)
}

// TestDeleteCompany_Master tests that master can delete companies
func (s *HierarchyTestSuite) TestDeleteCompany_Master() {
	// Create a company to delete
	companyToDelete := &models.Company{
		ID:               uuid.New(),
		Name:             "To Delete Company",
		Slug:             "to-delete",
		Email:            "delete@test.com",
		Phone:            "6666666666",
		SubscriptionPlan: "basic",
	}
	s.Require().NoError(s.testDB.DB.Create(companyToDelete).Error)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/companies/%s", companyToDelete.ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.masterToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusOK, w.Code)

	// Verify company is soft-deleted
	var deletedCompany models.Company
	err := s.testDB.DB.Unscoped().First(&deletedCompany, companyToDelete.ID).Error
	s.NoError(err)
	s.NotNil(deletedCompany.DeletedAt)
}

// TestDeleteCompany_Admin tests that admin cannot delete companies
func (s *HierarchyTestSuite) TestDeleteCompany_Admin() {
	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/companies/%s", s.company1.ID), nil)
	req.Header.Set("Authorization", "Bearer "+s.adminToken)
	w := httptest.NewRecorder()

	s.router.ServeHTTP(w, req)

	s.Equal(http.StatusForbidden, w.Code)
}

// Helper function
func strPtr(s string) *string {
	return &s
}
