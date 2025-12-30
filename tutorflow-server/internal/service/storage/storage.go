package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/tutorflow/tutorflow-server/internal/pkg/config"
)

// Service handles file storage operations
type Service struct {
	cfg      config.StorageConfig
	basePath string
	cdnBase  string
}

// NewService creates a new storage service
func NewService(cfg config.StorageConfig) *Service {
	return &Service{
		cfg:      cfg,
		basePath: cfg.LocalPath,
		cdnBase:  cfg.CDNBaseURL,
	}
}

// UploadFile uploads a file and returns the URL
func (s *Service) UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s-%d%s", uuid.New().String(), time.Now().UnixNano(), ext)

	// Create directory if not exists
	dir := filepath.Join(s.basePath, folder)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	// Create destination file
	dstPath := filepath.Join(dir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy file
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	// Return URL
	return fmt.Sprintf("%s/%s/%s", s.cdnBase, folder, filename), nil
}

// UploadImage uploads an image with validation
func (s *Service) UploadImage(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

	valid := false
	for _, allowed := range allowedExts {
		if ext == allowed {
			valid = true
			break
		}
	}
	if !valid {
		return "", fmt.Errorf("invalid image format: %s", ext)
	}

	// Check file size (10MB max)
	if file.Size > 10*1024*1024 {
		return "", fmt.Errorf("file too large: max 10MB")
	}

	return s.UploadFile(ctx, file, folder)
}

// UploadVideo uploads a video file
func (s *Service) UploadVideo(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".mp4", ".mov", ".avi", ".mkv", ".webm"}

	valid := false
	for _, allowed := range allowedExts {
		if ext == allowed {
			valid = true
			break
		}
	}
	if !valid {
		return "", fmt.Errorf("invalid video format: %s", ext)
	}

	// Check file size (500MB max)
	if file.Size > 500*1024*1024 {
		return "", fmt.Errorf("file too large: max 500MB")
	}

	return s.UploadFile(ctx, file, folder)
}

// UploadDocument uploads a document file
func (s *Service) UploadDocument(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := []string{".pdf", ".doc", ".docx", ".ppt", ".pptx", ".xls", ".xlsx", ".txt", ".zip"}

	valid := false
	for _, allowed := range allowedExts {
		if ext == allowed {
			valid = true
			break
		}
	}
	if !valid {
		return "", fmt.Errorf("invalid document format: %s", ext)
	}

	// Check file size (50MB max)
	if file.Size > 50*1024*1024 {
		return "", fmt.Errorf("file too large: max 50MB")
	}

	return s.UploadFile(ctx, file, folder)
}

// DeleteFile deletes a file by URL
func (s *Service) DeleteFile(ctx context.Context, url string) error {
	// Extract path from URL
	if url == "" {
		return nil
	}

	path := strings.TrimPrefix(url, s.cdnBase+"/")
	fullPath := filepath.Join(s.basePath, path)

	return os.Remove(fullPath)
}

// GetFilePath returns the local file path from URL
func (s *Service) GetFilePath(url string) string {
	path := strings.TrimPrefix(url, s.cdnBase+"/")
	return filepath.Join(s.basePath, path)
}

// FileExists checks if a file exists
func (s *Service) FileExists(url string) bool {
	path := s.GetFilePath(url)
	_, err := os.Stat(path)
	return err == nil
}
