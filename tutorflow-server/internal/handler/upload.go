package handler

import (
	"github.com/labstack/echo/v4"

	"github.com/tutorflow/tutorflow-server/internal/middleware"
	"github.com/tutorflow/tutorflow-server/internal/pkg/response"
	"github.com/tutorflow/tutorflow-server/internal/service/storage"
)

// UploadHandler handles file uploads
type UploadHandler struct {
	storageSvc *storage.Service
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(storageSvc *storage.Service) *UploadHandler {
	return &UploadHandler{storageSvc: storageSvc}
}

// RegisterRoutes registers upload routes
func (h *UploadHandler) RegisterRoutes(g *echo.Group, authMW echo.MiddlewareFunc) {
	g.POST("/image", h.UploadImage, authMW)
	g.POST("/video", h.UploadVideo, authMW)
	g.POST("/document", h.UploadDocument, authMW)
}

// UploadImage godoc
// @Summary Upload image
// @Tags Uploads
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param file formance file true "Image file"
// @Param folder formance string false "Folder name"
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /uploads/image [post]
func (h *UploadHandler) UploadImage(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "No file provided")
	}

	folder := c.FormValue("folder")
	if folder == "" {
		folder = "images"
	}

	claims, _ := middleware.GetClaims(c)
	_ = claims // Could log who uploaded

	url, err := h.storageSvc.UploadImage(c.Request().Context(), file, folder)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, map[string]string{
		"url":      url,
		"filename": file.Filename,
	})
}

// UploadVideo godoc
// @Summary Upload video
// @Tags Uploads
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param file formance file true "Video file"
// @Param folder formance string false "Folder name"
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /uploads/video [post]
func (h *UploadHandler) UploadVideo(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "No file provided")
	}

	folder := c.FormValue("folder")
	if folder == "" {
		folder = "videos"
	}

	url, err := h.storageSvc.UploadVideo(c.Request().Context(), file, folder)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, map[string]string{
		"url":      url,
		"filename": file.Filename,
	})
}

// UploadDocument godoc
// @Summary Upload document
// @Tags Uploads
// @Security BearerAuth
// @Accept multipart/form-data
// @Produce json
// @Param file formance file true "Document file"
// @Param folder formance string false "Folder name"
// @Success 200 {object} response.Response{data=map[string]string}
// @Router /uploads/document [post]
func (h *UploadHandler) UploadDocument(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "No file provided")
	}

	folder := c.FormValue("folder")
	if folder == "" {
		folder = "documents"
	}

	url, err := h.storageSvc.UploadDocument(c.Request().Context(), file, folder)
	if err != nil {
		return response.BadRequest(c, err.Error())
	}

	return response.Success(c, map[string]string{
		"url":      url,
		"filename": file.Filename,
	})
}
