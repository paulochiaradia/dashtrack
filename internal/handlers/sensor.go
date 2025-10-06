package handlers

import (
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/paulochiaradia/dashtrack/internal/logger"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
)

// SensorHandler lida com operações relacionadas a sensores
type SensorHandler struct {
	sensorRepo repository.SensorRepositoryInterface
}

// NewSensorHandler cria uma nova instância do handler de sensores
func NewSensorHandler(sensorRepo repository.SensorRepositoryInterface) *SensorHandler {
	return &SensorHandler{
		sensorRepo: sensorRepo,
	}
}

// RegisterSensor registra um novo sensor ESP32
func (h *SensorHandler) RegisterSensor(c *gin.Context) {
	var req struct {
		DeviceID    string            `json:"device_id" binding:"required"`
		Name        string            `json:"name" binding:"required"`
		Type        models.SensorType `json:"type" binding:"required"`
		Location    string            `json:"location"`
		Description string            `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid sensor registration request",
			zap.String("error", err.Error()),
			zap.String("client_ip", c.ClientIP()))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Obter user_id do contexto (middleware de autenticação)
	userID, exists := c.Get("user_id")
	if !exists {
		logger.Warn("User ID not found in context for sensor registration",
			zap.String("device_id", req.DeviceID))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Verificar se o device_id já existe
	existingSensor, err := h.sensorRepo.GetSensorByDeviceID(req.DeviceID)
	if err == nil && existingSensor != nil {
		logger.Warn("Attempt to register duplicate sensor",
			zap.String("device_id", req.DeviceID),
			zap.String("existing_sensor_id", existingSensor.ID.String()))
		c.JSON(http.StatusConflict, gin.H{"error": "Device ID already registered"})
		return
	}

	sensor := &models.Sensor{
		DeviceID:    req.DeviceID,
		Name:        req.Name,
		Type:        req.Type,
		Status:      models.SensorStatusActive,
		Location:    req.Location,
		Description: req.Description,
		UserID:      userID.(uuid.UUID),
	}

	if err := h.sensorRepo.CreateSensor(sensor); err != nil {
		logger.Error("Failed to register sensor",
			zap.String("device_id", req.DeviceID),
			zap.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register sensor"})
		return
	}

	logger.Info("Sensor registered successfully",
		zap.String("sensor_id", sensor.ID.String()),
		zap.String("device_id", req.DeviceID),
		zap.String("type", string(req.Type)),
		zap.String("user_id", userID.(uuid.UUID).String()))

	c.JSON(http.StatusCreated, gin.H{
		"message": "Sensor registered successfully",
		"sensor":  sensor,
	})
}

// ReceiveSensorData recebe dados de sensores ESP32 via HTTP POST
func (h *SensorHandler) ReceiveSensorData(c *gin.Context) {
	var payload models.SensorDataPayload

	if err := c.ShouldBindJSON(&payload); err != nil {
		logger.Warn("Invalid sensor data payload",
			zap.String("error", err.Error()),
			zap.String("client_ip", c.ClientIP()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payload format"})
		return
	}

	// Validar se o sensor existe
	sensor, err := h.sensorRepo.GetSensorByDeviceID(payload.DeviceID)
	if err != nil {
		logger.Warn("Data received from unregistered sensor",
			zap.String("device_id", payload.DeviceID),
			zap.String("type", string(payload.Type)))
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not registered"})
		return
	}

	// Validar se o tipo do sensor coincide
	if sensor.Type != payload.Type {
		logger.Warn("Sensor type mismatch",
			zap.String("device_id", payload.DeviceID),
			zap.String("registered_type", string(sensor.Type)),
			zap.String("payload_type", string(payload.Type)))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sensor type mismatch"})
		return
	}

	// Processar dados baseado no tipo de sensor
	switch payload.Type {
	case models.SensorTypeDHT11:
		err = h.processDHT11Data(sensor, payload)
	case models.SensorTypeGyroscope:
		err = h.processGyroscopeData(sensor, payload)
	case models.SensorTypeGPS:
		err = h.processGPSData(sensor, payload)
	default:
		logger.Warn("Unsupported sensor type",
			zap.String("device_id", payload.DeviceID),
			zap.String("type", string(payload.Type)))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported sensor type"})
		return
	}

	if err != nil {
		logger.Error("Failed to process sensor data",
			zap.String("device_id", payload.DeviceID),
			zap.String("type", string(payload.Type)),
			zap.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process sensor data"})
		return
	}

	logger.Info("Sensor data processed successfully",
		zap.String("device_id", payload.DeviceID),
		zap.String("type", string(payload.Type)),
		zap.Time("timestamp", payload.Timestamp))

	c.JSON(http.StatusOK, gin.H{"message": "Data received successfully"})
}

// processDHT11Data processa dados do sensor DHT11
func (h *SensorHandler) processDHT11Data(sensor *models.Sensor, payload models.SensorDataPayload) error {
	// Extrair dados do payload
	temperature, ok := payload.Data["temperature"].(float64)
	if !ok {
		return &ValidationError{Field: "temperature", Message: "temperature field is required and must be a number"}
	}

	humidity, ok := payload.Data["humidity"].(float64)
	if !ok {
		return &ValidationError{Field: "humidity", Message: "humidity field is required and must be a number"}
	}

	// Validações
	if temperature < -40 || temperature > 80 {
		return &ValidationError{Field: "temperature", Message: "temperature must be between -40 and 80 Celsius"}
	}

	if humidity < 0 || humidity > 100 {
		return &ValidationError{Field: "humidity", Message: "humidity must be between 0 and 100 percent"}
	}

	// Calcular heat index (índice de calor)
	heatIndex := calculateHeatIndex(temperature, humidity)

	reading := &models.DHT11Reading{
		SensorReading: models.SensorReading{
			SensorID:  sensor.ID,
			DeviceID:  payload.DeviceID,
			Timestamp: payload.Timestamp,
		},
		Temperature: temperature,
		Humidity:    humidity,
		HeatIndex:   heatIndex,
	}

	// Salvar leitura
	err := h.sensorRepo.CreateDHT11Reading(reading)
	if err != nil {
		return err
	}

	// Verificar alertas
	h.checkTemperatureAlerts(sensor, temperature, humidity)

	return nil
}

// processGyroscopeData processa dados do sensor giroscópio
func (h *SensorHandler) processGyroscopeData(sensor *models.Sensor, payload models.SensorDataPayload) error {
	// Extrair dados do payload
	accelX, _ := payload.Data["accel_x"].(float64)
	accelY, _ := payload.Data["accel_y"].(float64)
	accelZ, _ := payload.Data["accel_z"].(float64)
	gyroX, _ := payload.Data["gyro_x"].(float64)
	gyroY, _ := payload.Data["gyro_y"].(float64)
	gyroZ, _ := payload.Data["gyro_z"].(float64)

	// Calcular magnitude da aceleração
	magnitude := math.Sqrt(accelX*accelX + accelY*accelY + accelZ*accelZ)

	// Detectar vibração (threshold configurável)
	vibrationThreshold := 15.0 // m/s² - pode ser configurável por sensor
	isVibrating := magnitude > vibrationThreshold

	reading := &models.GyroscopeReading{
		SensorReading: models.SensorReading{
			SensorID:  sensor.ID,
			DeviceID:  payload.DeviceID,
			Timestamp: payload.Timestamp,
		},
		AccelX:      accelX,
		AccelY:      accelY,
		AccelZ:      accelZ,
		GyroX:       gyroX,
		GyroY:       gyroY,
		GyroZ:       gyroZ,
		Magnitude:   magnitude,
		IsVibrating: isVibrating,
	}

	// Salvar leitura
	err := h.sensorRepo.CreateGyroscopeReading(reading)
	if err != nil {
		return err
	}

	// Verificar alertas de vibração
	if isVibrating {
		h.checkVibrationAlerts(sensor, magnitude, vibrationThreshold)
	}

	return nil
}

// processGPSData processa dados do sensor GPS
func (h *SensorHandler) processGPSData(sensor *models.Sensor, payload models.SensorDataPayload) error {
	// Extrair dados do payload
	latitude, _ := payload.Data["latitude"].(float64)
	longitude, _ := payload.Data["longitude"].(float64)
	altitude, _ := payload.Data["altitude"].(float64)
	speed, _ := payload.Data["speed"].(float64)
	heading, _ := payload.Data["heading"].(float64)
	satellites, _ := payload.Data["satellites"].(float64)
	hdop, _ := payload.Data["hdop"].(float64)
	isValid, _ := payload.Data["is_valid"].(bool)

	// Validações básicas
	if latitude < -90 || latitude > 90 {
		return &ValidationError{Field: "latitude", Message: "latitude must be between -90 and 90"}
	}

	if longitude < -180 || longitude > 180 {
		return &ValidationError{Field: "longitude", Message: "longitude must be between -180 and 180"}
	}

	reading := &models.GPSReading{
		SensorReading: models.SensorReading{
			SensorID:  sensor.ID,
			DeviceID:  payload.DeviceID,
			Timestamp: payload.Timestamp,
		},
		Latitude:   latitude,
		Longitude:  longitude,
		Altitude:   altitude,
		Speed:      speed,
		Heading:    heading,
		Satellites: int(satellites),
		HDOP:       hdop,
		IsValid:    isValid,
	}

	// Salvar leitura
	err := h.sensorRepo.CreateGPSReading(reading)
	if err != nil {
		return err
	}

	// Verificar alertas de localização (se necessário)
	// h.checkLocationAlerts(sensor, latitude, longitude)

	return nil
}

// GetSensorData retorna dados de um sensor
func (h *SensorHandler) GetSensorData(c *gin.Context) {
	deviceID := c.Param("device_id")
	sensorType := c.Query("type")
	limitStr := c.DefaultQuery("limit", "100")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 1000 {
		limit = 100
	}

	// Verificar se o sensor existe
	sensor, err := h.sensorRepo.GetSensorByDeviceID(deviceID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sensor not found"})
		return
	}

	// Verificar permissão do usuário
	userID, exists := c.Get("user_id")
	if !exists || sensor.UserID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var data interface{}

	// Buscar dados baseado no tipo
	switch models.SensorType(sensorType) {
	case models.SensorTypeDHT11:
		data, err = h.sensorRepo.GetDHT11ReadingsByDevice(deviceID, limit)
	case models.SensorTypeGyroscope:
		data, err = h.sensorRepo.GetGyroscopeReadingsByDevice(deviceID, limit)
	case models.SensorTypeGPS:
		data, err = h.sensorRepo.GetGPSReadingsByDevice(deviceID, limit)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sensor type"})
		return
	}

	if err != nil {
		logger.Error("Failed to fetch sensor data",
			zap.String("device_id", deviceID),
			zap.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sensor": sensor,
		"data":   data,
	})
}

// GetMySensors retorna todos os sensores do usuário
func (h *SensorHandler) GetMySensors(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	sensors, err := h.sensorRepo.GetSensorsByUserID(userID.(uuid.UUID))
	if err != nil {
		logger.Error("Failed to fetch user sensors",
			zap.String("user_id", userID.(uuid.UUID).String()),
			zap.String("error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sensors"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sensors": sensors})
}

// Helper functions

// ValidationError representa um erro de validação
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// calculateHeatIndex calcula o índice de calor
func calculateHeatIndex(tempC, humidity float64) float64 {
	// Converter para Fahrenheit para o cálculo
	tempF := tempC*9/5 + 32

	// Fórmula simplificada do heat index
	if tempF < 80 {
		return tempC // Para temperaturas baixas, retorna a temperatura original
	}

	hi := -42.379 + 2.04901523*tempF + 10.14333127*humidity - 0.22475541*tempF*humidity
	hi += -0.00683783*tempF*tempF - 0.05481717*humidity*humidity + 0.00122874*tempF*tempF*humidity
	hi += 0.00085282*tempF*humidity*humidity - 0.00000199*tempF*tempF*humidity*humidity

	// Converter de volta para Celsius
	return (hi - 32) * 5 / 9
}

// checkTemperatureAlerts verifica alertas de temperatura
func (h *SensorHandler) checkTemperatureAlerts(sensor *models.Sensor, temperature, humidity float64) {
	// Exemplo de thresholds (podem ser configuráveis por sensor)
	if temperature > 35 { // Temperatura alta
		alert := &models.SensorAlert{
			SensorID:  sensor.ID,
			Type:      "temperature_high",
			Message:   "Temperature above safe threshold",
			Value:     temperature,
			Threshold: 35,
			Severity:  "medium",
		}
		h.sensorRepo.CreateSensorAlert(alert)
	}

	if humidity > 80 { // Umidade alta
		alert := &models.SensorAlert{
			SensorID:  sensor.ID,
			Type:      "humidity_high",
			Message:   "Humidity above safe threshold",
			Value:     humidity,
			Threshold: 80,
			Severity:  "low",
		}
		h.sensorRepo.CreateSensorAlert(alert)
	}
}

// checkVibrationAlerts verifica alertas de vibração
func (h *SensorHandler) checkVibrationAlerts(sensor *models.Sensor, magnitude, threshold float64) {
	severity := "low"
	if magnitude > threshold*2 {
		severity = "high"
	} else if magnitude > threshold*1.5 {
		severity = "medium"
	}

	alert := &models.SensorAlert{
		SensorID:  sensor.ID,
		Type:      "vibration_detected",
		Message:   "Vibration detected above threshold",
		Value:     magnitude,
		Threshold: threshold,
		Severity:  severity,
	}
	h.sensorRepo.CreateSensorAlert(alert)
}
