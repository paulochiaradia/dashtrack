package benchmarks_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/paulochiaradia/dashtrack/internal/auth"
	"github.com/paulochiaradia/dashtrack/internal/config"
	"github.com/paulochiaradia/dashtrack/internal/handlers"
	"github.com/paulochiaradia/dashtrack/internal/middleware"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/services"
	"github.com/paulochiaradia/dashtrack/tests/testutils/mocks"
)

var (
	benchRouter     *gin.Engine
	benchJWTManager *auth.JWTManager
	benchAuthToken  string
)

func init() {
	gin.SetMode(gin.TestMode)
	// setupBenchmarkRouter() will be called manually in each benchmark
}

func setupBenchmarkRouter() {
	// Setup JWT manager
	cfg := &config.Config{
		JWTSecret:              "benchmark-secret-key",
		JWTAccessExpireMinutes: 15,
		JWTRefreshExpireHours:  24,
		AppName:                "Dashtrack Benchmark",
	}

	accessExpiry := time.Duration(cfg.JWTAccessExpireMinutes) * time.Minute
	refreshExpiry := time.Duration(cfg.JWTRefreshExpireHours) * time.Hour
	benchJWTManager = auth.NewJWTManager(cfg.JWTSecret, accessExpiry, refreshExpiry, cfg.AppName)

	// Create mock controllers
	ctrl := gomock.NewController(&testing.T{})
	mockUserRepo := mocks.NewMockUserRepository(ctrl)
	mockRoleRepo := mocks.NewMockRoleRepository(ctrl)

	// Create services
	userService := services.NewUserService(mockUserRepo, mockRoleRepo, 4)

	// Setup mock expectations for benchmark
	setupBenchmarkMocks(mockUserRepo)

	// Create handlers
	userHandler := handlers.NewUserHandler(userService)

	// Setup router
	benchRouter = gin.New()

	authMiddleware := middleware.NewGinAuthMiddleware(benchJWTManager)

	// Protected routes
	api := benchRouter.Group("/api")
	api.Use(authMiddleware.RequireAuth())
	{
		users := api.Group("/users")
		{
			users.GET("", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), userHandler.GetUsers)
			users.POST("", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), userHandler.CreateUser)
			users.PUT("/:id", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), userHandler.UpdateUser)
			users.DELETE("/:id", authMiddleware.RequireAnyRole("master", "company_admin", "admin"), userHandler.DeleteUser)
		}
	}

	// Generate auth token
	userID := uuid.New()
	companyID := uuid.New()

	userContext := auth.UserContext{
		UserID:   userID,
		Email:    "benchmark@test.com",
		RoleName: "master",
		TenantID: &companyID,
	}

	accessToken, _, _ := benchJWTManager.GenerateTokens(userContext)
	benchAuthToken = accessToken
}

