package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/paulochiaradia/dashtrack/internal/middleware"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
	"github.com/paulochiaradia/dashtrack/internal/utils"
)

// ESP32DeviceHandler handles ESP32 device-related HTTP requests
type ESP32DeviceHandler struct {
	esp32Repo   *repository.ESP32DeviceRepository
	vehicleRepo *repository.VehicleRepository
	tracer      trace.Tracer
}

// NewESP32DeviceHandler creates a new ESP32 device handler
func NewESP32DeviceHandler(esp32Repo *repository.ESP32DeviceRepository, vehicleRepo *repository.VehicleRepository) *ESP32DeviceHandler {
	return &ESP32DeviceHandler{
		esp32Repo:   esp32Repo,
		vehicleRepo: vehicleRepo,
		tracer:      otel.Tracer("esp32-device-handler"),
	}
}

// CreateDevice creates a new ESP32 device
func (h *ESP32DeviceHandler) CreateDevice(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "ESP32DeviceHandler.CreateDevice")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	var req models.CreateESP32DeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate vehicle if provided
	if req.VehicleID != nil {
		vehicle, err := h.vehicleRepo.GetByID(ctx, *req.VehicleID, *companyID)
		if err != nil || vehicle == nil {
			utils.BadRequestResponse(c, "Invalid vehicle ID or vehicle does not belong to company")
			return
		}
	}

	device := &models.ESP32Device{
		CompanyID:        companyID,
		DeviceName:       req.DeviceName,
		DeviceID:         req.DeviceID,
		MACAddress:       req.MACAddress,
		IPAddress:        req.IPAddress,
		FirmwareVersion:  req.FirmwareVersion,
		HardwareRevision: req.HardwareRevision,
		WifiSSID:         req.WifiSSID,
		VehicleID:        req.VehicleID,
		InstallationDate: req.InstallationDate,
		Status:           "offline", // Default status
	}

	err = h.esp32Repo.Create(ctx, device)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to create ESP32 device")
		return
	}

	span.SetAttributes(
		attribute.String("device.id", device.ID.String()),
		attribute.String("device.device_id", device.DeviceID),
		attribute.String("company.id", companyID.String()),
	)

	utils.SuccessResponse(c, http.StatusCreated, "ESP32 device created successfully", device)
}

// GetDevices retrieves ESP32 devices for a company
func (h *ESP32DeviceHandler) GetDevices(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "ESP32DeviceHandler.GetDevices")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Parse filter parameters
	status := c.Query("status")
	vehicleIDStr := c.Query("vehicle_id")

	devices, err := h.esp32Repo.GetByCompany(ctx, *companyID, limit, offset)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve ESP32 devices")
		return
	}

	span.SetAttributes(
		attribute.String("company.id", companyID.String()),
		attribute.Int("devices.count", len(devices)),
	)

	utils.SuccessResponse(c, http.StatusOK, "ESP32 devices retrieved successfully", gin.H{
		"devices": devices,
		"limit":   limit,
		"offset":  offset,
		"count":   len(devices),
		"filters": gin.H{
			"status":     status,
			"vehicle_id": vehicleIDStr,
		},
	})
}

// GetDevice retrieves a specific ESP32 device
func (h *ESP32DeviceHandler) GetDevice(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "ESP32DeviceHandler.GetDevice")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	deviceIDStr := c.Param("id")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid device ID")
		return
	}

	device, err := h.esp32Repo.GetByID(ctx, deviceID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve ESP32 device")
		return
	}

	if device == nil {
		utils.NotFoundResponse(c, "ESP32 device not found")
		return
	}

	span.SetAttributes(
		attribute.String("device.id", device.ID.String()),
		attribute.String("device.device_id", device.DeviceID),
		attribute.String("company.id", companyID.String()),
	)

	utils.SuccessResponse(c, http.StatusOK, "ESP32 device retrieved successfully", device)
}

