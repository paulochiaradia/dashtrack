package routes

func (r *Router) setupAuthRoutes() {
	auth := r.engine.Group("/auth")
	auth.POST("/login", r.authHandler.LoginGin)
	auth.POST("/refresh", r.authHandler.RefreshTokenGin)
	auth.POST("/logout", r.authHandler.LogoutGin)
	auth.POST("/forgot-password", r.authHandler.ForgotPasswordGin)
	auth.POST("/reset-password", r.authHandler.ResetPasswordGin)
}
