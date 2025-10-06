package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/paulochiaradia/dashtrack/internal/logger"
	"github.com/paulochiaradia/dashtrack/internal/models"
)

// AuditService handles audit logging
type AuditService struct {
	db *sqlx.DB
}

// NewAuditService creates a new audit service
func NewAuditService(db *sqlx.DB) *AuditService {
	return &AuditService{
		db: db,
	}
}

// AuditAction represents an audit action
type AuditAction string

const (
	// Authentication actions
	ActionLogin          AuditAction = "LOGIN"
	ActionLoginFailed    AuditAction = "LOGIN_FAILED"
	ActionLogout         AuditAction = "LOGOUT"
	ActionPasswordChange AuditAction = "PASSWORD_CHANGE"
	ActionPasswordReset  AuditAction = "PASSWORD_RESET"
	Action2FAEnabled     AuditAction = "2FA_ENABLED"
	Action2FADisabled    AuditAction = "2FA_DISABLED"
	Action2FAVerified    AuditAction = "2FA_VERIFIED"
	Action2FAFailed      AuditAction = "2FA_FAILED"

	// User management actions
	ActionUserCreated     AuditAction = "USER_CREATED"
	ActionUserUpdated     AuditAction = "USER_UPDATED"
	ActionUserDeleted     AuditAction = "USER_DELETED"
	ActionUserActivated   AuditAction = "USER_ACTIVATED"
	ActionUserDeactivated AuditAction = "USER_DEACTIVATED"

	// Company actions
	ActionCompanyCreated AuditAction = "COMPANY_CREATED"
	ActionCompanyUpdated AuditAction = "COMPANY_UPDATED"
	ActionCompanyDeleted AuditAction = "COMPANY_DELETED"

	// Vehicle actions
	ActionVehicleCreated AuditAction = "VEHICLE_CREATED"
	ActionVehicleUpdated AuditAction = "VEHICLE_UPDATED"
	ActionVehicleDeleted AuditAction = "VEHICLE_DELETED"

	// ESP32 actions
	ActionDeviceRegistered   AuditAction = "DEVICE_REGISTERED"
	ActionDeviceUpdated      AuditAction = "DEVICE_UPDATED"
	ActionDeviceDeleted      AuditAction = "DEVICE_DELETED"
	ActionDeviceStatusChange AuditAction = "DEVICE_STATUS_CHANGE"

	// System actions
	ActionSystemConfigUpdate AuditAction = "SYSTEM_CONFIG_UPDATE"
	ActionBackupCreated      AuditAction = "BACKUP_CREATED"
	ActionSystemRestart      AuditAction = "SYSTEM_RESTART"

	// Security actions
	ActionRateLimitTriggered AuditAction = "RATE_LIMIT_TRIGGERED"
	ActionSuspiciousActivity AuditAction = "SUSPICIOUS_ACTIVITY"
	ActionPermissionDenied   AuditAction = "PERMISSION_DENIED"
)

