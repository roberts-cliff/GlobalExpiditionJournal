package models

import (
	"time"

	"gorm.io/gorm"
)

// ScrapbookEntry represents a memory/entry in a user's scrapbook for a country
type ScrapbookEntry struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	CountryID uint           `gorm:"not null;index" json:"country_id"`
	Title     string         `gorm:"size:255;not null" json:"title"`
	Notes     string         `gorm:"type:text" json:"notes,omitempty"`
	MediaURL  string         `gorm:"size:512" json:"media_url,omitempty"`
	MediaType string         `gorm:"size:50" json:"media_type,omitempty"`
	Tags      string         `gorm:"size:500" json:"tags,omitempty"` // Comma-separated tags
	VisitedAt time.Time      `json:"visited_at,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Country Country `gorm:"foreignKey:CountryID" json:"country,omitempty"`
}

// TableName specifies the table name for ScrapbookEntry
func (ScrapbookEntry) TableName() string {
	return "scrapbook_entries"
}

// BeforeCreate hook to set timestamps
func (s *ScrapbookEntry) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	if s.UpdatedAt.IsZero() {
		s.UpdatedAt = now
	}
	return nil
}

// BeforeUpdate hook to update timestamp
func (s *ScrapbookEntry) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return nil
}
