package handlers

import (
	"net/http"
	"os"
	"sort"
	"strings"

	"football-picking-league/backend/db"
	"football-picking-league/backend/models"
	"football-picking-league/backend/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// TeamRecord represents a team's record in the API response
type TeamRecord struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	ConferenceID  string  `json:"conferenceId"`
	Conference    string  `json:"conference,omitempty"`
	Wins          int     `json:"wins"`
	Losses        int     `json:"losses"`
	WinPercentage float64 `json:"winPercentage"`
}

// ConferenceTeams groups teams by conference in the API response
type ConferenceTeams struct {
	ConferenceID   string       `json:"id"`
	ConferenceName string       `json:"name"`
	Teams          []TeamRecord `json:"teams"`
}

// TeamsByConferenceResponse is the top-level API response structure
type TeamsByConferenceResponse struct {
	Conferences []ConferenceTeams `json:"conferences"`
}

// conferenceMap maps conference IDs to their names
var conferenceMap = map[string]string{
	"1": "SEC",
	"2": "Big Ten",
	"3": "ACC",
	"4": "Big 12",
	"5": "Pac-12",
}

// calculateWinPercentage calculates the win percentage (0.0 to 1.0)
func calculateWinPercentage(wins, losses int) float64 {
	totalGames := wins + losses
	if totalGames == 0 {
		return 0.0
	}
	return float64(wins) / float64(totalGames)
}

// GetTeamsRecordsHandler returns a handler function that fetches and returns team records grouped by conference
func GetTeamsRecordsHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the table name from environment variables or use a default
		tableName := os.Getenv("TEAMS_TABLE_NAME")
		if tableName == "" {
			tableName = "teams"
		}

		// Get all teams from the database using Scan operation
		// Note: For production with large datasets, consider using pagination
		result, err := dbClient.Scan(r.Context(), &dynamodb.ScanInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch teams: "+err.Error())
			return
		}

		// Convert the DynamoDB items to Team structs
		var dbTeams []models.Team
		for _, item := range result.Items {
			var team models.Team
			if err := attributevalue.UnmarshalMap(item, &team); err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process team data: "+err.Error())
				return
			}
			dbTeams = append(dbTeams, team)
		}

		// Group teams by conference and convert to API response format
		conferenceGroups := make(map[string][]TeamRecord)
		for _, team := range dbTeams {
			// Skip teams with unknown conference IDs
			if _, ok := conferenceMap[team.ConferenceID]; !ok {
				continue
			}

			record := TeamRecord{
				ID:            team.ID,
				Name:          team.Name,
				ConferenceID:  team.ConferenceID,
				Conference:    conferenceMap[team.ConferenceID],
				Wins:          team.Wins,
				Losses:        team.Losses,
				WinPercentage: calculateWinPercentage(team.Wins, team.Losses),
			}

			conferenceGroups[team.ConferenceID] = append(conferenceGroups[team.ConferenceID], record)
		}

		// Convert to response format and sort
		var response TeamsByConferenceResponse
		for confID, teams := range conferenceGroups {
			confName, ok := conferenceMap[confID]
			if !ok {
				continue
			}
			// Sort teams by win percentage (descending)
			sort.Slice(teams, func(i, j int) bool {
				if teams[i].WinPercentage != teams[j].WinPercentage {
					return teams[i].WinPercentage > teams[j].WinPercentage
				}
				// If win percentage is equal, sort by fewer losses
				if teams[i].Losses != teams[j].Losses {
					return teams[i].Losses < teams[j].Losses
				}
				// If still equal, sort by name
				return strings.Compare(teams[i].Name, teams[j].Name) < 0
			})

			response.Conferences = append(response.Conferences, ConferenceTeams{
				ConferenceID:   confID,
				ConferenceName: confName,
				Teams:          teams,
			})
		}

		// Sort conferences by name
		sort.Slice(response.Conferences, func(i, j int) bool {
			return strings.Compare(
				response.Conferences[i].ConferenceName,
				response.Conferences[j].ConferenceName,
			) < 0
		})

		// Respond with the sorted data
		utils.RespondWithJSON(w, http.StatusOK, response)
	}
}
