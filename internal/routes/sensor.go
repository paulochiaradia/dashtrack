package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/paulochiaradia/dashtrack/internal/handlers"
)

// SetupSensorRoutes configura as rotas relacionadas a sensores IoT
func (r *Router) SetupSensorRoutes(sensorHandler *handlers.SensorHandler) {
	// Use router's auth middleware (already configured with tokenService)
	authMiddleware := r.authMiddleware

	// Grupo de rotas para sensores (requer autenticação)
	sensorGroup := r.engine.Group("/api/v1/sensors")
	sensorGroup.Use(authMiddleware.RequireAuth())

	// Registrar um novo sensor
	sensorGroup.POST("/register", sensorHandler.RegisterSensor)

	// Listar sensores do usuário
	sensorGroup.GET("/my", sensorHandler.GetMySensors)

	// Obter dados de um sensor específico
	sensorGroup.GET("/:device_id/data", sensorHandler.GetSensorData)

	// Grupo de rotas para recepção de dados dos ESP32 (sem autenticação JWT)
	// Note: Para ESP32, usaremos autenticação baseada em device_id ou API key
	iotGroup := r.engine.Group("/api/v1/iot")

	// Receber dados de sensores ESP32
	iotGroup.POST("/data", sensorHandler.ReceiveSensorData)

	// Health check específico para IoT
	iotGroup.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "ok",
			"service":   "iot-gateway",
			"timestamp": time.Now().Format(time.RFC3339),
		})
	})
}
