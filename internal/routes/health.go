package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (r *Router) setupHealthRoutes() {
	// Health check
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"message":   "API is running with HOT RELOAD and GIN! 🚀",
			"version":   r.cfg.AppVersion,
			"database":  "connected",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})

	// Prometheus metrics
	r.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
