package domain

import (
	"context"
	"io"
	"mime/multipart"
)

// StorageService interface for file storage operations
type StorageService interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)
	UploadImage(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)
	UploadVideo(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)
	UploadDocument(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)
	DeleteFile(ctx context.Context, url string) error
	GetFilePath(url string) string
	FileExists(url string) bool
	UploadHLSFiles(ctx context.Context, localDir string, s3Prefix string) error
	GetFileStream(ctx context.Context, path string) (io.ReadCloser, string, error)
}
