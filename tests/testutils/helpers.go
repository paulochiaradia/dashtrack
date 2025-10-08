package testutils

import (
	"context"
	"database/sql"
	"time"

	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/services"
)

// TestDatabase provides utilities for test database operations
type TestDatabase struct {
	DB *sqlx.DB
	t  *testing.T
}

// NewTestDatabase creates a new test database helper
func NewTestDatabase(db *sqlx.DB, t *testing.T) *TestDatabase {
	return &TestDatabase{
		DB: db,
		t:  t,
	}
}

// CreateTestCompany creates a test company and returns its ID
func (td *TestDatabase) CreateTestCompany(name string) uuid.UUID {
	companyID := uuid.New()
	ctx := context.Background()

	query := `INSERT INTO companies (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`
	_, err := td.DB.ExecContext(ctx, query, companyID, name, time.Now(), time.Now())
	require.NoError(td.t, err, "Failed to create test company")

	return companyID
}

// CreateTestRole creates a test role and returns its ID
func (td *TestDatabase) CreateTestRole(name string) uuid.UUID {
	roleID := uuid.New()
	ctx := context.Background()

	query := `INSERT INTO roles (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4) ON CONFLICT (name) DO NOTHING`
	_, err := td.DB.ExecContext(ctx, query, roleID, name, time.Now(), time.Now())
	require.NoError(td.t, err, "Failed to create test role")

	// Get the actual role ID (in case it already existed)
	err = td.DB.Get(&roleID, "SELECT id FROM roles WHERE name = $1", name)
	require.NoError(td.t, err, "Failed to get role ID")

	return roleID
}

// CreateTestUser creates a test user and returns the user object
func (td *TestDatabase) CreateTestUser(name, email, roleName string, companyID *uuid.UUID) *models.User {
	userID := uuid.New()
	roleID := td.CreateTestRole(roleName)
	ctx := context.Background()

	query := `INSERT INTO users (id, name, email, password, role_id, company_id, active, created_at, updated_at)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := td.DB.ExecContext(ctx, query,
		userID, name, email, "$2a$10$test.hash", roleID, companyID,
		true, time.Now(), time.Now())
	require.NoError(td.t, err, "Failed to create test user")

	return &models.User{
		ID:        userID,
		Name:      name,
		Email:     email,
		Password:  "$2a$10$test.hash",
		Active:    true,
		CompanyID: companyID,
		Role: &models.Role{
			ID:   roleID,
			Name: roleName,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// CleanupTestData removes test data based on email pattern
func (td *TestDatabase) CleanupTestData(emailPattern string) {
	ctx := context.Background()

	// Delete users with matching email pattern
	_, _ = td.DB.ExecContext(ctx, "DELETE FROM users WHERE email LIKE $1", emailPattern)

	// Delete test companies
	_, _ = td.DB.ExecContext(ctx, "DELETE FROM companies WHERE name LIKE 'Test%' OR name LIKE 'E2E%'")
}

// GetUserByID retrieves a user by ID for testing
func (td *TestDatabase) GetUserByID(userID uuid.UUID) (*models.User, error) {
	ctx := context.Background()

	query := `SELECT u.id, u.name, u.email, u.password, u.active, u.created_at, u.updated_at,
			  r.id as role_id, r.name as role_name,
			  c.id as company_id, c.name as company_name
			  FROM users u
			  LEFT JOIN roles r ON u.role_id = r.id
			  LEFT JOIN companies c ON u.company_id = c.id
			  WHERE u.id = $1 AND u.deleted_at IS NULL`

	var user models.User
	var role models.Role
	var companyID sql.NullString
	var companyName sql.NullString

	err := td.DB.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.Name, &user.Email, &user.Password,
		&user.Active, &user.CreatedAt, &user.UpdatedAt,
		&role.ID, &role.Name,
		&companyID, &companyName,
	)

	if err != nil {
		return nil, err
	}

	user.Role = &role
	if companyID.Valid {
		companyUUID, _ := uuid.Parse(companyID.String)
		user.CompanyID = &companyUUID
	}

	return &user, nil
}

// AssertUserExists verifies that a user exists in the database
func (td *TestDatabase) AssertUserExists(userID uuid.UUID) {
	user, err := td.GetUserByID(userID)
	assert.NoError(td.t, err, "User should exist")
	assert.NotNil(td.t, user, "User should not be nil")
}

// AssertUserDeleted verifies that a user is soft deleted
func (td *TestDatabase) AssertUserDeleted(userID uuid.UUID) {
	ctx := context.Background()
	var deletedAt sql.NullTime

	query := `SELECT deleted_at FROM users WHERE id = $1`
	err := td.DB.QueryRowContext(ctx, query, userID).Scan(&deletedAt)

	assert.NoError(td.t, err, "Should be able to query user")
	assert.True(td.t, deletedAt.Valid, "User should be soft deleted")
	assert.False(td.t, deletedAt.Time.IsZero(), "Deleted_at should not be zero")
}

// GetUserCount returns the number of active users
func (td *TestDatabase) GetUserCount() int {
	ctx := context.Background()
	var count int

	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`
	err := td.DB.QueryRowContext(ctx, query).Scan(&count)
	require.NoError(td.t, err, "Failed to get user count")

	return count
}

