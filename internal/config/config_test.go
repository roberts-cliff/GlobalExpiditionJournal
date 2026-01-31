package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear any existing env vars
	os.Clearenv()

	cfg := Load()

	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected default host 0.0.0.0, got %s", cfg.Host)
	}
	if cfg.DBDriver != "sqlite" {
		t.Errorf("expected default DB driver sqlite, got %s", cfg.DBDriver)
	}
	if cfg.DatabaseURL != "globe_expedition.db" {
		t.Errorf("expected default database URL globe_expedition.db, got %s", cfg.DatabaseURL)
	}
	if cfg.SessionMaxAge != 86400 {
		t.Errorf("expected default session max age 86400, got %d", cfg.SessionMaxAge)
	}
}

func TestLoad_FromEnv(t *testing.T) {
	// Set environment variables
	os.Setenv("PORT", "3000")
	os.Setenv("DB_DRIVER", "postgres")
	os.Setenv("DATABASE_URL", "postgres://localhost/test")
	os.Setenv("SESSION_MAX_AGE", "3600")
	defer os.Clearenv()

	cfg := Load()

	if cfg.Port != "3000" {
		t.Errorf("expected port 3000, got %s", cfg.Port)
	}
	if cfg.DBDriver != "postgres" {
		t.Errorf("expected DB driver postgres, got %s", cfg.DBDriver)
	}
	if cfg.DatabaseURL != "postgres://localhost/test" {
		t.Errorf("expected database URL postgres://localhost/test, got %s", cfg.DatabaseURL)
	}
	if cfg.SessionMaxAge != 3600 {
		t.Errorf("expected session max age 3600, got %d", cfg.SessionMaxAge)
	}
}

func TestLoad_InvalidInt(t *testing.T) {
	os.Setenv("SESSION_MAX_AGE", "not-a-number")
	defer os.Clearenv()

	cfg := Load()

	// Should fall back to default when parse fails
	if cfg.SessionMaxAge != 86400 {
		t.Errorf("expected default session max age 86400 for invalid int, got %d", cfg.SessionMaxAge)
	}
}

func TestIsDevelopment(t *testing.T) {
	os.Clearenv()
	cfg := Load()

	if !cfg.IsDevelopment() {
		t.Error("expected IsDevelopment to be true with sqlite driver")
	}
	if cfg.IsProduction() {
		t.Error("expected IsProduction to be false with sqlite driver")
	}
}

func TestIsProduction(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	defer os.Clearenv()

	cfg := Load()

	if cfg.IsDevelopment() {
		t.Error("expected IsDevelopment to be false with postgres driver")
	}
	if !cfg.IsProduction() {
		t.Error("expected IsProduction to be true with postgres driver")
	}
}

func TestValidate_Development(t *testing.T) {
	os.Clearenv()
	cfg := Load()

	err := cfg.Validate()
	if err != nil {
		t.Errorf("expected no error in development mode, got %v", err)
	}
}

func TestValidate_Production_MissingLTI(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	os.Setenv("SESSION_SECRET", "secure-secret")
	defer os.Clearenv()

	cfg := Load()

	err := cfg.Validate()
	if err != ErrMissingLTIConfig {
		t.Errorf("expected ErrMissingLTIConfig, got %v", err)
	}
}

func TestValidate_Production_InsecureSecret(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	os.Setenv("LTI_CLIENT_ID", "test-client")
	// SESSION_SECRET not set, uses default
	defer os.Clearenv()

	cfg := Load()

	err := cfg.Validate()
	if err != ErrInsecureSessionSecret {
		t.Errorf("expected ErrInsecureSessionSecret, got %v", err)
	}
}

func TestValidate_Production_Valid(t *testing.T) {
	os.Setenv("DB_DRIVER", "postgres")
	os.Setenv("LTI_CLIENT_ID", "test-client")
	os.Setenv("SESSION_SECRET", "secure-production-secret")
	defer os.Clearenv()

	cfg := Load()

	err := cfg.Validate()
	if err != nil {
		t.Errorf("expected no error with valid production config, got %v", err)
	}
}
