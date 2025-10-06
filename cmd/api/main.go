package main

import (
	"log"
	"net/http"

	"github.com/paulochiaradia/dashtrack/internal/config"
	"github.com/paulochiaradia/dashtrack/internal/database"
	"github.com/paulochiaradia/dashtrack/internal/logger"
	"github.com/paulochiaradia/dashtrack/internal/routes"
	"github.com/paulochiaradia/dashtrack/internal/tracing"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()

	// Initialize structured logger
	err := logger.InitLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize tracing
	if cfg.ServerEnv != "test" { // Don't init tracing in test environment
		err = tracing.InitTracing(cfg.AppName, "http://jaeger:14268/api/traces")
		if err != nil {
			logger.Warn("Tracing disabled - Jaeger not available", zap.Error(err))
		} else {
			logger.Info("Tracing initialized successfully")
		}
	}

	// Initialize observability
	logger.Info("Starting application",
		zap.String("app_name", cfg.AppName),
		zap.String("version", cfg.AppVersion),
		zap.String("environment", cfg.ServerEnv),
	)

	db := database.NewDatabase(cfg.DBSource)
	defer db.Close()

	// Initialize router
	router := routes.NewRouter(db, cfg)

	// Log available endpoints
	logger.Info("Server configuration",
		zap.String("port", cfg.ServerPort),
		zap.String("database_status", "connected"),
		zap.Bool("jwt_enabled", true),
		zap.String("metrics_endpoint", "/metrics"),
	)

	logger.Info("Available endpoints documented",
		zap.Strings("public_endpoints", []string{
			"GET /health", "GET /metrics", "POST /api/v1/auth/login",
			"POST /api/v1/auth/refresh", "POST /api/v1/auth/logout",
			"POST /api/v1/auth/forgot-password", "POST /api/v1/auth/reset-password",
		}),
		zap.Strings("authenticated_endpoints", []string{
			"GET /api/v1/profile", "POST /api/v1/profile/change-password", "GET /api/v1/roles",
		}),
		zap.Strings("master_endpoints", []string{
			"POST /api/v1/master/companies", "GET /api/v1/master/companies", "DELETE /api/v1/master/companies/:id",
			"GET /api/v1/master/dashboard", "GET /api/v1/master/users", "POST /api/v1/master/users",
		}),
		zap.Strings("admin_endpoints", []string{
			"GET /api/v1/admin/users", "POST /api/v1/admin/users", "GET /api/v1/admin/users/:id",
			"PUT /api/v1/admin/users/:id", "DELETE /api/v1/admin/users/:id",
		}),
		zap.Strings("system_endpoints", []string{
			"GET /api/v1/system/users", "GET /api/v1/system/roles", "GET /api/v1/audit/logs",
		}),
		zap.Strings("manager_endpoints", []string{
			"GET /api/v1/manager/users",
		}),
	)

	// Start server
	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router.Engine(),
	}

	logger.Info("HTTP server starting", zap.String("address", server.Addr))
	logger.Fatal("Server stopped", zap.Error(server.ListenAndServe()))
}
