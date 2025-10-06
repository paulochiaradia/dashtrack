package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
	"github.com/paulochiaradia/dashtrack/internal/utils"
)

// CompanyHandler handles company-related HTTP requests
type CompanyHandler struct {
	companyRepo *repository.CompanyRepository
	tracer      trace.Tracer
}

// NewCompanyHandler creates a new company handler
func NewCompanyHandler(companyRepo *repository.CompanyRepository) *CompanyHandler {
	return &CompanyHandler{
		companyRepo: companyRepo,
		tracer:      otel.Tracer("company-handler"),
	}
}

// CreateCompany creates a new company (Master only)
func (h *CompanyHandler) CreateCompany(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "CompanyHandler.CreateCompany")
	defer span.End()

	// Check if user is master
	userContext, exists := c.Get("userContext")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User context not found")
		return
	}

	userCtx := userContext.(*models.UserContext)
	if !userCtx.IsMaster {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden", "Only master users can create companies")
		return
	}

	var req models.CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Check if slug already exists
	exists, err := h.companyRepo.CheckSlugExists(ctx, req.Slug, nil)
	if err != nil {
		span.RecordError(err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", "Failed to check slug availability")
		return
	}
	if exists {
		utils.ErrorResponse(c, http.StatusConflict, "Conflict", "Company slug already exists")
		return
	}

	company := &models.Company{
		Name:             req.Name,
		Slug:             req.Slug,
		Email:            req.Email,
		Phone:            req.Phone,
		Address:          req.Address,
		City:             req.City,
		State:            req.State,
		Country:          req.Country,
		SubscriptionPlan: req.SubscriptionPlan,
	}

	err = h.companyRepo.Create(ctx, company)
	if err != nil {
		span.RecordError(err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", "Failed to create company")
		return
	}

	span.SetAttributes(
		attribute.String("company.id", company.ID.String()),
		attribute.String("company.name", company.Name),
	)

	utils.SuccessResponse(c, http.StatusCreated, "Company created successfully", company)
}

// GetCompanies lists all companies (Master only)
func (h *CompanyHandler) GetCompanies(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "CompanyHandler.GetCompanies")
	defer span.End()

	// Check if user is master
	userContext, exists := c.Get("userContext")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User context not found")
		return
	}

	userCtx := userContext.(*models.UserContext)
	if !userCtx.IsMaster {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden", "Only master users can list all companies")
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	searchTerm := c.Query("search")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	var companies []models.Company
	if searchTerm != "" {
		companies, err = h.companyRepo.Search(ctx, searchTerm, limit, offset)
	} else {
		companies, err = h.companyRepo.List(ctx, limit, offset)
	}

	if err != nil {
		span.RecordError(err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", "Failed to retrieve companies")
		return
	}

	span.SetAttributes(attribute.Int("companies.count", len(companies)))

	utils.SuccessResponse(c, http.StatusOK, "Companies retrieved successfully", gin.H{
		"companies": companies,
		"limit":     limit,
		"offset":    offset,
		"count":     len(companies),
	})
}

// GetCompany retrieves a specific company
func (h *CompanyHandler) GetCompany(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "CompanyHandler.GetCompany")
	defer span.End()

	// Check user context
	userContext, exists := c.Get("userContext")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User context not found")
		return
	}

	userCtx := userContext.(*models.UserContext)
	companyIDStr := c.Param("id")

	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Bad Request", "Invalid company ID")
		return
	}

	// Check if user has access to this company
	if !userCtx.HasCompanyAccess(companyID) {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden", "Access denied to this company")
		return
	}

	company, err := h.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		span.RecordError(err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", "Failed to retrieve company")
		return
	}

	if company == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Not Found", "Company not found")
		return
	}

	span.SetAttributes(attribute.String("company.id", company.ID.String()))

	utils.SuccessResponse(c, http.StatusOK, "Company retrieved successfully", company)
}

