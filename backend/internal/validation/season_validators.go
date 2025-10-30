package validation

import (
	"strings"
	"time"
)

// ValidateCreateSeason validates season creation request
func ValidateCreateSeason(year int, name string) *ValidationError {
	details := make(map[string]string)

	currentYear := time.Now().Year()

	// Validate year
	if year == 0 {
		details["year"] = "Year is required"
	} else if year < 2000 || year > currentYear+10 {
		details["year"] = "Year must be between 2000 and 10 years in the future"
	}

	// Validate name
	if strings.TrimSpace(name) == "" {
		details["name"] = "Season name is required"
	} else if len(name) > 100 {
		details["name"] = "Season name must be less than 100 characters"
	}

	if len(details) > 0 {
		return NewValidationError("Validation failed", details)
	}

	return nil
}
