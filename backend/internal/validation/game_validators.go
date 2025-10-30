package validation

import (
	"fmt"
	"time"
)

// ValidateCreateGame validates game creation request
func ValidateCreateGame(weekID, homeTeamID, awayTeamID uint, gameTime time.Time, homeSpread, total float64) *ValidationError {
	details := make(map[string]string)

	// Validate week ID
	if weekID == 0 {
		details["week_id"] = "Week ID is required"
	}

	// Validate team IDs
	if homeTeamID == 0 {
		details["home_team_id"] = "Home team ID is required"
	}
	if awayTeamID == 0 {
		details["away_team_id"] = "Away team ID is required"
	}
	if homeTeamID == awayTeamID && homeTeamID != 0 {
		details["teams"] = "Home team and away team cannot be the same"
	}

	// Validate game time
	if gameTime.IsZero() {
		details["game_time"] = "Game time is required"
	} else if gameTime.Before(time.Now().Add(-24 * time.Hour)) {
		details["game_time"] = "Game time cannot be more than 24 hours in the past"
	}

	// Validate home spread (reasonable range)
	if homeSpread < -100 || homeSpread > 100 {
		details["home_spread"] = "Home spread must be between -100 and 100"
	}

	// Validate total (must be positive and reasonable)
	if total <= 0 {
		details["total"] = "Total must be greater than 0"
	} else if total > 200 {
		details["total"] = "Total must be less than or equal to 200"
	}

	if len(details) > 0 {
		return NewValidationError("Validation failed", details)
	}

	return nil
}

// ValidateUpdateGameResult validates game result update
func ValidateUpdateGameResult(homeScore, awayScore *int, isFinal bool) *ValidationError {
	details := make(map[string]string)

	// If marking as final, scores must be provided
	if isFinal {
		if homeScore == nil {
			details["home_score"] = "Home score is required when marking game as final"
		}
		if awayScore == nil {
			details["away_score"] = "Away score is required when marking game as final"
		}
	}

	// Validate score values if provided
	if homeScore != nil {
		if *homeScore < 0 {
			details["home_score"] = "Home score cannot be negative"
		} else if *homeScore > 200 {
			details["home_score"] = "Home score must be less than or equal to 200"
		}
	}

	if awayScore != nil {
		if *awayScore < 0 {
			details["away_score"] = "Away score cannot be negative"
		} else if *awayScore > 200 {
			details["away_score"] = "Away score must be less than or equal to 200"
		}
	}

	if len(details) > 0 {
		return NewValidationError("Validation failed", details)
	}

	return nil
}

// ValidateGameTime validates and parses game time string
func ValidateGameTime(gameTimeStr string) (time.Time, *ValidationError) {
	if gameTimeStr == "" {
		return time.Time{}, NewValidationError("Validation failed", map[string]string{
			"game_time": "Game time is required",
		})
	}

	gameTime, err := time.Parse(time.RFC3339, gameTimeStr)
	if err != nil {
		return time.Time{}, NewValidationError("Validation failed", map[string]string{
			"game_time": fmt.Sprintf("Invalid game time format. Expected ISO 8601/RFC3339 format: %v", err),
		})
	}

	return gameTime, nil
}
