package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/paulochiaradia/dashtrack/internal/auth"
	"github.com/paulochiaradia/dashtrack/internal/config"
	"github.com/paulochiaradia/dashtrack/internal/database"
	"github.com/paulochiaradia/dashtrack/internal/handlers"
	"github.com/paulochiaradia/dashtrack/internal/repository"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.LoadConfig()

	// Initialize observability (with simplified logger for now)
	log.Printf("Starting Dashtrack API v1.0.0")
	log.Printf("Environment: %s", "development") // TODO: Add to config

	db := database.NewDatabase(cfg.DBSource)
	defer db.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	// authLogRepo := repository.NewAuthLogRepository(db) // TODO: Use for authentication logging

	// Initialize JWT manager
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-jwt-key-change-in-production" // Default for development
		log.Println("Warning: Using default JWT secret. Set JWT_SECRET environment variable in production.")
	}
	jwtManager := auth.NewJWTManager(jwtSecret, 15*time.Minute, 24*time.Hour, "dashtrack-api")

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userRepo, roleRepo)
	roleHandler := handlers.NewRoleHandler(roleRepo)
	authHandler := handlers.NewAuthHandler(userRepo, jwtManager)

	// Setup routes
	mux := http.NewServeMux()

	// Observability endpoints
	mux.Handle("/metrics", promhttp.Handler()) // Prometheus metrics

	// Health check with more details
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{
			"status":"ok",
			"message":"API is running",
			"version":"1.0.0",
			"database":"connected",
			"timestamp":"%s"
		}`, "2023-01-01T00:00:00Z") // TODO: Use actual timestamp
	})

	// Authentication routes (public)
	mux.HandleFunc("/auth/login", authHandler.Login)
	mux.HandleFunc("/auth/refresh", authHandler.RefreshToken)
	mux.HandleFunc("/auth/logout", authHandler.Logout)

	// User profile routes (authenticated)
	authMiddleware := auth.NewAuthMiddleware(jwtManager)
	mux.Handle("/profile", authMiddleware.RequireAuth(http.HandlerFunc(authHandler.Me)))
	mux.Handle("/profile/change-password", authMiddleware.RequireAuth(http.HandlerFunc(authHandler.ChangePassword)))

	// Admin routes (admin only)
	mux.Handle("/admin/users", authMiddleware.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.ListUsers(w, r)
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.Handle("/admin/users/", authMiddleware.RequireRole("admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path
		path := strings.TrimPrefix(r.URL.Path, "/admin/users/")
		if path == "" {
			http.Error(w, "User ID required", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			userHandler.GetUser(w, r)
		case http.MethodPut:
			userHandler.UpdateUser(w, r)
		case http.MethodDelete:
			userHandler.DeleteUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Manager routes (admin and manager only)
	mux.Handle("/manager/users", authMiddleware.RequireAnyRole("admin", "manager")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			userHandler.ListUsers(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// Role routes (authenticated users can view roles)
	mux.Handle("/roles", authMiddleware.RequireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			roleHandler.ListRoles(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	// CORS middleware
	corsHandler := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			h.ServeHTTP(w, r)
		})
	}

	server := &http.Server{
		Addr:    ":" + cfg.APIPort,
		Handler: corsHandler(mux),
	}

	log.Printf("Server starting on port %s", cfg.APIPort)
	log.Printf("Database connected successfully")
	log.Printf("JWT authentication enabled")
	log.Printf("Observability enabled - metrics available at /metrics")
	log.Printf("Available endpoints:")
	log.Printf("  Public endpoints:")
	log.Printf("    GET    /health")
	log.Printf("    GET    /metrics")
	log.Printf("    POST   /auth/login")
	log.Printf("    POST   /auth/refresh")
	log.Printf("    POST   /auth/logout")
	log.Printf("  Authenticated endpoints:")
	log.Printf("    GET    /profile")
	log.Printf("    POST   /profile/change-password")
	log.Printf("    GET    /roles")
	log.Printf("  Manager endpoints (admin/manager):")
	log.Printf("    GET    /manager/users")
	log.Printf("  Admin endpoints (admin only):")
	log.Printf("    GET    /admin/users")
	log.Printf("    POST   /admin/users")
	log.Printf("    GET    /admin/users/{id}")
	log.Printf("    PUT    /admin/users/{id}")
	log.Printf("    DELETE /admin/users/{id}")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
