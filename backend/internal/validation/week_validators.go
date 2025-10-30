package validation

import (
	"fmt"
	"strings"
	"time"
)

// ValidateCreateWeek validates week creation request
func ValidateCreateWeek(seasonID uint, weekNumber int, name string, lockTime time.Time) *ValidationError {
	details := make(map[string]string)

	// Validate season ID
	if seasonID == 0 {
		details["season_id"] = "Season ID is required"
	}

	// Validate week number
	if weekNumber <= 0 {
		details["week_number"] = "Week number must be greater than 0"
	} else if weekNumber > 20 {
		details["week_number"] = "Week number must be less than or equal to 20"
	}

	// Validate name
	if strings.TrimSpace(name) == "" {
		details["name"] = "Week name is required"
	} else if len(name) > 100 {
		details["name"] = "Week name must be less than 100 characters"
	}

	// Validate lock time
	if lockTime.IsZero() {
		details["lock_time"] = "Lock time is required"
	} else if lockTime.Before(time.Now().Add(-7 * 24 * time.Hour)) {
		details["lock_time"] = "Lock time cannot be more than 7 days in the past"
	}

	if len(details) > 0 {
		return NewValidationError("Validation failed", details)
	}

	return nil
}

// ValidateLockTime validates and parses lock time string
func ValidateLockTime(lockTimeStr string) (time.Time, *ValidationError) {
	if lockTimeStr == "" {
		return time.Time{}, NewValidationError("Validation failed", map[string]string{
			"lock_time": "Lock time is required",
		})
	}

	lockTime, err := time.Parse(time.RFC3339, lockTimeStr)
	if err != nil {
		return time.Time{}, NewValidationError("Validation failed", map[string]string{
			"lock_time": fmt.Sprintf("Invalid lock time format. Expected ISO 8601/RFC3339 format: %v", err),
		})
	}

	return lockTime, nil
}

// ValidateWeekStatusTransition validates that a week can transition from one status to another
func ValidateWeekStatusTransition(currentStatus, newStatus string) *ValidationError {
	validTransitions := map[string][]string{
		"creating":  {"picking"},
		"picking":   {"scoring"},
		"scoring":   {"finished"},
		"finished":  {}, // No transitions allowed from finished
	}

	allowedTransitions, exists := validTransitions[currentStatus]
	if !exists {
		return NewValidationError("Invalid week status", map[string]string{
			"status": fmt.Sprintf("Unknown week status: %s", currentStatus),
		})
	}

	// Check if new status is in allowed transitions
	for _, allowed := range allowedTransitions {
		if allowed == newStatus {
			return nil
		}
	}

	return NewValidationError("Invalid status transition", map[string]string{
		"status": fmt.Sprintf("Cannot transition from %s to %s", currentStatus, newStatus),
	})
}

// ValidatePickDeadline validates and parses pick deadline string
func ValidatePickDeadline(pickDeadlineStr string) (time.Time, *ValidationError) {
	if pickDeadlineStr == "" {
		return time.Time{}, NewValidationError("Validation failed", map[string]string{
			"pick_deadline": "Pick deadline is required",
		})
	}

	pickDeadline, err := time.Parse(time.RFC3339, pickDeadlineStr)
	if err != nil {
		return time.Time{}, NewValidationError("Validation failed", map[string]string{
			"pick_deadline": fmt.Sprintf("Invalid pick deadline format. Expected ISO 8601/RFC3339 format: %v", err),
		})
	}

	// Pick deadline should be in the future
	if pickDeadline.Before(time.Now()) {
		return time.Time{}, NewValidationError("Validation failed", map[string]string{
			"pick_deadline": "Pick deadline must be in the future",
		})
	}

	return pickDeadline, nil
}
