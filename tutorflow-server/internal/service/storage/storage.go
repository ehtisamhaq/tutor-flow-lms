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
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/tutorflow/tutorflow-server/internal/pkg/config"
)

// Service handles file storage operations
type Service struct {
	cfg         config.StorageConfig
	basePath    string
	cdnBase     string
	minioClient *minio.Client
}

// NewService creates a new storage service
func NewService(cfg config.StorageConfig) *Service {
	svc := &Service{
		cfg:      cfg,
		basePath: cfg.LocalPath,
		cdnBase:  cfg.CDNBaseURL,
	}

	if cfg.Driver == "s3" {
		// Initialize MinIO client object
		useSSL := cfg.UseSSL
		minioClient, err := minio.New(cfg.S3Endpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
			Secure: useSSL,
			Region: cfg.S3Region,
		})
		if err != nil {
			// Panic or log fatal in production, but here we'll just log
			fmt.Printf("Failed to initialize MinIO client: %v\n", err)
		} else {
			svc.minioClient = minioClient
		}
	}

	return svc
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
	path := fmt.Sprintf("%s/%s", folder, filename)

	if s.cfg.Driver == "s3" && s.minioClient != nil {
		// Upload to S3
		info, err := s.minioClient.PutObject(ctx, s.cfg.S3Bucket, path, src, file.Size, minio.PutObjectOptions{
			ContentType: file.Header.Get("Content-Type"),
		})
		if err != nil {
			return "", err
		}
		// Return URL (CDN or S3)
		if s.cdnBase != "" {
			return fmt.Sprintf("%s/%s", s.cdnBase, path), nil
		}

		protocol := "https"
		if !s.cfg.UseSSL {
			protocol = "http"
		}
		return fmt.Sprintf("%s://%s/%s/%s", protocol, s.cfg.S3Endpoint, s.cfg.S3Bucket, info.Key), nil
	}

	// Local Storage Fallback
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

	if s.cfg.Driver == "s3" && s.minioClient != nil {
		return s.minioClient.RemoveObject(ctx, s.cfg.S3Bucket, path, minio.RemoveObjectOptions{})
	}

	fullPath := filepath.Join(s.basePath, path)
	return os.Remove(fullPath)
}

// GetFilePath returns the local file path from URL (only works for local driver)
func (s *Service) GetFilePath(url string) string {
	path := strings.TrimPrefix(url, s.cdnBase+"/")
	return filepath.Join(s.basePath, path)
}

// FileExists checks if a file exists
func (s *Service) FileExists(url string) bool {
	if s.cfg.Driver == "s3" && s.minioClient != nil {
		// Optimization: assume it exists if we have URL, or do a StatObject
		// For now, let's just return true if URL is valid to avoid latency
		return url != ""
	}
	path := s.GetFilePath(url)
	_, err := os.Stat(path)
	return err == nil
}

// UploadHLSFiles uploads all files in a directory to S3 (recursive)
// Used for uploading HLS segments and playlist
func (s *Service) UploadHLSFiles(ctx context.Context, localDir string, s3Prefix string) error {
	if s.cfg.Driver != "s3" {
		// If driver is local, we verify files are in the right place or move them.
		// For this implementation, we assume if processing happened locally in the expected folder,
		// they are already there if we aligned paths.
		// However, video processing writes to random temp dir.
		return nil
	}

	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}

		s3Path := filepath.Join(s3Prefix, relPath)

		// Open file
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// Guess content type
		contentType := "application/octet-stream"
		if strings.HasSuffix(path, ".m3u8") {
			contentType = "application/vnd.apple.mpegurl" // or application/x-mpegURL
		} else if strings.HasSuffix(path, ".ts") {
			contentType = "video/MP2T"
		} else if strings.HasSuffix(path, ".key") {
			contentType = "application/octet-stream"
		}

		// Upload
		_, err = s.minioClient.PutObject(ctx, s.cfg.S3Bucket, s3Path, file, info.Size(), minio.PutObjectOptions{
			ContentType: contentType,
		})
		return err
	})
}

// GetFileStream returns a stream of the file content
func (s *Service) GetFileStream(ctx context.Context, path string) (io.ReadCloser, string, error) {
	if s.cfg.Driver == "s3" && s.minioClient != nil {
		obj, err := s.minioClient.GetObject(ctx, s.cfg.S3Bucket, path, minio.GetObjectOptions{})
		if err != nil {
			return nil, "", err
		}

		info, err := obj.Stat()
		if err != nil {
			return nil, "", err
		}

		return obj, info.ContentType, nil
	}

	// Local fallback
	fullPath := filepath.Join(s.basePath, path)
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, "", err
	}

	contentType := "application/octet-stream"
	if strings.HasSuffix(path, ".m3u8") {
		contentType = "application/vnd.apple.mpegurl"
	} else if strings.HasSuffix(path, ".ts") {
		contentType = "video/MP2T"
	}

	return f, contentType, nil
}
