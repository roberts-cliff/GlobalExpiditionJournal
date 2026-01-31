package lti

import (
	"os"
	"testing"

	"globe-expedition-journal/internal/config"
	"globe-expedition-journal/internal/database"

	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) (*gorm.DB, func()) {
	os.Clearenv()
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DATABASE_URL", ":memory:")

	cfg := config.Load()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Run migrations for Platform
	err = db.AutoMigrate(&Platform{})
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db, func() {
		database.Close()
		os.Clearenv()
	}
}

func TestPlatformTableName(t *testing.T) {
	p := Platform{}
	if p.TableName() != "lti_platforms" {
		t.Errorf("expected table name 'lti_platforms', got '%s'", p.TableName())
	}
}

func TestPlatformRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlatformRepository(db)

	platform := &Platform{
		Issuer:        "https://canvas.example.com",
		ClientID:      "client-123",
		DeploymentID:  "deploy-456",
		JWKSEndpoint:  "https://canvas.example.com/.well-known/jwks",
		AuthEndpoint:  "https://canvas.example.com/api/lti/authorize",
		TokenEndpoint: "https://canvas.example.com/login/oauth2/token",
		Name:          "Example Canvas",
	}

	err := repo.Create(platform)
	if err != nil {
		t.Fatalf("failed to create platform: %v", err)
	}

	if platform.ID == 0 {
		t.Error("expected platform ID to be set")
	}
}

func TestPlatformRepository_FindByIssuer(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlatformRepository(db)

	// Create a platform
	platform := &Platform{
		Issuer:       "https://canvas.example.com",
		ClientID:     "client-123",
		JWKSEndpoint: "https://canvas.example.com/.well-known/jwks",
		AuthEndpoint: "https://canvas.example.com/api/lti/authorize",
	}
	repo.Create(platform)

	// Find by issuer
	found, err := repo.FindByIssuer("https://canvas.example.com")
	if err != nil {
		t.Fatalf("failed to find platform: %v", err)
	}

	if found.ClientID != "client-123" {
		t.Errorf("expected client ID 'client-123', got '%s'", found.ClientID)
	}
}

func TestPlatformRepository_FindByIssuer_NotFound(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlatformRepository(db)

	_, err := repo.FindByIssuer("https://nonexistent.com")
	if err == nil {
		t.Error("expected error for non-existent issuer")
	}
}

func TestPlatformRepository_FindByClientID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlatformRepository(db)

	platform := &Platform{
		Issuer:       "https://canvas.example.com",
		ClientID:     "unique-client-id",
		JWKSEndpoint: "https://canvas.example.com/.well-known/jwks",
		AuthEndpoint: "https://canvas.example.com/api/lti/authorize",
	}
	repo.Create(platform)

	found, err := repo.FindByClientID("unique-client-id")
	if err != nil {
		t.Fatalf("failed to find platform: %v", err)
	}

	if found.Issuer != "https://canvas.example.com" {
		t.Errorf("expected issuer 'https://canvas.example.com', got '%s'", found.Issuer)
	}
}

func TestPlatformRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlatformRepository(db)

	platform := &Platform{
		Issuer:       "https://canvas.example.com",
		ClientID:     "client-123",
		JWKSEndpoint: "https://canvas.example.com/.well-known/jwks",
		AuthEndpoint: "https://canvas.example.com/api/lti/authorize",
		Name:         "Original Name",
	}
	repo.Create(platform)

	// Update
	platform.Name = "Updated Name"
	err := repo.Update(platform)
	if err != nil {
		t.Fatalf("failed to update platform: %v", err)
	}

	// Verify
	found, _ := repo.FindByID(platform.ID)
	if found.Name != "Updated Name" {
		t.Errorf("expected name 'Updated Name', got '%s'", found.Name)
	}
}

