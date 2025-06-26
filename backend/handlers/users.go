package handlers

import (
	"context"
	"encoding/json"
	"football-picking-league/backend/db"
	"football-picking-league/backend/models"
	"football-picking-league/backend/utils"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

const (
	tableName    = "FootballLeague"
	entityTypePK = "USER"
)

// UserResponse represents the response structure for user data
type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// UsersResponse represents the response structure for multiple users
type UsersResponse struct {
	Users []UserResponse `json:"users"`
}

// GetAllUsersHandler returns a list of all users
func GetAllUsersHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Query the table using GSI1 to get all users
		result, err := dbClient.QueryItems(r.Context(), &dynamodb.QueryInput{
			TableName: aws.String(tableName),
			IndexName: aws.String("GSI1-EntityLookup"),
			KeyConditionExpression: aws.String("GSI1_PK = :pk"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":pk": &types.AttributeValueMemberS{Value: "USER"},
			},
		})

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error retrieving users")
			return
		}

		// Unmarshal the results into User models
		var users []models.User
		for _, item := range result.Items {
			if data, ok := item["data"].(*types.AttributeValueMemberM); ok {
				var user models.User
				err = attributevalue.UnmarshalMap(data.Value, &user)
				if err != nil {
					utils.RespondWithError(w, http.StatusInternalServerError, "Error processing user data")
					return
				}
				users = append(users, user)
			}
		}

		// Convert to response format
		response := UsersResponse{
			Users: make([]UserResponse, 0, len(users)),
		}

		for _, user := range users {
			response.Users = append(response.Users, UserResponse{
				ID:       user.ID,
				Username: user.Username,
				Email:    user.Email,
				Role:     user.Role,
			})
		}

		utils.RespondWithJSON(w, http.StatusOK, response)
	}
}

// GetUserByIDHandler returns a single user by ID
func GetUserByIDHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the ID from the URL parameters
		id := r.URL.Query().Get("id")
		if id == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "User ID is required")
			return
		}

		// Create the key for the query
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + id},
			"SK": &types.AttributeValueMemberS{Value: "PROFILE"},
		}

		// Get the item from DynamoDB
		item, err := dbClient.GetItem(r.Context(), tableName, key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error retrieving user")
			return
		}

		// If no user found
		if len(item) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}

		// Unmarshal the item into a User model
		var user models.User
		if data, ok := item["data"].(*types.AttributeValueMemberM); ok {
			err = attributevalue.UnmarshalMap(data.Value, &user)
			if err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "Error processing user data")
				return
			}
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid user data format")
			return
		}

		// Convert to response format and return
		utils.RespondWithJSON(w, http.StatusOK, UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		})
	}
}

// GetUserByUsernameHandler returns a single user by username
func GetUserByUsernameHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the username from the query parameters
		username := r.URL.Query().Get("username")
		if username == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Username is required")
			return
		}

		// First, look up the user ID using the username lookup table
		lookupKey := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USERNAME#" + username},
			"SK": &types.AttributeValueMemberS{Value: "LOOKUP"},
		}

		lookupItem, err := dbClient.GetItem(r.Context(), tableName, lookupKey)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error looking up username")
			return
		}

		if len(lookupItem) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}

		// Get the user ID from the lookup result
		userIDAttr, ok := lookupItem["user_id"].(*types.AttributeValueMemberS)
		if !ok {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid lookup data")
			return
		}
		userID := userIDAttr.Value

		// Now get the actual user data
		userKey := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "PROFILE"},
		}

		userItem, err := dbClient.GetItem(r.Context(), tableName, userKey)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error retrieving user")
			return
		}

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error retrieving user")
			return
		}

		// If no user found
		if len(userItem) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}

		// Unmarshal the user data
		var user models.User
		if data, ok := userItem["data"].(*types.AttributeValueMemberM); ok {
			err = attributevalue.UnmarshalMap(data.Value, &user)
			if err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "Error processing user data")
				return
			}
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid user data format")
			return
		}

		// Return the response
		utils.RespondWithJSON(w, http.StatusOK, UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		})
	}
}

// GetUserByEmailHandler returns a single user by email
func GetUserByEmailHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the email from the query parameters
		email := r.URL.Query().Get("email")
		if email == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Email is required")
			return
		}

		// First, look up the user ID using the email lookup table
		lookupKey := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "EMAIL#" + strings.ToLower(email)},
			"SK": &types.AttributeValueMemberS{Value: "LOOKUP"},
		}

		lookupItem, err := dbClient.GetItem(r.Context(), tableName, lookupKey)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error looking up email")
			return
		}

		if len(lookupItem) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}

		// Get the user ID from the lookup result
		userIDAttr, ok := lookupItem["user_id"].(*types.AttributeValueMemberS)
		if !ok {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid lookup data")
			return
		}
		userID := userIDAttr.Value

		// Now get the actual user data
		userKey := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + userID},
			"SK": &types.AttributeValueMemberS{Value: "PROFILE"},
		}

		userItem, err := dbClient.GetItem(r.Context(), tableName, userKey)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error retrieving user")
			return
		}

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error retrieving user")
			return
		}

		// If no user found
		if len(userItem) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}

		// Unmarshal the user data
		var user models.User
		if data, ok := userItem["data"].(*types.AttributeValueMemberM); ok {
			err = attributevalue.UnmarshalMap(data.Value, &user)
			if err != nil {
				utils.RespondWithError(w, http.StatusInternalServerError, "Error processing user data")
				return
			}
		} else {
			utils.RespondWithError(w, http.StatusInternalServerError, "Invalid user data format")
			return
		}

		// Return the response
		utils.RespondWithJSON(w, http.StatusOK, UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		})
	}
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// parseJSONBody parses the JSON body of a request into the target interface
func parseJSONBody(r *http.Request, target interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if err := json.Unmarshal(body, target); err != nil {
		return err
	}

	return nil
}

