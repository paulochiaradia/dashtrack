package routes

import (
	"github.com/paulochiaradia/dashtrack/internal/middleware"
)

func (r *Router) setupMasterRoutes() {
	// Create Gin middleware from auth middleware
	authMiddleware := middleware.NewGinAuthMiddleware(r.jwtManager)

	// Master routes (super-admin/master only)
	// Master = System Owner (full access to everything)
	master := r.engine.Group("/api/v1/master")
	master.Use(authMiddleware.RequireAuth())
	master.Use(authMiddleware.RequireRole("master")) // ONLY master role

	// Full User Management (master-only - can manage ALL users)
	master.GET("/users", r.userHandler.GetUsers)
	master.POST("/users", r.userHandler.CreateUser)
	master.GET("/users/:id", r.userHandler.GetUserByID)
	master.PUT("/users/:id", r.userHandler.UpdateUser)
	master.DELETE("/users/:id", r.userHandler.DeleteUser)

	// Company Management (master-only)
	// TODO: implement company handlers
	// master.GET("/companies", r.companyHandler.GetCompanies)
	// master.POST("/companies", r.companyHandler.CreateCompany)
	// master.GET("/companies/:id", r.companyHandler.GetCompany)
	// master.PUT("/companies/:id", r.companyHandler.UpdateCompany)
	// master.DELETE("/companies/:id", r.companyHandler.DeleteCompany)

	// System-wide Analytics (master-only)
	// TODO: implement analytics handlers
	// master.GET("/analytics/users", r.analyticsHandler.GetUserAnalytics)
	// master.GET("/analytics/companies", r.analyticsHandler.GetCompanyAnalytics)
	// master.GET("/analytics/system", r.analyticsHandler.GetSystemAnalytics)
}
