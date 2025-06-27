package handlers

import (
	"context"
	"errors"
	"football-picking-league/backend/db"
	"football-picking-league/backend/models"
	"football-picking-league/backend/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
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

		// Check if pick already exists for this user and game using efficient GSI query
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + req.UserID},
			"SK": &types.AttributeValueMemberS{Value: "PICK#" + req.GameID},
		}

		existingItem, err := dbClient.GetItem(r.Context(), "FootballLeague", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to check existing picks: "+err.Error())
			return
		}

		if len(existingItem) > 0 {
			utils.RespondWithError(w, http.StatusConflict, "A pick already exists for this user and game")
			return
		}

		// Create new pick
		now := time.Now()
		pickID := uuid.New().String()
		pick := models.Pick{
			ID:        pickID,
			UserID:    req.UserID,
			GameID:    req.GameID,
			Week:      req.Week,
			Pick:      req.Pick,
			Status:    "pending",
			Points:    0,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Marshal the pick data
		data, err := attributevalue.MarshalMap(pick)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process pick data")
			return
		}

		// Create the item to save using single table design
		item := map[string]types.AttributeValue{
			"PK":          &types.AttributeValueMemberS{Value: "USER#" + req.UserID},
			"SK":          &types.AttributeValueMemberS{Value: "PICK#" + req.GameID},
			"GSI1_PK":     &types.AttributeValueMemberS{Value: "PICK"},
			"GSI1_SK":     &types.AttributeValueMemberS{Value: "PICK#" + pickID},
			"GSI2_PK":     &types.AttributeValueMemberS{Value: "WEEK#" + strconv.Itoa(req.Week)},
			"GSI2_SK":     &types.AttributeValueMemberS{Value: "USER#" + req.UserID},
			"entity_type": &types.AttributeValueMemberS{Value: "PICK"},
			"id":          &types.AttributeValueMemberS{Value: pickID},
			"user_id":     &types.AttributeValueMemberS{Value: req.UserID},
			"game_id":     &types.AttributeValueMemberS{Value: req.GameID},
			"week":        &types.AttributeValueMemberN{Value: strconv.Itoa(req.Week)},
			"status":      &types.AttributeValueMemberS{Value: "pending"},
			"points":      &types.AttributeValueMemberN{Value: "0"},
			"created_at":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			"updated_at":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			"data":        &types.AttributeValueMemberM{Value: data},
		}

		// Add the pick to DynamoDB
		if err := dbClient.PutItem(r.Context(), "FootballLeague", item); err != nil {
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

		// Use GSI1 to find pick by ID efficiently
		result, err := dbClient.QueryItems(r.Context(), &dynamodb.QueryInput{
			TableName:              aws.String("FootballLeague"),
			IndexName:              aws.String("GSI1-EntityLookup"),
			KeyConditionExpression: aws.String("GSI1_PK = :pk AND GSI1_SK = :sk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "PICK"},
				":sk": &types.AttributeValueMemberS{Value: "PICK#" + pickID},
			},
			Limit: aws.Int32(1),
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to get pick: "+err.Error())
			return
		}

		if len(result.Items) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "Pick not found")
			return
		}

		// Unmarshal the result into a Pick
		var pick models.Pick
		if data, ok := result.Items[0]["data"].(*types.AttributeValueMemberM); ok {
			err = attributevalue.UnmarshalMap(data.Value, &pick)
			if err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process pick data: "+err.Error())
				return
			}
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid pick data format")
			return
		}

		// Return the response
		utils.RespondWithJSON(w, http.StatusOK, toPickResponse(pick))
	}
}

func GetAllPicksHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Use GSI1 to efficiently get all picks
		result, err := dbClient.QueryItems(r.Context(), &dynamodb.QueryInput{
			TableName:              aws.String("FootballLeague"),
			IndexName:              aws.String("GSI1-EntityLookup"),
			KeyConditionExpression: aws.String("GSI1_PK = :pk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "PICK"},
			},
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch picks: "+err.Error())
			return
		}

		// Process the results
		var picks []models.Pick
		for _, item := range result.Items {
			if data, ok := item["data"].(*types.AttributeValueMemberM); ok {
				var pick models.Pick
				err = attributevalue.UnmarshalMap(data.Value, &pick)
				if err == nil {
					picks = append(picks, pick)
				}
			}
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

		// Query using main table for user's picks efficiently
		result, err := dbClient.QueryItems(r.Context(), &dynamodb.QueryInput{
			TableName:              aws.String("FootballLeague"),
			KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "USER#" + userID},
				":sk": &types.AttributeValueMemberS{Value: "PICK#"},
			},
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch picks: "+err.Error())
			return
		}

		// Process the results
		var picks []models.Pick
		for _, item := range result.Items {
			if data, ok := item["data"].(*types.AttributeValueMemberM); ok {
				var pick models.Pick
				err = attributevalue.UnmarshalMap(data.Value, &pick)
				if err == nil {
					picks = append(picks, pick)
				}
			}
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

		// Use GSI2 to efficiently query picks by week
		result, err := dbClient.QueryItems(r.Context(), &dynamodb.QueryInput{
			TableName:              aws.String("FootballLeague"),
			IndexName:              aws.String("GSI2-UserPicks"),
			KeyConditionExpression: aws.String("GSI2_PK = :pk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "WEEK#" + week},
			},
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch picks: "+err.Error())
			return
		}

		// Process the results
		var picks []models.Pick
		for _, item := range result.Items {
			if data, ok := item["data"].(*types.AttributeValueMemberM); ok {
				var pick models.Pick
				err = attributevalue.UnmarshalMap(data.Value, &pick)
				if err == nil {
					picks = append(picks, pick)
				}
			}
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

		// Get user ID and game ID from URL parameters
		userID := r.URL.Query().Get("user_id")
		gameID := r.URL.Query().Get("game_id")
		if userID == "" || gameID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "User ID and Game ID are required")
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

		// Get the existing pick using efficient key
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "PICK#" + gameID},
		}

		result, err := dbClient.GetItem(r.Context(), "FootballLeague", key)
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
		if data, ok := result["data"].(*types.AttributeValueMemberM); ok {
			err = attributevalue.UnmarshalMap(data.Value, &existingPick)
			if err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process pick data: "+err.Error())
				return
			}
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid pick data format")
			return
		}

		// Update the pick
		existingPick.Pick = req.Pick
		existingPick.Week = req.Week
		existingPick.UpdatedAt = time.Now()

		// Marshal the updated pick data
		data, err := attributevalue.MarshalMap(existingPick)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process pick data")
			return
		}

		// Update the item in DynamoDB
		item := map[string]types.AttributeValue{
			"PK":          key["PK"],
			"SK":          key["SK"],
			"GSI1_PK":     &types.AttributeValueMemberS{Value: "PICK"},
			"GSI1_SK":     &types.AttributeValueMemberS{Value: "PICK#" + existingPick.ID},
			"GSI2_PK":     &types.AttributeValueMemberS{Value: "WEEK#" + strconv.Itoa(req.Week)},
			"GSI2_SK":     &types.AttributeValueMemberS{Value: "USER#" + userID},
			"entity_type": &types.AttributeValueMemberS{Value: "PICK"},
			"id":          &types.AttributeValueMemberS{Value: existingPick.ID},
			"user_id":     &types.AttributeValueMemberS{Value: userID},
			"game_id":     &types.AttributeValueMemberS{Value: gameID},
			"week":        &types.AttributeValueMemberN{Value: strconv.Itoa(req.Week)},
			"status":      &types.AttributeValueMemberS{Value: existingPick.Status},
			"points":      &types.AttributeValueMemberN{Value: strconv.Itoa(existingPick.Points)},
			"created_at":  result["created_at"], // Preserve created_at
			"updated_at":  &types.AttributeValueMemberS{Value: existingPick.UpdatedAt.Format(time.RFC3339)},
			"data":        &types.AttributeValueMemberM{Value: data},
		}

		if err := dbClient.PutItem(r.Context(), "FootballLeague", item); err != nil {
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

		// Get user ID and game ID from query parameters
		userID := r.URL.Query().Get("user_id")
		gameID := r.URL.Query().Get("game_id")
		if userID == "" || gameID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "User ID and Game ID are required")
			return
		}

		// Create the key for the pick
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "PICK#" + gameID},
		}

		// First check if the pick exists
		_, err := dbClient.GetItem(r.Context(), "FootballLeague", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Pick not found")
			return
		}

		// Delete the pick from DynamoDB
		if err := dbClient.DeleteItem(r.Context(), "FootballLeague", key); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete pick: "+err.Error())
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "Pick deleted successfully",
		})
	}
}

// Helper function to get picks by user and game - now optimized for single table
func getPicksByUserAndGame(dbClient db.DatabaseClient, ctx context.Context, userID, gameID string) ([]models.Pick, error) {
	// Use efficient direct key lookup
	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
		"SK": &types.AttributeValueMemberS{Value: "PICK#" + gameID},
	}

	result, err := dbClient.GetItem(ctx, "FootballLeague", key)
	if err != nil {
		return nil, err
	}

	var picks []models.Pick
	if len(result) > 0 {
		if data, ok := result["data"].(*types.AttributeValueMemberM); ok {
			var pick models.Pick
			err = attributevalue.UnmarshalMap(data.Value, &pick)
			if err == nil {
				picks = append(picks, pick)
			}
		}
	}

	return picks, nil
}