// CreateUserHandler handles creating a new user
func CreateUserHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var req CreateUserRequest
		if err := parseJSONBody(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		// Validate required fields
		if req.Username == "" || req.Email == "" || req.Role == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "Username, email, and role are required")
			return
		}

		// Check if user with email already exists
		existingUser, _ := findUserByEmail(dbClient, req.Email)
		if existingUser != nil {
			utils.RespondWithError(w, http.StatusConflict, "User with this email already exists")
			return
		}

		// Create new user
		now := time.Now()
		user := models.User{
			ID:        uuid.New().String(),
			Username:  req.Username,
			Email:     strings.ToLower(req.Email),
			Role:      req.Role,
			CreatedAt: now,
			UpdatedAt: now,
		}

		// Prepare the item for DynamoDB
		item := map[string]types.AttributeValue{
			"PK":           &types.AttributeValueMemberS{Value: "USER#" + user.ID},
			"SK":           &types.AttributeValueMemberS{Value: "PROFILE"},
			"GSI1_PK":      &types.AttributeValueMemberS{Value: "USER"},
			"GSI1_SK":      &types.AttributeValueMemberS{Value: "USER#" + user.ID},
			"entity_type":  &types.AttributeValueMemberS{Value: entityTypePK},
			"created_at":   &types.AttributeValueMemberS{Value: user.CreatedAt.Format(time.RFC3339)},
			"updated_at":   &types.AttributeValueMemberS{Value: user.UpdatedAt.Format(time.RFC3339)},
			"data": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"username": &types.AttributeValueMemberS{Value: user.Username},
					"email":    &types.AttributeValueMemberS{Value: user.Email},
					"role":     &types.AttributeValueMemberS{Value: user.Role},
				},
			},
		}

		// Put main user item in DynamoDB
		err := dbClient.PutItem(r.Context(), tableName, item)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create user")
			return
		}

		// Create username lookup record
		usernameItem := map[string]types.AttributeValue{
			"PK":      &types.AttributeValueMemberS{Value: "USERNAME#" + user.Username},
			"SK":      &types.AttributeValueMemberS{Value: "LOOKUP"},
			"user_id": &types.AttributeValueMemberS{Value: user.ID},
		}
		err = dbClient.PutItem(r.Context(), tableName, usernameItem)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create username lookup")
			return
		}

		// Create email lookup record
		emailItem := map[string]types.AttributeValue{
			"PK":      &types.AttributeValueMemberS{Value: "EMAIL#" + user.Email},
			"SK":      &types.AttributeValueMemberS{Value: "LOOKUP"},
			"user_id": &types.AttributeValueMemberS{Value: user.ID},
		}
		err = dbClient.PutItem(r.Context(), tableName, emailItem)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to create email lookup")
			return
		}

		// Return the created user
		utils.RespondWithJSON(w, http.StatusCreated, UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
			Role:     user.Role,
		})
	}
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Username *string `json:"username,omitempty"`
	Email    *string `json:"email,omitempty"`
	Role     *string `json:"role,omitempty"`
}

