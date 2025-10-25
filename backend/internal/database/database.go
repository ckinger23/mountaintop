package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/ckinger23/mountaintop/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Connect initializes the database connection and returns the DB instance
func Connect(dbPath string) (*gorm.DB, error) {
	// Configure GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Connect to SQLite (can easily swap to PostgreSQL later)
	db, err := gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection established")
	return db, nil
}

// Migrate runs all database migrations
func Migrate(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&models.User{},
		&models.Season{},
		&models.Week{},
		&models.Team{},
		&models.Game{},
		&models.Pick{},
	)

	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

// SeedData adds initial data for testing/development
func SeedData(db *gorm.DB) error {
	log.Println("Seeding database with initial data...")

	// Check if data already exists
	var count int64
	db.Model(&models.Team{}).Count(&count)
	if count > 0 {
		log.Println("Database already seeded, skipping...")
		return nil
	}

	// Load teams from JSON file
	data, err := os.ReadFile("data/teams.json")
	if err != nil {
		return fmt.Errorf("failed to read teams.json: %w", err)
	}

	var teams []models.Team
	if err := json.Unmarshal(data, &teams); err != nil {
		return fmt.Errorf("failed to parse teams.json: %w", err)
	}

	// Bulk insert all teams
	if err := db.Create(&teams).Error; err != nil {
		return fmt.Errorf("failed to seed teams: %w", err)
	}

	log.Printf("Database seeded successfully with %d teams", len(teams))
	return nil
}
