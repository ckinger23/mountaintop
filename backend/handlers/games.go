package handlers

import (
	"football-picking-league/backend/db"
	"football-picking-league/backend/models"
	"football-picking-league/backend/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// GameResponse represents the response structure for game data
type GameResponse struct {
	ID         string    `json:"id"`
	LeagueID   string    `json:"league_id"`
	Week       int       `json:"week"`
	HomeTeamID string    `json:"home_team_id"`
	AwayTeamID string    `json:"away_team_id"`
	GameDate   time.Time `json:"game_date"`
	Status     string    `json:"status"`           // pending/in_progress/completed
	Winner     string    `json:"winner,omitempty"` // home/away
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateGameRequest represents the request body for creating a game
type CreateGameRequest struct {
	LeagueID   string    `json:"league_id"`
	Week       int       `json:"week"`
	HomeTeamID string    `json:"home_team_id"`
	AwayTeamID string    `json:"away_team_id"`
	GameDate   time.Time `json:"game_date"`
	Status     string    `json:"status,omitempty"`
}

// UpdateGameRequest represents the request body for updating a game
type UpdateGameRequest struct {
	Status *string `json:"status,omitempty"`
	Winner *string `json:"winner,omitempty"`
}

// GamesResponse represents the response structure for multiple games
type GamesResponse struct {
	Games []GameResponse `json:"games"`
}

// GetAllGamesHandler returns all games for a specific league and optionally filtered by week
func GetAllGamesHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get league ID from URL parameters
		leagueID := r.URL.Query().Get("league_id")
		if leagueID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "League ID is required")
			return
		}

		// Get optional week parameter
		var week *int
		if weekStr := r.URL.Query().Get("week"); weekStr != "" {
			weekInt, err := strconv.Atoi(weekStr)
			if err != nil {
				utils.RespondWithError(w, http.StatusBadRequest, "Invalid week parameter")
				return
			}
			week = &weekInt
		}

		// Build the key condition for the query
		keyCondition := expression.Key("PK").Equal(expression.Value("LEAGUE#" + leagueID)).
			And(expression.Key("SK").BeginsWith("GAME#"))

		// Add week filter if provided
		var filter expression.ConditionBuilder
		if week != nil {
			filter = expression.Name("week").Equal(expression.Value(*week))
		}

		// Build the expression
		expr, err := expression.NewBuilder().
			WithKeyCondition(keyCondition).
			WithFilter(filter).
			Build()
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to build query")
			return
		}

		// Query the table
		result, err := dbClient.QueryItems(r.Context(), &dynamodb.QueryInput{
			TableName:                 aws.String("FootballLeague"),
			KeyConditionExpression:    expr.KeyCondition(),
			FilterExpression:          expr.Filter(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch games")
			return
		}

		// Process the results
		var games []models.Game
		for _, item := range result.Items {
			if data, ok := item["data"].(*types.AttributeValueMemberM); ok {
				var game models.Game
				err = attributevalue.UnmarshalMap(data.Value, &game)
				if err == nil {
					games = append(games, game)
				}
			}
		}

		// Convert to response format
		response := GamesResponse{
			Games: make([]GameResponse, 0, len(games)),
		}

		for _, game := range games {
			gameResp := GameResponse{
				ID:         game.ID,
				LeagueID:   game.LeagueID,
				Week:       game.Week,
				HomeTeamID: game.HomeTeamID,
				AwayTeamID: game.AwayTeamID,
				GameDate:   game.GameDate,
				Status:     game.Status,
				CreatedAt:  game.CreatedAt,
				UpdatedAt:  game.UpdatedAt,
			}
			// Only include winner if the game is completed
			if game.Status == "completed" {
				gameResp.Winner = game.Winner
			}
			response.Games = append(response.Games, gameResp)
		}

		utils.RespondWithJSON(w, http.StatusOK, response)
	}
}

// GetGameHandler returns a single game by ID
func GetGameHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the ID from the query parameters
		gameID := r.URL.Query().Get("id")
		if gameID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Game ID is required")
			return
		}

		// Use GSI1 to find the game efficiently
		result, err := dbClient.QueryItems(r.Context(), &dynamodb.QueryInput{
			TableName:              aws.String("FootballLeague"),
			IndexName:              aws.String("GSI1-EntityLookup"),
			KeyConditionExpression: aws.String("GSI1_PK = :pk AND GSI1_SK = :sk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "GAME"},
				":sk": &types.AttributeValueMemberS{Value: "GAME#" + gameID},
			},
			Limit: aws.Int32(1),
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch game")
			return
		}

		// Check if the item exists
		if len(result.Items) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "Game not found")
			return
		}

		item := result.Items[0]

		// Unmarshal the game data
		var game models.Game
		if data, ok := item["data"].(*types.AttributeValueMemberM); ok {
			err = attributevalue.UnmarshalMap(data.Value, &game)
			if err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "Error processing game data")
				return
			}
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid game data format")
			return
		}

		// Create the response
		response := GameResponse{
			ID:         game.ID,
			LeagueID:   game.LeagueID,
			Week:       game.Week,
			HomeTeamID: game.HomeTeamID,
			AwayTeamID: game.AwayTeamID,
			GameDate:   game.GameDate,
			Status:     game.Status,
			CreatedAt:  game.CreatedAt,
			UpdatedAt:  game.UpdatedAt,
		}

		// Only include winner if the game is completed
		if game.Status == "completed" {
			response.Winner = game.Winner
		}

		utils.RespondWithJSON(w, http.StatusOK, response)
	}
}

