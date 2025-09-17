package routes

import (
	"github.com/paulochiaradia/dashtrack/internal/middleware"
)

func (r *Router) setupAdminRoutes() {
	// Create Gin middleware from auth middleware
	authMiddleware := middleware.NewGinAuthMiddleware(r.jwtManager)

	// Admin routes (admin only)
	admin := r.engine.Group("/admin")
	admin.Use(authMiddleware.RequireAuth())
	admin.Use(authMiddleware.RequireRole("admin"))

	// User management
	admin.GET("/users", r.authHandler.GetUsersGin)
	admin.POST("/users", r.authHandler.CreateUserGin)
	admin.GET("/users/:id", r.authHandler.GetUserByIDGin)
	admin.PUT("/users/:id", r.authHandler.UpdateUserGin)
	admin.DELETE("/users/:id", r.authHandler.DeleteUserGin)

	// Store management (TODO: implement handlers)
	// admin.GET("/stores", r.storeHandler.GetStoresGin)
	// admin.POST("/stores", r.storeHandler.CreateStoreGin)
	// admin.PUT("/stores/:id", r.storeHandler.UpdateStoreGin)
	// admin.DELETE("/stores/:id", r.storeHandler.DeleteStoreGin)
}
