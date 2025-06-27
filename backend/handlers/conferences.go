package handlers

import (
	"encoding/json"
	"football-picking-league/backend/db"
	"football-picking-league/backend/models"
	"football-picking-league/backend/utils"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gorilla/mux"
)

// ConferenceResponse represents the response structure for conference data
type ConferenceResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateConferenceRequest represents the request body for creating a conference
type CreateConferenceRequest struct {
	Name string `json:"name"`
}

// ConferencesResponse represents the response structure for multiple conferences
type ConferencesResponse struct {
	Conferences []ConferenceResponse `json:"conferences"`
}

// GetConferenceHandler returns a single conference by ID
func GetConferenceHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the ID from the URL parameters
		id := r.URL.Query().Get("id")
		if id == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Conference ID is required")
			return
		}

		// Create the key for the query
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "CONFERENCE#" + id},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		}

		// Get the item from DynamoDB
		item, err := dbClient.GetItem(r.Context(), "FootballLeague", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch conference")
			return
		}

		// Check if the item exists
		if len(item) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "Conference not found")
			return
		}

		// Unmarshal the conference data
		var conference models.Conference
		if data, ok := item["data"].(*types.AttributeValueMemberM); ok {
			err = attributevalue.UnmarshalMap(data.Value, &conference)
			if err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "Error processing conference data")
				return
			}
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid conference data format")
			return
		}

		// Return the response
		utils.RespondWithJSON(w, http.StatusOK, ConferenceResponse{
			ID:        conference.ID,
			Name:      conference.Name,
			CreatedAt: conference.CreatedAt,
			UpdatedAt: conference.UpdatedAt,
		})
	}
}

// GetAllConferencesHandler returns a list of all conferences
func GetAllConferencesHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Query the table using GSI1-EntityLookup to get all conferences
		result, err := dbClient.QueryItems(r.Context(), &dynamodb.QueryInput{
			TableName:              aws.String("FootballLeague"),
			IndexName:              aws.String("GSI1-EntityLookup"),
			KeyConditionExpression: aws.String("GSI1_PK = :pk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "CONFERENCE"},
			},
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch conferences")
			return
		}

		// Process the results
		var conferences []models.Conference
		for _, item := range result.Items {
			if data, ok := item["data"].(*types.AttributeValueMemberM); ok {
				var conf models.Conference
				err = attributevalue.UnmarshalMap(data.Value, &conf)
				if err == nil {
					conferences = append(conferences, conf)
				}
			}
		}

		// Convert to response format
		response := ConferencesResponse{
			Conferences: make([]ConferenceResponse, 0, len(conferences)),
		}

		for _, conf := range conferences {
			response.Conferences = append(response.Conferences, ConferenceResponse{
				ID:        conf.ID,
				Name:      conf.Name,
				CreatedAt: conf.CreatedAt,
				UpdatedAt: conf.UpdatedAt,
			})
		}

		utils.RespondWithJSON(w, http.StatusOK, response)
	}
}

// CreateConferenceHandler handles creating a new conference
func CreateConferenceHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST method
		if r.Method != http.MethodPost {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Parse request body
		var req CreateConferenceRequest
		if err := utils.ParseJSONBody(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		// Validate required fields
		if strings.TrimSpace(req.Name) == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Conference name is required")
			return
		}

		// Generate a new UUID for the conference
		conferenceID := strings.ReplaceAll(strings.ToUpper(req.Name), " ", "_")

		now := time.Now()
		conference := models.Conference{
			ID:        conferenceID,
			Name:      req.Name,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Marshal the conference data
		data, err := attributevalue.MarshalMap(conference)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to process conference data")
			return
		}

		// Create the item to save
		item := map[string]types.AttributeValue{
			"PK":          &types.AttributeValueMemberS{Value: "CONFERENCE#" + conferenceID},
			"SK":          &types.AttributeValueMemberS{Value: "METADATA"},
			"GSI1_PK":     &types.AttributeValueMemberS{Value: "CONFERENCE"},
			"GSI1_SK":     &types.AttributeValueMemberS{Value: "CONFERENCE#" + conferenceID},
			"entity_type": &types.AttributeValueMemberS{Value: "CONFERENCE"},
			"id":          &types.AttributeValueMemberS{Value: conferenceID},
			"created_at":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			"updated_at":  &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
			"data":        &types.AttributeValueMemberM{Value: data},
		}

		// Save the conference to DynamoDB
		if err := dbClient.PutItem(r.Context(), "FootballLeague", item); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create conference: "+err.Error())
			return
		}

		// Return the created conference
		utils.RespondWithJSON(w, http.StatusCreated, ConferenceResponse{
			ID:        conference.ID,
			Name:      conference.Name,
			CreatedAt: conference.CreatedAt,
			UpdatedAt: conference.UpdatedAt,
		})
	}
}

