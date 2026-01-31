package api

import (
	"bytes"
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

func setupScrapbookTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.User{}, &models.Country{}, &models.ScrapbookEntry{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func seedScrapbookTestData(t *testing.T, db *gorm.DB) (*models.User, *models.Country) {
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

func createScrapbookTestRouter(db *gorm.DB, sm *lti.SessionManager) *gin.Engine {
	router := gin.New()
	handler := NewScrapbookHandler(db)

	auth := router.Group("/api/v1/scrapbook")
	auth.Use(middleware.AuthMiddleware(sm))
	{
		auth.GET("/entries", handler.ListEntries)
		auth.POST("/entries", handler.CreateEntry)
		auth.GET("/entries/:id", handler.GetEntry)
		auth.PUT("/entries/:id", handler.UpdateEntry)
		auth.DELETE("/entries/:id", handler.DeleteEntry)
		auth.GET("/countries/:countryId/entries", handler.GetEntriesByCountry)
		auth.GET("/stats", handler.GetStats)
	}

	return router
}

func TestScrapbookHandler_ListEntries_Empty(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, _ := seedScrapbookTestData(t, db)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/scrapbook/entries", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response ScrapbookEntryListResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Total != 0 {
		t.Errorf("expected 0 entries, got %d", response.Total)
	}
}

func TestScrapbookHandler_CreateEntry(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, country := seedScrapbookTestData(t, db)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	body := CreateScrapbookEntryRequest{
		CountryID: country.ID,
		Title:     "My Paris Trip",
		Notes:     "Amazing view from Eiffel Tower",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scrapbook/entries", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var response ScrapbookEntryResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Title != "My Paris Trip" {
		t.Errorf("expected title 'My Paris Trip', got '%s'", response.Title)
	}
	if response.Notes != "Amazing view from Eiffel Tower" {
		t.Errorf("expected notes 'Amazing view from Eiffel Tower', got '%s'", response.Notes)
	}
	if response.Country == nil {
		t.Error("expected country to be included")
	}
}

func TestScrapbookHandler_CreateEntry_WithMedia(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, country := seedScrapbookTestData(t, db)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	body := CreateScrapbookEntryRequest{
		CountryID: country.ID,
		Title:     "Cherry Blossoms",
		MediaURL:  "https://storage.example.com/photos/123.jpg",
		MediaType: "image/jpeg",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scrapbook/entries", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var response ScrapbookEntryResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.MediaURL != "https://storage.example.com/photos/123.jpg" {
		t.Errorf("expected mediaUrl to be set, got '%s'", response.MediaURL)
	}
	if response.MediaType != "image/jpeg" {
		t.Errorf("expected mediaType 'image/jpeg', got '%s'", response.MediaType)
	}
}

func TestScrapbookHandler_CreateEntry_WithDate(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, country := seedScrapbookTestData(t, db)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	visitDate := "2024-06-15T10:00:00Z"
	body := CreateScrapbookEntryRequest{
		CountryID: country.ID,
		Title:     "Summer Trip",
		VisitedAt: visitDate,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scrapbook/entries", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var response ScrapbookEntryResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.VisitedAt != visitDate {
		t.Errorf("expected visitedAt '%s', got '%s'", visitDate, response.VisitedAt)
	}
}

func TestScrapbookHandler_CreateEntry_InvalidCountry(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, _ := seedScrapbookTestData(t, db)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	body := CreateScrapbookEntryRequest{
		CountryID: 999, // Non-existent
		Title:     "Invalid",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scrapbook/entries", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestScrapbookHandler_CreateEntry_MissingTitle(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, country := seedScrapbookTestData(t, db)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	body := map[string]interface{}{
		"countryId": country.ID,
		// Missing title
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scrapbook/entries", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestScrapbookHandler_GetEntry(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, country := seedScrapbookTestData(t, db)

	entry := &models.ScrapbookEntry{
		UserID:    user.ID,
		CountryID: country.ID,
		Title:     "Test Entry",
		Notes:     "Test notes",
	}
	db.Create(entry)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/scrapbook/entries/1", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response ScrapbookEntryResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Title != "Test Entry" {
		t.Errorf("expected title 'Test Entry', got '%s'", response.Title)
	}
}

func TestScrapbookHandler_GetEntry_NotOwned(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, country := seedScrapbookTestData(t, db)

	// Create another user
	otherUser := &models.User{CanvasUserID: "other", CanvasInstanceURL: "https://canvas.example.com"}
	db.Create(otherUser)

	// Create entry for other user
	entry := &models.ScrapbookEntry{
		UserID:    otherUser.ID,
		CountryID: country.ID,
		Title:     "Other's Entry",
	}
	db.Create(entry)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/scrapbook/entries/1", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestScrapbookHandler_UpdateEntry(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, country := seedScrapbookTestData(t, db)

	entry := &models.ScrapbookEntry{
		UserID:    user.ID,
		CountryID: country.ID,
		Title:     "Old Title",
		Notes:     "Original notes",
	}
	db.Create(entry)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	body := UpdateScrapbookEntryRequest{
		Title: "New Title",
		Notes: "Updated notes",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/scrapbook/entries/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response ScrapbookEntryResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", response.Title)
	}
	if response.Notes != "Updated notes" {
		t.Errorf("expected notes 'Updated notes', got '%s'", response.Notes)
	}
}

func TestScrapbookHandler_DeleteEntry(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, country := seedScrapbookTestData(t, db)

	entry := &models.ScrapbookEntry{
		UserID:    user.ID,
		CountryID: country.ID,
		Title:     "To Delete",
	}
	db.Create(entry)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/scrapbook/entries/1", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify soft deleted
	var count int64
	db.Model(&models.ScrapbookEntry{}).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 entries after delete, got %d", count)
	}
}

func TestScrapbookHandler_GetEntriesByCountry(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, country := seedScrapbookTestData(t, db)

	// Create another country
	country2 := &models.Country{Name: "Germany", ISOCode: "DE", Region: "Europe"}
	db.Create(country2)

	// Create entries
	db.Create(&models.ScrapbookEntry{UserID: user.ID, CountryID: country.ID, Title: "Paris 1"})
	db.Create(&models.ScrapbookEntry{UserID: user.ID, CountryID: country.ID, Title: "Paris 2"})
	db.Create(&models.ScrapbookEntry{UserID: user.ID, CountryID: country2.ID, Title: "Berlin"})

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/scrapbook/countries/1/entries", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response struct {
		Entries []ScrapbookEntryResponse `json:"entries"`
	}
	json.Unmarshal(w.Body.Bytes(), &response)

	if len(response.Entries) != 2 {
		t.Errorf("expected 2 entries for country 1, got %d", len(response.Entries))
	}
}

