package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/paulochiaradia/dashtrack/internal/models"
)

// CompanyRepositoryInterface defines the contract for company repository
type CompanyRepositoryInterface interface {
	CountCompanies(ctx context.Context) (int, error)
	CountActiveCompanies(ctx context.Context) (int, error)
}

// CompanyRepository handles database operations for companies
type CompanyRepository struct {
	db     *sqlx.DB
	tracer trace.Tracer
}

// NewCompanyRepository creates a new company repository
func NewCompanyRepository(db *sqlx.DB) *CompanyRepository {
	return &CompanyRepository{
		db:     db,
		tracer: otel.Tracer("company-repository"),
	}
}

// Create creates a new company
func (r *CompanyRepository) Create(ctx context.Context, company *models.Company) error {
	ctx, span := r.tracer.Start(ctx, "CompanyRepository.Create",
		trace.WithAttributes(
			attribute.String("company.name", company.Name),
			attribute.String("company.slug", company.Slug),
		))
	defer span.End()

	// Set defaults
	company.ID = uuid.New()
	company.CreatedAt = time.Now()
	company.UpdatedAt = time.Now()
	if company.Status == "" {
		company.Status = "active"
	}
	if company.Country == "" {
		company.Country = "Brazil"
	}

	// Set subscription limits based on plan
	switch company.SubscriptionPlan {
	case "basic":
		company.MaxUsers = 10
		company.MaxVehicles = 5
		company.MaxSensors = 20
	case "premium":
		company.MaxUsers = 50
		company.MaxVehicles = 25
		company.MaxSensors = 100
	case "enterprise":
		company.MaxUsers = 500
		company.MaxVehicles = 100
		company.MaxSensors = 1000
	}

	query := `
		INSERT INTO companies (
			id, name, slug, email, phone, address, city, state, country,
			subscription_plan, max_users, max_vehicles, max_sensors, status,
			created_at, updated_at
		) VALUES (
			:id, :name, :slug, :email, :phone, :address, :city, :state, :country,
			:subscription_plan, :max_users, :max_vehicles, :max_sensors, :status,
			:created_at, :updated_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, company)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create company: %w", err)
	}

	span.SetAttributes(attribute.String("company.id", company.ID.String()))
	return nil
}

// GetByID retrieves a company by ID
func (r *CompanyRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Company, error) {
	ctx, span := r.tracer.Start(ctx, "CompanyRepository.GetByID",
		trace.WithAttributes(attribute.String("company.id", id.String())))
	defer span.End()

	var company models.Company
	query := `
		SELECT id, name, slug, email, phone, address, city, state, country,
			   subscription_plan, max_users, max_vehicles, max_sensors, status,
			   created_at, updated_at
		FROM companies 
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &company, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get company by ID: %w", err)
	}

	return &company, nil
}

// GetBySlug retrieves a company by slug
func (r *CompanyRepository) GetBySlug(ctx context.Context, slug string) (*models.Company, error) {
	ctx, span := r.tracer.Start(ctx, "CompanyRepository.GetBySlug",
		trace.WithAttributes(attribute.String("company.slug", slug)))
	defer span.End()

	var company models.Company
	query := `
		SELECT id, name, slug, email, phone, address, city, state, country,
			   subscription_plan, max_users, max_vehicles, max_sensors, status,
			   created_at, updated_at
		FROM companies 
		WHERE slug = $1
	`

	err := r.db.GetContext(ctx, &company, query, slug)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get company by slug: %w", err)
	}

	return &company, nil
}