// UpdateCompany updates a company
func (h *CompanyHandler) UpdateCompany(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "CompanyHandler.UpdateCompany")
	defer span.End()

	// Check user context
	userContext, exists := c.Get("userContext")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User context not found")
		return
	}

	userCtx := userContext.(*models.UserContext)
	companyIDStr := c.Param("id")

	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Bad Request", "Invalid company ID")
		return
	}

	// Check if user can manage this company
	if !userCtx.HasCompanyAccess(companyID) || !userCtx.CanManageCompany() {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden", "Access denied to manage this company")
		return
	}

	var req models.CreateCompanyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		utils.ValidationErrorResponse(c, err)
		return
	}

	// Get existing company
	company, err := h.companyRepo.GetByID(ctx, companyID)
	if err != nil {
		span.RecordError(err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", "Failed to retrieve company")
		return
	}

	if company == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Not Found", "Company not found")
		return
	}

	// Check if slug already exists (excluding current company)
	if req.Slug != company.Slug {
		exists, err := h.companyRepo.CheckSlugExists(ctx, req.Slug, &companyID)
		if err != nil {
			span.RecordError(err)
			utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", "Failed to check slug availability")
			return
		}
		if exists {
			utils.ErrorResponse(c, http.StatusConflict, "Conflict", "Company slug already exists")
			return
		}
	}

	// Update company fields
	company.Name = req.Name
	company.Slug = req.Slug
	company.Email = req.Email
	company.Phone = req.Phone
	company.Address = req.Address
	company.City = req.City
	company.State = req.State
	company.Country = req.Country

	// Only master can change subscription plan
	if userCtx.IsMaster {
		company.SubscriptionPlan = req.SubscriptionPlan
	}

	err = h.companyRepo.Update(ctx, company)
	if err != nil {
		span.RecordError(err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", "Failed to update company")
		return
	}

	span.SetAttributes(attribute.String("company.id", company.ID.String()))

	utils.SuccessResponse(c, http.StatusOK, "Company updated successfully", company)
}

// DeleteCompany soft deletes a company (Master only)
func (h *CompanyHandler) DeleteCompany(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "CompanyHandler.DeleteCompany")
	defer span.End()

	// Check if user is master
	userContext, exists := c.Get("userContext")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User context not found")
		return
	}

	userCtx := userContext.(*models.UserContext)
	if !userCtx.IsMaster {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden", "Only master users can delete companies")
		return
	}

	companyIDStr := c.Param("id")
	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Bad Request", "Invalid company ID")
		return
	}

	err = h.companyRepo.Delete(ctx, companyID)
	if err != nil {
		span.RecordError(err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", "Failed to delete company")
		return
	}

	span.SetAttributes(attribute.String("company.id", companyID.String()))

	utils.SuccessResponse(c, http.StatusOK, "Company deleted successfully", nil)
}

// GetCompanyStats retrieves company statistics
func (h *CompanyHandler) GetCompanyStats(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "CompanyHandler.GetCompanyStats")
	defer span.End()

	// Check user context
	userContext, exists := c.Get("userContext")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User context not found")
		return
	}

	userCtx := userContext.(*models.UserContext)
	companyIDStr := c.Param("id")

	companyID, err := uuid.Parse(companyIDStr)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Bad Request", "Invalid company ID")
		return
	}

	// Check if user has access to this company
	if !userCtx.HasCompanyAccess(companyID) {
		utils.ErrorResponse(c, http.StatusForbidden, "Forbidden", "Access denied to this company")
		return
	}

	stats, err := h.companyRepo.GetCompanyStats(ctx, companyID)
	if err != nil {
		span.RecordError(err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", "Failed to retrieve company statistics")
		return
	}

	span.SetAttributes(attribute.String("company.id", companyID.String()))

	utils.SuccessResponse(c, http.StatusOK, "Company statistics retrieved successfully", stats)
}

// GetMyCompany retrieves the current user's company
func (h *CompanyHandler) GetMyCompany(c *gin.Context) {
	ctx, span := h.tracer.Start(c.Request.Context(), "CompanyHandler.GetMyCompany")
	defer span.End()

	// Check user context
	userContext, exists := c.Get("userContext")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Unauthorized", "User context not found")
		return
	}

	userCtx := userContext.(*models.UserContext)
	if userCtx.CompanyID == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Not Found", "User is not associated with any company")
		return
	}

	company, err := h.companyRepo.GetByID(ctx, *userCtx.CompanyID)
	if err != nil {
		span.RecordError(err)
		utils.ErrorResponse(c, http.StatusInternalServerError, "Internal Server Error", "Failed to retrieve company")
		return
	}

	if company == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Not Found", "Company not found")
		return
	}

	// Get company statistics
	stats, err := h.companyRepo.GetCompanyStats(ctx, *userCtx.CompanyID)
	if err != nil {
		span.RecordError(err)
		// Don't fail the request if stats retrieval fails
		stats = &models.CompanyStats{}
	}

	span.SetAttributes(attribute.String("company.id", company.ID.String()))

	utils.SuccessResponse(c, http.StatusOK, "Company retrieved successfully", gin.H{
		"company": company,
		"stats":   stats,
	})
}