func TestScrapbookHandler_GetStats(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, country := seedScrapbookTestData(t, db)

	// Create another country
	country2 := &models.Country{Name: "Germany", ISOCode: "DE", Region: "Europe"}
	db.Create(country2)

	// Create entries
	db.Create(&models.ScrapbookEntry{UserID: user.ID, CountryID: country.ID, Title: "Entry 1"})
	db.Create(&models.ScrapbookEntry{UserID: user.ID, CountryID: country.ID, Title: "Entry 2", MediaURL: "http://photo.jpg"})
	db.Create(&models.ScrapbookEntry{UserID: user.ID, CountryID: country2.ID, Title: "Entry 3", MediaURL: "http://photo2.jpg"})

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/scrapbook/stats", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response ScrapbookStatsResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.TotalEntries != 3 {
		t.Errorf("expected 3 total entries, got %d", response.TotalEntries)
	}
	if response.CountriesDocumented != 2 {
		t.Errorf("expected 2 countries documented, got %d", response.CountriesDocumented)
	}
	if response.PhotosUploaded != 2 {
		t.Errorf("expected 2 photos uploaded, got %d", response.PhotosUploaded)
	}
}

func TestScrapbookHandler_ListEntries_WithData(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, country := seedScrapbookTestData(t, db)

	// Create entries
	db.Create(&models.ScrapbookEntry{UserID: user.ID, CountryID: country.ID, Title: "Entry 1"})
	db.Create(&models.ScrapbookEntry{UserID: user.ID, CountryID: country.ID, Title: "Entry 2"})

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/scrapbook/entries", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response ScrapbookEntryListResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Total != 2 {
		t.Errorf("expected 2 entries, got %d", response.Total)
	}
}

