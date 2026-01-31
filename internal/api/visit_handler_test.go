package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"globe-expedition-journal/internal/lti"
	"globe-expedition-journal/internal/middleware"
	"globe-expedition-journal/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupVisitTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Country{}, &models.Visit{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func seedVisitTestData(t *testing.T, db *gorm.DB) (*models.User, *models.Country) {
	user := &models.User{
		CanvasUserID:      "canvas-123",
		CanvasInstanceURL: "https://canvas.example.com",
		DisplayName:       "Test User",
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	country := &models.Country{
		Name:    "France",
		ISOCode: "FR",
		Region:  "Europe",
	}
	if err := db.Create(country).Error; err != nil {
		t.Fatalf("failed to create country: %v", err)
	}

	return user, country
}

func createVisitTestRouter(db *gorm.DB, sm *lti.SessionManager) *gin.Engine {
	router := gin.New()
	handler := NewVisitHandler(db)

	auth := router.Group("/api/v1")
	auth.Use(middleware.AuthMiddleware(sm))
	{
		auth.GET("/visits", handler.ListVisits)
		auth.POST("/visits", handler.CreateVisit)
		auth.GET("/visits/:id", handler.GetVisit)
		auth.PUT("/visits/:id", handler.UpdateVisit)
		auth.DELETE("/visits/:id", handler.DeleteVisit)
		auth.GET("/visits/country/:countryId", handler.GetVisitsByCountry)
	}

	return router
}

func TestVisitHandler_ListVisits_Empty(t *testing.T) {
	db := setupVisitTestDB(t)
	user, _ := seedVisitTestData(t, db)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createVisitTestRouter(db, sm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/visits", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response VisitListResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Total != 0 {
		t.Errorf("expected 0 visits, got %d", response.Total)
	}
}

func TestVisitHandler_CreateVisit(t *testing.T) {
	db := setupVisitTestDB(t)
	user, country := seedVisitTestData(t, db)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createVisitTestRouter(db, sm)

	body := CreateVisitRequest{
		CountryID: country.ID,
		Notes:     "Great trip!",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/visits", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var response VisitResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.CountryID != country.ID {
		t.Errorf("expected country ID %d, got %d", country.ID, response.CountryID)
	}
	if response.Notes != "Great trip!" {
		t.Errorf("expected notes 'Great trip!', got '%s'", response.Notes)
	}
	if response.Country == nil {
		t.Error("expected country to be included")
	}
}

func TestVisitHandler_CreateVisit_WithDate(t *testing.T) {
	db := setupVisitTestDB(t)
	user, country := seedVisitTestData(t, db)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createVisitTestRouter(db, sm)

	visitDate := "2024-06-15T10:00:00Z"
	body := CreateVisitRequest{
		CountryID: country.ID,
		VisitedAt: visitDate,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/visits", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var response VisitResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.VisitedAt != visitDate {
		t.Errorf("expected visitedAt '%s', got '%s'", visitDate, response.VisitedAt)
	}
}

func TestVisitHandler_CreateVisit_InvalidCountry(t *testing.T) {
	db := setupVisitTestDB(t)
	user, _ := seedVisitTestData(t, db)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createVisitTestRouter(db, sm)

	body := CreateVisitRequest{
		CountryID: 999, // Non-existent
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/visits", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestVisitHandler_GetVisit(t *testing.T) {
	db := setupVisitTestDB(t)
	user, country := seedVisitTestData(t, db)

	visit := &models.Visit{
		UserID:    user.ID,
		CountryID: country.ID,
		VisitedAt: time.Now(),
		Notes:     "Test visit",
	}
	db.Create(visit)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createVisitTestRouter(db, sm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/visits/1", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response VisitResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Notes != "Test visit" {
		t.Errorf("expected notes 'Test visit', got '%s'", response.Notes)
	}
}

func TestVisitHandler_GetVisit_NotOwned(t *testing.T) {
	db := setupVisitTestDB(t)
	user, country := seedVisitTestData(t, db)

	// Create another user
	otherUser := &models.User{CanvasUserID: "other", CanvasInstanceURL: "https://canvas.example.com"}
	db.Create(otherUser)

	// Create visit for other user
	visit := &models.Visit{
		UserID:    otherUser.ID,
		CountryID: country.ID,
		VisitedAt: time.Now(),
	}
	db.Create(visit)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createVisitTestRouter(db, sm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/visits/1", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestVisitHandler_UpdateVisit(t *testing.T) {
	db := setupVisitTestDB(t)
	user, country := seedVisitTestData(t, db)

	visit := &models.Visit{
		UserID:    user.ID,
		CountryID: country.ID,
		VisitedAt: time.Now(),
		Notes:     "Original notes",
	}
	db.Create(visit)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createVisitTestRouter(db, sm)

	body := UpdateVisitRequest{
		Notes: "Updated notes",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/visits/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response VisitResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Notes != "Updated notes" {
		t.Errorf("expected notes 'Updated notes', got '%s'", response.Notes)
	}
}

func TestVisitHandler_DeleteVisit(t *testing.T) {
	db := setupVisitTestDB(t)
	user, country := seedVisitTestData(t, db)

	visit := &models.Visit{
		UserID:    user.ID,
		CountryID: country.ID,
		VisitedAt: time.Now(),
	}
	db.Create(visit)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createVisitTestRouter(db, sm)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/visits/1", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify deleted
	var count int64
	db.Model(&models.Visit{}).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 visits after delete, got %d", count)
	}
}

func TestVisitHandler_GetVisitsByCountry(t *testing.T) {
	db := setupVisitTestDB(t)
	user, country := seedVisitTestData(t, db)

	// Create another country
	country2 := &models.Country{Name: "Germany", ISOCode: "DE", Region: "Europe"}
	db.Create(country2)

	// Create visits
	db.Create(&models.Visit{UserID: user.ID, CountryID: country.ID, VisitedAt: time.Now()})
	db.Create(&models.Visit{UserID: user.ID, CountryID: country.ID, VisitedAt: time.Now()})
	db.Create(&models.Visit{UserID: user.ID, CountryID: country2.ID, VisitedAt: time.Now()})

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createVisitTestRouter(db, sm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/visits/country/1", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response struct {
		Visits []VisitResponse `json:"visits"`
	}
	json.Unmarshal(w.Body.Bytes(), &response)

	if len(response.Visits) != 2 {
		t.Errorf("expected 2 visits for country 1, got %d", len(response.Visits))
	}
}

func TestVisitHandler_ListVisits_WithData(t *testing.T) {
	db := setupVisitTestDB(t)
	user, country := seedVisitTestData(t, db)

	// Create visits
	db.Create(&models.Visit{UserID: user.ID, CountryID: country.ID, VisitedAt: time.Now()})
	db.Create(&models.Visit{UserID: user.ID, CountryID: country.ID, VisitedAt: time.Now().Add(-24 * time.Hour)})

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createVisitTestRouter(db, sm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/visits", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response VisitListResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Total != 2 {
		t.Errorf("expected 2 visits, got %d", response.Total)
	}
}

func TestVisitHandler_Unauthenticated(t *testing.T) {
	db := setupVisitTestDB(t)
	sm := lti.NewSessionManager("test-secret", 3600)
	router := createVisitTestRouter(db, sm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/visits", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}