func TestPlatformRepository_Delete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlatformRepository(db)

	platform := &Platform{
		Issuer:       "https://canvas.example.com",
		ClientID:     "client-123",
		JWKSEndpoint: "https://canvas.example.com/.well-known/jwks",
		AuthEndpoint: "https://canvas.example.com/api/lti/authorize",
	}
	repo.Create(platform)

	err := repo.Delete(platform.ID)
	if err != nil {
		t.Fatalf("failed to delete platform: %v", err)
	}

	// Should not find deleted platform
	_, err = repo.FindByID(platform.ID)
	if err == nil {
		t.Error("expected error when finding deleted platform")
	}
}

func TestPlatformRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlatformRepository(db)

	// Create multiple platforms
	repo.Create(&Platform{
		Issuer:       "https://canvas1.example.com",
		ClientID:     "client-1",
		JWKSEndpoint: "https://canvas1.example.com/.well-known/jwks",
		AuthEndpoint: "https://canvas1.example.com/api/lti/authorize",
	})
	repo.Create(&Platform{
		Issuer:       "https://canvas2.example.com",
		ClientID:     "client-2",
		JWKSEndpoint: "https://canvas2.example.com/.well-known/jwks",
		AuthEndpoint: "https://canvas2.example.com/api/lti/authorize",
	})

	platforms, err := repo.List()
	if err != nil {
		t.Fatalf("failed to list platforms: %v", err)
	}

	if len(platforms) != 2 {
		t.Errorf("expected 2 platforms, got %d", len(platforms))
	}
}

func TestPlatformRepository_Upsert_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlatformRepository(db)

	platform := &Platform{
		Issuer:       "https://canvas.example.com",
		ClientID:     "client-123",
		JWKSEndpoint: "https://canvas.example.com/.well-known/jwks",
		AuthEndpoint: "https://canvas.example.com/api/lti/authorize",
	}

	err := repo.Upsert(platform)
	if err != nil {
		t.Fatalf("failed to upsert (create): %v", err)
	}

	if platform.ID == 0 {
		t.Error("expected platform ID to be set after upsert")
	}
}

func TestPlatformRepository_Upsert_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlatformRepository(db)

	// Create initial
	platform := &Platform{
		Issuer:       "https://canvas.example.com",
		ClientID:     "client-123",
		JWKSEndpoint: "https://canvas.example.com/.well-known/jwks",
		AuthEndpoint: "https://canvas.example.com/api/lti/authorize",
		Name:         "Original",
	}
	repo.Create(platform)
	originalID := platform.ID

	// Upsert with same issuer but different data
	updated := &Platform{
		Issuer:       "https://canvas.example.com",
		ClientID:     "client-456",
		JWKSEndpoint: "https://canvas.example.com/.well-known/jwks",
		AuthEndpoint: "https://canvas.example.com/api/lti/authorize",
		Name:         "Updated",
	}

	err := repo.Upsert(updated)
	if err != nil {
		t.Fatalf("failed to upsert (update): %v", err)
	}

	// Should have same ID
	if updated.ID != originalID {
		t.Errorf("expected ID %d, got %d", originalID, updated.ID)
	}

	// Verify updated values
	found, _ := repo.FindByIssuer("https://canvas.example.com")
	if found.ClientID != "client-456" {
		t.Errorf("expected client ID 'client-456', got '%s'", found.ClientID)
	}
	if found.Name != "Updated" {
		t.Errorf("expected name 'Updated', got '%s'", found.Name)
	}
}

func TestPlatformIssuerUnique(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := NewPlatformRepository(db)

	platform1 := &Platform{
		Issuer:       "https://canvas.example.com",
		ClientID:     "client-1",
		JWKSEndpoint: "https://canvas.example.com/.well-known/jwks",
		AuthEndpoint: "https://canvas.example.com/api/lti/authorize",
	}
	repo.Create(platform1)

	platform2 := &Platform{
		Issuer:       "https://canvas.example.com", // Same issuer
		ClientID:     "client-2",
		JWKSEndpoint: "https://canvas.example.com/.well-known/jwks",
		AuthEndpoint: "https://canvas.example.com/api/lti/authorize",
	}

	err := repo.Create(platform2)
	if err == nil {
		t.Error("expected error for duplicate issuer")
	}
}
