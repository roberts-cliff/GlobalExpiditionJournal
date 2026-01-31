package api

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"globe-expedition-journal/internal/lti"
	"globe-expedition-journal/internal/middleware"
	"globe-expedition-journal/internal/models"
	"globe-expedition-journal/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupUploadTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func setupUploadTestStorage(t *testing.T) (*storage.LocalStorage, func()) {
	tempDir, err := os.MkdirTemp("", "upload-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	config := storage.Config{
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

	s, err := storage.NewLocalStorage(config)
	if err != nil {
		os.RemoveAll(tempDir)
		t.Fatalf("failed to create storage: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(tempDir)
	}

	return s, cleanup
}

func seedUploadTestUser(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{
		CanvasUserID:      "canvas-123",
		CanvasInstanceURL: "https://canvas.example.com",
		DisplayName:       "Test User",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	return user
}

func createUploadTestRouter(s *storage.LocalStorage, sm *lti.SessionManager) *gin.Engine {
	router := gin.New()
	handler := NewUploadHandler(s)

	auth := router.Group("/api/v1")
	auth.Use(middleware.AuthMiddleware(sm))
	{
		auth.POST("/upload", handler.Upload)
		auth.DELETE("/upload/:filename", handler.Delete)
	}

	return router
}

func createMultipartRequest(t *testing.T, fieldName, filename, contentType string, content []byte) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fieldName, filename)
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(part, bytes.NewReader(content))
	if err != nil {
		return nil, err
	}
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Set the Content-Type for the file part
	// This is a bit hacky but works for testing
	if contentType != "" {
		// The actual content type is set in the multipart header
	}

	return req, nil
}

func TestUploadHandler_Upload_Success(t *testing.T) {
	db := setupUploadTestDB(t)
	user := seedUploadTestUser(t, db)
	s, cleanup := setupUploadTestStorage(t)
	defer cleanup()

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createUploadTestRouter(s, sm)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Create a form file with proper content type
	h := make(map[string][]string)
	h["Content-Disposition"] = []string{`form-data; name="file"; filename="test.jpg"`}
	h["Content-Type"] = []string{"image/jpeg"}
	part, _ := writer.CreatePart(h)
	part.Write([]byte("fake jpeg content"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var response UploadResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.URL == "" {
		t.Error("expected URL in response")
	}
}

func TestUploadHandler_Upload_NoFile(t *testing.T) {
	db := setupUploadTestDB(t)
	user := seedUploadTestUser(t, db)
	s, cleanup := setupUploadTestStorage(t)
	defer cleanup()

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createUploadTestRouter(s, sm)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestUploadHandler_Upload_InvalidFileType(t *testing.T) {
	db := setupUploadTestDB(t)
	user := seedUploadTestUser(t, db)
	s, cleanup := setupUploadTestStorage(t)
	defer cleanup()

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createUploadTestRouter(s, sm)

	// Create multipart form with PDF
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(map[string][]string)
	h["Content-Disposition"] = []string{`form-data; name="file"; filename="test.pdf"`}
	h["Content-Type"] = []string{"application/pdf"}
	part, _ := writer.CreatePart(h)
	part.Write([]byte("fake pdf content"))
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestUploadHandler_Upload_Unauthenticated(t *testing.T) {
	s, cleanup := setupUploadTestStorage(t)
	defer cleanup()

	sm := lti.NewSessionManager("test-secret", 3600)
	router := createUploadTestRouter(s, sm)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/upload", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestUploadHandler_Delete_Success(t *testing.T) {
	db := setupUploadTestDB(t)
	user := seedUploadTestUser(t, db)
	s, cleanup := setupUploadTestStorage(t)
	defer cleanup()

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	// Upload a file first
	url, _ := s.UploadWithMimeType(bytes.NewReader([]byte("test")), 4, "image/jpeg")
	filename := filepath.Base(url)

	router := createUploadTestRouter(s, sm)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/upload/"+filename, nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify file is deleted
	if s.Exists(filename) {
		t.Error("file should have been deleted")
	}
}

func TestUploadHandler_Delete_NotFound(t *testing.T) {
	db := setupUploadTestDB(t)
	user := seedUploadTestUser(t, db)
	s, cleanup := setupUploadTestStorage(t)
	defer cleanup()

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createUploadTestRouter(s, sm)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/upload/nonexistent.jpg", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestUploadHandler_Delete_Unauthenticated(t *testing.T) {
	s, cleanup := setupUploadTestStorage(t)
	defer cleanup()

	sm := lti.NewSessionManager("test-secret", 3600)
	router := createUploadTestRouter(s, sm)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/upload/test.jpg", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}
