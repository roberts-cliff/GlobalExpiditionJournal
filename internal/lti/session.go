package lti

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// SessionClaims represents the claims stored in a session token
type SessionClaims struct {
	jwt.RegisteredClaims

	UserID   uint   `json:"user_id"`
	CanvasID string `json:"canvas_id"`
	CourseID string `json:"course_id,omitempty"`
	Role     string `json:"role,omitempty"`
}

// SessionManager handles session creation and validation
type SessionManager struct {
	secret []byte
	maxAge time.Duration
}

// NewSessionManager creates a new session manager
func NewSessionManager(secret string, maxAgeSeconds int) *SessionManager {
	return &SessionManager{
		secret: []byte(secret),
		maxAge: time.Duration(maxAgeSeconds) * time.Second,
	}
}

// CreateToken creates a new session token for a user
func (m *SessionManager) CreateToken(userID uint, canvasID string, courseID string, role string) (string, error) {
	now := time.Now()
	claims := SessionClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.maxAge)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
		UserID:   userID,
		CanvasID: canvasID,
		CourseID: courseID,
		Role:     role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// ValidateToken validates a session token and returns the claims
func (m *SessionManager) ValidateToken(tokenString string) (*SessionClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &SessionClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secret, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*SessionClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
