package lti

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestNewJWKSHandler(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	handler := NewJWKSHandler(km)
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}

	if handler.GetKeyManager() != km {
		t.Error("key manager not set correctly")
	}
}

func TestJWKSHandler_HandleJWKS(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	handler := NewJWKSHandler(km)

	router := gin.New()
	router.GET("/.well-known/jwks.json", handler.HandleJWKS)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	// Verify response is valid JWKS
	var jwks JWKSResponse
	if err := json.Unmarshal(w.Body.Bytes(), &jwks); err != nil {
		t.Fatalf("response is not valid JWKS JSON: %v", err)
	}

	if len(jwks.Keys) != 1 {
		t.Errorf("expected 1 key, got %d", len(jwks.Keys))
	}
}

func TestJWKSHandler_ContentType(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	handler := NewJWKSHandler(km)

	router := gin.New()
	router.GET("/.well-known/jwks.json", handler.HandleJWKS)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("expected JSON content type, got %s", contentType)
	}
}

func TestJWKSHandler_CacheControl(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	handler := NewJWKSHandler(km)

	router := gin.New()
	router.GET("/.well-known/jwks.json", handler.HandleJWKS)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "public, max-age=3600" {
		t.Errorf("expected cache control header, got %s", cacheControl)
	}
}

func TestJWKSHandler_ResponseStructure(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	handler := NewJWKSHandler(km)

	router := gin.New()
	router.GET("/.well-known/jwks.json", handler.HandleJWKS)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var jwks JWKSResponse
	if err := json.Unmarshal(w.Body.Bytes(), &jwks); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	key := jwks.Keys[0]

	// Verify key structure
	if key.Kty != "RSA" {
		t.Errorf("expected kty=RSA, got %s", key.Kty)
	}
	if key.Use != "sig" {
		t.Errorf("expected use=sig, got %s", key.Use)
	}
	if key.Alg != "RS256" {
		t.Errorf("expected alg=RS256, got %s", key.Alg)
	}
	if key.Kid == "" {
		t.Error("expected kid to be set")
	}
	if key.N == "" {
		t.Error("expected n (modulus) to be set")
	}
	if key.E == "" {
		t.Error("expected e (exponent) to be set")
	}
}

func TestJWKSHandler_GetKeyManager(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	handler := NewJWKSHandler(km)

	if handler.GetKeyManager() != km {
		t.Error("GetKeyManager should return the same key manager")
	}
}

func TestJWKSHandler_MultipleCalls(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	handler := NewJWKSHandler(km)

	router := gin.New()
	router.GET("/.well-known/jwks.json", handler.HandleJWKS)

	// Make multiple calls and ensure consistency
	var firstKid string
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		var jwks JWKSResponse
		if err := json.Unmarshal(w.Body.Bytes(), &jwks); err != nil {
			t.Fatalf("call %d: failed to parse response: %v", i, err)
		}

		if i == 0 {
			firstKid = jwks.Keys[0].Kid
		} else if jwks.Keys[0].Kid != firstKid {
			t.Errorf("call %d: key ID changed, expected %s got %s", i, firstKid, jwks.Keys[0].Kid)
		}
	}
}
