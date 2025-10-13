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

// VehicleRepository handles database operations for vehicles
type VehicleRepository struct {
	db     *sqlx.DB
	tracer trace.Tracer
}

// NewVehicleRepository creates a new vehicle repository
func NewVehicleRepository(db *sqlx.DB) *VehicleRepository {
	return &VehicleRepository{
		db:     db,
		tracer: otel.Tracer("vehicle-repository"),
	}
}

// Create creates a new vehicle
func (r *VehicleRepository) Create(ctx context.Context, vehicle *models.Vehicle) error {
	ctx, span := r.tracer.Start(ctx, "VehicleRepository.Create",
		trace.WithAttributes(
			attribute.String("vehicle.license_plate", vehicle.LicensePlate),
			attribute.String("company.id", vehicle.CompanyID.String()),
		))
	defer span.End()

	vehicle.ID = uuid.New()
	vehicle.CreatedAt = time.Now()
	vehicle.UpdatedAt = time.Now()
	if vehicle.Status == "" {
		vehicle.Status = "active"
	}

	query := `
		INSERT INTO vehicles (
			id, company_id, team_id, license_plate, brand, model, year, color,
			vehicle_type, fuel_type, cargo_capacity, driver_id, helper_id, status,
			created_at, updated_at
		) VALUES (
			:id, :company_id, :team_id, :license_plate, :brand, :model, :year, :color,
			:vehicle_type, :fuel_type, :cargo_capacity, :driver_id, :helper_id, :status,
			:created_at, :updated_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, vehicle)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create vehicle: %w", err)
	}

	span.SetAttributes(attribute.String("vehicle.id", vehicle.ID.String()))
	return nil
}

// GetByID retrieves a vehicle by ID with company context
func (r *VehicleRepository) GetByID(ctx context.Context, id uuid.UUID, companyID uuid.UUID) (*models.Vehicle, error) {
	ctx, span := r.tracer.Start(ctx, "VehicleRepository.GetByID",
		trace.WithAttributes(
			attribute.String("vehicle.id", id.String()),
			attribute.String("company.id", companyID.String()),
		))
	defer span.End()

	var vehicle models.Vehicle
	query := `
		SELECT id, company_id, team_id, license_plate, brand, model, year, color,
			   vehicle_type, fuel_type, cargo_capacity, driver_id, helper_id, status,
			   created_at, updated_at
		FROM vehicles 
		WHERE id = $1 AND company_id = $2
	`

	err := r.db.GetContext(ctx, &vehicle, query, id, companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get vehicle by ID: %w", err)
	}

	return &vehicle, nil
}

// GetByLicensePlate retrieves a vehicle by license plate within company
func (r *VehicleRepository) GetByLicensePlate(ctx context.Context, licensePlate string, companyID uuid.UUID) (*models.Vehicle, error) {
	ctx, span := r.tracer.Start(ctx, "VehicleRepository.GetByLicensePlate",
		trace.WithAttributes(
			attribute.String("vehicle.license_plate", licensePlate),
			attribute.String("company.id", companyID.String()),
		))
	defer span.End()

	var vehicle models.Vehicle
	query := `
		SELECT id, company_id, team_id, license_plate, brand, model, year, color,
			   vehicle_type, fuel_type, cargo_capacity, driver_id, helper_id, status,
			   created_at, updated_at
		FROM vehicles 
		WHERE license_plate = $1 AND company_id = $2
	`

	err := r.db.GetContext(ctx, &vehicle, query, licensePlate, companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get vehicle by license plate: %w", err)
	}

	return &vehicle, nil
}

// GetByCompany retrieves all vehicles for a company
func (r *VehicleRepository) GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]models.Vehicle, error) {
	ctx, span := r.tracer.Start(ctx, "VehicleRepository.GetByCompany",
		trace.WithAttributes(
			attribute.String("company.id", companyID.String()),
			attribute.Int("limit", limit),
			attribute.Int("offset", offset),
		))
	defer span.End()

	var vehicles []models.Vehicle
	query := `
		SELECT id, company_id, team_id, license_plate, brand, model, year, color,
			   vehicle_type, fuel_type, cargo_capacity, driver_id, helper_id, status,
			   created_at, updated_at
		FROM vehicles 
		WHERE company_id = $1 AND status != 'deleted'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	err := r.db.SelectContext(ctx, &vehicles, query, companyID, limit, offset)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get vehicles by company: %w", err)
	}

	span.SetAttributes(attribute.Int("vehicles.count", len(vehicles)))
	return vehicles, nil
}

// GetByTeam retrieves all vehicles for a team
func (r *VehicleRepository) GetByTeam(ctx context.Context, teamID uuid.UUID, companyID uuid.UUID) ([]models.Vehicle, error) {
	ctx, span := r.tracer.Start(ctx, "VehicleRepository.GetByTeam",
		trace.WithAttributes(
			attribute.String("team.id", teamID.String()),
			attribute.String("company.id", companyID.String()),
		))
	defer span.End()

	var vehicles []models.Vehicle
	query := `
		SELECT id, company_id, team_id, license_plate, brand, model, year, color,
			   vehicle_type, fuel_type, cargo_capacity, driver_id, helper_id, status,
			   created_at, updated_at
		FROM vehicles 
		WHERE team_id = $1 AND company_id = $2 AND status != 'deleted'
		ORDER BY license_plate ASC
	`

	err := r.db.SelectContext(ctx, &vehicles, query, teamID, companyID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get vehicles by team: %w", err)
	}

	span.SetAttributes(attribute.Int("vehicles.count", len(vehicles)))
	return vehicles, nil
}

