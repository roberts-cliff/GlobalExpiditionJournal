package api

import (
	"net/http"
	"strconv"
	"time"

	"globe-expedition-journal/internal/middleware"
	"globe-expedition-journal/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// VisitHandler handles visit-related API endpoints
type VisitHandler struct {
	db *gorm.DB
}

// NewVisitHandler creates a new visit handler
func NewVisitHandler(db *gorm.DB) *VisitHandler {
	return &VisitHandler{db: db}
}

// VisitResponse represents a visit in API responses
type VisitResponse struct {
	ID        uint             `json:"id"`
	CountryID uint             `json:"countryId"`
	VisitedAt string           `json:"visitedAt"`
	Notes     string           `json:"notes,omitempty"`
	Country   *CountryResponse `json:"country,omitempty"`
}

// VisitListResponse represents the response for listing visits
type VisitListResponse struct {
	Visits []VisitResponse `json:"visits"`
	Total  int64           `json:"total"`
}

// CreateVisitRequest represents the request body for creating a visit
type CreateVisitRequest struct {
	CountryID uint   `json:"countryId" binding:"required"`
	VisitedAt string `json:"visitedAt"` // Optional, defaults to now
	Notes     string `json:"notes"`
}

// UpdateVisitRequest represents the request body for updating a visit
type UpdateVisitRequest struct {
	VisitedAt string `json:"visitedAt"`
	Notes     string `json:"notes"`
}

// toVisitResponse converts a model to a response
func toVisitResponse(v *models.Visit, includeCountry bool) VisitResponse {
	resp := VisitResponse{
		ID:        v.ID,
		CountryID: v.CountryID,
		VisitedAt: v.VisitedAt.Format(time.RFC3339),
		Notes:     v.Notes,
	}

	if includeCountry && v.Country.ID != 0 {
		country := toCountryResponse(&v.Country)
		resp.Country = &country
	}

	return resp
}

// ListVisits returns all visits for the authenticated user
// GET /api/v1/visits
func (h *VisitHandler) ListVisits(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var visits []models.Visit
	query := h.db.Where("user_id = ?", userID).Preload("Country")

	// Get total count
	var total int64
	h.db.Model(&models.Visit{}).Where("user_id = ?", userID).Count(&total)

	// Get visits (ordered by visit date, most recent first)
	if err := query.Order("visited_at DESC").Find(&visits).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch visits"})
		return
	}

	response := VisitListResponse{
		Visits: make([]VisitResponse, len(visits)),
		Total:  total,
	}

	for i, visit := range visits {
		response.Visits[i] = toVisitResponse(&visit, true)
	}

	c.JSON(http.StatusOK, response)
}

// GetVisit returns a specific visit
// GET /api/v1/visits/:id
func (h *VisitHandler) GetVisit(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visit ID"})
		return
	}

	var visit models.Visit
	if err := h.db.Preload("Country").Where("id = ? AND user_id = ?", id, userID).First(&visit).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "visit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch visit"})
		return
	}

	c.JSON(http.StatusOK, toVisitResponse(&visit, true))
}

// CreateVisit creates a new visit
// POST /api/v1/visits
func (h *VisitHandler) CreateVisit(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req CreateVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Verify country exists
	var country models.Country
	if err := h.db.First(&country, req.CountryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "country not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify country"})
		return
	}

	// Parse visit date or use current time
	visitedAt := time.Now()
	if req.VisitedAt != "" {
		parsed, err := time.Parse(time.RFC3339, req.VisitedAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visitedAt format, use RFC3339"})
			return
		}
		visitedAt = parsed
	}

	visit := models.Visit{
		UserID:    userID,
		CountryID: req.CountryID,
		VisitedAt: visitedAt,
		Notes:     req.Notes,
	}

	if err := h.db.Create(&visit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create visit"})
		return
	}

	// Load country for response
	visit.Country = country

	c.JSON(http.StatusCreated, toVisitResponse(&visit, true))
}

// UpdateVisit updates an existing visit
// PUT /api/v1/visits/:id
func (h *VisitHandler) UpdateVisit(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visit ID"})
		return
	}

	var req UpdateVisitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Find existing visit
	var visit models.Visit
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&visit).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "visit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch visit"})
		return
	}

	// Update fields
	if req.VisitedAt != "" {
		parsed, err := time.Parse(time.RFC3339, req.VisitedAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visitedAt format, use RFC3339"})
			return
		}
		visit.VisitedAt = parsed
	}
	visit.Notes = req.Notes

	if err := h.db.Save(&visit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update visit"})
		return
	}

	// Load country for response
	h.db.First(&visit.Country, visit.CountryID)

	c.JSON(http.StatusOK, toVisitResponse(&visit, true))
}

// DeleteVisit deletes a visit
// DELETE /api/v1/visits/:id
func (h *VisitHandler) DeleteVisit(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visit ID"})
		return
	}

	// Verify visit exists and belongs to user
	var visit models.Visit
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&visit).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "visit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch visit"})
		return
	}

	if err := h.db.Delete(&visit).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete visit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "visit deleted"})
}

// GetVisitsByCountry returns all visits for a specific country
// GET /api/v1/visits/country/:countryId
func (h *VisitHandler) GetVisitsByCountry(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	countryIDStr := c.Param("countryId")
	countryID, err := strconv.ParseUint(countryIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid country ID"})
		return
	}

	var visits []models.Visit
	if err := h.db.Where("user_id = ? AND country_id = ?", userID, countryID).
		Preload("Country").
		Order("visited_at DESC").
		Find(&visits).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch visits"})
		return
	}

	response := make([]VisitResponse, len(visits))
	for i, visit := range visits {
		response[i] = toVisitResponse(&visit, true)
	}

	c.JSON(http.StatusOK, gin.H{"visits": response})
}
