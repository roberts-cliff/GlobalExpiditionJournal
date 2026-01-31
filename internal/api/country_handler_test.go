package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"globe-expedition-journal/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupCountryTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.Country{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func seedCountries(t *testing.T, db *gorm.DB) {
	countries := []models.Country{
		{Name: "France", ISOCode: "FR", Region: "Europe"},
		{Name: "Germany", ISOCode: "DE", Region: "Europe"},
		{Name: "Japan", ISOCode: "JP", Region: "Asia"},
		{Name: "Brazil", ISOCode: "BR", Region: "South America"},
		{Name: "Canada", ISOCode: "CA", Region: "North America"},
	}

	for _, c := range countries {
		if err := db.Create(&c).Error; err != nil {
			t.Fatalf("failed to seed country: %v", err)
		}
	}
}

func TestCountryHandler_ListCountries(t *testing.T) {
	db := setupCountryTestDB(t)
	seedCountries(t, db)

	handler := NewCountryHandler(db)

	router := gin.New()
	router.GET("/api/v1/countries", handler.ListCountries)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/countries", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response CountryListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.Total != 5 {
		t.Errorf("expected 5 countries, got %d", response.Total)
	}

	if len(response.Countries) != 5 {
		t.Errorf("expected 5 countries in list, got %d", len(response.Countries))
	}

	// Should be ordered by name
	if response.Countries[0].Name != "Brazil" {
		t.Errorf("expected first country to be Brazil, got %s", response.Countries[0].Name)
	}
}

func TestCountryHandler_ListCountries_FilterByRegion(t *testing.T) {
	db := setupCountryTestDB(t)
	seedCountries(t, db)

	handler := NewCountryHandler(db)

	router := gin.New()
	router.GET("/api/v1/countries", handler.ListCountries)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/countries?region=Europe", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response CountryListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.Total != 2 {
		t.Errorf("expected 2 European countries, got %d", response.Total)
	}
}

func TestCountryHandler_GetCountry(t *testing.T) {
	db := setupCountryTestDB(t)
	seedCountries(t, db)

	handler := NewCountryHandler(db)

	router := gin.New()
	router.GET("/api/v1/countries/:id", handler.GetCountry)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/countries/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response CountryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.Name != "France" {
		t.Errorf("expected France, got %s", response.Name)
	}
}

func TestCountryHandler_GetCountry_NotFound(t *testing.T) {
	db := setupCountryTestDB(t)

	handler := NewCountryHandler(db)

	router := gin.New()
	router.GET("/api/v1/countries/:id", handler.GetCountry)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/countries/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestCountryHandler_GetCountry_InvalidID(t *testing.T) {
	db := setupCountryTestDB(t)

	handler := NewCountryHandler(db)

	router := gin.New()
	router.GET("/api/v1/countries/:id", handler.GetCountry)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/countries/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCountryHandler_GetCountryByCode(t *testing.T) {
	db := setupCountryTestDB(t)
	seedCountries(t, db)

	handler := NewCountryHandler(db)

	router := gin.New()
	router.GET("/api/v1/countries/code/:code", handler.GetCountryByCode)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/countries/code/JP", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response CountryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.Name != "Japan" {
		t.Errorf("expected Japan, got %s", response.Name)
	}
	if response.ISOCode != "JP" {
		t.Errorf("expected JP, got %s", response.ISOCode)
	}
}

func TestCountryHandler_GetCountryByCode_NotFound(t *testing.T) {
	db := setupCountryTestDB(t)
	seedCountries(t, db)

	handler := NewCountryHandler(db)

	router := gin.New()
	router.GET("/api/v1/countries/code/:code", handler.GetCountryByCode)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/countries/code/XX", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestCountryHandler_ListRegions(t *testing.T) {
	db := setupCountryTestDB(t)
	seedCountries(t, db)

	handler := NewCountryHandler(db)

	router := gin.New()
	router.GET("/api/v1/countries/regions", handler.ListRegions)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/countries/regions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response struct {
		Regions []string `json:"regions"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(response.Regions) != 4 {
		t.Errorf("expected 4 regions, got %d", len(response.Regions))
	}
}

func TestCountryHandler_SearchCountries(t *testing.T) {
	db := setupCountryTestDB(t)
	seedCountries(t, db)

	handler := NewCountryHandler(db)

	router := gin.New()
	router.GET("/api/v1/countries/search", handler.SearchCountries)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/countries/search?q=an", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response struct {
		Countries []CountryResponse `json:"countries"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Should match France, Germany, Japan, Canada (all contain "an")
	if len(response.Countries) != 4 {
		t.Errorf("expected 4 countries matching 'an', got %d", len(response.Countries))
	}
}

func TestCountryHandler_SearchCountries_ByCode(t *testing.T) {
	db := setupCountryTestDB(t)
	seedCountries(t, db)

	handler := NewCountryHandler(db)

	router := gin.New()
	router.GET("/api/v1/countries/search", handler.SearchCountries)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/countries/search?q=BR", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response struct {
		Countries []CountryResponse `json:"countries"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(response.Countries) != 1 {
		t.Errorf("expected 1 country matching 'BR', got %d", len(response.Countries))
	}
	if response.Countries[0].Name != "Brazil" {
		t.Errorf("expected Brazil, got %s", response.Countries[0].Name)
	}
}

func TestCountryHandler_SearchCountries_MissingQuery(t *testing.T) {
	db := setupCountryTestDB(t)

	handler := NewCountryHandler(db)

	router := gin.New()
	router.GET("/api/v1/countries/search", handler.SearchCountries)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/countries/search", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCountryHandler_ListCountries_Empty(t *testing.T) {
	db := setupCountryTestDB(t)
	// Don't seed any countries

	handler := NewCountryHandler(db)

	router := gin.New()
	router.GET("/api/v1/countries", handler.ListCountries)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/countries", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response CountryListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response.Total != 0 {
		t.Errorf("expected 0 countries, got %d", response.Total)
	}
}
