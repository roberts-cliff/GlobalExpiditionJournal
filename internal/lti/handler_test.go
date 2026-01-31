package lti

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"globe-expedition-journal/internal/config"
	"globe-expedition-journal/internal/database"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupHandlerTestDB(t *testing.T) (*Handler, func()) {
	os.Clearenv()
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DATABASE_URL", ":memory:")

	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Migrate platform table
	db.AutoMigrate(&Platform{})

	handler := NewHandler(db)

	return handler, func() {
		database.Close()
		os.Clearenv()
	}
}

func TestLoginInitiation_MissingIss(t *testing.T) {
	handler, cleanup := setupHandlerTestDB(t)
	defer cleanup()

	router := gin.New()
	router.GET("/lti/login", handler.LoginInitiation)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/login?login_hint=user123&target_link_uri=https://app.com", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "missing iss") {
		t.Errorf("expected error about missing iss, got %s", w.Body.String())
	}
}

func TestLoginInitiation_MissingLoginHint(t *testing.T) {
	handler, cleanup := setupHandlerTestDB(t)
	defer cleanup()

	router := gin.New()
	router.GET("/lti/login", handler.LoginInitiation)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/login?iss=https://canvas.example.com&target_link_uri=https://app.com", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "missing login_hint") {
		t.Errorf("expected error about missing login_hint, got %s", w.Body.String())
	}
}

func TestLoginInitiation_MissingTargetLinkURI(t *testing.T) {
	handler, cleanup := setupHandlerTestDB(t)
	defer cleanup()

	router := gin.New()
	router.GET("/lti/login", handler.LoginInitiation)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/login?iss=https://canvas.example.com&login_hint=user123", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "missing target_link_uri") {
		t.Errorf("expected error about missing target_link_uri, got %s", w.Body.String())
	}
}

func TestLoginInitiation_UnknownPlatform(t *testing.T) {
	handler, cleanup := setupHandlerTestDB(t)
	defer cleanup()

	router := gin.New()
	router.GET("/lti/login", handler.LoginInitiation)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/login?iss=https://unknown.com&login_hint=user123&target_link_uri=https://app.com", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "unknown platform") {
		t.Errorf("expected error about unknown platform, got %s", w.Body.String())
	}
}

func TestLoginInitiation_Success(t *testing.T) {
	handler, cleanup := setupHandlerTestDB(t)
	defer cleanup()

	// Register a platform
	platform := &Platform{
		Issuer:       "https://canvas.example.com",
		ClientID:     "client-123",
		JWKSEndpoint: "https://canvas.example.com/.well-known/jwks",
		AuthEndpoint: "https://canvas.example.com/api/lti/authorize",
	}
	handler.GetPlatformRepo().Create(platform)

	router := gin.New()
	router.GET("/lti/login", handler.LoginInitiation)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/login?iss=https://canvas.example.com&login_hint=user123&target_link_uri=https://app.com/launch", nil)
	req.Host = "localhost:8080"
	router.ServeHTTP(w, req)

	// Should redirect
	if w.Code != http.StatusFound {
		t.Errorf("expected status 302, got %d", w.Code)
	}

	// Check redirect location
	location := w.Header().Get("Location")
	if location == "" {
		t.Fatal("expected Location header")
	}

	redirectURL, err := url.Parse(location)
	if err != nil {
		t.Fatalf("failed to parse redirect URL: %v", err)
	}

	// Verify redirect URL components
	if redirectURL.Host != "canvas.example.com" {
		t.Errorf("expected redirect to canvas.example.com, got %s", redirectURL.Host)
	}

	query := redirectURL.Query()
	if query.Get("client_id") != "client-123" {
		t.Errorf("expected client_id 'client-123', got '%s'", query.Get("client_id"))
	}
	if query.Get("response_type") != "id_token" {
		t.Errorf("expected response_type 'id_token', got '%s'", query.Get("response_type"))
	}
	if query.Get("scope") != "openid" {
		t.Errorf("expected scope 'openid', got '%s'", query.Get("scope"))
	}
	if query.Get("login_hint") != "user123" {
		t.Errorf("expected login_hint 'user123', got '%s'", query.Get("login_hint"))
	}
	if query.Get("state") == "" {
		t.Error("expected state parameter")
	}
	if query.Get("nonce") == "" {
		t.Error("expected nonce parameter")
	}

	// Verify state was stored
	state := query.Get("state")
	stateData, ok := handler.GetStateStore().Peek(state)
	if !ok {
		t.Error("expected state to be stored")
	}
	if stateData.Nonce != query.Get("nonce") {
		t.Error("stored nonce should match nonce in redirect")
	}
}

func TestLoginInitiation_POST(t *testing.T) {
	handler, cleanup := setupHandlerTestDB(t)
	defer cleanup()

	// Register a platform
	platform := &Platform{
		Issuer:       "https://canvas.example.com",
		ClientID:     "client-123",
		JWKSEndpoint: "https://canvas.example.com/.well-known/jwks",
		AuthEndpoint: "https://canvas.example.com/api/lti/authorize",
	}
	handler.GetPlatformRepo().Create(platform)

	router := gin.New()
	router.POST("/lti/login", handler.LoginInitiation)

	form := url.Values{}
	form.Set("iss", "https://canvas.example.com")
	form.Set("login_hint", "user123")
	form.Set("target_link_uri", "https://app.com/launch")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/lti/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = "localhost:8080"
	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("expected status 302, got %d", w.Code)
	}
}

func TestLoginInitiation_ClientIDMismatch(t *testing.T) {
	handler, cleanup := setupHandlerTestDB(t)
	defer cleanup()

	// Register a platform
	platform := &Platform{
		Issuer:       "https://canvas.example.com",
		ClientID:     "client-123",
		JWKSEndpoint: "https://canvas.example.com/.well-known/jwks",
		AuthEndpoint: "https://canvas.example.com/api/lti/authorize",
	}
	handler.GetPlatformRepo().Create(platform)

	router := gin.New()
	router.GET("/lti/login", handler.LoginInitiation)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/lti/login?iss=https://canvas.example.com&login_hint=user123&target_link_uri=https://app.com&client_id=wrong-client", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "client_id mismatch") {
		t.Errorf("expected error about client_id mismatch, got %s", w.Body.String())
	}
}
