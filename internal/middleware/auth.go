package middleware

import (
	"net/http"
	"strings"

	"globe-expedition-journal/internal/lti"

	"github.com/gin-gonic/gin"
)

const (
	// ContextKeyUserID is the context key for the user ID
	ContextKeyUserID = "user_id"
	// ContextKeyCanvasID is the context key for the Canvas user ID
	ContextKeyCanvasID = "canvas_id"
	// ContextKeyCourseID is the context key for the course ID
	ContextKeyCourseID = "course_id"
	// ContextKeyRole is the context key for the user role
	ContextKeyRole = "role"
	// ContextKeyClaims is the context key for the full session claims
	ContextKeyClaims = "session_claims"
)

// AuthMiddleware creates a middleware that validates session tokens
func AuthMiddleware(sessionManager *lti.SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing or invalid authorization",
			})
			return
		}

		claims, err := sessionManager.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired session",
			})
			return
		}

		// Set claims in context for handlers
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyCanvasID, claims.CanvasID)
		c.Set(ContextKeyCourseID, claims.CourseID)
		c.Set(ContextKeyRole, claims.Role)
		c.Set(ContextKeyClaims, claims)

		c.Next()
	}
}

// OptionalAuthMiddleware creates a middleware that validates tokens if present
// but does not require authentication
func OptionalAuthMiddleware(sessionManager *lti.SessionManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.Next()
			return
		}

		claims, err := sessionManager.ValidateToken(token)
		if err != nil {
			// Token present but invalid - continue without auth
			c.Next()
			return
		}

		// Set claims in context for handlers
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyCanvasID, claims.CanvasID)
		c.Set(ContextKeyCourseID, claims.CourseID)
		c.Set(ContextKeyRole, claims.Role)
		c.Set(ContextKeyClaims, claims)

		c.Next()
	}
}

// RequireRole creates a middleware that requires a specific role
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(ContextKeyRole)
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "authentication required",
			})
			return
		}

		roleStr, ok := role.(string)
		if !ok || roleStr != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "insufficient permissions",
			})
			return
		}

		c.Next()
	}
}

// RequireInstructor creates a middleware that requires instructor role
func RequireInstructor() gin.HandlerFunc {
	return RequireRole("instructor")
}

// extractToken extracts the session token from cookie or Authorization header
func extractToken(c *gin.Context) string {
	// First, try to get from cookie
	if token, err := c.Cookie("session"); err == nil && token != "" {
		return token
	}

	// Fall back to Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	// Support "Bearer <token>" format
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return parts[1]
	}

	// Support token directly in header
	return authHeader
}

// GetUserID retrieves the user ID from the context
func GetUserID(c *gin.Context) (uint, bool) {
	val, exists := c.Get(ContextKeyUserID)
	if !exists {
		return 0, false
	}
	userID, ok := val.(uint)
	return userID, ok
}

// GetCanvasID retrieves the Canvas user ID from the context
func GetCanvasID(c *gin.Context) (string, bool) {
	val, exists := c.Get(ContextKeyCanvasID)
	if !exists {
		return "", false
	}
	canvasID, ok := val.(string)
	return canvasID, ok
}

// GetCourseID retrieves the course ID from the context
func GetCourseID(c *gin.Context) (string, bool) {
	val, exists := c.Get(ContextKeyCourseID)
	if !exists {
		return "", false
	}
	courseID, ok := val.(string)
	return courseID, ok
}

// GetRole retrieves the user role from the context
func GetRole(c *gin.Context) (string, bool) {
	val, exists := c.Get(ContextKeyRole)
	if !exists {
		return "", false
	}
	role, ok := val.(string)
	return role, ok
}

// GetSessionClaims retrieves the full session claims from the context
func GetSessionClaims(c *gin.Context) (*lti.SessionClaims, bool) {
	val, exists := c.Get(ContextKeyClaims)
	if !exists {
		return nil, false
	}
	claims, ok := val.(*lti.SessionClaims)
	return claims, ok
}

// IsAuthenticated checks if the request has valid authentication
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get(ContextKeyUserID)
	return exists
}

// IsInstructor checks if the authenticated user is an instructor
func IsInstructor(c *gin.Context) bool {
	role, ok := GetRole(c)
	return ok && role == "instructor"
}

// IsLearner checks if the authenticated user is a learner
func IsLearner(c *gin.Context) bool {
	role, ok := GetRole(c)
	return ok && role == "learner"
}
