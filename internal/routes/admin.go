package routes

import (
	"github.com/paulochiaradia/dashtrack/internal/middleware"
)

func (r *Router) setupAdminRoutes() {
	// Create Gin middleware from auth middleware
	authMiddleware := middleware.NewGinAuthMiddleware(r.jwtManager)

	// Admin routes (technical/operational administration)
	// Admin = System Administrator (technical operations)
	admin := r.engine.Group("/api/v1/admin")
	admin.Use(authMiddleware.RequireAuth())
	admin.Use(authMiddleware.RequireAdminRole()) // ONLY admin role (not master)

	// Technical User Management (admin-only)
	admin.GET("/users", r.userHandler.GetUsers)
	admin.POST("/users", r.userHandler.CreateUser)
	admin.GET("/users/:id", r.userHandler.GetUserByID)
	admin.PUT("/users/:id", r.userHandler.UpdateUser)
	admin.DELETE("/users/:id", r.userHandler.DeleteUser)

	// System Configuration (admin-only)
	// TODO: implement system config handlers
	// admin.GET("/system/config", r.systemHandler.GetSystemConfig)
	// admin.PUT("/system/config", r.systemHandler.UpdateSystemConfig)

	// Technical Monitoring (admin-only)
	// TODO: implement monitoring handlers
	// admin.GET("/system/health", r.systemHandler.GetHealthStatus)
	// admin.GET("/system/metrics", r.systemHandler.GetSystemMetrics)
	// admin.GET("/system/logs", r.systemHandler.GetSystemLogs)

	// Security Management (admin-only)
	// TODO: implement security config handlers
	// admin.GET("/security/config", r.securityHandler.GetSecurityConfig)
	// admin.PUT("/security/config", r.securityHandler.UpdateSecurityConfig)
}
