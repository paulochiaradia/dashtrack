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

// ESP32DeviceRepository handles database operations for ESP32 devices
type ESP32DeviceRepository struct {
	db     *sqlx.DB
	tracer trace.Tracer
}

// NewESP32DeviceRepository creates a new ESP32 device repository
func NewESP32DeviceRepository(db *sqlx.DB) *ESP32DeviceRepository {
	return &ESP32DeviceRepository{
		db:     db,
		tracer: otel.Tracer("esp32-device-repository"),
	}
}

// Create creates a new ESP32 device
func (r *ESP32DeviceRepository) Create(ctx context.Context, device *models.ESP32Device) error {
	ctx, span := r.tracer.Start(ctx, "ESP32DeviceRepository.Create",
		trace.WithAttributes(
			attribute.String("device.device_id", device.DeviceID),
			attribute.String("device.name", device.DeviceName),
		))
	defer span.End()

	// Add company_id attribute only if it's not nil
	if device.CompanyID != nil {
		span.SetAttributes(attribute.String("company.id", device.CompanyID.String()))
	}

	device.ID = uuid.New()
	device.CreatedAt = time.Now()
	device.UpdatedAt = time.Now()
	if device.Status == "" {
		device.Status = "active"
	}

	query := `
		INSERT INTO esp32_devices (
			id, company_id, device_id, device_name, firmware_version, hardware_revision,
			wifi_ssid, ip_address, mac_address, vehicle_id, installation_date,
			last_heartbeat, battery_level, signal_strength, status, created_at, updated_at
		) VALUES (
			:id, :company_id, :device_id, :device_name, :firmware_version, :hardware_revision,
			:wifi_ssid, :ip_address, :mac_address, :vehicle_id, :installation_date,
			:last_heartbeat, :battery_level, :signal_strength, :status, :created_at, :updated_at
		)
	`

	_, err := r.db.NamedExecContext(ctx, query, device)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to create ESP32 device: %w", err)
	}

	span.SetAttributes(attribute.String("device.id", device.ID.String()))
	return nil
}

// GetByID retrieves an ESP32 device by ID with company context
func (r *ESP32DeviceRepository) GetByID(ctx context.Context, id uuid.UUID, companyID uuid.UUID) (*models.ESP32Device, error) {
	ctx, span := r.tracer.Start(ctx, "ESP32DeviceRepository.GetByID",
		trace.WithAttributes(
			attribute.String("device.id", id.String()),
			attribute.String("company.id", companyID.String()),
		))
	defer span.End()

	var device models.ESP32Device
	query := `
		SELECT id, company_id, device_id, device_name, firmware_version, hardware_revision,
			   wifi_ssid, ip_address, mac_address, vehicle_id, installation_date,
			   last_heartbeat, battery_level, signal_strength, status, created_at, updated_at
		FROM esp32_devices 
		WHERE id = $1 AND company_id = $2
	`

	err := r.db.GetContext(ctx, &device, query, id, companyID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get ESP32 device by ID: %w", err)
	}

	return &device, nil
}

// GetByDeviceID retrieves an ESP32 device by device ID
func (r *ESP32DeviceRepository) GetByDeviceID(ctx context.Context, deviceID string) (*models.ESP32Device, error) {
	ctx, span := r.tracer.Start(ctx, "ESP32DeviceRepository.GetByDeviceID",
		trace.WithAttributes(attribute.String("device.device_id", deviceID)))
	defer span.End()

	var device models.ESP32Device
	query := `
		SELECT id, company_id, device_id, device_name, firmware_version, hardware_revision,
			   wifi_ssid, ip_address, mac_address, vehicle_id, installation_date,
			   last_heartbeat, battery_level, signal_strength, status, created_at, updated_at
		FROM esp32_devices 
		WHERE device_id = $1
	`

	err := r.db.GetContext(ctx, &device, query, deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get ESP32 device by device ID: %w", err)
	}

	return &device, nil
}

