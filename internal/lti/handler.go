package lti

import (
	"fmt"
	"net/http"
	"net/url"

	"globe-expedition-journal/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler handles LTI 1.3 endpoints
type Handler struct {
	db             *gorm.DB
	platformRepo   *PlatformRepository
	stateStore     *StateStore
	jwtValidator   *JWTValidator
	sessionManager *SessionManager
	frontendURL    string
}

// HandlerConfig holds configuration for the LTI handler
type HandlerConfig struct {
	SessionSecret string
	SessionMaxAge int
	FrontendURL   string
}

// NewHandler creates a new LTI handler
func NewHandler(db *gorm.DB) *Handler {
	return NewHandlerWithConfig(db, HandlerConfig{
		SessionSecret: "change-me-in-production",
		SessionMaxAge: 86400,
		FrontendURL:   "/",
	})
}

// NewHandlerWithConfig creates a new LTI handler with config
func NewHandlerWithConfig(db *gorm.DB, cfg HandlerConfig) *Handler {
	return &Handler{
		db:             db,
		platformRepo:   NewPlatformRepository(db),
		stateStore:     NewStateStore(),
		jwtValidator:   NewJWTValidator(),
		sessionManager: NewSessionManager(cfg.SessionSecret, cfg.SessionMaxAge),
		frontendURL:    cfg.FrontendURL,
	}
}

// LoginInitiation handles the OIDC login initiation request from the platform
// GET/POST /lti/login
func (h *Handler) LoginInitiation(c *gin.Context) {
	// Extract parameters (can come from query or form)
	iss := c.DefaultQuery("iss", c.PostForm("iss"))
	loginHint := c.DefaultQuery("login_hint", c.PostForm("login_hint"))
	targetLinkURI := c.DefaultQuery("target_link_uri", c.PostForm("target_link_uri"))
	clientID := c.DefaultQuery("client_id", c.PostForm("client_id"))
	ltiMessageHint := c.DefaultQuery("lti_message_hint", c.PostForm("lti_message_hint"))

	// Validate required parameters
	if iss == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing iss parameter"})
		return
	}
	if loginHint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing login_hint parameter"})
		return
	}
	if targetLinkURI == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing target_link_uri parameter"})
		return
	}

	// Find the platform by issuer
	platform, err := h.platformRepo.FindByIssuer(iss)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown platform issuer"})
		return
	}

	// If client_id provided, verify it matches
	if clientID != "" && clientID != platform.ClientID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id mismatch"})
		return
	}

	// Generate state and nonce
	state, err := GenerateState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate state"})
		return
	}

	nonce, err := GenerateNonce()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate nonce"})
		return
	}

	// Store state for later validation
	h.stateStore.Store(state, &StateData{
		Nonce:         nonce,
		TargetLinkURI: targetLinkURI,
		ClientID:      platform.ClientID,
	})

	// Build authorization redirect URL
	authURL, err := url.Parse(platform.AuthEndpoint)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid auth endpoint"})
		return
	}

	// Get the launch endpoint URL (where Canvas will redirect back)
	launchURL := getLaunchURL(c.Request)

	q := authURL.Query()
	q.Set("scope", "openid")
	q.Set("response_type", "id_token")
	q.Set("client_id", platform.ClientID)
	q.Set("redirect_uri", launchURL)
	q.Set("login_hint", loginHint)
	q.Set("state", state)
	q.Set("response_mode", "form_post")
	q.Set("nonce", nonce)
	q.Set("prompt", "none")
	if ltiMessageHint != "" {
		q.Set("lti_message_hint", ltiMessageHint)
	}
	authURL.RawQuery = q.Encode()

	// Redirect to platform authorization endpoint
	c.Redirect(http.StatusFound, authURL.String())
}

// Launch handles the LTI launch callback with id_token
// POST /lti/launch
func (h *Handler) Launch(c *gin.Context) {
	// Get id_token and state from form post
	idToken := c.PostForm("id_token")
	state := c.PostForm("state")

	if idToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id_token"})
		return
	}
	if state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing state"})
		return
	}

	// Retrieve and validate state
	stateData, ok := h.stateStore.Get(state)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid or expired state"})
		return
	}

	// Find platform by client ID
	platform, err := h.platformRepo.FindByClientID(stateData.ClientID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "platform not found"})
		return
	}

	// Validate the JWT token
	claims, err := h.jwtValidator.ValidateToken(idToken, platform, stateData.Nonce)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": fmt.Sprintf("token validation failed: %v", err)})
		return
	}

	// Find or create user
	user, err := h.findOrCreateUser(claims, platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process user"})
		return
	}

	// Determine role
	role := "learner"
	if claims.IsInstructor() {
		role = "instructor"
	}

	// Create session token
	sessionToken, err := h.sessionManager.CreateToken(
		user.ID,
		claims.Subject,
		claims.GetContextID(),
		role,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	// Set session cookie
	c.SetCookie(
		"session",
		sessionToken,
		int(h.sessionManager.maxAge.Seconds()),
		"/",
		"",
		c.Request.TLS != nil, // Secure if HTTPS
		true,                 // HttpOnly
	)

	// Redirect to frontend
	redirectURL := h.frontendURL
	if stateData.TargetLinkURI != "" {
		redirectURL = stateData.TargetLinkURI
	}
	c.Redirect(http.StatusFound, redirectURL)
}

// findOrCreateUser finds an existing user or creates a new one
func (h *Handler) findOrCreateUser(claims *LTIClaims, platform *Platform) (*models.User, error) {
	var user models.User

	// Try to find existing user
	err := h.db.Where("canvas_user_id = ? AND canvas_instance_url = ?",
		claims.Subject, platform.Issuer).First(&user).Error

	if err == gorm.ErrRecordNotFound {
		// Create new user
		user = models.User{
			CanvasUserID:      claims.Subject,
			CanvasInstanceURL: platform.Issuer,
			DisplayName:       claims.Name,
			Email:             claims.Email,
		}
		if err := h.db.Create(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}

	if err != nil {
		return nil, err
	}

	// Update user info if changed
	updated := false
	if claims.Name != "" && user.DisplayName != claims.Name {
		user.DisplayName = claims.Name
		updated = true
	}
	if claims.Email != "" && user.Email != claims.Email {
		user.Email = claims.Email
		updated = true
	}
	if updated {
		h.db.Save(&user)
	}

	return &user, nil
}

// GetStateStore returns the state store (for testing)
func (h *Handler) GetStateStore() *StateStore {
	return h.stateStore
}

// GetPlatformRepo returns the platform repository (for testing)
func (h *Handler) GetPlatformRepo() *PlatformRepository {
	return h.platformRepo
}

// GetSessionManager returns the session manager (for testing)
func (h *Handler) GetSessionManager() *SessionManager {
	return h.sessionManager
}

// getLaunchURL constructs the launch callback URL
func getLaunchURL(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil {
		scheme = "http"
	}
	// Check for forwarded headers (behind proxy)
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	}
	host := r.Host
	if fwdHost := r.Header.Get("X-Forwarded-Host"); fwdHost != "" {
		host = fwdHost
	}
	return fmt.Sprintf("%s://%s/lti/launch", scheme, host)
}
