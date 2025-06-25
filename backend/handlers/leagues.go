package handlers

import (
	"football-picking-league/backend/db"
	"football-picking-league/backend/models"
	"football-picking-league/backend/utils"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gorilla/mux"
)

// LeagueResponse represents the response structure for league data
type LeagueResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	AdminID   string    `json:"admin_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateLeagueRequest represents the request body for creating a league
type CreateLeagueRequest struct {
	Name    string `json:"name"`
	AdminID string `json:"admin_id"`
}

// LeaguesResponse represents the response structure for multiple leagues
type LeaguesResponse struct {
	Leagues []LeagueResponse `json:"leagues"`
}

// GetLeagueHandler returns a single league by ID
func GetLeagueHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the ID from the URL parameters
		id := r.URL.Query().Get("id")
		if id == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "League ID is required")
			return
		}

		// Create the key for the query
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "LEAGUE#" + id},
			"SK": &types.AttributeValueMemberS{Value: "METADATA#" + id},
		}

		// Get the item from DynamoDB
		item, err := dbClient.GetItem(r.Context(), "FootballLeague", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch league")
			return
		}

		// Check if the item exists
		if len(item) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "League not found")
			return
		}

		// Unmarshal the league data
		var league models.League
		if data, ok := item["data"].(*types.AttributeValueMemberM); ok {
			err = attributevalue.UnmarshalMap(data.Value, &league)
			if err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "Error processing league data")
				return
			}
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid league data format")
			return
		}

		// Return the response
		utils.RespondWithJSON(w, http.StatusOK, LeagueResponse{
			ID:        league.ID,
			Name:      league.Name,
			AdminID:   league.AdminID,
			CreatedAt: league.CreatedAt,
			UpdatedAt: league.UpdatedAt,
		})
	}
}

// GetAllLeaguesHandler returns a list of all leagues
func GetAllLeaguesHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Build the query to get all leagues
		expr, err := expression.NewBuilder().
			WithKeyCondition(expression.Key("entity_type").Equal(expression.Value("LEAGUE"))).
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
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch leagues")
			return
		}


		// Process the results
		var leagues []models.League
		for _, item := range result.Items {
			if data, ok := item["data"].(*types.AttributeValueMemberM); ok {
				var league models.League
				err = attributevalue.UnmarshalMap(data.Value, &league)
				if err == nil {
					leagues = append(leagues, league)
				}
			}
		}

		// Convert to response format
		response := LeaguesResponse{
			Leagues: make([]LeagueResponse, 0, len(leagues)),
		}

		for _, league := range leagues {
			response.Leagues = append(response.Leagues, LeagueResponse{
				ID:        league.ID,
				Name:      league.Name,
				AdminID:   league.AdminID,
				CreatedAt: league.CreatedAt,
				UpdatedAt: league.UpdatedAt,
			})
		}

		utils.RespondWithJSON(w, http.StatusOK, response)
	}
}

// CreateLeagueHandler handles creating a new league
func CreateLeagueHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST method
		if r.Method != http.MethodPost {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Parse request body
		var req CreateLeagueRequest
		if err := utils.ParseJSONBody(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		// Validate required fields
		if req.Name == "" || req.AdminID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "League name and admin ID are required")
			return
		}

		// Generate a league ID from the name (uppercase with underscores)
		leagueID := strings.ReplaceAll(strings.ToUpper(req.Name), " ", "_")

		now := time.Now()
		league := models.League{
			ID:        leagueID,
			Name:      req.Name,
			AdminID:   req.AdminID,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Marshal the league data
		data, err := attributevalue.MarshalMap(league)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process league data")
			return
		}

		// Create the item to save
		item := map[string]types.AttributeValue{
			"PK":          &types.AttributeValueMemberS{Value: "LEAGUE#" + leagueID},
			"SK":          &types.AttributeValueMemberS{Value: "METADATA#" + leagueID},
			"entity_type": &types.AttributeValueMemberS{Value: "LEAGUE"},
			"id":          &types.AttributeValueMemberS{Value: leagueID},
			"admin_id":    &types.AttributeValueMemberS{Value: req.AdminID},
			"created_at":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			"updated_at":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			"data":        &types.AttributeValueMemberM{Value: data},
		}

		// Save the league to DynamoDB
		if err := dbClient.PutItem(r.Context(), "FootballLeague", item); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create league: "+err.Error())
			return
		}

		// Return the created league
		utils.RespondWithJSON(w, http.StatusCreated, LeagueResponse{
			ID:        league.ID,
			Name:      league.Name,
			AdminID:   league.AdminID,
			CreatedAt: league.CreatedAt,
			UpdatedAt: league.UpdatedAt,
		})
	}
}

// UpdateLeagueRequest represents the request body for updating a league
type UpdateLeagueRequest struct {
	Name    *string `json:"name,omitempty"`
	AdminID *string `json:"admin_id,omitempty"`
}

// UpdateLeagueHandler handles updating a league
func UpdateLeagueHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the ID from the URL parameters
		vars := mux.Vars(r)
		id := vars["id"]
		if id == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "League ID is required")
			return
		}

		// Get the existing league first
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "LEAGUE#" + id},
			"SK": &types.AttributeValueMemberS{Value: "METADATA#" + id},
		}

		existingItem, err := dbClient.GetItem(r.Context(), "FootballLeague", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch league")
			return
		}

		if len(existingItem) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "League not found")
			return
		}

		// Parse the request body
		var req UpdateLeagueRequest
		if err := utils.ParseJSONBody(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		// Check if at least one field is being updated
		if req.Name == nil && req.AdminID == nil {
			utils.RespondWithError(w, http.StatusBadRequest, "No fields to update")
			return
		}

		// Unmarshal the existing data
		var existingData map[string]types.AttributeValue
		if data, ok := existingItem["data"].(*types.AttributeValueMemberM); ok {
			existingData = data.Value
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid league data format")
			return
		}

		// Update the fields if they are provided in the request
		if req.Name != nil {
			existingData["name"] = &types.AttributeValueMemberS{Value: *req.Name}
		}

		// Prepare the updated item
		updatedAt := time.Now().Format(time.RFC3339)
		item := map[string]types.AttributeValue{
			"PK":          key["PK"],
			"SK":          key["SK"],
			"entity_type": &types.AttributeValueMemberS{Value: "LEAGUE"},
			"id":          &types.AttributeValueMemberS{Value: id},
			"created_at":  existingItem["created_at"],
			"updated_at":  &types.AttributeValueMemberS{Value: updatedAt},
			"data":        &types.AttributeValueMemberM{Value: existingData},
		}

		// Update admin_id if provided
		if req.AdminID != nil {
			item["admin_id"] = &types.AttributeValueMemberS{Value: *req.AdminID}
		} else if adminID, ok := existingItem["admin_id"]; ok {
			item["admin_id"] = adminID
		}

		// Save the updated league to DynamoDB
		if err := dbClient.PutItem(r.Context(), "FootballLeague", item); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update league: "+err.Error())
			return
		}

		// Get the created_at time from the existing item
		var createdAt string
		if created, ok := existingItem["created_at"].(*types.AttributeValueMemberS); ok {
			createdAt = created.Value
		}

		// Parse the created_at time
		parsedCreatedAt, _ := time.Parse(time.RFC3339, createdAt)

		// Get the updated name (either new or existing)
		updatedName := ""
		if req.Name != nil {
			updatedName = *req.Name
		} else if name, ok := existingData["name"].(*types.AttributeValueMemberS); ok {
			updatedName = name.Value
		}

		// Get the updated admin ID (either new or existing)
		updatedAdminID := ""
		if req.AdminID != nil {
			updatedAdminID = *req.AdminID
		} else if adminID, ok := existingItem["admin_id"].(*types.AttributeValueMemberS); ok {
			updatedAdminID = adminID.Value
		}

		// Return the updated league
		utils.RespondWithJSON(w, http.StatusOK, LeagueResponse{
			ID:        id,
			Name:      updatedName,
			AdminID:   updatedAdminID,
			CreatedAt: parsedCreatedAt,
			UpdatedAt: time.Now(),
		})
	}
}

// DeleteLeagueHandler handles deleting a league by ID
func DeleteLeagueHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow DELETE method
		if r.Method != http.MethodDelete {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Get league ID from query parameters
		leagueID := r.URL.Query().Get("id")
		if leagueID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "League ID is required")
			return
		}

		// Create the key for the league item
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "LEAGUE#" + leagueID},
			"SK": &types.AttributeValueMemberS{Value: "METADATA#" + leagueID},
		}

		// First check if the league exists
		_, err := dbClient.GetItem(r.Context(), "FootballLeague", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "League not found")
			return
		}

		// Delete the league from DynamoDB
		if err := dbClient.DeleteItem(r.Context(), "FootballLeague", key); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete league: "+err.Error())
			return
		}

		// Return success response
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "League deleted successfully",
		})
	}
}