func TestScrapbookHandler_Unauthenticated(t *testing.T) {
	db := setupScrapbookTestDB(t)
	sm := lti.NewSessionManager("test-secret", 3600)
	router := createScrapbookTestRouter(db, sm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/scrapbook/entries", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestScrapbookHandler_GetEntry_InvalidID(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, _ := seedScrapbookTestData(t, db)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/scrapbook/entries/invalid", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestScrapbookHandler_CreateEntry_WithTags(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, country := seedScrapbookTestData(t, db)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	body := CreateScrapbookEntryRequest{
		CountryID: country.ID,
		Title:     "Museum Visit",
		Notes:     "Great art collection",
		Tags:      "museum,art,culture",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scrapbook/entries", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d: %s", w.Code, w.Body.String())
	}

	var response ScrapbookEntryResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Tags != "museum,art,culture" {
		t.Errorf("expected tags 'museum,art,culture', got '%s'", response.Tags)
	}
}

func TestScrapbookHandler_UpdateEntry_WithTags(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, country := seedScrapbookTestData(t, db)

	entry := &models.ScrapbookEntry{
		UserID:    user.ID,
		CountryID: country.ID,
		Title:     "Original Entry",
		Tags:      "original",
	}
	db.Create(entry)

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	body := UpdateScrapbookEntryRequest{
		Title: "Updated Entry",
		Tags:  "updated,new-tags",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/scrapbook/entries/1", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response ScrapbookEntryResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Tags != "updated,new-tags" {
		t.Errorf("expected tags 'updated,new-tags', got '%s'", response.Tags)
	}
}

func TestScrapbookHandler_ListEntries_FilterByTag(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, country := seedScrapbookTestData(t, db)

	// Create entries with different tags
	db.Create(&models.ScrapbookEntry{UserID: user.ID, CountryID: country.ID, Title: "Museum Visit", Tags: "museum,art"})
	db.Create(&models.ScrapbookEntry{UserID: user.ID, CountryID: country.ID, Title: "Food Tour", Tags: "food,culture"})
	db.Create(&models.ScrapbookEntry{UserID: user.ID, CountryID: country.ID, Title: "Art Gallery", Tags: "museum,art,gallery"})

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	// Filter by 'museum' tag
	req := httptest.NewRequest(http.MethodGet, "/api/v1/scrapbook/entries?tag=museum", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response ScrapbookEntryListResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Total != 2 {
		t.Errorf("expected 2 entries with 'museum' tag, got %d", response.Total)
	}
}

func TestScrapbookHandler_ListEntries_FilterByTag_NoMatch(t *testing.T) {
	db := setupScrapbookTestDB(t)
	user, country := seedScrapbookTestData(t, db)

	db.Create(&models.ScrapbookEntry{UserID: user.ID, CountryID: country.ID, Title: "Entry 1", Tags: "food,travel"})

	sm := lti.NewSessionManager("test-secret", 3600)
	token, _ := sm.CreateToken(user.ID, "canvas-123", "course-1", "learner")

	router := createScrapbookTestRouter(db, sm)

	// Filter by non-existent tag
	req := httptest.NewRequest(http.MethodGet, "/api/v1/scrapbook/entries?tag=nonexistent", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response ScrapbookEntryListResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.Total != 0 {
		t.Errorf("expected 0 entries, got %d", response.Total)
	}
}