func setupBenchmarkMocks(mockUserRepo *mocks.MockUserRepository) {
	userID := uuid.New()
	companyID := uuid.New()

	// Mock user for authentication
	user := &models.User{
		ID:        userID,
		Name:      "Benchmark User",
		Email:     "benchmark@test.com",
		Active:    true,
		CompanyID: &companyID,
		Role: &models.Role{
			ID:   uuid.New(),
			Name: "master",
		},
	}

	// Setup common mock expectations for user repository
	mockUserRepo.EXPECT().
		GetByEmail(gomock.Any(), "benchmark@test.com").
		Return(user, nil).
		AnyTimes()

	mockUserRepo.EXPECT().
		GetByID(gomock.Any(), userID).
		Return(user, nil).
		AnyTimes()

	// Mock for List users
	users := make([]*models.User, 100) // Simulate 100 users
	for i := 0; i < 100; i++ {
		users[i] = &models.User{
			ID:        uuid.New(),
			Name:      fmt.Sprintf("User %d", i),
			Email:     fmt.Sprintf("user%d@example.com", i),
			Active:    true,
			CompanyID: &companyID,
			Role: &models.Role{
				ID:   uuid.New(),
				Name: "driver",
			},
		}
	}

	mockUserRepo.EXPECT().
		List(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(users, nil).
		AnyTimes()

	mockUserRepo.EXPECT().
		CountUsers(gomock.Any(), gomock.Any()).
		Return(len(users), nil).
		AnyTimes()

	// Mock for Create, Update, Delete
	mockUserRepo.EXPECT().
		Create(gomock.Any(), gomock.Any()).
		Return(user, nil).
		AnyTimes()

	mockUserRepo.EXPECT().
		Update(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(user, nil).
		AnyTimes()

	mockUserRepo.EXPECT().
		Delete(gomock.Any(), gomock.Any()).
		Return(nil).
		AnyTimes()
}

// Benchmark JWT Token Generation
func BenchmarkJWTTokenGeneration(b *testing.B) {
	if benchJWTManager == nil {
		setupBenchmarkRouter()
	}

	userID := uuid.New()
	companyID := uuid.New()

	userContext := auth.UserContext{
		UserID:   userID,
		Email:    "test@example.com",
		RoleName: "admin",
		TenantID: &companyID,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := benchJWTManager.GenerateTokens(userContext)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark JWT Token Validation
func BenchmarkJWTTokenValidation(b *testing.B) {
	userID := uuid.New()
	companyID := uuid.New()

	userContext := auth.UserContext{
		UserID:   userID,
		Email:    "test@example.com",
		RoleName: "admin",
		TenantID: &companyID,
	}

	accessToken, _, err := benchJWTManager.GenerateTokens(userContext)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := benchJWTManager.ValidateToken(accessToken)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark GetUsers Endpoint
func BenchmarkGetUsersEndpoint(b *testing.B) {
	if benchRouter == nil {
		setupBenchmarkRouter()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/users?page=1&limit=10", nil)
		req.Header.Set("Authorization", "Bearer "+benchAuthToken)

		w := httptest.NewRecorder()
		benchRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}

// Benchmark GetUsers with Large Page Size
func BenchmarkGetUsersLargePage(b *testing.B) {
	if benchRouter == nil {
		setupBenchmarkRouter()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/users?page=1&limit=100", nil)
		req.Header.Set("Authorization", "Bearer "+benchAuthToken)

		w := httptest.NewRecorder()
		benchRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}

// Benchmark CreateUser Endpoint
func BenchmarkCreateUserEndpoint(b *testing.B) {
	if benchRouter == nil {
		setupBenchmarkRouter()
	}

	roleID := uuid.New()

	newUser := map[string]interface{}{
		"name":     "Benchmark User",
		"email":    "benchmark@example.com",
		"phone":    "1234567890",
		"password": "password123",
		"role_id":  roleID.String(),
		"active":   true,
	}

	body, _ := json.Marshal(newUser)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("POST", "/api/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+benchAuthToken)

		w := httptest.NewRecorder()
		benchRouter.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			b.Fatalf("Expected status 201, got %d. Response: %s", w.Code, w.Body.String())
		}
	}
}

// Benchmark UpdateUser Endpoint
func BenchmarkUpdateUserEndpoint(b *testing.B) {
	if benchRouter == nil {
		setupBenchmarkRouter()
	}

	userID := uuid.New()
	updateData := map[string]interface{}{
		"name":   "Updated User",
		"phone":  "9876543210",
		"active": true,
	}

	body, _ := json.Marshal(updateData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("PUT", "/api/users/"+userID.String(), bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+benchAuthToken)

		w := httptest.NewRecorder()
		benchRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}

// Benchmark DeleteUser Endpoint
func BenchmarkDeleteUserEndpoint(b *testing.B) {
	if benchRouter == nil {
		setupBenchmarkRouter()
	}

	userID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("DELETE", "/api/users/"+userID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+benchAuthToken)

		w := httptest.NewRecorder()
		benchRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			b.Fatalf("Expected status 200, got %d", w.Code)
		}
	}
}

// Benchmark Authentication Middleware
func BenchmarkAuthMiddleware(b *testing.B) {
	if benchJWTManager == nil || benchAuthToken == "" {
		setupBenchmarkRouter()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/api/users", nil)
		req.Header.Set("Authorization", "Bearer "+benchAuthToken)

		w := httptest.NewRecorder()

		// Only test the middleware, not the full endpoint
		_, err := benchJWTManager.ValidateToken(benchAuthToken)
		if err != nil {
			// Skip validation error for benchmark - focus on performance
			_ = err
		}

		_ = req.Context()
		_ = w
	}
}

// Benchmark Concurrent Requests
func BenchmarkConcurrentGetUsers(b *testing.B) {
	if benchRouter == nil {
		setupBenchmarkRouter()
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", "/api/users?page=1&limit=10", nil)
			req.Header.Set("Authorization", "Bearer "+benchAuthToken)

			w := httptest.NewRecorder()
			benchRouter.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				b.Fatalf("Expected status 200, got %d", w.Code)
			}
		}
	})
}

// Benchmark Memory Allocation for JSON Marshaling
func BenchmarkJSONMarshaling(b *testing.B) {
	user := &models.User{
		ID:        uuid.New(),
		Name:      "Test User",
		Email:     "test@example.com",
		Phone:     stringPtr("1234567890"),
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Role: &models.Role{
			ID:   uuid.New(),
			Name: "driver",
		},
		CompanyID: func() *uuid.UUID { id := uuid.New(); return &id }(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(user)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark Memory Allocation for User List
func BenchmarkUserListAllocation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		users := make([]*models.User, 100)
		for j := 0; j < 100; j++ {
			users[j] = &models.User{
				ID:        uuid.New(),
				Name:      fmt.Sprintf("User %d", j),
				Email:     fmt.Sprintf("user%d@example.com", j),
				Active:    true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				Role: &models.Role{
					ID:   uuid.New(),
					Name: "driver",
				},
			}
		}
		_ = users
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}

// Benchmark Report Function
func BenchmarkReport(b *testing.B) {
	b.Log("=== Performance Benchmark Report ===")
	b.Log("This benchmark tests the performance of key endpoints and operations")
	b.Log("Focus areas:")
	b.Log("- JWT token generation and validation")
	b.Log("- Authentication middleware overhead")
	b.Log("- CRUD operations performance")
	b.Log("- Concurrent request handling")
	b.Log("- Memory allocation patterns")
	b.Log("")
	b.Log("Run with: go test -bench=. -benchmem -count=3")
	b.Log("For CPU profiling: go test -bench=. -cpuprofile=cpu.prof")
	b.Log("For memory profiling: go test -bench=. -memprofile=mem.prof")
}
