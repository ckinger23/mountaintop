package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ckinger23/mountaintop/internal/app"
	"github.com/ckinger23/mountaintop/internal/models"
	"github.com/ckinger23/mountaintop/internal/validation"
	"github.com/go-chi/chi/v5"
)

type WeekRequest struct {
	SeasonID   uint   `json:"season_id"`
	WeekNumber int    `json:"week_number"`
	Name       string `json:"name"`
}

func CreateWeek(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req WeekRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			validation.RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON", nil)
			return
		}

		// Validate week data (basic validation)
		details := make(map[string]string)
		if req.SeasonID == 0 {
			details["season_id"] = "Season ID is required"
		}
		if req.WeekNumber <= 0 {
			details["week_number"] = "Week number must be greater than 0"
		} else if req.WeekNumber > 20 {
			details["week_number"] = "Week number must be less than or equal to 20"
		}
		if req.Name == "" {
			details["name"] = "Week name is required"
		}
		if len(details) > 0 {
			validation.RespondWithError(w, http.StatusBadRequest, "Validation failed", "VALIDATION_ERROR", details)
			return
		}

		// Check that season exists
		var season models.Season
		if err := a.DB.First(&season, req.SeasonID).Error; err != nil {
			validation.RespondWithError(w, http.StatusBadRequest, "Season not found", "SEASON_NOT_FOUND", map[string]string{
				"season_id": "The specified season does not exist",
			})
			return
		}

		week := models.Week{
			SeasonID:   req.SeasonID,
			WeekNumber: req.WeekNumber,
			Name:       req.Name,
			Status:     "creating",
		}

		if err := a.DB.Create(&week).Error; err != nil {
			validation.RespondWithError(w, http.StatusInternalServerError, "Error creating week", "DATABASE_ERROR", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(week)
	}
}

// UpdateWeek returns a handler for updating week details (admin only, only for weeks in 'creating' status)
func UpdateWeek(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		weekID := chi.URLParam(r, "id")

		var week models.Week
		if err := a.DB.First(&week, weekID).Error; err != nil {
			validation.RespondWithError(w, http.StatusNotFound, "Week not found", "WEEK_NOT_FOUND", nil)
			return
		}

		// Only allow editing weeks in 'creating' status
		if week.Status != "creating" {
			validation.RespondWithError(w, http.StatusBadRequest, "Cannot edit week", "INVALID_STATUS", map[string]string{
				"status": "Only weeks in 'creating' status can be edited",
			})
			return
		}

		var req WeekRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			validation.RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON", nil)
			return
		}

		// Validate week data
		details := make(map[string]string)
		if req.SeasonID == 0 {
			details["season_id"] = "Season ID is required"
		}
		if req.WeekNumber <= 0 {
			details["week_number"] = "Week number must be greater than 0"
		} else if req.WeekNumber > 20 {
			details["week_number"] = "Week number must be less than or equal to 20"
		}
		if req.Name == "" {
			details["name"] = "Week name is required"
		}
		if len(details) > 0 {
			validation.RespondWithError(w, http.StatusBadRequest, "Validation failed", "VALIDATION_ERROR", details)
			return
		}

		// Check that season exists
		var season models.Season
		if err := a.DB.First(&season, req.SeasonID).Error; err != nil {
			validation.RespondWithError(w, http.StatusBadRequest, "Season not found", "SEASON_NOT_FOUND", map[string]string{
				"season_id": "The specified season does not exist",
			})
			return
		}

		// Update week fields
		week.SeasonID = req.SeasonID
		week.WeekNumber = req.WeekNumber
		week.Name = req.Name

		if err := a.DB.Save(&week).Error; err != nil {
			validation.RespondWithError(w, http.StatusInternalServerError, "Error updating week", "DATABASE_ERROR", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(week)
	}
}

// OpenWeekForPicks transitions a week from 'creating' to 'picking' status
func OpenWeekForPicks(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		weekID := chi.URLParam(r, "id")

		var week models.Week
		if err := a.DB.Preload("Games").First(&week, weekID).Error; err != nil {
			validation.RespondWithError(w, http.StatusNotFound, "Week not found", "WEEK_NOT_FOUND", nil)
			return
		}

		// Validate current status
		if valErr := validation.ValidateWeekStatusTransition(week.Status, "picking"); valErr != nil {
			validation.RespondWithValidationError(w, valErr)
			return
		}

		// Require at least one game
		if len(week.Games) == 0 {
			validation.RespondWithError(w, http.StatusBadRequest, "Cannot open week for picks", "NO_GAMES", map[string]string{
				"games": "Week must have at least one game before opening for picks",
			})
			return
		}

		// Parse pick deadline from request
		var req struct {
			PickDeadline string `json:"pick_deadline"` // ISO 8601 format
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			validation.RespondWithError(w, http.StatusBadRequest, "Invalid request body", "INVALID_JSON", nil)
			return
		}

		pickDeadline, valErr := validation.ValidatePickDeadline(req.PickDeadline)
		if valErr != nil {
			validation.RespondWithValidationError(w, valErr)
			return
		}

		// Update week status and pick deadline
		week.Status = "picking"
		week.PickDeadline = &pickDeadline

		if err := a.DB.Save(&week).Error; err != nil {
			validation.RespondWithError(w, http.StatusInternalServerError, "Error updating week status", "DATABASE_ERROR", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(week)
	}
}

// LockWeek transitions a week from 'picking' to 'scoring' status
func LockWeek(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		weekID := chi.URLParam(r, "id")

		var week models.Week
		if err := a.DB.First(&week, weekID).Error; err != nil {
			validation.RespondWithError(w, http.StatusNotFound, "Week not found", "WEEK_NOT_FOUND", nil)
			return
		}

		// Validate current status
		if valErr := validation.ValidateWeekStatusTransition(week.Status, "scoring"); valErr != nil {
			validation.RespondWithValidationError(w, valErr)
			return
		}

		// Update week status
		week.Status = "scoring"

		if err := a.DB.Save(&week).Error; err != nil {
			validation.RespondWithError(w, http.StatusInternalServerError, "Error updating week status", "DATABASE_ERROR", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(week)
	}
}

// CompleteWeek transitions a week from 'scoring' to 'finished' status
func CompleteWeek(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		weekID := chi.URLParam(r, "id")

		var week models.Week
		if err := a.DB.Preload("Games").First(&week, weekID).Error; err != nil {
			validation.RespondWithError(w, http.StatusNotFound, "Week not found", "WEEK_NOT_FOUND", nil)
			return
		}

		// Validate current status
		if valErr := validation.ValidateWeekStatusTransition(week.Status, "finished"); valErr != nil {
			validation.RespondWithValidationError(w, valErr)
			return
		}

		// Verify all games are final
		for _, game := range week.Games {
			if !game.IsFinal {
				validation.RespondWithError(w, http.StatusBadRequest, "Cannot complete week", "GAMES_NOT_FINAL", map[string]string{
					"games": "All games must be marked as final before completing the week",
				})
				return
			}
		}

		// Update week status
		week.Status = "finished"

		if err := a.DB.Save(&week).Error; err != nil {
			validation.RespondWithError(w, http.StatusInternalServerError, "Error updating week status", "DATABASE_ERROR", nil)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(week)
	}
}
