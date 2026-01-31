package storage

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestStorage(t *testing.T) (*LocalStorage, func()) {
	// Create temp directory for tests
	tempDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	config := Config{
		Type:        "local",
		UploadsDir:  tempDir,
		MaxFileSize: 1024 * 1024, // 1MB for tests
		AllowedTypes: []string{
			"image/jpeg",
			"image/png",
			"image/gif",
			"image/webp",
		},
		BaseURL: "/uploads",
	}

	storage, err := NewLocalStorage(config)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to create storage: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return storage, cleanup
}

func TestNewLocalStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	config := DefaultConfig()
	config.UploadsDir = tempDir

	storage, err := NewLocalStorage(config)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	if storage == nil {
		t.Fatal("storage should not be nil")
	}
}

func TestNewLocalStorage_CreatesDirectory(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "storage-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Use a subdirectory that doesn't exist
	uploadsDir := filepath.Join(tempDir, "uploads", "nested")

	config := DefaultConfig()
	config.UploadsDir = uploadsDir

	_, err = NewLocalStorage(config)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
		t.Error("uploads directory should have been created")
	}
}

func TestLocalStorage_Upload(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	content := []byte("test file content")
	reader := bytes.NewReader(content)

	url, err := storage.Upload("test.jpg", reader, int64(len(content)))
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	if url == "" {
		t.Error("URL should not be empty")
	}

	if !strings.HasPrefix(url, "/uploads/") {
		t.Errorf("URL should start with /uploads/, got %s", url)
	}

	if !strings.HasSuffix(url, ".jpg") {
		t.Errorf("URL should end with .jpg, got %s", url)
	}
}

func TestLocalStorage_Upload_UniqueFilenames(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	content := []byte("test content")

	url1, _ := storage.Upload("test.jpg", bytes.NewReader(content), int64(len(content)))
	url2, _ := storage.Upload("test.jpg", bytes.NewReader(content), int64(len(content)))

	if url1 == url2 {
		t.Error("uploads should have unique filenames")
	}
}

func TestLocalStorage_Upload_FileTooLarge(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// Create content larger than max size (1MB in test config)
	content := make([]byte, 2*1024*1024) // 2MB

	_, err := storage.Upload("large.jpg", bytes.NewReader(content), int64(len(content)))
	if err != ErrFileTooLarge {
		t.Errorf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestLocalStorage_UploadWithMimeType(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	content := []byte("test image content")
	reader := bytes.NewReader(content)

	url, err := storage.UploadWithMimeType(reader, int64(len(content)), "image/png")
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	if !strings.HasSuffix(url, ".png") {
		t.Errorf("URL should end with .png, got %s", url)
	}
}

func TestLocalStorage_UploadWithMimeType_InvalidType(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	content := []byte("test content")
	reader := bytes.NewReader(content)

	_, err := storage.UploadWithMimeType(reader, int64(len(content)), "application/pdf")
	if err != ErrInvalidFileType {
		t.Errorf("expected ErrInvalidFileType, got %v", err)
	}
}

func TestLocalStorage_UploadWithMimeType_FileTooLarge(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	content := make([]byte, 2*1024*1024) // 2MB

	_, err := storage.UploadWithMimeType(bytes.NewReader(content), int64(len(content)), "image/jpeg")
	if err != ErrFileTooLarge {
		t.Errorf("expected ErrFileTooLarge, got %v", err)
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	content := []byte("test content")
	url, _ := storage.Upload("test.jpg", bytes.NewReader(content), int64(len(content)))

	// Extract filename from URL
	filename := filepath.Base(url)

	// Verify file exists
	if !storage.Exists(filename) {
		t.Fatal("file should exist before delete")
	}

	// Delete
	err := storage.Delete(filename)
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	// Verify file is gone
	if storage.Exists(filename) {
		t.Error("file should not exist after delete")
	}
}

func TestLocalStorage_Delete_NotFound(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	err := storage.Delete("nonexistent.jpg")
	if err != ErrFileNotFound {
		t.Errorf("expected ErrFileNotFound, got %v", err)
	}
}

func TestLocalStorage_Exists(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	content := []byte("test content")
	url, _ := storage.Upload("test.jpg", bytes.NewReader(content), int64(len(content)))

	filename := filepath.Base(url)

	if !storage.Exists(filename) {
		t.Error("file should exist")
	}

	if storage.Exists("nonexistent.jpg") {
		t.Error("nonexistent file should not exist")
	}
}

func TestLocalStorage_GetURL(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	url := storage.GetURL("test-file.jpg")

	if url != "/uploads/test-file.jpg" {
		t.Errorf("expected /uploads/test-file.jpg, got %s", url)
	}
}

func TestLocalStorage_GetFilePath(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	path := storage.GetFilePath("test-file.jpg")

	if !strings.HasSuffix(path, "test-file.jpg") {
		t.Errorf("path should end with test-file.jpg, got %s", path)
	}
}

func TestConfig_IsAllowedType(t *testing.T) {
	config := DefaultConfig()

	tests := []struct {
		mimeType string
		allowed  bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"image/gif", true},
		{"image/webp", true},
		{"IMAGE/JPEG", true}, // Case insensitive
		{"application/pdf", false},
		{"text/html", false},
		{"", false},
	}

	for _, tt := range tests {
		result := config.IsAllowedType(tt.mimeType)
		if result != tt.allowed {
			t.Errorf("IsAllowedType(%s) = %v, want %v", tt.mimeType, result, tt.allowed)
		}
	}
}

func TestGetExtensionForMimeType(t *testing.T) {
	tests := []struct {
		mimeType string
		ext      string
	}{
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"application/pdf", ""},
		{"", ""},
	}

	for _, tt := range tests {
		result := GetExtensionForMimeType(tt.mimeType)
		if result != tt.ext {
			t.Errorf("GetExtensionForMimeType(%s) = %s, want %s", tt.mimeType, result, tt.ext)
		}
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test.jpg", "test.jpg"},
		{"my file.jpg", "my_file.jpg"},
		{"../../../etc/passwd", "passwd"},
		{"/path/to/file.jpg", "file.jpg"},
	}

	for _, tt := range tests {
		result := SanitizeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("SanitizeFilename(%s) = %s, want %s", tt.input, result, tt.expected)
		}
	}
}