// GetUserCountByCompany returns the number of active users in a company
func (td *TestDatabase) GetUserCountByCompany(companyID uuid.UUID) int {
	ctx := context.Background()
	var count int

	query := `SELECT COUNT(*) FROM users WHERE company_id = $1 AND deleted_at IS NULL`
	err := td.DB.QueryRowContext(ctx, query, companyID).Scan(&count)
	require.NoError(td.t, err, "Failed to get user count by company")

	return count
}

// TestDataBuilder provides a fluent interface for creating test data
type TestDataBuilder struct {
	td        *TestDatabase
	users     []*models.User
	companies []uuid.UUID
	roles     []uuid.UUID
}

// NewTestDataBuilder creates a new test data builder
func NewTestDataBuilder(td *TestDatabase) *TestDataBuilder {
	return &TestDataBuilder{
		td:        td,
		users:     make([]*models.User, 0),
		companies: make([]uuid.UUID, 0),
		roles:     make([]uuid.UUID, 0),
	}
}

// WithCompany adds a company to the test data
func (tdb *TestDataBuilder) WithCompany(name string) *TestDataBuilder {
	companyID := tdb.td.CreateTestCompany(name)
	tdb.companies = append(tdb.companies, companyID)
	return tdb
}

// WithRole adds a role to the test data
func (tdb *TestDataBuilder) WithRole(name string) *TestDataBuilder {
	roleID := tdb.td.CreateTestRole(name)
	tdb.roles = append(tdb.roles, roleID)
	return tdb
}

// WithUser adds a user to the test data
func (tdb *TestDataBuilder) WithUser(name, email, role string, companyIndex int) *TestDataBuilder {
	var companyID *uuid.UUID
	if companyIndex >= 0 && companyIndex < len(tdb.companies) {
		companyID = &tdb.companies[companyIndex]
	}

	user := tdb.td.CreateTestUser(name, email, role, companyID)
	tdb.users = append(tdb.users, user)
	return tdb
}

// Build returns the created test data
func (tdb *TestDataBuilder) Build() *TestDataSet {
	return &TestDataSet{
		Users:     tdb.users,
		Companies: tdb.companies,
		Roles:     tdb.roles,
	}
}

// TestDataSet contains created test data
type TestDataSet struct {
	Users     []*models.User
	Companies []uuid.UUID
	Roles     []uuid.UUID
}

// GetUser returns a user by index
func (tds *TestDataSet) GetUser(index int) *models.User {
	if index >= 0 && index < len(tds.Users) {
		return tds.Users[index]
	}
	return nil
}

// GetCompany returns a company ID by index
func (tds *TestDataSet) GetCompany(index int) uuid.UUID {
	if index >= 0 && index < len(tds.Companies) {
		return tds.Companies[index]
	}
	return uuid.Nil
}

// GetUserByRole returns the first user with the specified role
func (tds *TestDataSet) GetUserByRole(role string) *models.User {
	for _, user := range tds.Users {
		if user.Role != nil && user.Role.Name == role {
			return user
		}
	}
	return nil
}

// GetUsersByRole returns all users with the specified role
func (tds *TestDataSet) GetUsersByRole(role string) []*models.User {
	var result []*models.User
	for _, user := range tds.Users {
		if user.Role != nil && user.Role.Name == role {
			result = append(result, user)
		}
	}
	return result
}

// AssertionHelpers provides common assertion helpers for tests
type AssertionHelpers struct {
	t *testing.T
}

// NewAssertionHelpers creates a new assertion helpers instance
func NewAssertionHelpers(t *testing.T) *AssertionHelpers {
	return &AssertionHelpers{t: t}
}

