package api

import (
	"net/http"
	"strconv"

	"globe-expedition-journal/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CountryHandler handles country-related API endpoints
type CountryHandler struct {
	db *gorm.DB
}

// NewCountryHandler creates a new country handler
func NewCountryHandler(db *gorm.DB) *CountryHandler {
	return &CountryHandler{db: db}
}

// CountryResponse represents a country in API responses
type CountryResponse struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	ISOCode string `json:"isoCode"`
	Region  string `json:"region,omitempty"`
}

// CountryListResponse represents the response for listing countries
type CountryListResponse struct {
	Countries []CountryResponse `json:"countries"`
	Total     int64             `json:"total"`
}

// toCountryResponse converts a model to a response
func toCountryResponse(c *models.Country) CountryResponse {
	return CountryResponse{
		ID:      c.ID,
		Name:    c.Name,
		ISOCode: c.ISOCode,
		Region:  c.Region,
	}
}

// ListCountries returns all countries
// GET /api/v1/countries
func (h *CountryHandler) ListCountries(c *gin.Context) {
	// Optional filters
	region := c.Query("region")

	var countries []models.Country
	query := h.db.Model(&models.Country{})

	if region != "" {
		query = query.Where("region = ?", region)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get countries (ordered by name)
	if err := query.Order("name ASC").Find(&countries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch countries"})
		return
	}

	response := CountryListResponse{
		Countries: make([]CountryResponse, len(countries)),
		Total:     total,
	}

	for i, country := range countries {
		response.Countries[i] = toCountryResponse(&country)
	}

	c.JSON(http.StatusOK, response)
}

// GetCountry returns a specific country by ID
// GET /api/v1/countries/:id
func (h *CountryHandler) GetCountry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid country ID"})
		return
	}

	var country models.Country
	if err := h.db.First(&country, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "country not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch country"})
		return
	}

	c.JSON(http.StatusOK, toCountryResponse(&country))
}

// GetCountryByCode returns a country by ISO code
// GET /api/v1/countries/code/:code
func (h *CountryHandler) GetCountryByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing country code"})
		return
	}

	var country models.Country
	if err := h.db.Where("iso_code = ?", code).First(&country).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "country not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch country"})
		return
	}

	c.JSON(http.StatusOK, toCountryResponse(&country))
}

// ListRegions returns all unique regions
// GET /api/v1/countries/regions
func (h *CountryHandler) ListRegions(c *gin.Context) {
	var regions []string
	if err := h.db.Model(&models.Country{}).Distinct().Pluck("region", &regions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch regions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"regions": regions})
}

// SearchCountries searches countries by name
// GET /api/v1/countries/search?q=query
func (h *CountryHandler) SearchCountries(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing search query"})
		return
	}

	var countries []models.Country
	searchPattern := "%" + query + "%"

	if err := h.db.Where("name LIKE ? OR iso_code LIKE ?", searchPattern, searchPattern).
		Order("name ASC").
		Limit(20).
		Find(&countries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search countries"})
		return
	}

	response := make([]CountryResponse, len(countries))
	for i, country := range countries {
		response[i] = toCountryResponse(&country)
	}

	c.JSON(http.StatusOK, gin.H{"countries": response})
}
