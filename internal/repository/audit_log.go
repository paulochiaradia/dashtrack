package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/paulochiaradia/dashtrack/internal/models"
)

// AuditLogRepositoryInterface defines the contract for audit log repository
type AuditLogRepositoryInterface interface {
	Create(ctx context.Context, log *models.AuditLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error)
	List(ctx context.Context, filter *models.AuditLogFilter) ([]*models.AuditLog, error)
	Count(ctx context.Context, filter *models.AuditLogFilter) (int64, error)
	GetStats(ctx context.Context, filter *models.AuditLogFilter) (*models.AuditLogStats, error)
	GetByTraceID(ctx context.Context, traceID string) ([]*models.AuditLog, error)
	DeleteOldLogs(ctx context.Context, olderThan time.Time) (int64, error)
}

// AuditLogRepository handles audit log database operations
type AuditLogRepository struct {
	db *sqlx.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *sqlx.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create inserts a new audit log entry
func (r *AuditLogRepository) Create(ctx context.Context, log *models.AuditLog) error {
	query := `
		INSERT INTO audit_logs (
			id, user_id, user_email, company_id, action, resource, resource_id,
			method, path, ip_address, user_agent, changes, metadata,
			success, error_message, status_code, duration_ms, trace_id, span_id, created_at
		) VALUES (
			:id, :user_id, :user_email, :company_id, :action, :resource, :resource_id,
			:method, :path, :ip_address, :user_agent, :changes, :metadata,
			:success, :error_message, :status_code, :duration_ms, :trace_id, :span_id, :created_at
		)`

	// Convert maps to JSON
	changesJSON, err := json.Marshal(log.Changes)
	if err != nil {
		return fmt.Errorf("failed to marshal changes: %w", err)
	}

	metadataJSON, err := json.Marshal(log.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Prepare data for insertion
	data := map[string]interface{}{
		"id":            log.ID,
		"user_id":       log.UserID,
		"user_email":    log.UserEmail,
		"company_id":    log.CompanyID,
		"action":        log.Action,
		"resource":      log.Resource,
		"resource_id":   log.ResourceID,
		"method":        log.Method,
		"path":          log.Path,
		"ip_address":    log.IPAddress,
		"user_agent":    log.UserAgent,
		"changes":       changesJSON,
		"metadata":      metadataJSON,
		"success":       log.Success,
		"error_message": log.ErrorMessage,
		"status_code":   log.StatusCode,
		"duration_ms":   log.DurationMs,
		"trace_id":      log.TraceID,
		"span_id":       log.SpanID,
		"created_at":    log.CreatedAt,
	}

	_, err = r.db.NamedExecContext(ctx, query, data)
	return err
}

// GetByID retrieves an audit log by ID
func (r *AuditLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error){
	query := `
		SELECT 
			id, user_id, user_email, company_id, action, resource, resource_id,
			method, path, ip_address, user_agent, changes, metadata,
			success, error_message, status_code, duration_ms, trace_id, span_id, created_at
		FROM audit_logs
		WHERE id = $1`

	var log models.AuditLog
	var changesJSON, metadataJSON []byte

	err := r.db.QueryRowxContext(ctx, query, id).Scan(
		&log.ID, &log.UserID, &log.UserEmail, &log.CompanyID, &log.Action, &log.Resource, &log.ResourceID,
		&log.Method, &log.Path, &log.IPAddress, &log.UserAgent, &changesJSON, &metadataJSON,
		&log.Success, &log.ErrorMessage, &log.StatusCode, &log.DurationMs, &log.TraceID, &log.SpanID, &log.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	// Unmarshal JSON fields
	if changesJSON != nil {
		if err := json.Unmarshal(changesJSON, &log.Changes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal changes: %w", err)
		}
	}

	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &log.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &log, nil
}

// List retrieves audit logs with filters
func (r *AuditLogRepository) List(ctx context.Context, filter *models.AuditLogFilter) ([]*models.AuditLog, error) {
	query := `
		SELECT 
			id, user_id, user_email, company_id, action, resource, resource_id,
			method, path, ip_address, user_agent, changes, metadata,
			success, error_message, status_code, duration_ms, trace_id, span_id, created_at
		FROM audit_logs
		WHERE 1=1`

	args := []interface{}{}
	argCount := 1

	// Apply filters
	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *filter.UserID)
		argCount++
	}

	if filter.CompanyID != nil {
		query += fmt.Sprintf(" AND company_id = $%d", argCount)
		args = append(args, *filter.CompanyID)
		argCount++
	}

	if filter.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argCount)
		args = append(args, *filter.Action)
		argCount++
	}

	if filter.Resource != nil {
		query += fmt.Sprintf(" AND resource = $%d", argCount)
		args = append(args, *filter.Resource)
		argCount++
	}

	if filter.ResourceID != nil {
		query += fmt.Sprintf(" AND resource_id = $%d", argCount)
		args = append(args, *filter.ResourceID)
		argCount++
	}

	if filter.Success != nil {
		query += fmt.Sprintf(" AND success = $%d", argCount)
		args = append(args, *filter.Success)
		argCount++
	}

	if filter.From != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *filter.From)
		argCount++
	}

	if filter.To != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *filter.To)
		argCount++
	}

	// Order by created_at desc
	query += " ORDER BY created_at DESC"

	// Apply limit and offset
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filter.Limit)
		argCount++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryxContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		var changesJSON, metadataJSON []byte

		err := rows.Scan(
			&log.ID, &log.UserID, &log.UserEmail, &log.CompanyID, &log.Action, &log.Resource, &log.ResourceID,
			&log.Method, &log.Path, &log.IPAddress, &log.UserAgent, &changesJSON, &metadataJSON,
			&log.Success, &log.ErrorMessage, &log.StatusCode, &log.DurationMs, &log.TraceID, &log.SpanID, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Unmarshal JSON fields
		if changesJSON != nil {
			json.Unmarshal(changesJSON, &log.Changes)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &log.Metadata)
		}

		logs = append(logs, &log)
	}

	return logs, rows.Err()
}

