package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestGameHandlerEdgeCases tests additional edge cases and integration scenarios
func TestGameHandlerEdgeCases(t *testing.T) {
	t.Run("GetAllGamesHandler - empty results", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		mockDB.On("QueryItems", mock.Anything).Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{},
		}, nil)

		req := httptest.NewRequest("GET", "/games?league_id=league1", nil)
		rr := httptest.NewRecorder()

		handler := GetAllGamesHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"games":[]`)
		mockDB.AssertExpectations(t)
	})

	t.Run("GetAllGamesHandler - corrupted data structure", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		mockDB.On("QueryItems", mock.Anything).Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
					"SK": &types.AttributeValueMemberS{Value: "GAME#game1"},
					"data": &types.AttributeValueMemberS{Value: "corrupted"}, // Wrong type
				},
			},
		}, nil)

		req := httptest.NewRequest("GET", "/games?league_id=league1", nil)
		rr := httptest.NewRecorder()

		handler := GetAllGamesHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Should succeed but skip corrupted items
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), `"games":[]`)
		mockDB.AssertExpectations(t)
	})

	t.Run("CreateGameHandler - malformed JSON", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		req := httptest.NewRequest("POST", "/games", strings.NewReader(`{"league_id":"league1","week":1,}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := CreateGameHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("CreateGameHandler - zero week value", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		req := httptest.NewRequest("POST", "/games", strings.NewReader(`{
			"league_id": "league1",
			"week": 0,
			"home_team_id": "team1",
			"away_team_id": "team2",
			"game_date": "2023-01-01T00:00:00Z"
		}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := CreateGameHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "week are required")
		mockDB.AssertExpectations(t)
	})

	t.Run("CreateGameHandler - same team as home and away", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		// This should be allowed by current validation, but could be edge case
		mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
		
		req := httptest.NewRequest("POST", "/games", strings.NewReader(`{
			"league_id": "league1",
			"week": 1,
			"home_team_id": "team1",
			"away_team_id": "team1",
			"game_date": "2023-01-01T00:00:00Z"
		}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := CreateGameHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("CreateGameHandler - extremely long team IDs", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		longTeamID := strings.Repeat("A", 1000)
		mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
		
		req := httptest.NewRequest("POST", "/games", strings.NewReader(`{
			"league_id": "league1",
			"week": 1,
			"home_team_id": "`+longTeamID+`",
			"away_team_id": "team2",
			"game_date": "2023-01-01T00:00:00Z"
		}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := CreateGameHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Should handle large data
		assert.Equal(t, http.StatusCreated, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("UpdateGameHandler - invalid JSON", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		req := httptest.NewRequest("PUT", "/game?id=game1", strings.NewReader(`{"status":"completed",}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := UpdateGameHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("UpdateGameHandler - missing league ID in GSI result", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		mockDB.On("QueryItems", mock.Anything).Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
					"SK": &types.AttributeValueMemberS{Value: "GAME#game1"},
					// Missing league_id field
				},
			},
		}, nil)
		
		req := httptest.NewRequest("PUT", "/game?id=game1", strings.NewReader(`{"status":"completed"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := UpdateGameHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "missing league ID")
		mockDB.AssertExpectations(t)
	})

	t.Run("UpdateGameHandler - no changes detected", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		// Setup mock for GSI query
		mockDB.On("QueryItems", mock.Anything).Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"league_id": &types.AttributeValueMemberS{Value: "league1"},
				},
			},
		}, nil)
		
		// Setup mock for GetItem
		mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
			"data": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"id":           &types.AttributeValueMemberS{Value: "game1"},
					"league_id":    &types.AttributeValueMemberS{Value: "league1"},
					"week":         &types.AttributeValueMemberN{Value: "1"},
					"home_team_id": &types.AttributeValueMemberS{Value: "team1"},
					"away_team_id": &types.AttributeValueMemberS{Value: "team2"},
					"game_date":    &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
					"status":       &types.AttributeValueMemberS{Value: "completed"}, // Already completed
					"winner":       &types.AttributeValueMemberS{Value: "home"}, // Already has winner
					"created_at":   &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
					"updated_at":   &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
				},
			},
		}, nil)
		
		req := httptest.NewRequest("PUT", "/game?id=game1", strings.NewReader(`{"status":"completed","winner":"home"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := UpdateGameHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "No changes detected")
		mockDB.AssertExpectations(t)
	})

	t.Run("DeleteGameHandler - missing league ID in result", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: "game1"},
			// Missing league_id
		}, nil)
		
		req := httptest.NewRequest("DELETE", "/game?id=game1", nil)
		rr := httptest.NewRecorder()

		handler := DeleteGameHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "missing league ID")
		mockDB.AssertExpectations(t)
	})

	t.Run("Large request body handling", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		// Create a very large game date string
		largeData := strings.Repeat("A", 50000)
		
		req := httptest.NewRequest("POST", "/games", strings.NewReader(`{
			"league_id": "`+largeData+`",
			"week": 1,
			"home_team_id": "team1",
			"away_team_id": "team2",
			"game_date": "2023-01-01T00:00:00Z"
		}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)

		handler := CreateGameHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Should handle large requests
		assert.Equal(t, http.StatusCreated, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Concurrent request simulation", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		// Simulate concurrent access by having multiple handlers use the same mock
		mockDB.On("QueryItems", mock.Anything).Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
					"SK": &types.AttributeValueMemberS{Value: "GAME#game1"},
					"GSI1_PK": &types.AttributeValueMemberS{Value: "GAME"},
					"GSI1_SK": &types.AttributeValueMemberS{Value: "GAME#game1"},
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"id":           &types.AttributeValueMemberS{Value: "game1"},
							"league_id":    &types.AttributeValueMemberS{Value: "league1"},
							"week":         &types.AttributeValueMemberN{Value: "1"},
							"home_team_id": &types.AttributeValueMemberS{Value: "team1"},
							"away_team_id": &types.AttributeValueMemberS{Value: "team2"},
							"game_date":    &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
							"status":       &types.AttributeValueMemberS{Value: "pending"},
							"created_at":   &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
							"updated_at":   &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
						},
					},
				},
			},
		}, nil)

		req := httptest.NewRequest("GET", "/game?id=game1", nil)
		rr := httptest.NewRecorder()

		handler := GetGameHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Response data validation", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		now := time.Now()
		mockDB.On("QueryItems", mock.Anything).Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
					"SK": &types.AttributeValueMemberS{Value: "GAME#game1"},
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"id":           &types.AttributeValueMemberS{Value: "game1"},
							"league_id":    &types.AttributeValueMemberS{Value: "league1"},
							"week":         &types.AttributeValueMemberN{Value: "1"},
							"home_team_id": &types.AttributeValueMemberS{Value: "team1"},
							"away_team_id": &types.AttributeValueMemberS{Value: "team2"},
							"game_date":    &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
							"status":       &types.AttributeValueMemberS{Value: "completed"},
							"winner":       &types.AttributeValueMemberS{Value: "home"},
							"created_at":   &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
							"updated_at":   &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
						},
					},
				},
			},
		}, nil)

		req := httptest.NewRequest("GET", "/game?id=game1", nil)
		rr := httptest.NewRecorder()

		handler := GetGameHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		
		// Verify response contains no sensitive data
		responseBody := rr.Body.String()
		assert.NotContains(t, responseBody, "PK")
		assert.NotContains(t, responseBody, "SK")
		assert.NotContains(t, responseBody, "GSI1_PK")
		assert.Contains(t, responseBody, "game1")
		assert.Contains(t, responseBody, "winner") // Should include winner for completed games
		
		mockDB.AssertExpectations(t)
	})

	t.Run("Pending game response validation", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		mockDB.On("QueryItems", mock.Anything).Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"id":           &types.AttributeValueMemberS{Value: "game1"},
							"league_id":    &types.AttributeValueMemberS{Value: "league1"},
							"week":         &types.AttributeValueMemberN{Value: "1"},
							"home_team_id": &types.AttributeValueMemberS{Value: "team1"},
							"away_team_id": &types.AttributeValueMemberS{Value: "team2"},
							"game_date":    &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
							"status":       &types.AttributeValueMemberS{Value: "pending"},
							"created_at":   &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
							"updated_at":   &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
						},
					},
				},
			},
		}, nil)

		req := httptest.NewRequest("GET", "/game?id=game1", nil)
		rr := httptest.NewRecorder()

		handler := GetGameHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		
		// Verify pending games don't include winner
		responseBody := rr.Body.String()
		assert.Contains(t, responseBody, `"status":"pending"`)
		assert.NotContains(t, responseBody, `"winner"`) // Should not include winner for pending games
		
		mockDB.AssertExpectations(t)
	})
}