// UpdateUserHandler handles updating a user by ID
func UpdateUserHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the ID from the URL parameters
		id := r.URL.Query().Get("id")
		if id == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "User ID is required")
			return
		}

		// Parse the request body
		var req UpdateUserRequest
		if err := parseJSONBody(r, &req); err != nil {
			utils.RespondWithError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		// Check if at least one field is being updated
		if req.Username == nil && req.Email == nil && req.Role == nil {
			utils.RespondWithError(w, http.StatusBadRequest, "No fields to update")
			return
		}

		// Get the existing user first
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + id},
			"SK": &types.AttributeValueMemberS{Value: "PROFILE"},
		}

		existingItem, err := dbClient.GetItem(r.Context(), tableName, key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to fetch user")
			return
		}

		if len(existingItem) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}

		// Get existing data
		var existingData map[string]types.AttributeValue
		if data, ok := existingItem["data"].(*types.AttributeValueMemberM); ok {
			existingData = data.Value
		} else {
			existingData = make(map[string]types.AttributeValue)
		}

		// Check for email uniqueness if email is being updated
		if req.Email != nil {
			existingUser, _ := findUserByEmail(dbClient, *req.Email)
			if existingUser != nil && existingUser.ID != id {
				utils.RespondWithError(w, http.StatusConflict, "Email already in use by another user")
				return
			}
		}

		// Update the fields if they are provided in the request
		if req.Username != nil {
			existingData["username"] = &types.AttributeValueMemberS{Value: *req.Username}
		}
		if req.Email != nil {
			existingData["email"] = &types.AttributeValueMemberS{Value: strings.ToLower(*req.Email)}
		}
		if req.Role != nil {
			existingData["role"] = &types.AttributeValueMemberS{Value: *req.Role}
		}

		// Prepare the updated item
		updatedAt := time.Now()
		item := map[string]types.AttributeValue{
			"PK":           key["PK"],
			"SK":           key["SK"],
			"GSI1_PK":      &types.AttributeValueMemberS{Value: "USER"},
			"GSI1_SK":      &types.AttributeValueMemberS{Value: "USER#" + id},
			"entity_type":  &types.AttributeValueMemberS{Value: entityTypePK},
			"id":           &types.AttributeValueMemberS{Value: id},
			"updated_at":   &types.AttributeValueMemberS{Value: updatedAt.Format(time.RFC3339)},
			"created_at":   existingItem["created_at"], // Preserve created_at
			"data":         &types.AttributeValueMemberM{Value: existingData},
		}


		// Save the updated user to DynamoDB
		if err := dbClient.PutItem(r.Context(), tableName, item); err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to update user: "+err.Error())
			return
		}

		// Get the updated values for the response
		updatedUser := models.User{
			ID: id,
		}

		// Set username (new or existing)
		if req.Username != nil {
			updatedUser.Username = *req.Username
		} else if username, ok := existingData["username"].(*types.AttributeValueMemberS); ok {
			updatedUser.Username = username.Value
		}

		// Set email (new or existing)
		if req.Email != nil {
			updatedUser.Email = strings.ToLower(*req.Email)
		} else if email, ok := existingData["email"].(*types.AttributeValueMemberS); ok {
			updatedUser.Email = email.Value
		}

		// Set role (new or existing)
		if req.Role != nil {
			updatedUser.Role = *req.Role
		} else if role, ok := existingData["role"].(*types.AttributeValueMemberS); ok {
			updatedUser.Role = role.Value
		}

		// Set timestamps
		updatedUser.UpdatedAt = updatedAt
		if createdAt, ok := existingItem["created_at"].(*types.AttributeValueMemberS); ok {
			if t, err := time.Parse(time.RFC3339, createdAt.Value); err == nil {
				updatedUser.CreatedAt = t
			}
		}

		// Return the updated user
		utils.RespondWithJSON(w, http.StatusOK, UserResponse{
			ID:       updatedUser.ID,
			Username: updatedUser.Username,
			Email:    updatedUser.Email,
			Role:     updatedUser.Role,
		})
	}
}

// DeleteUserHandler handles deleting a user by ID
func DeleteUserHandler(dbClient db.DatabaseClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the ID from the URL parameters
		id := r.URL.Query().Get("id")
		if id == "" {
			utils.RespondWithError(w, http.StatusBadRequest, "User ID is required")
			return
		}

		// First check if user exists
		key := map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#" + id},
			"SK": &types.AttributeValueMemberS{Value: "PROFILE"},
		}

		// Get the item to verify it exists and get the email for cleanup
		item, err := dbClient.GetItem(r.Context(), tableName, key)
		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Error checking user existence")
			return
		}

		if len(item) == 0 {
			utils.RespondWithError(w, http.StatusNotFound, "User not found")
			return
		}

		// Extract email for GSI cleanup if needed
		if data, ok := item["data"].(*types.AttributeValueMemberM); ok {
			if email, ok := data.Value["email"].(*types.AttributeValueMemberS); ok {
				_ = email.Value // Use email.Value if needed for cleanup
			}
		}

		// Delete the user using the DeleteItem method
		err = dbClient.DeleteItem(r.Context(), tableName, key)

		if err != nil {
			utils.RespondWithError(w, http.StatusInternalServerError, "Failed to delete user")
			return
		}

		// If you have any related data (like sessions, tokens, etc.), delete them here
		// For example, you might want to delete any sessions or auth tokens

		utils.RespondWithJSON(w, http.StatusOK, map[string]string{
			"message": "User deleted successfully",
		})
	}
}

// Helper function to find user by email
func findUserByEmail(dbClient db.DatabaseClient, email string) (*models.User, error) {
	filt := expression.Name("data.email").Equal(expression.Value(strings.ToLower(email))).
		And(expression.Name("entity_type").Equal(expression.Value(entityTypePK)))

	expr, err := expression.NewBuilder().WithFilter(filt).Build()
	if err != nil {
		return nil, err
	}

	input := &dynamodb.ScanInput{
		TableName:                 aws.String(tableName),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		Limit:                     aws.Int32(1),
	}

	result, err := dbClient.Scan(context.TODO(), input)
	if err != nil || len(result.Items) == 0 {
		return nil, err
	}

	var user models.User
	if data, ok := result.Items[0]["data"].(*types.AttributeValueMemberM); ok {
		err = attributevalue.UnmarshalMap(data.Value, &user)
		if err != nil {
			return nil, err
		}
		return &user, nil
	}

	return nil, nil
}
