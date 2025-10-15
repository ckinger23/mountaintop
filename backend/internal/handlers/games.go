package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ckinger23/mountaintop/internal/database"
	"github.com/ckinger23/mountaintop/internal/models"
	"github.com/go-chi/chi/v5"
)

// GetGames returns all games for a specific week
func GetGames(w http.ResponseWriter, r *http.Request) {
	weekID := r.URL.Query().Get("week_id")

	var games []models.Game
	query := database.DB.Preload("HomeTeam").Preload("AwayTeam").Preload("Week")

	if weekID != "" {
		query = query.Where("week_id = ?", weekID)
	}

	if err := query.Find(&games).Error; err != nil {
		http.Error(w, "Error fetching games", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}

// GetGame returns a single game by ID
func GetGame(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "id")

	var game models.Game
	if err := database.DB.Preload("HomeTeam").Preload("AwayTeam").Preload("Week").First(&game, gameID).Error; err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game)
}

type CreateGameRequest struct {
	WeekID     uint    `json:"week_id"`
	HomeTeamID uint    `json:"home_team_id"`
	AwayTeamID uint    `json:"away_team_id"`
	GameTime   string  `json:"game_time"` // ISO 8601 format
	HomeSpread float64 `json:"home_spread"`
}

// CreateGame creates a new game (admin only)
func CreateGame(w http.ResponseWriter, r *http.Request) {
	var req CreateGameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Parse game time
	// gameTime, err := time.Parse(time.RFC3339, req.GameTime)
	// if err != nil {
	// 	http.Error(w, "Invalid game_time format", http.StatusBadRequest)
	// 	return
	// }

	game := models.Game{
		WeekID:     req.WeekID,
		HomeTeamID: req.HomeTeamID,
		AwayTeamID: req.AwayTeamID,
		// GameTime:   gameTime,
		HomeSpread: req.HomeSpread,
	}

	if err := database.DB.Create(&game).Error; err != nil {
		http.Error(w, "Error creating game", http.StatusInternalServerError)
		return
	}

	// Load relationships
	database.DB.Preload("HomeTeam").Preload("AwayTeam").Preload("Week").First(&game, game.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(game)
}

type UpdateGameResultRequest struct {
	HomeScore int  `json:"home_score"`
	AwayScore int  `json:"away_score"`
	IsFinal   bool `json:"is_final"`
}

// UpdateGameResult updates the score and determines winner (admin only)
func UpdateGameResult(w http.ResponseWriter, r *http.Request) {
	gameID := chi.URLParam(r, "id")

	var req UpdateGameResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var game models.Game
	if err := database.DB.First(&game, gameID).Error; err != nil {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	}

	// Update game result
	game.HomeScore = &req.HomeScore
	game.AwayScore = &req.AwayScore
	game.IsFinal = req.IsFinal

	// Determine winner
	if req.HomeScore > req.AwayScore {
		game.WinnerTeamID = &game.HomeTeamID
	} else if req.AwayScore > req.HomeScore {
		game.WinnerTeamID = &game.AwayTeamID
	}
	// If tied, WinnerTeamID remains nil

	if err := database.DB.Save(&game).Error; err != nil {
		http.Error(w, "Error updating game", http.StatusInternalServerError)
		return
	}

	// If game is final, calculate pick results
	if game.IsFinal {
		if err := calculatePickResults(game.ID); err != nil {
			http.Error(w, "Error calculating pick results", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(game)
}

// calculatePickResults updates all picks for a game once it's final
func calculatePickResults(gameID uint) error {
	var game models.Game
	if err := database.DB.First(&game, gameID).Error; err != nil {
		return err
	}

	var picks []models.Pick
	if err := database.DB.Where("game_id = ?", gameID).Find(&picks).Error; err != nil {
		return err
	}

	for _, pick := range picks {
		isCorrect := pick.PickedTeamID == *game.WinnerTeamID
		pick.IsCorrect = &isCorrect

		// Simple scoring: 1 point for correct pick
		if isCorrect {
			pick.PointsEarned = 1
		} else {
			pick.PointsEarned = 0
		}

		database.DB.Save(&pick)
	}

	return nil
}

// GetWeeks returns all weeks, optionally filtered by season
func GetWeeks(w http.ResponseWriter, r *http.Request) {
	seasonID := r.URL.Query().Get("season_id")

	var weeks []models.Week
	query := database.DB.Preload("Season")

	if seasonID != "" {
		query = query.Where("season_id = ?", seasonID)
	}

	if err := query.Order("week_number ASC").Find(&weeks).Error; err != nil {
		http.Error(w, "Error fetching weeks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(weeks)
}

// GetCurrentWeek returns the current active week
func GetCurrentWeek(w http.ResponseWriter, r *http.Request) {
	var season models.Season
	if err := database.DB.Where("is_active = ?", true).First(&season).Error; err != nil {
		http.Error(w, "No active season found", http.StatusNotFound)
		return
	}

	var week models.Week
	// This is simplified - you'd want logic to determine current week based on dates
	if err := database.DB.Where("season_id = ?", season.ID).Order("week_number DESC").First(&week).Error; err != nil {
		http.Error(w, "No weeks found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(week)
}

// GetTeams returns all teams
func GetTeams(w http.ResponseWriter, r *http.Request) {
	var teams []models.Team
	if err := database.DB.Order("name ASC").Find(&teams).Error; err != nil {
		http.Error(w, "Error fetching teams", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(teams)
}

// GetSeasons returns all seasons
func GetSeasons(w http.ResponseWriter, r *http.Request) {
	var seasons []models.Season
	if err := database.DB.Order("year DESC").Find(&seasons).Error; err != nil {
		http.Error(w, "Error fetching seasons", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(seasons)
}

// GetLeaderboard returns the current standings
func GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	seasonIDStr := r.URL.Query().Get("season_id")

	var leaderboard []models.LeaderboardEntry

	query := `
		SELECT 
			u.id as user_id,
			u.username,
			u.display_name,
			COALESCE(SUM(p.points_earned), 0) as total_points,
			COALESCE(SUM(CASE WHEN p.is_correct = 1 THEN 1 ELSE 0 END), 0) as correct_picks,
			COUNT(p.id) as total_picks,
			COALESCE(CAST(SUM(CASE WHEN p.is_correct = 1 THEN 1 ELSE 0 END) AS FLOAT) / NULLIF(COUNT(p.id), 0), 0) as win_pct
		FROM users u
		LEFT JOIN picks p ON u.id = p.user_id
		LEFT JOIN games g ON p.game_id = g.id
		LEFT JOIN weeks w ON g.week_id = w.id
	`

	if seasonIDStr != "" {
		seasonID, _ := strconv.Atoi(seasonIDStr)
		query += ` WHERE w.season_id = ?`
		database.DB.Raw(query+` GROUP BY u.id ORDER BY total_points DESC`, seasonID).Scan(&leaderboard)
	} else {
		database.DB.Raw(query + ` GROUP BY u.id ORDER BY total_points DESC`).Scan(&leaderboard)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(leaderboard)
}
