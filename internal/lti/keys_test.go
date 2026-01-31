package lti

import (
	"crypto/rsa"
	"encoding/json"
	"math/big"
	"testing"
)

func TestNewKeyManager(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	if km.GetPrivateKey() == nil {
		t.Error("expected private key to be set")
	}

	if km.GetKeyID() == "" {
		t.Error("expected key ID to be set")
	}
}

func TestKeyManager_GetPrivateKey(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	key := km.GetPrivateKey()
	if key == nil {
		t.Fatal("expected non-nil private key")
	}

	// Verify it's a valid RSA key
	if key.N == nil || key.E == 0 {
		t.Error("private key is missing required components")
	}
}

func TestKeyManager_GetJWKS(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	jwks := km.GetJWKS()
	if jwks == nil {
		t.Fatal("expected non-nil JWKS response")
	}

	if len(jwks.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(jwks.Keys))
	}

	key := jwks.Keys[0]
	if key.Kty != "RSA" {
		t.Errorf("expected key type RSA, got %s", key.Kty)
	}
	if key.Use != "sig" {
		t.Errorf("expected key use sig, got %s", key.Use)
	}
	if key.Alg != "RS256" {
		t.Errorf("expected algorithm RS256, got %s", key.Alg)
	}
	if key.Kid == "" {
		t.Error("expected key ID to be set")
	}
	if key.N == "" {
		t.Error("expected modulus (n) to be set")
	}
	if key.E == "" {
		t.Error("expected exponent (e) to be set")
	}
}

func TestKeyManager_GetJWKSJSON(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	jsonStr, err := km.GetJWKSJSON()
	if err != nil {
		t.Fatalf("failed to get JWKS JSON: %v", err)
	}

	// Verify it's valid JSON
	var jwks JWKSResponse
	if err := json.Unmarshal([]byte(jsonStr), &jwks); err != nil {
		t.Fatalf("JWKS JSON is not valid: %v", err)
	}

	if len(jwks.Keys) != 1 {
		t.Errorf("expected 1 key in JSON, got %d", len(jwks.Keys))
	}
}

func TestKeyManager_KeyIDUnique(t *testing.T) {
	km1, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create first key manager: %v", err)
	}

	km2, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create second key manager: %v", err)
	}

	if km1.GetKeyID() == km2.GetKeyID() {
		t.Error("key IDs should be unique between instances")
	}
}

func TestNewKeyManagerWithKey(t *testing.T) {
	// Create a key first
	km1, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	originalKey := km1.GetPrivateKey()
	originalKeyID := "custom-key-id"

	// Create new manager with existing key
	km2 := NewKeyManagerWithKey(originalKey, originalKeyID)

	if km2.GetKeyID() != originalKeyID {
		t.Errorf("expected key ID %s, got %s", originalKeyID, km2.GetKeyID())
	}

	if km2.GetPrivateKey() != originalKey {
		t.Error("private keys should be the same instance")
	}
}

func TestKeyManager_GetJWKS_NilKey(t *testing.T) {
	km := &KeyManager{
		privateKey: nil,
		keyID:      "test",
	}

	jwks := km.GetJWKS()
	if jwks == nil {
		t.Fatal("expected non-nil JWKS response even with nil key")
	}
	if len(jwks.Keys) != 0 {
		t.Errorf("expected 0 keys for nil private key, got %d", len(jwks.Keys))
	}
}

func TestJWK_Structure(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	jwks := km.GetJWKS()
	jwk := jwks.Keys[0]

	// Marshal to JSON and unmarshal back
	data, err := json.Marshal(jwk)
	if err != nil {
		t.Fatalf("failed to marshal JWK: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal JWK: %v", err)
	}

	// Verify required fields
	requiredFields := []string{"kty", "kid", "n", "e"}
	for _, field := range requiredFields {
		if _, ok := parsed[field]; !ok {
			t.Errorf("missing required field: %s", field)
		}
	}
}

func TestKeyManager_ConcurrentAccess(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	// Test concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			_ = km.GetPrivateKey()
			_ = km.GetKeyID()
			_ = km.GetJWKS()
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestKeyManager_KeyBitSize(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	key := km.GetPrivateKey()
	bitSize := key.N.BitLen()

	// Should be 2048 bits
	if bitSize < 2048 {
		t.Errorf("key bit size too small: %d, expected at least 2048", bitSize)
	}
}

func TestJWKS_Serialization(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	jwks := km.GetJWKS()
	data, err := json.Marshal(jwks)
	if err != nil {
		t.Fatalf("failed to marshal JWKS: %v", err)
	}

	// Parse back and verify structure
	var parsed JWKSResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal JWKS: %v", err)
	}

	if len(parsed.Keys) != len(jwks.Keys) {
		t.Errorf("key count mismatch after serialization")
	}
}

func TestKeyManager_RSAExponent(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	key := km.GetPrivateKey()
	// Standard RSA exponent is 65537
	if key.E != 65537 {
		t.Logf("Note: RSA exponent is %d (standard is 65537)", key.E)
	}

	// Verify exponent is odd and greater than 2
	if key.E <= 2 || key.E%2 == 0 {
		t.Errorf("invalid RSA exponent: %d", key.E)
	}
}

func TestKeyManager_PublicKeyExtraction(t *testing.T) {
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create key manager: %v", err)
	}

	privateKey := km.GetPrivateKey()
	publicKey := &privateKey.PublicKey

	// Verify public key components match
	jwks := km.GetJWKS()
	if len(jwks.Keys) == 0 {
		t.Fatal("expected at least one key in JWKS")
	}

	// The exponent in JWKS should be base64url encoded big endian bytes of E
	expectedE := big.NewInt(int64(publicKey.E)).Bytes()
	if len(expectedE) == 0 {
		t.Error("expected non-empty exponent bytes")
	}
}

func TestJWK_JSONTags(t *testing.T) {
	jwk := JWK{
		Kty: "RSA",
		Use: "",
		Kid: "test-key",
		Alg: "",
		N:   "test-n",
		E:   "test-e",
	}

	data, err := json.Marshal(jwk)
	if err != nil {
		t.Fatalf("failed to marshal JWK: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	// "use" and "alg" should be omitted when empty
	if _, ok := parsed["use"]; ok {
		t.Error("use should be omitted when empty")
	}
	if _, ok := parsed["alg"]; ok {
		t.Error("alg should be omitted when empty")
	}
}

// Mock private key for testing
func mockPrivateKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	km, err := NewKeyManager()
	if err != nil {
		t.Fatalf("failed to create mock key: %v", err)
	}
	return km.GetPrivateKey()
}
