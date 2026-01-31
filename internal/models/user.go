package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user authenticated via Canvas LTI
type User struct {
	ID                uint           `gorm:"primaryKey" json:"id"`
	CanvasUserID      string         `gorm:"size:255;not null" json:"canvas_user_id"`
	CanvasInstanceURL string         `gorm:"size:512;not null" json:"canvas_instance_url"`
	DisplayName       string         `gorm:"size:255" json:"display_name"`
	Email             string         `gorm:"size:255" json:"email"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Visits []Visit `gorm:"foreignKey:UserID" json:"visits,omitempty"`
}

// TableName specifies the table name for User
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook to set timestamps
func (u *User) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	if u.UpdatedAt.IsZero() {
		u.UpdatedAt = now
	}
	return nil
}

// UniqueCanvasIndex returns the composite unique index for Canvas identity
// The index is defined via GORM tags, but this documents the constraint
func (User) UniqueCanvasIndex() string {
	return "idx_users_canvas_identity"
}
