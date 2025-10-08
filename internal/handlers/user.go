package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/services"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// getUserContext extracts UserContext from gin.Context
func (h *UserHandler) getUserContext(c *gin.Context) *models.UserContext {
	userContext, exists := c.Get("userContext")
	if !exists {
		return nil
	}
	return userContext.(*models.UserContext)
}

// GetUsers handles GET /users with multi-tenant support
func (h *UserHandler) GetUsers(c *gin.Context) {
	userContext := h.getUserContext(c)
	if userContext == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse query parameters
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	var active *bool
	if activeStr := c.Query("active"); activeStr != "" {
		if a, err := strconv.ParseBool(activeStr); err == nil {
			active = &a
		}
	}

	req := services.UserListRequest{
		Page:   page,
		Limit:  limit,
		Active: active,
	}

	response, err := h.userService.GetUsers(c.Request.Context(), userContext, req)
	if err != nil {
		if err == services.ErrInsufficientPermissions {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUserByID handles GET /users/:id
func (h *UserHandler) GetUserByID(c *gin.Context) {
	userContext := h.getUserContext(c)
	if userContext == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userContext, userID)
	if err != nil {
		if err == services.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if err == services.ErrInsufficientPermissions {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// CreateUser handles POST /users
func (h *UserHandler) CreateUser(c *gin.Context) {
	userContext := h.getUserContext(c)
	if userContext == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.CreateUser(c.Request.Context(), userContext, req)
	if err != nil {
		// Add detailed error logging
		fmt.Printf("ERROR: CreateUser failed: %v\n", err)
		switch err {
		case services.ErrInsufficientPermissions:
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		case services.ErrEmailAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		case services.ErrInvalidRole:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		case services.ErrInvalidCompany:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company"})
		case services.ErrRoleRequiresCompany:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Role requires company assignment"})
		case services.ErrRoleProhibitsCompany:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Role prohibits company assignment"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, user)
}

// UpdateUser handles PUT /users/:id
func (h *UserHandler) UpdateUser(c *gin.Context) {
	userContext := h.getUserContext(c)
	if userContext == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), userContext, userID, req)
	if err != nil {
		switch err {
		case services.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case services.ErrInsufficientPermissions:
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		case services.ErrEmailAlreadyExists:
			c.JSON(http.StatusConflict, gin.H{"error": "Email already exists"})
		case services.ErrCannotModifyOwnRole:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot modify own role"})
		case services.ErrInvalidRole:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser handles DELETE /users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	userContext := h.getUserContext(c)
	if userContext == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), userContext, userID)
	if err != nil {
		switch err {
		case services.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case services.ErrInsufficientPermissions:
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		case services.ErrCannotDeleteSelf:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete yourself"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// TransferUserToCompany handles PATCH /master/users/:id/transfer - Master only
func (h *UserHandler) TransferUserToCompany(c *gin.Context) {
	userContext := h.getUserContext(c)
	if userContext == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Only master can transfer users between companies
	if userContext.Role != "master" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only master users can transfer users between companies"})
		return
	}

	userIDStr := c.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req models.TransferUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	companyID, err := uuid.Parse(req.CompanyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid company ID"})
		return
	}

	// Perform the transfer
	err = h.userService.TransferUserToCompany(c.Request.Context(), userID, companyID, req.Reason)
	if err != nil {
		switch err {
		case services.ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		case services.ErrCompanyNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Company not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "User transferred successfully",
		"user_id":    userID,
		"company_id": companyID,
		"reason":     req.Reason,
	})
}
