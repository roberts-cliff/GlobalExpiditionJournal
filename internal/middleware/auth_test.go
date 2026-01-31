package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"globe-expedition-journal/internal/lti"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func createTestSessionManager() *lti.SessionManager {
	return lti.NewSessionManager("test-secret-key-12345", 3600)
}

func createTestToken(sm *lti.SessionManager, userID uint, canvasID, courseID, role string) string {
	token, _ := sm.CreateToken(userID, canvasID, courseID, role)
	return token
}

func TestAuthMiddleware_ValidCookie(t *testing.T) {
	sm := createTestSessionManager()
	token := createTestToken(sm, 123, "canvas-1", "course-1", "learner")

	router := gin.New()
	router.Use(AuthMiddleware(sm))
	router.GET("/test", func(c *gin.Context) {
		userID, _ := GetUserID(c)
		c.JSON(200, gin.H{"user_id": userID})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_ValidBearerToken(t *testing.T) {
	sm := createTestSessionManager()
	token := createTestToken(sm, 456, "canvas-2", "course-2", "instructor")

	router := gin.New()
	router.Use(AuthMiddleware(sm))
	router.GET("/test", func(c *gin.Context) {
		role, _ := GetRole(c)
		c.JSON(200, gin.H{"role": role})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	sm := createTestSessionManager()

	router := gin.New()
	router.Use(AuthMiddleware(sm))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	sm := createTestSessionManager()

	router := gin.New()
	router.Use(AuthMiddleware(sm))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "invalid-token"})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	sm1 := createTestSessionManager()
	sm2 := lti.NewSessionManager("different-secret", 3600)
	token := createTestToken(sm1, 123, "canvas", "course", "learner")

	router := gin.New()
	router.Use(AuthMiddleware(sm2)) // Different secret
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestOptionalAuthMiddleware_WithToken(t *testing.T) {
	sm := createTestSessionManager()
	token := createTestToken(sm, 789, "canvas-3", "course-3", "learner")

	router := gin.New()
	router.Use(OptionalAuthMiddleware(sm))
	router.GET("/test", func(c *gin.Context) {
		if IsAuthenticated(c) {
			userID, _ := GetUserID(c)
			c.JSON(200, gin.H{"user_id": userID})
		} else {
			c.JSON(200, gin.H{"anonymous": true})
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestOptionalAuthMiddleware_WithoutToken(t *testing.T) {
	sm := createTestSessionManager()

	router := gin.New()
	router.Use(OptionalAuthMiddleware(sm))
	router.GET("/test", func(c *gin.Context) {
		if IsAuthenticated(c) {
			c.JSON(200, gin.H{"authenticated": true})
		} else {
			c.JSON(200, gin.H{"anonymous": true})
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestOptionalAuthMiddleware_InvalidToken(t *testing.T) {
	sm := createTestSessionManager()

	router := gin.New()
	router.Use(OptionalAuthMiddleware(sm))
	router.GET("/test", func(c *gin.Context) {
		if IsAuthenticated(c) {
			c.JSON(200, gin.H{"authenticated": true})
		} else {
			c.JSON(200, gin.H{"anonymous": true})
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "bad-token"})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should continue as anonymous, not fail
	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRequireRole_Authorized(t *testing.T) {
	sm := createTestSessionManager()
	token := createTestToken(sm, 1, "canvas", "course", "instructor")

	router := gin.New()
	router.Use(AuthMiddleware(sm))
	router.Use(RequireRole("instructor"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestRequireRole_Unauthorized(t *testing.T) {
	sm := createTestSessionManager()
	token := createTestToken(sm, 1, "canvas", "course", "learner")

	router := gin.New()
	router.Use(AuthMiddleware(sm))
	router.Use(RequireRole("instructor"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status 403, got %d", w.Code)
	}
}

func TestRequireRole_NoAuth(t *testing.T) {
	router := gin.New()
	router.Use(RequireRole("instructor"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestRequireInstructor(t *testing.T) {
	sm := createTestSessionManager()
	token := createTestToken(sm, 1, "canvas", "course", "instructor")

	router := gin.New()
	router.Use(AuthMiddleware(sm))
	router.Use(RequireInstructor())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestGetUserID(t *testing.T) {
	sm := createTestSessionManager()
	token := createTestToken(sm, 42, "canvas", "course", "learner")

	router := gin.New()
	router.Use(AuthMiddleware(sm))
	router.GET("/test", func(c *gin.Context) {
		userID, ok := GetUserID(c)
		if !ok {
			t.Error("expected user ID to be present")
		}
		if userID != 42 {
			t.Errorf("expected user ID 42, got %d", userID)
		}
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
}

func TestGetCanvasID(t *testing.T) {
	sm := createTestSessionManager()
	token := createTestToken(sm, 1, "my-canvas-id", "course", "learner")

	router := gin.New()
	router.Use(AuthMiddleware(sm))
	router.GET("/test", func(c *gin.Context) {
		canvasID, ok := GetCanvasID(c)
		if !ok {
			t.Error("expected canvas ID to be present")
		}
		if canvasID != "my-canvas-id" {
			t.Errorf("expected canvas ID 'my-canvas-id', got '%s'", canvasID)
		}
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
}

func TestGetCourseID(t *testing.T) {
	sm := createTestSessionManager()
	token := createTestToken(sm, 1, "canvas", "course-xyz", "learner")

	router := gin.New()
	router.Use(AuthMiddleware(sm))
	router.GET("/test", func(c *gin.Context) {
		courseID, ok := GetCourseID(c)
		if !ok {
			t.Error("expected course ID to be present")
		}
		if courseID != "course-xyz" {
			t.Errorf("expected course ID 'course-xyz', got '%s'", courseID)
		}
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
}

func TestGetSessionClaims(t *testing.T) {
	sm := createTestSessionManager()
	token := createTestToken(sm, 99, "canvas-99", "course-99", "instructor")

	router := gin.New()
	router.Use(AuthMiddleware(sm))
	router.GET("/test", func(c *gin.Context) {
		claims, ok := GetSessionClaims(c)
		if !ok {
			t.Error("expected claims to be present")
		}
		if claims.UserID != 99 {
			t.Errorf("expected user ID 99 in claims, got %d", claims.UserID)
		}
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
}

func TestIsAuthenticated(t *testing.T) {
	sm := createTestSessionManager()
	token := createTestToken(sm, 1, "canvas", "course", "learner")

	router := gin.New()
	router.Use(OptionalAuthMiddleware(sm))
	router.GET("/test", func(c *gin.Context) {
		auth := IsAuthenticated(c)
		c.JSON(200, gin.H{"authenticated": auth})
	})

	// With token
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: token})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Without token
	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
}

func TestIsInstructor(t *testing.T) {
	sm := createTestSessionManager()

	tests := []struct {
		role     string
		expected bool
	}{
		{"instructor", true},
		{"learner", false},
		{"admin", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			token := createTestToken(sm, 1, "canvas", "course", tt.role)

			router := gin.New()
			router.Use(AuthMiddleware(sm))
			router.GET("/test", func(c *gin.Context) {
				result := IsInstructor(c)
				if result != tt.expected {
					t.Errorf("IsInstructor() = %v, want %v for role %s", result, tt.expected, tt.role)
				}
				c.JSON(200, gin.H{"ok": true})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.AddCookie(&http.Cookie{Name: "session", Value: token})
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		})
	}
}

func TestIsLearner(t *testing.T) {
	sm := createTestSessionManager()

	tests := []struct {
		role     string
		expected bool
	}{
		{"learner", true},
		{"instructor", false},
		{"student", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			token := createTestToken(sm, 1, "canvas", "course", tt.role)

			router := gin.New()
			router.Use(AuthMiddleware(sm))
			router.GET("/test", func(c *gin.Context) {
				result := IsLearner(c)
				if result != tt.expected {
					t.Errorf("IsLearner() = %v, want %v for role %s", result, tt.expected, tt.role)
				}
				c.JSON(200, gin.H{"ok": true})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.AddCookie(&http.Cookie{Name: "session", Value: token})
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
		})
	}
}

func TestExtractToken_DirectHeader(t *testing.T) {
	sm := createTestSessionManager()
	token := createTestToken(sm, 1, "canvas", "course", "learner")

	router := gin.New()
	router.Use(AuthMiddleware(sm))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", token) // Without "Bearer " prefix
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
}

func TestExtractToken_CookiePriority(t *testing.T) {
	sm := createTestSessionManager()
	cookieToken := createTestToken(sm, 100, "cookie-user", "course", "learner")
	headerToken := createTestToken(sm, 200, "header-user", "course", "learner")

	router := gin.New()
	router.Use(AuthMiddleware(sm))
	router.GET("/test", func(c *gin.Context) {
		userID, _ := GetUserID(c)
		// Cookie should take priority
		if userID != 100 {
			t.Errorf("expected cookie token to take priority, got user ID %d", userID)
		}
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: cookieToken})
	req.Header.Set("Authorization", "Bearer "+headerToken)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
}

func TestGetHelpers_NoAuth(t *testing.T) {
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		_, ok1 := GetUserID(c)
		_, ok2 := GetCanvasID(c)
		_, ok3 := GetCourseID(c)
		_, ok4 := GetRole(c)
		_, ok5 := GetSessionClaims(c)

		if ok1 || ok2 || ok3 || ok4 || ok5 {
			t.Error("expected all getters to return false when not authenticated")
		}
		c.JSON(200, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
}
