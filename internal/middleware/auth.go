package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

		// Try to validate token with both formats
		// First try the new TokenService format
		if m.validateTokenServiceToken(c, tokenString) {
			return // Successfully validated
		}

		// Fallback to JWTManager format
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

// validateTokenServiceToken validates tokens created by TokenService (newer format)
func (m *GinAuthMiddleware) validateTokenServiceToken(c *gin.Context, tokenString string) bool {
	// Parse token with MapClaims (TokenService format)
	// Use the same secret key that TokenService uses
	secretKey := []byte("your-super-secret-jwt-key-change-in-production-make-it-longer-and-more-secure-2024")

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return secretKey, nil
	})

	if err != nil || !token.Valid {
		return false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false
	}

	// Extract claims from TokenService format
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return false
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return false
	}

	email, _ := claims["email"].(string)
	role, _ := claims["role"].(string) // TokenService uses "role" instead of "role_name"

	// Set user context using TokenService format
	c.Set("user_id", userID.String())
	c.Set("email", email)
	c.Set("role_name", role) // Map "role" to "role_name" for compatibility

	// Create user context for multitenant middleware
	userContext := &models.UserContext{
		UserID:   userID,
		Role:     role,
		IsMaster: role == "master",
	}
	c.Set("userContext", userContext)

	c.Next()
	return true
}
