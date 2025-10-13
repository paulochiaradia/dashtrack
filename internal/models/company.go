package models

import (
	"time"

	"github.com/google/uuid"
)

// Company represents a company/organization in the multi-tenant system
type Company struct {
	ID               uuid.UUID `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	Slug             string    `json:"slug" db:"slug"`
	Email            string    `json:"email" db:"email"`
	Phone            *string   `json:"phone" db:"phone"`
	Address          *string   `json:"address" db:"address"`
	City             *string   `json:"city" db:"city"`
	State            *string   `json:"state" db:"state"`
	Country          string    `json:"country" db:"country"`
	SubscriptionPlan string    `json:"subscription_plan" db:"subscription_plan"`
	MaxUsers         int       `json:"max_users" db:"max_users"`
	MaxVehicles      int       `json:"max_vehicles" db:"max_vehicles"`
	MaxSensors       int       `json:"max_sensors" db:"max_sensors"`
	Status           string    `json:"status" db:"status"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// Team represents a team within a company
type Team struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	CompanyID   uuid.UUID  `json:"company_id" db:"company_id"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description" db:"description"`
	ManagerID   *uuid.UUID `json:"manager_id" db:"manager_id"`
	Status      string     `json:"status" db:"status"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`

	// Populated fields (not in DB)
	Manager *User        `json:"manager,omitempty"`
	Members []TeamMember `json:"members,omitempty"`
	Company *Company     `json:"company,omitempty"`
}

// TeamMember represents the many-to-many relationship between teams and users
type TeamMember struct {
	ID         uuid.UUID `json:"id" db:"id"`
	TeamID     uuid.UUID `json:"team_id" db:"team_id"`
	UserID     uuid.UUID `json:"user_id" db:"user_id"`
	RoleInTeam string    `json:"role_in_team" db:"role_in_team"`
	JoinedAt   time.Time `json:"joined_at" db:"joined_at"`

	// Populated fields
	User *User `json:"user,omitempty"`
	Team *Team `json:"team,omitempty"`
}

// Vehicle represents a company vehicle with IoT sensors
type Vehicle struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	CompanyID     uuid.UUID  `json:"company_id" db:"company_id"`
	TeamID        *uuid.UUID `json:"team_id" db:"team_id"`
	LicensePlate  string     `json:"license_plate" db:"license_plate"`
	Brand         string     `json:"brand" db:"brand"`
	Model         string     `json:"model" db:"model"`
	Year          int        `json:"year" db:"year"`
	Color         *string    `json:"color" db:"color"`
	VehicleType   string     `json:"vehicle_type" db:"vehicle_type"`
	FuelType      string     `json:"fuel_type" db:"fuel_type"`
	CargoCapacity *float64   `json:"cargo_capacity" db:"cargo_capacity"`
	DriverID      *uuid.UUID `json:"driver_id" db:"driver_id"`
	HelperID      *uuid.UUID `json:"helper_id" db:"helper_id"`
	Status        string     `json:"status" db:"status"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`

	// Populated fields
	Company      *Company      `json:"company,omitempty"`
	Team         *Team         `json:"team,omitempty"`
	Driver       *User         `json:"driver,omitempty"`
	Helper       *User         `json:"helper,omitempty"`
	Sensors      []Sensor      `json:"sensors,omitempty"`
	ESP32Devices []ESP32Device `json:"esp32_devices,omitempty"`
}

