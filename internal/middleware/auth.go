package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/paulochiaradia/dashtrack/internal/auth"
	"github.com/paulochiaradia/dashtrack/internal/models"
)

type GinAuthMiddleware struct {
	jwtManager auth.JWTManagerInterface
}

func NewGinAuthMiddleware(jwtManager auth.JWTManagerInterface) *GinAuthMiddleware {
	return &GinAuthMiddleware{
		jwtManager: jwtManager,
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

		// Validate token using JWTManager
		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", claims.UserID.String())
		c.Set("email", claims.Email)
		c.Set("name", claims.Name)
		c.Set("role_id", claims.RoleID.String())
		c.Set("role_name", claims.RoleName)
		c.Set("user_role", claims.RoleName) // For compatibility with UserHandler
		if claims.TenantID != nil {
			c.Set("tenant_id", claims.TenantID.String())
			c.Set("company_id", claims.TenantID.String()) // For compatibility with UserHandler
		}

		// Create user context for multitenant middleware
		userContext := &models.UserContext{
			UserID:    claims.UserID,
			CompanyID: claims.TenantID, // TenantID maps to CompanyID
			Role:      claims.RoleName,
			IsMaster:  claims.RoleName == "master",
		}
		c.Set("userContext", userContext)

		c.Next()
	}
} // RequireRole middleware ensures the user has the specified role
func (m *GinAuthMiddleware) RequireRole(role string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role_name")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
			c.Abort()
			return
		}

		if userRole.(string) != role {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyRole middleware ensures the user has at least one of the specified roles
func (m *GinAuthMiddleware) RequireAnyRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role_name")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
			c.Abort()
			return
		}

		userRoleStr := userRole.(string)
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
func (m *GinAuthMiddleware) RequireAdminRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("role_name")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
			c.Abort()
			return
		}

		userRoleStr := userRole.(string)
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
