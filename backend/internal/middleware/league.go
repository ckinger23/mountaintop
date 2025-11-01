package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/ckinger23/mountaintop/internal/models"
	"gorm.io/gorm"
)

const LeagueContextKey contextKey = "league_id"

// LeagueContext extracts league_id from query params and validates membership
// Usage: Add to routes that need to be scoped to a specific league
// Example: /api/games?league_id=1
func LeagueContext(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user from context (assumes AuthMiddleware has already run)
			claims, ok := GetUserFromContext(r)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Get league_id from query params
			leagueIDStr := r.URL.Query().Get("league_id")
			if leagueIDStr == "" {
				http.Error(w, "league_id query parameter is required", http.StatusBadRequest)
				return
			}

			leagueID, err := strconv.ParseUint(leagueIDStr, 10, 32)
			if err != nil {
				http.Error(w, "Invalid league_id", http.StatusBadRequest)
				return
			}

			// Verify user is a member of this league
			var membership models.LeagueMembership
			if err := db.Where("league_id = ? AND user_id = ?", leagueID, claims.UserID).First(&membership).Error; err != nil {
				http.Error(w, "You are not a member of this league", http.StatusForbidden)
				return
			}

			// Add league_id to context
			ctx := context.WithValue(r.Context(), LeagueContextKey, uint(leagueID))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LeagueOwnerOnly ensures the user is the owner of the league
// Must be used after LeagueContext middleware
func LeagueOwnerOnly(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetUserFromContext(r)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			leagueID, ok := r.Context().Value(LeagueContextKey).(uint)
			if !ok {
				http.Error(w, "League context not found", http.StatusInternalServerError)
				return
			}

			// Check if user is the owner
			var league models.League
			if err := db.First(&league, leagueID).Error; err != nil {
				http.Error(w, "League not found", http.StatusNotFound)
				return
			}

			if league.OwnerID != claims.UserID {
				http.Error(w, "Only the league owner can perform this action", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetLeagueIDFromContext extracts league_id from the request context
func GetLeagueIDFromContext(ctx context.Context) (uint, bool) {
	leagueID, ok := ctx.Value(LeagueContextKey).(uint)
	return leagueID, ok
}