// GetByCompany retrieves all ESP32 devices for a company
func (r *ESP32DeviceRepository) GetByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int) ([]models.ESP32Device, error) {
	ctx, span := r.tracer.Start(ctx, "ESP32DeviceRepository.GetByCompany",
		trace.WithAttributes(
			attribute.String("company.id", companyID.String()),
			attribute.Int("limit", limit),
			attribute.Int("offset", offset),
		))
	defer span.End()

	var devices []models.ESP32Device
	query := `
		SELECT id, company_id, device_id, device_name, firmware_version, hardware_revision,
			   wifi_ssid, ip_address, mac_address, vehicle_id, installation_date,
			   last_heartbeat, battery_level, signal_strength, status, created_at, updated_at
		FROM esp32_devices 
		WHERE company_id = $1 AND status != 'deleted'
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	err := r.db.SelectContext(ctx, &devices, query, companyID, limit, offset)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get ESP32 devices by company: %w", err)
	}

	span.SetAttributes(attribute.Int("devices.count", len(devices)))
	return devices, nil
}

// GetByVehicle retrieves all ESP32 devices for a vehicle
func (r *ESP32DeviceRepository) GetByVehicle(ctx context.Context, vehicleID uuid.UUID, companyID uuid.UUID) ([]models.ESP32Device, error) {
	ctx, span := r.tracer.Start(ctx, "ESP32DeviceRepository.GetByVehicle",
		trace.WithAttributes(
			attribute.String("vehicle.id", vehicleID.String()),
			attribute.String("company.id", companyID.String()),
		))
	defer span.End()

	var devices []models.ESP32Device
	query := `
		SELECT id, company_id, device_id, device_name, firmware_version, hardware_revision,
			   wifi_ssid, ip_address, mac_address, vehicle_id, installation_date,
			   last_heartbeat, battery_level, signal_strength, status, created_at, updated_at
		FROM esp32_devices 
		WHERE vehicle_id = $1 AND company_id = $2 AND status != 'deleted'
		ORDER BY device_name ASC
	`

	err := r.db.SelectContext(ctx, &devices, query, vehicleID, companyID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get ESP32 devices by vehicle: %w", err)
	}

	span.SetAttributes(attribute.Int("devices.count", len(devices)))
	return devices, nil
}

// Update updates an ESP32 device
func (r *ESP32DeviceRepository) Update(ctx context.Context, device *models.ESP32Device) error {
	ctx, span := r.tracer.Start(ctx, "ESP32DeviceRepository.Update",
		trace.WithAttributes(attribute.String("device.id", device.ID.String())))
	defer span.End()

	device.UpdatedAt = time.Now()

	query := `
		UPDATE esp32_devices SET
			device_name = :device_name,
			firmware_version = :firmware_version,
			hardware_revision = :hardware_revision,
			wifi_ssid = :wifi_ssid,
			ip_address = :ip_address,
			mac_address = :mac_address,
			vehicle_id = :vehicle_id,
			installation_date = :installation_date,
			last_heartbeat = :last_heartbeat,
			battery_level = :battery_level,
			signal_strength = :signal_strength,
			status = :status,
			updated_at = :updated_at
		WHERE id = :id AND company_id = :company_id
	`

	result, err := r.db.NamedExecContext(ctx, query, device)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update ESP32 device: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("ESP32 device not found or not authorized")
	}

	return nil
}

// UpdateHeartbeat updates the heartbeat timestamp and status for an ESP32 device
func (r *ESP32DeviceRepository) UpdateHeartbeat(ctx context.Context, deviceID string, batteryLevel *float64, signalStrength *int) error {
	ctx, span := r.tracer.Start(ctx, "ESP32DeviceRepository.UpdateHeartbeat",
		trace.WithAttributes(attribute.String("device.device_id", deviceID)))
	defer span.End()

	query := `
		UPDATE esp32_devices SET
			last_heartbeat = NOW(),
			battery_level = COALESCE($2, battery_level),
			signal_strength = COALESCE($3, signal_strength),
			status = 'online',
			updated_at = NOW()
		WHERE device_id = $1
	`

	result, err := r.db.ExecContext(ctx, query, deviceID, batteryLevel, signalStrength)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update ESP32 device heartbeat: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("ESP32 device not found")
	}

	return nil
}

// UpdateVehicleAssignment assigns or unassigns an ESP32 device to/from a vehicle
func (r *ESP32DeviceRepository) UpdateVehicleAssignment(ctx context.Context, deviceID uuid.UUID, companyID uuid.UUID, vehicleID *uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "ESP32DeviceRepository.UpdateVehicleAssignment",
		trace.WithAttributes(
			attribute.String("device.id", deviceID.String()),
			attribute.String("company.id", companyID.String()),
		))
	defer span.End()

	query := `
		UPDATE esp32_devices SET
			vehicle_id = $1,
			updated_at = NOW()
		WHERE id = $2 AND company_id = $3
	`

	result, err := r.db.ExecContext(ctx, query, vehicleID, deviceID, companyID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to update ESP32 device vehicle assignment: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("ESP32 device not found or not authorized")
	}

	return nil
}

// Delete soft deletes an ESP32 device
func (r *ESP32DeviceRepository) Delete(ctx context.Context, id uuid.UUID, companyID uuid.UUID) error {
	ctx, span := r.tracer.Start(ctx, "ESP32DeviceRepository.Delete",
		trace.WithAttributes(
			attribute.String("device.id", id.String()),
			attribute.String("company.id", companyID.String()),
		))
	defer span.End()

	query := `
		UPDATE esp32_devices 
		SET status = 'deleted', updated_at = NOW() 
		WHERE id = $1 AND company_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, id, companyID)
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to delete ESP32 device: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("ESP32 device not found or not authorized")
	}

	return nil
}

