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

// ScrapbookHandler handles scrapbook entry API endpoints
type ScrapbookHandler struct {
	db *gorm.DB
}

// NewScrapbookHandler creates a new scrapbook handler
func NewScrapbookHandler(db *gorm.DB) *ScrapbookHandler {
	return &ScrapbookHandler{db: db}
}

// ScrapbookEntryResponse represents a scrapbook entry in API responses
type ScrapbookEntryResponse struct {
	ID        uint             `json:"id"`
	CountryID uint             `json:"countryId"`
	Title     string           `json:"title"`
	Notes     string           `json:"notes,omitempty"`
	MediaURL  string           `json:"mediaUrl,omitempty"`
	MediaType string           `json:"mediaType,omitempty"`
	Tags      string           `json:"tags,omitempty"`
	VisitedAt string           `json:"visitedAt,omitempty"`
	CreatedAt string           `json:"createdAt"`
	UpdatedAt string           `json:"updatedAt"`
	Country   *CountryResponse `json:"country,omitempty"`
}

// ScrapbookEntryListResponse represents the response for listing entries
type ScrapbookEntryListResponse struct {
	Entries []ScrapbookEntryResponse `json:"entries"`
	Total   int64                    `json:"total"`
}

// CreateScrapbookEntryRequest represents the request body for creating an entry
type CreateScrapbookEntryRequest struct {
	CountryID uint   `json:"countryId" binding:"required"`
	Title     string `json:"title" binding:"required"`
	Notes     string `json:"notes"`
	MediaURL  string `json:"mediaUrl"`
	MediaType string `json:"mediaType"`
	Tags      string `json:"tags"`
	VisitedAt string `json:"visitedAt"`
}

// UpdateScrapbookEntryRequest represents the request body for updating an entry
type UpdateScrapbookEntryRequest struct {
	Title     string `json:"title"`
	Notes     string `json:"notes"`
	MediaURL  string `json:"mediaUrl"`
	MediaType string `json:"mediaType"`
	Tags      string `json:"tags"`
	VisitedAt string `json:"visitedAt"`
}

// ScrapbookStatsResponse represents user statistics
type ScrapbookStatsResponse struct {
	TotalEntries        int64 `json:"totalEntries"`
	CountriesDocumented int64 `json:"countriesDocumented"`
	PhotosUploaded      int64 `json:"photosUploaded"`
}

// toScrapbookEntryResponse converts a model to a response
func toScrapbookEntryResponse(e *models.ScrapbookEntry, includeCountry bool) ScrapbookEntryResponse {
	resp := ScrapbookEntryResponse{
		ID:        e.ID,
		CountryID: e.CountryID,
		Title:     e.Title,
		Notes:     e.Notes,
		MediaURL:  e.MediaURL,
		MediaType: e.MediaType,
		Tags:      e.Tags,
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
		UpdatedAt: e.UpdatedAt.Format(time.RFC3339),
	}

	if !e.VisitedAt.IsZero() {
		resp.VisitedAt = e.VisitedAt.Format(time.RFC3339)
	}

	if includeCountry && e.Country.ID != 0 {
		country := toCountryResponse(&e.Country)
		resp.Country = &country
	}

	return resp
}

// ListEntries returns all scrapbook entries for the authenticated user
// GET /api/v1/scrapbook/entries
// Query params: tag (optional) - filter by tag using LIKE match
func (h *ScrapbookHandler) ListEntries(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var entries []models.ScrapbookEntry
	query := h.db.Where("user_id = ?", userID).Preload("Country")

	// Filter by tag if provided
	tagFilter := c.Query("tag")
	if tagFilter != "" {
		query = query.Where("tags LIKE ?", "%"+tagFilter+"%")
	}

	// Get total count (with tag filter if applied)
	var total int64
	countQuery := h.db.Model(&models.ScrapbookEntry{}).Where("user_id = ?", userID)
	if tagFilter != "" {
		countQuery = countQuery.Where("tags LIKE ?", "%"+tagFilter+"%")
	}
	countQuery.Count(&total)

	// Get entries (ordered by creation date, most recent first)
	if err := query.Order("created_at DESC").Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch entries"})
		return
	}

	response := ScrapbookEntryListResponse{
		Entries: make([]ScrapbookEntryResponse, len(entries)),
		Total:   total,
	}

	for i, entry := range entries {
		response.Entries[i] = toScrapbookEntryResponse(&entry, true)
	}

	c.JSON(http.StatusOK, response)
}