// CreateGameHandler handles creating a new game
func CreateGameHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST method
		if r.Method != http.MethodPost {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Parse request body
		var req CreateGameRequest
		if err := utils.ParseJSONBody(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		// Validate required fields
		if req.LeagueID == "" || req.HomeTeamID == "" || req.AwayTeamID == "" || req.Week == 0 {
			utils.RespondWithError(w, http.StatusBadRequest, "League ID, home team ID, away team ID, and week are required")
			return
		}

		// Validate status if provided
		if req.Status != "" && !isValidStatus(req.Status) {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid status. Must be 'pending', 'in_progress', or 'completed'")
			return
		}

		// Generate a new game ID
		gameID := uuid.New().String()
		now := time.Now()

		// Set default status if not provided
		if req.Status == "" {
			req.Status = "pending"
		}

		// Create the game object
		game := models.Game{
			ID:         gameID,
			LeagueID:   req.LeagueID,
			Week:       req.Week,
			HomeTeamID: req.HomeTeamID,
			AwayTeamID: req.AwayTeamID,
			GameDate:   req.GameDate,
			Status:     req.Status,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		// Marshal the game data
		data, err := attributevalue.MarshalMap(game)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process game data")
			return
		}

		// Create the item to save
		item := map[string]types.AttributeValue{
			"PK":          &types.AttributeValueMemberS{Value: "LEAGUE#" + req.LeagueID},
			"SK":          &types.AttributeValueMemberS{Value: "GAME#" + gameID},
			"GSI1_PK":     &types.AttributeValueMemberS{Value: "GAME"},
			"GSI1_SK":     &types.AttributeValueMemberS{Value: "GAME#" + gameID},
			"entity_type": &types.AttributeValueMemberS{Value: "GAME"},
			"id":          &types.AttributeValueMemberS{Value: gameID},
			"league_id":   &types.AttributeValueMemberS{Value: req.LeagueID},
			"week":        &types.AttributeValueMemberN{Value: strconv.Itoa(req.Week)},
			"status":      &types.AttributeValueMemberS{Value: req.Status},
			"created_at":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			"updated_at":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			"data":        &types.AttributeValueMemberM{Value: data},
		}

		// Save the game to DynamoDB
		if err := dbClient.PutItem(r.Context(), "FootballLeague", item); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create game: "+err.Error())
			return
		}

		// Return the created game
		utils.RespondWithJSON(w, http.StatusCreated, GameResponse{
			ID:         game.ID,
			LeagueID:   game.LeagueID,
			Week:       game.Week,
			HomeTeamID: game.HomeTeamID,
			AwayTeamID: game.AwayTeamID,
			GameDate:   game.GameDate,
			Status:     game.Status,
			CreatedAt:  game.CreatedAt,
			UpdatedAt:  game.UpdatedAt,
		})
	}
}

