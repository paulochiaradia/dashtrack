package routes

import (
	"github.com/paulochiaradia/dashtrack/internal/middleware"
)

// setupSecurityRoutes sets up security-related routes
func (r *Router) setupSecurityRoutes() {
	// Create auth middleware
	authMiddleware := middleware.NewGinAuthMiddleware(r.jwtManager)

	// Create rate limiter
	rateLimiter := middleware.NewRateLimiter(r.db)

	// Protected security routes
	security := r.engine.Group("/api/v1/security")
	security.Use(authMiddleware.RequireAuth())
	security.Use(rateLimiter.RateLimitMiddleware())
	{
		// Logout
		security.POST("/logout", r.securityHandler.Logout)

		// 2FA Management
		twoFA := security.Group("/2fa")
		{
			twoFA.GET("/status", r.securityHandler.Get2FAStatus)
			twoFA.POST("/setup", r.securityHandler.Setup2FA)
			twoFA.POST("/enable", r.securityHandler.Enable2FA)
			twoFA.POST("/disable", r.securityHandler.Disable2FA)
			twoFA.POST("/verify", r.securityHandler.Verify2FA)
			twoFA.POST("/backup-codes", r.securityHandler.GenerateBackupCodes)
		}

		// Note: Audit logs moved to /api/v1/audit/* (shared master/admin routes)
		// This allows both master (business oversight) and admin (technical) access
	}
}