// MarkOfflineDevices marks devices as offline if they haven't sent heartbeat in specified duration
func (r *ESP32DeviceRepository) MarkOfflineDevices(ctx context.Context, timeoutMinutes int) error {
	ctx, span := r.tracer.Start(ctx, "ESP32DeviceRepository.MarkOfflineDevices",
		trace.WithAttributes(attribute.Int("timeout_minutes", timeoutMinutes)))
	defer span.End()

	query := `
		UPDATE esp32_devices 
		SET status = 'offline', updated_at = NOW()
		WHERE status = 'online' 
		AND (last_heartbeat IS NULL OR last_heartbeat < NOW() - INTERVAL '%d minutes')
	`

	result, err := r.db.ExecContext(ctx, fmt.Sprintf(query, timeoutMinutes))
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to mark offline devices: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	span.SetAttributes(attribute.Int64("devices_marked_offline", rowsAffected))

	return nil
}

// GetOnlineDevices retrieves all online ESP32 devices for a company
func (r *ESP32DeviceRepository) GetOnlineDevices(ctx context.Context, companyID uuid.UUID) ([]models.ESP32Device, error) {
	ctx, span := r.tracer.Start(ctx, "ESP32DeviceRepository.GetOnlineDevices",
		trace.WithAttributes(attribute.String("company.id", companyID.String())))
	defer span.End()

	var devices []models.ESP32Device
	query := `
		SELECT id, company_id, device_id, device_name, firmware_version, hardware_revision,
			   wifi_ssid, ip_address, mac_address, vehicle_id, installation_date,
			   last_heartbeat, battery_level, signal_strength, status, created_at, updated_at
		FROM esp32_devices 
		WHERE company_id = $1 AND status = 'online'
		ORDER BY last_heartbeat DESC
	`

	err := r.db.SelectContext(ctx, &devices, query, companyID)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get online ESP32 devices: %w", err)
	}

	span.SetAttributes(attribute.Int("devices.count", len(devices)))
	return devices, nil
}

