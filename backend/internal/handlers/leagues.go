package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ckinger23/mountaintop/internal/app"
	"github.com/ckinger23/mountaintop/internal/middleware"
	"github.com/ckinger23/mountaintop/internal/models"
	"github.com/go-chi/chi/v5"
)

// CreateLeagueRequest is the request body for creating a league
type CreateLeagueRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// JoinLeagueRequest is the request body for joining a league by code
type JoinLeagueRequest struct {
	Code string `json:"code"`
}

// CreateLeague creates a new league owned by the authenticated user
func CreateLeague(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.GetUserFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req CreateLeagueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate input
		if req.Name == "" {
			http.Error(w, "League name is required", http.StatusBadRequest)
			return
		}

		// Check if user already owns a league (uniqueIndex constraint on OwnerID)
		var existingLeague models.League
		if err := a.DB.Where("owner_id = ?", claims.UserID).First(&existingLeague).Error; err == nil {
			http.Error(w, "You already own a league. Each user can only own one league.", http.StatusBadRequest)
			return
		}

		// Generate unique league code
		code, err := generateLeagueCode()
		if err != nil {
			http.Error(w, "Failed to generate league code", http.StatusInternalServerError)
			return
		}

		// Create league
		league := models.League{
			Name:        req.Name,
			Code:        code,
			Description: req.Description,
			OwnerID:     claims.UserID,
			IsPublic:    req.IsPublic,
			IsActive:    true,
		}

		if err := a.DB.Create(&league).Error; err != nil {
			http.Error(w, "Failed to create league: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Create membership for the owner
		membership := models.LeagueMembership{
			LeagueID: league.ID,
			UserID:   claims.UserID,
			Role:     "owner",
			JoinedAt: time.Now(),
		}

		if err := a.DB.Create(&membership).Error; err != nil {
			http.Error(w, "Failed to create league membership", http.StatusInternalServerError)
			return
		}

		// Preload owner for response
		a.DB.Preload("Owner").First(&league, league.ID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(league)
	}
}

// GetMyLeagues returns all leagues the authenticated user is a member of
func GetMyLeagues(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.GetUserFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get all memberships for this user
		var memberships []models.LeagueMembership
		if err := a.DB.Where("user_id = ?", claims.UserID).Preload("League.Owner").Find(&memberships).Error; err != nil {
			http.Error(w, "Failed to fetch leagues", http.StatusInternalServerError)
			return
		}

		// Extract leagues from memberships
		leagues := make([]models.League, len(memberships))
		for i, m := range memberships {
			leagues[i] = m.League
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(leagues)
	}
}

// GetLeague returns a specific league by ID
func GetLeague(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.GetUserFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		leagueID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
		if err != nil {
			http.Error(w, "Invalid league ID", http.StatusBadRequest)
			return
		}

		var league models.League
		if err := a.DB.Preload("Owner").Preload("Members.User").First(&league, leagueID).Error; err != nil {
			http.Error(w, "League not found", http.StatusNotFound)
			return
		}

		// Verify user is a member of this league
		var membership models.LeagueMembership
		if err := a.DB.Where("league_id = ? AND user_id = ?", leagueID, claims.UserID).First(&membership).Error; err != nil {
			http.Error(w, "You are not a member of this league", http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(league)
	}
}

// UpdateLeague updates league settings (owner only)
func UpdateLeague(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.GetUserFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		leagueID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
		if err != nil {
			http.Error(w, "Invalid league ID", http.StatusBadRequest)
			return
		}

		var league models.League
		if err := a.DB.First(&league, leagueID).Error; err != nil {
			http.Error(w, "League not found", http.StatusNotFound)
			return
		}

		// Only owner can update league settings
		if league.OwnerID != claims.UserID {
			http.Error(w, "Only the league owner can update settings", http.StatusForbidden)
			return
		}

		var req CreateLeagueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Update fields
		if req.Name != "" {
			league.Name = req.Name
		}
		league.Description = req.Description
		league.IsPublic = req.IsPublic

		if err := a.DB.Save(&league).Error; err != nil {
			http.Error(w, "Failed to update league", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(league)
	}
}

// JoinLeague allows a user to join a league by code
func JoinLeague(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.GetUserFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req JoinLeagueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Normalize code (uppercase, no spaces)
		code := strings.ToUpper(strings.TrimSpace(req.Code))
		if code == "" {
			http.Error(w, "League code is required", http.StatusBadRequest)
			return
		}

		// Find league by code
		var league models.League
		if err := a.DB.Where("code = ?", code).First(&league).Error; err != nil {
			http.Error(w, "League not found with that code", http.StatusNotFound)
			return
		}

		// Check if league is active
		if !league.IsActive {
			http.Error(w, "This league is no longer active", http.StatusBadRequest)
			return
		}

		// Check if user is already a member
		var existingMembership models.LeagueMembership
		if err := a.DB.Where("league_id = ? AND user_id = ?", league.ID, claims.UserID).First(&existingMembership).Error; err == nil {
			http.Error(w, "You are already a member of this league", http.StatusBadRequest)
			return
		}

		// Create membership
		membership := models.LeagueMembership{
			LeagueID: league.ID,
			UserID:   claims.UserID,
			Role:     "member",
			JoinedAt: time.Now(),
		}

		if err := a.DB.Create(&membership).Error; err != nil {
			http.Error(w, "Failed to join league", http.StatusInternalServerError)
			return
		}

		// Preload owner for response
		a.DB.Preload("Owner").First(&league, league.ID)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(league)
	}
}

// BrowsePublicLeagues returns a list of public leagues users can join
func BrowsePublicLeagues(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := middleware.GetUserFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var leagues []models.League
		if err := a.DB.Where("is_public = ? AND is_active = ?", true, true).Preload("Owner").Find(&leagues).Error; err != nil {
			http.Error(w, "Failed to fetch public leagues", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(leagues)
	}
}

// LeaveLeague allows a user to leave a league (except owner)
func LeaveLeague(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.GetUserFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		leagueID, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 32)
		if err != nil {
			http.Error(w, "Invalid league ID", http.StatusBadRequest)
			return
		}

		var league models.League
		if err := a.DB.First(&league, leagueID).Error; err != nil {
			http.Error(w, "League not found", http.StatusNotFound)
			return
		}

		// Owner cannot leave their own league
		if league.OwnerID == claims.UserID {
			http.Error(w, "League owners cannot leave their league. Delete the league instead.", http.StatusBadRequest)
			return
		}

		// Delete membership
		result := a.DB.Where("league_id = ? AND user_id = ?", leagueID, claims.UserID).Delete(&models.LeagueMembership{})
		if result.Error != nil {
			http.Error(w, "Failed to leave league", http.StatusInternalServerError)
			return
		}

		if result.RowsAffected == 0 {
			http.Error(w, "You are not a member of this league", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// generateLeagueCode generates a random league code in format "CFB-XXXX"
func generateLeagueCode() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	code := hex.EncodeToString(bytes)
	return fmt.Sprintf("CFB-%s", strings.ToUpper(code)), nil
}
