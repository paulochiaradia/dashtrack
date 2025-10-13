package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/services"
)

type GinAuthMiddleware struct {
	tokenService *services.TokenService
}

func NewGinAuthMiddleware(tokenService *services.TokenService) *GinAuthMiddleware {
	return &GinAuthMiddleware{
		tokenService: tokenService,
	}
}

// RequireAuth middleware ensures the request has a valid JWT token
func (m *GinAuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

		// Validate token using TokenService
		user, err := m.tokenService.ValidateAccessToken(c.Request.Context(), tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", user.ID.String())
		c.Set("email", user.Email)
		c.Set("name", user.Name)
		c.Set("role_id", user.RoleID.String())
		c.Set("role_name", user.Role.Name)
		c.Set("user_role", user.Role.Name) // For compatibility with UserHandler
		if user.CompanyID != nil {
			c.Set("tenant_id", user.CompanyID.String())
			c.Set("company_id", user.CompanyID.String()) // For compatibility with UserHandler
		}

		// Create user context for multitenant middleware
		userContext := &models.UserContext{
			UserID:    user.ID,
			CompanyID: user.CompanyID,
			Role:      user.Role.Name,
			IsMaster:  user.Role.Name == "master",
		}
		c.Set("userContext", userContext)

		c.Next()
	}
} // RequireRole middleware ensures the user has the specified role
// Master role has universal access to all routes
func (m *GinAuthMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role_name")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
			c.Abort()
			return
		}

		userRoleStr := userRole.(string)

		// Master role has universal access
		if userRoleStr == "master" {
			c.Next()
			return
		}

		if userRoleStr != role {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole middleware ensures the user has at least one of the specified roles
// Master role has universal access to all routes
func (m *GinAuthMiddleware) RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role_name")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
			c.Abort()
			return
		}

		userRoleStr := userRole.(string)

		// Master role has universal access
		if userRoleStr == "master" {
			c.Next()
			return
		}

		for _, role := range roles {
			if userRoleStr == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}

// RequireAdminRole middleware ensures the user has admin role (technical/operational admin)
// Master role has universal access to all routes
func (m *GinAuthMiddleware) RequireAdminRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role_name")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
			c.Abort()
			return
		}

		userRoleStr := userRole.(string)

		// Master role has universal access
		if userRoleStr == "master" {
			c.Next()
			return
		}

		if userRoleStr != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin role required for technical operations"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireMasterRole middleware ensures the user has master role (business owner)
func (m *GinAuthMiddleware) RequireMasterRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role_name")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
			c.Abort()
			return
		}

		userRoleStr := userRole.(string)
		if userRoleStr != "master" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Master role required for business operations"})
			c.Abort()
			return
		}

		c.Next()
	}
}
