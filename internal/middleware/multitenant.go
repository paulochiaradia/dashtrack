package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
	"github.com/paulochiaradia/dashtrack/internal/utils"
)

// Global middleware functions that can be used directly
func RequireMasterRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext, exists := c.Get("userContext")
		if !exists {
			utils.UnauthorizedResponse(c, "User context not found")
			c.Abort()
			return
		}

		userCtx := userContext.(*models.UserContext)
		if userCtx.Role != "master" {
			utils.ForbiddenResponse(c, "Master role required")
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequireCompanyAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext, exists := c.Get("userContext")
		if !exists {
			utils.UnauthorizedResponse(c, "User context not found")
			c.Abort()
			return
		}

		userCtx := userContext.(*models.UserContext)
		if userCtx.CompanyID == nil {
			utils.ForbiddenResponse(c, "Company access required")
			c.Abort()
			return
		}

		// Store company ID in context for easy access
		c.Set("companyID", *userCtx.CompanyID)
		c.Next()
	}
}

func RequireCompanyAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext, exists := c.Get("userContext")
		if !exists {
			utils.UnauthorizedResponse(c, "User context not found")
			c.Abort()
			return
		}

		userCtx := userContext.(*models.UserContext)
		// Master sempre tem acesso, company_admin só tem acesso se for da mesma empresa
		if userCtx.Role != "company_admin" && userCtx.Role != "master" {
			utils.ForbiddenResponse(c, "Company admin role required")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireDriverOrHelper ensures user is a driver or helper
func RequireDriverOrHelper() gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext, exists := c.Get("userContext")
		if !exists {
			utils.UnauthorizedResponse(c, "User context not found")
			c.Abort()
			return
		}

		userCtx := userContext.(*models.UserContext)
		// Master e company_admin sempre têm acesso, drivers e helpers só aos seus veículos
		if !userCtx.IsMaster && userCtx.Role != "company_admin" && userCtx.Role != "driver" && userCtx.Role != "helper" {
			utils.ForbiddenResponse(c, "Driver or helper role required")
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireVehicleAccess ensures user has access to specific vehicle
func RequireVehicleAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		userContext, exists := c.Get("userContext")
		if !exists {
			utils.UnauthorizedResponse(c, "User context not found")
			c.Abort()
			return
		}

		userCtx := userContext.(*models.UserContext)

		// Master sempre tem acesso a todos os veículos
		if userCtx.IsMaster {
			c.Next()
			return
		}

		// Company admin tem acesso a todos os veículos da empresa
		if userCtx.Role == "company_admin" {
			c.Next()
			return
		}

		// Para drivers e helpers, verificação será feita no repository
		// baseado na atribuição de veículos específicos
		if userCtx.Role == "driver" || userCtx.Role == "helper" {
			c.Set("requireVehicleCheck", true)
			c.Next()
			return
		}

		utils.ForbiddenResponse(c, "Insufficient permissions for vehicle access")
		c.Abort()
	}
}

// MultiTenantMiddleware provides multi-tenant context and authorization
type MultiTenantMiddleware struct {
	userRepo *repository.UserRepository
	tracer   trace.Tracer
}

// NewMultiTenantMiddleware creates a new multi-tenant middleware
func NewMultiTenantMiddleware(userRepo *repository.UserRepository) *MultiTenantMiddleware {
	return &MultiTenantMiddleware{
		userRepo: userRepo,
		tracer:   otel.Tracer("multitenant-middleware"),
	}
}

// RequireCompanyAccess ensures the user has access to the specified company
func (m *MultiTenantMiddleware) RequireCompanyAccess() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, span := m.tracer.Start(c.Request.Context(), "MultiTenantMiddleware.RequireCompanyAccess")
		defer span.End()

		// Get user context from previous middleware
		userContext, exists := c.Get("userContext")
		if !exists {
			utils.UnauthorizedResponse(c, "User context not found")
			c.Abort()
			return
		}

		userCtx := userContext.(*models.UserContext)

		// Check if company ID is provided in URL params
		companyIDStr := c.Param("companyId")
		if companyIDStr == "" {
			// Try alternative parameter names
			companyIDStr = c.Param("company_id")
		}

		if companyIDStr != "" {
			companyID, err := uuid.Parse(companyIDStr)
			if err != nil {
				utils.BadRequestResponse(c, "Invalid company ID format")
				c.Abort()
				return
			}

			// Check if user has access to this company
			if !userCtx.HasCompanyAccess(companyID) {
				span.SetAttributes(
					attribute.String("user.id", userCtx.UserID.String()),
					attribute.String("company.id", companyID.String()),
					attribute.Bool("access_denied", true),
				)
				utils.ForbiddenResponse(c, "Access denied to this company")
				c.Abort()
				return
			}

			// Add company ID to context for handlers to use
			c.Set("companyID", companyID)
			span.SetAttributes(attribute.String("company.id", companyID.String()))
		}

		span.SetAttributes(attribute.String("user.id", userCtx.UserID.String()))
		c.Next()
	}
}

// RequireMasterRole ensures the user has master role
func (m *MultiTenantMiddleware) RequireMasterRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, span := m.tracer.Start(c.Request.Context(), "MultiTenantMiddleware.RequireMasterRole")
		defer span.End()

		userContext, exists := c.Get("userContext")
		if !exists {
			utils.UnauthorizedResponse(c, "User context not found")
			c.Abort()
			return
		}

		userCtx := userContext.(*models.UserContext)
		if !userCtx.IsMaster {
			span.SetAttributes(
				attribute.String("user.id", userCtx.UserID.String()),
				attribute.String("user.role", userCtx.Role),
				attribute.Bool("access_denied", true),
			)
			utils.ForbiddenResponse(c, "Master role required")
			c.Abort()
			return
		}

		span.SetAttributes(
			attribute.String("user.id", userCtx.UserID.String()),
			attribute.Bool("is_master", true),
		)
		c.Next()
	}
}

