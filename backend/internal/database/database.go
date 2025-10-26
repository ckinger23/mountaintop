package database

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ckinger23/mountaintop/internal/models"
	"golang.org/x/crypto/bcrypt"
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

	log.Printf("Seeded %d teams", len(teams))

	// Seed admin user
	if err := seedAdminUser(db); err != nil {
		return fmt.Errorf("failed to seed admin user: %w", err)
	}

	// Seed development season, weeks, and games
	if err := seedDevelopmentData(db); err != nil {
		return fmt.Errorf("failed to seed development data: %w", err)
	}

	log.Println("Database seeded successfully")
	return nil
}

// seedAdminUser creates a default admin user for local development
func seedAdminUser(db *gorm.DB) error {
	// Hash the default password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin user
	admin := models.User{
		Username:     "admin",
		Email:        "admin@example.com",
		PasswordHash: string(hashedPassword),
		DisplayName:  "Admin User",
		IsAdmin:      true,
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	log.Println("Created admin user (email: admin@example.com, password: admin123)")
	return nil
}

// seedDevelopmentData creates a sample season with weeks and games for local development
func seedDevelopmentData(db *gorm.DB) error {
	// Create current season
	season := models.Season{
		Year:     2025,
		Name:     "2025 Season",
		IsActive: true,
	}
	if err := db.Create(&season).Error; err != nil {
		return fmt.Errorf("failed to create season: %w", err)
	}
	log.Printf("Created season: %s", season.Name)

	// Get some teams for sample games
	var teams []models.Team
	if err := db.Limit(10).Find(&teams).Error; err != nil {
		return fmt.Errorf("failed to fetch teams: %w", err)
	}

	if len(teams) < 4 {
		return fmt.Errorf("not enough teams to create sample games")
	}

	// Create 3 weeks with sample games
	now := time.Now()
	for weekNum := 1; weekNum <= 3; weekNum++ {
		week := models.Week{
			SeasonID:   season.ID,
			WeekNumber: weekNum,
			Name:       fmt.Sprintf("Week %d", weekNum),
			// Lock time is 7 days from now for week 1, then +7 days for each subsequent week
			LockTime:   now.AddDate(0, 0, 7*weekNum),
		}
		if err := db.Create(&week).Error; err != nil {
			return fmt.Errorf("failed to create week %d: %w", weekNum, err)
		}
		log.Printf("Created %s", week.Name)

		// Create 3-4 sample games per week
		gamesPerWeek := []struct {
			homeIdx  int
			awayIdx  int
			spread   float64
			total    float64
		}{
			{0, 1, -3.5, 50.5},
			{2, 3, -7.0, 55.5},
			{4, 5, 2.5, 48.5},
			{6, 7, -10.5, 52.5},
		}

		for _, gameData := range gamesPerWeek {
			if gameData.homeIdx >= len(teams) || gameData.awayIdx >= len(teams) {
				continue
			}

			game := models.Game{
				WeekID:     week.ID,
				HomeTeamID: teams[gameData.homeIdx].ID,
				AwayTeamID: teams[gameData.awayIdx].ID,
				// Games are 5 days from now for week 1, then +7 days for each week
				GameTime:   now.AddDate(0, 0, 5+(7*(weekNum-1))),
				HomeSpread: gameData.spread,
				Total:      gameData.total,
				IsFinal:    false,
			}
			if err := db.Create(&game).Error; err != nil {
				return fmt.Errorf("failed to create game: %w", err)
			}
		}
		log.Printf("Created 4 games for Week %d", weekNum)
	}

	return nil
}
