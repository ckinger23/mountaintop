package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/ckinger23/mountaintop/internal/app"
	"github.com/ckinger23/mountaintop/internal/leaderboard"
	"github.com/ckinger23/mountaintop/internal/models"
	"github.com/ckinger23/mountaintop/internal/validation"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

// GetGames returns a handler for fetching all games for a specific week
func GetGames(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		weekID := r.URL.Query().Get("week_id")

		var games []models.Game
		// Preload() loads related data from other tables
		// only works with Find(), FIrst(), and Scan()
		// Solves n+1 query problem
		query := a.DB.Preload("HomeTeam").Preload("AwayTeam").Preload("Week")

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
}

// GetGame returns a handler for fetching a single game by ID
func GetGame(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameID := chi.URLParam(r, "id")

		var game models.Game
		if err := a.DB.Preload("HomeTeam").Preload("AwayTeam").Preload("Week").First(&game, gameID).Error; err != nil {
			http.Error(w, "Game not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(game)
	}
}

type CreateGameRequest struct {
	WeekID     uint    `json:"week_id"`
	HomeTeamID uint    `json:"home_team_id"`
	AwayTeamID uint    `json:"away_team_id"`
	GameTime   string  `json:"game_time"` // ISO 8601 format
	HomeSpread float64 `json:"home_spread"`
	Total      float64 `json:"total"`
}

// CreateGame returns a handler for creating a new game (admin only)
func CreateGame(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateGameRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			validation.RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON", nil)
			return
		}

		// Validate and parse game time
		gameTime, valErr := validation.ValidateGameTime(req.GameTime)
		if valErr != nil {
			validation.RespondWithValidationError(w, valErr)
			return
		}

		// Validate game data
		if valErr := validation.ValidateCreateGame(req.WeekID, req.HomeTeamID, req.AwayTeamID, gameTime, req.HomeSpread, req.Total); valErr != nil {
			validation.RespondWithValidationError(w, valErr)
			return
		}

		// Check that week exists
		var week models.Week
		if err := a.DB.First(&week, req.WeekID).Error; err != nil {
			validation.RespondWithError(w, http.StatusBadRequest, "Week not found", "WEEK_NOT_FOUND", map[string]string{
				"week_id": "The specified week does not exist",
			})
			return
		}

		// Check that teams exist
		var homeTeam, awayTeam models.Team
		if err := a.DB.First(&homeTeam, req.HomeTeamID).Error; err != nil {
			validation.RespondWithError(w, http.StatusBadRequest, "Home team not found", "TEAM_NOT_FOUND", map[string]string{
				"home_team_id": "The specified home team does not exist",
			})
			return
		}
		if err := a.DB.First(&awayTeam, req.AwayTeamID).Error; err != nil {
			validation.RespondWithError(w, http.StatusBadRequest, "Away team not found", "TEAM_NOT_FOUND", map[string]string{
				"away_team_id": "The specified away team does not exist",
			})
			return
		}

		game := models.Game{
			WeekID:     req.WeekID,
			HomeTeamID: req.HomeTeamID,
			AwayTeamID: req.AwayTeamID,
			GameTime:   gameTime,
			HomeSpread: req.HomeSpread,
			Total:      req.Total,
		}

		if err := a.DB.Create(&game).Error; err != nil {
			validation.RespondWithError(w, http.StatusInternalServerError, "Error creating game", "DATABASE_ERROR", nil)
			return
		}

		// Load relationships
		a.DB.Preload("HomeTeam").Preload("AwayTeam").Preload("Week").First(&game, game.ID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(game)
	}
}

