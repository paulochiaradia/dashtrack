package routes

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/paulochiaradia/dashtrack/internal/auth"
	"github.com/paulochiaradia/dashtrack/internal/config"
	"github.com/paulochiaradia/dashtrack/internal/handlers"
	"github.com/paulochiaradia/dashtrack/internal/middleware"
	"github.com/paulochiaradia/dashtrack/internal/repository"
	"github.com/paulochiaradia/dashtrack/internal/services"
)

// Router struct holds all dependencies for the router
type Router struct {
	engine           *gin.Engine
	cfg              *config.Config
	db               *sqlx.DB
	authHandler      *handlers.AuthHandler
	userHandler      *handlers.UserHandler
	sensorHandler    *handlers.SensorHandler
	companyHandler   *handlers.CompanyHandler
	teamHandler      *handlers.TeamHandler
	vehicleHandler   *handlers.VehicleHandler
	esp32Handler     *handlers.ESP32DeviceHandler
	securityHandler  *handlers.SecurityHandler
	sessionHandler   *handlers.SessionHandler
	dashboardHandler *handlers.DashboardHandler
	jwtManager       *auth.JWTManager
	authMiddleware   *middleware.GinAuthMiddleware
}

// NewRouter creates and configures a new router
func NewRouter(db *sql.DB, cfg *config.Config) *Router {
	if cfg.ServerEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	sqlxDB := sqlx.NewDb(db, "postgres")

	// Repositories
	userRepo := repository.NewUserRepository(sqlxDB)
	roleRepo := repository.NewRoleRepository(db)
	authLogRepo := repository.NewAuthLogRepository(db)
	sensorRepo := repository.NewSensorRepository(sqlxDB)
	companyRepo := repository.NewCompanyRepository(sqlxDB)
	teamRepo := repository.NewTeamRepository(sqlxDB)
	vehicleRepo := repository.NewVehicleRepository(sqlxDB)
	esp32Repo := repository.NewESP32DeviceRepository(sqlxDB)

	sessionRepo := repository.NewSessionRepository(sqlxDB)

	// Services
	accessExpiry := time.Duration(cfg.JWTAccessExpireMinutes) * time.Minute
	refreshExpiry := time.Duration(cfg.JWTRefreshExpireHours) * time.Hour
	jwtManager := auth.NewJWTManager(cfg.JWTSecret, accessExpiry, refreshExpiry, cfg.AppName)
	tokenService := services.NewTokenService(sqlxDB, cfg.JWTSecret, accessExpiry, refreshExpiry)
	twoFactorService := services.NewTwoFactorService(sqlxDB)
	auditService := services.NewAuditService(sqlxDB)
	sessionManager := services.NewSessionManager(sqlxDB)
	userService := services.NewUserService(userRepo, roleRepo, cfg.BcryptCost)

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo, authLogRepo, jwtManager, tokenService, cfg.BcryptCost)
	userHandler := handlers.NewUserHandler(userService)
	sensorHandler := handlers.NewSensorHandler(sensorRepo)
	companyHandler := handlers.NewCompanyHandler(companyRepo)
	teamHandler := handlers.NewTeamHandler(teamRepo, userRepo)
	vehicleHandler := handlers.NewVehicleHandler(vehicleRepo, teamRepo)
	esp32Handler := handlers.NewESP32DeviceHandler(esp32Repo, vehicleRepo)
	securityHandler := handlers.NewSecurityHandler(tokenService, twoFactorService, auditService)
	sessionHandler := handlers.NewSessionHandler(sessionManager)
	dashboardHandler := handlers.NewDashboardHandler(userRepo, authLogRepo, sessionRepo, companyRepo)

	// Middleware
	authMiddleware := middleware.NewGinAuthMiddleware(jwtManager)

	router := &Router{
		engine:           gin.New(),
		cfg:              cfg,
		db:               sqlxDB,
		authHandler:      authHandler,
		userHandler:      userHandler,
		sensorHandler:    sensorHandler,
		companyHandler:   companyHandler,
		teamHandler:      teamHandler,
		vehicleHandler:   vehicleHandler,
		esp32Handler:     esp32Handler,
		securityHandler:  securityHandler,
		sessionHandler:   sessionHandler,
		dashboardHandler: dashboardHandler,
		jwtManager:       jwtManager,
		authMiddleware:   authMiddleware,
	}

	router.setupMiddleware()
	router.setupRoutes()

	return router
}

func (r *Router) setupMiddleware() {
	r.engine.Use(gin.Recovery())
	// TODO: Add other middlewares when they are implemented
	// r.engine.Use(middleware.GinLoggingMiddleware())
	// r.engine.Use(middleware.CORSMiddleware())
	// r.engine.Use(middleware.RateLimitMiddleware())
	// r.engine.Use(middleware.SecurityHeaders())
}

func (r *Router) setupRoutes() {
	// API v1 routes
	v1 := r.engine.Group("/api/v1")

	// Public routes (no authentication required)
	public := v1.Group("/auth")
	{
		public.POST("/login", r.authHandler.LoginGin)
		public.POST("/refresh", r.authHandler.RefreshTokenGin)
	}

	// Protected routes (authentication required)
	protected := v1.Group("")
	protected.Use(r.authMiddleware.RequireAuth())
	{
		// Auth routes
		protected.POST("/auth/logout", r.authHandler.LogoutGin)
		protected.POST("/auth/change-password", r.authHandler.ChangePasswordGin)

		// User routes with role-based access
		userRoutes := protected.Group("/users")
		{
			userRoutes.GET("", r.userHandler.GetUsers)          // List users
			userRoutes.GET("/:id", r.userHandler.GetUserByID)   // Get user by ID
			userRoutes.PUT("/:id", r.userHandler.UpdateUser)    // Update user
			userRoutes.DELETE("/:id", r.userHandler.DeleteUser) // Delete user
		}

		// Admin and Company Admin routes (roles that can create users)
		adminRoutes := protected.Group("")
		adminRoutes.Use(r.authMiddleware.RequireAnyRole("admin", "company_admin"))
		{
			adminRoutes.POST("/users", r.userHandler.CreateUser) // Create user
		}
		// Master-only routes
		masterRoutes := protected.Group("")
		masterRoutes.Use(r.authMiddleware.RequireRole("master"))
		{
			// Master can access all companies' data
		}
	}
	// Setup additional role-based routes
	r.setupProtectedRoutes()
	r.setupMasterRoutes()
	r.setupCompanyAdminRoutes()
	r.setupAdminRoutes()
	r.setupManagerRoutes()
	r.setupHealthRoutes()
	r.setupSecurityRoutes()
	r.setupSessionRoutes()
}

// Engine returns the gin engine
func (r *Router) Engine() *gin.Engine {
	return r.engine
}
