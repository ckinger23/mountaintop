package models

import (
	"time"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	Username     string `gorm:"uniqueIndex;not null" json:"username"`
	Email        string `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash string `gorm:"not null" json:"-"`
	IsAdmin      bool   `gorm:"default:false" json:"is_admin"`
	DisplayName  string `json:"display_name"`
	
	// Relationships
	Picks []Pick `gorm:"foreignKey:UserID" json:"picks,omitempty"`
}

// Season represents a CFB season (e.g., 2024, 2025)
type Season struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	Year      int    `gorm:"uniqueIndex;not null" json:"year"`
	Name      string `json:"name"` // e.g., "2024 Regular Season"
	IsActive  bool   `gorm:"default:false" json:"is_active"`
	
	// Relationships
	Weeks []Week `gorm:"foreignKey:SeasonID" json:"weeks,omitempty"`
}

// Week represents a week in the season
type Week struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	SeasonID    uint      `gorm:"not null" json:"season_id"`
	WeekNumber  int       `gorm:"not null" json:"week_number"`
	Name        string    `json:"name"` // e.g., "Week 8"
	LockTime    time.Time `json:"lock_time"` // When picks must be submitted by
	
	// Relationships
	Season Season `gorm:"foreignKey:SeasonID" json:"season,omitempty"`
	Games  []Game `gorm:"foreignKey:WeekID" json:"games,omitempty"`
}

// Team represents a college football team
type Team struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	Name         string `gorm:"uniqueIndex;not null" json:"name"`
	Abbreviation string `gorm:"uniqueIndex" json:"abbreviation"`
	LogoURL      string `json:"logo_url"`
	Conference   string `json:"conference"`
}

// Game represents a single game
type Game struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	WeekID      uint       `gorm:"not null" json:"week_id"`
	HomeTeamID  uint       `gorm:"not null" json:"home_team_id"`
	AwayTeamID  uint       `gorm:"not null" json:"away_team_id"`
	GameTime    time.Time  `json:"game_time"`
	HomeSpread  float64    `json:"home_spread"` // Negative means home team favored
	Total       float64    `json:"total"`       // Over/under line for combined score

	// Game Results (null until game is final)
	IsFinal      bool  `gorm:"default:false" json:"is_final"`
	HomeScore    *int  `json:"home_score"`
	AwayScore    *int  `json:"away_score"`
	WinnerTeamID *uint `json:"winner_team_id"` // null for tie
	
	// Relationships
	Week     Week   `gorm:"foreignKey:WeekID" json:"week,omitempty"`
	HomeTeam Team   `gorm:"foreignKey:HomeTeamID" json:"home_team,omitempty"`
	AwayTeam Team   `gorm:"foreignKey:AwayTeamID" json:"away_team,omitempty"`
	Picks    []Pick `gorm:"foreignKey:GameID" json:"picks,omitempty"`
}

// Pick represents a user's pick for a game
type Pick struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	UserID         uint   `gorm:"not null;uniqueIndex:idx_user_game" json:"user_id"`
	GameID         uint   `gorm:"not null;uniqueIndex:idx_user_game" json:"game_id"`
	PickedTeamID   uint   `gorm:"not null" json:"picked_team_id"`
	PickedOverUnder string `json:"picked_over_under"` // "over" or "under"
	Confidence     int    `json:"confidence"`         // Optional: for confidence pools

	// Scoring (one point for spread pick, one point for over/under pick)
	SpreadCorrect     *bool `json:"spread_correct"`      // null until game is final
	OverUnderCorrect  *bool `json:"over_under_correct"`  // null until game is final
	PointsEarned      int   `gorm:"default:0" json:"points_earned"`
	
	// Relationships
	User        User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Game        Game `gorm:"foreignKey:GameID" json:"game,omitempty"`
	PickedTeam  Team `gorm:"foreignKey:PickedTeamID" json:"picked_team,omitempty"`
}

// Leaderboard is a view/calculated model for displaying standings
type LeaderboardEntry struct {
	UserID      uint   `json:"user_id"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	TotalPoints int    `json:"total_points"`
	CorrectPicks int   `json:"correct_picks"`
	TotalPicks  int    `json:"total_picks"`
	WinPct      float64 `json:"win_pct"`
}
