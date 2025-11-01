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
		&models.League{},           // NEW: Must come before User (foreign key)
		&models.LeagueMembership{}, // NEW
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

	// After migrations, backfill existing data with default league
	if err := backfillDefaultLeague(db); err != nil {
		return fmt.Errorf("failed to backfill default league: %w", err)
	}

	return nil
}

// backfillDefaultLeague creates a default league and migrates existing data to it
func backfillDefaultLeague(db *gorm.DB) error {
	// Check if default league already exists
	var existingLeague models.League
	if err := db.Where("code = ?", "DEFAULT-LEAGUE").First(&existingLeague).Error; err == nil {
		// Default league already exists, skip
		return nil
	}

	// Check if there's any data that needs migration
	var seasonCount, pickCount int64
	db.Model(&models.Season{}).Count(&seasonCount)
	db.Model(&models.Pick{}).Count(&pickCount)

	// If no seasons or picks exist, no need to backfill (fresh install)
	if seasonCount == 0 && pickCount == 0 {
		log.Println("Fresh install detected, skipping default league backfill")
		return nil
	}

	log.Println("Backfilling existing data with default league...")

	// Find the admin user (or first user) to be the league owner
	var adminUser models.User
	if err := db.Where("is_admin = ?", true).First(&adminUser).Error; err != nil {
		// If no admin, use first user
		if err := db.First(&adminUser).Error; err != nil {
			return fmt.Errorf("no users found to assign as league owner: %w", err)
		}
	}

	// Create default league
	defaultLeague := models.League{
		Name:        "Main League",
		Code:        "DEFAULT-LEAGUE",
		Description: "Default league for existing data",
		OwnerID:     adminUser.ID,
		IsPublic:    true,
		IsActive:    true,
	}

	if err := db.Create(&defaultLeague).Error; err != nil {
		return fmt.Errorf("failed to create default league: %w", err)
	}

	log.Printf("Created default league (ID: %d, Owner: %s)", defaultLeague.ID, adminUser.Username)

	// Update all existing seasons to belong to default league
	if err := db.Model(&models.Season{}).Where("league_id = 0 OR league_id IS NULL").Update("league_id", defaultLeague.ID).Error; err != nil {
		return fmt.Errorf("failed to update seasons: %w", err)
	}

	// Update all existing picks to belong to default league
	if err := db.Model(&models.Pick{}).Where("league_id = 0 OR league_id IS NULL").Update("league_id", defaultLeague.ID).Error; err != nil {
		return fmt.Errorf("failed to update picks: %w", err)
	}

	// Create league memberships for all existing users
	var allUsers []models.User
	if err := db.Find(&allUsers).Error; err != nil {
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	for _, user := range allUsers {
		role := "member"
		if user.ID == adminUser.ID {
			role = "owner"
		}

		membership := models.LeagueMembership{
			LeagueID: defaultLeague.ID,
			UserID:   user.ID,
			Role:     role,
			JoinedAt: time.Now(),
		}

		if err := db.Create(&membership).Error; err != nil {
			log.Printf("Warning: failed to create membership for user %s: %v", user.Username, err)
		}
	}

	log.Printf("Backfilled %d users as members of default league", len(allUsers))
	log.Println("Default league backfill completed successfully")

	return nil
}

// SeedMode determines what data to seed
type SeedMode string

const (
	SeedModeMinimal SeedMode = "minimal" // Only teams and superuser
	SeedModeFull    SeedMode = "full"    // Teams, superuser, league, season, weeks, games
)

// SeedData adds initial data for testing/development
// mode: "minimal" = teams + superuser only, "full" = teams + superuser + league + season
func SeedData(db *gorm.DB, mode SeedMode) error {
	log.Printf("Seeding database with initial data (mode: %s)...", mode)

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

	// Seed admin user (and optionally league based on mode)
	if err := seedAdminUser(db, mode); err != nil {
		return fmt.Errorf("failed to seed admin user: %w", err)
	}

	// Only seed development data in full mode
	if mode == SeedModeFull {
		if err := seedDevelopmentData(db); err != nil {
			return fmt.Errorf("failed to seed development data: %w", err)
		}
	}

	log.Println("Database seeded successfully")
	return nil
}

// seedAdminUser creates a default admin user and optionally their league for local development
func seedAdminUser(db *gorm.DB, mode SeedMode) error {
	// Hash the default password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Create admin user
	admin := models.User{
		Username:      "admin",
		Email:         "admin@example.com",
		PasswordHash:  string(hashedPassword),
		DisplayName:   "Admin User",
		IsAdmin:       true,
		IsGlobalAdmin: true, // NEW: Make first admin a global admin
	}

	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	log.Println("Created admin user (email: admin@example.com, password: admin123)")

	// Only create league in full mode
	if mode == SeedModeFull {
		// Create a default league owned by admin
		league := models.League{
			Name:        "Dev League",
			Code:        "DEV-2025",
			Description: "Development/Testing League",
			OwnerID:     admin.ID,
			IsPublic:    true,
			IsActive:    true,
		}

		if err := db.Create(&league).Error; err != nil {
			return fmt.Errorf("failed to create league: %w", err)
		}

		log.Printf("Created league: %s (Code: %s)", league.Name, league.Code)

		// Create admin's membership in their league
		membership := models.LeagueMembership{
			LeagueID: league.ID,
			UserID:   admin.ID,
			Role:     "owner",
			JoinedAt: time.Now(),
		}

		if err := db.Create(&membership).Error; err != nil {
			return fmt.Errorf("failed to create admin membership: %w", err)
		}
	} else {
		log.Println("Minimal mode: skipping league creation")
	}

	return nil
}

// seedDevelopmentData creates a sample season with weeks and games for local development
func seedDevelopmentData(db *gorm.DB) error {
	// Get the dev league
	var league models.League
	if err := db.Where("code = ?", "DEV-2025").First(&league).Error; err != nil {
		return fmt.Errorf("failed to find dev league: %w", err)
	}

	// Create current season (associated with the dev league)
	season := models.Season{
		LeagueID: league.ID, // NEW: Associate season with league
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
			Status:     "creating", // New weeks start in creating status
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