// List retrieves all companies with pagination
func (r *CompanyRepository) List(ctx context.Context, limit, offset int) ([]models.Company, error) {
	ctx, span := r.tracer.Start(ctx, "CompanyRepository.List",
		trace.WithAttributes(
			attribute.Int("limit", limit),
			attribute.Int("offset", offset),
		))
	defer span.End()

	var companies []models.Company
	query := `
		SELECT id, name, slug, email, phone, address, city, state, country,
			   subscription_plan, max_users, max_vehicles, max_sensors, status,
			   created_at, updated_at
		FROM companies 
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	err := r.db.SelectContext(ctx, &companies, query, limit, offset)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to list companies: %w", err)
	}

	span.SetAttributes(attribute.Int("companies.count", len(companies)))
	return companies, nil
}

// Update updates a company
func (r *CompanyRepository) Update(ctx context.Context, company *models.Company) error {
	ctx, span := r.tracer.Start(ctx, "CompanyRepository.Update",
		trace.WithAttributes(attribute.String("company.id", company.ID.String())))
	defer span.End()

	company.UpdatedAt = time.Now()

	query := `
		UPDATE companies SET
			name = :name,
			email = :email,
			phone = :phone,
			address = :address,
			city = :city,
			state = :state,
			country = :country,
			subscription_plan = :subscription_plan,
			max_users = :max_users,
			max_vehicles = :max_vehicles,
			max_sensors = :max_sensors,
			status = :status,
			updated_at = :updated_at
		WHERE id = :id
	`

	result, err := r.db.NamedExecContext(ctx, query, company)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update company: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("company not found")
	}

	return nil
}

// Delete soft deletes a company
func (r *CompanyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "CompanyRepository.Delete",
		trace.WithAttributes(attribute.String("company.id", id.String())))
	defer span.End()

	// Soft delete: mark as inactive instead of deleting
	query := `UPDATE companies SET status = 'inactive', updated_at = NOW() WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete company: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("company not found")
	}

	return nil
}

// GetCompanyStats returns statistical data for a company
func (r *CompanyRepository) GetCompanyStats(ctx context.Context, companyID uuid.UUID) (*models.CompanyStats, error) {
	ctx, span := r.tracer.Start(ctx, "CompanyRepository.GetCompanyStats",
		trace.WithAttributes(attribute.String("company.id", companyID.String())))
	defer span.End()

	stats := &models.CompanyStats{}

	// Get vehicle stats
	vehicleQuery := `
		SELECT 
			COUNT(*) as total_vehicles,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_vehicles
		FROM vehicles 
		WHERE company_id = $1
	`
	err := r.db.GetContext(ctx, stats, vehicleQuery, companyID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get vehicle stats: %w", err)
	}

	// Get sensor stats
	sensorQuery := `
		SELECT 
			COUNT(*) as total_sensors,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active_sensors
		FROM sensors 
		WHERE company_id = $1
	`
	var sensorStats struct {
		TotalSensors  int `db:"total_sensors"`
		ActiveSensors int `db:"active_sensors"`
	}
	err = r.db.GetContext(ctx, &sensorStats, sensorQuery, companyID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get sensor stats: %w", err)
	}
	stats.TotalSensors = sensorStats.TotalSensors
	stats.ActiveSensors = sensorStats.ActiveSensors

	// Get user stats
	userQuery := `
		SELECT 
			COUNT(*) as total_users,
			COUNT(CASE WHEN active = true THEN 1 END) as active_users
		FROM users 
		WHERE company_id = $1
	`
	var userStats struct {
		TotalUsers  int `db:"total_users"`
		ActiveUsers int `db:"active_users"`
	}
	err = r.db.GetContext(ctx, &userStats, userQuery, companyID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}
	stats.TotalUsers = userStats.TotalUsers
	stats.ActiveUsers = userStats.ActiveUsers

	// Get today's trip stats
	tripQuery := `
		SELECT 
			COUNT(*) as total_trips_today,
			COALESCE(SUM(distance_km), 0) as total_distance_today
		FROM vehicle_trips vt
		JOIN vehicles v ON vt.vehicle_id = v.id
		WHERE v.company_id = $1 
		AND DATE(vt.start_time) = CURRENT_DATE
	`
	var tripStats struct {
		TotalTripsToday    int     `db:"total_trips_today"`
		TotalDistanceToday float64 `db:"total_distance_today"`
	}
	err = r.db.GetContext(ctx, &tripStats, tripQuery, companyID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get trip stats: %w", err)
	}
	stats.TotalTripsToday = tripStats.TotalTripsToday
	stats.TotalDistanceToday = tripStats.TotalDistanceToday

	// Get alert stats
	alertQuery := `
		SELECT COUNT(*) as active_alerts_count
		FROM sensor_alerts sa
		JOIN sensors s ON sa.sensor_id = s.id
		WHERE s.company_id = $1 
		AND sa.status = 'active'
	`
	err = r.db.GetContext(ctx, &stats.ActiveAlertsCount, alertQuery, companyID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get alert stats: %w", err)
	}

	// Get ESP32 device stats
	esp32Query := `
		SELECT 
			COUNT(CASE WHEN status = 'online' THEN 1 END) as esp32_devices_online,
			COUNT(CASE WHEN status = 'offline' THEN 1 END) as esp32_devices_offline
		FROM esp32_devices 
		WHERE company_id = $1
	`
	var esp32Stats struct {
		ESP32DevicesOnline  int `db:"esp32_devices_online"`
		ESP32DevicesOffline int `db:"esp32_devices_offline"`
	}
	err = r.db.GetContext(ctx, &esp32Stats, esp32Query, companyID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get ESP32 stats: %w", err)
	}
	stats.ESP32DevicesOnline = esp32Stats.ESP32DevicesOnline
	stats.ESP32DevicesOffline = esp32Stats.ESP32DevicesOffline

	return stats, nil
}

