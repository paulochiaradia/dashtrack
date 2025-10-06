package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/paulochiaradia/dashtrack/internal/models"
)

// SensorRepositoryInterface define os métodos para operações com sensores
type SensorRepositoryInterface interface {
	// Sensor CRUD
	CreateSensor(sensor *models.Sensor) error
	GetSensorByID(id uuid.UUID) (*models.Sensor, error)
	GetSensorByDeviceID(deviceID string) (*models.Sensor, error)
	GetSensorsByUserID(userID uuid.UUID) ([]*models.Sensor, error)
	UpdateSensor(sensor *models.Sensor) error
	DeleteSensor(id uuid.UUID) error
	UpdateSensorLastSeen(deviceID string) error

	// DHT11 Readings
	CreateDHT11Reading(reading *models.DHT11Reading) error
	GetDHT11ReadingsByDevice(deviceID string, limit int) ([]*models.DHT11Reading, error)
	GetDHT11ReadingsByTimeRange(deviceID string, start, end time.Time) ([]*models.DHT11Reading, error)
	GetLatestDHT11Reading(deviceID string) (*models.DHT11Reading, error)

	// Gyroscope Readings
	CreateGyroscopeReading(reading *models.GyroscopeReading) error
	GetGyroscopeReadingsByDevice(deviceID string, limit int) ([]*models.GyroscopeReading, error)
	GetGyroscopeReadingsByTimeRange(deviceID string, start, end time.Time) ([]*models.GyroscopeReading, error)
	GetLatestGyroscopeReading(deviceID string) (*models.GyroscopeReading, error)

	// GPS Readings
	CreateGPSReading(reading *models.GPSReading) error
	GetGPSReadingsByDevice(deviceID string, limit int) ([]*models.GPSReading, error)
	GetGPSReadingsByTimeRange(deviceID string, start, end time.Time) ([]*models.GPSReading, error)
	GetLatestGPSReading(deviceID string) (*models.GPSReading, error)

	// Alerts
	CreateSensorAlert(alert *models.SensorAlert) error
	GetActiveAlertsBySensor(sensorID uuid.UUID) ([]*models.SensorAlert, error)
	ResolveSensorAlert(alertID uuid.UUID) error

	// Statistics
	GetSensorStats(sensorID uuid.UUID) (*models.SensorStats, error)
}

// SensorRepository implementa SensorRepositoryInterface
type SensorRepository struct {
	db *sqlx.DB
}

// NewSensorRepository cria uma nova instância do repositório de sensores
func NewSensorRepository(db *sqlx.DB) SensorRepositoryInterface {
	return &SensorRepository{db: db}
}

// CreateSensor cria um novo sensor
func (r *SensorRepository) CreateSensor(sensor *models.Sensor) error {
	if sensor.ID == uuid.Nil {
		sensor.ID = uuid.New()
	}
	sensor.CreatedAt = time.Now()
	sensor.UpdatedAt = time.Now()

	query := `
		INSERT INTO sensors (id, device_id, name, type, status, location, description, user_id, created_at, updated_at)
		VALUES (:id, :device_id, :name, :type, :status, :location, :description, :user_id, :created_at, :updated_at)`

	_, err := r.db.NamedExec(query, sensor)
	return err
}

// GetSensorByID busca um sensor pelo ID
func (r *SensorRepository) GetSensorByID(id uuid.UUID) (*models.Sensor, error) {
	var sensor models.Sensor
	query := `SELECT * FROM sensors WHERE id = $1`
	err := r.db.Get(&sensor, query, id)
	if err != nil {
		return nil, err
	}
	return &sensor, nil
}

// GetSensorByDeviceID busca um sensor pelo device_id
func (r *SensorRepository) GetSensorByDeviceID(deviceID string) (*models.Sensor, error) {
	var sensor models.Sensor
	query := `SELECT * FROM sensors WHERE device_id = $1`
	err := r.db.Get(&sensor, query, deviceID)
	if err != nil {
		return nil, err
	}
	return &sensor, nil
}

// GetSensorsByUserID busca todos os sensores de um usuário
func (r *SensorRepository) GetSensorsByUserID(userID uuid.UUID) ([]*models.Sensor, error) {
	var sensors []*models.Sensor
	query := `SELECT * FROM sensors WHERE user_id = $1 ORDER BY created_at DESC`
	err := r.db.Select(&sensors, query, userID)
	return sensors, err
}

// UpdateSensor atualiza um sensor
func (r *SensorRepository) UpdateSensor(sensor *models.Sensor) error {
	sensor.UpdatedAt = time.Now()
	query := `
		UPDATE sensors 
		SET name = :name, type = :type, status = :status, location = :location, 
		    description = :description, updated_at = :updated_at
		WHERE id = :id`
	_, err := r.db.NamedExec(query, sensor)
	return err
}

