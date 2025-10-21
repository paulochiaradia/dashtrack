package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"

	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/services"
	"github.com/paulochiaradia/dashtrack/tests/testutils/mocks"
)

// UserServiceTestSuite defines the test suite for UserService
type UserServiceTestSuite struct {
	suite.Suite
	userService  *services.UserService
	mockUserRepo *mocks.MockUserRepository
	mockRoleRepo *mocks.MockRoleRepository
}

// adapter to add the missing Search method so the generated mock satisfies the repository interface.
// The Search implementation is a stub because tests in this file don't use Search directly.
type userRepoAdapter struct {
	*mocks.MockUserRepository
}

func (u *userRepoAdapter) Search(ctx context.Context, companyID *uuid.UUID, query string, limit, offset int) ([]*models.User, error) {
	return nil, nil
}

func (suite *UserServiceTestSuite) SetupTest() {
	ctrl := gomock.NewController(suite.T())
	suite.mockUserRepo = mocks.NewMockUserRepository(ctrl)
	suite.mockRoleRepo = mocks.NewMockRoleRepository(ctrl)

	// Create UserService with bcryptCost parameter, wrapping the mock to provide the missing Search method.
	suite.userService = services.NewUserService(&userRepoAdapter{suite.mockUserRepo}, suite.mockRoleRepo, bcrypt.DefaultCost)
}

func (suite *UserServiceTestSuite) TestCreateUser_Success() {
	ctx := context.Background()
	companyID := uuid.New()
	currentUser := &models.UserContext{
		UserID:    uuid.New(),
		CompanyID: &companyID,
		Role:      "company_admin",
		IsMaster:  false,
	}

	// Mock role repository to return valid role
	roleID := uuid.New()
	expectedRole := &models.Role{
		ID:   roleID,
		Name: "driver",
	}

	createReq := models.CreateUserRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
		Phone:    "1234567890",
		RoleID:   roleID.String(),
	}

	// Setup expectations
	// 1. Role validation
	suite.mockRoleRepo.EXPECT().
		GetByID(ctx, roleID).
		Return(expectedRole, nil)

	// 2. Email uniqueness check
	suite.mockUserRepo.EXPECT().
		GetByEmail(ctx, createReq.Email).
		Return(nil, nil) // No existing user

	// 3. User creation
	suite.mockUserRepo.EXPECT().
		Create(ctx, gomock.Any()).
		Return(nil)

	// 4. Fetch created user by ID (this happens at the end of CreateUser)
	createdUser := &models.User{
		ID:        uuid.New(),
		Name:      createReq.Name,
		Email:     createReq.Email,
		Phone:     &createReq.Phone,
		RoleID:    roleID,
		CompanyID: &companyID,
		Active:    true,
		Role:      expectedRole,
	}
	suite.mockUserRepo.EXPECT().
		GetByID(ctx, gomock.Any()).
		Return(createdUser, nil)

	// Test
	user, err := suite.userService.CreateUser(ctx, currentUser, createReq)

	// Assertions
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), user)
	assert.Equal(suite.T(), createReq.Name, user.Name)
	assert.Equal(suite.T(), createReq.Email, user.Email)
	assert.Equal(suite.T(), companyID, *user.CompanyID)
}

func (suite *UserServiceTestSuite) TestCreateUser_InsufficientPermissions() {
	ctx := context.Background()
	companyID := uuid.New()

	// Admin trying to create company_admin
	currentUser := &models.UserContext{
		UserID:    uuid.New(),
		CompanyID: &companyID,
		Role:      "admin",
		IsMaster:  false,
	}

	// Mock role for company_admin (higher level than admin)
	companyAdminRoleID := uuid.New()
	companyAdminRole := &models.Role{
		ID:   companyAdminRoleID,
		Name: "company_admin",
	}

	createReq := models.CreateUserRequest{
		Name:     "Test Admin",
		Email:    "admin@example.com",
		Password: "password123",
		RoleID:   companyAdminRoleID.String(),
	}

	// Setup expectation - should get role to validate permissions
	suite.mockRoleRepo.EXPECT().
		GetByID(ctx, companyAdminRoleID).
		Return(companyAdminRole, nil)

	// Test
	user, err := suite.userService.CreateUser(ctx, currentUser, createReq)

	// Assertions
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "cannot create user with role company_admin")
	assert.Nil(suite.T(), user)
}

