package handlers

import (
	"encoding/json"
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
	"github.com/gorilla/mux"
)

// LeaderboardResponse represents the response structure for leaderboard entries
type LeaderboardResponse struct {
	Entries []models.LeaderboardEntry `json:"entries"`
}

// GetAllLeaderboardEntriesHandler returns all leaderboard entries
func GetAllLeaderboardEntriesHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get all leaderboard entries
		input := &dynamodb.ScanInput{
			TableName: aws.String("LeaderboardEntries"), // Make sure this matches your DynamoDB table name
		}

		result, err := dbClient.Scan(r.Context(), input)

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch leaderboard entries")
			return
		}

		var entries []models.LeaderboardEntry
		err = attributevalue.UnmarshalListOfMaps(result.Items, &entries)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process leaderboard entries")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, LeaderboardResponse{Entries: entries})
	}
}

// GetLeaderboardEntryHandler returns a specific leaderboard entry by user ID and week
func GetLeaderboardEntryHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID := vars["userId"]
		weekStr := vars["week"]

		week, err := strconv.Atoi(weekStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid week format")
			return
		}

		key := map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
			"week":    &types.AttributeValueMemberN{Value: strconv.Itoa(week)},
		}

		result, err := dbClient.GetItem(r.Context(), "LeaderboardEntries", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch leaderboard entry")
			return
		}

		if len(result) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "Leaderboard entry not found")
			return
		}

		var entry models.LeaderboardEntry
		err = attributevalue.UnmarshalMap(result, &entry)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process leaderboard entry")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, entry)
	}
}

// GetLeaderboardEntriesByWeekHandler returns all leaderboard entries for a specific week
func GetLeaderboardEntriesByWeekHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		weekStr := vars["week"]

		week, err := strconv.Atoi(weekStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid week format")
			return
		}

		// Query the GSI for week
		input := &dynamodb.ScanInput{
			TableName:        aws.String("LeaderboardEntries"),
			FilterExpression: aws.String("#week = :week"),
			ExpressionAttributeNames: map[string]string{
				"#week": "week",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":week": &types.AttributeValueMemberN{Value: strconv.Itoa(week)},
			},
		}

		result, err := dbClient.Scan(r.Context(), input)

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch leaderboard entries")
			return
		}

		var entries []models.LeaderboardEntry
		err = attributevalue.UnmarshalListOfMaps(result.Items, &entries)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process leaderboard entries")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, LeaderboardResponse{Entries: entries})
	}
}

// CreateLeaderboardEntryRequest represents the request body for creating a leaderboard entry
type CreateLeaderboardEntryRequest struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Points    int    `json:"points"`
	Correct   int    `json:"correct"`
	Incorrect int    `json:"incorrect"`
	Week      int    `json:"week"`
}

// CreateLeaderboardEntryHandler creates a new leaderboard entry
func CreateLeaderboardEntryHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreateLeaderboardEntryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		defer r.Body.Close()

		// Validate required fields
		if req.UserID == "" || req.Username == "" || req.Week == 0 {
			utils.RespondWithError(w, http.StatusBadRequest, "Missing required fields")
			return
		}

		entry := models.LeaderboardEntry{
			UserID:    req.UserID,
			Username:  req.Username,
			Points:    req.Points,
			Correct:   req.Correct,
			Incorrect: req.Incorrect,
			Week:      req.Week,
		}

		item, err := attributevalue.MarshalMap(entry)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create leaderboard entry")
			return
		}

		err = dbClient.PutItem(r.Context(), "LeaderboardEntries", item)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to save leaderboard entry")
			return
		}

		utils.RespondWithJSON(w, http.StatusCreated, entry)
	}
}

// UpdateLeaderboardEntryRequest represents the request body for updating a leaderboard entry
type UpdateLeaderboardEntryRequest struct {
	Points    *int `json:"points,omitempty"`
	Correct   *int `json:"correct,omitempty"`
	Incorrect *int `json:"incorrect,omitempty"`
}

// UpdateLeaderboardEntryHandler updates an existing leaderboard entry
func UpdateLeaderboardEntryHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID := vars["userId"]
		weekStr := vars["week"]

		week, err := strconv.Atoi(weekStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid week format")
			return
		}

		var req UpdateLeaderboardEntryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}
		defer r.Body.Close()

		// Build update expression
		updateExpr := "SET "
		exprAttrValues := make(map[string]types.AttributeValue)
		exprAttrNames := make(map[string]string)
		first := true

		if req.Points != nil {
			updateExpr += "#p = :p"
			exprAttrValues[":p"] = &types.AttributeValueMemberN{Value: strconv.Itoa(*req.Points)}
			exprAttrNames["#p"] = "points"
			first = false
		}

		if req.Correct != nil {
			if !first {
				updateExpr += ", "
			}
			updateExpr += "#c = :c"
			exprAttrValues[":c"] = &types.AttributeValueMemberN{Value: strconv.Itoa(*req.Correct)}
			exprAttrNames["#c"] = "correct"
			first = false
		}

		if req.Incorrect != nil {
			if !first {
				updateExpr += ", "
			}
			updateExpr += "#i = :i"
			exprAttrValues[":i"] = &types.AttributeValueMemberN{Value: strconv.Itoa(*req.Incorrect)}
			exprAttrNames["#i"] = "incorrect"
		}

		key := map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
			"week":    &types.AttributeValueMemberN{Value: strconv.Itoa(week)},
		}

		// Add updated_at
		updateExpr += ", #u = :u"
		exprAttrValues[":u"] = &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)}
		exprAttrNames["#u"] = "updated_at"

		// For simplicity, we'll use PutItem instead of UpdateItem since our DatabaseClient interface doesn't have UpdateItem
		// First get the existing item
		existingItem, err := dbClient.GetItem(r.Context(), "LeaderboardEntries", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch existing leaderboard entry")
			return
		}

		// Update the fields
		if req.Points != nil {
			existingItem["points"] = &types.AttributeValueMemberN{Value: strconv.Itoa(*req.Points)}
		}
		if req.Correct != nil {
			existingItem["correct"] = &types.AttributeValueMemberN{Value: strconv.Itoa(*req.Correct)}
		}
		if req.Incorrect != nil {
			existingItem["incorrect"] = &types.AttributeValueMemberN{Value: strconv.Itoa(*req.Incorrect)}
		}
		existingItem["updated_at"] = &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)}

		// Save the updated item
		err = dbClient.PutItem(r.Context(), "LeaderboardEntries", existingItem)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update leaderboard entry")
			return
		}

		// Return the updated entry
		result, err := dbClient.GetItem(r.Context(), "LeaderboardEntries", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch updated leaderboard entry")
			return
		}

		var entry models.LeaderboardEntry
		err = attributevalue.UnmarshalMap(result, &entry)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process leaderboard entry")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, entry)
	}
}

// DeleteLeaderboardEntryHandler deletes a leaderboard entry
func DeleteLeaderboardEntryHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userID := vars["userId"]
		weekStr := vars["week"]

		week, err := strconv.Atoi(weekStr)
		if err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid week format")
			return
		}

		key := map[string]types.AttributeValue{
			"user_id": &types.AttributeValueMemberS{Value: userID},
			"week":    &types.AttributeValueMemberN{Value: strconv.Itoa(week)},
		}

		// First check if the entry exists
		result, err := dbClient.GetItem(r.Context(), "LeaderboardEntries", key)
		if err != nil || len(result) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "Leaderboard entry not found")
			return
		}

		err = dbClient.DeleteItem(r.Context(), "LeaderboardEntries", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete leaderboard entry")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Leaderboard entry deleted successfully"})
	}
}
