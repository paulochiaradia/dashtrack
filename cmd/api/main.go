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
			logger.Warn("Failed to initialize tracing", zap.Error(err))
		} else {
			logger.Info("Tracing initialized", zap.String("service", cfg.AppName))
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
			"GET /health", "GET /metrics", "POST /auth/login",
			"POST /auth/refresh", "POST /auth/logout",
			"POST /auth/forgot-password", "POST /auth/reset-password",
		}),
		zap.Strings("authenticated_endpoints", []string{
			"GET /profile", "POST /profile/change-password", "GET /roles",
		}),
		zap.Strings("manager_endpoints", []string{
			"GET /manager/users",
		}),
		zap.Strings("admin_endpoints", []string{
			"GET /admin/users", "POST /admin/users", "GET /admin/users/:id",
			"PUT /admin/users/:id", "DELETE /admin/users/:id",
		}),
	)

	// Start server
	server := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router.GetEngine(),
	}

	logger.Info("HTTP server starting", zap.String("address", server.Addr))
	logger.Fatal("Server stopped", zap.Error(server.ListenAndServe()))
}
