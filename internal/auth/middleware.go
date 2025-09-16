package auth

import (
	"context"
	"net/http"
	"strings"
)

// ContextKey represents the key for context values
type ContextKey string

const (
	// UserContextKey is the key for user context in request context
	UserContextKey ContextKey = "user_context"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	jwtManager *JWTManager
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(jwtManager *JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// RequireAuth middleware that requires valid JWT token
func (a *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractTokenFromHeader(r)
		if token == "" {
			http.Error(w, "Authorization token required", http.StatusUnauthorized)
			return
		}

		claims, err := a.jwtManager.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Add user context to request
		userContext := claims.ToUserContext()
		ctx := context.WithValue(r.Context(), UserContextKey, userContext)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// RequireRole middleware that requires specific role
func (a *AuthMiddleware) RequireRole(roleName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userContext, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "User context not found", http.StatusInternalServerError)
				return
			}

			if userContext.RoleName != roleName {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyRole middleware that requires any of the specified roles
func (a *AuthMiddleware) RequireAnyRole(roleNames ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userContext, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "User context not found", http.StatusInternalServerError)
				return
			}

			hasPermission := false
			for _, roleName := range roleNames {
				if userContext.RoleName == roleName {
					hasPermission = true
					break
				}
			}

			if !hasPermission {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// OptionalAuth middleware that adds user context if token is present but doesn't require it
func (a *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractTokenFromHeader(r)
		if token != "" {
			claims, err := a.jwtManager.ValidateToken(token)
			if err == nil {
				userContext := claims.ToUserContext()
				ctx := context.WithValue(r.Context(), UserContextKey, userContext)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// extractTokenFromHeader extracts Bearer token from Authorization header
func extractTokenFromHeader(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer token format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// GetUserFromContext extracts user context from request context
func GetUserFromContext(ctx context.Context) (UserContext, bool) {
	userContext, ok := ctx.Value(UserContextKey).(UserContext)
	return userContext, ok
}

// IsAdmin checks if the current user is an admin
func IsAdmin(ctx context.Context) bool {
	userContext, ok := GetUserFromContext(ctx)
	if !ok {
		return false
	}
	return userContext.RoleName == "admin"
}

// IsManager checks if the current user is a manager or admin
func IsManager(ctx context.Context) bool {
	userContext, ok := GetUserFromContext(ctx)
	if !ok {
		return false
	}
	return userContext.RoleName == "manager" || userContext.RoleName == "admin"
}

// IsUser checks if the current user is authenticated (any role)
func IsUser(ctx context.Context) bool {
	_, ok := GetUserFromContext(ctx)
	return ok
}
