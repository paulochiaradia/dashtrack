package routes

import (
	"github.com/gin-gonic/gin"
)

// setupVehicleRoutes configures vehicle management routes
func (r *Router) setupVehicleRoutes(api *gin.RouterGroup) {
	// Company Admin vehicle routes (full CRUD)
	companyAdmin := api.Group("/company-admin/vehicles")
	companyAdmin.Use(r.authMiddleware.RequireAuth())
	companyAdmin.Use(r.authMiddleware.RequireRole("company_admin"))
	{
		companyAdmin.POST("", r.vehicleHandler.CreateVehicle)                                     // Create vehicle
		companyAdmin.GET("", r.vehicleHandler.GetVehicles)                                        // List vehicles
		companyAdmin.GET("/:id", r.vehicleHandler.GetVehicle)                                     // Get vehicle details
		companyAdmin.PUT("/:id", r.vehicleHandler.UpdateVehicle)                                  // Update vehicle
		companyAdmin.DELETE("/:id", r.vehicleHandler.DeleteVehicle)                               // Delete vehicle (soft delete)
		companyAdmin.PUT("/:id/assign", r.vehicleHandler.AssignUsers)                             // Assign driver/helper
		companyAdmin.GET("/:id/assignment-history", r.vehicleHandler.GetVehicleAssignmentHistory) // Get assignment history
	}

	// Admin vehicle routes (read-only + assign)
	admin := api.Group("/admin/vehicles")
	admin.Use(r.authMiddleware.RequireAuth())
	admin.Use(r.authMiddleware.RequireRole("admin"))
	{
		admin.GET("", r.vehicleHandler.GetVehicles)                                        // List vehicles
		admin.GET("/:id", r.vehicleHandler.GetVehicle)                                     // Get vehicle details
		admin.PUT("/:id/assign", r.vehicleHandler.AssignUsers)                             // Assign driver/helper
		admin.GET("/:id/assignment-history", r.vehicleHandler.GetVehicleAssignmentHistory) // Get assignment history
	}

	// Manager vehicle routes (read-only for their teams)
	manager := api.Group("/manager/vehicles")
	manager.Use(r.authMiddleware.RequireAuth())
	manager.Use(r.authMiddleware.RequireRole("manager"))
	{
		manager.GET("", r.vehicleHandler.GetVehicles)    // List vehicles (filtered by manager's teams)
		manager.GET("/:id", r.vehicleHandler.GetVehicle) // Get vehicle details
	}

	// Driver/Assistant routes (read-only for assigned vehicles)
	user := api.Group("/vehicles")
	user.Use(r.authMiddleware.RequireAuth())
	{
		user.GET("/my-vehicle", r.vehicleHandler.GetMyVehicle) // Get vehicle assigned to current user
	}
}