// UpdateDevice updates an ESP32 device
func (h *ESP32DeviceHandler) UpdateDevice(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "ESP32DeviceHandler.UpdateDevice")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	deviceIDStr := c.Param("id")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid device ID")
		return
	}

	var req models.CreateESP32DeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Get existing device
	device, err := h.esp32Repo.GetByID(ctx, deviceID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve ESP32 device")
		return
	}

	if device == nil {
		utils.NotFoundResponse(c, "ESP32 device not found")
		return
	}

	// Validate vehicle if provided
	if req.VehicleID != nil {
		vehicle, err := h.vehicleRepo.GetByID(ctx, *req.VehicleID, *companyID)
		if err != nil || vehicle == nil {
			utils.BadRequestResponse(c, "Invalid vehicle ID or vehicle does not belong to company")
			return
		}
	}

	// Update device fields
	device.DeviceName = req.DeviceName
	device.DeviceID = req.DeviceID
	device.MACAddress = req.MACAddress
	device.IPAddress = req.IPAddress
	device.FirmwareVersion = req.FirmwareVersion
	device.HardwareRevision = req.HardwareRevision
	device.WifiSSID = req.WifiSSID
	device.VehicleID = req.VehicleID
	device.InstallationDate = req.InstallationDate

	err = h.esp32Repo.Update(ctx, device)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to update ESP32 device")
		return
	}

	span.SetAttributes(
		attribute.String("device.id", device.ID.String()),
		attribute.String("device.device_id", device.DeviceID),
	)

	utils.SuccessResponse(c, http.StatusOK, "ESP32 device updated successfully", device)
}

// DeleteDevice deletes an ESP32 device
func (h *ESP32DeviceHandler) DeleteDevice(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "ESP32DeviceHandler.DeleteDevice")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	deviceIDStr := c.Param("id")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid device ID")
		return
	}

	err = h.esp32Repo.Delete(ctx, deviceID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to delete ESP32 device")
		return
	}

	span.SetAttributes(attribute.String("device.id", deviceID.String()))

	utils.SuccessResponse(c, http.StatusOK, "ESP32 device deleted successfully", nil)
}

// GetDeviceStats retrieves statistics for ESP32 devices
func (h *ESP32DeviceHandler) GetDeviceStats(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "ESP32DeviceHandler.GetDeviceStats")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	stats, err := h.esp32Repo.GetDeviceStatistics(ctx, companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve ESP32 device statistics")
		return
	}

	span.SetAttributes(attribute.String("company.id", companyID.String()))

	utils.SuccessResponse(c, http.StatusOK, "ESP32 device statistics retrieved successfully", stats)
}

// UpdateDeviceStatus updates the status of an ESP32 device
func (h *ESP32DeviceHandler) UpdateDeviceStatus(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "ESP32DeviceHandler.UpdateDeviceStatus")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	deviceIDStr := c.Param("id")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid device ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate status
	validStatuses := []string{"online", "offline", "maintenance", "error"}
	isValid := false
	for _, status := range validStatuses {
		if req.Status == status {
			isValid = true
			break
		}
	}

	if !isValid {
		utils.BadRequestResponse(c, "Invalid status. Must be one of: online, offline, maintenance, error")
		return
	}

	// Update device status manually
	device, err := h.esp32Repo.GetByID(ctx, deviceID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve ESP32 device")
		return
	}

	if device == nil {
		utils.NotFoundResponse(c, "ESP32 device not found")
		return
	}

	device.Status = req.Status
	err = h.esp32Repo.Update(ctx, device)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to update ESP32 device status")
		return
	}

	span.SetAttributes(
		attribute.String("device.id", deviceID.String()),
		attribute.String("device.status", req.Status),
	)

	utils.SuccessResponse(c, http.StatusOK, "ESP32 device status updated successfully", gin.H{
		"device_id":  deviceID,
		"status":     req.Status,
		"updated_at": time.Now(),
	})
}

// GetDeviceByDeviceID retrieves an ESP32 device by its device ID (for ESP32 communication)
func (h *ESP32DeviceHandler) GetDeviceByDeviceID(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "ESP32DeviceHandler.GetDeviceByDeviceID")
	defer span.End()

	deviceID := c.Param("deviceId")
	if deviceID == "" {
		utils.BadRequestResponse(c, "Device ID is required")
		return
	}

	device, err := h.esp32Repo.GetByDeviceID(ctx, deviceID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve ESP32 device")
		return
	}

	if device == nil {
		utils.NotFoundResponse(c, "ESP32 device not found")
		return
	}

	span.SetAttributes(
		attribute.String("device.id", device.ID.String()),
		attribute.String("device.device_id", device.DeviceID),
	)

	utils.SuccessResponse(c, http.StatusOK, "ESP32 device retrieved successfully", device)
}

