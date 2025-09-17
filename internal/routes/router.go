package routes

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/paulochiaradia/dashtrack/internal/auth"
	"github.com/paulochiaradia/dashtrack/internal/config"
	"github.com/paulochiaradia/dashtrack/internal/handlers"
	"github.com/paulochiaradia/dashtrack/internal/middleware"
	"github.com/paulochiaradia/dashtrack/internal/repository"
)

type Router struct {
	engine      *gin.Engine
	cfg         *config.Config
	authHandler *handlers.AuthHandler
	jwtManager  *auth.JWTManager
}

func NewRouter(db *sql.DB, cfg *config.Config) *Router {
	// Set Gin mode based on environment
	if cfg.ServerEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	authLogRepo := repository.NewAuthLogRepository(db)

	// Initialize JWT manager with config
	accessExpiry := time.Duration(cfg.JWTAccessExpireMinutes) * time.Minute
	refreshExpiry := time.Duration(cfg.JWTRefreshExpireHours) * time.Hour
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, accessExpiry, refreshExpiry, cfg.AppName)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(userRepo, authLogRepo, jwtManager, cfg.BcryptCost)

	router := &Router{
		engine:      gin.New(),
		cfg:         cfg,
		authHandler: authHandler,
		jwtManager:  jwtManager,
	}

	router.setupMiddleware()
	router.setupRoutes()

	return router
}

func (r *Router) setupMiddleware() {
	// Structured logging middleware
	r.engine.Use(middleware.GinLoggingMiddleware())

	// Metrics middleware
	r.engine.Use(middleware.GinMetricsMiddleware())

	// Tracing middleware
	r.engine.Use(middleware.GinTracingMiddleware())

	// Recovery middleware
	r.engine.Use(gin.Recovery())

	// CORS middleware
	r.engine.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
}

func (r *Router) setupRoutes() {
	// Health and monitoring
	r.setupHealthRoutes()

	// Authentication routes
	r.setupAuthRoutes()

	// Protected routes
	r.setupProtectedRoutes()

	// Admin routes
	r.setupAdminRoutes()

	// Manager routes
	r.setupManagerRoutes()
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