// UpdateGameHandler handles updating a game
func UpdateGameHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow PUT method
		if r.Method != http.MethodPut {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Get game ID from query parameters
		gameID := r.URL.Query().Get("id")
		if gameID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Game ID is required")
			return
		}

		// Parse request body
		var req UpdateGameRequest
		if err := utils.ParseJSONBody(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		// Validate that at least one field is being updated
		if req.Status == nil && req.Winner == nil {
			utils.RespondWithError(w, http.StatusBadRequest, "No fields to update")
			return
		}

		// Validate status if provided
		if req.Status != nil && !isValidStatus(*req.Status) {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid status. Must be 'pending', 'in_progress', or 'completed'")
			return
		}

		// Validate winner if provided
		if req.Winner != nil && *req.Winner != "home" && *req.Winner != "away" {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid winner. Must be 'home' or 'away'")
			return
		}

		// First, find the game using GSI to get its league ID
		gsiResult, err := dbClient.QueryItems(r.Context(), &dynamodb.QueryInput{
			TableName:              aws.String("FootballLeague"),
			IndexName:              aws.String("GSI1-EntityLookup"),
			KeyConditionExpression: aws.String("GSI1_PK = :pk AND GSI1_SK = :sk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "GAME"},
				":sk": &types.AttributeValueMemberS{Value: "GAME#" + gameID},
			},
			Limit: aws.Int32(1),
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to find game")
			return
		}

		if len(gsiResult.Items) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "Game not found")
			return
		}

		// Extract league ID from the result
		var leagueID string
		if leagueIDVal, ok := gsiResult.Items[0]["league_id"]; ok {
			if leagueIDStr, ok := leagueIDVal.(*types.AttributeValueMemberS); ok {
				leagueID = leagueIDStr.Value
			}
		}

		if leagueID == "" {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid game data: missing league ID")
			return
		}

		// Now get the game using its actual key
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "LEAGUE#" + leagueID},
			"SK": &types.AttributeValueMemberS{Value: "GAME#" + gameID},
		}

		// Get the existing game
		item, err := dbClient.GetItem(r.Context(), "FootballLeague", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch game")
			return
		}

		// Check if the game exists
		if len(item) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "Game not found")
			return
		}

		// Unmarshal the existing game data
		var game models.Game
		if data, ok := item["data"].(*types.AttributeValueMemberM); ok {
			err = attributevalue.UnmarshalMap(data.Value, &game)
			if err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "Error processing game data")
				return
			}
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid game data format")
			return
		}

		// Update fields if provided
		updated := false
		if req.Status != nil && game.Status != *req.Status {
			game.Status = *req.Status
			updated = true
		}

		// Only update winner if the game is being marked as completed
		if req.Winner != nil && game.Status == "completed" && game.Winner != *req.Winner {
			game.Winner = *req.Winner
			updated = true
		}

		// If no fields were actually updated, return early
		if !updated {
			utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "No changes detected"})
			return
		}

		// Update the updated_at timestamp
		game.UpdatedAt = time.Now()

		// Marshal the updated game data
		updatedData, err := attributevalue.MarshalMap(game)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process game data")
			return
		}

		// Update the item in DynamoDB
		updateItem := map[string]types.AttributeValue{
			"PK":   &types.AttributeValueMemberS{Value: "LEAGUE#" + game.LeagueID},
			"SK":   &types.AttributeValueMemberS{Value: "GAME#" + gameID},
			"data": &types.AttributeValueMemberM{Value: updatedData},
		}

		// Add updated fields to the update expression
		updateBuilder := expression.Set(
			expression.Name("status"),
			expression.Value(game.Status),
		).Set(
			expression.Name("updated_at"),
			expression.Value(game.UpdatedAt.Format(time.RFC3339)),
		)

		expr, err := expression.NewBuilder().
			WithUpdate(updateBuilder).
			Build()
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to build update expression")
			return
		}

		// Convert expression names to DynamoDB attribute values
		exprNames := make(map[string]types.AttributeValue)
		for k, v := range expr.Names() {
			exprNames[k] = &types.AttributeValueMemberS{Value: v}
		}

		// Add the update expression to the update item
		updateItem["update_expression"] = &types.AttributeValueMemberS{Value: *expr.Update()}
		updateItem["expression_attribute_names"] = &types.AttributeValueMemberM{Value: exprNames}
		updateItem["expression_attribute_values"] = &types.AttributeValueMemberM{Value: expr.Values()}

		// Update the game in DynamoDB
		if err := dbClient.PutItem(r.Context(), "FootballLeague", updateItem); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update game: "+err.Error())
			return
		}

		// Return the updated game
		response := GameResponse{
			ID:         game.ID,
			LeagueID:   game.LeagueID,
			Week:       game.Week,
			HomeTeamID: game.HomeTeamID,
			AwayTeamID: game.AwayTeamID,
			GameDate:   game.GameDate,
			Status:     game.Status,
			CreatedAt:  game.CreatedAt,
			UpdatedAt:  game.UpdatedAt,
		}

		// Only include winner if the game is completed
		if game.Status == "completed" {
			response.Winner = game.Winner
		}

		utils.RespondWithJSON(w, http.StatusOK, response)
	}
}

// DeleteGameHandler handles deleting a game by ID
func DeleteGameHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow DELETE method
		if r.Method != http.MethodDelete {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Get game ID from query parameters
		gameID := r.URL.Query().Get("id")
		if gameID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Game ID is required")
			return
		}

		// First, get the game to find its league ID
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "GAME#" + gameID},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		}

		// Get the game to find its league ID
		item, err := dbClient.GetItem(r.Context(), "FootballLeague", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch game")
			return
		}

		// Check if the game exists
		if len(item) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "Game not found")
			return
		}

		// Extract the league ID from the item
		var leagueID string
		if leagueIDVal, ok := item["league_id"]; ok {
			if leagueIDStr, ok := leagueIDVal.(*types.AttributeValueMemberS); ok {
				leagueID = leagueIDStr.Value
			}
		}

		if leagueID == "" {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid game data: missing league ID")
			return
		}

		// Create the key for the delete operation
		deleteKey := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "LEAGUE#" + leagueID},
			"SK": &types.AttributeValueMemberS{Value: "GAME#" + gameID},
		}

		// Delete the game from DynamoDB
		if err := dbClient.DeleteItem(r.Context(), "FootballLeague", deleteKey); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete game: "+err.Error())
			return
		}

		// Return success response
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "Game deleted successfully",
		})
	}
}

// Helper function to validate game status
func isValidStatus(status string) bool {
	switch status {
	case "pending", "in_progress", "completed":
		return true
	default:
		return false
	}
}
