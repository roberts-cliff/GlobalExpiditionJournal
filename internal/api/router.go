package api

import (
	"log"

	"globe-expedition-journal/internal/lti"
	"globe-expedition-journal/internal/middleware"
	"globe-expedition-journal/internal/storage"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewRouter creates and configures the Gin router
func NewRouter() *gin.Engine {
	router := gin.Default()

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", HealthCheck)
	}

	return router
}

// RouterConfig holds configuration for the router
type RouterConfig struct {
	SessionSecret string
	SessionMaxAge int
	DemoMode      bool   // Enable demo login without LTI
	UploadsDir    string // Directory for file uploads
}

// DefaultRouterConfig returns the default router configuration
func DefaultRouterConfig() RouterConfig {
	return RouterConfig{
		SessionSecret: "change-me-in-production",
		SessionMaxAge: 86400,
		DemoMode:      true,        // Enable by default for dev
		UploadsDir:    "./uploads", // Default uploads directory
	}
}

// NewRouterWithDB creates a router with database-dependent handlers
func NewRouterWithDB(db *gorm.DB) *gin.Engine {
	return NewRouterWithConfig(db, DefaultRouterConfig())
}

// NewRouterWithConfig creates a router with custom configuration
func NewRouterWithConfig(db *gorm.DB, cfg RouterConfig) *gin.Engine {
	router := gin.Default()

	// CORS middleware for development
	if cfg.DemoMode {
		router.Use(corsMiddleware())
	}

	// Create session manager for auth middleware
	sessionManager := lti.NewSessionManager(cfg.SessionSecret, cfg.SessionMaxAge)

	// API v1 routes - public
	v1 := router.Group("/api/v1")
	{
		v1.GET("/health", HealthCheck)
	}

	// Demo routes (dev mode only)
	if cfg.DemoMode {
		demoHandler := NewDemoHandler(db, sessionManager)
		demo := router.Group("/api/v1/demo")
		{
			demo.POST("/login", demoHandler.DemoLogin)
		}
		log.Println("Demo mode enabled: POST /api/v1/demo/login")
	}

	// Country routes (public, read-only)
	countryHandler := NewCountryHandler(db)
	countries := router.Group("/api/v1/countries")
	{
		countries.GET("", countryHandler.ListCountries)
		countries.GET("/regions", countryHandler.ListRegions)
		countries.GET("/search", countryHandler.SearchCountries)
		countries.GET("/code/:code", countryHandler.GetCountryByCode)
		countries.GET("/:id", countryHandler.GetCountry)
	}

	// API v1 routes - authenticated
	userHandler := NewUserHandler(db)
	visitHandler := NewVisitHandler(db)
	scrapbookHandler := NewScrapbookHandler(db)
	v1Auth := router.Group("/api/v1")
	v1Auth.Use(middleware.AuthMiddleware(sessionManager))
	{
		v1Auth.GET("/me", userHandler.GetMe)
		v1Auth.POST("/logout", userHandler.Logout)

		// Visit routes
		v1Auth.GET("/visits", visitHandler.ListVisits)
		v1Auth.POST("/visits", visitHandler.CreateVisit)
		v1Auth.GET("/visits/:id", visitHandler.GetVisit)
		v1Auth.PUT("/visits/:id", visitHandler.UpdateVisit)
		v1Auth.DELETE("/visits/:id", visitHandler.DeleteVisit)
		v1Auth.GET("/visits/country/:countryId", visitHandler.GetVisitsByCountry)

		// Scrapbook routes
		v1Auth.GET("/scrapbook/entries", scrapbookHandler.ListEntries)
		v1Auth.POST("/scrapbook/entries", scrapbookHandler.CreateEntry)
		v1Auth.GET("/scrapbook/entries/:id", scrapbookHandler.GetEntry)
		v1Auth.PUT("/scrapbook/entries/:id", scrapbookHandler.UpdateEntry)
		v1Auth.DELETE("/scrapbook/entries/:id", scrapbookHandler.DeleteEntry)
		v1Auth.GET("/scrapbook/countries/:countryId/entries", scrapbookHandler.GetEntriesByCountry)
		v1Auth.GET("/scrapbook/stats", scrapbookHandler.GetStats)
	}

	// File upload handling
	storageConfig := storage.DefaultConfig()
	storageConfig.UploadsDir = cfg.UploadsDir
	localStorage, err := storage.NewLocalStorage(storageConfig)
	if err != nil {
		log.Printf("Warning: failed to initialize storage: %v", err)
	} else {
		uploadHandler := NewUploadHandler(localStorage)
		v1Auth := router.Group("/api/v1")
		v1Auth.Use(middleware.AuthMiddleware(sessionManager))
		{
			v1Auth.POST("/upload", uploadHandler.Upload)
			v1Auth.DELETE("/upload/:filename", uploadHandler.Delete)
		}

		// Static file serving for uploads
		router.Static("/uploads", cfg.UploadsDir)
		log.Printf("Serving uploads from: %s", cfg.UploadsDir)
	}

	// Initialize key manager for JWKS
	keyManager, err := lti.NewKeyManager()
	if err != nil {
		log.Printf("Warning: failed to initialize key manager: %v", err)
	}

	// LTI routes
	ltiHandler := lti.NewHandlerWithConfig(db, lti.HandlerConfig{
		SessionSecret: cfg.SessionSecret,
		SessionMaxAge: cfg.SessionMaxAge,
		FrontendURL:   "/",
	})
	ltiGroup := router.Group("/lti")
	{
		ltiGroup.GET("/login", ltiHandler.LoginInitiation)
		ltiGroup.POST("/login", ltiHandler.LoginInitiation)
		ltiGroup.POST("/launch", ltiHandler.Launch)
	}

	// JWKS endpoint (well-known)
	if keyManager != nil {
		jwksHandler := lti.NewJWKSHandler(keyManager)
		wellKnown := router.Group("/.well-known")
		{
			wellKnown.GET("/jwks.json", jwksHandler.HandleJWKS)
		}
	}

	return router
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// HealthCheck handles the health check endpoint
func HealthCheck(c *gin.Context) {
	c.JSON(200, HealthResponse{Status: "healthy"})
}

// corsMiddleware adds CORS headers for development
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == "" {
			origin = "*"
		}

		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
