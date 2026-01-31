package models

import (
	"os"
	"testing"
	"time"

	"globe-expedition-journal/internal/config"
	"globe-expedition-journal/internal/database"
)

func setupTestDB(t *testing.T) func() {
	os.Clearenv()
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DATABASE_URL", ":memory:")

	cfg := config.Load()
	_, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	// Run migrations
	err = database.Migrate(AllModels()...)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return func() {
		database.Close()
		os.Clearenv()
	}
}

func TestAllModels(t *testing.T) {
	models := AllModels()
	if len(models) != 4 {
		t.Errorf("expected 4 models, got %d", len(models))
	}
}

func TestUserTableName(t *testing.T) {
	u := User{}
	if u.TableName() != "users" {
		t.Errorf("expected table name 'users', got '%s'", u.TableName())
	}
}

func TestCountryTableName(t *testing.T) {
	c := Country{}
	if c.TableName() != "countries" {
		t.Errorf("expected table name 'countries', got '%s'", c.TableName())
	}
}

func TestVisitTableName(t *testing.T) {
	v := Visit{}
	if v.TableName() != "visits" {
		t.Errorf("expected table name 'visits', got '%s'", v.TableName())
	}
}

func TestUserCreate(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := User{
		CanvasUserID:      "12345",
		CanvasInstanceURL: "https://canvas.example.com",
		DisplayName:       "Test User",
		Email:             "test@example.com",
	}

	result := database.GetDB().Create(&user)
	if result.Error != nil {
		t.Fatalf("failed to create user: %v", result.Error)
	}

	if user.ID == 0 {
		t.Error("expected user ID to be set after create")
	}

	if user.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if user.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestCountryCreate(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	country := Country{
		Name:    "France",
		ISOCode: "FR",
		Region:  "Europe",
	}

	result := database.GetDB().Create(&country)
	if result.Error != nil {
		t.Fatalf("failed to create country: %v", result.Error)
	}

	if country.ID == 0 {
		t.Error("expected country ID to be set after create")
	}
}

func TestCountryISOCodeUnique(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	country1 := Country{Name: "France", ISOCode: "FR", Region: "Europe"}
	database.GetDB().Create(&country1)

	country2 := Country{Name: "Fake France", ISOCode: "FR", Region: "Europe"}
	result := database.GetDB().Create(&country2)

	if result.Error == nil {
		t.Error("expected error for duplicate ISO code")
	}
}

func TestVisitCreate(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Create user and country first
	user := User{
		CanvasUserID:      "12345",
		CanvasInstanceURL: "https://canvas.example.com",
	}
	database.GetDB().Create(&user)

	country := Country{Name: "France", ISOCode: "FR", Region: "Europe"}
	database.GetDB().Create(&country)

	// Create visit
	visit := Visit{
		UserID:    user.ID,
		CountryID: country.ID,
		VisitedAt: time.Now(),
		Notes:     "Amazing trip!",
	}

	result := database.GetDB().Create(&visit)
	if result.Error != nil {
		t.Fatalf("failed to create visit: %v", result.Error)
	}

	if visit.ID == 0 {
		t.Error("expected visit ID to be set after create")
	}

	if visit.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}
}

func TestVisitWithRelationships(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Create user and country
	user := User{
		CanvasUserID:      "12345",
		CanvasInstanceURL: "https://canvas.example.com",
		DisplayName:       "Test User",
	}
	database.GetDB().Create(&user)

	country := Country{Name: "France", ISOCode: "FR", Region: "Europe"}
	database.GetDB().Create(&country)

	// Create visit
	visit := Visit{
		UserID:    user.ID,
		CountryID: country.ID,
		VisitedAt: time.Now(),
	}
	database.GetDB().Create(&visit)

	// Load visit with relationships
	var loadedVisit Visit
	database.GetDB().Preload("User").Preload("Country").First(&loadedVisit, visit.ID)

	if loadedVisit.User.DisplayName != "Test User" {
		t.Errorf("expected user display name 'Test User', got '%s'", loadedVisit.User.DisplayName)
	}

	if loadedVisit.Country.Name != "France" {
		t.Errorf("expected country name 'France', got '%s'", loadedVisit.Country.Name)
	}
}

func TestUserWithVisits(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	// Create user
	user := User{
		CanvasUserID:      "12345",
		CanvasInstanceURL: "https://canvas.example.com",
	}
	database.GetDB().Create(&user)

	// Create countries
	france := Country{Name: "France", ISOCode: "FR", Region: "Europe"}
	germany := Country{Name: "Germany", ISOCode: "DE", Region: "Europe"}
	database.GetDB().Create(&france)
	database.GetDB().Create(&germany)

	// Create visits
	database.GetDB().Create(&Visit{UserID: user.ID, CountryID: france.ID, VisitedAt: time.Now()})
	database.GetDB().Create(&Visit{UserID: user.ID, CountryID: germany.ID, VisitedAt: time.Now()})

	// Load user with visits
	var loadedUser User
	database.GetDB().Preload("Visits").First(&loadedUser, user.ID)

	if len(loadedUser.Visits) != 2 {
		t.Errorf("expected 2 visits, got %d", len(loadedUser.Visits))
	}
}
