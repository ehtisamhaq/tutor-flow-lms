package handler

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/usecase/reports"
)

// ReportHandler handles reports and recently viewed HTTP requests
type ReportHandler struct {
	reportUC *reports.UseCase
}

// NewReportHandler creates a new report handler
func NewReportHandler(reportUC *reports.UseCase) *ReportHandler {
	return &ReportHandler{reportUC: reportUC}
}

// RegisterRoutes registers report routes
func (h *ReportHandler) RegisterRoutes(g *echo.Group, authMW, adminMW echo.MiddlewareFunc) {
	r := g.Group("/reports", authMW)

	// Recently Viewed
	r.GET("/recently-viewed", h.GetRecentlyViewed)
	r.POST("/recently-viewed/:courseId", h.TrackView)
	r.DELETE("/recently-viewed", h.ClearRecentlyViewed)

	// Exports (Admin only for full reports)
	r.POST("/export", h.ExportData, adminMW)

	// Scheduled Reports
	r.POST("/scheduled", h.CreateScheduledReport)
	r.GET("/scheduled", h.GetMyScheduledReports)
	r.PUT("/scheduled/:id", h.UpdateScheduledReport)
	r.DELETE("/scheduled/:id", h.DeleteScheduledReport)
}

// Recently Viewed

func (h *ReportHandler) TrackView(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	if err := h.reportUC.TrackView(c.Request().Context(), claims.UserID, courseID); err != nil {
		return response.InternalError(c, "Failed to track view")
	}

	return response.Success(c, nil)
}

func (h *ReportHandler) GetRecentlyViewed(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)
	limit := 10
	if l := c.QueryParam("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	items, err := h.reportUC.GetRecentlyViewed(c.Request().Context(), claims.UserID, limit)
	if err != nil {
		return response.InternalError(c, "Failed to get recently viewed")
	}

	return response.Success(c, items)
}

func (h *ReportHandler) ClearRecentlyViewed(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)
	if err := h.reportUC.ClearRecentlyViewed(c.Request().Context(), claims.UserID); err != nil {
		return response.InternalError(c, "Failed to clear recently viewed")
	}
	return response.NoContent(c)
}

// Exports

func (h *ReportHandler) ExportData(c echo.Context) error {
	var req domain.ExportRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	result, err := h.reportUC.ExportData(c.Request().Context(), req)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	c.Response().Header().Set(echo.HeaderContentType, result.ContentType)
	c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+result.Filename)
	return c.Blob(http.StatusOK, result.ContentType, result.Data)
}

// Scheduled Reports

func (h *ReportHandler) CreateScheduledReport(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)
	var report domain.ScheduledReport
	if err := c.Bind(&report); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	created, err := h.reportUC.CreateScheduledReport(c.Request().Context(), claims.UserID, &report)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, created)
}

func (h *ReportHandler) GetMyScheduledReports(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)
	reports, err := h.reportUC.GetMyScheduledReports(c.Request().Context(), claims.UserID)
	if err != nil {
		return response.InternalError(c, "Failed to get scheduled reports")
	}
	return response.Success(c, reports)
}

func (h *ReportHandler) UpdateScheduledReport(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid report ID")
	}

	var update domain.ScheduledReport
	if err := c.Bind(&update); err != nil {
		return response.BadRequest(c, "Invalid request body")
	}

	updated, err := h.reportUC.UpdateScheduledReport(c.Request().Context(), claims.UserID, id, &update)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, updated)
}

func (h *ReportHandler) DeleteScheduledReport(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid report ID")
	}

	if err := h.reportUC.DeleteScheduledReport(c.Request().Context(), claims.UserID, id); err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.NoContent(c)
}
