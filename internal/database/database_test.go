package database

import (
	"os"
	"testing"

	"globe-expedition-journal/internal/config"
)

func TestConnect_SQLite(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DATABASE_URL", ":memory:")
	defer os.Clearenv()

	cfg := config.Load()
	db, err := Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to SQLite: %v", err)
	}
	defer Close()

	if db == nil {
		t.Fatal("expected db to be non-nil")
	}

	// Verify connection works
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("failed to get sql.DB: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}
}

func TestConnect_UnsupportedDriver(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_DRIVER", "mysql")
	defer os.Clearenv()

	cfg := config.Load()
	_, err := Connect(cfg)
	if err == nil {
		t.Fatal("expected error for unsupported driver")
	}
}

func TestGetDB_BeforeConnect(t *testing.T) {
	// Reset global DB
	DB = nil

	db := GetDB()
	if db != nil {
		t.Error("expected GetDB to return nil before Connect")
	}
}

func TestGetDB_AfterConnect(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DATABASE_URL", ":memory:")
	defer os.Clearenv()
	defer func() { DB = nil }()

	cfg := config.Load()
	_, err := Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer Close()

	db := GetDB()
	if db == nil {
		t.Error("expected GetDB to return non-nil after Connect")
	}
}

func TestMigrate_NotConnected(t *testing.T) {
	// Reset global DB
	DB = nil

	err := Migrate()
	if err == nil {
		t.Error("expected error when migrating without connection")
	}
}

func TestClose_NotConnected(t *testing.T) {
	// Reset global DB
	DB = nil

	err := Close()
	if err != nil {
		t.Errorf("expected no error when closing nil connection, got %v", err)
	}
}

// TestModel is a simple model for migration testing
type TestModel struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"size:255"`
}

func TestMigrate_Success(t *testing.T) {
	os.Clearenv()
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DATABASE_URL", ":memory:")
	defer os.Clearenv()
	defer func() { DB = nil }()

	cfg := config.Load()
	_, err := Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	defer Close()

	err = Migrate(&TestModel{})
	if err != nil {
		t.Errorf("failed to migrate: %v", err)
	}

	// Verify table was created
	if !DB.Migrator().HasTable(&TestModel{}) {
		t.Error("expected test_models table to exist after migration")
	}
}
