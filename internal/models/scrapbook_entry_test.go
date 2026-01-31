package models

import (
	"testing"
	"time"

	"globe-expedition-journal/internal/database"
)

func TestScrapbookEntryTableName(t *testing.T) {
	s := ScrapbookEntry{}
	if s.TableName() != "scrapbook_entries" {
		t.Errorf("expected table name 'scrapbook_entries', got '%s'", s.TableName())
	}
}

func TestScrapbookEntryCreate(t *testing.T) {
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

	// Create scrapbook entry
	entry := ScrapbookEntry{
		UserID:    user.ID,
		CountryID: country.ID,
		Title:     "My Paris Trip",
		Notes:     "Amazing view from Eiffel Tower",
		VisitedAt: time.Now(),
	}

	result := database.GetDB().Create(&entry)
	if result.Error != nil {
		t.Fatalf("failed to create scrapbook entry: %v", result.Error)
	}

	if entry.ID == 0 {
		t.Error("expected entry ID to be set after create")
	}

	if entry.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	if entry.UpdatedAt.IsZero() {
		t.Error("expected UpdatedAt to be set")
	}
}

func TestScrapbookEntryWithMedia(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := User{
		CanvasUserID:      "12345",
		CanvasInstanceURL: "https://canvas.example.com",
	}
	database.GetDB().Create(&user)

	country := Country{Name: "Japan", ISOCode: "JP", Region: "Asia"}
	database.GetDB().Create(&country)

	entry := ScrapbookEntry{
		UserID:    user.ID,
		CountryID: country.ID,
		Title:     "Cherry Blossoms",
		Notes:     "Beautiful spring in Tokyo",
		MediaURL:  "https://storage.example.com/photos/123.jpg",
		MediaType: "image/jpeg",
		VisitedAt: time.Now(),
	}

	result := database.GetDB().Create(&entry)
	if result.Error != nil {
		t.Fatalf("failed to create scrapbook entry with media: %v", result.Error)
	}

	// Verify media fields are saved
	var loaded ScrapbookEntry
	database.GetDB().First(&loaded, entry.ID)

	if loaded.MediaURL != "https://storage.example.com/photos/123.jpg" {
		t.Errorf("expected media URL to be saved, got '%s'", loaded.MediaURL)
	}

	if loaded.MediaType != "image/jpeg" {
		t.Errorf("expected media type 'image/jpeg', got '%s'", loaded.MediaType)
	}
}

func TestScrapbookEntryWithRelationships(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := User{
		CanvasUserID:      "12345",
		CanvasInstanceURL: "https://canvas.example.com",
		DisplayName:       "Test User",
	}
	database.GetDB().Create(&user)

	country := Country{Name: "France", ISOCode: "FR", Region: "Europe"}
	database.GetDB().Create(&country)

	entry := ScrapbookEntry{
		UserID:    user.ID,
		CountryID: country.ID,
		Title:     "Eiffel Tower Visit",
		VisitedAt: time.Now(),
	}
	database.GetDB().Create(&entry)

	// Load entry with relationships
	var loaded ScrapbookEntry
	database.GetDB().Preload("User").Preload("Country").First(&loaded, entry.ID)

	if loaded.User.DisplayName != "Test User" {
		t.Errorf("expected user display name 'Test User', got '%s'", loaded.User.DisplayName)
	}

	if loaded.Country.Name != "France" {
		t.Errorf("expected country name 'France', got '%s'", loaded.Country.Name)
	}
}

func TestScrapbookEntryUpdate(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := User{
		CanvasUserID:      "12345",
		CanvasInstanceURL: "https://canvas.example.com",
	}
	database.GetDB().Create(&user)

	country := Country{Name: "Germany", ISOCode: "DE", Region: "Europe"}
	database.GetDB().Create(&country)

	entry := ScrapbookEntry{
		UserID:    user.ID,
		CountryID: country.ID,
		Title:     "Old Title",
		Notes:     "Original notes",
	}
	database.GetDB().Create(&entry)

	originalUpdatedAt := entry.UpdatedAt

	// Small delay to ensure UpdatedAt changes
	time.Sleep(10 * time.Millisecond)

	// Update the entry
	entry.Title = "New Title"
	entry.Notes = "Updated notes"
	database.GetDB().Save(&entry)

	// Verify update
	var loaded ScrapbookEntry
	database.GetDB().First(&loaded, entry.ID)

	if loaded.Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", loaded.Title)
	}

	if loaded.Notes != "Updated notes" {
		t.Errorf("expected notes 'Updated notes', got '%s'", loaded.Notes)
	}

	if !loaded.UpdatedAt.After(originalUpdatedAt) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestScrapbookEntrySoftDelete(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := User{
		CanvasUserID:      "12345",
		CanvasInstanceURL: "https://canvas.example.com",
	}
	database.GetDB().Create(&user)

	country := Country{Name: "Italy", ISOCode: "IT", Region: "Europe"}
	database.GetDB().Create(&country)

	entry := ScrapbookEntry{
		UserID:    user.ID,
		CountryID: country.ID,
		Title:     "Rome Trip",
	}
	database.GetDB().Create(&entry)

	// Soft delete
	database.GetDB().Delete(&entry)

	// Should not find with normal query
	var notFound ScrapbookEntry
	result := database.GetDB().First(&notFound, entry.ID)
	if result.Error == nil {
		t.Error("expected entry to be soft deleted")
	}

	// Should find with Unscoped
	var found ScrapbookEntry
	database.GetDB().Unscoped().First(&found, entry.ID)
	if found.ID == 0 {
		t.Error("expected to find soft deleted entry with Unscoped")
	}

	if found.DeletedAt.Time.IsZero() {
		t.Error("expected DeletedAt to be set")
	}
}