// DeleteSensor remove um sensor
func (r *SensorRepository) DeleteSensor(id uuid.UUID) error {
	query := `DELETE FROM sensors WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// UpdateSensorLastSeen atualiza o último visto do sensor
func (r *SensorRepository) UpdateSensorLastSeen(deviceID string) error {
	now := time.Now()
	query := `UPDATE sensors SET last_seen = $1 WHERE device_id = $2`
	_, err := r.db.Exec(query, now, deviceID)
	return err
}

// CreateDHT11Reading cria uma nova leitura DHT11
func (r *SensorRepository) CreateDHT11Reading(reading *models.DHT11Reading) error {
	if reading.ID == uuid.Nil {
		reading.ID = uuid.New()
	}
	reading.CreatedAt = time.Now()

	query := `
		INSERT INTO dht11_readings (id, sensor_id, device_id, temperature, humidity, heat_index, timestamp, created_at)
		VALUES (:id, :sensor_id, :device_id, :temperature, :humidity, :heat_index, :timestamp, :created_at)`

	_, err := r.db.NamedExec(query, reading)
	if err != nil {
		return err
	}

	// Atualizar last_seen do sensor
	return r.UpdateSensorLastSeen(reading.DeviceID)
}

// GetDHT11ReadingsByDevice busca leituras DHT11 por device_id
func (r *SensorRepository) GetDHT11ReadingsByDevice(deviceID string, limit int) ([]*models.DHT11Reading, error) {
	var readings []*models.DHT11Reading
	query := `
		SELECT * FROM dht11_readings 
		WHERE device_id = $1 
		ORDER BY timestamp DESC 
		LIMIT $2`
	err := r.db.Select(&readings, query, deviceID, limit)
	return readings, err
}

// GetDHT11ReadingsByTimeRange busca leituras DHT11 por período
func (r *SensorRepository) GetDHT11ReadingsByTimeRange(deviceID string, start, end time.Time) ([]*models.DHT11Reading, error) {
	var readings []*models.DHT11Reading
	query := `
		SELECT * FROM dht11_readings 
		WHERE device_id = $1 AND timestamp BETWEEN $2 AND $3 
		ORDER BY timestamp DESC`
	err := r.db.Select(&readings, query, deviceID, start, end)
	return readings, err
}

// GetLatestDHT11Reading busca a última leitura DHT11
func (r *SensorRepository) GetLatestDHT11Reading(deviceID string) (*models.DHT11Reading, error) {
	var reading models.DHT11Reading
	query := `
		SELECT * FROM dht11_readings 
		WHERE device_id = $1 
		ORDER BY timestamp DESC 
		LIMIT 1`
	err := r.db.Get(&reading, query, deviceID)
	if err != nil {
		return nil, err
	}
	return &reading, nil
}

// CreateGyroscopeReading cria uma nova leitura de giroscópio
func (r *SensorRepository) CreateGyroscopeReading(reading *models.GyroscopeReading) error {
	if reading.ID == uuid.Nil {
		reading.ID = uuid.New()
	}
	reading.CreatedAt = time.Now()

	query := `
		INSERT INTO gyroscope_readings (id, sensor_id, device_id, accel_x, accel_y, accel_z, 
		                               gyro_x, gyro_y, gyro_z, magnitude, is_vibrating, timestamp, created_at)
		VALUES (:id, :sensor_id, :device_id, :accel_x, :accel_y, :accel_z, 
		        :gyro_x, :gyro_y, :gyro_z, :magnitude, :is_vibrating, :timestamp, :created_at)`

	_, err := r.db.NamedExec(query, reading)
	if err != nil {
		return err
	}

	return r.UpdateSensorLastSeen(reading.DeviceID)
}

// GetGyroscopeReadingsByDevice busca leituras de giroscópio por device_id
func (r *SensorRepository) GetGyroscopeReadingsByDevice(deviceID string, limit int) ([]*models.GyroscopeReading, error) {
	var readings []*models.GyroscopeReading
	query := `
		SELECT * FROM gyroscope_readings 
		WHERE device_id = $1 
		ORDER BY timestamp DESC 
		LIMIT $2`
	err := r.db.Select(&readings, query, deviceID, limit)
	return readings, err
}

// GetGyroscopeReadingsByTimeRange busca leituras de giroscópio por período
func (r *SensorRepository) GetGyroscopeReadingsByTimeRange(deviceID string, start, end time.Time) ([]*models.GyroscopeReading, error) {
	var readings []*models.GyroscopeReading
	query := `
		SELECT * FROM gyroscope_readings 
		WHERE device_id = $1 AND timestamp BETWEEN $2 AND $3 
		ORDER BY timestamp DESC`
	err := r.db.Select(&readings, query, deviceID, start, end)
	return readings, err
}

// GetLatestGyroscopeReading busca a última leitura de giroscópio
func (r *SensorRepository) GetLatestGyroscopeReading(deviceID string) (*models.GyroscopeReading, error) {
	var reading models.GyroscopeReading
	query := `
		SELECT * FROM gyroscope_readings 
		WHERE device_id = $1 
		ORDER BY timestamp DESC 
		LIMIT 1`
	err := r.db.Get(&reading, query, deviceID)
	if err != nil {
		return nil, err
	}
	return &reading, nil
}

// CreateGPSReading cria uma nova leitura GPS
func (r *SensorRepository) CreateGPSReading(reading *models.GPSReading) error {
	if reading.ID == uuid.Nil {
		reading.ID = uuid.New()
	}
	reading.CreatedAt = time.Now()

	query := `
		INSERT INTO gps_readings (id, sensor_id, device_id, latitude, longitude, altitude, 
		                         speed, heading, satellites, hdop, is_valid, timestamp, created_at)
		VALUES (:id, :sensor_id, :device_id, :latitude, :longitude, :altitude, 
		        :speed, :heading, :satellites, :hdop, :is_valid, :timestamp, :created_at)`

	_, err := r.db.NamedExec(query, reading)
	if err != nil {
		return err
	}

	return r.UpdateSensorLastSeen(reading.DeviceID)
}

// GetGPSReadingsByDevice busca leituras GPS por device_id
func (r *SensorRepository) GetGPSReadingsByDevice(deviceID string, limit int) ([]*models.GPSReading, error) {
	var readings []*models.GPSReading
	query := `
		SELECT * FROM gps_readings 
		WHERE device_id = $1 
		ORDER BY timestamp DESC 
		LIMIT $2`
	err := r.db.Select(&readings, query, deviceID, limit)
	return readings, err
}

// GetGPSReadingsByTimeRange busca leituras GPS por período
func (r *SensorRepository) GetGPSReadingsByTimeRange(deviceID string, start, end time.Time) ([]*models.GPSReading, error) {
	var readings []*models.GPSReading
	query := `
		SELECT * FROM gps_readings 
		WHERE device_id = $1 AND timestamp BETWEEN $2 AND $3 
		ORDER BY timestamp DESC`
	err := r.db.Select(&readings, query, deviceID, start, end)
	return readings, err
}

// GetLatestGPSReading busca a última leitura GPS
func (r *SensorRepository) GetLatestGPSReading(deviceID string) (*models.GPSReading, error) {
	var reading models.GPSReading
	query := `
		SELECT * FROM gps_readings 
		WHERE device_id = $1 
		ORDER BY timestamp DESC 
		LIMIT 1`
	err := r.db.Get(&reading, query, deviceID)
	if err != nil {
		return nil, err
	}
	return &reading, nil
}

// CreateSensorAlert cria um novo alerta
func (r *SensorRepository) CreateSensorAlert(alert *models.SensorAlert) error {
	if alert.ID == uuid.Nil {
		alert.ID = uuid.New()
	}
	alert.CreatedAt = time.Now()

	query := `
		INSERT INTO sensor_alerts (id, sensor_id, type, message, value, threshold, severity, created_at)
		VALUES (:id, :sensor_id, :type, :message, :value, :threshold, :severity, :created_at)`

	_, err := r.db.NamedExec(query, alert)
	return err
}

// GetActiveAlertsBySensor busca alertas ativos de um sensor
func (r *SensorRepository) GetActiveAlertsBySensor(sensorID uuid.UUID) ([]*models.SensorAlert, error) {
	var alerts []*models.SensorAlert
	query := `
		SELECT * FROM sensor_alerts 
		WHERE sensor_id = $1 AND resolved_at IS NULL 
		ORDER BY created_at DESC`
	err := r.db.Select(&alerts, query, sensorID)
	return alerts, err
}

// ResolveSensorAlert resolve um alerta
func (r *SensorRepository) ResolveSensorAlert(alertID uuid.UUID) error {
	now := time.Now()
	query := `UPDATE sensor_alerts SET resolved_at = $1 WHERE id = $2`
	_, err := r.db.Exec(query, now, alertID)
	return err
}

// GetSensorStats busca estatísticas de um sensor
func (r *SensorRepository) GetSensorStats(sensorID uuid.UUID) (*models.SensorStats, error) {
	var stats models.SensorStats

	// Buscar informações do sensor
	sensor, err := r.GetSensorByID(sensorID)
	if err != nil {
		return nil, err
	}

	stats.SensorID = sensorID

	// Determinar a tabela baseada no tipo de sensor
	var tableName string
	switch sensor.Type {
	case models.SensorTypeDHT11:
		tableName = "dht11_readings"
	case models.SensorTypeGyroscope:
		tableName = "gyroscope_readings"
	case models.SensorTypeGPS:
		tableName = "gps_readings"
	default:
		return nil, fmt.Errorf("unsupported sensor type: %s", sensor.Type)
	}

	// Buscar estatísticas
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as readings_count,
			MIN(timestamp) as first_reading,
			MAX(timestamp) as last_reading
		FROM %s 
		WHERE sensor_id = $1`, tableName)

	row := r.db.QueryRow(query, sensorID)
	var firstReading, lastReading sql.NullTime
	err = row.Scan(&stats.ReadingsCount, &firstReading, &lastReading)
	if err != nil {
		return nil, err
	}

	if firstReading.Valid {
		stats.FirstReading = firstReading.Time
	}
	if lastReading.Valid {
		stats.LastReading = lastReading.Time
		// Considerar online se teve reading nos últimos 5 minutos
		stats.IsOnline = time.Since(lastReading.Time) < 5*time.Minute
	}

	return &stats, nil
}
