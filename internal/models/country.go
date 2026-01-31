package models

// Country represents a country in the world
type Country struct {
	ID      uint   `gorm:"primaryKey" json:"id"`
	Name    string `gorm:"size:255;not null" json:"name"`
	ISOCode string `gorm:"size:3;uniqueIndex;not null" json:"iso_code"` // ISO 3166-1 alpha-2 or alpha-3
	Region  string `gorm:"size:100" json:"region"`                      // e.g., "Europe", "Asia", "Africa"

	// Relationships
	Visits []Visit `gorm:"foreignKey:CountryID" json:"visits,omitempty"`
}

// TableName specifies the table name for Country
func (Country) TableName() string {
	return "countries"
}
