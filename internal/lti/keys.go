package lti

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"sync"
)

// KeyManager handles RSA key pairs for LTI tool signing
type KeyManager struct {
	mu         sync.RWMutex
	privateKey *rsa.PrivateKey
	keyID      string
}

// JWKSResponse represents a JWKS (JSON Web Key Set) response
type JWKSResponse struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`           // Key type (RSA)
	Use string `json:"use,omitempty"` // Key use (sig for signing)
	Kid string `json:"kid"`           // Key ID
	Alg string `json:"alg,omitempty"` // Algorithm
	N   string `json:"n"`             // RSA modulus (base64url)
	E   string `json:"e"`             // RSA exponent (base64url)
}

// NewKeyManager creates a new key manager with a generated RSA key pair
func NewKeyManager() (*KeyManager, error) {
	// Generate a 2048-bit RSA key pair
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Generate a random key ID
	keyIDBytes := make([]byte, 8)
	if _, err := rand.Read(keyIDBytes); err != nil {
		return nil, fmt.Errorf("failed to generate key ID: %w", err)
	}
	keyID := base64.RawURLEncoding.EncodeToString(keyIDBytes)

	return &KeyManager{
		privateKey: privateKey,
		keyID:      keyID,
	}, nil
}

// NewKeyManagerWithKey creates a key manager with an existing private key
func NewKeyManagerWithKey(privateKey *rsa.PrivateKey, keyID string) *KeyManager {
	return &KeyManager{
		privateKey: privateKey,
		keyID:      keyID,
	}
}

// GetPrivateKey returns the private key for signing
func (km *KeyManager) GetPrivateKey() *rsa.PrivateKey {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.privateKey
}

// GetKeyID returns the key ID
func (km *KeyManager) GetKeyID() string {
	km.mu.RLock()
	defer km.mu.RUnlock()
	return km.keyID
}

// GetJWKS returns the public key in JWKS format
func (km *KeyManager) GetJWKS() *JWKSResponse {
	km.mu.RLock()
	defer km.mu.RUnlock()

	if km.privateKey == nil {
		return &JWKSResponse{Keys: []JWK{}}
	}

	publicKey := &km.privateKey.PublicKey

	// Encode modulus (n) as base64url
	nBytes := publicKey.N.Bytes()
	nBase64 := base64.RawURLEncoding.EncodeToString(nBytes)

	// Encode exponent (e) as base64url
	eBytes := big.NewInt(int64(publicKey.E)).Bytes()
	eBase64 := base64.RawURLEncoding.EncodeToString(eBytes)

	jwk := JWK{
		Kty: "RSA",
		Use: "sig",
		Kid: km.keyID,
		Alg: "RS256",
		N:   nBase64,
		E:   eBase64,
	}

	return &JWKSResponse{
		Keys: []JWK{jwk},
	}
}

// GetJWKSJSON returns the JWKS as a JSON string
func (km *KeyManager) GetJWKSJSON() (string, error) {
	jwks := km.GetJWKS()
	data, err := json.Marshal(jwks)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JWKS: %w", err)
	}
	return string(data), nil
}
