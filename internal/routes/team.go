package routes

func (r *Router) setupTeamRoutes() {
	authMiddleware := r.authMiddleware

	// ==================================================
	// COMPANY ADMIN ROUTES - Full Team Management
	// ==================================================
	companyAdmin := r.engine.Group("/api/v1/company-admin/teams")
	companyAdmin.Use(authMiddleware.RequireAuth())
	companyAdmin.Use(authMiddleware.RequireRole("company_admin"))

	// CRUD Operations
	companyAdmin.GET("", r.teamHandler.GetTeams)          // List all teams
	companyAdmin.POST("", r.teamHandler.CreateTeam)       // Create team
	companyAdmin.GET("/:id", r.teamHandler.GetTeam)       // Get team details
	companyAdmin.PUT("/:id", r.teamHandler.UpdateTeam)    // Update team
	companyAdmin.DELETE("/:id", r.teamHandler.DeleteTeam) // Delete team

	// Member Management
	companyAdmin.GET("/:id/members", r.teamHandler.GetMembers)                    // List team members
	companyAdmin.POST("/:id/members", r.teamHandler.AddMember)                    // Add member to team
	companyAdmin.DELETE("/:id/members/:userId", r.teamHandler.RemoveMember)       // Remove member from team
	companyAdmin.PUT("/:id/members/:userId/role", r.teamHandler.UpdateMemberRole) // Update member role

	// Statistics & Analytics
	companyAdmin.GET("/:id/stats", r.teamHandler.GetTeamStats)       // Team statistics
	companyAdmin.GET("/:id/vehicles", r.teamHandler.GetTeamVehicles) // Vehicles assigned to team

	// Vehicle Assignment
	companyAdmin.POST("/:id/vehicles/:vehicleId", r.teamHandler.AssignVehicleToTeam)       // Assign vehicle to team
	companyAdmin.DELETE("/:id/vehicles/:vehicleId", r.teamHandler.UnassignVehicleFromTeam) // Unassign vehicle from team

	// ==================================================
	// ADMIN ROUTES - Team Management within Company
	// ==================================================
	admin := r.engine.Group("/api/v1/admin/teams")
	admin.Use(authMiddleware.RequireAuth())
	admin.Use(authMiddleware.RequireRole("admin"))

	// Admins can view and manage teams
	admin.GET("", r.teamHandler.GetTeams)               // List teams
	admin.GET("/:id", r.teamHandler.GetTeam)            // Get team details
	admin.GET("/:id/members", r.teamHandler.GetMembers) // List team members
	admin.GET("/:id/stats", r.teamHandler.GetTeamStats) // Team statistics

	// ==================================================
	// MANAGER ROUTES - View Teams and Members
	// ==================================================
	manager := r.engine.Group("/api/v1/manager/teams")
	manager.Use(authMiddleware.RequireAuth())
	manager.Use(authMiddleware.RequireRole("manager"))

	// Managers can view teams (read-only)
	manager.GET("", r.teamHandler.GetTeams)               // List teams
	manager.GET("/:id", r.teamHandler.GetTeam)            // Get team details
	manager.GET("/:id/members", r.teamHandler.GetMembers) // List team members

	// ==================================================
	// USER ROUTES - View Own Teams
	// ==================================================
	user := r.engine.Group("/api/v1/teams")
	user.Use(authMiddleware.RequireAuth())

	// Any authenticated user can view teams they belong to
	user.GET("/my-teams", r.teamHandler.GetMyTeams) // Get current user's teams
}
