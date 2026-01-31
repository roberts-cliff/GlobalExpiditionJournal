package models

// AllModels returns all models for migration
func AllModels() []interface{} {
	return []interface{}{
		&User{},
		&Country{},
		&Visit{},
		&ScrapbookEntry{},
	}
}
