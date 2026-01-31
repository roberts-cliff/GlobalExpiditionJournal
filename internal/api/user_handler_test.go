package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"globe-expedition-journal/internal/lti"
	"globe-expedition-journal/internal/middleware"
	"globe-expedition-journal/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestDB(t *testing.T) *gorm.DB {
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

func createTestUser(t *testing.T, db *gorm.DB) *models.User {
	user := &models.User{
		CanvasUserID:      "canvas-123",
		CanvasInstanceURL: "https://canvas.example.com",
		DisplayName:       "Test User",
		Email:             "test@example.com",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

func TestUserHandler_GetMe_Authenticated(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-456", "learner")

	handler := NewUserHandler(db)

	router := gin.New()
	router.Use(middleware.AuthMiddleware(sm))
	router.GET("/api/v1/me", handler.GetMe)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response MeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.ID != user.ID {
		t.Errorf("expected ID %d, got %d", user.ID, response.ID)
	}
	if response.CanvasID != "canvas-123" {
		t.Errorf("expected CanvasID 'canvas-123', got '%s'", response.CanvasID)
	}
	if response.CourseID != "course-456" {
		t.Errorf("expected CourseID 'course-456', got '%s'", response.CourseID)
	}
	if response.Role != "learner" {
		t.Errorf("expected Role 'learner', got '%s'", response.Role)
	}
	if response.DisplayName != "Test User" {
		t.Errorf("expected DisplayName 'Test User', got '%s'", response.DisplayName)
	}
}

func TestUserHandler_GetMe_Unauthenticated(t *testing.T) {
	db := setupTestDB(t)
	handler := NewUserHandler(db)

	router := gin.New()
	router.GET("/api/v1/me", handler.GetMe)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestUserHandler_GetMe_UserNotFound(t *testing.T) {
	db := setupTestDB(t)

	sm := lti.NewSessionManager("test-secret", 3600)
	// Token for non-existent user
	token, _ := sm.CreateToken(999, "canvas-999", "course-456", "learner")

	handler := NewUserHandler(db)

	router := gin.New()
	router.Use(middleware.AuthMiddleware(sm))
	router.GET("/api/v1/me", handler.GetMe)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestUserHandler_Logout(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-456", "learner")

	handler := NewUserHandler(db)

	router := gin.New()
	router.Use(middleware.AuthMiddleware(sm))
	router.POST("/api/v1/logout", handler.Logout)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Check that the session cookie was cleared
	cookies := w.Result().Cookies()
	found := false
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			found = true
			if cookie.MaxAge >= 0 {
				t.Error("expected session cookie to have negative MaxAge")
			}
		}
	}
	if !found {
		t.Error("expected session cookie to be set (for clearing)")
	}
}

func TestUserHandler_GetMe_InstructorRole(t *testing.T) {
	db := setupTestDB(t)
	user := createTestUser(t, db)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-789", "instructor")

	handler := NewUserHandler(db)

	router := gin.New()
	router.Use(middleware.AuthMiddleware(sm))
	router.GET("/api/v1/me", handler.GetMe)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response MeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.Role != "instructor" {
		t.Errorf("expected Role 'instructor', got '%s'", response.Role)
	}
}
