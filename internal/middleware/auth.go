package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/paulochiaradia/dashtrack/internal/auth"
)

type GinAuthMiddleware struct {
	jwtManager *auth.JWTManager
}

func NewGinAuthMiddleware(jwtManager *auth.JWTManager) *GinAuthMiddleware {
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

		// Validate token
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

		c.Next()
	}
}

// RequireRole middleware ensures the user has the specified role
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