// Search searches ESP32 devices by device ID, name, or MAC address
func (r *ESP32DeviceRepository) Search(ctx context.Context, companyID uuid.UUID, searchTerm string, limit, offset int) ([]models.ESP32Device, error) {
	ctx, span := r.tracer.Start(ctx, "ESP32DeviceRepository.Search",
		trace.WithAttributes(
			attribute.String("company.id", companyID.String()),
			attribute.String("search_term", searchTerm),
			attribute.Int("limit", limit),
			attribute.Int("offset", offset),
		))
	defer span.End()

	var devices []models.ESP32Device
	searchPattern := "%" + strings.ToLower(searchTerm) + "%"

	query := `
		SELECT id, company_id, device_id, device_name, firmware_version, hardware_revision,
			   wifi_ssid, ip_address, mac_address, vehicle_id, installation_date,
			   last_heartbeat, battery_level, signal_strength, status, created_at, updated_at
		FROM esp32_devices 
		WHERE company_id = $1 
		AND (LOWER(device_id) LIKE $2 OR LOWER(device_name) LIKE $2 OR LOWER(mac_address) LIKE $2)
		AND status != 'deleted'
		ORDER BY device_name ASC
		LIMIT $3 OFFSET $4
	`

	err := r.db.SelectContext(ctx, &devices, query, companyID, searchPattern, limit, offset)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to search ESP32 devices: %w", err)
	}

	span.SetAttributes(attribute.Int("devices.count", len(devices)))
	return devices, nil
}

// CheckDeviceIDExists checks if a device ID already exists
func (r *ESP32DeviceRepository) CheckDeviceIDExists(ctx context.Context, deviceID string, excludeID *uuid.UUID) (bool, error) {
	ctx, span := r.tracer.Start(ctx, "ESP32DeviceRepository.CheckDeviceIDExists",
		trace.WithAttributes(attribute.String("device.device_id", deviceID)))
	defer span.End()

	query := `SELECT COUNT(*) FROM esp32_devices WHERE device_id = $1 AND status != 'deleted'`
	args := []interface{}{deviceID}

	if excludeID != nil {
		query += ` AND id != $2`
		args = append(args, *excludeID)
	}

	var count int
	err := r.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("failed to check device ID existence: %w", err)
	}

	return count > 0, nil
}

// GetDeviceStatistics retrieves statistics for ESP32 devices
func (r *ESP32DeviceRepository) GetDeviceStatistics(ctx context.Context, companyID *uuid.UUID) (map[string]interface{}, error) {
	ctx, span := r.tracer.Start(ctx, "ESP32DeviceRepository.GetDeviceStatistics")
	defer span.End()

	query := `
		SELECT 
			COUNT(*) as total_devices,
			COUNT(CASE WHEN status = 'online' THEN 1 END) as online_devices,
			COUNT(CASE WHEN status = 'offline' THEN 1 END) as offline_devices,
			COUNT(CASE WHEN status = 'inactive' THEN 1 END) as inactive_devices,
			AVG(CASE WHEN battery_level IS NOT NULL THEN battery_level END) as avg_battery_level,
			AVG(CASE WHEN signal_strength IS NOT NULL THEN signal_strength END) as avg_signal_strength
		FROM esp32_devices 
		WHERE status != 'deleted'`

	args := []interface{}{}
	if companyID != nil {
		query += ` AND company_id = $1`
		args = append(args, *companyID)
		span.SetAttributes(attribute.String("company.id", companyID.String()))
	}

	var stats struct {
		TotalDevices      int      `db:"total_devices"`
		OnlineDevices     int      `db:"online_devices"`
		OfflineDevices    int      `db:"offline_devices"`
		InactiveDevices   int      `db:"inactive_devices"`
		AvgBatteryLevel   *float64 `db:"avg_battery_level"`
		AvgSignalStrength *float64 `db:"avg_signal_strength"`
	}

	err := r.db.GetContext(ctx, &stats, query, args...)
	if err != nil {
		span.RecordError(err)
		return nil, fmt.Errorf("failed to get device statistics: %w", err)
	}

	result := map[string]interface{}{
		"total_devices":       stats.TotalDevices,
		"online_devices":      stats.OnlineDevices,
		"offline_devices":     stats.OfflineDevices,
		"inactive_devices":    stats.InactiveDevices,
		"avg_battery_level":   stats.AvgBatteryLevel,
		"avg_signal_strength": stats.AvgSignalStrength,
	}

	return result, nil
}
