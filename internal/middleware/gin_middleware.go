package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/paulochiaradia/dashtrack/internal/logger"
	"github.com/paulochiaradia/dashtrack/internal/metrics"
	"github.com/paulochiaradia/dashtrack/internal/tracing"
	"go.uber.org/zap"
)

// GinLoggingMiddleware provides structured logging for Gin
func GinLoggingMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request with structured data
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.String("client_ip", clientIP),
			zap.Int("status_code", statusCode),
			zap.Duration("latency", latency),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("body_size", c.Writer.Size()),
		)
	})
}

// GinMetricsMiddleware captures metrics for Gin requests
func GinMetricsMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Record metrics
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		path := c.FullPath() // Use route pattern instead of actual path

		if path == "" {
			path = c.Request.URL.Path // Fallback for unregistered routes
		}

		metrics.HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	})
}

// GinTracingMiddleware adds tracing to Gin requests
func GinTracingMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		ctx, span := tracing.StartSpan(c.Request.Context(), c.Request.Method+" "+c.FullPath())
		defer span.End()

		// Update request context
		c.Request = c.Request.WithContext(ctx)

		// Process request
		c.Next()

		// Add span attributes after processing
		span.SetAttributes(
		// HTTP attributes would go here
		)
	})
}
