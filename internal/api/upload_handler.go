package api

import (
	"net/http"

	"globe-expedition-journal/internal/middleware"
	"globe-expedition-journal/internal/storage"

	"github.com/gin-gonic/gin"
)

// UploadHandler handles file upload API endpoints
type UploadHandler struct {
	storage *storage.LocalStorage
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(s *storage.LocalStorage) *UploadHandler {
	return &UploadHandler{storage: s}
}

// UploadResponse represents the response after a successful upload
type UploadResponse struct {
	URL      string `json:"url"`
	Filename string `json:"filename"`
}

// Upload handles file uploads
// POST /api/v1/upload
func (h *UploadHandler) Upload(c *gin.Context) {
	_, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	// Get uploaded file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
		return
	}
	defer file.Close()

	// Get content type from header
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		// Try to detect from file extension or default
		contentType = "application/octet-stream"
	}

	// Validate file type
	config := h.storage.GetConfig()
	if !config.IsAllowedType(contentType) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":        "invalid file type",
			"allowedTypes": config.AllowedTypes,
		})
		return
	}

	// Validate file size
	if header.Size > config.MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "file too large",
			"maxSize": config.MaxFileSize,
		})
		return
	}

	// Upload file
	url, err := h.storage.UploadWithMimeType(file, header.Size, contentType)
	if err != nil {
		if err == storage.ErrFileTooLarge {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
			return
		}
		if err == storage.ErrInvalidFileType {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
		return
	}

	c.JSON(http.StatusCreated, UploadResponse{
		URL:      url,
		Filename: header.Filename,
	})
}

// Delete handles file deletion
// DELETE /api/v1/upload/:filename
func (h *UploadHandler) Delete(c *gin.Context) {
	_, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename required"})
		return
	}

	err := h.storage.Delete(filename)
	if err != nil {
		if err == storage.ErrFileNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "file deleted"})
}
