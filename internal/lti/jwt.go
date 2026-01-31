package lti

import (
	"context"
	"fmt"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// LTIClaims represents the claims in an LTI 1.3 id_token
type LTIClaims struct {
	jwt.RegisteredClaims

	// User info
	Name   string `json:"name,omitempty"`
	Email  string `json:"email,omitempty"`
	Locale string `json:"locale,omitempty"`

	// LTI claims (using full URIs as keys)
	MessageType        string                 `json:"https://purl.imsglobal.org/spec/lti/claim/message_type,omitempty"`
	Version            string                 `json:"https://purl.imsglobal.org/spec/lti/claim/version,omitempty"`
	DeploymentID       string                 `json:"https://purl.imsglobal.org/spec/lti/claim/deployment_id,omitempty"`
	TargetLinkURI      string                 `json:"https://purl.imsglobal.org/spec/lti/claim/target_link_uri,omitempty"`
	ResourceLink       map[string]interface{} `json:"https://purl.imsglobal.org/spec/lti/claim/resource_link,omitempty"`
	Roles              []string               `json:"https://purl.imsglobal.org/spec/lti/claim/roles,omitempty"`
	Context            map[string]interface{} `json:"https://purl.imsglobal.org/spec/lti/claim/context,omitempty"`
	LaunchPresentation map[string]interface{} `json:"https://purl.imsglobal.org/spec/lti/claim/launch_presentation,omitempty"`
	Custom             map[string]interface{} `json:"https://purl.imsglobal.org/spec/lti/claim/custom,omitempty"`

	// Nonce for replay protection
	Nonce string `json:"nonce,omitempty"`

	// Platform instance claim
	ToolPlatform map[string]interface{} `json:"https://purl.imsglobal.org/spec/lti/claim/tool_platform,omitempty"`
}

// GetContextID returns the context (course) ID if present
func (c *LTIClaims) GetContextID() string {
	if c.Context == nil {
		return ""
	}
	if id, ok := c.Context["id"].(string); ok {
		return id
	}
	return ""
}

// GetContextLabel returns the context (course) label if present
func (c *LTIClaims) GetContextLabel() string {
	if c.Context == nil {
		return ""
	}
	if label, ok := c.Context["label"].(string); ok {
		return label
	}
	return ""
}

// HasRole checks if the user has a specific role
func (c *LTIClaims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// IsInstructor returns true if user has an instructor role
func (c *LTIClaims) IsInstructor() bool {
	instructorRoles := []string{
		"http://purl.imsglobal.org/vocab/lis/v2/membership#Instructor",
		"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Instructor",
	}
	for _, role := range instructorRoles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

// IsLearner returns true if user has a learner role
func (c *LTIClaims) IsLearner() bool {
	learnerRoles := []string{
		"http://purl.imsglobal.org/vocab/lis/v2/membership#Learner",
		"http://purl.imsglobal.org/vocab/lis/v2/institution/person#Student",
	}
	for _, role := range learnerRoles {
		if c.HasRole(role) {
			return true
		}
	}
	return false
}

// JWTValidator validates LTI id_tokens
type JWTValidator struct {
	jwksCache map[string]keyfunc.Keyfunc
}

// NewJWTValidator creates a new JWT validator
func NewJWTValidator() *JWTValidator {
	return &JWTValidator{
		jwksCache: make(map[string]keyfunc.Keyfunc),
	}
}

// ValidateToken validates an LTI id_token and returns the claims
func (v *JWTValidator) ValidateToken(tokenString string, platform *Platform, expectedNonce string) (*LTIClaims, error) {
	// Get or create JWKS keyfunc for this platform
	kf, err := v.getKeyfunc(platform.JWKSEndpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to get JWKS: %w", err)
	}

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &LTIClaims{}, kf.KeyfuncCtx(context.Background()),
		jwt.WithIssuer(platform.Issuer),
		jwt.WithAudience(platform.ClientID),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*LTIClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate nonce
	if claims.Nonce != expectedNonce {
		return nil, fmt.Errorf("nonce mismatch")
	}

	// Validate LTI message type
	if claims.MessageType != "LtiResourceLinkRequest" && claims.MessageType != "LtiDeepLinkingRequest" {
		return nil, fmt.Errorf("unsupported message type: %s", claims.MessageType)
	}

	return claims, nil
}

// getKeyfunc gets or creates a JWKS keyfunc for the given endpoint
func (v *JWTValidator) getKeyfunc(jwksURL string) (keyfunc.Keyfunc, error) {
	if kf, ok := v.jwksCache[jwksURL]; ok {
		return kf, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	kf, err := keyfunc.NewDefaultCtx(ctx, []string{jwksURL})
	if err != nil {
		return nil, err
	}

	v.jwksCache[jwksURL] = kf
	return kf, nil
}