// ESP32Device represents an ESP32 IoT device
type ESP32Device struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	CompanyID        *uuid.UUID `json:"company_id" db:"company_id"`
	DeviceID         string     `json:"device_id" db:"device_id"`
	DeviceName       string     `json:"device_name" db:"device_name"`
	FirmwareVersion  *string    `json:"firmware_version" db:"firmware_version"`
	HardwareRevision *string    `json:"hardware_revision" db:"hardware_revision"`
	WifiSSID         *string    `json:"wifi_ssid" db:"wifi_ssid"`
	IPAddress        *string    `json:"ip_address" db:"ip_address"`
	MACAddress       *string    `json:"mac_address" db:"mac_address"`
	VehicleID        *uuid.UUID `json:"vehicle_id" db:"vehicle_id"`
	InstallationDate *time.Time `json:"installation_date" db:"installation_date"`
	LastHeartbeat    *time.Time `json:"last_heartbeat" db:"last_heartbeat"`
	BatteryLevel     *float64   `json:"battery_level" db:"battery_level"`
	SignalStrength   *int       `json:"signal_strength" db:"signal_strength"`
	Status           string     `json:"status" db:"status"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`

	// Populated fields
	Company *Company `json:"company,omitempty"`
	Vehicle *Vehicle `json:"vehicle,omitempty"`
	Sensors []Sensor `json:"sensors,omitempty"`
}

// VehicleTrip represents a vehicle trip/delivery
type VehicleTrip struct {
	ID              uuid.UUID  `json:"id" db:"id"`
	VehicleID       uuid.UUID  `json:"vehicle_id" db:"vehicle_id"`
	DriverID        *uuid.UUID `json:"driver_id" db:"driver_id"`
	HelperID        *uuid.UUID `json:"helper_id" db:"helper_id"`
	StartLocation   *string    `json:"start_location" db:"start_location"`
	EndLocation     *string    `json:"end_location" db:"end_location"`
	StartLatitude   *float64   `json:"start_latitude" db:"start_latitude"`
	StartLongitude  *float64   `json:"start_longitude" db:"start_longitude"`
	EndLatitude     *float64   `json:"end_latitude" db:"end_latitude"`
	EndLongitude    *float64   `json:"end_longitude" db:"end_longitude"`
	StartTime       time.Time  `json:"start_time" db:"start_time"`
	EndTime         *time.Time `json:"end_time" db:"end_time"`
	DistanceKm      *float64   `json:"distance_km" db:"distance_km"`
	DurationMinutes *int       `json:"duration_minutes" db:"duration_minutes"`
	FuelConsumption *float64   `json:"fuel_consumption" db:"fuel_consumption"`
	Status          string     `json:"status" db:"status"`
	Notes           *string    `json:"notes" db:"notes"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`

	// Populated fields
	Vehicle *Vehicle `json:"vehicle,omitempty"`
	Driver  *User    `json:"driver,omitempty"`
	Helper  *User    `json:"helper,omitempty"`
}