// GetByDriver retrieves vehicles assigned to a driver
func (r *VehicleRepository) GetByDriver(ctx context.Context, driverID uuid.UUID, companyID uuid.UUID) ([]models.Vehicle, error) {
	ctx, span := r.tracer.Start(ctx, "VehicleRepository.GetByDriver",
		trace.WithAttributes(
			attribute.String("driver.id", driverID.String()),
			attribute.String("company.id", companyID.String()),
		))
	defer span.End()

	var vehicles []models.Vehicle
	query := `
		SELECT id, company_id, team_id, license_plate, brand, model, year, color,
			   vehicle_type, fuel_type, cargo_capacity, driver_id, helper_id, status,
			   created_at, updated_at
		FROM vehicles 
		WHERE driver_id = $1 AND company_id = $2 AND status != 'deleted'
		ORDER BY license_plate ASC
	`

	err := r.db.SelectContext(ctx, &vehicles, query, driverID, companyID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get vehicles by driver: %w", err)
	}

	span.SetAttributes(attribute.Int("vehicles.count", len(vehicles)))
	return vehicles, nil
}

// Update updates a vehicle
func (r *VehicleRepository) Update(ctx context.Context, vehicle *models.Vehicle) error {
	ctx, span := r.tracer.Start(ctx, "VehicleRepository.Update",
		trace.WithAttributes(attribute.String("vehicle.id", vehicle.ID.String())))
	defer span.End()

	vehicle.UpdatedAt = time.Now()

	query := `
		UPDATE vehicles SET
			team_id = :team_id,
			license_plate = :license_plate,
			brand = :brand,
			model = :model,
			year = :year,
			color = :color,
			vehicle_type = :vehicle_type,
			fuel_type = :fuel_type,
			cargo_capacity = :cargo_capacity,
			driver_id = :driver_id,
			helper_id = :helper_id,
			status = :status,
			updated_at = :updated_at
		WHERE id = :id AND company_id = :company_id
	`

	result, err := r.db.NamedExecContext(ctx, query, vehicle)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update vehicle: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("vehicle not found or not authorized")
	}

	return nil
}

// UpdateAssignment updates vehicle assignments (driver, helper, team)
func (r *VehicleRepository) UpdateAssignment(ctx context.Context, vehicleID, companyID uuid.UUID, driverID, helperID, teamID *uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "VehicleRepository.UpdateAssignment",
		trace.WithAttributes(
			attribute.String("vehicle.id", vehicleID.String()),
			attribute.String("company.id", companyID.String()),
		))
	defer span.End()

	query := `
		UPDATE vehicles SET
			driver_id = $1,
			helper_id = $2,
			team_id = $3,
			updated_at = NOW()
		WHERE id = $4 AND company_id = $5
	`

	result, err := r.db.ExecContext(ctx, query, driverID, helperID, teamID, vehicleID, companyID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update vehicle assignment: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("vehicle not found or not authorized")
	}

	return nil
}