// CheckSlugExists checks if a company slug already exists
func (r *CompanyRepository) CheckSlugExists(ctx context.Context, slug string, excludeID *uuid.UUID) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "CompanyRepository.CheckSlugExists",
		trace.WithAttributes(attribute.String("company.slug", slug)))
	defer span.End()

	query := `SELECT COUNT(*) FROM companies WHERE slug = $1`
	args := []interface{}{slug}

	if excludeID != nil {
		query += ` AND id != $2`
		args = append(args, *excludeID)
	}

	var count int
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("failed to check slug existence: %w", err)
	}

	return count > 0, nil
}

// Search searches companies by name, email, or slug
func (r *CompanyRepository) Search(ctx context.Context, searchTerm string, limit, offset int) ([]models.Company, error) {
	ctx, span := r.tracer.Start(ctx, "CompanyRepository.Search",
		trace.WithAttributes(
			attribute.String("search_term", searchTerm),
			attribute.Int("limit", limit),
			attribute.Int("offset", offset),
		))
	defer span.End()

	var companies []models.Company
	searchPattern := "%" + strings.ToLower(searchTerm) + "%"

	query := `
		SELECT id, name, slug, email, phone, address, city, state, country,
			   subscription_plan, max_users, max_vehicles, max_sensors, status,
			   created_at, updated_at
		FROM companies 
		WHERE (LOWER(name) LIKE $1 OR LOWER(email) LIKE $1 OR LOWER(slug) LIKE $1)
		AND status != 'deleted'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	err := r.db.SelectContext(ctx, &companies, query, searchPattern, limit, offset)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to search companies: %w", err)
	}

	span.SetAttributes(attribute.Int("companies.count", len(companies)))
	return companies, nil
}

// CountCompanies counts total companies
func (r *CompanyRepository) CountCompanies(ctx context.Context) (int, error) {
	ctx, span := r.tracer.Start(ctx, "CompanyRepository.CountCompanies")
	defer span.End()

	query := "SELECT COUNT(*) FROM companies"
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to count companies: %w", err)
	}

	return count, nil
}

// CountActiveCompanies counts active companies
func (r *CompanyRepository) CountActiveCompanies(ctx context.Context) (int, error) {
	ctx, span := r.tracer.Start(ctx, "CompanyRepository.CountActiveCompanies")
	defer span.End()

	// Temporarily count all companies since active column might not exist
	query := "SELECT COUNT(*) FROM companies"
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		span.RecordError(err)
		return 0, fmt.Errorf("failed to count active companies: %w", err)
	}

	return count, nil
}