// LogEntry represents an audit log entry input
type LogEntry struct {
	UserID       *uuid.UUID             `json:"user_id,omitempty"`
	Action       AuditAction            `json:"action"`
	Resource     string                 `json:"resource"`
	ResourceID   *string                `json:"resource_id,omitempty"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Success      bool                   `json:"success"`
	ErrorMessage *string                `json:"error_message,omitempty"`
}

// Log creates an audit log entry
func (as *AuditService) Log(ctx context.Context, entry *LogEntry) error {
	auditLog := &models.AuditLog{
		ID:           uuid.New(),
		UserID:       entry.UserID,
		Action:       string(entry.Action),
		Resource:     entry.Resource,
		ResourceID:   entry.ResourceID,
		IPAddress:    entry.IPAddress,
		UserAgent:    entry.UserAgent,
		Details:      entry.Details,
		Success:      entry.Success,
		ErrorMessage: entry.ErrorMessage,
		CreatedAt:    time.Now(),
	}

	// Store asynchronously to avoid blocking main flow
	go func() {
		err := as.storeAuditLog(context.Background(), auditLog)
		if err != nil {
			logger.Error("Failed to store audit log",
				zap.Error(err),
				zap.String("action", string(entry.Action)),
				zap.String("resource", entry.Resource),
			)
		}
	}()

	return nil
}

// LogAuthentication logs authentication events
func (as *AuditService) LogAuthentication(ctx context.Context, userID *uuid.UUID, action AuditAction, ipAddress, userAgent string, success bool, errorMsg *string, details map[string]interface{}) error {
	return as.Log(ctx, &LogEntry{
		UserID:       userID,
		Action:       action,
		Resource:     "authentication",
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Details:      details,
		Success:      success,
		ErrorMessage: errorMsg,
	})
}

// LogUserAction logs user management actions
func (as *AuditService) LogUserAction(ctx context.Context, actorID *uuid.UUID, action AuditAction, targetUserID string, ipAddress, userAgent string, success bool, errorMsg *string, details map[string]interface{}) error {
	return as.Log(ctx, &LogEntry{
		UserID:       actorID,
		Action:       action,
		Resource:     "users",
		ResourceID:   &targetUserID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Details:      details,
		Success:      success,
		ErrorMessage: errorMsg,
	})
}

// LogCompanyAction logs company management actions
func (as *AuditService) LogCompanyAction(ctx context.Context, userID *uuid.UUID, action AuditAction, companyID string, ipAddress, userAgent string, success bool, errorMsg *string, details map[string]interface{}) error {
	return as.Log(ctx, &LogEntry{
		UserID:       userID,
		Action:       action,
		Resource:     "companies",
		ResourceID:   &companyID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Details:      details,
		Success:      success,
		ErrorMessage: errorMsg,
	})
}

// LogVehicleAction logs vehicle management actions
func (as *AuditService) LogVehicleAction(ctx context.Context, userID *uuid.UUID, action AuditAction, vehicleID string, ipAddress, userAgent string, success bool, errorMsg *string, details map[string]interface{}) error {
	return as.Log(ctx, &LogEntry{
		UserID:       userID,
		Action:       action,
		Resource:     "vehicles",
		ResourceID:   &vehicleID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Details:      details,
		Success:      success,
		ErrorMessage: errorMsg,
	})
}

// LogDeviceAction logs ESP32 device actions
func (as *AuditService) LogDeviceAction(ctx context.Context, userID *uuid.UUID, action AuditAction, deviceID string, ipAddress, userAgent string, success bool, errorMsg *string, details map[string]interface{}) error {
	return as.Log(ctx, &LogEntry{
		UserID:       userID,
		Action:       action,
		Resource:     "devices",
		ResourceID:   &deviceID,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Details:      details,
		Success:      success,
		ErrorMessage: errorMsg,
	})
}

// LogSecurityEvent logs security-related events
func (as *AuditService) LogSecurityEvent(ctx context.Context, userID *uuid.UUID, action AuditAction, ipAddress, userAgent string, details map[string]interface{}) error {
	return as.Log(ctx, &LogEntry{
		UserID:    userID,
		Action:    action,
		Resource:  "security",
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   details,
		Success:   true, // Security events are logged regardless
	})
}

// GetAuditLogs retrieves audit logs with filtering and pagination
func (as *AuditService) GetAuditLogs(ctx context.Context, filters *AuditLogFilters) ([]*models.AuditLog, int, error) {
	// Build query conditions
	conditions := []string{"1=1"}
	args := []interface{}{}
	argIndex := 1

	if filters.UserID != nil {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, *filters.UserID)
		argIndex++
	}

	if filters.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, filters.Action)
		argIndex++
	}

	if filters.Resource != "" {
		conditions = append(conditions, fmt.Sprintf("resource = $%d", argIndex))
		args = append(args, filters.Resource)
		argIndex++
	}

	if filters.IPAddress != "" {
		conditions = append(conditions, fmt.Sprintf("ip_address = $%d", argIndex))
		args = append(args, filters.IPAddress)
		argIndex++
	}

	if !filters.StartDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, filters.StartDate)
		argIndex++
	}

	if !filters.EndDate.IsZero() {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, filters.EndDate)
		argIndex++
	}

	if filters.Success != nil {
		conditions = append(conditions, fmt.Sprintf("success = $%d", argIndex))
		args = append(args, *filters.Success)
		argIndex++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total records
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs WHERE %s", whereClause)
	var total int
	err := as.db.GetContext(ctx, &total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Get records with pagination
	query := fmt.Sprintf(`
		SELECT id, user_id, action, resource, resource_id, ip_address, user_agent,
			   details, success, error_message, created_at
		FROM audit_logs
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, filters.Limit, filters.Offset)

	var logs []*models.AuditLog
	err = as.db.SelectContext(ctx, &logs, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get audit logs: %w", err)
	}

	return logs, total, nil
}

// AuditLogFilters represents filters for audit log queries
type AuditLogFilters struct {
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Action    string     `json:"action,omitempty"`
	Resource  string     `json:"resource,omitempty"`
	IPAddress string     `json:"ip_address,omitempty"`
	StartDate time.Time  `json:"start_date,omitempty"`
	EndDate   time.Time  `json:"end_date,omitempty"`
	Success   *bool      `json:"success,omitempty"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}

// CleanupOldLogs removes audit logs older than specified duration
func (as *AuditService) CleanupOldLogs(ctx context.Context, retentionDays int) error {
	query := `
		DELETE FROM audit_logs
		WHERE created_at < NOW() - INTERVAL '%d days'
	`

	result, err := as.db.ExecContext(ctx, fmt.Sprintf(query, retentionDays))
	if err != nil {
		return fmt.Errorf("failed to cleanup old audit logs: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	logger.Info("Cleaned up old audit logs",
		zap.Int("retention_days", retentionDays),
		zap.Int64("rows_deleted", rowsAffected),
	)

	return nil
}

// storeAuditLog stores an audit log entry in the database
func (as *AuditService) storeAuditLog(ctx context.Context, log *models.AuditLog) error {
	// Convert details to JSON
	var detailsJSON []byte
	var err error
	if log.Details != nil {
		detailsJSON, err = json.Marshal(log.Details)
		if err != nil {
			return fmt.Errorf("failed to marshal details: %w", err)
		}
	}

	query := `
		INSERT INTO audit_logs (
			id, user_id, action, resource, resource_id, ip_address, user_agent,
			details, success, error_message, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	_, err = as.db.ExecContext(ctx, query,
		log.ID, log.UserID, log.Action, log.Resource, log.ResourceID,
		log.IPAddress, log.UserAgent, detailsJSON, log.Success,
		log.ErrorMessage, log.CreatedAt,
	)

	return err
}
