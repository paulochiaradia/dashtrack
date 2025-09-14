package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

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

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userRepo, roleRepo)
	roleHandler := handlers.NewRoleHandler(roleRepo)

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

	// User routes
	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			userHandler.ListUsers(w, r)
		case http.MethodPost:
			userHandler.CreateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path
		path := strings.TrimPrefix(r.URL.Path, "/users/")
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
	})

	// Role routes
	mux.HandleFunc("/roles", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			roleHandler.ListRoles(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

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
	log.Printf("Observability enabled - metrics available at /metrics")
	log.Printf("Available endpoints:")
	log.Printf("  GET    /health")
	log.Printf("  GET    /metrics")
	log.Printf("  GET    /roles")
	log.Printf("  GET    /users")
	log.Printf("  POST   /users")
	log.Printf("  GET    /users/{id}")
	log.Printf("  PUT    /users/{id}")
	log.Printf("  DELETE /users/{id}")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("could not start server: %s\n", err)
	}
}