// UpdateGame returns a handler for updating game details (admin only)
func UpdateGame(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameID := chi.URLParam(r, "id")

		var game models.Game
		if err := a.DB.First(&game, gameID).Error; err != nil {
			validation.RespondWithError(w, http.StatusNotFound, "Game not found", "GAME_NOT_FOUND", nil)
			return
		}

		var req CreateGameRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			validation.RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON", nil)
			return
		}

		// Validate and parse game time
		gameTime, valErr := validation.ValidateGameTime(req.GameTime)
		if valErr != nil {
			validation.RespondWithValidationError(w, valErr)
			return
		}

		// Validate game data
		if valErr := validation.ValidateCreateGame(req.WeekID, req.HomeTeamID, req.AwayTeamID, gameTime, req.HomeSpread, req.Total); valErr != nil {
			validation.RespondWithValidationError(w, valErr)
			return
		}

		// Check that week exists
		var week models.Week
		if err := a.DB.First(&week, req.WeekID).Error; err != nil {
			validation.RespondWithError(w, http.StatusBadRequest, "Week not found", "WEEK_NOT_FOUND", map[string]string{
				"week_id": "The specified week does not exist",
			})
			return
		}

		// Check that teams exist
		var homeTeam, awayTeam models.Team
		if err := a.DB.First(&homeTeam, req.HomeTeamID).Error; err != nil {
			validation.RespondWithError(w, http.StatusBadRequest, "Home team not found", "TEAM_NOT_FOUND", map[string]string{
				"home_team_id": "The specified home team does not exist",
			})
			return
		}
		if err := a.DB.First(&awayTeam, req.AwayTeamID).Error; err != nil {
			validation.RespondWithError(w, http.StatusBadRequest, "Away team not found", "TEAM_NOT_FOUND", map[string]string{
				"away_team_id": "The specified away team does not exist",
			})
			return
		}

		// Update game fields
		game.WeekID = req.WeekID
		game.HomeTeamID = req.HomeTeamID
		game.AwayTeamID = req.AwayTeamID
		game.GameTime = gameTime
		game.HomeSpread = req.HomeSpread
		game.Total = req.Total

		if err := a.DB.Save(&game).Error; err != nil {
			validation.RespondWithError(w, http.StatusInternalServerError, "Error updating game", "DATABASE_ERROR", nil)
			return
		}

		// Load relationships
		a.DB.Preload("HomeTeam").Preload("AwayTeam").Preload("Week").First(&game, game.ID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(game)
	}
}

