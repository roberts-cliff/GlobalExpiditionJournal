package api

import (
	"net/http"

	"globe-expedition-journal/internal/lti"
	"globe-expedition-journal/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DemoHandler handles demo/development endpoints
type DemoHandler struct {
	db             *gorm.DB
	sessionManager *lti.SessionManager
}

// NewDemoHandler creates a new demo handler
func NewDemoHandler(db *gorm.DB, sessionManager *lti.SessionManager) *DemoHandler {
	return &DemoHandler{
		db:             db,
		sessionManager: sessionManager,
	}
}

// DemoLoginRequest represents the demo login request
type DemoLoginRequest struct {
	Name string `json:"name"`
	Role string `json:"role"` // "instructor" or "learner"
}

// DemoLogin creates a demo session without LTI (dev mode only)
// POST /api/v1/demo/login
func (h *DemoHandler) DemoLogin(c *gin.Context) {
	var req DemoLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Name = "Demo Explorer"
		req.Role = "learner"
	}

	if req.Name == "" {
		req.Name = "Demo Explorer"
	}
	if req.Role == "" {
		req.Role = "learner"
	}

	// Find or create demo user
	var user models.User
	demoCanvasID := "demo-user-001"
	demoInstance := "demo.local"

	err := h.db.Where("canvas_user_id = ? AND canvas_instance_url = ?",
		demoCanvasID, demoInstance).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		user = models.User{
			CanvasUserID:      demoCanvasID,
			CanvasInstanceURL: demoInstance,
			DisplayName:       req.Name,
			Email:             "demo@example.com",
		}
		if err := h.db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create demo user"})
			return
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "database error"})
		return
	}

	// Update name if different
	if user.DisplayName != req.Name {
		user.DisplayName = req.Name
		h.db.Save(&user)
	}

	// Create session token
	token, err := h.sessionManager.CreateToken(
		user.ID,
		demoCanvasID,
		"demo-course-001",
		req.Role,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	// Set session cookie
	c.SetCookie(
		"session",
		token,
		86400, // 24 hours
		"/",
		"",
		false, // Not secure for local dev
		true,  // HttpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Demo session created",
		"user": MeResponse{
			ID:          user.ID,
			CanvasID:    demoCanvasID,
			CourseID:    "demo-course-001",
			Role:        req.Role,
			DisplayName: user.DisplayName,
			Email:       user.Email,
		},
	})
}