func TestScrapbookEntryListByCountry(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := User{
		CanvasUserID:      "12345",
		CanvasInstanceURL: "https://canvas.example.com",
	}
	database.GetDB().Create(&user)

	france := Country{Name: "France", ISOCode: "FR", Region: "Europe"}
	germany := Country{Name: "Germany", ISOCode: "DE", Region: "Europe"}
	database.GetDB().Create(&france)
	database.GetDB().Create(&germany)

	// Create entries for France
	database.GetDB().Create(&ScrapbookEntry{UserID: user.ID, CountryID: france.ID, Title: "Paris 1"})
	database.GetDB().Create(&ScrapbookEntry{UserID: user.ID, CountryID: france.ID, Title: "Paris 2"})
	// Create entry for Germany
	database.GetDB().Create(&ScrapbookEntry{UserID: user.ID, CountryID: germany.ID, Title: "Berlin"})

	// Query entries for France only
	var entries []ScrapbookEntry
	database.GetDB().Where("country_id = ?", france.ID).Find(&entries)

	if len(entries) != 2 {
		t.Errorf("expected 2 entries for France, got %d", len(entries))
	}
}

func TestScrapbookEntryListByUser(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user1 := User{CanvasUserID: "user1", CanvasInstanceURL: "https://canvas.example.com"}
	user2 := User{CanvasUserID: "user2", CanvasInstanceURL: "https://canvas.example.com"}
	database.GetDB().Create(&user1)
	database.GetDB().Create(&user2)

	country := Country{Name: "France", ISOCode: "FR", Region: "Europe"}
	database.GetDB().Create(&country)

	// Create entries for different users
	database.GetDB().Create(&ScrapbookEntry{UserID: user1.ID, CountryID: country.ID, Title: "User1 Entry 1"})
	database.GetDB().Create(&ScrapbookEntry{UserID: user1.ID, CountryID: country.ID, Title: "User1 Entry 2"})
	database.GetDB().Create(&ScrapbookEntry{UserID: user2.ID, CountryID: country.ID, Title: "User2 Entry"})

	// Query entries for user1 only
	var entries []ScrapbookEntry
	database.GetDB().Where("user_id = ?", user1.ID).Find(&entries)

	if len(entries) != 2 {
		t.Errorf("expected 2 entries for user1, got %d", len(entries))
	}

	for _, e := range entries {
		if e.UserID != user1.ID {
			t.Errorf("expected all entries to belong to user1")
		}
	}
}

func TestScrapbookEntryWithTags(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := User{
		CanvasUserID:      "12345",
		CanvasInstanceURL: "https://canvas.example.com",
	}
	database.GetDB().Create(&user)

	country := Country{Name: "Spain", ISOCode: "ES", Region: "Europe"}
	database.GetDB().Create(&country)

	// Create entry with tags
	entry := ScrapbookEntry{
		UserID:    user.ID,
		CountryID: country.ID,
		Title:     "Barcelona Adventure",
		Notes:     "Visited La Sagrada Familia",
		Tags:      "architecture,museum,travel",
	}

	result := database.GetDB().Create(&entry)
	if result.Error != nil {
		t.Fatalf("failed to create scrapbook entry with tags: %v", result.Error)
	}

	// Verify tags are saved
	var loaded ScrapbookEntry
	database.GetDB().First(&loaded, entry.ID)

	if loaded.Tags != "architecture,museum,travel" {
		t.Errorf("expected tags 'architecture,museum,travel', got '%s'", loaded.Tags)
	}
}

func TestScrapbookEntryFilterByTag(t *testing.T) {
	cleanup := setupTestDB(t)
	defer cleanup()

	user := User{
		CanvasUserID:      "12345",
		CanvasInstanceURL: "https://canvas.example.com",
	}
	database.GetDB().Create(&user)

	country := Country{Name: "France", ISOCode: "FR", Region: "Europe"}
	database.GetDB().Create(&country)

	// Create entries with different tags
	database.GetDB().Create(&ScrapbookEntry{UserID: user.ID, CountryID: country.ID, Title: "Museum Visit", Tags: "museum,art"})
	database.GetDB().Create(&ScrapbookEntry{UserID: user.ID, CountryID: country.ID, Title: "Food Tour", Tags: "food,culture"})
	database.GetDB().Create(&ScrapbookEntry{UserID: user.ID, CountryID: country.ID, Title: "Art Gallery", Tags: "museum,art,gallery"})

	// Query entries containing 'museum' tag using LIKE
	var entries []ScrapbookEntry
	database.GetDB().Where("tags LIKE ?", "%museum%").Find(&entries)

	if len(entries) != 2 {
		t.Errorf("expected 2 entries with 'museum' tag, got %d", len(entries))
	}
}
