package models

import (
	"time"

	"github.com/google/uuid"
)

// SensorType representa os tipos de sensores suportados
type SensorType string

const (
	SensorTypeDHT11     SensorType = "dht11"
	SensorTypeGyroscope SensorType = "gyroscope"
	SensorTypeGPS       SensorType = "gps_neo6v2"
	SensorTypeGeneric   SensorType = "generic"
)

// SensorStatus representa o status do sensor
type SensorStatus string

const (
	SensorStatusActive   SensorStatus = "active"
	SensorStatusInactive SensorStatus = "inactive"
	SensorStatusError    SensorStatus = "error"
)

// Sensor representa um dispositivo sensor ESP32
type Sensor struct {
	ID          uuid.UUID    `json:"id" db:"id"`
	DeviceID    string       `json:"device_id" db:"device_id"`
	Name        string       `json:"name" db:"name"`
	Type        SensorType   `json:"type" db:"type"`
	Status      SensorStatus `json:"status" db:"status"`
	Location    string       `json:"location" db:"location"`
	Description string       `json:"description" db:"description"`
	UserID      uuid.UUID    `json:"user_id" db:"user_id"`
	CreatedAt   time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at"`
	LastSeen    *time.Time   `json:"last_seen" db:"last_seen"`
}

// SensorReading representa uma leitura base de sensor
type SensorReading struct {
	ID        uuid.UUID `json:"id" db:"id"`
	SensorID  uuid.UUID `json:"sensor_id" db:"sensor_id"`
	DeviceID  string    `json:"device_id" db:"device_id"`
	Timestamp time.Time `json:"timestamp" db:"timestamp"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// DHT11Reading representa leitura do sensor de temperatura/umidade DHT11
type DHT11Reading struct {
	SensorReading
	Temperature float64 `json:"temperature" db:"temperature"` // Celsius
	Humidity    float64 `json:"humidity" db:"humidity"`       // Percentage
	HeatIndex   float64 `json:"heat_index" db:"heat_index"`   // Calculated heat index
}

// GyroscopeReading representa leitura do sensor giroscópio (vibração)
type GyroscopeReading struct {
	SensorReading
	AccelX      float64 `json:"accel_x" db:"accel_x"`           // m/s²
	AccelY      float64 `json:"accel_y" db:"accel_y"`           // m/s²
	AccelZ      float64 `json:"accel_z" db:"accel_z"`           // m/s²
	GyroX       float64 `json:"gyro_x" db:"gyro_x"`             // rad/s
	GyroY       float64 `json:"gyro_y" db:"gyro_y"`             // rad/s
	GyroZ       float64 `json:"gyro_z" db:"gyro_z"`             // rad/s
	Magnitude   float64 `json:"magnitude" db:"magnitude"`       // Overall acceleration magnitude
	IsVibrating bool    `json:"is_vibrating" db:"is_vibrating"` // Threshold-based vibration detection
}

// GPSReading representa leitura do sensor GPS NEO-6V2
type GPSReading struct {
	SensorReading
	Latitude   float64 `json:"latitude" db:"latitude"`     // Decimal degrees
	Longitude  float64 `json:"longitude" db:"longitude"`   // Decimal degrees
	Altitude   float64 `json:"altitude" db:"altitude"`     // Meters above sea level
	Speed      float64 `json:"speed" db:"speed"`           // km/h
	Heading    float64 `json:"heading" db:"heading"`       // Degrees from North
	Satellites int     `json:"satellites" db:"satellites"` // Number of satellites
	HDOP       float64 `json:"hdop" db:"hdop"`             // Horizontal Dilution of Precision
	IsValid    bool    `json:"is_valid" db:"is_valid"`     // GPS fix validity
}

// SensorDataPayload representa o payload genérico recebido do ESP32
type SensorDataPayload struct {
	DeviceID  string                 `json:"device_id"`
	Type      SensorType             `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// SensorStats representa estatísticas de um sensor
type SensorStats struct {
	SensorID      uuid.UUID `json:"sensor_id"`
	ReadingsCount int64     `json:"readings_count"`
	LastReading   time.Time `json:"last_reading"`
	FirstReading  time.Time `json:"first_reading"`
	IsOnline      bool      `json:"is_online"`
}

// SensorAlert representa alertas baseados em thresholds
type SensorAlert struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	SensorID   uuid.UUID  `json:"sensor_id" db:"sensor_id"`
	Type       string     `json:"type" db:"type"` // temperature_high, vibration_detected, gps_out_of_bounds
	Message    string     `json:"message" db:"message"`
	Value      float64    `json:"value" db:"value"`
	Threshold  float64    `json:"threshold" db:"threshold"`
	Severity   string     `json:"severity" db:"severity"` // low, medium, high, critical
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	ResolvedAt *time.Time `json:"resolved_at" db:"resolved_at"`
}
