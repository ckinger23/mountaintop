package handlers

import (
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"football-picking-league/backend/db"
	"football-picking-league/backend/models"
	"football-picking-league/backend/utils"
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
		tableName := os.Getenv("DYNAMODB_TABLE")
		if tableName == "" {
			tableName = "football-picking-league"
		}

		// Build the query for teams using the GSI-EntityType index
		// In our single-table design, teams have entity_type = 'TEAM'
		keyEx := expression.Key("entity_type").Equal(expression.Value("TEAM"))
		expr, err := expression.NewBuilder().
			WithKeyCondition(keyEx).
			Build()
		if err != nil {
			log.Printf("Error building query expression: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to build query expression")
			return
		}

		// Query the table using the GSI-EntityType index
		result, err := dbClient.QueryItems(r.Context(), &dynamodb.QueryInput{
			TableName:                 aws.String(tableName),
			IndexName:                 aws.String("GSI-EntityType"),
			KeyConditionExpression:    expr.KeyCondition(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
		})
		if err != nil {
			log.Printf("Error querying teams: %v", err)
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch teams")
			return
		}

		if result.Items == nil {
			log.Println("No teams found in the database")
			utils.RespondWithJSON(w, http.StatusOK, TeamsByConferenceResponse{Conferences: []ConferenceTeams{}})
			return
		}

		// Convert the DynamoDB items to Team structs
		var dbTeams []models.Team
		for _, item := range result.Items {
			// Unmarshal the team data
			// Note: In our single-table design, the team ID is part of the PK (PK = "TEAM#{teamId}")
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
