package benchmarks_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/models"
)

func BenchmarkHealthEndpoint(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"status":  "ok",
			"message": "API is running",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", "/health", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
		}
	})
}

func BenchmarkGetUsers(b *testing.B) {
	users := make([]models.User, 100)
	for i := 0; i < 100; i++ {
		users[i] = models.User{
			ID:     uuid.New(),
			Name:   "Test User",
			Email:  "test@example.com",
			Active: true,
		}
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(users)
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("GET", "/users", nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
		}
	})
}

func BenchmarkCreateUser(b *testing.B) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var createReq models.CreateUserRequest
		json.NewDecoder(r.Body).Decode(&createReq)
		
		user := models.User{
			ID:     uuid.New(),
			Name:   createReq.Name,
			Email:  createReq.Email,
			Active: true,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	})

	createReq := models.CreateUserRequest{
		Name:     "Benchmark User",
		Email:    "benchmark@example.com",
		Password: "password123",
		RoleID:   uuid.New().String(),
	}

	body, _ := json.Marshal(createReq)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
		}
	})
}

func BenchmarkJSONEncoding(b *testing.B) {
	user := models.User{
		ID:     uuid.New(),
		Name:   "Test User",
		Email:  "test@example.com",
		Active: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(user)
	}
}

func BenchmarkUUIDGeneration(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		uuid.New()
	}
}
