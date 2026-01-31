package seed

import (
	"log"

	"globe-expedition-journal/internal/models"

	"gorm.io/gorm"
)

// Countries populates the countries table with initial data
func Countries(db *gorm.DB) error {
	var count int64
	db.Model(&models.Country{}).Count(&count)
	if count > 0 {
		log.Printf("Countries already seeded (%d records)", count)
		return nil
	}

	countries := []models.Country{
		// Europe
		{Name: "France", ISOCode: "FR", Region: "Europe"},
		{Name: "Germany", ISOCode: "DE", Region: "Europe"},
		{Name: "Italy", ISOCode: "IT", Region: "Europe"},
		{Name: "Spain", ISOCode: "ES", Region: "Europe"},
		{Name: "United Kingdom", ISOCode: "GB", Region: "Europe"},
		{Name: "Netherlands", ISOCode: "NL", Region: "Europe"},
		{Name: "Belgium", ISOCode: "BE", Region: "Europe"},
		{Name: "Switzerland", ISOCode: "CH", Region: "Europe"},
		{Name: "Austria", ISOCode: "AT", Region: "Europe"},
		{Name: "Portugal", ISOCode: "PT", Region: "Europe"},
		{Name: "Greece", ISOCode: "GR", Region: "Europe"},
		{Name: "Sweden", ISOCode: "SE", Region: "Europe"},
		{Name: "Norway", ISOCode: "NO", Region: "Europe"},
		{Name: "Denmark", ISOCode: "DK", Region: "Europe"},
		{Name: "Finland", ISOCode: "FI", Region: "Europe"},
		{Name: "Ireland", ISOCode: "IE", Region: "Europe"},
		{Name: "Poland", ISOCode: "PL", Region: "Europe"},
		{Name: "Czech Republic", ISOCode: "CZ", Region: "Europe"},
		{Name: "Hungary", ISOCode: "HU", Region: "Europe"},
		{Name: "Croatia", ISOCode: "HR", Region: "Europe"},

		// Asia
		{Name: "Japan", ISOCode: "JP", Region: "Asia"},
		{Name: "China", ISOCode: "CN", Region: "Asia"},
		{Name: "South Korea", ISOCode: "KR", Region: "Asia"},
		{Name: "India", ISOCode: "IN", Region: "Asia"},
		{Name: "Thailand", ISOCode: "TH", Region: "Asia"},
		{Name: "Vietnam", ISOCode: "VN", Region: "Asia"},
		{Name: "Indonesia", ISOCode: "ID", Region: "Asia"},
		{Name: "Malaysia", ISOCode: "MY", Region: "Asia"},
		{Name: "Singapore", ISOCode: "SG", Region: "Asia"},
		{Name: "Philippines", ISOCode: "PH", Region: "Asia"},
		{Name: "Taiwan", ISOCode: "TW", Region: "Asia"},

		// North America
		{Name: "United States", ISOCode: "US", Region: "North America"},
		{Name: "Canada", ISOCode: "CA", Region: "North America"},
		{Name: "Mexico", ISOCode: "MX", Region: "North America"},

		// South America
		{Name: "Brazil", ISOCode: "BR", Region: "South America"},
		{Name: "Argentina", ISOCode: "AR", Region: "South America"},
		{Name: "Chile", ISOCode: "CL", Region: "South America"},
		{Name: "Colombia", ISOCode: "CO", Region: "South America"},
		{Name: "Peru", ISOCode: "PE", Region: "South America"},
		{Name: "Ecuador", ISOCode: "EC", Region: "South America"},

		// Africa
		{Name: "South Africa", ISOCode: "ZA", Region: "Africa"},
		{Name: "Egypt", ISOCode: "EG", Region: "Africa"},
		{Name: "Morocco", ISOCode: "MA", Region: "Africa"},
		{Name: "Kenya", ISOCode: "KE", Region: "Africa"},
		{Name: "Nigeria", ISOCode: "NG", Region: "Africa"},
		{Name: "Ghana", ISOCode: "GH", Region: "Africa"},
		{Name: "Tanzania", ISOCode: "TZ", Region: "Africa"},

		// Oceania
		{Name: "Australia", ISOCode: "AU", Region: "Oceania"},
		{Name: "New Zealand", ISOCode: "NZ", Region: "Oceania"},
		{Name: "Fiji", ISOCode: "FJ", Region: "Oceania"},

		// Middle East
		{Name: "United Arab Emirates", ISOCode: "AE", Region: "Middle East"},
		{Name: "Israel", ISOCode: "IL", Region: "Middle East"},
		{Name: "Turkey", ISOCode: "TR", Region: "Middle East"},
		{Name: "Saudi Arabia", ISOCode: "SA", Region: "Middle East"},
		{Name: "Jordan", ISOCode: "JO", Region: "Middle East"},
	}

	for _, country := range countries {
		if err := db.Create(&country).Error; err != nil {
			log.Printf("Warning: failed to seed country %s: %v", country.Name, err)
		}
	}

	log.Printf("Seeded %d countries", len(countries))
	return nil
}
