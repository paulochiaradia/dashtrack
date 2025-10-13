package routes

// setupSystemRoutes sets up routes accessible by both master and admin
func (r *Router) setupSystemRoutes() {
	// Create Gin middleware from auth middleware
	authMiddleware := r.authMiddleware

	// System routes (shared between master and admin)
	// Both roles can access these, but with different purposes:
	// - Master: Business oversight and decision making
	// - Admin: Technical monitoring and troubleshooting
	system := r.engine.Group("/api/v1/system")
	system.Use(authMiddleware.RequireAuth())
	system.Use(authMiddleware.RequireAnyRole("admin", "master"))
	{
		// User information (both can view, but different contexts)
		system.GET("/users", r.userHandler.GetUsers)
		system.GET("/users/:id", r.userHandler.GetUserByID)

		// Role management (both need to understand roles)
		system.GET("/roles", r.authHandler.GetRolesGin)

		// Basic system information
		// TODO: implement system info handlers
		// system.GET("/info", r.systemHandler.GetSystemInfo)
		// system.GET("/version", r.systemHandler.GetVersion)
	}

	// Audit routes (both master and admin need audit access)
	audit := r.engine.Group("/api/v1/audit")
	audit.Use(authMiddleware.RequireAuth())
	audit.Use(authMiddleware.RequireAnyRole("admin", "master"))
	{
		// Audit logs (both roles need this for different reasons)
		// Master: Business compliance and oversight
		// Admin: Technical troubleshooting and security
		audit.GET("/logs", r.securityHandler.GetAuditLogs)

		// TODO: implement more specific audit endpoints
		// audit.GET("/security", r.auditHandler.GetSecurityLogs)
		// audit.GET("/business", r.auditHandler.GetBusinessLogs)
		// audit.GET("/technical", r.auditHandler.GetTechnicalLogs)
	}
}
