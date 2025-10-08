package middleware

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/paulochiaradia/dashtrack/internal/logger"
	"github.com/paulochiaradia/dashtrack/internal/models"
)

// RateLimiter represents a rate limiter with database backing
type RateLimiter struct {
	db       *sqlx.DB
	cache    map[string]*RateLimitCache
	mutex    sync.RWMutex
	rules    []models.RateLimitRule
	lastSync time.Time
}

// RateLimitCache represents in-memory cache for rate limiting
type RateLimitCache struct {
	Count     int
	ResetTime time.Time
	Blocked   bool
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(db *sqlx.DB) *RateLimiter {
	rl := &RateLimiter{
		db:    db,
		cache: make(map[string]*RateLimitCache),
	}

	// Load rules from database
	rl.syncRules()

	// Start background sync
	go rl.backgroundSync()

	return rl
}

// parseInterval converts PostgreSQL interval string to time.Duration
func parseInterval(interval string) time.Duration {
	// Handle common PostgreSQL interval formats
	// Examples: "00:01:00", "01:00:00", "00:05:00"
	if strings.Contains(interval, ":") {
		parts := strings.Split(interval, ":")
		if len(parts) == 3 {
			hours, _ := strconv.Atoi(parts[0])
			minutes, _ := strconv.Atoi(parts[1])
			seconds, _ := strconv.Atoi(parts[2])
			return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
		}
	}

	// Default to 1 minute if parsing fails
	return time.Minute
}

// RateLimitMiddleware creates a rate limiting middleware
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		userID := rl.getUserID(c)
		path := c.Request.URL.Path
		method := c.Request.Method

		// Find applicable rule
		rule := rl.findRule(path, method)
		if rule == nil {
			c.Next()
			return
		}

		// Generate cache key
		var key string
		if rule.UserBased && userID != nil {
			key = fmt.Sprintf("user:%s:%s:%s", userID.String(), path, method)
		} else {
			key = fmt.Sprintf("ip:%s:%s:%s", clientIP, path, method)
		}

		// Check rate limit
		blocked, err := rl.checkRateLimit(key, rule, clientIP, userID)
		if err != nil {
			logger.Error("Rate limit check failed", zap.Error(err))
			c.Next()
			return
		}

		if blocked {
			// Log rate limit event
			rl.logRateLimitEvent(userID, clientIP, path, method, true, rule.ID, c.Request.UserAgent())

			windowSize := parseInterval(rule.WindowSize)
			c.Header("X-RateLimit-Limit", strconv.Itoa(rule.MaxRequests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", strconv.Itoa(int(windowSize.Seconds())))

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"retry_after": int(windowSize.Seconds()),
			})
			c.Abort()
			return
		}

		// Log successful request
		rl.logRateLimitEvent(userID, clientIP, path, method, false, rule.ID, c.Request.UserAgent())

		c.Next()
	}
}

// checkRateLimit checks if the request should be blocked
func (rl *RateLimiter) checkRateLimit(key string, rule *models.RateLimitRule, clientIP string, userID *uuid.UUID) (bool, error) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowSize := parseInterval(rule.WindowSize)

	// Get or create cache entry
	cache, exists := rl.cache[key]
	if !exists {
		cache = &RateLimitCache{
			Count:     0,
			ResetTime: now.Add(windowSize),
		}
		rl.cache[key] = cache
	}

	// Reset if window expired
	if now.After(cache.ResetTime) {
		cache.Count = 0
		cache.ResetTime = now.Add(windowSize)
		cache.Blocked = false
	}

	// Increment counter
	cache.Count++

	// Check if limit exceeded
	if cache.Count > rule.MaxRequests {
		cache.Blocked = true
		return true, nil
	}

	return false, nil
}

// findRule finds the applicable rate limit rule for the request
func (rl *RateLimiter) findRule(path, method string) *models.RateLimitRule {
	for _, rule := range rl.rules {
		if !rule.Active {
			continue
		}

		// Simple pattern matching - can be enhanced with regex
		if rule.Method == "ANY" || rule.Method == method {
			if rl.pathMatches(path, rule.Path) {
				return &rule
			}
		}
	}
	return nil
}

// pathMatches checks if path matches the rule pattern
func (rl *RateLimiter) pathMatches(path, pattern string) bool {
	// Simple wildcard matching - can be enhanced
	if pattern == "*" {
		return true
	}

	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		prefix := pattern[:len(pattern)-1]
		return len(path) >= len(prefix) && path[:len(prefix)] == prefix
	}

	return path == pattern
}

// getUserID extracts user ID from context
func (rl *RateLimiter) getUserID(c *gin.Context) *uuid.UUID {
	if userIDStr, exists := c.Get("user_id"); exists {
		if userID, err := uuid.Parse(userIDStr.(string)); err == nil {
			return &userID
		}
	}
	return nil
}

// logRateLimitEvent logs a rate limit event to the database
func (rl *RateLimiter) logRateLimitEvent(userID *uuid.UUID, clientIP, path, method string, blocked bool, ruleID uuid.UUID, userAgent string) {
	go func() {
		query := `
			INSERT INTO rate_limit_events (user_id, ip_address, path, method, user_agent, blocked, rule_id, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
		`

		_, err := rl.db.Exec(query, userID, clientIP, path, method, userAgent, blocked, ruleID)
		if err != nil {
			logger.Error("Failed to log rate limit event", zap.Error(err))
		}
	}()
}

// syncRules loads rate limit rules from database
func (rl *RateLimiter) syncRules() {
	query := `SELECT id, name, path, method, max_requests, window_size, user_based, active, created_at, updated_at FROM rate_limit_rules WHERE active = true`

	err := rl.db.Select(&rl.rules, query)
	if err != nil {
		logger.Error("Failed to sync rate limit rules", zap.Error(err))
		return
	}

	rl.lastSync = time.Now()
	logger.Info("Rate limit rules synced", zap.Int("count", len(rl.rules)))
}

// backgroundSync periodically syncs rules from database
func (rl *RateLimiter) backgroundSync() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.syncRules()
		rl.cleanupCache()
	}
}

// cleanupCache removes expired cache entries
func (rl *RateLimiter) cleanupCache() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	for key, cache := range rl.cache {
		if now.After(cache.ResetTime.Add(time.Hour)) { // Keep cache for 1 hour after reset
			delete(rl.cache, key)
		}
	}
}
