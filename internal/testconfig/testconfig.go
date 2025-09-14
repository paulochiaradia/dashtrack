package testconfig

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rubenv/sql-migrate"
)

// TestConfig holds the test configuration
type TestConfig struct {
	DatabaseURL string
}

// NewTestConfig creates a new test configuration
func NewTestConfig() *TestConfig {
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://test:test@localhost:5433/test_dashtrack?sslmode=disable"
	}

	return &TestConfig{
		DatabaseURL: dbURL,
	}
}

// SetupTestDatabase creates a test database and runs migrations
func SetupTestDatabase(t *testing.T) (*sql.DB, func()) {
	config := NewTestConfig()
	
	db, err := sql.Open("pgx", config.DatabaseURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Run migrations
	migrations := &migrate.FileMigrationSource{
		Dir: "../../internal/database/migrations",
	}

	n, err := migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	if n > 0 {
		fmt.Printf("Applied %d migrations to test database\n", n)
	}

	// Return cleanup function
	cleanup := func() {
		// Rollback migrations
		migrate.Exec(db, "postgres", migrations, migrate.Down)
		db.Close()
	}

	return db, cleanup
}

// TruncateTables truncates all tables for clean test state
func TruncateTables(db *sql.DB) error {
	tables := []string{"auth_logs", "user_sessions", "users", "roles"}
	
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %v", table, err)
		}
	}
	
	return nil
}

// SeedTestData inserts test data for testing
func SeedTestData(db *sql.DB) error {
	// Insert test roles
	_, err := db.Exec(`
		INSERT INTO roles (id, name, description) VALUES 
		('550e8400-e29b-41d4-a716-446655440001', 'admin', 'Administrator'),
		('550e8400-e29b-41d4-a716-446655440002', 'driver', 'Driver'),
		('550e8400-e29b-41d4-a716-446655440003', 'helper', 'Helper')
	`)
	if err != nil {
		return fmt.Errorf("failed to seed roles: %v", err)
	}

	// Insert test users
	_, err = db.Exec(`
		INSERT INTO users (id, name, email, password, phone, role_id, active) VALUES 
		('550e8400-e29b-41d4-a716-446655440011', 'Test Admin', 'admin@test.com', 'hashedpassword', '123456789', '550e8400-e29b-41d4-a716-446655440001', true),
		('550e8400-e29b-41d4-a716-446655440012', 'Test Driver', 'driver@test.com', 'hashedpassword', '987654321', '550e8400-e29b-41d4-a716-446655440002', true)
	`)
	if err != nil {
		return fmt.Errorf("failed to seed users: %v", err)
	}

	return nil
}
