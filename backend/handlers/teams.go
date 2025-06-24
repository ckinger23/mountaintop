package handlers

import (
	"football-picking-league/backend/db"
	"football-picking-league/backend/utils"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Team struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ConferenceID string `json:"conference_id"`
}

type ConferenceTeams struct {
	Name  string   `json:"name"`
	Teams []string `json:"teams"`
}

type TeamsResponse struct {
	Conferences []ConferenceTeams `json:"conferences"`
}

func GetAllTeamsHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Build the query to get all teams
		expr, err := expression.NewBuilder().
			WithKeyCondition(expression.Key("entity_type").Equal(expression.Value("TEAM"))).
			Build()
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to build query")
			return
		}

		// Query the table using the GSI-EntityType index
		result, err := dbClient.QueryItems(r.Context(), &dynamodb.QueryInput{
			TableName:                 aws.String("FootballLeague"),
			IndexName:                 aws.String("GSI-EntityType"),
			KeyConditionExpression:    expr.KeyCondition(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch teams")
			return
		}

		// Convert DynamoDB items to Team structs
		var teams []Team
		for _, item := range result.Items {
			var team Team
			if id, ok := item["id"]; ok {
				if idVal, ok := id.(*types.AttributeValueMemberS); ok {
					team.ID = idVal.Value
				}
			}
			if data, ok := item["data"]; ok {
				if dataMap, ok := data.(*types.AttributeValueMemberM); ok {
					if name, ok := dataMap.Value["name"].(*types.AttributeValueMemberS); ok {
						team.Name = name.Value
					}
				}
			}
			if confID, ok := item["conference_id"]; ok {
				if confIDVal, ok := confID.(*types.AttributeValueMemberS); ok {
					team.ConferenceID = confIDVal.Value
				}
			}
			teams = append(teams, team)
		}

		// Group teams by conference
		conferenceMap := map[string][]string{
			"1": make([]string, 0), // SEC
			"2": make([]string, 0), // Big Ten
			"3": make([]string, 0), // Big 12
			"4": make([]string, 0), // ACC
		}

		conferenceNames := map[string]string{
			"1": "SEC",
			"2": "Big Ten",
			"3": "Big 12",
			"4": "ACC",
		}

		for _, team := range teams {
			if teams, exists := conferenceMap[team.ConferenceID]; exists {
				conferenceMap[team.ConferenceID] = append(teams, team.Name)
			}
		}

		// Convert to response format
		var response TeamsResponse
		for confID, teamNames := range conferenceMap {
			response.Conferences = append(response.Conferences, ConferenceTeams{
				Name:  conferenceNames[confID],
				Teams: teamNames,
			})
		}

		utils.RespondWithJSON(w, http.StatusOK, response)
	}
}

func GetSingleTeamHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get and validate team name from query parameters
		teamName := strings.TrimSpace(r.URL.Query().Get("name"))
		if teamName == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Query parameter 'name' is required and cannot be empty")
			return
		}

		// Create filter expression for the query
		filt := expression.Name("data").AttributeExists().And(
			expression.Name("data.name").Equal(expression.Value(teamName)),
		)

		expr, err := expression.NewBuilder().WithFilter(filt).Build()
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to build query expression")
			return
		}

		// Query DynamoDB
		result, err := dbClient.Scan(r.Context(), &dynamodb.ScanInput{
			TableName:                 aws.String("teams"),
			FilterExpression:          expr.Filter(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to query team: "+err.Error())
			return
		}

		// Check if any team was found
		if len(result.Items) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "Team not found")
			return
		}

		// Convert the first matching item to Team struct
		var team Team
		item := result.Items[0]

		if id, ok := item["id"]; ok {
			if idVal, ok := id.(*types.AttributeValueMemberS); ok {
				team.ID = idVal.Value
			}
		}

		if data, ok := item["data"]; ok {
			if dataMap, ok := data.(*types.AttributeValueMemberM); ok {
				if name, ok := dataMap.Value["name"].(*types.AttributeValueMemberS); ok {
					team.Name = name.Value
				}
			}
		}

		if confID, ok := item["conference_id"]; ok {
			if confIDVal, ok := confID.(*types.AttributeValueMemberS); ok {
				team.ConferenceID = confIDVal.Value
			}
		}

		utils.RespondWithJSON(w, http.StatusOK, team)
	}
}
