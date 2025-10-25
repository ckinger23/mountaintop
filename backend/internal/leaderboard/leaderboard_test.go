package leaderboard

import (
	"testing"
	"time"

	"github.com/ckinger23/mountaintop/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Auto-migrate all models
	err = db.AutoMigrate(
		&models.User{},
		&models.Season{},
		&models.Week{},
		&models.Team{},
		&models.Game{},
		&models.Pick{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// seedTestData creates sample data for testing
func seedTestData(t *testing.T, db *gorm.DB) {
	// Create users
	alice := models.User{Username: "alice", Email: "alice@example.com", DisplayName: "Alice", PasswordHash: "hash1"}
	bob := models.User{Username: "bob", Email: "bob@example.com", DisplayName: "Bob", PasswordHash: "hash2"}
	charlie := models.User{Username: "charlie", Email: "charlie@example.com", DisplayName: "Charlie", PasswordHash: "hash3"}

	if err := db.Create(&alice).Error; err != nil {
		t.Fatalf("Failed to create alice: %v", err)
	}
	if err := db.Create(&bob).Error; err != nil {
		t.Fatalf("Failed to create bob: %v", err)
	}
	if err := db.Create(&charlie).Error; err != nil {
		t.Fatalf("Failed to create charlie: %v", err)
	}

	// Create seasons
	season2024 := models.Season{Year: 2024, Name: "2024 Season", IsActive: false}
	season2025 := models.Season{Year: 2025, Name: "2025 Season", IsActive: true}
	db.Create(&season2024)
	db.Create(&season2025)

	// Create weeks
	week1_2024 := models.Week{SeasonID: season2024.ID, WeekNumber: 1, Name: "Week 1", LockTime: time.Now()}
	week1_2025 := models.Week{SeasonID: season2025.ID, WeekNumber: 1, Name: "Week 1", LockTime: time.Now()}
	db.Create(&week1_2024)
	db.Create(&week1_2025)

	// Create teams
	teams := []models.Team{
		{Name: "Team A", Abbreviation: "TEA"},
		{Name: "Team B", Abbreviation: "TEB"},
		{Name: "Team C", Abbreviation: "TEC"},
		{Name: "Team D", Abbreviation: "TED"},
	}
	for _, team := range teams {
		db.Create(&team)
	}

	// Create games for 2024
	homeScore1 := 28
	awayScore1 := 21
	game1_2024 := models.Game{
		WeekID:       week1_2024.ID,
		HomeTeamID:   teams[0].ID,
		AwayTeamID:   teams[1].ID,
		GameTime:     time.Now(),
		HomeSpread:   -7.0,
		IsFinal:      true,
		HomeScore:    &homeScore1,
		AwayScore:    &awayScore1,
		WinnerTeamID: &teams[0].ID,
	}
	db.Create(&game1_2024)

	homeScore2 := 14
	awayScore2 := 35
	game2_2024 := models.Game{
		WeekID:       week1_2024.ID,
		HomeTeamID:   teams[2].ID,
		AwayTeamID:   teams[3].ID,
		GameTime:     time.Now(),
		HomeSpread:   3.5,
		IsFinal:      true,
		HomeScore:    &homeScore2,
		AwayScore:    &awayScore2,
		WinnerTeamID: &teams[3].ID,
	}
	db.Create(&game2_2024)

	// Create games for 2025
	homeScore3 := 42
	awayScore3 := 38
	game1_2025 := models.Game{
		WeekID:       week1_2025.ID,
		HomeTeamID:   teams[0].ID,
		AwayTeamID:   teams[1].ID,
		GameTime:     time.Now(),
		HomeSpread:   -3.0,
		IsFinal:      true,
		HomeScore:    &homeScore3,
		AwayScore:    &awayScore3,
		WinnerTeamID: &teams[0].ID,
	}
	db.Create(&game1_2025)

	// Create picks for 2024
	// Alice: 2 correct picks
	isCorrect1 := true
	db.Create(&models.Pick{
		UserID:       alice.ID,
		GameID:       game1_2024.ID,
		PickedTeamID: teams[0].ID,
		IsCorrect:    &isCorrect1,
		PointsEarned: 1,
	})
	db.Create(&models.Pick{
		UserID:       alice.ID,
		GameID:       game2_2024.ID,
		PickedTeamID: teams[3].ID,
		IsCorrect:    &isCorrect1,
		PointsEarned: 1,
	})

	// Bob: 1 correct, 1 incorrect
	isCorrect2 := true
	isIncorrect := false
	db.Create(&models.Pick{
		UserID:       bob.ID,
		GameID:       game1_2024.ID,
		PickedTeamID: teams[0].ID,
		IsCorrect:    &isCorrect2,
		PointsEarned: 1,
	})
	db.Create(&models.Pick{
		UserID:       bob.ID,
		GameID:       game2_2024.ID,
		PickedTeamID: teams[2].ID,
		IsCorrect:    &isIncorrect,
		PointsEarned: 0,
	})

	// Create picks for 2025
	// Alice: 1 correct pick
	isCorrect3 := true
	db.Create(&models.Pick{
		UserID:       alice.ID,
		GameID:       game1_2025.ID,
		PickedTeamID: teams[0].ID,
		IsCorrect:    &isCorrect3,
		PointsEarned: 1,
	})

	// Charlie: 0 correct picks in 2025
	isIncorrect2 := false
	db.Create(&models.Pick{
		UserID:       charlie.ID,
		GameID:       game1_2025.ID,
		PickedTeamID: teams[1].ID,
		IsCorrect:    &isIncorrect2,
		PointsEarned: 0,
	})
}

func TestGetLeaderboard_AllSeasons(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)

	entries, err := GetLeaderboard(db, nil)

	assert.NoError(t, err)
	assert.Len(t, entries, 3, "Should return all 3 users")

	// Alice should be first (3 total points across both seasons)
	assert.Equal(t, "alice", entries[0].Username)
	assert.Equal(t, 3, entries[0].TotalPoints)
	assert.Equal(t, 3, entries[0].CorrectPicks)
	assert.Equal(t, 3, entries[0].TotalPicks)
	assert.InDelta(t, 1.0, entries[0].WinPct, 0.01)

	// Bob should be second (1 point)
	assert.Equal(t, "bob", entries[1].Username)
	assert.Equal(t, 1, entries[1].TotalPoints)
	assert.Equal(t, 1, entries[1].CorrectPicks)
	assert.Equal(t, 2, entries[1].TotalPicks)
	assert.InDelta(t, 0.5, entries[1].WinPct, 0.01)

	// Charlie should be third (0 points)
	assert.Equal(t, "charlie", entries[2].Username)
	assert.Equal(t, 0, entries[2].TotalPoints)
	assert.Equal(t, 0, entries[2].CorrectPicks)
	assert.Equal(t, 1, entries[2].TotalPicks)
	assert.InDelta(t, 0.0, entries[2].WinPct, 0.01)
}

func TestGetLeaderboard_FilteredBySeason(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)

	// Get 2024 season ID
	var season2024 models.Season
	db.Where("year = ?", 2024).First(&season2024)

	entries, err := GetLeaderboard(db, &season2024.ID)

	assert.NoError(t, err)
	assert.Len(t, entries, 2, "Should return only users with picks in 2024")

	// Alice should be first for 2024 (2 points)
	assert.Equal(t, "alice", entries[0].Username)
	assert.Equal(t, 2, entries[0].TotalPoints)
	assert.Equal(t, 2, entries[0].CorrectPicks)
	assert.Equal(t, 2, entries[0].TotalPicks)
	assert.InDelta(t, 1.0, entries[0].WinPct, 0.01)

	// Bob should be second (1 point)
	assert.Equal(t, "bob", entries[1].Username)
	assert.Equal(t, 1, entries[1].TotalPoints)
	assert.Equal(t, 1, entries[1].CorrectPicks)
	assert.Equal(t, 2, entries[1].TotalPicks)
	assert.InDelta(t, 0.5, entries[1].WinPct, 0.01)
}

func TestGetLeaderboard_2025Season(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)

	// Get 2025 season ID
	var season2025 models.Season
	db.Where("year = ?", 2025).First(&season2025)

	entries, err := GetLeaderboard(db, &season2025.ID)

	assert.NoError(t, err)
	assert.Len(t, entries, 2, "Should return only users with picks in 2025")

	// Alice should be first for 2025 (1 point)
	assert.Equal(t, "alice", entries[0].Username)
	assert.Equal(t, 1, entries[0].TotalPoints)
	assert.Equal(t, 1, entries[0].CorrectPicks)
	assert.Equal(t, 1, entries[0].TotalPicks)

	// Charlie should be second (0 points but has a pick)
	assert.Equal(t, "charlie", entries[1].Username)
	assert.Equal(t, 0, entries[1].TotalPoints)
	assert.Equal(t, 0, entries[1].CorrectPicks)
	assert.Equal(t, 1, entries[1].TotalPicks)
}

