package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/services"
)

// AuditHandler handles audit log HTTP requests
type AuditHandler struct {
	auditService *services.AuditService
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(auditService *services.AuditService) *AuditHandler {
	return &AuditHandler{
		auditService: auditService,
	}
}

// GetLogs handles GET /api/v1/audit/logs
func (h *AuditHandler) GetLogs(c *gin.Context) {
	filter := &models.AuditLogFilter{}

	// Parse query parameters
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id format"})
			return
		}
		filter.UserID = &userID
	}

	if companyIDStr := c.Query("company_id"); companyIDStr != "" {
		companyID, err := uuid.Parse(companyIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company_id format"})
			return
		}
		filter.CompanyID = &companyID
	}

	if action := c.Query("action"); action != "" {
		filter.Action = &action
	}

	if resource := c.Query("resource"); resource != "" {
		filter.Resource = &resource
	}

	if resourceID := c.Query("resource_id"); resourceID != "" {
		filter.ResourceID = &resourceID
	}

	if successStr := c.Query("success"); successStr != "" {
		success := successStr == "true"
		filter.Success = &success
	}

	// Parse date range
	if fromStr := c.Query("from"); fromStr != "" {
		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from date format (use RFC3339)"})
			return
		}
		filter.From = &from
	}

	if toStr := c.Query("to"); toStr != "" {
		to, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to date format (use RFC3339)"})
			return
		}
		filter.To = &to
	}

	// Parse pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil || limit < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit"})
			return
		}
		filter.Limit = limit
	} else {
		filter.Limit = 50 // Default
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offset"})
			return
		}
		filter.Offset = offset
	}

	// Get logs
	logs, total, err := h.auditService.GetLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetLogByID handles GET /api/v1/audit/logs/:id
func (h *AuditHandler) GetLogByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid log ID"})
		return
	}

	log, err := h.auditService.GetLogByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit log"})
		return
	}

	if log == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Audit log not found"})
		return
	}

	c.JSON(http.StatusOK, log)
}

// GetStats handles GET /api/v1/audit/stats
func (h *AuditHandler) GetStats(c *gin.Context) {
	filter := &models.AuditLogFilter{}

	// Parse date range
	if fromStr := c.Query("from"); fromStr != "" {
		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from date format (use RFC3339)"})
			return
		}
		filter.From = &from
	}

	if toStr := c.Query("to"); toStr != "" {
		to, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to date format (use RFC3339)"})
			return
		}
		filter.To = &to
	}

	stats, err := h.auditService.GetStats(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit statistics"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTimeline handles GET /api/v1/audit/timeline
func (h *AuditHandler) GetTimeline(c *gin.Context) {
	filter := &models.AuditLogFilter{
		Limit: 100, // Timeline limited to 100 most recent
	}

	// Parse date range
	if fromStr := c.Query("from"); fromStr != "" {
		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid from date format"})
			return
		}
		filter.From = &from
	}

	if toStr := c.Query("to"); toStr != "" {
		to, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid to date format"})
			return
		}
		filter.To = &to
	}

	// Optional filters
	if resource := c.Query("resource"); resource != "" {
		filter.Resource = &resource
	}

	logs, _, err := h.auditService.GetLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve timeline"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"timeline": logs,
		"count":    len(logs),
	})
}

// GetUserLogs handles GET /api/v1/audit/users/:id/logs
func (h *AuditHandler) GetUserLogs(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	filter := &models.AuditLogFilter{
		UserID: &userID,
		Limit:  50,
	}

	// Parse pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	logs, total, err := h.auditService.GetLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetResourceLogs handles GET /api/v1/audit/resources/:type
func (h *AuditHandler) GetResourceLogs(c *gin.Context) {
	resourceType := c.Param("type")

	filter := &models.AuditLogFilter{
		Resource: &resourceType,
		Limit:    50,
	}

	// Parse pagination
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	logs, total, err := h.auditService.GetLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve resource logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":     logs,
		"total":    total,
		"resource": resourceType,
		"limit":    filter.Limit,
		"offset":   filter.Offset,
	})
}

// GetByTraceID handles GET /api/v1/audit/traces/:traceId
func (h *AuditHandler) GetByTraceID(c *gin.Context) {
	traceID := c.Param("traceId")

	logs, err := h.auditService.GetByTraceID(c.Request.Context(), traceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve logs by trace ID"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":     logs,
		"trace_id": traceID,
		"count":    len(logs),
	})
}

// ExportLogs handles GET /api/v1/audit/export
func (h *AuditHandler) ExportLogs(c *gin.Context) {
	format := c.Query("format")
	if format == "" {
		format = "json"
	}

	if format != "json" && format != "csv" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid format. Use 'json' or 'csv'"})
		return
	}

	filter := &models.AuditLogFilter{}

	// Parse filters (same as GetLogs)
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err == nil {
			filter.UserID = &userID
		}
	}

	if action := c.Query("action"); action != "" {
		filter.Action = &action
	}

	if resource := c.Query("resource"); resource != "" {
		filter.Resource = &resource
	}

	if fromStr := c.Query("from"); fromStr != "" {
		from, err := time.Parse(time.RFC3339, fromStr)
		if err == nil {
			filter.From = &from
		}
	}

	if toStr := c.Query("to"); toStr != "" {
		to, err := time.Parse(time.RFC3339, toStr)
		if err == nil {
			filter.To = &to
		}
	}

	data, err := h.auditService.ExportLogs(c.Request.Context(), filter, format)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to export logs"})
		return
	}

	// Set appropriate headers
	filename := "audit_logs_" + time.Now().Format("20060102_150405")
	if format == "json" {
		c.Header("Content-Disposition", "attachment; filename="+filename+".json")
		c.Data(http.StatusOK, "application/json", data)
	} else {
		c.Header("Content-Disposition", "attachment; filename="+filename+".csv")
		c.Data(http.StatusOK, "text/csv", data)
	}
}
