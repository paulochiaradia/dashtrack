package routes

import (
	"github.com/paulochiaradia/dashtrack/internal/middleware"
)

func (r *Router) setupMultiTenantRoutes() {
	// Create Gin middleware from auth middleware
	authMiddleware := middleware.NewGinAuthMiddleware(r.jwtManager)

	// Multi-tenant API routes - require authentication and company context
	api := r.engine.Group("/api/v1")
	api.Use(authMiddleware.RequireAuth())

	// Master-only routes (for business and system ownership)
	// Master = System Owner (business operations)
	master := api.Group("/master")
	master.Use(authMiddleware.RequireMasterRole())
	{
		// Company management (business operations - master only)
		master.POST("/companies", r.companyHandler.CreateCompany)
		master.GET("/companies", r.companyHandler.GetCompanies)
		master.GET("/companies/:id", r.companyHandler.GetCompany)
		master.PUT("/companies/:id", r.companyHandler.UpdateCompany)
		master.DELETE("/companies/:id", r.companyHandler.DeleteCompany)
		master.GET("/companies/:id/stats", r.companyHandler.GetCompanyStats)

		// Business Analytics & Dashboard (master only)
		master.GET("/dashboard", r.dashboardHandler.GetDashboard)
		master.GET("/analytics/global", r.companyHandler.GetCompanies) // TODO: implement global analytics

		// User Management with Business Context (master can manage ALL users across companies)
		master.GET("/users", r.userHandler.GetUsers)
		master.POST("/users", r.userHandler.CreateUser)
		master.GET("/users/:id", r.userHandler.GetUserByID)
		master.PUT("/users/:id", r.userHandler.UpdateUser)
		master.DELETE("/users/:id", r.userHandler.DeleteUser)

		// Billing & Business Operations (master only)
		// TODO: implement billing handlers
		// master.GET("/billing", r.billingHandler.GetBilling)
		// master.POST("/billing/plans", r.billingHandler.CreatePlan)
		// master.GET("/revenue", r.billingHandler.GetRevenue)

		// System-wide Configuration (business policies - master only)
		// TODO: implement business config handlers
		// master.GET("/policies", r.policyHandler.GetPolicies)
		// master.PUT("/policies", r.policyHandler.UpdatePolicies)
	}

	// Company-scoped routes - require company access
	company := api.Group("/company")
	company.Use(middleware.RequireCompanyAccess())
	{
		// Company info (read-only for regular users)
		company.GET("/info", r.companyHandler.GetMyCompany)
		// Using existing method for dashboard
		company.GET("/dashboard", r.companyHandler.GetMyCompany)

		// Company admin routes (require company admin role)
		companyAdmin := company.Group("/admin")
		companyAdmin.Use(middleware.RequireCompanyAdmin())
		{
			// Using existing update method
			companyAdmin.PUT("/info", r.companyHandler.UpdateCompany)
			// Using existing stats method
			companyAdmin.GET("/stats", r.companyHandler.GetCompanyStats)
		}

		// Team management
		teams := company.Group("/teams")
		{
			teams.GET("", r.teamHandler.GetTeams)
			teams.GET("/:id", r.teamHandler.GetTeam)
			teams.GET("/:id/members", r.teamHandler.GetMembers)
		}

		// Team admin routes (require company admin or team manager role)
		teamsAdmin := teams.Group("")
		teamsAdmin.Use(middleware.RequireCompanyAdmin()) // TODO: Add team manager check
		{
			teamsAdmin.POST("", r.teamHandler.CreateTeam)
			teamsAdmin.PUT("/:id", r.teamHandler.UpdateTeam)
			teamsAdmin.DELETE("/:id", r.teamHandler.DeleteTeam)
			teamsAdmin.POST("/:id/members", r.teamHandler.AddMember)
			teamsAdmin.DELETE("/:id/members/:userId", r.teamHandler.RemoveMember)
		}

		// Vehicle management
		vehicles := company.Group("/vehicles")
		{
			vehicles.GET("", r.vehicleHandler.GetVehicles)
			vehicles.GET("/:id", r.vehicleHandler.GetVehicle)
			vehicles.GET("/stats", r.vehicleHandler.GetVehicleStats)
		}

		// Vehicle admin routes (require company admin role)
		vehiclesAdmin := vehicles.Group("")
		vehiclesAdmin.Use(middleware.RequireCompanyAdmin())
		{
			vehiclesAdmin.POST("", r.vehicleHandler.CreateVehicle)
			vehiclesAdmin.PUT("/:id", r.vehicleHandler.UpdateVehicle)
			vehiclesAdmin.DELETE("/:id", r.vehicleHandler.DeleteVehicle)
			vehiclesAdmin.POST("/:id/assign-team", r.vehicleHandler.AssignVehicleToTeam)
		}

		// ESP32 Device management
		devices := company.Group("/devices")
		{
			devices.GET("", r.esp32Handler.GetDevices)
			devices.GET("/:id", r.esp32Handler.GetDevice)
			devices.GET("/stats", r.esp32Handler.GetDeviceStats)
		}

		// ESP32 Device admin routes (require company admin role)
		devicesAdmin := devices.Group("")
		devicesAdmin.Use(middleware.RequireCompanyAdmin())
		{
			devicesAdmin.POST("", r.esp32Handler.CreateDevice)
			devicesAdmin.PUT("/:id", r.esp32Handler.UpdateDevice)
			devicesAdmin.DELETE("/:id", r.esp32Handler.DeleteDevice)
			devicesAdmin.PUT("/:id/status", r.esp32Handler.UpdateDeviceStatus)
			devicesAdmin.POST("/:id/assign-vehicle", r.esp32Handler.AssignDeviceToVehicle)
		}
	}
}

func (r *Router) setupESP32PublicRoutes() {
	// Public ESP32 routes (for device communication)
	esp32 := r.engine.Group("/api/v1/esp32")
	{
		// Device registration and status
		esp32.POST("/register", r.esp32Handler.RegisterDevice)
		esp32.GET("/device/:deviceId", r.esp32Handler.GetDeviceByDeviceID)
		esp32.PUT("/device/:deviceId/status", r.esp32Handler.UpdateDeviceStatus)
	}
}
