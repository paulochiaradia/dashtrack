package routes

func (r *Router) setupCompanyAdminRoutes() {
	// Use router's auth middleware (already configured with tokenService)
	authMiddleware := r.authMiddleware

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

	// NOTE: Team management routes moved to internal/routes/team.go (r.setupTeamRoutes)
	// NOTE: Vehicle management routes will be implemented separately
}
