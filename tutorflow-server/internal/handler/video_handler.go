package handler

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tutorflow/tutorflow-server/internal/domain"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// VideoHandler handles video and DRM-related HTTP requests
type VideoHandler struct {
	videoUC domain.VideoUseCase
}

// NewVideoHandler creates a new video handler
func NewVideoHandler(uc domain.VideoUseCase) *VideoHandler {
	return &VideoHandler{videoUC: uc}
}

// RegisterRoutes registers video/DRM routes
func (h *VideoHandler) RegisterRoutes(e *echo.Group, authMiddleware echo.MiddlewareFunc) {
	// Authenticated routes
	videos := e.Group("/videos", authMiddleware)
	// Upload & Processing
	videos.POST("/lessons/:lessonId/upload", h.UploadVideo)
	videos.GET("/lessons/:lessonId/status", h.GetProcessingStatus)
	videos.DELETE("/lessons/:lessonId", h.DeleteVideo)

	// Playback
	videos.GET("/lessons/:lessonId/playback", h.GetPlaybackURL)

	// DRM routes
	drm := e.Group("/drm", authMiddleware)
	drm.POST("/authorize", h.AuthorizePlayback)
	drm.POST("/heartbeat", h.Heartbeat)
	drm.GET("/devices", h.GetDevices)
	drm.DELETE("/devices/:deviceId", h.RemoveDevice)

	// Key delivery (no auth - uses token)
	e.GET("/drm/key/:token", h.GetEncryptionKey)

	// Admin routes
	admin := e.Group("/admin/videos", authMiddleware)
	admin.POST("/:id/encrypt", h.EnableEncryption)
	admin.POST("/:id/rotate-key", h.RotateKey)

	// Stream Proxy
	e.GET("/videos/stream/:videoId/*", h.ServeHLS)
}

// UploadVideoRequest represents a video upload request
type UploadVideoRequest struct {
	FileURL string `json:"file_url"`
}

// UploadVideo uploads a video for a lesson
func (h *VideoHandler) UploadVideo(c echo.Context) error {
	// For now, accept url in the request body for testing
	// We use an inline struct or the UpgradeVideoRequest type, but let's stick to what was working or simple.
	// The problem in previous edit was removing the struct closing brace.
	// Let's just fix the function body to use new req structure if needed, or stick to old one but pass context.

	// Revert to using UploadVideoRequest but with context
	// Actually, I changed the input to take LessonID in body in previous attempt?
	// Original code took LessonID from Param.
	// Let's stick to Param for LessonID as that's RESTful.

	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid lesson ID"},
		})
	}

	// Check for multipart file first
	file, err := c.FormFile("video")
	if err == nil {
		fmt.Printf("UploadVideo: Found multipart file 'video', size: %d\n", file.Size)
		// Multipart upload
		asset, err := h.videoUC.UploadVideoFile(c.Request().Context(), lessonID, file)
		if err != nil {
			fmt.Printf("UploadVideo: UploadVideoFile failed: %v\n", err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"success": false,
				"error":   map[string]string{"message": err.Error()},
			})
		}
		return c.JSON(http.StatusCreated, map[string]interface{}{
			"success": true,
			"data":    asset,
		})
	}

	fmt.Printf("UploadVideo: FormFile('video') failed: %v, falling back to JSON\n", err)

	// Fallback to JSON body (URL)
	var req UploadVideoRequest
	if err := c.Bind(&req); err != nil {
		fmt.Printf("UploadVideo: JSON bind failed: %v\n", err)
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid request body or missing file"},
		})
	}

	fmt.Printf("UploadVideo: Falling back to URL: '%s'\n", req.FileURL)
	asset, err := h.videoUC.UploadVideo(c.Request().Context(), lessonID, req.FileURL)
	if err != nil {
		fmt.Printf("UploadVideo: UploadVideo failed: %v\n", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    asset,
	})
}

// GetProcessingStatus returns video processing status
func (h *VideoHandler) GetProcessingStatus(c echo.Context) error {
	id, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid lesson ID"},
		})
	}

	// Get video asset by lesson ID
	status, err := h.videoUC.GetProcessingStatus(c.Request().Context(), id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Video not found"},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    status,
	})
}

// DeleteVideo removes a video associated with a lesson
func (h *VideoHandler) DeleteVideo(c echo.Context) error {
	id, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid lesson ID"},
		})
	}

	if err := h.videoUC.DeleteVideo(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Failed to delete video"},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Video deleted successfully",
	})
}

// GetPlaybackURL returns a signed playback URL
func (h *VideoHandler) GetPlaybackURL(c echo.Context) error {
	lessonID, err := uuid.Parse(c.Param("lessonId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid lesson ID"},
		})
	}

	userID := getUserIDFromContext(c)
	deviceID := c.QueryParam("device_id")

	url, err := h.videoUC.GetPlaybackURL(c.Request().Context(), lessonID, userID, deviceID)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]string{
			"url": url,
		},
	})
}

