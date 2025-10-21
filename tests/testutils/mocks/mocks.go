package mocks

import (
	"context"
	"reflect"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/services"
)

// MockUserRepository is a mock of UserRepositoryInterface interface.
type MockUserRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUserRepositoryMockRecorder
}

// MockUserRepositoryMockRecorder is the mock recorder for MockUserRepository.
type MockUserRepositoryMockRecorder struct {
	mock *MockUserRepository
}

// NewMockUserRepository creates a new mock instance.
func NewMockUserRepository(ctrl *gomock.Controller) *MockUserRepository {
	mock := &MockUserRepository{ctrl: ctrl}
	mock.recorder = &MockUserRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserRepository) EXPECT() *MockUserRepositoryMockRecorder {
	return m.recorder
}

// CountByCompanyAndRoles mocks base method.
func (m *MockUserRepository) CountByCompanyAndRoles(ctx context.Context, companyID *uuid.UUID, roles []string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CountByCompanyAndRoles", ctx, companyID, roles)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CountByCompanyAndRoles indicates an expected call of CountByCompanyAndRoles.
func (mr *MockUserRepositoryMockRecorder) CountByCompanyAndRoles(ctx, companyID, roles interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountByCompanyAndRoles", reflect.TypeOf((*MockUserRepository)(nil).CountByCompanyAndRoles), ctx, companyID, roles)
}

// Create mocks base method.
func (m *MockUserRepository) Create(ctx context.Context, user *models.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", ctx, user)
	ret0, _ := ret[0].(error)
	return ret0
}

// Create indicates an expected call of Create.
func (mr *MockUserRepositoryMockRecorder) Create(ctx, user interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockUserRepository)(nil).Create), ctx, user)
}

// Delete mocks base method.
func (m *MockUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Delete", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// Delete indicates an expected call of Delete.
func (mr *MockUserRepositoryMockRecorder) Delete(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Delete", reflect.TypeOf((*MockUserRepository)(nil).Delete), ctx, id)
}

// GetByEmail mocks base method.
func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByEmail", ctx, email)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByEmail indicates an expected call of GetByEmail.
func (mr *MockUserRepositoryMockRecorder) GetByEmail(ctx, email interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByEmail", reflect.TypeOf((*MockUserRepository)(nil).GetByEmail), ctx, email)
}

// GetByID mocks base method.
func (m *MockUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockUserRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockUserRepository)(nil).GetByID), ctx, id)
}

// ListByCompanyAndRoles mocks base method.
func (m *MockUserRepository) ListByCompanyAndRoles(ctx context.Context, companyID *uuid.UUID, roles []string, limit, offset int) ([]*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListByCompanyAndRoles", ctx, companyID, roles, limit, offset)
	ret0, _ := ret[0].([]*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListByCompanyAndRoles indicates an expected call of ListByCompanyAndRoles.
func (mr *MockUserRepositoryMockRecorder) ListByCompanyAndRoles(ctx, companyID, roles, limit, offset interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByCompanyAndRoles", reflect.TypeOf((*MockUserRepository)(nil).ListByCompanyAndRoles), ctx, companyID, roles, limit, offset)
}

// Update mocks base method.
func (m *MockUserRepository) Update(ctx context.Context, id uuid.UUID, req models.UpdateUserRequest) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, id, req)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Update indicates an expected call of Update.
func (mr *MockUserRepositoryMockRecorder) Update(ctx, id, req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockUserRepository)(nil).Update), ctx, id, req)
}

// List mocks base method.
func (m *MockUserRepository) List(ctx context.Context, limit, offset int, active *bool, roleID *uuid.UUID) ([]*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, limit, offset, active, roleID)
	ret0, _ := ret[0].([]*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockUserRepositoryMockRecorder) List(ctx, limit, offset, active, roleID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockUserRepository)(nil).List), ctx, limit, offset, active, roleID)
}

// CountUsers mocks base method.
func (m *MockUserRepository) CountUsers(ctx context.Context, companyID *uuid.UUID) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CountUsers", ctx, companyID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CountUsers indicates an expected call of CountUsers.
func (mr *MockUserRepositoryMockRecorder) CountUsers(ctx, companyID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountUsers", reflect.TypeOf((*MockUserRepository)(nil).CountUsers), ctx, companyID)
}

// CountActiveUsers mocks base method.
func (m *MockUserRepository) CountActiveUsers(ctx context.Context, companyID *uuid.UUID) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CountActiveUsers", ctx, companyID)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CountActiveUsers indicates an expected call of CountActiveUsers.
func (mr *MockUserRepositoryMockRecorder) CountActiveUsers(ctx, companyID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountActiveUsers", reflect.TypeOf((*MockUserRepository)(nil).CountActiveUsers), ctx, companyID)
}