// AssertValidUser validates that a user object has expected values
func (ah *AssertionHelpers) AssertValidUser(user *models.User) {
	assert.NotNil(ah.t, user, "User should not be nil")
	assert.NotEqual(ah.t, uuid.Nil, user.ID, "User ID should not be nil")
	assert.NotEmpty(ah.t, user.Name, "User name should not be empty")
	assert.NotEmpty(ah.t, user.Email, "User email should not be empty")
	assert.NotNil(ah.t, user.Role, "User role should not be nil")
	assert.NotEmpty(ah.t, user.Role.Name, "Role name should not be empty")
}

// AssertUserBelongsToCompany validates that a user belongs to the specified company
func (ah *AssertionHelpers) AssertUserBelongsToCompany(user *models.User, companyID uuid.UUID) {
	assert.NotNil(ah.t, user.CompanyID, "User should have a company ID")
	assert.Equal(ah.t, companyID, *user.CompanyID, "User should belong to the specified company")
}

// AssertUserHasRole validates that a user has the specified role
func (ah *AssertionHelpers) AssertUserHasRole(user *models.User, expectedRole string) {
	assert.NotNil(ah.t, user.Role, "User should have a role")
	assert.Equal(ah.t, expectedRole, user.Role.Name, "User should have the expected role")
}

// AssertPaginatedResponse validates a paginated response structure
func (ah *AssertionHelpers) AssertPaginatedResponse(response *services.UserListResponse, expectedTotal int) {
	assert.NotNil(ah.t, response, "Response should not be nil")
	assert.Equal(ah.t, expectedTotal, response.Total, "Total count should match expected")
	assert.LessOrEqual(ah.t, len(response.Users), response.Limit, "Users count should not exceed limit")
	assert.GreaterOrEqual(ah.t, response.Page, 1, "Page should be at least 1")
	assert.GreaterOrEqual(ah.t, response.TotalPages, 1, "Total pages should be at least 1")
}

// TimeHelpers provides utilities for time-related testing
type TimeHelpers struct{}

// NewTimeHelpers creates a new time helpers instance
func NewTimeHelpers() *TimeHelpers {
	return &TimeHelpers{}
}

// IsRecent checks if a time is within the last few seconds (useful for created_at/updated_at)
func (th *TimeHelpers) IsRecent(t time.Time, withinSeconds int) bool {
	return time.Since(t) <= time.Duration(withinSeconds)*time.Second
}

// AssertTimeIsRecent asserts that a time is recent
func (th *TimeHelpers) AssertTimeIsRecent(t *testing.T, timestamp time.Time, message string) {
	assert.True(t, th.IsRecent(timestamp, 5), message)
}

// StringPtr returns a pointer to a string (helper for optional fields)
func StringPtr(s string) *string {
	return &s
}

// UUIDPtr returns a pointer to a UUID (helper for optional fields)
func UUIDPtr(id uuid.UUID) *uuid.UUID {
	return &id
}

// BoolPtr returns a pointer to a bool (helper for optional fields)
func BoolPtr(b bool) *bool {
	return &b
}

// TestLogger provides structured logging for tests
type TestLogger struct {
	t *testing.T
}

// NewTestLogger creates a new test logger
func NewTestLogger(t *testing.T) *TestLogger {
	return &TestLogger{t: t}
}

// Info logs an info message
func (tl *TestLogger) Info(msg string, args ...interface{}) {
	tl.t.Logf("INFO: "+msg, args...)
}

// Error logs an error message
func (tl *TestLogger) Error(msg string, args ...interface{}) {
	tl.t.Logf("ERROR: "+msg, args...)
}

// Debug logs a debug message
func (tl *TestLogger) Debug(msg string, args ...interface{}) {
	tl.t.Logf("DEBUG: "+msg, args...)
}

// Step logs a test step
func (tl *TestLogger) Step(step int, description string, args ...interface{}) {
	tl.t.Logf("STEP %d: "+description, append([]interface{}{step}, args...)...)
}

// Success logs a success message
func (tl *TestLogger) Success(msg string, args ...interface{}) {
	tl.t.Logf("✅ SUCCESS: "+msg, args...)
}

// Failure logs a failure message
func (tl *TestLogger) Failure(msg string, args ...interface{}) {
	tl.t.Logf("❌ FAILURE: "+msg, args...)
}