// RegisterDevice registers a new ESP32 device (simplified endpoint for ESP32 self-registration)
func (h *ESP32DeviceHandler) RegisterDevice(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "ESP32DeviceHandler.RegisterDevice")
	defer span.End()

	var req struct {
		DeviceID         string  `json:"device_id" binding:"required"`
		MacAddress       string  `json:"mac_address" binding:"required"`
		IPAddress        *string `json:"ip_address"`
		FirmwareVersion  *string `json:"firmware_version"`
		HardwareRevision *string `json:"hardware_revision"`
		DeviceName       *string `json:"device_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Check if device already exists
	existingDevice, err := h.esp32Repo.GetByDeviceID(ctx, req.DeviceID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to check device existence")
		return
	}

	if existingDevice != nil {
		// Update existing device with new information
		if req.IPAddress != nil {
			existingDevice.IPAddress = req.IPAddress
		}
		if req.FirmwareVersion != nil {
			existingDevice.FirmwareVersion = req.FirmwareVersion
		}
		if req.HardwareRevision != nil {
			existingDevice.HardwareRevision = req.HardwareRevision
		}
		existingDevice.Status = "online"
		existingDevice.LastHeartbeat = &time.Time{}
		*existingDevice.LastHeartbeat = time.Now()

		err = h.esp32Repo.Update(ctx, existingDevice)
		if err != nil {
			span.RecordError(err)
			utils.InternalServerErrorResponse(c, "Failed to update ESP32 device")
			return
		}

		utils.SuccessResponse(c, http.StatusOK, "ESP32 device updated successfully", existingDevice)
		return
	}

	// Create new device (without company assignment - will need to be assigned later)
	name := req.DeviceID
	if req.DeviceName != nil && *req.DeviceName != "" {
		name = *req.DeviceName
	}

	device := &models.ESP32Device{
		DeviceName:       name,
		DeviceID:         req.DeviceID,
		MACAddress:       &req.MacAddress,
		IPAddress:        req.IPAddress,
		FirmwareVersion:  req.FirmwareVersion,
		HardwareRevision: req.HardwareRevision,
		Status:           "active",
	}

	now := time.Now()
	device.LastHeartbeat = &now

	err = h.esp32Repo.Create(ctx, device)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to register ESP32 device")
		return
	}

	span.SetAttributes(
		attribute.String("device.id", device.ID.String()),
		attribute.String("device.device_id", device.DeviceID),
	)

	utils.SuccessResponse(c, http.StatusCreated, "ESP32 device registered successfully", device)
}

// AssignDeviceToVehicle assigns an ESP32 device to a vehicle
func (h *ESP32DeviceHandler) AssignDeviceToVehicle(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "ESP32DeviceHandler.AssignDeviceToVehicle")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	deviceIDStr := c.Param("id")
	deviceID, err := uuid.Parse(deviceIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid device ID")
		return
	}

	var req struct {
		VehicleID *uuid.UUID `json:"vehicle_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Get existing device
	device, err := h.esp32Repo.GetByID(ctx, deviceID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve ESP32 device")
		return
	}

	if device == nil {
		utils.NotFoundResponse(c, "ESP32 device not found")
		return
	}

	// Validate vehicle if provided
	if req.VehicleID != nil {
		vehicle, err := h.vehicleRepo.GetByID(ctx, *req.VehicleID, *companyID)
		if err != nil || vehicle == nil {
			utils.BadRequestResponse(c, "Invalid vehicle ID or vehicle does not belong to company")
			return
		}
	}

	// Update device vehicle assignment
	device.VehicleID = req.VehicleID

	err = h.esp32Repo.Update(ctx, device)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to assign ESP32 device to vehicle")
		return
	}

	span.SetAttributes(
		attribute.String("device.id", device.ID.String()),
		attribute.String("device.device_id", device.DeviceID),
	)

	utils.SuccessResponse(c, http.StatusOK, "ESP32 device assigned to vehicle successfully", device)
}