// GetByCompany mocks base method.
func (m *MockUserRepository) GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByCompany", ctx, companyID, limit, offset)
	ret0, _ := ret[0].([]*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByCompany indicates an expected call of GetByCompany.
func (mr *MockUserRepositoryMockRecorder) GetByCompany(ctx, companyID, limit, offset interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByCompany", reflect.TypeOf((*MockUserRepository)(nil).GetByCompany), ctx, companyID, limit, offset)
}

// ListByRoles mocks base method.
func (m *MockUserRepository) ListByRoles(ctx context.Context, roles []string, limit, offset int) ([]*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListByRoles", ctx, roles, limit, offset)
	ret0, _ := ret[0].([]*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListByRoles indicates an expected call of ListByRoles.
func (mr *MockUserRepositoryMockRecorder) ListByRoles(ctx, roles, limit, offset interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListByRoles", reflect.TypeOf((*MockUserRepository)(nil).ListByRoles), ctx, roles, limit, offset)
}

// UpdateLoginAttempts mocks base method.
func (m *MockUserRepository) UpdateLoginAttempts(ctx context.Context, id uuid.UUID, attempts int, blockedUntil *time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateLoginAttempts", ctx, id, attempts, blockedUntil)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateLoginAttempts indicates an expected call of UpdateLoginAttempts.
func (mr *MockUserRepositoryMockRecorder) UpdateLoginAttempts(ctx, id, attempts, blockedUntil interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateLoginAttempts", reflect.TypeOf((*MockUserRepository)(nil).UpdateLoginAttempts), ctx, id, attempts, blockedUntil)
}

// UpdateLastLogin mocks base method.
func (m *MockUserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateLastLogin", ctx, id)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateLastLogin indicates an expected call of UpdateLastLogin.
func (mr *MockUserRepositoryMockRecorder) UpdateLastLogin(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateLastLogin", reflect.TypeOf((*MockUserRepository)(nil).UpdateLastLogin), ctx, id)
}

// UpdateCompany mocks base method.
func (m *MockUserRepository) UpdateCompany(ctx context.Context, userID, companyID uuid.UUID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateCompany", ctx, userID, companyID)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateCompany indicates an expected call of UpdateCompany.
func (mr *MockUserRepositoryMockRecorder) UpdateCompany(ctx, userID, companyID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateCompany", reflect.TypeOf((*MockUserRepository)(nil).UpdateCompany), ctx, userID, companyID)
}

// UpdatePassword mocks base method.
func (m *MockUserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, hashedPassword string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePassword", ctx, id, hashedPassword)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePassword indicates an expected call of UpdatePassword.
func (mr *MockUserRepositoryMockRecorder) UpdatePassword(ctx, id, hashedPassword interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePassword", reflect.TypeOf((*MockUserRepository)(nil).UpdatePassword), ctx, id, hashedPassword)
}

// GetUserContext mocks base method.
func (m *MockUserRepository) GetUserContext(ctx context.Context, userID uuid.UUID) (*models.UserContext, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserContext", ctx, userID)
	ret0, _ := ret[0].(*models.UserContext)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserContext indicates an expected call of GetUserContext.
func (mr *MockUserRepositoryMockRecorder) GetUserContext(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserContext", reflect.TypeOf((*MockUserRepository)(nil).GetUserContext), ctx, userID)
}

// MockRoleRepository is a mock of RoleRepositoryInterface interface.
type MockRoleRepository struct {
	ctrl     *gomock.Controller
	recorder *MockRoleRepositoryMockRecorder
}

// MockRoleRepositoryMockRecorder is the mock recorder for MockRoleRepository.
type MockRoleRepositoryMockRecorder struct {
	mock *MockRoleRepository
}

// NewMockRoleRepository creates a new mock instance.
func NewMockRoleRepository(ctrl *gomock.Controller) *MockRoleRepository {
	mock := &MockRoleRepository{ctrl: ctrl}
	mock.recorder = &MockRoleRepositoryMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRoleRepository) EXPECT() *MockRoleRepositoryMockRecorder {
	return m.recorder
}

// GetByName mocks base method.
func (m *MockRoleRepository) GetByName(ctx context.Context, name string) (*models.Role, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByName", ctx, name)
	ret0, _ := ret[0].(*models.Role)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByName indicates an expected call of GetByName.
func (mr *MockRoleRepositoryMockRecorder) GetByName(ctx, name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByName", reflect.TypeOf((*MockRoleRepository)(nil).GetByName), ctx, name)
}

// GetAll mocks base method.
func (m *MockRoleRepository) GetAll(ctx context.Context) ([]*models.Role, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll", ctx)
	ret0, _ := ret[0].([]*models.Role)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll.
func (mr *MockRoleRepositoryMockRecorder) GetAll(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockRoleRepository)(nil).GetAll), ctx)
}

// GetByID mocks base method.
func (m *MockRoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Role, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, id)
	ret0, _ := ret[0].(*models.Role)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockRoleRepositoryMockRecorder) GetByID(ctx, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockRoleRepository)(nil).GetByID), ctx, id)
}

// MockJWTManager is a mock of JWTManager interface.
type MockJWTManager struct {
	ctrl     *gomock.Controller
	recorder *MockJWTManagerMockRecorder
}

// MockJWTManagerMockRecorder is the mock recorder for MockJWTManager.
type MockJWTManagerMockRecorder struct {
	mock *MockJWTManager
}

// NewMockJWTManager creates a new mock instance.
func NewMockJWTManager(ctrl *gomock.Controller) *MockJWTManager {
	mock := &MockJWTManager{ctrl: ctrl}
	mock.recorder = &MockJWTManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockJWTManager) EXPECT() *MockJWTManagerMockRecorder {
	return m.recorder
}

// GenerateToken mocks base method.
func (m *MockJWTManager) GenerateToken(userID uuid.UUID, email, role string, companyID *uuid.UUID) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GenerateToken", userID, email, role, companyID)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GenerateToken indicates an expected call of GenerateToken.
func (mr *MockJWTManagerMockRecorder) GenerateToken(userID, email, role, companyID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateToken", reflect.TypeOf((*MockJWTManager)(nil).GenerateToken), userID, email, role, companyID)
}

// ValidateToken mocks base method - DEPRECATED: Use TokenService instead
// func (m *MockJWTManager) ValidateToken(tokenString string) (*models.UserContext, error) {
// 	m.ctrl.T.Helper()
// 	ret := m.ctrl.Call(m, "ValidateToken", tokenString)
// 	ret0, _ := ret[0].(*models.UserContext)
// 	ret1, _ := ret[1].(error)
// 	return ret0, ret1
// }

// ValidateToken indicates an expected call of ValidateToken.
// func (mr *MockJWTManagerMockRecorder) ValidateToken(tokenString interface{}) *gomock.Call {
// 	mr.mock.ctrl.T.Helper()
// 	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateToken", reflect.TypeOf((*MockJWTManager)(nil).ValidateToken), tokenString)
// }

// GenerateTokens mocks base method - DEPRECATED: Use TokenService instead
// func (m *MockJWTManager) GenerateTokens(userContext models.UserContext) (string, string, error) {
// 	m.ctrl.T.Helper()
// 	ret := m.ctrl.Call(m, "GenerateTokens", userContext)
// 	ret0, _ := ret[0].(string)
// 	ret1, _ := ret[1].(string)
// 	ret2, _ := ret[2].(error)
// 	return ret0, ret1, ret2
// }

// GenerateTokens indicates an expected call of GenerateTokens.
// func (mr *MockJWTManagerMockRecorder) GenerateTokens(userContext interface{}) *gomock.Call {
// 	mr.mock.ctrl.T.Helper()
// 	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GenerateTokens", reflect.TypeOf((*MockJWTManager)(nil).GenerateTokens), userContext)
// }

// ValidateRefreshToken mocks base method.
func (m *MockJWTManager) ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateRefreshToken", tokenString)
	ret0, _ := ret[0].(uuid.UUID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidateRefreshToken indicates an expected call of ValidateRefreshToken.
func (mr *MockJWTManagerMockRecorder) ValidateRefreshToken(tokenString interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateRefreshToken", reflect.TypeOf((*MockJWTManager)(nil).ValidateRefreshToken), tokenString)
}

// RefreshToken mocks base method - DEPRECATED: Use TokenService instead
// func (m *MockJWTManager) RefreshToken(refreshTokenString string, userContext models.UserContext) (string, string, error) {
// 	m.ctrl.T.Helper()
// 	ret := m.ctrl.Call(m, "RefreshToken", refreshTokenString, userContext)
// 	ret0, _ := ret[0].(string)
// 	ret1, _ := ret[1].(string)
// 	ret2, _ := ret[2].(error)
// 	return ret0, ret1, ret2
// }

// RefreshToken indicates an expected call of RefreshToken.
// func (mr *MockJWTManagerMockRecorder) RefreshToken(refreshTokenString, userContext interface{}) *gomock.Call {
// 	mr.mock.ctrl.T.Helper()
// 	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshToken", reflect.TypeOf((*MockJWTManager)(nil).RefreshToken), refreshTokenString, userContext)
// }

// GeneratePasswordResetToken mocks base method.
func (m *MockJWTManager) GeneratePasswordResetToken(userID uuid.UUID, email string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GeneratePasswordResetToken", userID, email)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GeneratePasswordResetToken indicates an expected call of GeneratePasswordResetToken.
func (mr *MockJWTManagerMockRecorder) GeneratePasswordResetToken(userID, email interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GeneratePasswordResetToken", reflect.TypeOf((*MockJWTManager)(nil).GeneratePasswordResetToken), userID, email)
}

// ValidatePasswordResetToken mocks base method.
func (m *MockJWTManager) ValidatePasswordResetToken(tokenString string) (uuid.UUID, string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidatePasswordResetToken", tokenString)
	ret0, _ := ret[0].(uuid.UUID)
	ret1, _ := ret[1].(string)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// ValidatePasswordResetToken indicates an expected call of ValidatePasswordResetToken.
func (mr *MockJWTManagerMockRecorder) ValidatePasswordResetToken(tokenString interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidatePasswordResetToken", reflect.TypeOf((*MockJWTManager)(nil).ValidatePasswordResetToken), tokenString)
}

// MockUserService is a mock of UserService interface.
type MockUserService struct {
	ctrl     *gomock.Controller
	recorder *MockUserServiceMockRecorder
}

// MockUserServiceMockRecorder is the mock recorder for MockUserService.
type MockUserServiceMockRecorder struct {
	mock *MockUserService
}

// NewMockUserService creates a new mock instance.
func NewMockUserService(ctrl *gomock.Controller) *MockUserService {
	mock := &MockUserService{ctrl: ctrl}
	mock.recorder = &MockUserServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserService) EXPECT() *MockUserServiceMockRecorder {
	return m.recorder
}

// GetByEmail mocks base method.
func (m *MockUserService) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByEmail", ctx, email)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByEmail indicates an expected call of GetByEmail.
func (mr *MockUserServiceMockRecorder) GetByEmail(ctx, email interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByEmail", reflect.TypeOf((*MockUserService)(nil).GetByEmail), ctx, email)
}

// GetByID mocks base method.
func (m *MockUserService) GetByID(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetByID", ctx, userID)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetByID indicates an expected call of GetByID.
func (mr *MockUserServiceMockRecorder) GetByID(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetByID", reflect.TypeOf((*MockUserService)(nil).GetByID), ctx, userID)
}

// ValidatePassword mocks base method.
func (m *MockUserService) ValidatePassword(hashedPassword, password string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidatePassword", hashedPassword, password)
	ret0, _ := ret[0].(bool)
	return ret0
}

// ValidatePassword indicates an expected call of ValidatePassword.
func (mr *MockUserServiceMockRecorder) ValidatePassword(hashedPassword, password interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidatePassword", reflect.TypeOf((*MockUserService)(nil).ValidatePassword), hashedPassword, password)
}

// GetUsers mocks base method.
func (m *MockUserService) GetUsers(ctx context.Context, req services.UserListRequest, currentUser *models.UserContext) (*services.UserListResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUsers", ctx, req, currentUser)
	ret0, _ := ret[0].(*services.UserListResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUsers indicates an expected call of GetUsers.
func (mr *MockUserServiceMockRecorder) GetUsers(ctx, req, currentUser interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUsers", reflect.TypeOf((*MockUserService)(nil).GetUsers), ctx, req, currentUser)
}

// CreateUser mocks base method.
func (m *MockUserService) CreateUser(ctx context.Context, req models.CreateUserRequest, currentUser *models.UserContext) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", ctx, req, currentUser)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockUserServiceMockRecorder) CreateUser(ctx, req, currentUser interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockUserService)(nil).CreateUser), ctx, req, currentUser)
}

// UpdateUser mocks base method.
func (m *MockUserService) UpdateUser(ctx context.Context, userID uuid.UUID, req models.UpdateUserRequest, currentUser *models.UserContext) (*models.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateUser", ctx, userID, req, currentUser)
	ret0, _ := ret[0].(*models.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateUser indicates an expected call of UpdateUser.
func (mr *MockUserServiceMockRecorder) UpdateUser(ctx, userID, req, currentUser interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateUser", reflect.TypeOf((*MockUserService)(nil).UpdateUser), ctx, userID, req, currentUser)
}

// DeleteUser mocks base method.
func (m *MockUserService) DeleteUser(ctx context.Context, userID uuid.UUID, currentUser *models.UserContext) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteUser", ctx, userID, currentUser)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteUser indicates an expected call of DeleteUser.
func (mr *MockUserServiceMockRecorder) DeleteUser(ctx, userID, currentUser interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteUser", reflect.TypeOf((*MockUserService)(nil).DeleteUser), ctx, userID, currentUser)
}
