package filestore

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type FileStore struct {
	uploadDir string
}

func New(uploadDir string) (*FileStore, error) {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}
	return &FileStore{uploadDir: uploadDir}, nil
}

func (fs *FileStore) SaveFile(file *multipart.FileHeader, subdir string) (string, error) {
	// Validate file type
	if !isAllowedFileType(file.Header.Get("Content-Type")) {
		return "", fmt.Errorf("unsupported file type: %s", file.Header.Get("Content-Type"))
	}

	// Create subdirectory if it doesn't exist
	uploadPath := fs.uploadDir
	if subdir != "" {
		uploadPath = filepath.Join(fs.uploadDir, subdir)
		if err := os.MkdirAll(uploadPath, 0o755); err != nil {
			return "", fmt.Errorf("failed to create upload subdirectory: %w", err)
		}
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s-%s%s",
		uuid.New().String(),
		time.Now().Format("20060102-150405"),
		ext)

	// Create file path
	filePath := filepath.Join(uploadPath, filename)

	// Open source file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file contents
	if _, err = io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	// Return relative path (including subdirectory)
	if subdir != "" {
		return filepath.Join(subdir, filename), nil
	}
	return filename, nil
}

func isAllowedFileType(contentType string) bool {
	allowedTypes := map[string]bool{
		"image/jpeg":       true,
		"image/png":        true,
		"image/gif":        true,
		"video/mp4":        true,
		"video/webm":       true,
		"video/quicktime":  true,
		"video/x-msvideo":  true, // AVI
		"video/x-matroska": true, // MKV
	}
	return allowedTypes[strings.ToLower(contentType)]
}

// DeleteFile deletes a file from the upload directory
func (fs *FileStore) DeleteFile(filename string) error {
	filePath := filepath.Join(fs.uploadDir, filename)
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}
