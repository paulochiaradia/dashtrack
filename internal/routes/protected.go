package routes

func (r *Router) setupProtectedRoutes() {
	// Create Gin middleware from auth middleware
	authMiddleware := r.authMiddleware

	// Protected routes (require authentication)
	protected := r.engine.Group("/api/v1")
	protected.Use(authMiddleware.RequireAuth())
	protected.GET("/profile", r.authHandler.MeGin)
	protected.POST("/profile/change-password", r.authHandler.ChangePasswordGin)
	protected.GET("/roles", r.authHandler.GetRolesGin)

	// Dashboard for all authenticated users (role-based filtering happens inside handler)
	protected.GET("/dashboard", r.dashboardHandler.GetDashboard)
}