// UpdateConferenceHandler handles updating a conference
func UpdateConferenceHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the ID from the URL parameters
		vars := mux.Vars(r)
		id := vars["id"]
		if id == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Conference ID is required")
			return
		}

		// Get the existing conference first
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "CONFERENCE#" + id},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		}

		existingItem, err := dbClient.GetItem(r.Context(), "FootballLeague", key)
		if err != nil {
			utils.RespondWithError(w, http.StatusNotFound, "Conference not found")
			return
		}

		// Parse the request body
		var req CreateConferenceRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		// Validate required fields
		if strings.TrimSpace(req.Name) == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Conference name is required")
			return
		}

		// Unmarshal the existing data
		var existingData map[string]types.AttributeValue
		if data, ok := existingItem["data"].(*types.AttributeValueMemberM); ok {
			existingData = data.Value
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid conference data")
			return
		}

		// Update the name in the data
		existingData["name"] = &types.AttributeValueMemberS{Value: req.Name}

		// Prepare the updated item
		updatedAt := time.Now().Format(time.RFC3339)
		item := map[string]types.AttributeValue{
			"PK":          key["PK"],
			"SK":          key["SK"],
			"entity_type": &types.AttributeValueMemberS{Value: "CONFERENCE"},
			"id":          &types.AttributeValueMemberS{Value: id},
			"created_at":  existingItem["created_at"],
			"updated_at":  &types.AttributeValueMemberS{Value: updatedAt},
			"data":        &types.AttributeValueMemberM{Value: existingData},
		}

		// Save the updated conference to DynamoDB
		if err := dbClient.PutItem(r.Context(), "FootballLeague", item); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update conference: "+err.Error())
			return
		}

		// Get the created_at time from the existing item
		var createdAt string
		if created, ok := existingItem["created_at"].(*types.AttributeValueMemberS); ok {
			createdAt = created.Value
		}

		// Parse the created_at time
		parsedCreatedAt, _ := time.Parse(time.RFC3339, createdAt)

		// Return the updated conference
		utils.RespondWithJSON(w, http.StatusOK, ConferenceResponse{
			ID:        id,
			Name:      req.Name,
			CreatedAt: parsedCreatedAt,
			UpdatedAt: time.Now(),
		})
	}
}

// DeleteConferenceHandler handles deleting a conference by ID
func DeleteConferenceHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow DELETE method
		if r.Method != http.MethodDelete {
			utils.RespondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
			return
		}

		// Get conference ID from query parameters
		conferenceID := r.URL.Query().Get("id")
		if conferenceID == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Conference ID is required")
			return
		}

		// Create the key for the conference item
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "CONFERENCE#" + conferenceID},
			"SK": &types.AttributeValueMemberS{Value: "METADATA"},
		}

		// First check if the conference exists
		result, err := dbClient.GetItem(r.Context(), "FootballLeague", key)
		if err != nil || len(result) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "Conference not found")
			return
		}

		// Delete the conference from DynamoDB
		if err := dbClient.DeleteItem(r.Context(), "FootballLeague", key); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete conference: "+err.Error())
			return
		}

		// Return success response
		utils.RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "Conference deleted successfully",
		})
	}
}