// CompanySetting represents per-company configuration
type CompanySetting struct {
	ID           uuid.UUID `json:"id" db:"id"`
	CompanyID    uuid.UUID `json:"company_id" db:"company_id"`
	SettingKey   string    `json:"setting_key" db:"setting_key"`
	SettingValue *string   `json:"setting_value" db:"setting_value"`
	SettingType  string    `json:"setting_type" db:"setting_type"`
	Description  *string   `json:"description" db:"description"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// CreateCompanyRequest represents request to create a new company
type CreateCompanyRequest struct {
	Name             string  `json:"name" binding:"required,min=2,max=255"`
	Slug             string  `json:"slug" binding:"required,min=2,max=100"`
	Email            string  `json:"email" binding:"required,email"`
	Phone            *string `json:"phone"`
	Address          *string `json:"address"`
	City             *string `json:"city"`
	State            *string `json:"state"`
	Country          string  `json:"country"`
	SubscriptionPlan string  `json:"subscription_plan" binding:"required,oneof=basic premium enterprise"`
}

// CreateTeamRequest represents request to create a new team
type CreateTeamRequest struct {
	Name        string     `json:"name" binding:"required,min=2,max=255"`
	Description *string    `json:"description"`
	ManagerID   *uuid.UUID `json:"manager_id"`
}

// CreateVehicleRequest represents request to create a new vehicle
type CreateVehicleRequest struct {
	TeamID        *uuid.UUID `json:"team_id"`
	LicensePlate  string     `json:"license_plate" binding:"required,min=3,max=20"`
	Brand         string     `json:"brand" binding:"required"`
	Model         string     `json:"model" binding:"required"`
	Year          int        `json:"year" binding:"required,min=1900,max=2100"`
	Color         *string    `json:"color"`
	VehicleType   string     `json:"vehicle_type" binding:"required,oneof=truck van car motorcycle bus"`
	FuelType      string     `json:"fuel_type" binding:"required,oneof=gasoline diesel electric hybrid cng"`
	CargoCapacity *float64   `json:"cargo_capacity"`
	DriverID      *uuid.UUID `json:"driver_id"`
	HelperID      *uuid.UUID `json:"helper_id"`
}

// CreateESP32DeviceRequest represents request to register a new ESP32 device
type CreateESP32DeviceRequest struct {
	DeviceID         string     `json:"device_id" binding:"required,min=3,max=255"`
	DeviceName       string     `json:"device_name" binding:"required,min=2,max=255"`
	FirmwareVersion  *string    `json:"firmware_version"`
	HardwareRevision *string    `json:"hardware_revision"`
	WifiSSID         *string    `json:"wifi_ssid"`
	IPAddress        *string    `json:"ip_address"`
	MACAddress       *string    `json:"mac_address"`
	VehicleID        *uuid.UUID `json:"vehicle_id"`
	InstallationDate *time.Time `json:"installation_date"`
}

// UpdateVehicleAssignmentRequest represents request to assign/unassign users to vehicle
type UpdateVehicleAssignmentRequest struct {
	DriverID *uuid.UUID `json:"driver_id"`
	HelperID *uuid.UUID `json:"helper_id"`
	TeamID   *uuid.UUID `json:"team_id"`
}

// VehicleDashboardData represents real-time dashboard data for a vehicle
type VehicleDashboardData struct {
	Vehicle          Vehicle                `json:"vehicle"`
	CurrentLocation  *GPSReading            `json:"current_location,omitempty"`
	LatestSensorData map[string]interface{} `json:"latest_sensor_data"`
	ActiveTrip       *VehicleTrip           `json:"active_trip,omitempty"`
	TodayStats       VehicleDailyStats      `json:"today_stats"`
	Alerts           []SensorAlert          `json:"alerts"`
	ESP32Status      []ESP32Device          `json:"esp32_status"`
}

// VehicleDailyStats represents daily statistics for a vehicle
type VehicleDailyStats struct {
	TotalTrips         int     `json:"total_trips"`
	TotalDistanceKm    float64 `json:"total_distance_km"`
	TotalDurationHours float64 `json:"total_duration_hours"`
	FuelConsumption    float64 `json:"fuel_consumption"`
	AverageSpeed       float64 `json:"average_speed"`
	AlertsCount        int     `json:"alerts_count"`
}

// CompanyDashboardData represents dashboard data for a company
type CompanyDashboardData struct {
	Company      Company                `json:"company"`
	TotalStats   CompanyStats           `json:"total_stats"`
	ActiveTrips  []VehicleTrip          `json:"active_trips"`
	RecentAlerts []SensorAlert          `json:"recent_alerts"`
	Vehicles     []VehicleDashboardData `json:"vehicles"`
	Teams        []Team                 `json:"teams"`
}

// CompanyStats represents overall statistics for a company
type CompanyStats struct {
	TotalVehicles       int     `json:"total_vehicles"`
	ActiveVehicles      int     `json:"active_vehicles"`
	TotalSensors        int     `json:"total_sensors"`
	ActiveSensors       int     `json:"active_sensors"`
	TotalUsers          int     `json:"total_users"`
	ActiveUsers         int     `json:"active_users"`
	TotalTripsToday     int     `json:"total_trips_today"`
	TotalDistanceToday  float64 `json:"total_distance_today"`
	ActiveAlertsCount   int     `json:"active_alerts_count"`
	ESP32DevicesOnline  int     `json:"esp32_devices_online"`
	ESP32DevicesOffline int     `json:"esp32_devices_offline"`
}