// RequireCompanyAdmin ensures the user can manage company resources
func (m *MultiTenantMiddleware) RequireCompanyAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, span := m.tracer.Start(c.Request.Context(), "MultiTenantMiddleware.RequireCompanyAdmin")
		defer span.End()

		userContext, exists := c.Get("userContext")
		if !exists {
			utils.UnauthorizedResponse(c, "User context not found")
			c.Abort()
			return
		}

		userCtx := userContext.(*models.UserContext)
		if !userCtx.CanManageCompany() {
			span.SetAttributes(
				attribute.String("user.id", userCtx.UserID.String()),
				attribute.String("user.role", userCtx.Role),
				attribute.Bool("access_denied", true),
			)
			utils.ForbiddenResponse(c, "Company admin privileges required")
			c.Abort()
			return
		}

		span.SetAttributes(
			attribute.String("user.id", userCtx.UserID.String()),
			attribute.String("user.role", userCtx.Role),
		)
		c.Next()
	}
}

// InjectUserContext retrieves and injects user context into the request
func (m *MultiTenantMiddleware) InjectUserContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, span := m.tracer.Start(c.Request.Context(), "MultiTenantMiddleware.InjectUserContext")
		defer span.End()

		// Get user ID from existing auth middleware
		userIDInterface, exists := c.Get("userID")
		if !exists {
			utils.UnauthorizedResponse(c, "User ID not found in context")
			c.Abort()
			return
		}

		userID, ok := userIDInterface.(uuid.UUID)
		if !ok {
			utils.UnauthorizedResponse(c, "Invalid user ID format in context")
			c.Abort()
			return
		}

		// Get user context from repository
		userContext, err := m.userRepo.GetUserContext(c.Request.Context(), userID)
		if err != nil {
			span.RecordError(err)
			utils.InternalServerErrorResponse(c, "Failed to retrieve user context")
			c.Abort()
			return
		}

		if userContext == nil {
			utils.UnauthorizedResponse(c, "User not found or inactive")
			c.Abort()
			return
		}

		// Inject user context into request
		c.Set("userContext", userContext)

		span.SetAttributes(
			attribute.String("user.id", userContext.UserID.String()),
			attribute.String("user.role", userContext.Role),
			attribute.Bool("is_master", userContext.IsMaster),
		)

		if userContext.CompanyID != nil {
			span.SetAttributes(attribute.String("company.id", userContext.CompanyID.String()))
		}

		c.Next()
	}
}

// RequireCompanyScope ensures operations are scoped to user's company
func (m *MultiTenantMiddleware) RequireCompanyScope() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, span := m.tracer.Start(c.Request.Context(), "MultiTenantMiddleware.RequireCompanyScope")
		defer span.End()

		userContext, exists := c.Get("userContext")
		if !exists {
			utils.UnauthorizedResponse(c, "User context not found")
			c.Abort()
			return
		}

		userCtx := userContext.(*models.UserContext)

		// Master users can access any scope
		if userCtx.IsMaster {
			span.SetAttributes(attribute.Bool("master_bypass", true))
			c.Next()
			return
		}

		// Company users must have a company ID
		if userCtx.CompanyID == nil {
			utils.ForbiddenResponse(c, "User is not associated with any company")
			c.Abort()
			return
		}

		// Add user's company ID as the scope for all operations
		c.Set("scopeCompanyID", *userCtx.CompanyID)

		span.SetAttributes(
			attribute.String("user.id", userCtx.UserID.String()),
			attribute.String("scope.company.id", userCtx.CompanyID.String()),
		)

		c.Next()
	}
}

// GetCompanyIDFromContext retrieves company ID from context with fallback logic
func GetCompanyIDFromContext(c *gin.Context) (*uuid.UUID, error) {
	// First try to get from URL parameter
	if companyIDInterface, exists := c.Get("companyID"); exists {
		if companyID, ok := companyIDInterface.(uuid.UUID); ok {
			return &companyID, nil
		}
	}

	// Then try to get from scope (user's company)
	if scopeCompanyIDInterface, exists := c.Get("scopeCompanyID"); exists {
		if scopeCompanyID, ok := scopeCompanyIDInterface.(uuid.UUID); ok {
			return &scopeCompanyID, nil
		}
	}

	// Finally try to get from user context
	if userContextInterface, exists := c.Get("userContext"); exists {
		if userContext, ok := userContextInterface.(*models.UserContext); ok {
			return userContext.CompanyID, nil
		}
	}

	return nil, nil
}

// ExtractUserContext is a helper to get user context from gin context
func ExtractUserContext(c *gin.Context) (*models.UserContext, bool) {
	userContext, exists := c.Get("userContext")
	if !exists {
		return nil, false
	}

	userCtx, ok := userContext.(*models.UserContext)
	return userCtx, ok
}

// GetUserIDFromContext retrieves user ID from context
func GetUserIDFromContext(c *gin.Context) (*uuid.UUID, error) {
	// Try to get from user context
	if userContextInterface, exists := c.Get("userContext"); exists {
		if userContext, ok := userContextInterface.(*models.UserContext); ok {
			return &userContext.UserID, nil
		}
	}

	return nil, nil
}