// GetEntry returns a specific scrapbook entry
// GET /api/v1/scrapbook/entries/:id
func (h *ScrapbookHandler) GetEntry(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry ID"})
		return
	}

	var entry models.ScrapbookEntry
	if err := h.db.Preload("Country").Where("id = ? AND user_id = ?", id, userID).First(&entry).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch entry"})
		return
	}

	c.JSON(http.StatusOK, toScrapbookEntryResponse(&entry, true))
}

// CreateEntry creates a new scrapbook entry
// POST /api/v1/scrapbook/entries
func (h *ScrapbookHandler) CreateEntry(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var req CreateScrapbookEntryRequest
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

	entry := models.ScrapbookEntry{
		UserID:    userID,
		CountryID: req.CountryID,
		Title:     req.Title,
		Notes:     req.Notes,
		MediaURL:  req.MediaURL,
		MediaType: req.MediaType,
		Tags:      req.Tags,
	}

	// Parse visit date if provided
	if req.VisitedAt != "" {
		parsed, err := time.Parse(time.RFC3339, req.VisitedAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visitedAt format, use RFC3339"})
			return
		}
		entry.VisitedAt = parsed
	}

	if err := h.db.Create(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create entry"})
		return
	}

	// Load country for response
	entry.Country = country

	c.JSON(http.StatusCreated, toScrapbookEntryResponse(&entry, true))
}

// UpdateEntry updates an existing scrapbook entry
// PUT /api/v1/scrapbook/entries/:id
func (h *ScrapbookHandler) UpdateEntry(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry ID"})
		return
	}

	var req UpdateScrapbookEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Find existing entry
	var entry models.ScrapbookEntry
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&entry).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch entry"})
		return
	}

	// Update fields if provided
	if req.Title != "" {
		entry.Title = req.Title
	}
	entry.Notes = req.Notes
	entry.MediaURL = req.MediaURL
	entry.MediaType = req.MediaType
	entry.Tags = req.Tags

	if req.VisitedAt != "" {
		parsed, err := time.Parse(time.RFC3339, req.VisitedAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visitedAt format, use RFC3339"})
			return
		}
		entry.VisitedAt = parsed
	}

	if err := h.db.Save(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update entry"})
		return
	}

	// Load country for response
	h.db.First(&entry.Country, entry.CountryID)

	c.JSON(http.StatusOK, toScrapbookEntryResponse(&entry, true))
}

// DeleteEntry deletes a scrapbook entry
// DELETE /api/v1/scrapbook/entries/:id
func (h *ScrapbookHandler) DeleteEntry(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry ID"})
		return
	}

	// Verify entry exists and belongs to user
	var entry models.ScrapbookEntry
	if err := h.db.Where("id = ? AND user_id = ?", id, userID).First(&entry).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch entry"})
		return
	}

	if err := h.db.Delete(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete entry"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "entry deleted"})
}

// GetEntriesByCountry returns all scrapbook entries for a specific country
// GET /api/v1/scrapbook/countries/:countryId/entries
func (h *ScrapbookHandler) GetEntriesByCountry(c *gin.Context) {
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

	var entries []models.ScrapbookEntry
	if err := h.db.Where("user_id = ? AND country_id = ?", userID, countryID).
		Preload("Country").
		Order("created_at DESC").
		Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch entries"})
		return
	}

	response := make([]ScrapbookEntryResponse, len(entries))
	for i, entry := range entries {
		response[i] = toScrapbookEntryResponse(&entry, true)
	}

	c.JSON(http.StatusOK, gin.H{"entries": response})
}

// GetStats returns scrapbook statistics for the authenticated user
// GET /api/v1/scrapbook/stats
func (h *ScrapbookHandler) GetStats(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	var stats ScrapbookStatsResponse

	// Total entries
	h.db.Model(&models.ScrapbookEntry{}).Where("user_id = ?", userID).Count(&stats.TotalEntries)

	// Countries documented (distinct countries with entries)
	h.db.Model(&models.ScrapbookEntry{}).
		Where("user_id = ?", userID).
		Distinct("country_id").
		Count(&stats.CountriesDocumented)

	// Photos uploaded (entries with media_url)
	h.db.Model(&models.ScrapbookEntry{}).
		Where("user_id = ? AND media_url != ''", userID).
		Count(&stats.PhotosUploaded)

	c.JSON(http.StatusOK, stats)
}
