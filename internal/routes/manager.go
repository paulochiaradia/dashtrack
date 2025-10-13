package routes

func (r *Router) setupManagerRoutes() {
	// Create Gin middleware from auth middleware
	authMiddleware := r.authMiddleware

	// Manager routes (manager and admin)
	manager := r.engine.Group("/api/v1/manager")
	manager.Use(authMiddleware.RequireAuth())
	manager.Use(authMiddleware.RequireAnyRole("manager", "admin"))

	// User management (limited to same store/company)
	manager.GET("/users", r.userHandler.GetUsers)

	// Team management (TODO: implement handlers)
	// manager.GET("/teams", r.teamHandler.GetTeamsGin)
	// manager.POST("/teams", r.teamHandler.CreateTeamGin)
	// manager.PUT("/teams/:id", r.teamHandler.UpdateTeamGin)
	// manager.DELETE("/teams/:id", r.teamHandler.DeleteTeamGin)

	// Team member management
	// manager.POST("/teams/:id/members", r.teamHandler.AddMemberGin)
	// manager.DELETE("/teams/:team_id/members/:user_id", r.teamHandler.RemoveMemberGin)
}
