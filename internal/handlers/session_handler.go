package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/logger"
	"github.com/paulochiaradia/dashtrack/internal/services"
	"go.uber.org/zap"
)

// SessionHandler handles session management endpoints
type SessionHandler struct {
	sessionManager *services.SessionManager
}

// NewSessionHandler creates a new session handler
func NewSessionHandler(sessionManager *services.SessionManager) *SessionHandler {
	return &SessionHandler{
		sessionManager: sessionManager,
	}
}

// GetSessionDashboard returns comprehensive session information
func (sh *SessionHandler) GetSessionDashboard(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	dashboard, err := sh.sessionManager.GetUserSessionDashboard(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get session dashboard", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve session information"})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// GetActiveSession returns all active sessions for the user
func (sh *SessionHandler) GetActiveSessions(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	sessions, err := sh.sessionManager.GetActiveSessionsForUser(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get active sessions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve active sessions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"active_sessions": sessions,
		"count":           len(sessions),
	})
}

// RevokeSession allows user to revoke a specific session
func (sh *SessionHandler) RevokeSession(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	sessionIDStr := c.Param("sessionId")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid session ID"})
		return
	}

	// Verify session belongs to user
	sessions, err := sh.sessionManager.GetActiveSessionsForUser(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get active sessions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify session ownership"})
		return
	}

	found := false
	for _, session := range sessions {
		if session.ID == sessionID {
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found or does not belong to user"})
		return
	}

	// Revoke the session
	err = sh.sessionManager.RevokeOldestSessions(c.Request.Context(), []uuid.UUID{sessionID}, "user_requested")
	if err != nil {
		logger.Error("Failed to revoke session", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke session"})
		return
	}

	logger.Info("Session revoked by user",
		zap.String("user_id", userID.String()),
		zap.String("session_id", sessionID.String()))

	c.JSON(http.StatusOK, gin.H{"message": "Session revoked successfully"})
}

// GetSessionMetrics returns detailed session metrics for the user
func (sh *SessionHandler) GetSessionMetrics(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	metrics, err := sh.sessionManager.GetUserSessionMetrics(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to get session metrics", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve session metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetSecurityAlerts returns security alerts for the user
func (sh *SessionHandler) GetSecurityAlerts(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	alerts, err := sh.sessionManager.DetectSuspiciousActivity(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to detect suspicious activity", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve security alerts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"security_alerts": alerts,
		"count":           len(alerts),
		"has_concerns":    len(alerts) > 0,
	})
}
