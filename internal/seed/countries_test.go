package seed

import (
	"testing"

	"globe-expedition-journal/internal/models"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.Country{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func TestCountries(t *testing.T) {
	db := setupTestDB(t)

	err := Countries(db)
	if err != nil {
		t.Fatalf("failed to seed countries: %v", err)
	}

	var count int64
	db.Model(&models.Country{}).Count(&count)

	if count == 0 {
		t.Error("expected countries to be seeded")
	}

	if count < 50 {
		t.Errorf("expected at least 50 countries, got %d", count)
	}
}

func TestCountries_Idempotent(t *testing.T) {
	db := setupTestDB(t)

	Countries(db)
	Countries(db)

	var count int64
	db.Model(&models.Country{}).Count(&count)

	if count > 60 {
		t.Errorf("seeding should be idempotent, got %d countries", count)
	}
}

func TestCountries_VerifyData(t *testing.T) {
	db := setupTestDB(t)
	Countries(db)

	var usa models.Country
	if err := db.Where("iso_code = ?", "US").First(&usa).Error; err != nil {
		t.Error("expected United States to be seeded")
	}
	if usa.Name != "United States" {
		t.Errorf("expected name 'United States', got '%s'", usa.Name)
	}

	var japan models.Country
	if err := db.Where("iso_code = ?", "JP").First(&japan).Error; err != nil {
		t.Error("expected Japan to be seeded")
	}
	if japan.Region != "Asia" {
		t.Errorf("expected Japan region 'Asia', got '%s'", japan.Region)
	}
}

func TestCountries_AllRegions(t *testing.T) {
	db := setupTestDB(t)
	Countries(db)

	var regions []string
	db.Model(&models.Country{}).Distinct().Pluck("region", &regions)

	expectedRegions := []string{
		"Europe", "Asia", "North America", "South America",
		"Africa", "Oceania", "Middle East",
	}

	regionMap := make(map[string]bool)
	for _, r := range regions {
		regionMap[r] = true
	}

	for _, expected := range expectedRegions {
		if !regionMap[expected] {
			t.Errorf("expected region '%s' to be seeded", expected)
		}
	}
}
