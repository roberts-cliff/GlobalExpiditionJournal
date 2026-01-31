package lti

import (
	"testing"
	"time"
)

func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager("test-secret", 3600)
	if sm == nil {
		t.Fatal("expected session manager to be created")
	}
	if sm.maxAge != time.Hour {
		t.Errorf("expected maxAge to be 1 hour, got %v", sm.maxAge)
	}
}

func TestSessionManager_CreateAndValidateToken(t *testing.T) {
	sm := NewSessionManager("test-secret-key-12345", 3600)

	// Create token
	token, err := sm.CreateToken(123, "canvas-user-1", "course-456", "instructor")
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	// Validate token
	claims, err := sm.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.UserID != 123 {
		t.Errorf("expected UserID 123, got %d", claims.UserID)
	}
	if claims.CanvasID != "canvas-user-1" {
		t.Errorf("expected CanvasID 'canvas-user-1', got '%s'", claims.CanvasID)
	}
	if claims.CourseID != "course-456" {
		t.Errorf("expected CourseID 'course-456', got '%s'", claims.CourseID)
	}
	if claims.Role != "instructor" {
		t.Errorf("expected Role 'instructor', got '%s'", claims.Role)
	}
}

func TestSessionManager_ValidateToken_InvalidToken(t *testing.T) {
	sm := NewSessionManager("test-secret", 3600)

	_, err := sm.ValidateToken("invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestSessionManager_ValidateToken_WrongSecret(t *testing.T) {
	sm1 := NewSessionManager("secret-1", 3600)
	sm2 := NewSessionManager("secret-2", 3600)

	token, err := sm1.CreateToken(1, "user", "course", "learner")
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	_, err = sm2.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error when validating with wrong secret")
	}
}

func TestSessionManager_ValidateToken_ExpiredToken(t *testing.T) {
	// Create session manager with 1 second expiry
	sm := NewSessionManager("test-secret", 1)

	token, err := sm.CreateToken(1, "user", "course", "learner")
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Wait for token to expire
	time.Sleep(2 * time.Second)

	_, err = sm.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestSessionManager_CreateToken_EmptyOptionalFields(t *testing.T) {
	sm := NewSessionManager("test-secret", 3600)

	token, err := sm.CreateToken(1, "user", "", "")
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	claims, err := sm.ValidateToken(token)
	if err != nil {
		t.Fatalf("failed to validate token: %v", err)
	}

	if claims.CourseID != "" {
		t.Errorf("expected empty CourseID, got '%s'", claims.CourseID)
	}
	if claims.Role != "" {
		t.Errorf("expected empty Role, got '%s'", claims.Role)
	}
}