func TestQueryBuilder_Chaining(t *testing.T) {
	db := setupTestDB(t)
	seedTestData(t, db)

	var season2024 models.Season
	db.Where("year = ?", 2024).First(&season2024)

	// Test the query builder with method chaining
	query := NewQuery(db).ForSeason(season2024.ID)
	entries, err := query.Execute()

	assert.NoError(t, err)
	assert.Len(t, entries, 2, "Should return only users with picks in 2024")
	assert.Equal(t, "alice", entries[0].Username)
	assert.Equal(t, 2, entries[0].TotalPoints)
}

func TestGetLeaderboard_EmptyDatabase(t *testing.T) {
	db := setupTestDB(t)

	entries, err := GetLeaderboard(db, nil)

	assert.NoError(t, err)
	assert.Len(t, entries, 0, "Should return empty array for empty database")
}

func TestGetLeaderboard_UsersWithNoPicks(t *testing.T) {
	db := setupTestDB(t)

	// Create users but no picks
	users := []models.User{
		{Username: "alice", Email: "alice@example.com", DisplayName: "Alice", PasswordHash: "hash1"},
		{Username: "bob", Email: "bob@example.com", DisplayName: "Bob", PasswordHash: "hash2"},
	}
	for _, user := range users {
		db.Create(&user)
	}

	entries, err := GetLeaderboard(db, nil)

	assert.NoError(t, err)
	assert.Len(t, entries, 2)

	// All users should have 0 points
	for _, entry := range entries {
		assert.Equal(t, 0, entry.TotalPoints)
		assert.Equal(t, 0, entry.CorrectPicks)
		assert.Equal(t, 0, entry.TotalPicks)
		assert.Equal(t, 0.0, entry.WinPct)
	}
}
