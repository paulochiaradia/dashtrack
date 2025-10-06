package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/paulochiaradia/dashtrack/internal/logger"
	"github.com/paulochiaradia/dashtrack/internal/models"
	"github.com/paulochiaradia/dashtrack/internal/repository"
)

// DashboardHandler handles dashboard-related requests
type DashboardHandler struct {
	userRepo    repository.UserRepositoryInterface
	authLogRepo repository.AuthLogRepositoryInterface
	sessionRepo repository.SessionRepositoryInterface
	companyRepo repository.CompanyRepositoryInterface
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(
	userRepo repository.UserRepositoryInterface,
	authLogRepo repository.AuthLogRepositoryInterface,
	sessionRepo repository.SessionRepositoryInterface,
	companyRepo repository.CompanyRepositoryInterface,
) *DashboardHandler {
	return &DashboardHandler{
		userRepo:    userRepo,
		authLogRepo: authLogRepo,
		sessionRepo: sessionRepo,
		companyRepo: companyRepo,
	}
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	// User Statistics
	TotalUsers    int `json:"total_users"`
	ActiveUsers   int `json:"active_users"`
	InactiveUsers int `json:"inactive_users"`
	OnlineUsers   int `json:"online_users"`

	// Login Statistics
	TotalLogins      int `json:"total_logins"`
	SuccessfulLogins int `json:"successful_logins"`
	FailedLogins     int `json:"failed_logins"`
	TodayLogins      int `json:"today_logins"`
	WeekLogins       int `json:"week_logins"`
	MonthLogins      int `json:"month_logins"`

	// Session Statistics
	ActiveSessions         int     `json:"active_sessions"`
	AverageSessionDuration float64 `json:"average_session_duration_minutes"`

	// Company Statistics (for master only)
	TotalCompanies  *int `json:"total_companies,omitempty"`
	ActiveCompanies *int `json:"active_companies,omitempty"`

	// Time Range
	DateRange struct {
		From time.Time `json:"from"`
		To   time.Time `json:"to"`
	} `json:"date_range"`
}

// UserActivity represents user activity data
type UserActivity struct {
	UserID       uuid.UUID  `json:"user_id"`
	UserName     string     `json:"user_name"`
	UserEmail    string     `json:"user_email"`
	Role         string     `json:"role"`
	CompanyID    *uuid.UUID `json:"company_id,omitempty"`
	CompanyName  *string    `json:"company_name,omitempty"`
	LastLogin    *time.Time `json:"last_login"`
	IsOnline     bool       `json:"is_online"`
	SessionCount int        `json:"session_count"`
	LoginCount   int        `json:"login_count"`
	FailedLogins int        `json:"failed_logins"`
	TotalTime    int        `json:"total_time_minutes"` // Total logged time in minutes
}

// DashboardResponse represents the complete dashboard response
type DashboardResponse struct {
	Stats          DashboardStats       `json:"stats"`
	UserActivities []UserActivity       `json:"user_activities"`
	RecentLogins   []models.RecentLogin `json:"recent_logins"`
	TopUsers       []UserActivity       `json:"top_users"` // Most active users
}

// GetDashboard returns dashboard data based on user role
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	// Get user context
	userContext, exists := c.Get("userContext")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not found"})
		return
	}

	ctx := userContext.(*models.UserContext)

	// Get date range from query parameters (optional)
	days := c.DefaultQuery("days", "30") // Default to last 30 days
	daysInt, err := strconv.Atoi(days)
	if err != nil || daysInt <= 0 {
		daysInt = 30
	}

	to := time.Now()
	from := to.AddDate(0, 0, -daysInt)

	var response DashboardResponse
	var dashboardErr error

	// Route to appropriate dashboard based on role
	switch ctx.Role {
	case "master":
		response, dashboardErr = h.getMasterDashboard(c.Request.Context(), from, to)
	case "admin", "company_admin":
		if ctx.CompanyID == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Company ID required for company admin dashboard"})
			return
		}
		response, dashboardErr = h.getCompanyDashboard(c.Request.Context(), *ctx.CompanyID, from, to)
	case "driver", "helper":
		response, dashboardErr = h.getUserDashboard(c.Request.Context(), ctx.UserID, from, to)
	default:
		c.JSON(http.StatusForbidden, gin.H{"error": "Dashboard access not allowed for this role"})
		return
	}

	if dashboardErr != nil {
		logger.Error("Failed to get dashboard data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get dashboard data"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// getMasterDashboard returns dashboard data for master user (all system data)
func (h *DashboardHandler) getMasterDashboard(ctx context.Context, from, to time.Time) (DashboardResponse, error) {
	// Get all users statistics
	totalUsers, err := h.userRepo.CountUsers(ctx, nil) // nil = all companies
	if err != nil {
		return DashboardResponse{}, err
	}

	activeUsers, err := h.userRepo.CountActiveUsers(ctx, nil)
	if err != nil {
		return DashboardResponse{}, err
	}

	// Get login statistics
	totalLogins, err := h.authLogRepo.CountLogins(ctx, nil, from, to)
	if err != nil {
		return DashboardResponse{}, err
	}

	successfulLogins, err := h.authLogRepo.CountSuccessfulLogins(ctx, nil, from, to)
	if err != nil {
		return DashboardResponse{}, err
	}

	failedLogins, err := h.authLogRepo.CountFailedLogins(ctx, nil, from, to)
	if err != nil {
		return DashboardResponse{}, err
	}

	// Get company statistics
	totalCompanies, err := h.companyRepo.CountCompanies(ctx)
	if err != nil {
		return DashboardResponse{}, err
	}

	activeCompanies, err := h.companyRepo.CountActiveCompanies(ctx)
	if err != nil {
		return DashboardResponse{}, err
	}

	// Get user activities (all users)
	userActivities, err := h.getUserActivities(ctx, nil, from, to, 50) // Top 50 users
	if err != nil {
		return DashboardResponse{}, err
	}

	// Get recent logins (all users)
	recentLogins, err := h.getRecentLogins(ctx, nil, from, to, 20) // Last 20 logins
	if err != nil {
		return DashboardResponse{}, err
	}

	// Build response
	stats := DashboardStats{
		TotalUsers:       totalUsers,
		ActiveUsers:      activeUsers,
		InactiveUsers:    totalUsers - activeUsers,
		TotalLogins:      totalLogins,
		SuccessfulLogins: successfulLogins,
		FailedLogins:     failedLogins,
		TotalCompanies:   &totalCompanies,
		ActiveCompanies:  &activeCompanies,
	}
	stats.DateRange.From = from
	stats.DateRange.To = to

	return DashboardResponse{
		Stats:          stats,
		UserActivities: userActivities,
		RecentLogins:   recentLogins,
		TopUsers:       userActivities[:min(10, len(userActivities))], // Top 10 most active
	}, nil
}

// getCompanyDashboard returns dashboard data for company admin (only their company data)
func (h *DashboardHandler) getCompanyDashboard(ctx context.Context, companyID uuid.UUID, from, to time.Time) (DashboardResponse, error) {
	// Get company users statistics
	totalUsers, err := h.userRepo.CountUsers(ctx, &companyID)
	if err != nil {
		return DashboardResponse{}, err
	}

	activeUsers, err := h.userRepo.CountActiveUsers(ctx, &companyID)
	if err != nil {
		return DashboardResponse{}, err
	}

	// Get login statistics for company users
	totalLogins, err := h.authLogRepo.CountLogins(ctx, &companyID, from, to)
	if err != nil {
		return DashboardResponse{}, err
	}

	successfulLogins, err := h.authLogRepo.CountSuccessfulLogins(ctx, &companyID, from, to)
	if err != nil {
		return DashboardResponse{}, err
	}

	failedLogins, err := h.authLogRepo.CountFailedLogins(ctx, &companyID, from, to)
	if err != nil {
		return DashboardResponse{}, err
	}

	// Get user activities (company users only)
	userActivities, err := h.getUserActivities(ctx, &companyID, from, to, 50)
	if err != nil {
		return DashboardResponse{}, err
	}

	// Get recent logins (company users only)
	recentLogins, err := h.getRecentLogins(ctx, &companyID, from, to, 20)
	if err != nil {
		return DashboardResponse{}, err
	}

	// Build response
	stats := DashboardStats{
		TotalUsers:       totalUsers,
		ActiveUsers:      activeUsers,
		InactiveUsers:    totalUsers - activeUsers,
		TotalLogins:      totalLogins,
		SuccessfulLogins: successfulLogins,
		FailedLogins:     failedLogins,
		// No company stats for company admin
	}
	stats.DateRange.From = from
	stats.DateRange.To = to

	return DashboardResponse{
		Stats:          stats,
		UserActivities: userActivities,
		RecentLogins:   recentLogins,
		TopUsers:       userActivities[:min(5, len(userActivities))], // Top 5 most active
	}, nil
}

// getUserDashboard returns dashboard data for individual user (only their own data)
func (h *DashboardHandler) getUserDashboard(ctx context.Context, userID uuid.UUID, from, to time.Time) (DashboardResponse, error) {
	// Get user's own statistics
	totalLogins, err := h.authLogRepo.CountUserLogins(ctx, userID, from, to)
	if err != nil {
		return DashboardResponse{}, err
	}

	successfulLogins, err := h.authLogRepo.CountUserSuccessfulLogins(ctx, userID, from, to)
	if err != nil {
		return DashboardResponse{}, err
	}

	failedLogins, err := h.authLogRepo.CountUserFailedLogins(ctx, userID, from, to)
	if err != nil {
		return DashboardResponse{}, err
	}

	// Get user's activity data
	userActivity, err := h.getUserActivity(ctx, userID, from, to)
	if err != nil {
		return DashboardResponse{}, err
	}

	// Get user's recent logins
	recentLogins, err := h.getUserRecentLogins(ctx, userID, from, to, 20)
	if err != nil {
		return DashboardResponse{}, err
	}

	// Build response with only user's data
	stats := DashboardStats{
		TotalUsers:       1, // Only themselves
		ActiveUsers:      1,
		InactiveUsers:    0,
		TotalLogins:      totalLogins,
		SuccessfulLogins: successfulLogins,
		FailedLogins:     failedLogins,
	}
	stats.DateRange.From = from
	stats.DateRange.To = to

	userActivities := []UserActivity{}
	if userActivity != nil {
		userActivities = append(userActivities, *userActivity)
	}

	return DashboardResponse{
		Stats:          stats,
		UserActivities: userActivities,
		RecentLogins:   recentLogins,
		TopUsers:       userActivities, // Only themselves
	}, nil
}

// Helper function to get min of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getUserActivities gets user activity data for a company or all users
func (h *DashboardHandler) getUserActivities(ctx context.Context, companyID *uuid.UUID, from, to time.Time, limit int) ([]UserActivity, error) {
	// Get recent logins as user activities
	recentLogins, err := h.getRecentLogins(ctx, companyID, from, to, limit)
	if err != nil {
		return nil, err
	}

	// Convert recent logins to user activities
	activities := make([]UserActivity, len(recentLogins))
	for i, login := range recentLogins {
		activities[i] = UserActivity{
			UserID:       login.UserID,
			UserName:     login.UserName,
			UserEmail:    login.UserEmail,
			Role:         "", // Could be fetched separately if needed
			CompanyID:    login.CompanyID,
			CompanyName:  login.CompanyName,
			LastLogin:    &login.LoginTime,
			IsOnline:     false, // Would need session data to determine
			SessionCount: 0,     // Would need session data
			LoginCount:   1,     // This represents one login
			FailedLogins: 0,     // Would need to query separately
			TotalTime:    0,     // Would need session duration data
		}
	}

	return activities, nil
}

// getRecentLogins gets recent login data for a company or all users
func (h *DashboardHandler) getRecentLogins(ctx context.Context, companyID *uuid.UUID, from, to time.Time, limit int) ([]models.RecentLogin, error) {
	// Get successful logins from auth_logs with user information
	recentLogins, err := h.authLogRepo.GetRecentSuccessfulLogins(ctx, companyID, from, to, limit)
	if err != nil {
		return nil, err
	}

	return recentLogins, nil
}

// getUserActivity gets activity data for a specific user
func (h *DashboardHandler) getUserActivity(ctx context.Context, userID uuid.UUID, from, to time.Time) (*UserActivity, error) {
	// Get user information
	user, err := h.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get user's login count in the period
	loginCount, err := h.authLogRepo.CountUserLogins(ctx, userID, from, to)
	if err != nil {
		return nil, err
	}

	return &UserActivity{
		UserID:       userID,
		UserName:     user.Name,
		UserEmail:    user.Email,
		Role:         "", // Could add role info if needed
		CompanyID:    user.CompanyID,
		LoginCount:   loginCount,
		FailedLogins: 0, // Could query separately
		TotalTime:    0, // Would need session data
	}, nil
}

// getUserRecentLogins gets recent login data for a specific user
func (h *DashboardHandler) getUserRecentLogins(ctx context.Context, userID uuid.UUID, from, to time.Time, limit int) ([]models.RecentLogin, error) {
	// Get recent logins for the specific user
	recentLogins, err := h.authLogRepo.GetUserRecentSuccessfulLogins(ctx, userID, from, to, limit)
	if err != nil {
		return nil, err
	}

	return recentLogins, nil
}
