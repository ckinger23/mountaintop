package database

import (
	"fmt"
	"log"

	"github.com/ckinger23/mountaintop/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Connect initializes the database connection
func Connect(dbPath string) error {
	var err error

	// Configure GORM
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Connect to SQLite (can easily swap to PostgreSQL later)
	DB, err = gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection established")
	return nil
}

// Migrate runs all database migrations
func Migrate() error {
	log.Println("Running database migrations...")

	err := DB.AutoMigrate(
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
func SeedData() error {
	log.Println("Seeding database with initial data...")

	// Check if data already exists
	var count int64
	DB.Model(&models.Team{}).Count(&count)
	if count > 0 {
		log.Println("Database already seeded, skipping...")
		return nil
	}

	// Add some sample teams
	teams := []models.Team{
		{Name: "Alabama", Abbreviation: "ALA", Conference: "SEC"},
		{Name: "Georgia", Abbreviation: "UGA", Conference: "SEC"},
		{Name: "Ohio State", Abbreviation: "OSU", Conference: "Big Ten"},
		{Name: "Michigan", Abbreviation: "MICH", Conference: "Big Ten"},
		{Name: "Texas", Abbreviation: "TEX", Conference: "SEC"},
		{Name: "USC", Abbreviation: "USC", Conference: "Big Ten"},
		{Name: "Oregon", Abbreviation: "ORE", Conference: "Big Ten"},
		{Name: "Penn State", Abbreviation: "PSU", Conference: "Big Ten"},
	}

	for _, team := range teams {
		if err := DB.Create(&team).Error; err != nil {
			return fmt.Errorf("failed to seed team %s: %w", team.Name, err)
		}
	}

	log.Println("Database seeded successfully")
	return nil
}
