package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ckinger23/mountaintop/internal/app"
	"github.com/ckinger23/mountaintop/internal/middleware"
	"github.com/ckinger23/mountaintop/internal/models"
	"github.com/go-chi/chi/v5"
)

type SubmitPickRequest struct {
	GameID       uint `json:"game_id"`
	PickedTeamID uint `json:"picked_team_id"`
	Confidence   int  `json:"confidence"`
}

// SubmitPick returns a handler for creating or updating a user's pick for a game
func SubmitPick(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.GetUserFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req SubmitPickRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Get the game to check lock time
		var game models.Game
		if err := a.DB.Preload("Week").First(&game, req.GameID).Error; err != nil {
			http.Error(w, "Game not found", http.StatusNotFound)
			return
		}

		// Check if picks are still allowed (before lock time)
		if time.Now().After(game.Week.LockTime) {
			http.Error(w, "Picks are locked for this week", http.StatusForbidden)
			return
		}

		// Check if game has already started or is final
		if game.IsFinal || time.Now().After(game.GameTime) {
			http.Error(w, "Cannot pick a game that has already started", http.StatusForbidden)
			return
		}

		// Validate that picked team is in the game
		if req.PickedTeamID != game.HomeTeamID && req.PickedTeamID != game.AwayTeamID {
			http.Error(w, "Invalid team selection for this game", http.StatusBadRequest)
			return
		}

		// Check if pick already exists
		var existingPick models.Pick
		err := a.DB.Where("user_id = ? AND game_id = ?", claims.UserID, req.GameID).First(&existingPick).Error

		if err == nil {
			// Update existing pick
			existingPick.PickedTeamID = req.PickedTeamID
			existingPick.Confidence = req.Confidence

			if err := a.DB.Save(&existingPick).Error; err != nil {
				http.Error(w, "Error updating pick", http.StatusInternalServerError)
				return
			}

			a.DB.Preload("Game").Preload("PickedTeam").First(&existingPick, existingPick.ID)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(existingPick)
			return
		}

		// Create new pick
		pick := models.Pick{
			UserID:       claims.UserID,
			GameID:       req.GameID,
			PickedTeamID: req.PickedTeamID,
			Confidence:   req.Confidence,
		}

		if err := a.DB.Create(&pick).Error; err != nil {
			http.Error(w, "Error creating pick", http.StatusInternalServerError)
			return
		}

		// Load relationships
		a.DB.Preload("Game").Preload("PickedTeam").First(&pick, pick.ID)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(pick)
	}
}

// GetMyPicks returns a handler for fetching all picks for the authenticated user
func GetMyPicks(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.GetUserFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		weekID := r.URL.Query().Get("week_id")

		var picks []models.Pick
		query := a.DB.Where("user_id = ?", claims.UserID).
			Preload("Game.HomeTeam").
			Preload("Game.AwayTeam").
			Preload("Game.Week").
			Preload("PickedTeam")

		if weekID != "" {
			query = query.Joins("JOIN games ON picks.game_id = games.id").
				Where("games.week_id = ?", weekID)
		}

		if err := query.Find(&picks).Error; err != nil {
			http.Error(w, "Error fetching picks", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(picks)
	}
}

// GetPicksForUser returns a handler for fetching all picks for a specific user (viewable by anyone)
func GetPicksForUser(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userId")
		weekID := r.URL.Query().Get("week_id")

		var picks []models.Pick
		query := a.DB.Where("user_id = ?", userID).
			Preload("Game.HomeTeam").
			Preload("Game.AwayTeam").
			Preload("Game.Week").
			Preload("PickedTeam").
			Preload("User")

		if weekID != "" {
			query = query.Joins("JOIN games ON picks.game_id = games.id").
				Where("games.week_id = ?", weekID)
		}

		if err := query.Find(&picks).Error; err != nil {
			http.Error(w, "Error fetching picks", http.StatusInternalServerError)
			return
		}

		// Only show picks if games have started or are final
		var visiblePicks []models.Pick
		for _, pick := range picks {
			if pick.Game.IsFinal || time.Now().After(pick.Game.GameTime) {
				visiblePicks = append(visiblePicks, pick)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(visiblePicks)
	}
}

// GetAllPicksForWeek returns a handler for fetching all users' picks for a week (admin or after lock)
func GetAllPicksForWeek(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		weekID := chi.URLParam(r, "weekId")

		// Get the week to check lock time
		var week models.Week
		if err := a.DB.First(&week, weekID).Error; err != nil {
			http.Error(w, "Week not found", http.StatusNotFound)
			return
		}

		// Check if week is locked
		if time.Now().Before(week.LockTime) {
			// Only admins can view before lock time
			claims, ok := middleware.GetUserFromContext(r)
			if !ok || !claims.IsAdmin {
				http.Error(w, "Picks not yet visible", http.StatusForbidden)
				return
			}
		}

		var picks []models.Pick
		if err := a.DB.
			Joins("JOIN games ON picks.game_id = games.id").
			Where("games.week_id = ?", weekID).
			Preload("User").
			Preload("Game.HomeTeam").
			Preload("Game.AwayTeam").
			Preload("PickedTeam").
			Find(&picks).Error; err != nil {
			http.Error(w, "Error fetching picks", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(picks)
	}
}

// GetPickStats returns a handler for fetching statistics about a user's picks
func GetPickStats(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := chi.URLParam(r, "userId")

		type Stats struct {
			TotalPicks     int64   `json:"total_picks"`
			CorrectPicks   int64   `json:"correct_picks"`
			IncorrectPicks int64   `json:"incorrect_picks"`
			WinPercentage  float64 `json:"win_percentage"`
			TotalPoints    int     `json:"total_points"`
		}

		var stats Stats

		a.DB.Model(&models.Pick{}).
			Where("user_id = ? AND is_correct IS NOT NULL", userID).
			Count(&stats.TotalPicks)

		a.DB.Model(&models.Pick{}).
			Where("user_id = ? AND is_correct = ?", userID, true).
			Count(&stats.CorrectPicks)

		stats.IncorrectPicks = stats.TotalPicks - stats.CorrectPicks

		if stats.TotalPicks > 0 {
			stats.WinPercentage = float64(stats.CorrectPicks) / float64(stats.TotalPicks) * 100
		}

		// Scan maps to any struct, not just GORM model
		// ignores GORM tags, hooks, associations
		// used for custom queries, aggregations, joins
		// Can't use Preload
		a.DB.Model(&models.Pick{}).
			Where("user_id = ?", userID).
			Select("COALESCE(SUM(points_earned), 0)").
			Scan(&stats.TotalPoints)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(stats)
	}
}
