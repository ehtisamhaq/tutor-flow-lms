package handler

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tutorflow/tutorflow-server/internal/domain"
	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/usecase/certificate"
)

var _ = domain.Certificate{}

// CertificateHandler handles certificate HTTP requests
type CertificateHandler struct {
	certUC *certificate.UseCase
}

// NewCertificateHandler creates a new certificate handler
func NewCertificateHandler(certUC *certificate.UseCase) *CertificateHandler {
	return &CertificateHandler{certUC: certUC}
}

// RegisterRoutes registers certificate routes
func (h *CertificateHandler) RegisterRoutes(g *echo.Group, authMW echo.MiddlewareFunc) {
	certs := g.Group("/certificates")
	certs.GET("/verify/:number", h.VerifyCertificate) // Public endpoint
	certs.GET("/my", h.GetMyCertificates, authMW)
	certs.GET("/:id", h.GetCertificate, authMW)
	certs.POST("/request/:courseId", h.RequestCertificate, authMW)
	certs.GET("/:id/data", h.GetCertificateData, authMW)
}

// VerifyCertificate godoc
// @Summary Verify a certificate (public)
// @Tags Certificates
// @Param number path string true "Certificate Number"
// @Success 200 {object} response.Response{data=certificate.CertificateVerification}
// @Router /certificates/verify/{number} [get]
func (h *CertificateHandler) VerifyCertificate(c echo.Context) error {
	number := c.Param("number")
	if number == "" {
		return response.BadRequest(c, "Certificate number required")
	}

	verification, err := h.certUC.VerifyCertificate(c.Request().Context(), number)
	if err != nil {
		return response.NotFound(c, "Certificate not found or invalid")
	}

	return response.Success(c, verification)
}

// GetMyCertificates godoc
// @Summary Get my certificates
// @Tags Certificates
// @Security BearerAuth
// @Success 200 {object} response.Response
// @Router /certificates/my [get]
func (h *CertificateHandler) GetMyCertificates(c echo.Context) error {
	claims, _ := middleware.GetClaims(c)

	page, limit := 1, 20
	certs, total, err := h.certUC.GetMyCertificates(c.Request().Context(), claims.UserID, page, limit)
	if err != nil {
		return response.InternalError(c, "Failed to get certificates")
	}

	return response.Paginated(c, certs, page, limit, total)
}

// GetCertificate godoc
// @Summary Get certificate by ID
// @Tags Certificates
// @Security BearerAuth
// @Param id path string true "Certificate ID"
// @Success 200 {object} response.Response{data=domain.Certificate}
// @Router /certificates/{id} [get]
func (h *CertificateHandler) GetCertificate(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid certificate ID")
	}

	cert, err := h.certUC.GetCertificate(c.Request().Context(), id)
	if err != nil || cert == nil {
		return response.NotFound(c, "Certificate not found")
	}

	return response.Success(c, cert)
}

// RequestCertificate godoc
// @Summary Request certificate for a completed course
// @Tags Certificates
// @Security BearerAuth
// @Param courseId path string true "Course ID"
// @Success 201 {object} response.Response{data=domain.Certificate}
// @Router /certificates/request/{courseId} [post]
func (h *CertificateHandler) RequestCertificate(c echo.Context) error {
	courseID, err := uuid.Parse(c.Param("courseId"))
	if err != nil {
		return response.BadRequest(c, "Invalid course ID")
	}

	claims, _ := middleware.GetClaims(c)

	cert, err := h.certUC.RequestCertificate(c.Request().Context(), claims.UserID, courseID)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Created(c, cert)
}

// GetCertificateData godoc
// @Summary Get certificate data for PDF generation
// @Tags Certificates
// @Security BearerAuth
// @Param id path string true "Certificate ID"
// @Success 200 {object} response.Response{data=certificate.CertificateData}
// @Router /certificates/{id}/data [get]
func (h *CertificateHandler) GetCertificateData(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return response.BadRequest(c, "Invalid certificate ID")
	}

	// Get base URL for verification link
	baseURL := c.Scheme() + "://" + c.Request().Host + "/api/v1/certificates"

	data, err := h.certUC.GetCertificateData(c.Request().Context(), id, baseURL)
	if err != nil {
		return response.NotFound(c, "Certificate not found")
	}

	return response.Success(c, data)
}
