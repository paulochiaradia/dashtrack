package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/paulochiaradia/dashtrack/internal/logger"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/services"
)

// SecurityHandler handles security-related endpoints
type SecurityHandler struct {
	tokenService     *services.TokenService
	twoFactorService *services.TwoFactorService
	auditService     *services.AuditService
}

// NewSecurityHandler creates a new security handler
func NewSecurityHandler(
	tokenService *services.TokenService,
	twoFactorService *services.TwoFactorService,
	auditService *services.AuditService,
) *SecurityHandler {
	return &SecurityHandler{
		tokenService:     tokenService,
		twoFactorService: twoFactorService,
		auditService:     auditService,
	}
}

// RefreshToken handles token refresh requests
func (sh *SecurityHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("Invalid refresh token request format", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	logger.Info("Refresh token request received",
		zap.String("client_ip", c.ClientIP()),
		zap.String("refresh_token_prefix", req.RefreshToken[:20]+"..."))

	clientIP := c.ClientIP()
	userAgent := c.Request.UserAgent()

	// Refresh token pair
	tokenPair, err := sh.tokenService.RefreshTokenPair(c.Request.Context(), req.RefreshToken, clientIP, userAgent)
	if err != nil {
		// Log failed refresh attempt
		logger.Error("Failed to refresh token",
			zap.Error(err),
			zap.String("client_ip", clientIP),
			zap.String("refresh_token_prefix", req.RefreshToken[:20]+"..."))

		errorMsg := err.Error()
		sh.auditService.LogAuthentication(c.Request.Context(), nil, services.ActionLoginFailed, clientIP, userAgent, false, &errorMsg, map[string]interface{}{
			"reason": "invalid_refresh_token",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	// Log successful token refresh
	sh.auditService.LogAuthentication(c.Request.Context(), nil, services.ActionLogin, clientIP, userAgent, true, nil, map[string]interface{}{
		"method": "refresh_token",
	})

	c.JSON(http.StatusOK, tokenPair)
}

// Logout handles logout requests
func (sh *SecurityHandler) Logout(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Revoke all user sessions
	err = sh.tokenService.RevokeAllUserSessions(c.Request.Context(), userID)
	if err != nil {
		logger.Error("Failed to revoke user sessions", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}

	// Log logout
	sh.auditService.LogAuthentication(c.Request.Context(), &userID, services.ActionLogout, c.ClientIP(), c.Request.UserAgent(), true, nil, nil)

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Setup2FA initiates 2FA setup
func (sh *SecurityHandler) Setup2FA(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get user email for account name
	email, _ := c.Get("email")
	accountName := email.(string)

	// Setup 2FA
	setup, err := sh.twoFactorService.SetupTwoFactor(c.Request.Context(), userID, "DashTrack", accountName)
	if err != nil {
		errorMsg := err.Error()
		sh.auditService.LogAuthentication(c.Request.Context(), &userID, services.Action2FAEnabled, c.ClientIP(), c.Request.UserAgent(), false, &errorMsg, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log 2FA setup attempt
	sh.auditService.LogAuthentication(c.Request.Context(), &userID, services.Action2FAEnabled, c.ClientIP(), c.Request.UserAgent(), true, nil, map[string]interface{}{
		"step": "setup_initiated",
	})

	c.JSON(http.StatusOK, setup)
}

// Enable2FA enables 2FA after verification
func (sh *SecurityHandler) Enable2FA(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.TwoFactorSetupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Enable 2FA
	err = sh.twoFactorService.EnableTwoFactor(c.Request.Context(), userID, req.Code)
	if err != nil {
		errorMsg := err.Error()
		sh.auditService.LogAuthentication(c.Request.Context(), &userID, services.Action2FAEnabled, c.ClientIP(), c.Request.UserAgent(), false, &errorMsg, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log successful 2FA enablement
	sh.auditService.LogAuthentication(c.Request.Context(), &userID, services.Action2FAEnabled, c.ClientIP(), c.Request.UserAgent(), true, nil, nil)

	c.JSON(http.StatusOK, gin.H{"message": "2FA enabled successfully"})
}

// Disable2FA disables 2FA
func (sh *SecurityHandler) Disable2FA(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.TwoFactorVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Disable 2FA
	err = sh.twoFactorService.DisableTwoFactor(c.Request.Context(), userID, req.Code)
	if err != nil {
		errorMsg := err.Error()
		sh.auditService.LogAuthentication(c.Request.Context(), &userID, services.Action2FADisabled, c.ClientIP(), c.Request.UserAgent(), false, &errorMsg, nil)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log 2FA disablement
	sh.auditService.LogAuthentication(c.Request.Context(), &userID, services.Action2FADisabled, c.ClientIP(), c.Request.UserAgent(), true, nil, nil)

	c.JSON(http.StatusOK, gin.H{"message": "2FA disabled successfully"})
}

// Verify2FA verifies a 2FA code during login
func (sh *SecurityHandler) Verify2FA(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.TwoFactorVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Verify 2FA code
	valid, err := sh.twoFactorService.VerifyTwoFactor(c.Request.Context(), userID, req.Code)
	if err != nil {
		errorMsg := err.Error()
		sh.auditService.LogAuthentication(c.Request.Context(), &userID, services.Action2FAFailed, c.ClientIP(), c.Request.UserAgent(), false, &errorMsg, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "2FA verification failed"})
		return
	}

	if !valid {
		sh.auditService.LogAuthentication(c.Request.Context(), &userID, services.Action2FAFailed, c.ClientIP(), c.Request.UserAgent(), false, nil, map[string]interface{}{
			"reason": "invalid_code",
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid 2FA code"})
		return
	}

	// Log successful 2FA verification
	sh.auditService.LogAuthentication(c.Request.Context(), &userID, services.Action2FAVerified, c.ClientIP(), c.Request.UserAgent(), true, nil, nil)

	c.JSON(http.StatusOK, gin.H{"verified": true})
}

// GenerateBackupCodes generates new backup codes
func (sh *SecurityHandler) GenerateBackupCodes(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Generate backup codes
	codes, err := sh.twoFactorService.GenerateBackupCodes(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Log backup codes generation
	sh.auditService.LogAuthentication(c.Request.Context(), &userID, services.Action2FAEnabled, c.ClientIP(), c.Request.UserAgent(), true, nil, map[string]interface{}{
		"action": "backup_codes_generated",
	})

	c.JSON(http.StatusOK, gin.H{
		"backup_codes": codes,
		"message":      "New backup codes generated. Store them safely!",
	})
}

// GetAuditLogs retrieves audit logs with filtering
func (sh *SecurityHandler) GetAuditLogs(c *gin.Context) {
	// Parse query parameters
	filters := &services.AuditLogFilters{
		Limit:  50, // Default limit
		Offset: 0,
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			filters.Offset = (page - 1) * filters.Limit
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 500 {
			filters.Limit = limit
		}
	}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			filters.UserID = &userID
		}
	}

	if action := c.Query("action"); action != "" {
		filters.Action = action
	}

	if resource := c.Query("resource"); resource != "" {
		filters.Resource = resource
	}

	if ipAddress := c.Query("ip_address"); ipAddress != "" {
		filters.IPAddress = ipAddress
	}

	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters.StartDate = startDate
		}
	}

	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filters.EndDate = endDate.Add(24 * time.Hour) // End of day
		}
	}

	if successStr := c.Query("success"); successStr != "" {
		if success, err := strconv.ParseBool(successStr); err == nil {
			filters.Success = &success
		}
	}

	// Get audit logs
	logs, total, err := sh.auditService.GetAuditLogs(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"logs": logs,
			"pagination": gin.H{
				"total":  total,
				"limit":  filters.Limit,
				"offset": filters.Offset,
				"pages":  (total + filters.Limit - 1) / filters.Limit,
			},
		},
	})
}

// Get2FAStatus returns the current 2FA status for the user
func (sh *SecurityHandler) Get2FAStatus(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Check 2FA status
	enabled, err := sh.twoFactorService.IsTwoFactorEnabled(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check 2FA status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"enabled": enabled,
	})
}
