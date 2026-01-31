package storage

import (
	"errors"
	"io"
	"path/filepath"
	"strings"
)

// Common errors
var (
	ErrFileTooLarge    = errors.New("file exceeds maximum size limit")
	ErrInvalidFileType = errors.New("file type not allowed")
	ErrFileNotFound    = errors.New("file not found")
)

// Storage defines the interface for file storage operations
type Storage interface {
	// Upload stores a file and returns its URL
	Upload(filename string, content io.Reader, size int64) (string, error)

	// Delete removes a file from storage
	Delete(filename string) error

	// GetURL returns the public URL for a stored file
	GetURL(filename string) string

	// Exists checks if a file exists in storage
	Exists(filename string) bool
}

// Config holds storage configuration
type Config struct {
	Type         string   // "local" or "s3"
	UploadsDir   string   // Local directory for uploads
	MaxFileSize  int64    // Maximum file size in bytes
	AllowedTypes []string // Allowed MIME types
	BaseURL      string   // Base URL for serving files
}

// DefaultConfig returns default storage configuration
func DefaultConfig() Config {
	return Config{
		Type:        "local",
		UploadsDir:  "./uploads",
		MaxFileSize: 10 * 1024 * 1024, // 10MB
		AllowedTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
		},
		BaseURL: "/uploads",
	}
}

// IsAllowedType checks if a MIME type is in the allowed list
func (c Config) IsAllowedType(mimeType string) bool {
	mimeType = strings.ToLower(strings.TrimSpace(mimeType))
	for _, allowed := range c.AllowedTypes {
		if strings.ToLower(allowed) == mimeType {
			return true
		}
	}
	return false
}

// GetExtensionForMimeType returns the file extension for a MIME type
func GetExtensionForMimeType(mimeType string) string {
	switch strings.ToLower(mimeType) {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}

// SanitizeFilename removes potentially dangerous characters from filenames
func SanitizeFilename(filename string) string {
	// Get just the base name without path
	filename = filepath.Base(filename)
	// Replace spaces and special chars
	filename = strings.ReplaceAll(filename, " ", "_")
	return filename
}
