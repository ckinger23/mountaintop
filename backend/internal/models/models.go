package models

import (
	"time"
	"gorm.io/gorm"
)

// League represents a picks league/pool
type League struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Name        string `gorm:"not null" json:"name"`                      // "Carter's CFB League"
	Code        string `gorm:"uniqueIndex;not null" json:"code"`          // "CFB-2025-XY7K" (auto-generated)
	Description string `json:"description"`                               // Optional
	OwnerID     uint   `gorm:"not null;uniqueIndex" json:"owner_id"`      // User who created it (uniqueIndex = can only own 1)
	IsPublic    bool   `gorm:"default:false" json:"is_public"`            // Public leagues show in browse
	IsActive    bool   `gorm:"default:true" json:"is_active"`

	// Relationships
	Owner       User                 `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Members     []LeagueMembership   `gorm:"foreignKey:LeagueID;constraint:OnDelete:CASCADE" json:"members,omitempty"`
	Seasons     []Season             `gorm:"foreignKey:LeagueID;constraint:OnDelete:CASCADE" json:"seasons,omitempty"`
}

// LeagueMembership represents a user's membership in a league
type LeagueMembership struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	LeagueID  uint      `gorm:"not null;uniqueIndex:idx_league_user" json:"league_id"` // Unique per league+user
	UserID    uint      `gorm:"not null;uniqueIndex:idx_league_user" json:"user_id"`
	Role      string    `gorm:"default:'member'" json:"role"` // "owner" or "member"
	JoinedAt  time.Time `gorm:"not null" json:"joined_at"`

	// Relationships
	League    League `gorm:"foreignKey:LeagueID" json:"league,omitempty"`
	User      User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// User represents a user in the system
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	Username      string `gorm:"uniqueIndex;not null" json:"username"`
	Email         string `gorm:"uniqueIndex;not null" json:"email"`
	PasswordHash  string `gorm:"not null" json:"-"`
	IsAdmin       bool   `gorm:"default:false" json:"is_admin"` // Legacy - kept for backward compatibility
	IsGlobalAdmin bool   `gorm:"default:false" json:"is_global_admin"` // NEW: Superuser (you)
	DisplayName   string `json:"display_name"`

	// Relationships
	Picks       []Pick               `gorm:"foreignKey:UserID" json:"picks,omitempty"`
	OwnedLeague *League              `gorm:"foreignKey:OwnerID" json:"owned_league,omitempty"` // NEW: Can own ONE league
	Memberships []LeagueMembership   `gorm:"foreignKey:UserID" json:"memberships,omitempty"`    // NEW: Can join MULTIPLE leagues
}

// Season represents a CFB season (e.g., 2024, 2025)
type Season struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	LeagueID  uint   `gorm:"not null;index" json:"league_id"` // NEW: Which league owns this season
	Year      int    `gorm:"not null" json:"year"`            // Removed uniqueIndex - multiple leagues can have same year
	Name      string `json:"name"`                            // e.g., "2024 Regular Season"
	IsActive  bool   `gorm:"default:false" json:"is_active"`

	// Relationships
	League    League `gorm:"foreignKey:LeagueID" json:"league,omitempty"` // NEW
	Weeks     []Week `gorm:"foreignKey:SeasonID" json:"weeks,omitempty"`
}

// Week represents a week in the season
type Week struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	SeasonID     uint       `gorm:"not null" json:"season_id"`
	WeekNumber   int        `gorm:"not null" json:"week_number"`
	Name         string     `json:"name"` // e.g., "Week 8"
	Status       string     `gorm:"default:'creating'" json:"status"` // creating, picking, scoring, finished
	PickDeadline *time.Time `json:"pick_deadline"` // When picks must be submitted by (set when transitioning to 'picking')

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

	LeagueID        uint   `gorm:"not null;index" json:"league_id"`                        // NEW: Denormalized for easier queries
	UserID          uint   `gorm:"not null;uniqueIndex:idx_league_user_game" json:"user_id"` // Updated index
	GameID          uint   `gorm:"not null;uniqueIndex:idx_league_user_game" json:"game_id"` // Updated index
	PickedTeamID    uint   `gorm:"not null" json:"picked_team_id"`
	PickedOverUnder string `json:"picked_over_under"` // "over" or "under"
	Confidence      int    `json:"confidence"`         // Optional: for confidence pools

	// Scoring (one point for spread pick, one point for over/under pick)
	SpreadCorrect     *bool `json:"spread_correct"`      // null until game is final
	OverUnderCorrect  *bool `json:"over_under_correct"`  // null until game is final
	PointsEarned      int   `gorm:"default:0" json:"points_earned"`

	// Relationships
	League      League `gorm:"foreignKey:LeagueID" json:"league,omitempty"`      // NEW
	User        User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Game        Game   `gorm:"foreignKey:GameID" json:"game,omitempty"`
	PickedTeam  Team   `gorm:"foreignKey:PickedTeamID" json:"picked_team,omitempty"`
}

// Leaderboard is a view/calculated model for displaying standings
type LeaderboardEntry struct {
	LeagueID     uint    `json:"league_id"`      // NEW: Which league this leaderboard is for
	UserID       uint    `json:"user_id"`
	Username     string  `json:"username"`
	DisplayName  string  `json:"display_name"`
	TotalPoints  int     `json:"total_points"`
	CorrectPicks int     `json:"correct_picks"`
	TotalPicks   int     `json:"total_picks"`
	WinPct       float64 `json:"win_pct"`
}
