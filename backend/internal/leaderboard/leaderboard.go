package leaderboard

import (
	"github.com/ckinger23/mountaintop/internal/models"
	"gorm.io/gorm"
)

// Query builds and executes a leaderboard query with optional season and league filters
type Query struct {
	db       *gorm.DB
	seasonID *uint
	leagueID *uint
}

// NewQuery creates a new leaderboard query builder
func NewQuery(db *gorm.DB) *Query {
	return &Query{db: db}
}

// ForSeason filters the leaderboard to a specific season
func (q *Query) ForSeason(seasonID uint) *Query {
	q.seasonID = &seasonID
	return q
}

// ForLeague filters the leaderboard to a specific league
func (q *Query) ForLeague(leagueID uint) *Query {
	q.leagueID = &leagueID
	return q
}

// Execute runs the leaderboard query and returns results ordered by total points
func (q *Query) Execute() ([]models.LeaderboardEntry, error) {
	var results []models.LeaderboardEntry

	// Start with base query selecting from users
	// Note: We count total correct picks (spread + over/under), but total_picks is number of games
	// Each game can contribute 0-2 points (1 for spread, 1 for over/under)
	query := q.db.Table("users u").
		Select(`
			u.id as user_id,
			u.username,
			u.display_name,
			COALESCE(SUM(p.points_earned), 0) as total_points,
			COALESCE(SUM(CASE WHEN p.spread_correct = 1 THEN 1 ELSE 0 END) + SUM(CASE WHEN p.over_under_correct = 1 THEN 1 ELSE 0 END), 0) as correct_picks,
			COUNT(p.id) as total_picks,
			COALESCE(CAST(SUM(p.points_earned) AS FLOAT) / NULLIF(COUNT(p.id) * 2, 0), 0) as win_pct
		`).
		Joins("LEFT JOIN picks p ON u.id = p.user_id").
		Joins("LEFT JOIN games g ON p.game_id = g.id").
		Joins("LEFT JOIN weeks w ON g.week_id = w.id")

	// Apply season filter if specified
	if q.seasonID != nil {
		query = query.Where("w.season_id = ? OR p.id IS NULL", *q.seasonID)
	}

	// Apply league filter if specified
	if q.leagueID != nil {
		query = query.Where("p.league_id = ? OR p.id IS NULL", *q.leagueID)
	}

	// Group by user and order by points
	query = query.Group("u.id").Order("total_points DESC")

	// Execute query
	if err := query.Scan(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}

// GetLeaderboard is a convenience function for the most common use case
func GetLeaderboard(db *gorm.DB, seasonID *uint, leagueID *uint) ([]models.LeaderboardEntry, error) {
	query := NewQuery(db)

	if seasonID != nil {
		query = query.ForSeason(*seasonID)
	}

	if leagueID != nil {
		query = query.ForLeague(*leagueID)
	}

	return query.Execute()
}