// Delete soft deletes a vehicle
func (r *VehicleRepository) Delete(ctx context.Context, id uuid.UUID, companyID uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "VehicleRepository.Delete",
		trace.WithAttributes(
			attribute.String("vehicle.id", id.String()),
			attribute.String("company.id", companyID.String()),
		))
	defer span.End()

	query := `
		UPDATE vehicles 
		SET deleted_at = NOW(), updated_at = NOW() 
		WHERE id = $1 AND company_id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, id, companyID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete vehicle: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("vehicle not found or not authorized")
	}

	return nil
}

// GetVehicleDashboardData retrieves comprehensive dashboard data for a vehicle
func (r *VehicleRepository) GetVehicleDashboardData(ctx context.Context, vehicleID, companyID uuid.UUID) (*models.VehicleDashboardData, error) {
	ctx, span := r.tracer.Start(ctx, "VehicleRepository.GetVehicleDashboardData",
		trace.WithAttributes(
			attribute.String("vehicle.id", vehicleID.String()),
			attribute.String("company.id", companyID.String()),
		))
	defer span.End()

	// Get vehicle basic info
	vehicle, err := r.GetByID(ctx, vehicleID, companyID)
	if err != nil {
		return nil, err
	}
	if vehicle == nil {
		return nil, fmt.Errorf("vehicle not found")
	}

	data := &models.VehicleDashboardData{
		Vehicle:          *vehicle,
		LatestSensorData: make(map[string]interface{}),
	}

	// Get today's statistics
	statsQuery := `
		SELECT 
			COUNT(*) as total_trips,
			COALESCE(SUM(distance_km), 0) as total_distance_km,
			COALESCE(SUM(duration_minutes), 0) as total_duration_minutes,
			COALESCE(SUM(fuel_consumption), 0) as fuel_consumption
		FROM vehicle_trips 
		WHERE vehicle_id = $1 AND DATE(start_time) = CURRENT_DATE
	`

	var stats struct {
		TotalTrips         int     `db:"total_trips"`
		TotalDistanceKm    float64 `db:"total_distance_km"`
		TotalDurationHours float64 `db:"total_duration_minutes"`
		FuelConsumption    float64 `db:"fuel_consumption"`
	}

	err = r.db.GetContext(ctx, &stats, statsQuery, vehicleID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get vehicle stats: %w", err)
	}

	data.TodayStats = models.VehicleDailyStats{
		TotalTrips:         stats.TotalTrips,
		TotalDistanceKm:    stats.TotalDistanceKm,
		TotalDurationHours: stats.TotalDurationHours / 60, // Convert minutes to hours
		FuelConsumption:    stats.FuelConsumption,
	}

	// Calculate average speed
	if data.TodayStats.TotalDurationHours > 0 {
		data.TodayStats.AverageSpeed = data.TodayStats.TotalDistanceKm / data.TodayStats.TotalDurationHours
	}

	// Get active trip
	activeTrip, err := r.GetActiveTrip(ctx, vehicleID)
	if err != nil {
		span.RecordError(err)
	} else {
		data.ActiveTrip = activeTrip
	}

	// Get alerts count
	alertsQuery := `
		SELECT COUNT(*) 
		FROM sensor_alerts sa
		JOIN sensors s ON sa.sensor_id = s.id
		WHERE s.vehicle_id = $1 AND sa.status = 'active'
	`

	err = r.db.GetContext(ctx, &data.TodayStats.AlertsCount, alertsQuery, vehicleID)
	if err != nil {
		span.RecordError(err)
	}

	return data, nil
}

// GetActiveTrip retrieves the currently active trip for a vehicle
func (r *VehicleRepository) GetActiveTrip(ctx context.Context, vehicleID uuid.UUID) (*models.VehicleTrip, error) {
	ctx, span := r.tracer.Start(ctx, "VehicleRepository.GetActiveTrip",
		trace.WithAttributes(attribute.String("vehicle.id", vehicleID.String())))
	defer span.End()

	var trip models.VehicleTrip
	query := `
		SELECT id, vehicle_id, driver_id, helper_id, start_location, end_location,
			   start_latitude, start_longitude, end_latitude, end_longitude,
			   start_time, end_time, distance_km, duration_minutes, fuel_consumption,
			   status, notes, created_at, updated_at
		FROM vehicle_trips 
		WHERE vehicle_id = $1 AND status = 'active'
		ORDER BY start_time DESC
		LIMIT 1
	`

	err := r.db.GetContext(ctx, &trip, query, vehicleID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get active trip: %w", err)
	}

	return &trip, nil
}

// Search searches vehicles by license plate, brand, or model
func (r *VehicleRepository) Search(ctx context.Context, companyID uuid.UUID, searchTerm string, limit, offset int) ([]models.Vehicle, error) {
	ctx, span := r.tracer.Start(ctx, "VehicleRepository.Search",
		trace.WithAttributes(
			attribute.String("company.id", companyID.String()),
			attribute.String("search_term", searchTerm),
			attribute.Int("limit", limit),
			attribute.Int("offset", offset),
		))
	defer span.End()

	var vehicles []models.Vehicle
	searchPattern := "%" + strings.ToLower(searchTerm) + "%"

	query := `
		SELECT id, company_id, team_id, license_plate, brand, model, year, color,
			   vehicle_type, fuel_type, cargo_capacity, driver_id, helper_id, status,
			   created_at, updated_at
		FROM vehicles 
		WHERE company_id = $1 
		AND (LOWER(license_plate) LIKE $2 OR LOWER(brand) LIKE $2 OR LOWER(model) LIKE $2)
		AND status != 'deleted'
		ORDER BY license_plate ASC
		LIMIT $3 OFFSET $4
	`

	err := r.db.SelectContext(ctx, &vehicles, query, companyID, searchPattern, limit, offset)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to search vehicles: %w", err)
	}

	span.SetAttributes(attribute.Int("vehicles.count", len(vehicles)))
	return vehicles, nil
}

// CheckLicensePlateExists checks if a license plate already exists within a company
func (r *VehicleRepository) CheckLicensePlateExists(ctx context.Context, licensePlate string, companyID uuid.UUID, excludeID *uuid.UUID) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "VehicleRepository.CheckLicensePlateExists",
		trace.WithAttributes(
			attribute.String("license_plate", licensePlate),
			attribute.String("company.id", companyID.String()),
		))
	defer span.End()

	query := `SELECT COUNT(*) FROM vehicles WHERE license_plate = $1 AND company_id = $2 AND status != 'deleted'`
	args := []interface{}{licensePlate, companyID}

	if excludeID != nil {
		query += ` AND id != $3`
		args = append(args, *excludeID)
	}

	var count int
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("failed to check license plate existence: %w", err)
	}

	return count > 0, nil
}