func (suite *UserServiceTestSuite) TestGetUsers_Master_CanSeeAll() {
	ctx := context.Background()
	currentUser := &models.UserContext{
		UserID:   uuid.New(),
		Role:     "master",
		IsMaster: true,
	}

	req := services.UserListRequest{
		Page:  1,
		Limit: 10,
	}

	// Mock repository response
	expectedUsers := []*models.User{
		{
			ID:    uuid.New(),
			Name:  "User 1",
			Email: "user1@example.com",
			Role:  &models.Role{Name: "admin"},
		},
		{
			ID:    uuid.New(),
			Name:  "User 2",
			Email: "user2@example.com",
			Role:  &models.Role{Name: "driver"},
		},
	}

	// For master users, call List with appropriate parameters
	suite.mockUserRepo.EXPECT().
		List(ctx, 10, 0, req.Active, gomock.Any()).
		Return(expectedUsers, nil)

	suite.mockUserRepo.EXPECT().
		CountUsers(ctx, gomock.Any()).
		Return(2, nil)

	// Test
	result, err := suite.userService.GetUsers(ctx, currentUser, req)

	// Assertions
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result.Users, 2)
	assert.Equal(suite.T(), 2, result.Total)
	assert.Equal(suite.T(), 1, result.Page)
	assert.Equal(suite.T(), 10, result.Limit)
}

func (suite *UserServiceTestSuite) TestUpdateUser_Success() {
	ctx := context.Background()
	userID := uuid.New()
	companyID := uuid.New()

	currentUser := &models.UserContext{
		UserID:    uuid.New(),
		CompanyID: &companyID,
		Role:      "admin",
		IsMaster:  false,
	}

	updateReq := models.UpdateUserRequest{
		Name:   "Updated Name",
		Email:  "updated@example.com",
		Phone:  "9876543210",
		Active: boolPtr(true),
	}

	// Mock getting existing user
	existingUser := &models.User{
		ID:        userID,
		Name:      "Original Name",
		Email:     "original@example.com",
		CompanyID: &companyID,
		Role:      &models.Role{Name: "driver"},
	}

	suite.mockUserRepo.EXPECT().GetByID(ctx, userID).Return(existingUser, nil)

	// Mock email uniqueness check (since email is changing)
	suite.mockUserRepo.EXPECT().
		GetByEmail(ctx, updateReq.Email).
		Return(nil, nil) // No existing user with this email

	// Mock update
	updatedUser := &models.User{
		ID:        userID,
		Name:      updateReq.Name,
		Email:     updateReq.Email,
		Phone:     &updateReq.Phone,
		Active:    *updateReq.Active,
		CompanyID: &companyID,
		Role:      &models.Role{Name: "driver"},
		UpdatedAt: time.Now(),
	}

	suite.mockUserRepo.EXPECT().Update(ctx, userID, updateReq).Return(updatedUser, nil)

	// Test
	result, err := suite.userService.UpdateUser(ctx, currentUser, userID, updateReq)

	// Assertions
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), updateReq.Name, result.Name)
	assert.Equal(suite.T(), updateReq.Email, result.Email)
}

func (suite *UserServiceTestSuite) TestDeleteUser_Success() {
	ctx := context.Background()
	userID := uuid.New()
	companyID := uuid.New()

	currentUser := &models.UserContext{
		UserID:    uuid.New(), // Different from userID being deleted
		CompanyID: &companyID,
		Role:      "admin",
		IsMaster:  false,
	}

	// Mock getting existing user
	existingUser := &models.User{
		ID:        userID,
		Name:      "User to Delete",
		Email:     "delete@example.com",
		CompanyID: &companyID,
		Role:      &models.Role{Name: "driver"},
	}

	suite.mockUserRepo.EXPECT().GetByID(ctx, userID).Return(existingUser, nil)
	suite.mockUserRepo.EXPECT().Delete(ctx, userID).Return(nil)

	// Test
	err := suite.userService.DeleteUser(ctx, currentUser, userID)

	// Assertions
	assert.NoError(suite.T(), err)
}

func (suite *UserServiceTestSuite) TestDeleteUser_CannotDeleteSelf() {
	ctx := context.Background()
	userID := uuid.New()
	companyID := uuid.New()

	// User trying to delete themselves
	currentUser := &models.UserContext{
		UserID:    userID, // Same as userID being deleted
		CompanyID: &companyID,
		Role:      "company_admin", // Changed to company_admin so they can delete themselves
		IsMaster:  false,
	}

	// Mock getting existing user (needed for permission checks)
	existingUser := &models.User{
		ID:        userID,
		Name:      "Self User",
		Email:     "self@example.com",
		CompanyID: &companyID,
		Role:      &models.Role{Name: "company_admin"}, // Same role
	}

	suite.mockUserRepo.EXPECT().GetByID(ctx, userID).Return(existingUser, nil)

	// Test
	err := suite.userService.DeleteUser(ctx, currentUser, userID)

	// Assertions
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "cannot delete your own account")
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}

// Run the test suite
func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}
