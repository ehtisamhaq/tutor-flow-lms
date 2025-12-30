package handler

import (
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/usecase/admin"
)

// AdminHandler handles admin dashboard HTTP requests
type AdminHandler struct {
	adminUC *admin.UseCase
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(adminUC *admin.UseCase) *AdminHandler {
	return &AdminHandler{adminUC: adminUC}
}

// RegisterRoutes registers admin dashboard routes
func (h *AdminHandler) RegisterRoutes(g *echo.Group, authMW, adminMW echo.MiddlewareFunc) {
	a := g.Group("/admin", authMW, adminMW)
	a.GET("/dashboard", h.GetDashboard)
	a.GET("/revenue-chart", h.GetRevenueChart)
	a.GET("/top-courses", h.GetTopCourses)
	a.GET("/top-instructors", h.GetTopInstructors)
	a.GET("/recent-orders", h.GetRecentOrders)
	a.GET("/recent-users", h.GetRecentUsers)
	a.GET("/system-health", h.GetSystemHealth)
}

// GetDashboard godoc
// @Summary Get dashboard statistics
// @Tags Admin
// @Security BearerAuth
// @Success 200 {object} response.Response{data=admin.DashboardStats}
// @Router /admin/dashboard [get]
func (h *AdminHandler) GetDashboard(c echo.Context) error {
	stats, err := h.adminUC.GetDashboardStats(c.Request().Context())
	if err != nil {
		return response.InternalError(c, "Failed to get dashboard stats")
	}

	return response.Success(c, stats)
}

// GetRevenueChart godoc
// @Summary Get revenue chart data
// @Tags Admin
// @Security BearerAuth
// @Param period query string false "Period (week, month, year)"
// @Success 200 {object} response.Response{data=admin.RevenueChartData}
// @Router /admin/revenue-chart [get]
func (h *AdminHandler) GetRevenueChart(c echo.Context) error {
	period := c.QueryParam("period")
	if period == "" {
		period = "month"
	}

	data, err := h.adminUC.GetRevenueChart(c.Request().Context(), period)
	if err != nil {
		return response.InternalError(c, "Failed to get revenue chart")
	}

	return response.Success(c, data)
}

// GetTopCourses godoc
// @Summary Get top performing courses
// @Tags Admin
// @Security BearerAuth
// @Param limit query int false "Limit (default 10)"
// @Param sort_by query string false "Sort by (revenue, rating, enrollments)"
// @Success 200 {object} response.Response{data=[]admin.TopCourse}
// @Router /admin/top-courses [get]
func (h *AdminHandler) GetTopCourses(c echo.Context) error {
	limit := 10
	if l := c.QueryParam("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}
	sortBy := c.QueryParam("sort_by")

	courses, err := h.adminUC.GetTopCourses(c.Request().Context(), limit, sortBy)
	if err != nil {
		return response.InternalError(c, "Failed to get top courses")
	}

	return response.Success(c, courses)
}

// GetTopInstructors godoc
// @Summary Get top performing instructors
// @Tags Admin
// @Security BearerAuth
// @Param limit query int false "Limit (default 10)"
// @Success 200 {object} response.Response{data=[]admin.TopInstructor}
// @Router /admin/top-instructors [get]
func (h *AdminHandler) GetTopInstructors(c echo.Context) error {
	limit := 10
	if l := c.QueryParam("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	instructors, err := h.adminUC.GetTopInstructors(c.Request().Context(), limit)
	if err != nil {
		return response.InternalError(c, "Failed to get top instructors")
	}

	return response.Success(c, instructors)
}

// GetRecentOrders godoc
// @Summary Get recent orders
// @Tags Admin
// @Security BearerAuth
// @Param limit query int false "Limit (default 10)"
// @Success 200 {object} response.Response{data=[]admin.RecentOrder}
// @Router /admin/recent-orders [get]
func (h *AdminHandler) GetRecentOrders(c echo.Context) error {
	limit := 10
	if l := c.QueryParam("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	orders, err := h.adminUC.GetRecentOrders(c.Request().Context(), limit)
	if err != nil {
		return response.InternalError(c, "Failed to get recent orders")
	}

	return response.Success(c, orders)
}

// GetRecentUsers godoc
// @Summary Get recently registered users
// @Tags Admin
// @Security BearerAuth
// @Param limit query int false "Limit (default 10)"
// @Success 200 {object} response.Response{data=[]admin.RecentUser}
// @Router /admin/recent-users [get]
func (h *AdminHandler) GetRecentUsers(c echo.Context) error {
	limit := 10
	if l := c.QueryParam("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	users, err := h.adminUC.GetRecentUsers(c.Request().Context(), limit)
	if err != nil {
		return response.InternalError(c, "Failed to get recent users")
	}

	return response.Success(c, users)
}

// GetSystemHealth godoc
// @Summary Get system health status
// @Tags Admin
// @Security BearerAuth
// @Success 200 {object} response.Response{data=admin.SystemHealth}
// @Router /admin/system-health [get]
func (h *AdminHandler) GetSystemHealth(c echo.Context) error {
	health, err := h.adminUC.GetSystemHealth(c.Request().Context())
	if err != nil {
		return response.InternalError(c, "Failed to get system health")
	}

	return response.Success(c, health)
}
