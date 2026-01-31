package lti

import (
	"time"

	"gorm.io/gorm"
)

// Platform represents an LTI 1.3 platform (e.g., Canvas LMS instance)
type Platform struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	Issuer        string         `gorm:"size:512;uniqueIndex;not null" json:"issuer"` // e.g., https://canvas.instructure.com
	ClientID      string         `gorm:"size:255;not null" json:"client_id"`
	DeploymentID  string         `gorm:"size:255" json:"deployment_id"`
	JWKSEndpoint  string         `gorm:"size:512;not null" json:"jwks_endpoint"` // Platform's public keys
	AuthEndpoint  string         `gorm:"size:512;not null" json:"auth_endpoint"` // OIDC authorization URL
	TokenEndpoint string         `gorm:"size:512" json:"token_endpoint"`         // For LTI Advantage services
	Name          string         `gorm:"size:255" json:"name"`                   // Friendly name
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName specifies the table name for Platform
func (Platform) TableName() string {
	return "lti_platforms"
}

// PlatformRepository handles database operations for platforms
type PlatformRepository struct {
	db *gorm.DB
}

// NewPlatformRepository creates a new platform repository
func NewPlatformRepository(db *gorm.DB) *PlatformRepository {
	return &PlatformRepository{db: db}
}

// Create adds a new platform registration
func (r *PlatformRepository) Create(platform *Platform) error {
	return r.db.Create(platform).Error
}

// FindByIssuer finds a platform by its issuer URL
func (r *PlatformRepository) FindByIssuer(issuer string) (*Platform, error) {
	var platform Platform
	err := r.db.Where("issuer = ?", issuer).First(&platform).Error
	if err != nil {
		return nil, err
	}
	return &platform, nil
}

// FindByID finds a platform by ID
func (r *PlatformRepository) FindByID(id uint) (*Platform, error) {
	var platform Platform
	err := r.db.First(&platform, id).Error
	if err != nil {
		return nil, err
	}
	return &platform, nil
}

// FindByClientID finds a platform by client ID
func (r *PlatformRepository) FindByClientID(clientID string) (*Platform, error) {
	var platform Platform
	err := r.db.Where("client_id = ?", clientID).First(&platform).Error
	if err != nil {
		return nil, err
	}
	return &platform, nil
}

// Update updates an existing platform
func (r *PlatformRepository) Update(platform *Platform) error {
	return r.db.Save(platform).Error
}

// Delete soft-deletes a platform
func (r *PlatformRepository) Delete(id uint) error {
	return r.db.Delete(&Platform{}, id).Error
}

// List returns all registered platforms
func (r *PlatformRepository) List() ([]Platform, error) {
	var platforms []Platform
	err := r.db.Find(&platforms).Error
	return platforms, err
}

// Upsert creates or updates a platform based on issuer
func (r *PlatformRepository) Upsert(platform *Platform) error {
	existing, err := r.FindByIssuer(platform.Issuer)
	if err == nil {
		// Update existing
		platform.ID = existing.ID
		platform.CreatedAt = existing.CreatedAt
		return r.Update(platform)
	}
	// Create new
	return r.Create(platform)
}
