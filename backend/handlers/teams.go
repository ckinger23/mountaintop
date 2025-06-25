package handlers

import (
	"football-picking-league/backend/db"
	"football-picking-league/backend/utils"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gorilla/mux"
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

// CreateTeamRequest represents the request body for creating a team
type CreateTeamRequest struct {
	Name         string `json:"name"`
	ConferenceID string `json:"conference_id"`
}

// CreateTeamHandler handles creating a new team
func CreateTeamHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST method
		if r.Method != http.MethodPost {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Parse request body
		var req CreateTeamRequest
		if err := utils.ParseJSONBody(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		// Validate required fields
		if req.Name == "" || req.ConferenceID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Team name and conference ID are required")
			return
		}

		// Generate a new team ID
		teamID := "TEAM#" + strings.ReplaceAll(req.Name, " ", "_")

		// Create the team item for DynamoDB
		item := map[string]types.AttributeValue{
			"PK":            &types.AttributeValueMemberS{Value: teamID},
			"SK":            &types.AttributeValueMemberS{Value: "METADATA"},
			"entity_type":   &types.AttributeValueMemberS{Value: "TEAM"},
			"id":            &types.AttributeValueMemberS{Value: teamID},
			"conference_id": &types.AttributeValueMemberS{Value: req.ConferenceID},
			"data": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"name": &types.AttributeValueMemberS{Value: req.Name},
				},
			},
		}

		// Save the team to DynamoDB
		if err := dbClient.PutItem(r.Context(), "FootballLeague", item); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create team: "+err.Error())
			return
		}

		// Return the created team
		team := Team{
			ID:           teamID,
			Name:         req.Name,
			ConferenceID: req.ConferenceID,
		}

		utils.RespondWithJSON(w, http.StatusCreated, team)
	}
}

// UpdateTeamRequest represents the request body for updating a team
type UpdateTeamRequest struct {
	Name         *string `json:"name,omitempty"`
	ConferenceID *string `json:"conference_id,omitempty"`
}

// UpdateTeamHandler handles updating a team
func UpdateTeamHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the ID from the URL parameters
		vars := mux.Vars(r)
		id := vars["id"]
		if id == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Team ID is required")
			return
		}

		// Get the existing team first
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "TEAM#" + id},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		}

		existingItem, err := dbClient.GetItem(r.Context(), "FootballLeague", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch team")
			return
		}

		if len(existingItem) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "Team not found")
			return
		}

		// Parse the request body
		var req UpdateTeamRequest
		if err := utils.ParseJSONBody(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		// Check if at least one field is being updated
		if req.Name == nil && req.ConferenceID == nil {
			utils.RespondWithError(w, http.StatusBadRequest, "No fields to update")
			return
		}

		// Get existing data
		var existingData map[string]types.AttributeValue
		if data, ok := existingItem["data"].(*types.AttributeValueMemberM); ok {
			existingData = data.Value
		} else {
			existingData = make(map[string]types.AttributeValue)
		}

		// Update the fields if they are provided in the request
		if req.Name != nil {
			existingData["name"] = &types.AttributeValueMemberS{Value: *req.Name}
		}

		// Prepare the updated item
		updatedAt := time.Now().Format(time.RFC3339)
		item := map[string]types.AttributeValue{
			"PK":            key["PK"],
			"SK":            key["SK"],
			"entity_type":   &types.AttributeValueMemberS{Value: "TEAM"},
			"id":            &types.AttributeValueMemberS{Value: id},
			"conference_id": existingItem["conference_id"],
			"updated_at":    &types.AttributeValueMemberS{Value: updatedAt},
			"data":          &types.AttributeValueMemberM{Value: existingData},
		}

		// Update conference_id if provided
		if req.ConferenceID != nil {
			item["conference_id"] = &types.AttributeValueMemberS{Value: *req.ConferenceID}
		}

		// Save the updated team to DynamoDB
		if err := dbClient.PutItem(r.Context(), "FootballLeague", item); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update team: "+err.Error())
			return
		}

		// Get the updated name (either new or existing)
		updatedName := ""
		if req.Name != nil {
			updatedName = *req.Name
		} else if name, ok := existingData["name"].(*types.AttributeValueMemberS); ok {
			updatedName = name.Value
		}

		// Get the updated conference ID (either new or existing)
		updatedConfID := ""
		if req.ConferenceID != nil {
			updatedConfID = *req.ConferenceID
		} else if confID, ok := existingItem["conference_id"].(*types.AttributeValueMemberS); ok {
			updatedConfID = confID.Value
		}

		// Return the updated team
		utils.RespondWithJSON(w, http.StatusOK, Team{
			ID:           id,
			Name:         updatedName,
			ConferenceID: updatedConfID,
		})
	}
}

// DeleteTeamHandler handles deleting a team by ID
func DeleteTeamHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow DELETE method
		if r.Method != http.MethodDelete {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Get team ID from query parameters
		teamID := r.URL.Query().Get("id")
		if teamID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Team ID is required")
			return
		}

		// Create the key for the team item
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: teamID},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		}

		// First check if the team exists
		_, err := dbClient.GetItem(r.Context(), "FootballLeague", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Team not found")
			return
		}

		// Delete the team from DynamoDB
		if err := dbClient.DeleteItem(r.Context(), "FootballLeague", key); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete team: "+err.Error())
			return
		}

		// Return success response
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "Team deleted successfully",
		})
	}
}
