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
		// Use GSI1 to efficiently get all leaderboard entries
		result, err := dbClient.QueryItems(r.Context(), &dynamodb.QueryInput{
			TableName:              aws.String("FootballLeague"),
			IndexName:              aws.String("GSI1-EntityLookup"),
			KeyConditionExpression: aws.String("GSI1_PK = :pk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "LEADERBOARD_ENTRY"},
			},
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch leaderboard entries")
			return
		}

		// Process the results
		var entries []models.LeaderboardEntry
		for _, item := range result.Items {
			if data, ok := item["data"].(*types.AttributeValueMemberM); ok {
				var entry models.LeaderboardEntry
				err = attributevalue.UnmarshalMap(data.Value, &entry)
				if err == nil {
					entries = append(entries, entry)
				}
			}
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

		// Use efficient key lookup with single table design
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "LEADERBOARD#WEEK#" + strconv.Itoa(week)},
		}

		result, err := dbClient.GetItem(r.Context(), "FootballLeague", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch leaderboard entry")
			return
		}

		if len(result) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "Leaderboard entry not found")
			return
		}

		// Unmarshal the leaderboard entry data
		var entry models.LeaderboardEntry
		if data, ok := result["data"].(*types.AttributeValueMemberM); ok {
			err = attributevalue.UnmarshalMap(data.Value, &entry)
			if err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process leaderboard entry")
				return
			}
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid leaderboard entry data format")
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

		// Use GSI2 to efficiently query leaderboard entries by week
		result, err := dbClient.QueryItems(r.Context(), &dynamodb.QueryInput{
			TableName:              aws.String("FootballLeague"),
			IndexName:              aws.String("GSI2-UserPicks"),
			KeyConditionExpression: aws.String("GSI2_PK = :pk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "LEADERBOARD_WEEK#" + strconv.Itoa(week)},
			},
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch leaderboard entries")
			return
		}

		// Process the results
		var entries []models.LeaderboardEntry
		for _, item := range result.Items {
			if data, ok := item["data"].(*types.AttributeValueMemberM); ok {
				var entry models.LeaderboardEntry
				err = attributevalue.UnmarshalMap(data.Value, &entry)
				if err == nil {
					entries = append(entries, entry)
				}
			}
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

		now := time.Now()
		entry := models.LeaderboardEntry{
			UserID:    req.UserID,
			Username:  req.Username,
			Points:    req.Points,
			Correct:   req.Correct,
			Incorrect: req.Incorrect,
			Week:      req.Week,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Marshal the entry data
		data, err := attributevalue.MarshalMap(entry)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process entry data")
			return
		}

		// Create the item using single table design
		item := map[string]types.AttributeValue{
			"PK":          &types.AttributeValueMemberS{Value: "USER#" + req.UserID},
			"SK":          &types.AttributeValueMemberS{Value: "LEADERBOARD#WEEK#" + strconv.Itoa(req.Week)},
			"GSI1_PK":     &types.AttributeValueMemberS{Value: "LEADERBOARD_ENTRY"},
			"GSI1_SK":     &types.AttributeValueMemberS{Value: "USER#" + req.UserID + "#WEEK#" + strconv.Itoa(req.Week)},
			"GSI2_PK":     &types.AttributeValueMemberS{Value: "LEADERBOARD_WEEK#" + strconv.Itoa(req.Week)},
			"GSI2_SK":     &types.AttributeValueMemberS{Value: "USER#" + req.UserID},
			"entity_type": &types.AttributeValueMemberS{Value: "LEADERBOARD_ENTRY"},
			"user_id":     &types.AttributeValueMemberS{Value: req.UserID},
			"username":    &types.AttributeValueMemberS{Value: req.Username},
			"week":        &types.AttributeValueMemberN{Value: strconv.Itoa(req.Week)},
			"points":      &types.AttributeValueMemberN{Value: strconv.Itoa(req.Points)},
			"correct":     &types.AttributeValueMemberN{Value: strconv.Itoa(req.Correct)},
			"incorrect":   &types.AttributeValueMemberN{Value: strconv.Itoa(req.Incorrect)},
			"created_at":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			"updated_at":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			"data":        &types.AttributeValueMemberM{Value: data},
		}

		err = dbClient.PutItem(r.Context(), "FootballLeague", item)
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

		// Use efficient key lookup with single table design
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "LEADERBOARD#WEEK#" + strconv.Itoa(week)},
		}

		// Get the existing item
		existingItem, err := dbClient.GetItem(r.Context(), "FootballLeague", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch existing leaderboard entry")
			return
		}

		if len(existingItem) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "Leaderboard entry not found")
			return
		}

		// Unmarshal existing data to get current values
		var existingEntry models.LeaderboardEntry
		if data, ok := existingItem["data"].(*types.AttributeValueMemberM); ok {
			err = attributevalue.UnmarshalMap(data.Value, &existingEntry)
			if err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process existing entry")
				return
			}
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid entry data format")
			return
		}

		// Update the fields if provided
		if req.Points != nil {
			existingEntry.Points = *req.Points
		}
		if req.Correct != nil {
			existingEntry.Correct = *req.Correct
		}
		if req.Incorrect != nil {
			existingEntry.Incorrect = *req.Incorrect
		}
		existingEntry.UpdatedAt = time.Now()

		// Marshal the updated entry data
		data, err := attributevalue.MarshalMap(existingEntry)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process updated entry data")
			return
		}

		// Update the item
		updatedItem := map[string]types.AttributeValue{
			"PK":          key["PK"],
			"SK":          key["SK"],
			"GSI1_PK":     &types.AttributeValueMemberS{Value: "LEADERBOARD_ENTRY"},
			"GSI1_SK":     &types.AttributeValueMemberS{Value: "USER#" + userID + "#WEEK#" + strconv.Itoa(week)},
			"GSI2_PK":     &types.AttributeValueMemberS{Value: "LEADERBOARD_WEEK#" + strconv.Itoa(week)},
			"GSI2_SK":     &types.AttributeValueMemberS{Value: "USER#" + userID},
			"entity_type": &types.AttributeValueMemberS{Value: "LEADERBOARD_ENTRY"},
			"user_id":     &types.AttributeValueMemberS{Value: userID},
			"username":    existingItem["username"], // Preserve username
			"week":        &types.AttributeValueMemberN{Value: strconv.Itoa(week)},
			"points":      &types.AttributeValueMemberN{Value: strconv.Itoa(existingEntry.Points)},
			"correct":     &types.AttributeValueMemberN{Value: strconv.Itoa(existingEntry.Correct)},
			"incorrect":   &types.AttributeValueMemberN{Value: strconv.Itoa(existingEntry.Incorrect)},
			"created_at":  existingItem["created_at"], // Preserve created_at
			"updated_at":  &types.AttributeValueMemberS{Value: existingEntry.UpdatedAt.Format(time.RFC3339)},
			"data":        &types.AttributeValueMemberM{Value: data},
		}

		err = dbClient.PutItem(r.Context(), "FootballLeague", updatedItem)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update leaderboard entry")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, existingEntry)
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

		// Use efficient key with single table design
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "LEADERBOARD#WEEK#" + strconv.Itoa(week)},
		}

		// First check if the entry exists
		result, err := dbClient.GetItem(r.Context(), "FootballLeague", key)
		if err != nil || len(result) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "Leaderboard entry not found")
			return
		}

		err = dbClient.DeleteItem(r.Context(), "FootballLeague", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete leaderboard entry")
			return
		}

		utils.RespondWithJSON(w, http.StatusOK, map[string]string{"message": "Leaderboard entry deleted successfully"})
	}
}
