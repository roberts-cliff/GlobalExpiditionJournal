package models

import (
	"time"

	"gorm.io/gorm"
)

// Visit represents a user's visit to a country
type Visit struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"not null;index" json:"user_id"`
	CountryID uint           `gorm:"not null;index" json:"country_id"`
	VisitedAt time.Time      `gorm:"not null" json:"visited_at"`
	Notes     string         `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Country Country `gorm:"foreignKey:CountryID" json:"country,omitempty"`
}

// TableName specifies the table name for Visit
func (Visit) TableName() string {
	return "visits"
}

// BeforeCreate hook to set timestamps
func (v *Visit) BeforeCreate(tx *gorm.DB) error {
	if v.CreatedAt.IsZero() {
		v.CreatedAt = time.Now()
	}
	if v.VisitedAt.IsZero() {
		v.VisitedAt = time.Now()
	}
	return nil
}
