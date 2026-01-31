package lti

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// JWKSHandler handles the JWKS endpoint
type JWKSHandler struct {
	keyManager *KeyManager
}

// NewJWKSHandler creates a new JWKS handler
func NewJWKSHandler(keyManager *KeyManager) *JWKSHandler {
	return &JWKSHandler{
		keyManager: keyManager,
	}
}

// HandleJWKS serves the public keys in JWKS format
// GET /.well-known/jwks.json
func (h *JWKSHandler) HandleJWKS(c *gin.Context) {
	jwks := h.keyManager.GetJWKS()

	// Set appropriate headers for JWKS
	c.Header("Cache-Control", "public, max-age=3600")
	c.JSON(http.StatusOK, jwks)
}

// GetKeyManager returns the key manager (for signing operations)
func (h *JWKSHandler) GetKeyManager() *KeyManager {
	return h.keyManager
}
