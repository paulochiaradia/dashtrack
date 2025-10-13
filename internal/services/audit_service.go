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
	"github.com/paulochiaradia/dashtrack/internal/metrics"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
)

// AuditService handles audit logging
type AuditService struct {
	db   *sqlx.DB
	repo repository.AuditLogRepositoryInterface
}

// NewAuditService creates a new audit service
func NewAuditService(db *sqlx.DB) *AuditService {
	return &AuditService{
		db:   db,
		repo: repository.NewAuditLogRepository(db),
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

// LogHTTPRequest logs HTTP requests automatically (used by audit middleware)
func (as *AuditService) LogHTTPRequest(ctx context.Context, auditLog *models.AuditLog) error {
	// Store asynchronously to avoid blocking main flow
	go func() {
		err := as.storeAuditLog(context.Background(), auditLog)
		if err != nil {
			method := "UNKNOWN"
			if auditLog.Method != nil {
				method = *auditLog.Method
			}
			path := "UNKNOWN"
			if auditLog.Path != nil {
				path = *auditLog.Path
			}

			logger.Error("Failed to store HTTP audit log",
				zap.Error(err),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("action", auditLog.Action),
			)

			// Increment error metric
			metrics.IncrementDatabaseWriteError()
		} else {
			// Increment successful write metric
			metrics.IncrementDatabaseWrite()
		}
	}()

	return nil
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
	return as.repo.Create(ctx, log)
}

// GetLogs retrieves audit logs with filters
func (as *AuditService) GetLogs(ctx context.Context, filter *models.AuditLogFilter) ([]*models.AuditLog, int64, error) {
	// Set default limit if not specified
	if filter.Limit == 0 {
		filter.Limit = 50
	}

	// Get logs
	logs, err := as.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	total, err := as.repo.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

// GetLogByID retrieves a specific audit log
func (as *AuditService) GetLogByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error) {
	return as.repo.GetByID(ctx, id)
}

// GetStats retrieves audit log statistics
func (as *AuditService) GetStats(ctx context.Context, filter *models.AuditLogFilter) (*models.AuditLogStats, error) {
	return as.repo.GetStats(ctx, filter)
}

// GetByTraceID retrieves all logs for a Jaeger trace
func (as *AuditService) GetByTraceID(ctx context.Context, traceID string) ([]*models.AuditLog, error) {
	return as.repo.GetByTraceID(ctx, traceID)
}

// ExportLogs exports audit logs to JSON or CSV format
func (as *AuditService) ExportLogs(ctx context.Context, filter *models.AuditLogFilter, format string) ([]byte, error) {
	// Get all logs without pagination for export
	filter.Limit = 0
	filter.Offset = 0

	logs, err := as.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	switch format {
	case "json":
		return json.MarshalIndent(logs, "", "  ")
	case "csv":
		return exportToCSV(logs)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// exportToCSV converts audit logs to CSV format
func exportToCSV(logs []*models.AuditLog) ([]byte, error) {
	var csv string

	// Header
	csv += "ID,Timestamp,User ID,User Email,Company ID,Action,Resource,Resource ID,Method,Path,IP Address,Success,Status Code,Duration (ms),Trace ID\n"

	// Rows
	for _, log := range logs {
		userID := ""
		if log.UserID != nil {
			userID = log.UserID.String()
		}

		userEmail := ""
		if log.UserEmail != nil {
			userEmail = *log.UserEmail
		}

		companyID := ""
		if log.CompanyID != nil {
			companyID = log.CompanyID.String()
		}

		resourceID := ""
		if log.ResourceID != nil {
			resourceID = *log.ResourceID
		}

		method := ""
		if log.Method != nil {
			method = *log.Method
		}

		path := ""
		if log.Path != nil {
			path = *log.Path
		}

		statusCode := ""
		if log.StatusCode != nil {
			statusCode = fmt.Sprintf("%d", *log.StatusCode)
		}

		durationMs := ""
		if log.DurationMs != nil {
			durationMs = fmt.Sprintf("%d", *log.DurationMs)
		}

		traceID := ""
		if log.TraceID != nil {
			traceID = *log.TraceID
		}

		csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%t,%s,%s,%s\n",
			log.ID.String(),
			log.CreatedAt.Format(time.RFC3339),
			userID,
			userEmail,
			companyID,
			log.Action,
			log.Resource,
			resourceID,
			method,
			path,
			log.IPAddress,
			log.Success,
			statusCode,
			durationMs,
			traceID,
		)
	}

	return []byte(csv), nil
}
