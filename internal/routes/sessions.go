package routes

// setupSessionRoutes sets up session management routes
func (r *Router) setupSessionRoutes() {
	// Create auth middleware
	authMiddleware := r.authMiddleware

	// Session management routes (protected)
	sessions := r.engine.Group("/api/v1/sessions")
	sessions.Use(authMiddleware.RequireAuth())
	{
		sessions.GET("/dashboard", r.sessionHandler.GetSessionDashboard)
		sessions.GET("/active", r.sessionHandler.GetActiveSessions)
		sessions.DELETE("/:sessionId", r.sessionHandler.RevokeSession)
		sessions.GET("/metrics", r.sessionHandler.GetSessionMetrics)
		sessions.GET("/security-alerts", r.sessionHandler.GetSecurityAlerts)
	}
}
