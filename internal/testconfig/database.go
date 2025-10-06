package testconfig

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	pgMigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type TestDB struct {
	DB       *gorm.DB
	SqlDB    *sql.DB
	TestName string
}

// SetupTestDB creates a test database for integration testing
func SetupTestDB(testName string) (*TestDB, error) {
	// Use test database configuration
	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5432")
	user := getEnv("TEST_DB_USER", "dashtrack_user")
	password := getEnv("TEST_DB_PASSWORD", "dashtrack_password")
	dbname := fmt.Sprintf("dashtrack_test_%s", testName)

	// Create database if it doesn't exist
	if err := createTestDatabase(host, port, user, password, dbname); err != nil {
		return nil, fmt.Errorf("failed to create test database: %v", err)
	}

	// Connect to the test database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=UTC",
		host, port, user, password, dbname)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %v", err)
	}

	testDB := &TestDB{
		DB:       db,
		SqlDB:    sqlDB,
		TestName: testName,
	}

	// Run migrations
	if err := testDB.RunMigrations(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	return testDB, nil
}

// createTestDatabase creates a new test database
func createTestDatabase(host, port, user, password, dbname string) error {
	// Connect to postgres database to create test database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		host, port, user, password)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	// Drop database if exists
	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbname))
	if err != nil {
		return err
	}

	// Create new database
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
	return err
}

// RunMigrations runs database migrations for testing
func (tdb *TestDB) RunMigrations() error {
	driver, err := pgMigrate.WithInstance(tdb.SqlDB, &pgMigrate.Config{})
	if err != nil {
		return err
	}

	// Assuming migrations are in the migrations directory
	migrationsPath := "file://../../migrations"
	m, err := migrate.NewWithDatabaseInstance(migrationsPath, "postgres", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

// SeedTestData seeds the database with test data
func (tdb *TestDB) SeedTestData() error {
	// Insert test roles
	roles := []map[string]interface{}{
		{"id": "f47ac10b-58cc-4372-a567-0e02b2c3d479", "name": "master", "description": "Master role with full access"},
		{"id": "f47ac10b-58cc-4372-a567-0e02b2c3d480", "name": "company_admin", "description": "Company administrator"},
		{"id": "f47ac10b-58cc-4372-a567-0e02b2c3d481", "name": "driver", "description": "Vehicle driver"},
		{"id": "f47ac10b-58cc-4372-a567-0e02b2c3d482", "name": "helper", "description": "Driver helper"},
	}

	for _, role := range roles {
		err := tdb.DB.Exec("INSERT INTO roles (id, name, description, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW()) ON CONFLICT (id) DO NOTHING",
			role["id"], role["name"], role["description"]).Error
		if err != nil {
			return err
		}
	}

	// Insert test companies
	companies := []map[string]interface{}{
		{
			"id":                "c47ac10b-58cc-4372-a567-0e02b2c3d479",
			"name":              "Test Company 1",
			"slug":              "test-company-1",
			"email":             "contact@testcompany1.com",
			"country":           "Brazil",
			"subscription_plan": "premium",
			"status":            "active",
		},
		{
			"id":                "c47ac10b-58cc-4372-a567-0e02b2c3d480",
			"name":              "Test Company 2",
			"slug":              "test-company-2",
			"email":             "contact@testcompany2.com",
			"country":           "Brazil",
			"subscription_plan": "basic",
			"status":            "active",
		},
	}

	for _, company := range companies {
		err := tdb.DB.Exec(`INSERT INTO companies (id, name, slug, email, country, subscription_plan, status, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?, ?, NOW(), NOW()) ON CONFLICT (id) DO NOTHING`,
			company["id"], company["name"], company["slug"], company["email"],
			company["country"], company["subscription_plan"], company["status"]).Error
		if err != nil {
			return err
		}
	}

	// Insert test users
	users := []map[string]interface{}{
		{
			"id":         "u47ac10b-58cc-4372-a567-0e02b2c3d479",
			"name":       "Master User",
			"email":      "master@dashtrack.com",
			"password":   "$2a$10$N9qo8uLOickgx2ZMRZoMye.xjW6sBhOdOjq.8IHdgeXd7.dqW5WKO", // hashed "password123"
			"role_id":    "f47ac10b-58cc-4372-a567-0e02b2c3d479",                         // master role
			"company_id": nil,
		},
		{
			"id":         "u47ac10b-58cc-4372-a567-0e02b2c3d480",
			"name":       "Company 1 Admin",
			"email":      "admin@company1.com",
			"password":   "$2a$10$N9qo8uLOickgx2ZMRZoMye.xjW6sBhOdOjq.8IHdgeXd7.dqW5WKO",
			"role_id":    "f47ac10b-58cc-4372-a567-0e02b2c3d480", // company_admin role
			"company_id": "c47ac10b-58cc-4372-a567-0e02b2c3d479", // test company 1
		},
		{
			"id":         "u47ac10b-58cc-4372-a567-0e02b2c3d481",
			"name":       "Driver User",
			"email":      "driver@company1.com",
			"password":   "$2a$10$N9qo8uLOickgx2ZMRZoMye.xjW6sBhOdOjq.8IHdgeXd7.dqW5WKO",
			"role_id":    "f47ac10b-58cc-4372-a567-0e02b2c3d481", // driver role
			"company_id": "c47ac10b-58cc-4372-a567-0e02b2c3d479", // test company 1
		},
		{
			"id":         "u47ac10b-58cc-4372-a567-0e02b2c3d482",
			"name":       "Helper User",
			"email":      "helper@company1.com",
			"password":   "$2a$10$N9qo8uLOickgx2ZMRZoMye.xjW6sBhOdOjq.8IHdgeXd7.dqW5WKO",
			"role_id":    "f47ac10b-58cc-4372-a567-0e02b2c3d482", // helper role
			"company_id": "c47ac10b-58cc-4372-a567-0e02b2c3d479", // test company 1
		},
	}

	for _, user := range users {
		err := tdb.DB.Exec(`INSERT INTO users (id, name, email, password, role_id, company_id, created_at, updated_at) 
			VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW()) ON CONFLICT (id) DO NOTHING`,
			user["id"], user["name"], user["email"], user["password"],
			user["role_id"], user["company_id"]).Error
		if err != nil {
			return err
		}
	}

	return nil
}

// CleanupTestData removes test data from database
func (tdb *TestDB) CleanupTestData() error {
	tables := []string{
		"auth_logs", "esp32_devices", "vehicle_assignments", "vehicles",
		"users", "companies", "roles",
	}

	for _, table := range tables {
		err := tdb.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error
		if err != nil {
			log.Printf("Warning: failed to truncate table %s: %v", table, err)
		}
	}

	return nil
}

// TearDown closes the database connection and drops the test database
func (tdb *TestDB) TearDown() error {
	dbname := fmt.Sprintf("dashtrack_test_%s", tdb.TestName)

	// Close the connection
	if err := tdb.SqlDB.Close(); err != nil {
		return err
	}

	// Drop the test database
	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5432")
	user := getEnv("TEST_DB_USER", "dashtrack_user")
	password := getEnv("TEST_DB_PASSWORD", "dashtrack_password")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=disable",
		host, port, user, password)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbname))
	return err
}

// TestMain runs setup and teardown for all tests in a package
func TestMain(m *testing.M) {
	// Setup code here if needed
	code := m.Run()
	// Teardown code here if needed
	os.Exit(code)
}

// Helper function to get environment variables with default values
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
