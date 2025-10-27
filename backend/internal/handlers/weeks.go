package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/ckinger23/mountaintop/internal/app"
	"github.com/ckinger23/mountaintop/internal/models"
	"github.com/go-chi/chi/v5"
)

type WeekRequest struct {
	SeasonID   uint      `json:"season_id"`
	WeekNumber int       `json:"week_number"`
	Name       string    `json:"name"`      // e.g., "Week 8"
	LockTime   time.Time `json:"lock_time"` // When picks must be submitted by
}

func CreateWeek(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req WeekRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		week := models.Week{
			SeasonID:   req.SeasonID,
			WeekNumber: req.WeekNumber,
			Name:       req.Name,
			LockTime:   req.LockTime,
		}

		if err := a.DB.Create(&week).Error; err != nil {
			http.Error(w, "Error creating week", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(week)
	}
}

// UpdateWeek returns a handler for updating week details (admin only)
func UpdateWeek(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		weekID := chi.URLParam(r, "id")

		var req WeekRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		var week models.Week
		if err := a.DB.First(&week, weekID).Error; err != nil {
			http.Error(w, "Week not found", http.StatusNotFound)
			return
		}

		// Update week fields
		week.SeasonID = req.SeasonID
		week.WeekNumber = req.WeekNumber
		week.Name = req.Name
		week.LockTime = req.LockTime

		if err := a.DB.Save(&week).Error; err != nil {
			http.Error(w, "Error updating week", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(week)
	}
}
