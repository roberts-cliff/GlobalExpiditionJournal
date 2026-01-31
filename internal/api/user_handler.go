package api

import (
	"net/http"

	"globe-expedition-journal/internal/middleware"
	"globe-expedition-journal/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserHandler handles user-related API endpoints
type UserHandler struct {
	db *gorm.DB
}

// NewUserHandler creates a new user handler
func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// MeResponse represents the response for the /me endpoint
type MeResponse struct {
	ID          uint   `json:"id"`
	CanvasID    string `json:"canvasId"`
	CourseID    string `json:"courseId"`
	Role        string `json:"role"`
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"email,omitempty"`
}

// GetMe returns the current authenticated user's information
// GET /api/v1/me
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	canvasID, _ := middleware.GetCanvasID(c)
	courseID, _ := middleware.GetCourseID(c)
	role, _ := middleware.GetRole(c)

	// Get full user info from database
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	response := MeResponse{
		ID:          user.ID,
		CanvasID:    canvasID,
		CourseID:    courseID,
		Role:        role,
		DisplayName: user.DisplayName,
		Email:       user.Email,
	}

	c.JSON(http.StatusOK, response)
}

// Logout clears the session cookie
// POST /api/v1/logout
func (h *UserHandler) Logout(c *gin.Context) {
	// Clear the session cookie
	c.SetCookie(
		"session",
		"",
		-1, // Expire immediately
		"/",
		"",
		c.Request.TLS != nil,
		true,
	)

	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}
