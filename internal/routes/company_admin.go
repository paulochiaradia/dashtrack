package routes

import (
	"github.com/paulochiaradia/dashtrack/internal/middleware"
)

func (r *Router) setupCompanyAdminRoutes() {
	// Create Gin middleware from auth middleware
	authMiddleware := middleware.NewGinAuthMiddleware(r.jwtManager)

	// Company Admin routes (company_admin only)
	// Company Admin = Company Administrator (manages their company)
	companyAdmin := r.engine.Group("/api/v1/company-admin")
	companyAdmin.Use(authMiddleware.RequireAuth())
	companyAdmin.Use(authMiddleware.RequireRole("company_admin")) // ONLY company_admin role

	// Company User Management (company_admin-only - can manage users in their company)
	companyAdmin.GET("/users", r.userHandler.GetUsers)
	companyAdmin.POST("/users", r.userHandler.CreateUser)
	companyAdmin.GET("/users/:id", r.userHandler.GetUserByID)
	companyAdmin.PUT("/users/:id", r.userHandler.UpdateUser)
	companyAdmin.DELETE("/users/:id", r.userHandler.DeleteUser)

	// Company Settings (company_admin-only)
	// TODO: implement company settings handlers
	// companyAdmin.GET("/settings", r.companyHandler.GetCompanySettings)
	// companyAdmin.PUT("/settings", r.companyHandler.UpdateCompanySettings)

	// Team Management (company_admin-only)
	// TODO: implement team handlers
	// companyAdmin.GET("/teams", r.teamHandler.GetTeams)
	// companyAdmin.POST("/teams", r.teamHandler.CreateTeam)
	// companyAdmin.GET("/teams/:id", r.teamHandler.GetTeam)
	// companyAdmin.PUT("/teams/:id", r.teamHandler.UpdateTeam)
	// companyAdmin.DELETE("/teams/:id", r.teamHandler.DeleteTeam)

	// Vehicle Management (company_admin-only)
	// TODO: implement vehicle handlers for company admin
	// companyAdmin.GET("/vehicles", r.vehicleHandler.GetCompanyVehicles)
	// companyAdmin.POST("/vehicles", r.vehicleHandler.CreateVehicle)
	// companyAdmin.GET("/vehicles/:id", r.vehicleHandler.GetVehicle)
	// companyAdmin.PUT("/vehicles/:id", r.vehicleHandler.UpdateVehicle)
	// companyAdmin.DELETE("/vehicles/:id", r.vehicleHandler.DeleteVehicle)
}
