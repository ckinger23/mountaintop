package models

import (
	"time"
)

// User represents a league user
// dynamodbav == dynamodb attribute value
type User struct {
	ID        string    `dynamodbav:"id"`
	Username  string    `dynamodbav:"username"`
	Email     string    `dynamodbav:"email"`
	Role      string    `dynamodbav:"role"` // admin/player
	CreatedAt time.Time `dynamodbav:"created_at"`
	UpdatedAt time.Time `dynamodbav:"updated_at"`
}

// Conference represents a team conference (e.g. NFC, AFC)
type Conference struct {
	ID        string    `dynamodbav:"id"`
	Name      string    `dynamodbav:"name"`
	CreatedAt time.Time `dynamodbav:"created_at"`
	UpdatedAt time.Time `dynamodbav:"updated_at"`
}

// Team represents a football team
type Team struct {
	ID           string    `dynamodbav:"id"`
	Name         string    `dynamodbav:"name"`
	ConferenceID string    `dynamodbav:"conference_id"`
	Wins         int       `dynamodbav:"wins"`
	Losses       int       `dynamodbav:"losses"`
	CreatedAt    time.Time `dynamodbav:"created_at"`
	UpdatedAt    time.Time `dynamodbav:"updated_at"`
}

// League represents a picking league
type League struct {
	ID        string    `dynamodbav:"id"`
	Name      string    `dynamodbav:"name"`
	AdminID   string    `dynamodbav:"admin_id"`
	CreatedAt time.Time `dynamodbav:"created_at"`
	UpdatedAt time.Time `dynamodbav:"updated_at"`
}

// Game represents a football game
type Game struct {
	ID         string    `dynamodbav:"id"`
	LeagueID   string    `dynamodbav:"league_id"`
	Week       int       `dynamodbav:"week"`
	HomeTeamID string    `dynamodbav:"home_team_id"`
	AwayTeamID string    `dynamodbav:"away_team_id"`
	GameDate   time.Time `dynamodbav:"game_date"`
	Status     string    `dynamodbav:"status"` // pending/in_progress/completed
	Winner     string    `dynamodbav:"winner"` // home/away
	CreatedAt  time.Time `dynamodbav:"created_at"`
	UpdatedAt  time.Time `dynamodbav:"updated_at"`
}

// Pick represents a user's game prediction
type Pick struct {
	ID        string    `dynamodbav:"id"`
	UserID    string    `dynamodbav:"user_id"`
	GameID    string    `dynamodbav:"game_id"`
	Week      int       `dynamodbav:"week"`
	Pick      string    `dynamodbav:"pick"`   // home/away
	Status    string    `dynamodbav:"status"` // pending/correct/incorrect
	Points    int       `dynamodbav:"points"`
	CreatedAt time.Time `dynamodbav:"created_at"`
	UpdatedAt time.Time `dynamodbav:"updated_at"`
}

// LeaderboardEntry represents a user's leaderboard position
type LeaderboardEntry struct {
	UserID    string `dynamodbav:"user_id"`
	Username  string `dynamodbav:"username"`
	Points    int    `dynamodbav:"points"`
	Correct   int    `dynamodbav:"correct"`
	Incorrect int    `dynamodbav:"incorrect"`
	Week      int    `dynamodbav:"week"`
}
