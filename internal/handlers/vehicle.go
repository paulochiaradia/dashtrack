package handlers

import (
	"net/http"
	"strconv"

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

// VehicleHandler handles vehicle-related HTTP requests
type VehicleHandler struct {
	vehicleRepo *repository.VehicleRepository
	teamRepo    *repository.TeamRepository
	tracer      trace.Tracer
}

// NewVehicleHandler creates a new vehicle handler
func NewVehicleHandler(vehicleRepo *repository.VehicleRepository, teamRepo *repository.TeamRepository) *VehicleHandler {
	return &VehicleHandler{
		vehicleRepo: vehicleRepo,
		teamRepo:    teamRepo,
		tracer:      otel.Tracer("vehicle-handler"),
	}
}

// CreateVehicle creates a new vehicle
func (h *VehicleHandler) CreateVehicle(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "VehicleHandler.CreateVehicle")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	var req models.CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Validate team if provided
	if req.TeamID != nil {
		team, err := h.teamRepo.GetByID(ctx, *req.TeamID, *companyID)
		if err != nil || team == nil {
			utils.BadRequestResponse(c, "Invalid team ID or team does not belong to company")
			return
		}
	}

	vehicle := &models.Vehicle{
		CompanyID:    *companyID,
		LicensePlate: req.LicensePlate,
		Brand:        req.Brand,
		Model:        req.Model,
		Year:         req.Year,
		Color:        req.Color,
		VehicleType:  req.VehicleType,
		FuelType:     req.FuelType,
		CapacityKg:   req.CapacityKg,
		DriverID:     req.DriverID,
		HelperID:     req.HelperID,
		TeamID:       req.TeamID,
		Status:       "active", // Default status
	}

	err = h.vehicleRepo.Create(ctx, vehicle)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to create vehicle")
		return
	}

	span.SetAttributes(
		attribute.String("vehicle.id", vehicle.ID.String()),
		attribute.String("vehicle.license_plate", vehicle.LicensePlate),
		attribute.String("company.id", companyID.String()),
	)

	utils.SuccessResponse(c, http.StatusCreated, "Vehicle created successfully", vehicle)
}

// GetVehicles retrieves vehicles for a company
func (h *VehicleHandler) GetVehicles(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "VehicleHandler.GetVehicles")
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

	// Parse filter parameters (for future use)
	status := c.Query("status")
	teamIDStr := c.Query("team_id")
	vehicleType := c.Query("vehicle_type")

	vehicles, err := h.vehicleRepo.GetByCompany(ctx, *companyID, limit, offset)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve vehicles")
		return
	}

	span.SetAttributes(
		attribute.String("company.id", companyID.String()),
		attribute.Int("vehicles.count", len(vehicles)),
	)

	utils.SuccessResponse(c, http.StatusOK, "Vehicles retrieved successfully", gin.H{
		"vehicles": vehicles,
		"limit":    limit,
		"offset":   offset,
		"count":    len(vehicles),
		"filters": gin.H{
			"status":       status,
			"team_id":      teamIDStr,
			"vehicle_type": vehicleType,
		},
	})
}

// GetVehicle retrieves a specific vehicle
func (h *VehicleHandler) GetVehicle(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "VehicleHandler.GetVehicle")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	vehicleIDStr := c.Param("id")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid vehicle ID")
		return
	}

	vehicle, err := h.vehicleRepo.GetByID(ctx, vehicleID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve vehicle")
		return
	}

	if vehicle == nil {
		utils.NotFoundResponse(c, "Vehicle not found")
		return
	}

	span.SetAttributes(
		attribute.String("vehicle.id", vehicle.ID.String()),
		attribute.String("vehicle.license_plate", vehicle.LicensePlate),
		attribute.String("company.id", companyID.String()),
	)

	utils.SuccessResponse(c, http.StatusOK, "Vehicle retrieved successfully", vehicle)
}

