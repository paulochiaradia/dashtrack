package repositories_test

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
)

// UserRepositoryTestSuite defines the test suite for UserRepository
type UserRepositoryTestSuite struct {
	suite.Suite
	db   *sqlx.DB
	mock sqlmock.Sqlmock
	repo *repository.UserRepository
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	// Create a mock database connection
	mockDB, mock, err := sqlmock.New()
	suite.Require().NoError(err)

	suite.db = sqlx.NewDb(mockDB, "sqlmock")
	suite.mock = mock
	suite.repo = repository.NewUserRepository(suite.db)
}

func (suite *UserRepositoryTestSuite) TearDownTest() {
	suite.db.Close()
}

func (suite *UserRepositoryTestSuite) TestCreate_Success() {
	ctx := context.Background()
	userID := uuid.New()
	companyID := uuid.New()
	roleID := uuid.New()

	user := &models.User{
		ID:        userID,
		Name:      "Test User",
		Email:     "test@example.com",
		Phone:     &[]string{"1234567890"}[0],
		CompanyID: &companyID,
		RoleID:    roleID,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock the INSERT query
	expectedQuery := `INSERT INTO users \( id, name, email, password, phone, cpf, avatar, role_id, company_id, active, dashboard_config, api_token, created_at, updated_at, password_changed_at \)`
	suite.mock.ExpectExec(expectedQuery).
		WithArgs(
			sqlmock.AnyArg(), // id - generated in Create method
			user.Name,
			user.Email,
			sqlmock.AnyArg(), // password
			user.Phone,
			sqlmock.AnyArg(), // cpf
			sqlmock.AnyArg(), // avatar
			user.RoleID,
			user.CompanyID,
			user.Active,
			sqlmock.AnyArg(), // dashboard_config
			sqlmock.AnyArg(), // api_token
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
			sqlmock.AnyArg(), // password_changed_at
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	// Test
	err := suite.repo.Create(ctx, user)

	// Assertions
	assert.NoError(suite.T(), err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *UserRepositoryTestSuite) TestCreate_DatabaseError() {
	ctx := context.Background()
	userID := uuid.New()
	companyID := uuid.New()
	roleID := uuid.New()

	user := &models.User{
		ID:        userID,
		Name:      "Test User",
		Email:     "test@example.com",
		Phone:     &[]string{"1234567890"}[0],
		CompanyID: &companyID,
		RoleID:    roleID,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Mock the INSERT query to return an error
	expectedQuery := `INSERT INTO users \( id, name, email, password, phone, cpf, avatar, role_id, company_id, active, dashboard_config, api_token, created_at, updated_at, password_changed_at \)`
	suite.mock.ExpectExec(expectedQuery).
		WithArgs(
			sqlmock.AnyArg(), // id - generated in Create method
			user.Name,
			user.Email,
			sqlmock.AnyArg(), // password
			user.Phone,
			sqlmock.AnyArg(), // cpf
			sqlmock.AnyArg(), // avatar
			user.RoleID,
			user.CompanyID,
			user.Active,
			sqlmock.AnyArg(), // dashboard_config
			sqlmock.AnyArg(), // api_token
			sqlmock.AnyArg(), // created_at
			sqlmock.AnyArg(), // updated_at
			sqlmock.AnyArg(), // password_changed_at
		).
		WillReturnError(sql.ErrConnDone)

	// Test
	err := suite.repo.Create(ctx, user)

	// Assertions
	assert.Error(suite.T(), err)
	assert.Contains(suite.T(), err.Error(), "failed to create user")
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *UserRepositoryTestSuite) TestGetByID_Success() {
	ctx := context.Background()
	userID := uuid.New()
	companyID := uuid.New()
	roleID := uuid.New()

	expectedUser := &models.User{
		ID:        userID,
		Name:      "Test User",
		Email:     "test@example.com",
		Phone:     &[]string{"1234567890"}[0],
		CompanyID: &companyID,
		RoleID:    roleID,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create rows with the expected data
	rows := sqlmock.NewRows([]string{
		"id", "name", "email", "password", "phone", "cpf", "avatar", "role_id", "company_id",
		"active", "last_login", "dashboard_config", "api_token", "login_attempts",
		"blocked_until", "password_changed_at", "created_at", "updated_at",
		"role_id", "role_name", "role_description", "role_created_at", "role_updated_at",
	}).AddRow(
		expectedUser.ID,
		expectedUser.Name,
		expectedUser.Email,
		"hashed_password",
		expectedUser.Phone,
		nil, // cpf
		nil, // avatar
		expectedUser.RoleID,
		expectedUser.CompanyID,
		expectedUser.Active,
		nil,                    // last_login
		nil,                    // dashboard_config
		nil,                    // api_token
		0,                      // login_attempts
		nil,                    // blocked_until
		expectedUser.CreatedAt, // password_changed_at
		expectedUser.CreatedAt,
		expectedUser.UpdatedAt,
		expectedUser.RoleID,    // role_id
		"driver",               // role_name
		"Driver role",          // role_description
		expectedUser.CreatedAt, // role_created_at
		expectedUser.UpdatedAt, // role_updated_at
	)

	// Mock the SELECT query
	expectedQuery := `SELECT u.id, u.name, u.email, u.password, u.phone, u.cpf, u.avatar, u.role_id, u.company_id, u.active, u.last_login, u.dashboard_config, u.api_token, u.login_attempts, u.blocked_until, u.password_changed_at, u.created_at, u.updated_at, r.id, r.name, r.description, r.created_at, r.updated_at FROM users u JOIN roles r ON u.role_id = r.id WHERE u.id = \$1`
	suite.mock.ExpectQuery(expectedQuery).
		WithArgs(userID).
		WillReturnRows(rows)

	// Test
	result, err := suite.repo.GetByID(ctx, userID)

	// Assertions
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedUser.ID, result.ID)
	assert.Equal(suite.T(), expectedUser.Name, result.Name)
	assert.Equal(suite.T(), expectedUser.Email, result.Email)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *UserRepositoryTestSuite) TestGetByID_NotFound() {
	ctx := context.Background()
	userID := uuid.New()

	// Mock the SELECT query to return no rows
	expectedQuery := regexp.QuoteMeta("SELECT u.id, u.name, u.email, u.password, u.phone, u.cpf, u.avatar, u.role_id, u.company_id")
	suite.mock.ExpectQuery(expectedQuery).
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	// Test
	result, err := suite.repo.GetByID(ctx, userID)

	// Assertions
	assert.NoError(suite.T(), err)
	assert.Nil(suite.T(), result)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *UserRepositoryTestSuite) TestGetByEmail_Success() {
	ctx := context.Background()
	userID := uuid.New()
	companyID := uuid.New()
	roleID := uuid.New()
	email := "test@example.com"

	expectedUser := &models.User{
		ID:        userID,
		Name:      "Test User",
		Email:     email,
		Phone:     &[]string{"1234567890"}[0],
		CompanyID: &companyID,
		RoleID:    roleID,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Create rows with the expected data
	rows := sqlmock.NewRows([]string{
		"id", "name", "email", "password", "phone", "cpf", "avatar", "role_id", "company_id",
		"active", "last_login", "dashboard_config", "api_token", "login_attempts",
		"blocked_until", "password_changed_at", "created_at", "updated_at",
		"role_id", "role_name", "role_description", "role_created_at", "role_updated_at",
	}).AddRow(
		expectedUser.ID,
		expectedUser.Name,
		expectedUser.Email,
		"hashed_password",
		expectedUser.Phone,
		nil, // cpf
		nil, // avatar
		expectedUser.RoleID,
		expectedUser.CompanyID,
		expectedUser.Active,
		nil,                    // last_login
		nil,                    // dashboard_config
		nil,                    // api_token
		0,                      // login_attempts
		nil,                    // blocked_until
		expectedUser.CreatedAt, // password_changed_at
		expectedUser.CreatedAt,
		expectedUser.UpdatedAt,
		expectedUser.RoleID,    // role_id
		"driver",               // role_name
		"Driver role",          // role_description
		expectedUser.CreatedAt, // role_created_at
		expectedUser.UpdatedAt, // role_updated_at
	)

	// Mock the SELECT query
	expectedQuery := `SELECT u.id, u.name, u.email, u.password, u.phone, u.cpf, u.avatar, u.role_id, u.company_id, u.active, u.last_login, u.dashboard_config, u.api_token, u.login_attempts, u.blocked_until, u.password_changed_at, u.created_at, u.updated_at, r.id, r.name, r.description, r.created_at, r.updated_at FROM users u JOIN roles r ON u.role_id = r.id WHERE u.email = \$1`
	suite.mock.ExpectQuery(expectedQuery).
		WithArgs(email).
		WillReturnRows(rows)

	// Test
	result, err := suite.repo.GetByEmail(ctx, email)

	// Assertions
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedUser.Email, result.Email)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *UserRepositoryTestSuite) TestUpdate_Success() {
	ctx := context.Background()
	userID := uuid.New()

	// Use a pointer to bool for Active field
	active := true
	updateReq := models.UpdateUserRequest{
		Name:   "Updated Name",
		Email:  "updated@example.com",
		Phone:  "9999999999",
		Active: &active,
	}

	// Mock the UPDATE query
	expectedUpdateQuery := `UPDATE users SET name = \$1, email = \$2, phone = \$3, active = \$4, updated_at = \$5 WHERE id = \$6`

	suite.mock.ExpectExec(expectedUpdateQuery).
		WithArgs(
			updateReq.Name,
			updateReq.Email,
			updateReq.Phone,
			updateReq.Active,
			sqlmock.AnyArg(), // updated_at
			userID,
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Mock the SELECT query for the updated user
	companyID := uuid.New()
	roleID := uuid.New()
	rows := sqlmock.NewRows([]string{
		"id", "name", "email", "password", "phone", "cpf", "avatar", "role_id", "company_id",
		"active", "last_login", "dashboard_config", "api_token", "login_attempts",
		"blocked_until", "password_changed_at", "created_at", "updated_at",
		"role_id", "role_name", "role_description", "role_created_at", "role_updated_at",
	}).AddRow(
		userID,
		updateReq.Name,
		updateReq.Email,
		"hashed_password",
		updateReq.Phone,
		"12345678901",
		"",
		roleID,
		companyID,
		*updateReq.Active,
		(*time.Time)(nil),
		"",
		"",
		0,
		(*time.Time)(nil),
		time.Now(), // password_changed_at is not nullable
		time.Now(),
		time.Now(),
		roleID,
		"Admin",
		"Administrator role",
		time.Now(),
		time.Now(),
	)

	expectedSelectQuery := regexp.QuoteMeta("SELECT u.id, u.name, u.email, u.password, u.phone, u.cpf, u.avatar, u.role_id, u.company_id")
	suite.mock.ExpectQuery(expectedSelectQuery).
		WithArgs(userID).
		WillReturnRows(rows)

	// Test
	result, err := suite.repo.Update(ctx, userID, updateReq)

	// Assertions
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), updateReq.Name, result.Name)
	assert.Equal(suite.T(), updateReq.Email, result.Email)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *UserRepositoryTestSuite) TestDelete_Success() {
	ctx := context.Background()
	userID := uuid.New()

	// Mock the UPDATE query (soft delete)
	expectedQuery := `UPDATE users SET active = false, updated_at = \$1 WHERE id = \$2`
	suite.mock.ExpectExec(expectedQuery).
		WithArgs(sqlmock.AnyArg(), userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// Test
	err := suite.repo.Delete(ctx, userID)

	// Assertions
	assert.NoError(suite.T(), err)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func (suite *UserRepositoryTestSuite) TestList_Success() {
	ctx := context.Background()
	limit := 10
	offset := 0
	active := true

	userID1 := uuid.New()
	userID2 := uuid.New()
	companyID := uuid.New()
	roleID := uuid.New()

	// Create rows with all fields matching the actual query
	rows := sqlmock.NewRows([]string{
		"id", "name", "email", "phone", "cpf", "avatar", "role_id", "company_id",
		"active", "last_login", "dashboard_config", "login_attempts",
		"blocked_until", "password_changed_at", "created_at", "updated_at",
		"role_id", "role_name", "role_description", "role_created_at", "role_updated_at",
	}).
		AddRow(userID1, "User 1", "user1@example.com", "1111111111", "11111111111", "", roleID, companyID,
			true, (*time.Time)(nil), "", 0, (*time.Time)(nil), time.Now(), time.Now(), time.Now(),
			roleID, "Admin", "Administrator role", time.Now(), time.Now()).
		AddRow(userID2, "User 2", "user2@example.com", "2222222222", "22222222222", "", roleID, companyID,
			true, (*time.Time)(nil), "", 0, (*time.Time)(nil), time.Now(), time.Now(), time.Now(),
			roleID, "Admin", "Administrator role", time.Now(), time.Now())

	// Mock the SELECT query matching the actual implementation
	expectedQuery := regexp.QuoteMeta("SELECT u.id, u.name, u.email, u.phone, u.cpf, u.avatar, u.role_id, u.company_id")
	suite.mock.ExpectQuery(expectedQuery).
		WithArgs(true, limit, offset).
		WillReturnRows(rows)

	// Test
	result, err := suite.repo.List(ctx, limit, offset, &active, nil)

	// Assertions
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), result, 2)
	assert.Equal(suite.T(), "User 1", result[0].Name)
	assert.Equal(suite.T(), "User 2", result[1].Name)
	assert.NoError(suite.T(), suite.mock.ExpectationsWereMet())
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
