package app

import (
	"gorm.io/gorm"
)

// App holds all application dependencies
type App struct {
	DB *gorm.DB
}

// NewApp creates a new App instance with the provided dependencies
func NewApp(db *gorm.DB) *App {
	return &App{
		DB: db,
	}
}
