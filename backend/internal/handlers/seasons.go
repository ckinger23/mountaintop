package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ckinger23/mountaintop/internal/app"
	"github.com/ckinger23/mountaintop/internal/models"
	"github.com/ckinger23/mountaintop/internal/validation"
)

type CreateSeasonRequest struct {
	LeagueID uint   `json:"league_id"` // NEW: Which league this season belongs to
	Year     int    `json:"year"`
	Name     string `json:"name"` // e.g., "2024 Regular Season"
	IsActive bool   `json:"is_active"`
}

func CreateSeason(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateSeasonRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			validation.RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON", nil)
			return
		}

		// Validate season data
		if valErr := validation.ValidateCreateSeason(req.Year, req.Name); valErr != nil {
			validation.RespondWithValidationError(w, valErr)
			return
		}

		// Check if season with this year already exists for this league
		var existingSeason models.Season
		if err := a.DB.Where("league_id = ? AND year = ?", req.LeagueID, req.Year).First(&existingSeason).Error; err == nil {
			validation.RespondWithError(w, http.StatusConflict, "Season already exists", "SEASON_EXISTS", map[string]string{
				"year": "A season for this year already exists for this league",
			})
			return
		}

		season := models.Season{
			LeagueID: req.LeagueID,
			Year:     req.Year,
			Name:     req.Name,
			IsActive: req.IsActive,
		}

		if err := a.DB.Create(&season).Error; err != nil {
			validation.RespondWithError(w, http.StatusInternalServerError, "Error creating season", "DATABASE_ERROR", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(season)
	}
}
