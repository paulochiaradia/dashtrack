package middleware

import (
	"bytes"
	"context"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/metrics"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/services"
)

// AuditMiddleware creates a middleware that automatically logs all HTTP requests
func AuditMiddleware(auditService *services.AuditService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip health and metrics endpoints
		if shouldSkipAudit(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Record start time
		start := time.Now()

		// Extract user context (set by auth middleware)
		userID := extractUserID(c)
		companyID := extractCompanyID(c)
		userEmail := extractUserEmail(c)

		// Extract Jaeger tracing context
		traceID := c.GetString("trace_id")
		spanID := c.GetString("span_id")

		// Capture request body for CREATE/UPDATE operations
		var requestBody map[string]interface{}
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			requestBody = captureRequestBody(c)
		}

		// Process the request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Milliseconds()

		// Extract resource information from path
		action := mapMethodToAction(c.Request.Method)
		resource := extractResource(c.Request.URL.Path)
		resourceIDUUID := extractResourceID(c)

		// Convert resourceID UUID to string
		var resourceID *string
		if resourceIDUUID != nil {
			idStr := resourceIDUUID.String()
			resourceID = &idStr
		}

		// Determine success based on status code
		statusCode := c.Writer.Status()
		success := statusCode < 400

		// Get error message if request failed
		var errorMessagePtr *string
		if !success {
			if err, exists := c.Get("error"); exists {
				errMsg := err.(error).Error()
				errorMessagePtr = &errMsg
			}
		}

		// Convert to pointers for model
		method := c.Request.Method
		path := c.Request.URL.Path
		userEmailStr := userEmail
		traceIDStr := traceID
		spanIDStr := spanID

		// Build metadata
		metadata := buildMetadata(c, requestBody)

		// Create audit log entry
		auditLog := &models.AuditLog{
			ID:           uuid.New(), // Generate unique ID for audit log
			UserID:       &userID,
			UserEmail:    &userEmailStr,
			CompanyID:    companyID,
			Action:       action,
			Resource:     resource,
			ResourceID:   resourceID,
			Method:       &method,
			Path:         &path,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Success:      success,
			ErrorMessage: errorMessagePtr,
			StatusCode:   &statusCode,
			DurationMs:   &duration,
			TraceID:      &traceIDStr,
			SpanID:       &spanIDStr,
			Metadata:     metadata,
			CreatedAt:    time.Now(),
		}

		// Log asynchronously (don't block the response)
		go func() {
			ctx := context.Background()
			if err := auditService.LogHTTPRequest(ctx, auditLog); err != nil {
				// Log error but don't fail the request
				// The error is already logged inside the service
			}
		}()

		// Increment Prometheus metrics
		incrementAuditMetrics(action, resource, userEmail, success, duration, c.Request.Method, statusCode, c.Writer.Size())
	}
}

// shouldSkipAudit checks if the path should be excluded from audit logs
func shouldSkipAudit(path string) bool {
	skipPaths := []string{
		"/health",
		"/metrics",
		"/favicon.ico",
	}

	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}

	return false
}

// extractUserID extracts user ID from context (set by auth middleware)
func extractUserID(c *gin.Context) uuid.UUID {
	userIDStr := c.GetString("user_id")
	if userIDStr == "" {
		return uuid.Nil
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil
	}

	return userID
}

// extractCompanyID extracts company ID from context
func extractCompanyID(c *gin.Context) *uuid.UUID {
	companyIDStr := c.GetString("company_id")
	if companyIDStr == "" {
		return nil
	}

	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		return nil
	}

	return &companyID
}

// extractUserEmail extracts user email from context
func extractUserEmail(c *gin.Context) string {
	email := c.GetString("email")
	if email == "" {
		return "anonymous"
	}
	return email
}

// mapMethodToAction maps HTTP method to audit action
func mapMethodToAction(method string) string {
	actionMap := map[string]string{
		"GET":    "READ",
		"POST":   "CREATE",
		"PUT":    "UPDATE",
		"PATCH":  "UPDATE",
		"DELETE": "DELETE",
	}

	if action, exists := actionMap[method]; exists {
		return action
	}

	return "UNKNOWN"
}

