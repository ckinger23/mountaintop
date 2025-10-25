package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ckinger23/mountaintop/internal/app"
	"github.com/ckinger23/mountaintop/internal/models"
)

type CreateSeasonRequest struct {
	Year     int    `json:"year"`
	Name     string `json:"name"` // e.g., "2024 Regular Season"
	IsActive bool   `json:"is_active"`
}

func CreateSeason(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateSeasonRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		season := models.Season{
			Year:     req.Year,
			Name:     req.Name,
			IsActive: req.IsActive,
			// Do not set weeks
		}

		if err := a.DB.Create(&season).Error; err != nil {
			http.Error(w, "Error creating season", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(season)

	}
}