// DeleteGame returns a handler for deleting a game (admin only)
func DeleteGame(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameID := chi.URLParam(r, "id")

		var game models.Game
		if err := a.DB.First(&game, gameID).Error; err != nil {
			validation.RespondWithError(w, http.StatusNotFound, "Game not found", "GAME_NOT_FOUND", nil)
			return
		}

		// Don't allow deleting games that are final or have picks
		if game.IsFinal {
			validation.RespondWithError(w, http.StatusForbidden, "Cannot delete a final game", "GAME_IS_FINAL", map[string]string{
				"game_id": "This game is marked as final and cannot be deleted",
			})
			return
		}

		var pickCount int64
		a.DB.Model(&models.Pick{}).Where("game_id = ?", gameID).Count(&pickCount)
		if pickCount > 0 {
			validation.RespondWithError(w, http.StatusForbidden, "Cannot delete a game with existing picks", "HAS_PICKS", map[string]string{
				"game_id": "This game has user picks and cannot be deleted",
			})
			return
		}

		if err := a.DB.Delete(&game).Error; err != nil {
			validation.RespondWithError(w, http.StatusInternalServerError, "Error deleting game", "DATABASE_ERROR", nil)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

type UpdateGameResultRequest struct {
	HomeScore int  `json:"home_score"`
	AwayScore int  `json:"away_score"`
	IsFinal   bool `json:"is_final"`
}

// UpdateGameResult returns a handler for updating the score and determining winner (admin only)
func UpdateGameResult(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameID := chi.URLParam(r, "id")

		var req UpdateGameResultRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			validation.RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON", nil)
			return
		}

		// Validate scores
		homeScore := &req.HomeScore
		awayScore := &req.AwayScore
		if valErr := validation.ValidateUpdateGameResult(homeScore, awayScore, req.IsFinal); valErr != nil {
			validation.RespondWithValidationError(w, valErr)
			return
		}

		// Check game exists before starting transaction
		var game models.Game
		if err := a.DB.First(&game, gameID).Error; err != nil {
			validation.RespondWithError(w, http.StatusNotFound, "Game not found", "GAME_NOT_FOUND", nil)
			return
		}

		// Start transaction for updating game and calculating picks
		tx := a.DB.Begin()
		if tx.Error != nil {
			validation.RespondWithError(w, http.StatusInternalServerError, "Error starting transaction", "DATABASE_ERROR", nil)
			return
		}

		// Ensure rollback on panic or error
		defer func() {
			if r := recover(); r != nil {
				tx.Rollback()
				validation.RespondWithError(w, http.StatusInternalServerError, "Internal server error", "PANIC", nil)
			}
		}()

		// Update game result
		game.HomeScore = homeScore
		game.AwayScore = awayScore
		game.IsFinal = req.IsFinal

		// Determine winner
		if req.HomeScore > req.AwayScore {
			game.WinnerTeamID = &game.HomeTeamID
		} else if req.AwayScore > req.HomeScore {
			game.WinnerTeamID = &game.AwayTeamID
		} else {
			// If tied, WinnerTeamID remains nil
			game.WinnerTeamID = nil
		}

		if err := tx.Save(&game).Error; err != nil {
			tx.Rollback()
			validation.RespondWithError(w, http.StatusInternalServerError, "Error updating game", "DATABASE_ERROR", nil)
			return
		}

		// If game is final, calculate pick results within the same transaction
		if game.IsFinal {
			if err := calculatePickResults(tx, game.ID); err != nil {
				tx.Rollback()
				validation.RespondWithError(w, http.StatusInternalServerError, "Error calculating pick results", "SCORING_ERROR", nil)
				return
			}
		}

		// Commit transaction
		if err := tx.Commit().Error; err != nil {
			validation.RespondWithError(w, http.StatusInternalServerError, "Error committing transaction", "DATABASE_ERROR", nil)
			return
		}

		// Reload game with relationships after successful commit
		a.DB.Preload("HomeTeam").Preload("AwayTeam").Preload("Week").First(&game, game.ID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(game)
	}
}

// calculatePickResults updates all picks for a game once it's final
// This function should be called within a transaction
func calculatePickResults(tx *gorm.DB, gameID uint) error {
	var game models.Game
	if err := tx.First(&game, gameID).Error; err != nil {
		return err
	}

	// Verify game has scores and winner
	if game.HomeScore == nil || game.AwayScore == nil {
		return nil // Nothing to calculate without scores
	}

	var picks []models.Pick
	if err := tx.Where("game_id = ?", gameID).Find(&picks).Error; err != nil {
		return err
	}

	// If no picks, nothing to do
	if len(picks) == 0 {
		return nil
	}

	// Calculate actual total score
	actualTotal := float64(*game.HomeScore + *game.AwayScore)

	// Update all picks in a batch
	for i := range picks {
		// Check spread pick correctness
		spreadCorrect := false
		if game.WinnerTeamID != nil {
			spreadCorrect = picks[i].PickedTeamID == *game.WinnerTeamID
		}
		picks[i].SpreadCorrect = &spreadCorrect

		// Check over/under pick correctness
		var overUnderCorrect bool
		if picks[i].PickedOverUnder == "over" {
			overUnderCorrect = actualTotal > game.Total
		} else if picks[i].PickedOverUnder == "under" {
			overUnderCorrect = actualTotal < game.Total
		}
		picks[i].OverUnderCorrect = &overUnderCorrect

		// Scoring: 1 point for correct spread, 1 point for correct over/under (max 2 points per game)
		picks[i].PointsEarned = 0
		if spreadCorrect {
			picks[i].PointsEarned++
		}
		if overUnderCorrect {
			picks[i].PointsEarned++
		}

		// Save each pick within the transaction
		if err := tx.Save(&picks[i]).Error; err != nil {
			return err
		}
	}

	return nil
}

// GetWeeks returns a handler for fetching all weeks, optionally filtered by season
func GetWeeks(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		seasonID := r.URL.Query().Get("season_id")

		var weeks []models.Week
		query := a.DB.Preload("Season")

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
}

// GetCurrentWeek returns a handler for fetching the current active week
func GetCurrentWeek(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var season models.Season
		if err := a.DB.Where("is_active = ?", true).First(&season).Error; err != nil {
			http.Error(w, "No active season found", http.StatusNotFound)
			return
		}

		var week models.Week
		// This is simplified - you'd want logic to determine current week based on dates
		if err := a.DB.Where("season_id = ?", season.ID).Order("week_number DESC").First(&week).Error; err != nil {
			http.Error(w, "No weeks found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(week)
	}
}

// GetTeams returns a handler for fetching all teams
func GetTeams(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var teams []models.Team
		if err := a.DB.Order("name ASC").Find(&teams).Error; err != nil {
			http.Error(w, "Error fetching teams", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(teams)
	}
}

// GetSeasons returns a handler for fetching all seasons
func GetSeasons(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var seasons []models.Season
		if err := a.DB.Order("year DESC").Find(&seasons).Error; err != nil {
			http.Error(w, "Error fetching seasons", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(seasons)
	}
}

// GetLeaderboard returns a handler for fetching the current standings
func GetLeaderboard(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get season Id from query param
		seasonIDStr := r.URL.Query().Get("season_id")

		var seasonID *uint
		if seasonIDStr != "" {
			id, err := strconv.ParseUint(seasonIDStr, 10, 32)
			if err != nil {
				http.Error(w, "Invalid season_id parameter", http.StatusBadRequest)
				return
			}
			uid := uint(id)
			seasonID = &uid
		}

		entries, err := leaderboard.GetLeaderboard(a.DB, seasonID)
		if err != nil {
			http.Error(w, "Error fetching leaderboard", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(entries)
	}
}
