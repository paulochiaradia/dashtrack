package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
)

var (
	ErrUserNotFound            = errors.New("user not found")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrCannotModifyOwnRole     = errors.New("cannot modify your own role")
	ErrCompanyMismatch         = errors.New("user does not belong to your company")
	ErrEmailAlreadyExists      = errors.New("email already exists")
	ErrInvalidRole             = errors.New("invalid role")
	ErrInvalidCompany          = errors.New("invalid company")
	ErrCannotDeleteSelf        = errors.New("cannot delete yourself")
)

// UserService handles user business logic with multi-tenant permissions
type UserService struct {
	userRepo   repository.UserRepositoryInterface
	roleRepo   repository.RoleRepositoryInterface
	bcryptCost int
}

// NewUserService creates a new user service
func NewUserService(userRepo repository.UserRepositoryInterface, roleRepo repository.RoleRepositoryInterface, bcryptCost int) *UserService {
	return &UserService{
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		bcryptCost: bcryptCost,
	}
}

// UserListRequest represents request parameters for listing users
type UserListRequest struct {
	Page   int   `json:"page" form:"page" binding:"min=1"`
	Limit  int   `json:"limit" form:"limit" binding:"min=1,max=100"`
	Active *bool `json:"active" form:"active"`
}

// UserListResponse represents paginated user list response
type UserListResponse struct {
	Users      []*models.User `json:"users"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
}

// GetUsers retrieves users based on the requesting user's permissions
func (s *UserService) GetUsers(ctx context.Context, requesterContext *models.UserContext, req UserListRequest) (*UserListResponse, error) {
	offset := (req.Page - 1) * req.Limit

	var users []*models.User
	var total int
	var err error

	switch requesterContext.Role {
	case "master":
		// Master can see all users
		users, err = s.userRepo.List(ctx, req.Limit, offset, req.Active, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to list all users: %w", err)
		}
		total, err = s.userRepo.CountUsers(ctx, nil)

	case "company_admin":
		// Company admin can see users from their company
		if requesterContext.CompanyID == nil {
			return nil, ErrInsufficientPermissions
		}

		// Company admins can see all roles in their company
		roles := []string{"company_admin", "manager", "driver", "helper"}
		users, err = s.userRepo.ListByCompanyAndRoles(ctx, requesterContext.CompanyID, roles, req.Limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to list company users: %w", err)
		}
		total, err = s.userRepo.CountByCompanyAndRoles(ctx, requesterContext.CompanyID, roles)

	case "admin":
		// General admin can only see drivers and helpers (no company restriction)
		roles := []string{"driver", "helper"}
		users, err = s.userRepo.ListByRoles(ctx, roles, req.Limit, offset)
		if err != nil {
			return nil, fmt.Errorf("failed to list drivers and helpers: %w", err)
		}
		total, err = s.userRepo.CountByCompanyAndRoles(ctx, nil, roles)

	default:
		return nil, ErrInsufficientPermissions
	}

	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	totalPages := (total + req.Limit - 1) / req.Limit

	return &UserListResponse{
		Users:      users,
		Total:      total,
		Page:       req.Page,
		Limit:      req.Limit,
		TotalPages: totalPages,
	}, nil
}

// GetUserByID retrieves a user by ID with permission checks
func (s *UserService) GetUserByID(ctx context.Context, requesterContext *models.UserContext, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Check permissions
	if !s.canAccessUser(requesterContext, user) {
		return nil, ErrInsufficientPermissions
	}

	// Remove sensitive data
	user.Password = ""
	return user, nil
}

// CreateUser creates a new user with permission checks
func (s *UserService) CreateUser(ctx context.Context, requesterContext *models.UserContext, req models.CreateUserRequest) (*models.User, error) {
	// Check if requester can create users
	if !s.canCreateUser(requesterContext, req.RoleID) {
		return nil, ErrInsufficientPermissions
	}

	// Parse role ID
	roleID, err := uuid.Parse(req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	// Get role to validate
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return nil, errors.New("role not found")
	}

	// Validate role creation permissions
	if !s.canCreateUserWithRole(requesterContext, role.Name) {
		return nil, fmt.Errorf("cannot create user with role %s", role.Name)
	}

	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), s.bcryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Set company ID based on requester
	var companyID *uuid.UUID
	if req.CompanyID != nil {
		parsedCompanyID, err := uuid.Parse(*req.CompanyID)
		if err != nil {
			return nil, fmt.Errorf("invalid company ID: %w", err)
		}
		companyID = &parsedCompanyID
	} else if requesterContext.CompanyID != nil {
		// If no company specified but requester has one, use requester's company
		companyID = requesterContext.CompanyID
	}

	// Validate company permissions
	if !s.canAssignToCompany(requesterContext, companyID) {
		return nil, ErrInsufficientPermissions
	}

	// Create user
	user := &models.User{
		ID:                uuid.New(),
		Name:              req.Name,
		Email:             req.Email,
		Password:          string(hashedPassword),
		Phone:             &req.Phone,
		CPF:               &req.CPF,
		RoleID:            roleID,
		CompanyID:         companyID,
		Active:            true,
		LoginAttempts:     0,
		PasswordChangedAt: time.Now(),
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	err = s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Fetch created user with role information
	createdUser, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created user: %w", err)
	}

	// Remove sensitive data
	createdUser.Password = ""
	return createdUser, nil
}

// UpdateUser updates a user with permission checks
func (s *UserService) UpdateUser(ctx context.Context, requesterContext *models.UserContext, userID uuid.UUID, req models.UpdateUserRequest) (*models.User, error) {
	// Get existing user
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if existingUser == nil {
		return nil, ErrUserNotFound
	}

	// Check permissions
	if !s.canModifyUser(requesterContext, existingUser) {
		return nil, ErrInsufficientPermissions
	}

	// Check if trying to modify own role
	if requesterContext.UserID == userID && req.RoleID != "" {
		return nil, ErrCannotModifyOwnRole
	}

	// If email is being changed, check uniqueness
	if req.Email != "" && req.Email != existingUser.Email {
		emailUser, err := s.userRepo.GetByEmail(ctx, req.Email)
		if err != nil {
			return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
		}
		if emailUser != nil && emailUser.ID != userID {
			return nil, ErrEmailAlreadyExists
		}
	}

	// Update user
	updatedUser, err := s.userRepo.Update(ctx, userID, req)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Remove sensitive data
	updatedUser.Password = ""
	return updatedUser, nil
}

// DeleteUser deletes a user with permission checks
func (s *UserService) DeleteUser(ctx context.Context, requesterContext *models.UserContext, userID uuid.UUID) error {
	// Get existing user
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	if existingUser == nil {
		return ErrUserNotFound
	}

	// Check permissions
	if !s.canDeleteUser(requesterContext, existingUser) {
		return ErrInsufficientPermissions
	}

	// Prevent self-deletion
	if requesterContext.UserID == userID {
		return errors.New("cannot delete your own account")
	}

	return s.userRepo.Delete(ctx, userID)
}

// Permission helper methods

func (s *UserService) canAccessUser(requesterContext *models.UserContext, targetUser *models.User) bool {
	switch requesterContext.Role {
	case "master":
		return true
	case "company_admin":
		return requesterContext.CompanyID != nil &&
			targetUser.CompanyID != nil &&
			*requesterContext.CompanyID == *targetUser.CompanyID
	case "admin":
		return targetUser.Role != nil && (targetUser.Role.Name == "driver" || targetUser.Role.Name == "helper")
	case "driver", "helper":
		return requesterContext.UserID == targetUser.ID
	default:
		return false
	}
}

func (s *UserService) canCreateUser(requesterContext *models.UserContext, roleID string) bool {
	switch requesterContext.Role {
	case "master":
		return true
	case "company_admin":
		return true // Can create users in their company
	case "admin":
		return true // Can create drivers and helpers
	default:
		return false
	}
}

func (s *UserService) canCreateUserWithRole(requesterContext *models.UserContext, roleName string) bool {
	switch requesterContext.Role {
	case "master":
		return true // Can create any role
	case "company_admin":
		// Can create company_admin, manager, driver, helper in their company
		allowedRoles := []string{"company_admin", "manager", "driver", "helper"}
		for _, allowed := range allowedRoles {
			if roleName == allowed {
				return true
			}
		}
		return false
	case "admin":
		// Can only create driver and helper
		return roleName == "driver" || roleName == "helper"
	default:
		return false
	}
}

func (s *UserService) canModifyUser(requesterContext *models.UserContext, targetUser *models.User) bool {
	switch requesterContext.Role {
	case "master":
		return true
	case "company_admin":
		return requesterContext.CompanyID != nil &&
			targetUser.CompanyID != nil &&
			*requesterContext.CompanyID == *targetUser.CompanyID
	case "admin":
		return targetUser.Role != nil && (targetUser.Role.Name == "driver" || targetUser.Role.Name == "helper")
	case "driver", "helper":
		return requesterContext.UserID == targetUser.ID
	default:
		return false
	}
}

func (s *UserService) canDeleteUser(requesterContext *models.UserContext, targetUser *models.User) bool {
	switch requesterContext.Role {
	case "master":
		return true
	case "company_admin":
		// Can delete users from their company except other company_admins unless it's themselves
		if requesterContext.CompanyID == nil || targetUser.CompanyID == nil {
			return false
		}
		if *requesterContext.CompanyID != *targetUser.CompanyID {
			return false
		}
		// Cannot delete other company admins
		if targetUser.Role != nil && targetUser.Role.Name == "company_admin" && requesterContext.UserID != targetUser.ID {
			return false
		}
		return true
	case "admin":
		return targetUser.Role != nil && (targetUser.Role.Name == "driver" || targetUser.Role.Name == "helper")
	default:
		return false
	}
}

func (s *UserService) canAssignToCompany(requesterContext *models.UserContext, companyID *uuid.UUID) bool {
	switch requesterContext.Role {
	case "master":
		return true // Can assign to any company
	case "company_admin":
		// Can only assign to their own company
		return requesterContext.CompanyID != nil &&
			companyID != nil &&
			*requesterContext.CompanyID == *companyID
	case "admin":
		// General admin can create users without company restriction
		return true
	default:
		return false
	}
}
