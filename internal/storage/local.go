package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// LocalStorage implements Storage interface for local filesystem
type LocalStorage struct {
	config Config
}

// NewLocalStorage creates a new LocalStorage instance
func NewLocalStorage(config Config) (*LocalStorage, error) {
	// Ensure uploads directory exists
	if err := os.MkdirAll(config.UploadsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create uploads directory: %w", err)
	}

	return &LocalStorage{config: config}, nil
}

// Upload stores a file locally and returns its URL
func (s *LocalStorage) Upload(filename string, content io.Reader, size int64) (string, error) {
	// Validate file size
	if size > s.config.MaxFileSize {
		return "", ErrFileTooLarge
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".bin"
	}
	uniqueName := uuid.New().String() + ext

	// Create full path
	fullPath := filepath.Join(s.config.UploadsDir, uniqueName)

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content with size limit
	written, err := io.CopyN(file, content, s.config.MaxFileSize+1)
	if err != nil && err != io.EOF {
		os.Remove(fullPath) // Cleanup on error
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Double-check size (in case Content-Length header was wrong)
	if written > s.config.MaxFileSize {
		os.Remove(fullPath)
		return "", ErrFileTooLarge
	}

	return s.GetURL(uniqueName), nil
}

// UploadWithMimeType stores a file with proper extension based on MIME type
func (s *LocalStorage) UploadWithMimeType(content io.Reader, size int64, mimeType string) (string, error) {
	// Validate file type
	if !s.config.IsAllowedType(mimeType) {
		return "", ErrInvalidFileType
	}

	// Validate file size
	if size > s.config.MaxFileSize {
		return "", ErrFileTooLarge
	}

	// Generate unique filename with proper extension
	ext := GetExtensionForMimeType(mimeType)
	if ext == "" {
		return "", ErrInvalidFileType
	}
	uniqueName := uuid.New().String() + ext

	// Create full path
	fullPath := filepath.Join(s.config.UploadsDir, uniqueName)

	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Copy content with size limit
	written, err := io.CopyN(file, content, s.config.MaxFileSize+1)
	if err != nil && err != io.EOF {
		os.Remove(fullPath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	if written > s.config.MaxFileSize {
		os.Remove(fullPath)
		return "", ErrFileTooLarge
	}

	return s.GetURL(uniqueName), nil
}

// Delete removes a file from local storage
func (s *LocalStorage) Delete(filename string) error {
	// Extract just the filename from URL if needed
	filename = filepath.Base(filename)
	fullPath := filepath.Join(s.config.UploadsDir, filename)

	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return ErrFileNotFound
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetURL returns the public URL for a stored file
func (s *LocalStorage) GetURL(filename string) string {
	filename = filepath.Base(filename)
	return s.config.BaseURL + "/" + filename
}

// Exists checks if a file exists in local storage
func (s *LocalStorage) Exists(filename string) bool {
	filename = filepath.Base(filename)
	fullPath := filepath.Join(s.config.UploadsDir, filename)
	_, err := os.Stat(fullPath)
	return err == nil
}

// GetFilePath returns the full filesystem path for a file
func (s *LocalStorage) GetFilePath(filename string) string {
	filename = filepath.Base(filename)
	return filepath.Join(s.config.UploadsDir, filename)
}

// GetConfig returns the storage configuration
func (s *LocalStorage) GetConfig() Config {
	return s.config
}