// UpdateVehicle updates a vehicle
func (h *VehicleHandler) UpdateVehicle(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "VehicleHandler.UpdateVehicle")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	vehicleIDStr := c.Param("id")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid vehicle ID")
		return
	}

	var req models.CreateVehicleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Get existing vehicle
	vehicle, err := h.vehicleRepo.GetByID(ctx, vehicleID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve vehicle")
		return
	}

	if vehicle == nil {
		utils.NotFoundResponse(c, "Vehicle not found")
		return
	}

	// Validate team if provided
	if req.TeamID != nil {
		team, err := h.teamRepo.GetByID(ctx, *req.TeamID, *companyID)
		if err != nil || team == nil {
			utils.BadRequestResponse(c, "Invalid team ID or team does not belong to company")
			return
		}
	}

	// Update vehicle fields
	vehicle.LicensePlate = req.LicensePlate
	vehicle.Brand = req.Brand
	vehicle.Model = req.Model
	vehicle.Year = req.Year
	vehicle.Color = req.Color
	vehicle.VehicleType = req.VehicleType
	vehicle.FuelType = req.FuelType
	vehicle.CapacityKg = req.CapacityKg
	vehicle.DriverID = req.DriverID
	vehicle.HelperID = req.HelperID
	vehicle.TeamID = req.TeamID

	err = h.vehicleRepo.Update(ctx, vehicle)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to update vehicle")
		return
	}

	span.SetAttributes(
		attribute.String("vehicle.id", vehicle.ID.String()),
		attribute.String("vehicle.license_plate", vehicle.LicensePlate),
	)

	utils.SuccessResponse(c, http.StatusOK, "Vehicle updated successfully", vehicle)
}

// DeleteVehicle deletes a vehicle
func (h *VehicleHandler) DeleteVehicle(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "VehicleHandler.DeleteVehicle")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	vehicleIDStr := c.Param("id")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid vehicle ID")
		return
	}

	err = h.vehicleRepo.Delete(ctx, vehicleID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to delete vehicle")
		return
	}

	span.SetAttributes(attribute.String("vehicle.id", vehicleID.String()))

	utils.SuccessResponse(c, http.StatusOK, "Vehicle deleted successfully", nil)
}

// GetVehicleStats retrieves statistics for vehicles
func (h *VehicleHandler) GetVehicleStats(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "VehicleHandler.GetVehicleStats")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	// Get basic vehicle count as stats
	vehicles, err := h.vehicleRepo.GetByCompany(ctx, *companyID, 1000, 0)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve vehicle statistics")
		return
	}

	// Calculate basic statistics
	stats := map[string]interface{}{
		"total_vehicles": len(vehicles),
		"active":         0,
		"inactive":       0,
		"maintenance":    0,
	}

	for _, vehicle := range vehicles {
		switch vehicle.Status {
		case "active":
			stats["active"] = stats["active"].(int) + 1
		case "inactive":
			stats["inactive"] = stats["inactive"].(int) + 1
		case "maintenance":
			stats["maintenance"] = stats["maintenance"].(int) + 1
		}
	}

	span.SetAttributes(attribute.String("company.id", companyID.String()))

	utils.SuccessResponse(c, http.StatusOK, "Vehicle statistics retrieved successfully", stats)
}

// AssignVehicleToTeam assigns a vehicle to a team
func (h *VehicleHandler) AssignVehicleToTeam(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "VehicleHandler.AssignVehicleToTeam")
	defer span.End()

	// Get company ID from context
	companyID, err := middleware.GetCompanyIDFromContext(c)
	if err != nil || companyID == nil {
		utils.BadRequestResponse(c, "Company context required")
		return
	}

	vehicleIDStr := c.Param("id")
	vehicleID, err := uuid.Parse(vehicleIDStr)
	if err != nil {
		utils.BadRequestResponse(c, "Invalid vehicle ID")
		return
	}

	var req struct {
		TeamID *uuid.UUID `json:"team_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Get existing vehicle
	vehicle, err := h.vehicleRepo.GetByID(ctx, vehicleID, *companyID)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to retrieve vehicle")
		return
	}

	if vehicle == nil {
		utils.NotFoundResponse(c, "Vehicle not found")
		return
	}

	// Validate team if provided
	if req.TeamID != nil {
		team, err := h.teamRepo.GetByID(ctx, *req.TeamID, *companyID)
		if err != nil || team == nil {
			utils.BadRequestResponse(c, "Invalid team ID or team does not belong to company")
			return
		}
	}

	// Update vehicle team assignment
	vehicle.TeamID = req.TeamID

	err = h.vehicleRepo.Update(ctx, vehicle)
	if err != nil {
		span.RecordError(err)
		utils.InternalServerErrorResponse(c, "Failed to assign vehicle to team")
		return
	}

	span.SetAttributes(
		attribute.String("vehicle.id", vehicle.ID.String()),
		attribute.String("vehicle.license_plate", vehicle.LicensePlate),
	)

	utils.SuccessResponse(c, http.StatusOK, "Vehicle assigned to team successfully", vehicle)
}