// extractResource extracts resource name from URL path
func extractResource(path string) string {
	// Remove /api/v1 prefix
	path = strings.TrimPrefix(path, "/api/v1/")

	// Split by /
	parts := strings.Split(path, "/")

	if len(parts) == 0 {
		return "unknown"
	}

	// Get first segment as resource
	resource := parts[0]

	// Handle special cases
	switch resource {
	case "auth":
		if len(parts) > 1 {
			return "auth_" + parts[1] // auth_login, auth_logout, etc
		}
		return "auth"
	case "profile":
		return "user_profile"
	case "master", "admin", "manager", "company-admin":
		// For role-based routes, get the actual resource
		if len(parts) > 1 {
			return parts[1] // users, companies, etc
		}
		return resource
	default:
		return resource
	}
}

// extractResourceID extracts resource ID from URL parameters
func extractResourceID(c *gin.Context) *uuid.UUID {
	// Try common ID parameter names
	idParams := []string{"id", "userId", "companyId", "vehicleId", "teamId"}

	for _, param := range idParams {
		if idStr := c.Param(param); idStr != "" {
			if id, err := uuid.Parse(idStr); err == nil {
				return &id
			}
		}
	}

	return nil
}

// captureRequestBody captures and parses request body
func captureRequestBody(c *gin.Context) map[string]interface{} {
	// Read body
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil
	}

	// Restore body for next handlers
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Try to parse as JSON
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		return nil
	}

	// Restore body again
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Remove sensitive fields
	sanitizeBody(body)

	return body
}

// sanitizeBody removes sensitive information from request body
func sanitizeBody(body map[string]interface{}) {
	sensitiveFields := []string{
		"password",
		"new_password",
		"old_password",
		"current_password",
		"token",
		"secret",
		"api_key",
		"credit_card",
		"ssn",
		"cpf",
	}

	for _, field := range sensitiveFields {
		if _, exists := body[field]; exists {
			body[field] = "***REDACTED***"
		}
	}
}

// buildMetadata builds metadata object with additional context
func buildMetadata(c *gin.Context, requestBody map[string]interface{}) map[string]interface{} {
	metadata := make(map[string]interface{})

	// Add query parameters
	if len(c.Request.URL.RawQuery) > 0 {
		metadata["query_params"] = c.Request.URL.RawQuery
	}

	// Add request body (sanitized)
	if requestBody != nil {
		metadata["request_body"] = requestBody
	}

	// Add user role if available
	if role := c.GetString("role"); role != "" {
		metadata["user_role"] = role
	}

	// Add referer if available
	if referer := c.Request.Referer(); referer != "" {
		metadata["referer"] = referer
	}

	// Add response size
	metadata["response_size"] = c.Writer.Size()

	return metadata
}

// incrementAuditMetrics increments Prometheus metrics for audit logs
func incrementAuditMetrics(action, resource, userEmail string, success bool, duration int64, method string, statusCode, responseSize int) {
	// Get role from context (defaulting to "unknown" if not set)
	role := "unknown"

	// Increment main action counter
	metrics.IncrementAuditAction(action, resource, role, success)

	// Observe action duration
	metrics.ObserveAuditActionDuration(action, resource, method, duration)

	// Increment user action counter
	metrics.IncrementUserAction(userEmail, action, resource)

	// Increment resource access counter
	metrics.IncrementResourceAccess(resource, action, method)

	// Track HTTP status codes
	metrics.IncrementHTTPStatusCode(method, resource, statusCode)

	// Observe response size
	metrics.ObserveResponseSize(method, resource, statusCode, responseSize)

	// Increment error counter if request failed
	if !success {
		errorType := "http_error"
		if statusCode == 401 {
			errorType = "unauthorized"
		} else if statusCode == 403 {
			errorType = "forbidden"
		} else if statusCode == 404 {
			errorType = "not_found"
		} else if statusCode == 429 {
			errorType = "rate_limit"
		} else if statusCode >= 500 {
			errorType = "server_error"
		}

		metrics.IncrementAuditError(action, resource, errorType, statusCode)
	}

	// Detect suspicious activity
	metrics.DetectSuspiciousActivity(action, resource, statusCode, userEmail)

	// Track authentication events separately
	if resource == "auth_login" || resource == "auth_logout" {
		metrics.IncrementAuthenticationEvent(action, success, "0.0.0.0") // IP will be added later
	}
}
