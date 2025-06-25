package handlers

import (
	"context"
	"errors"
	"football-picking-league/backend/db"
	"football-picking-league/backend/models"
	"football-picking-league/backend/utils"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

// PickRequest represents the request body for creating/updating a pick
type PickRequest struct {
	UserID string `json:"user_id"`
	GameID string `json:"game_id"`
	Week   int    `json:"week"`
	Pick   string `json:"pick"` // home/away
}

// PickResponse represents the response structure for pick data
type PickResponse struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	GameID    string    `json:"game_id"`
	Week      int       `json:"week"`
	Pick      string    `json:"pick"`
	Status    string    `json:"status"`
	Points    int       `json:"points"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// PicksResponse represents the response structure for multiple picks
type PicksResponse struct {
	Picks []PickResponse `json:"picks"`
}

// validatePick ensures the pick data is valid
func validatePick(pick *PickRequest) error {
	if pick.UserID == "" {
		return errors.New("user_id is required")
	}
	if pick.GameID == "" {
		return errors.New("game_id is required")
	}
	if pick.Week <= 0 {
		return errors.New("week must be greater than 0")
	}
	if pick.Pick != "home" && pick.Pick != "away" {
		return errors.New("pick must be either 'home' or 'away'")
	}
	return nil
}

// toPickResponse converts a models.Pick to a PickResponse
func toPickResponse(pick models.Pick) PickResponse {
	return PickResponse{
		ID:        pick.ID,
		UserID:    pick.UserID,
		GameID:    pick.GameID,
		Week:      pick.Week,
		Pick:      pick.Pick,
		Status:    pick.Status,
		Points:    pick.Points,
		CreatedAt: pick.CreatedAt,
		UpdatedAt: pick.UpdatedAt,
	}
}

func SubmitPickHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Parse request body
		var req PickRequest
		if err := utils.ParseJSONBody(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		// Validate pick
		if err := validatePick(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Check if pick already exists for this user and game
		existingPicks, err := getPicksByUserAndGame(dbClient, r.Context(), req.UserID, req.GameID)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check existing picks: "+err.Error())
			return
		}

		if len(existingPicks) > 0 {
			utils.RespondWithError(w, http.StatusConflict, "A pick already exists for this user and game")
			return
		}

		// Create new pick
		now := time.Now()
		pick := models.Pick{
			ID:        uuid.New().String(),
			UserID:    req.UserID,
			GameID:    req.GameID,
			Week:      req.Week,
			Pick:      req.Pick,
			Status:    "pending",
			Points:    0,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Convert to DynamoDB item
		item, err := attributevalue.MarshalMap(pick)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to marshal pick: "+err.Error())
			return
		}

		// Add the pick to DynamoDB
		if err := dbClient.PutItem(r.Context(), "Picks", item); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save pick: "+err.Error())
			return
		}

		// Return the created pick
		utils.RespondWithJSON(w, http.StatusCreated, toPickResponse(pick))
	}
}

func GetPickHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get pick ID from query parameters
		pickID := r.URL.Query().Get("id")
		if pickID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Pick ID is required")
			return
		}

		// Create the key for the query
		key := map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: pickID},
		}

		// Get the item from DynamoDB
		result, err := dbClient.GetItem(r.Context(), "Picks", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get pick: "+err.Error())
			return
		}

		if len(result) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "Pick not found")
			return
		}

		// Unmarshal the result into a Pick
		var pick models.Pick
		if err := attributevalue.UnmarshalMap(result, &pick); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process pick data: "+err.Error())
			return
		}

		// Return the response
		utils.RespondWithJSON(w, http.StatusOK, toPickResponse(pick))
	}
}

func GetAllPicksHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Scan the entire Picks table
		result, err := dbClient.Scan(r.Context(), &dynamodb.ScanInput{
			TableName: aws.String("Picks"),
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch picks: "+err.Error())
			return
		}

		// Unmarshal the results
		var picks []models.Pick
		if err := attributevalue.UnmarshalListOfMaps(result.Items, &picks); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process picks: "+err.Error())
			return
		}

		// Convert to response format
		response := PicksResponse{
			Picks: make([]PickResponse, 0, len(picks)),
		}

		for _, pick := range picks {
			response.Picks = append(response.Picks, toPickResponse(pick))
		}

		utils.RespondWithJSON(w, http.StatusOK, response)
	}
}

func GetAllPicksByUserHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get user ID from query parameters
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "User ID is required")
			return
		}

		// Build the query
		expr, err := expression.NewBuilder().
			WithKeyCondition(expression.Key("user_id").Equal(expression.Value(userID))).
			Build()
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to build query: "+err.Error())
			return
		}

		// Query the Picks table
		result, err := dbClient.QueryItems(r.Context(), &dynamodb.QueryInput{
			TableName:                 aws.String("Picks"),
			KeyConditionExpression:    expr.KeyCondition(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch picks: "+err.Error())
			return
		}

		// Unmarshal the results
		var picks []models.Pick
		if err := attributevalue.UnmarshalListOfMaps(result.Items, &picks); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process picks: "+err.Error())
			return
		}

		// Convert to response format
		response := PicksResponse{
			Picks: make([]PickResponse, 0, len(picks)),
		}

		for _, pick := range picks {
			response.Picks = append(response.Picks, toPickResponse(pick))
		}

		utils.RespondWithJSON(w, http.StatusOK, response)
	}
}

func GetAllPicksByWeekHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get week from query parameters
		week := r.URL.Query().Get("week")
		if week == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Week is required")
			return
		}

		// Build the query
		expr, err := expression.NewBuilder().
			WithFilter(expression.Name("week").Equal(expression.Value(week))).
			Build()
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to build query: "+err.Error())
			return
		}

		// Scan the Picks table with filter
		result, err := dbClient.Scan(r.Context(), &dynamodb.ScanInput{
			TableName:                 aws.String("Picks"),
			FilterExpression:          expr.Filter(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch picks: "+err.Error())
			return
		}

		// Unmarshal the results
		var picks []models.Pick
		if err := attributevalue.UnmarshalListOfMaps(result.Items, &picks); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process picks: "+err.Error())
			return
		}

		// Convert to response format
		response := PicksResponse{
			Picks: make([]PickResponse, 0, len(picks)),
		}

		for _, pick := range picks {
			response.Picks = append(response.Picks, toPickResponse(pick))
		}

		utils.RespondWithJSON(w, http.StatusOK, response)
	}
}
func UpdatePickHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Get pick ID from URL parameters
		pickID := r.URL.Query().Get("id")
		if pickID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Pick ID is required")
			return
		}

		// Parse request body
		var req PickRequest
		if err := utils.ParseJSONBody(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		// Validate pick
		if err := validatePick(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// First, get the existing pick
		key := map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: pickID},
		}

		result, err := dbClient.GetItem(r.Context(), "Picks", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get pick: "+err.Error())
			return
		}

		if len(result) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "Pick not found")
			return
		}

		// Unmarshal the existing pick
		var existingPick models.Pick
		if err := attributevalue.UnmarshalMap(result, &existingPick); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process pick data: "+err.Error())
			return
		}

		// Update the pick
		existingPick.Pick = req.Pick
		existingPick.Week = req.Week
		existingPick.UpdatedAt = time.Now()

		// Convert to DynamoDB item
		item, err := attributevalue.MarshalMap(existingPick)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to marshal pick: "+err.Error())
			return
		}

		// Update the pick in DynamoDB
		if err := dbClient.PutItem(r.Context(), "Picks", item); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update pick: "+err.Error())
			return
		}

		// Return the updated pick
		utils.RespondWithJSON(w, http.StatusOK, toPickResponse(existingPick))
	}
}

func DeletePickHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Get pick ID from query parameters
		pickID := r.URL.Query().Get("id")
		if pickID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Pick ID is required")
			return
		}

		// First check if the pick exists
		key := map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: pickID},
		}

		_, err := dbClient.GetItem(r.Context(), "Picks", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Pick not found")
			return
		}

		// Delete the pick from DynamoDB
		if err := dbClient.DeleteItem(r.Context(), "Picks", key); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete pick: "+err.Error())
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "Pick deleted successfully",
		})
	}
}

// Helper function to get picks by user and game
func getPicksByUserAndGame(dbClient db.DatabaseClient, ctx context.Context, userID, gameID string) ([]models.Pick, error) {
	// Build the query
	expr, err := expression.NewBuilder().
		WithFilter(
			expression.Name("user_id").Equal(expression.Value(userID)).
				And(expression.Name("game_id").Equal(expression.Value(gameID))),
		).
		Build()

	if err != nil {
		return nil, err
	}

	// Scan the Picks table with filter
	result, err := dbClient.Scan(ctx, &dynamodb.ScanInput{
		TableName:                 aws.String("Picks"),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})

	if err != nil {
		return nil, err
	}

	// Unmarshal the results
	var picks []models.Pick
	if err := attributevalue.UnmarshalListOfMaps(result.Items, &picks); err != nil {
		return nil, err
	}

	return picks, nil
}
