package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	// Server settings
	Port string
	Host string

	// Database settings
	DBDriver    string // "sqlite" or "postgres"
	DatabaseURL string

	// LTI 1.3 settings
	LTIIssuer        string
	LTIClientID      string
	LTIDeploymentID  string
	LTIJWKSEndpoint  string
	LTIAuthEndpoint  string
	LTITokenEndpoint string

	// Session settings
	SessionSecret string
	SessionMaxAge int

	// Development settings
	DemoMode bool // Enable demo login without LTI

	// Storage settings
	StorageType string // "local" or "s3"
	UploadsDir  string // Local directory for uploads
	MaxFileSize int64  // Maximum file size in bytes
}

// Load reads configuration from environment variables with sensible defaults
func Load() *Config {
	return &Config{
		// Server
		Port: getEnv("PORT", "8080"),
		Host: getEnv("HOST", "0.0.0.0"),

		// Database
		DBDriver:    getEnv("DB_DRIVER", "sqlite"),
		DatabaseURL: getEnv("DATABASE_URL", "globe_expedition.db"),

		// LTI 1.3
		LTIIssuer:        getEnv("LTI_ISSUER", ""),
		LTIClientID:      getEnv("LTI_CLIENT_ID", ""),
		LTIDeploymentID:  getEnv("LTI_DEPLOYMENT_ID", ""),
		LTIJWKSEndpoint:  getEnv("LTI_JWKS_ENDPOINT", ""),
		LTIAuthEndpoint:  getEnv("LTI_AUTH_ENDPOINT", ""),
		LTITokenEndpoint: getEnv("LTI_TOKEN_ENDPOINT", ""),

		// Session
		SessionSecret: getEnv("SESSION_SECRET", "change-me-in-production"),
		SessionMaxAge: getEnvInt("SESSION_MAX_AGE", 86400), // 24 hours

		// Development - demo mode enabled by default for SQLite
		DemoMode: getEnvBool("DEMO_MODE", true),

		// Storage
		StorageType: getEnv("STORAGE_TYPE", "local"),
		UploadsDir:  getEnv("UPLOADS_DIR", "./uploads"),
		MaxFileSize: getEnvInt64("MAX_FILE_SIZE", 10*1024*1024), // 10MB default
	}
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

// getEnvInt retrieves an environment variable as int or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getEnvBool retrieves an environment variable as bool or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// getEnvInt64 retrieves an environment variable as int64 or returns a default value
func getEnvInt64(key string, defaultValue int64) int64 {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// IsDevelopment returns true if running with SQLite (dev mode)
func (c *Config) IsDevelopment() bool {
	return c.DBDriver == "sqlite"
}

// IsProduction returns true if running with PostgreSQL (prod mode)
func (c *Config) IsProduction() bool {
	return c.DBDriver == "postgres"
}

// Validate checks that required configuration is present
func (c *Config) Validate() error {
	// In production, require LTI configuration
	if c.IsProduction() {
		if c.LTIClientID == "" {
			return ErrMissingLTIConfig
		}
		if c.SessionSecret == "change-me-in-production" {
			return ErrInsecureSessionSecret
		}
	}
	return nil
}