// AuthorizeRequest represents a DRM authorization request
type AuthorizeRequest struct {
	LessonID uuid.UUID `json:"lesson_id"`
	CourseID uuid.UUID `json:"course_id"`
	DeviceID string    `json:"device_id"`
}

// AuthorizePlayback authorizes video playback
func (h *VideoHandler) AuthorizePlayback(c echo.Context) error {
	userID := getUserIDFromContext(c)

	var req AuthorizeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid request body"},
		})
	}

	// Validate device limit
	if err := h.videoUC.ValidateDeviceLimit(c.Request().Context(), userID); err != nil {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	// Generate signed URL
	url, err := h.videoUC.GetPlaybackURL(c.Request().Context(), req.LessonID, userID, req.DeviceID)
	if err != nil {
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"authorized": true,
			"signed_url": url,
		},
	})
}

// HeartbeatRequest represents a session heartbeat
type HeartbeatRequest struct {
	LessonID uuid.UUID `json:"lesson_id"`
	DeviceID string    `json:"device_id"`
}

// Heartbeat validates an active session
func (h *VideoHandler) Heartbeat(c echo.Context) error {
	var req HeartbeatRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid request body"},
		})
	}

	// For now, just return success
	// In production, validate the session is still valid
	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
	})
}

// GetDevices returns user's registered devices
func (h *VideoHandler) GetDevices(c echo.Context) error {
	userID := getUserIDFromContext(c)

	devices, err := h.videoUC.GetUserDevices(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Failed to get devices"},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    devices,
	})
}

// RemoveDevice removes a device from user's account
func (h *VideoHandler) RemoveDevice(c echo.Context) error {
	userID := getUserIDFromContext(c)
	deviceID := c.Param("deviceId")

	if err := h.videoUC.RemoveDevice(c.Request().Context(), userID, deviceID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Device removed",
	})
}

// GetEncryptionKey returns the encryption key for a video
func (h *VideoHandler) GetEncryptionKey(c echo.Context) error {
	token := c.Param("token")

	key, err := h.videoUC.GetEncryptionKey(c.Request().Context(), token)
	if err != nil {
		return c.NoContent(http.StatusForbidden)
	}

	// Return raw key bytes
	c.Response().Header().Set("Content-Type", "application/octet-stream")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	return c.Blob(http.StatusOK, "application/octet-stream", key)
}

// EnableEncryptionRequest represents an encryption request
type EnableEncryptionRequest struct {
	EncryptionType domain.HLSEncryptionType `json:"encryption_type"`
}

// EnableEncryption enables encryption for a video (admin)
func (h *VideoHandler) EnableEncryption(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid video ID"},
		})
	}

	var req EnableEncryptionRequest
	if err := c.Bind(&req); err != nil {
		req.EncryptionType = domain.HLSEncryptionAES128
	}

	if err := h.videoUC.EnableEncryption(c.Request().Context(), id, req.EncryptionType); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Encryption enabled",
	})
}

// RotateKey rotates the encryption key for a video (admin)
func (h *VideoHandler) RotateKey(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": "Invalid video ID"},
		})
	}

	if err := h.videoUC.RotateEncryptionKey(c.Request().Context(), id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"success": false,
			"error":   map[string]string{"message": err.Error()},
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Encryption key rotated",
	})
}

// ServeHLS serves HLS content from S3 with token validation and manifest rewriting
func (h *VideoHandler) ServeHLS(c echo.Context) error {
	videoID, err := uuid.Parse(c.Param("videoId"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid video id"})
	}
	file := c.Param("*")
	token := c.QueryParam("token")

	// Validate token
	if token == "" {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "missing token"})
	}
	if err := h.videoUC.ValidatePlayback(c.Request().Context(), token); err != nil {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "invalid or expired token"})
	}

	// Fetch stream
	stream, contentType, err := h.videoUC.GetVideoSegment(c.Request().Context(), videoID, file)
	if err != nil {
		return c.NoContent(http.StatusNotFound)
	}
	defer stream.Close()

	// Rewrite m3u8 playlist to inject token into key URI and segments
	if strings.HasSuffix(file, ".m3u8") {
		content, err := io.ReadAll(stream)
		if err != nil {
			return err
		}

		lines := strings.Split(string(content), "\n")
		var newLines []string
		for _, line := range lines {
			if strings.HasPrefix(line, "#EXT-X-KEY") {
				// Replace key URI with tokenized URI
				// Search for known pattern /key/<videoID> and replace with /key/<token>
				search := fmt.Sprintf("/key/%s", videoID.String())
				replace := fmt.Sprintf("/key/%s", token)
				line = strings.Replace(line, search, replace, 1)
			} else if !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
				// Append token to segment URL
				if strings.Contains(line, "?") {
					line = fmt.Sprintf("%s&token=%s", line, token)
				} else {
					line = fmt.Sprintf("%s?token=%s", line, token)
				}
			}
			newLines = append(newLines, line)
		}

		output := strings.Join(newLines, "\n")
		return c.Blob(http.StatusOK, contentType, []byte(output))
	}

	return c.Stream(http.StatusOK, contentType, stream)
}