// Count returns the total count of audit logs matching the filter
func (r *AuditLogRepository) Count(ctx context.Context, filter *models.AuditLogFilter) (int64, error) {
	query := "SELECT COUNT(*) FROM audit_logs WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	// Apply same filters as List
	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argCount)
		args = append(args, *filter.UserID)
		argCount++
	}

	if filter.CompanyID != nil {
		query += fmt.Sprintf(" AND company_id = $%d", argCount)
		args = append(args, *filter.CompanyID)
		argCount++
	}

	if filter.Action != nil {
		query += fmt.Sprintf(" AND action = $%d", argCount)
		args = append(args, *filter.Action)
		argCount++
	}

	if filter.Resource != nil {
		query += fmt.Sprintf(" AND resource = $%d", argCount)
		args = append(args, *filter.Resource)
		argCount++
	}

	if filter.Success != nil {
		query += fmt.Sprintf(" AND success = $%d", argCount)
		args = append(args, *filter.Success)
		argCount++
	}

	if filter.From != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *filter.From)
		argCount++
	}

	if filter.To != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *filter.To)
	}

	var count int64
	err := r.db.GetContext(ctx, &count, query, args...)
	return count, err
}

// GetStats returns aggregated statistics for audit logs
func (r *AuditLogRepository) GetStats(ctx context.Context, filter *models.AuditLogFilter) (*models.AuditLogStats, error) {
	// Base stats query
	var stats models.AuditLogStats

	// Total actions and success rate
	query := "SELECT COUNT(*) as total, AVG(CASE WHEN success THEN 1.0 ELSE 0.0 END) as success_rate, AVG(duration_ms) as avg_duration FROM audit_logs WHERE 1=1"
	args := []interface{}{}
	argCount := 1

	// Apply time filters
	if filter.From != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argCount)
		args = append(args, *filter.From)
		argCount++
	}

	if filter.To != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argCount)
		args = append(args, *filter.To)
	}

	var total int64
	var successRate, avgDuration sql.NullFloat64
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&total, &successRate, &avgDuration)
	if err != nil {
		return nil, err
	}

	stats.TotalActions = total
	if successRate.Valid {
		stats.SuccessRate = successRate.Float64
	}
	if avgDuration.Valid {
		stats.AvgDurationMs = avgDuration.Float64
	}

	// Actions by type
	stats.ActionsByType = make(map[string]int64)
	actionQuery := "SELECT action, COUNT(*) FROM audit_logs"
	if filter.From != nil || filter.To != nil {
		actionQuery += " WHERE 1=1"
		if filter.From != nil {
			actionQuery += " AND created_at >= $1"
		}
		if filter.To != nil {
			actionQuery += " AND created_at <= $2"
		}
	}
	actionQuery += " GROUP BY action"

	rows, err := r.db.QueryContext(ctx, actionQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var action string
		var count int64
		if err := rows.Scan(&action, &count); err != nil {
			return nil, err
		}
		stats.ActionsByType[action] = count
	}

	return &stats, nil
}

// GetByTraceID retrieves all audit logs for a specific Jaeger trace ID
func (r *AuditLogRepository) GetByTraceID(ctx context.Context, traceID string) ([]*models.AuditLog, error) {
	query := `
		SELECT 
			id, user_id, user_email, company_id, action, resource, resource_id,
			method, path, ip_address, user_agent, changes, metadata,
			success, error_message, status_code, duration_ms, trace_id, span_id, created_at
		FROM audit_logs
		WHERE trace_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryxContext(ctx, query, traceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		var log models.AuditLog
		var changesJSON, metadataJSON []byte

		err := rows.Scan(
			&log.ID, &log.UserID, &log.UserEmail, &log.CompanyID, &log.Action, &log.Resource, &log.ResourceID,
			&log.Method, &log.Path, &log.IPAddress, &log.UserAgent, &changesJSON, &metadataJSON,
			&log.Success, &log.ErrorMessage, &log.StatusCode, &log.DurationMs, &log.TraceID, &log.SpanID, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if changesJSON != nil {
			json.Unmarshal(changesJSON, &log.Changes)
		}
		if metadataJSON != nil {
			json.Unmarshal(metadataJSON, &log.Metadata)
		}

		logs = append(logs, &log)
	}

	return logs, rows.Err()
}

// DeleteOldLogs deletes audit logs older than the specified date
func (r *AuditLogRepository) DeleteOldLogs(ctx context.Context, olderThan time.Time) (int64, error) {
	query := "DELETE FROM audit_logs WHERE created_at < $1"
	result, err := r.db.ExecContext(ctx, query, olderThan)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	return rowsAffected, err
}
