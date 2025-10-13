package routes

import (
	"github.com/gin-gonic/gin"
)

// setupAuditRoutes configures audit log routes (Master and Admin only)
func (router *Router) setupAuditRoutes(api *gin.RouterGroup) {
	audit := api.Group("/audit")
	audit.Use(router.authMiddleware.RequireAuth()) // Require authentication

	// All audit endpoints require master or admin role
	// TODO: Add role-based middleware for master/admin only

	// List audit logs with filters
	audit.GET("/logs", router.auditHandler.GetLogs)

	// Get specific audit log
	audit.GET("/logs/:id", router.auditHandler.GetLogByID)

	// Get audit statistics
	audit.GET("/stats", router.auditHandler.GetStats)

	// Get audit timeline
	audit.GET("/timeline", router.auditHandler.GetTimeline)

	// Get logs for specific user
	audit.GET("/users/:id/logs", router.auditHandler.GetUserLogs)

	// Get logs for specific resource type
	audit.GET("/resources/:type", router.auditHandler.GetResourceLogs)

	// Get logs by Jaeger trace ID
	audit.GET("/traces/:traceId", router.auditHandler.GetByTraceID)

	// Export audit logs (JSON or CSV)
	audit.GET("/export", router.auditHandler.ExportLogs)
}
