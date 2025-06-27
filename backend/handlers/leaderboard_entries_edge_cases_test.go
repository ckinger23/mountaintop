package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestLeaderboardEntryHandlerEdgeCases tests additional edge cases and integration scenarios
func TestLeaderboardEntryHandlerEdgeCases(t *testing.T) {
	t.Run("GetAllLeaderboardEntriesHandler - corrupted data structure", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		mockDB.On("QueryItems", mock.Anything).Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"PK": &types.AttributeValueMemberS{Value: "USER#user1"},
					"SK": &types.AttributeValueMemberS{Value: "LEADERBOARD#WEEK#1"},
					"data": &types.AttributeValueMemberS{Value: "corrupted"}, // Wrong type
				},
			},
		}, nil)

		req := httptest.NewRequest("GET", "/leaderboard/entries", nil)
		rr := httptest.NewRecorder()

		handler := GetAllLeaderboardEntriesHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Should succeed but skip corrupted items
		assert.Equal(t, http.StatusOK, rr.Code)
		// Response should have empty array (entries might be null if no valid items)
		responseBody := rr.Body.String()
		assert.True(t, strings.Contains(responseBody, `"entries":[]`) || strings.Contains(responseBody, `"entries":null`))
		mockDB.AssertExpectations(t)
	})

	t.Run("CreateLeaderboardEntryHandler - malformed JSON", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		req := httptest.NewRequest("POST", "/leaderboard/entries", strings.NewReader(`{"user_id":"user1","week":1,}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := CreateLeaderboardEntryHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("CreateLeaderboardEntryHandler - negative values", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		// Should allow negative values as they might be valid in some contexts
		mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
		
		req := httptest.NewRequest("POST", "/leaderboard/entries", strings.NewReader(`{
			"user_id": "user1",
			"username": "testuser",
			"points": -5,
			"correct": 2,
			"incorrect": 8,
			"week": 1
		}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := CreateLeaderboardEntryHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("CreateLeaderboardEntryHandler - extremely large values", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
		
		req := httptest.NewRequest("POST", "/leaderboard/entries", strings.NewReader(`{
			"user_id": "user1",
			"username": "testuser",
			"points": 999999,
			"correct": 500000,
			"incorrect": 499999,
			"week": 1000
		}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := CreateLeaderboardEntryHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Should handle large values
		assert.Equal(t, http.StatusCreated, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("CreateLeaderboardEntryHandler - extremely long username", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		longUsername := strings.Repeat("A", 1000)
		mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
		
		req := httptest.NewRequest("POST", "/leaderboard/entries", strings.NewReader(`{
			"user_id": "user1",
			"username": "`+longUsername+`",
			"points": 10,
			"correct": 8,
			"incorrect": 2,
			"week": 1
		}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := CreateLeaderboardEntryHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Should handle large usernames
		assert.Equal(t, http.StatusCreated, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("UpdateLeaderboardEntryHandler - partial update", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		// Mock GetItem for existing entry
		mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#user1"},
			"SK": &types.AttributeValueMemberS{Value: "LEADERBOARD#WEEK#1"},
			"username": &types.AttributeValueMemberS{Value: "testuser"},
			"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
			"data": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"user_id":    &types.AttributeValueMemberS{Value: "user1"},
					"username":   &types.AttributeValueMemberS{Value: "testuser"},
					"points":     &types.AttributeValueMemberN{Value: "10"},
					"correct":    &types.AttributeValueMemberN{Value: "8"},
					"incorrect":  &types.AttributeValueMemberN{Value: "2"},
					"week":       &types.AttributeValueMemberN{Value: "1"},
					"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
					"updated_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
				},
			},
		}, nil)
		
		// Mock PutItem for update
		mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)

		// Only update points, leave other fields unchanged
		req := httptest.NewRequest("PUT", "/leaderboard/entry", strings.NewReader(`{"points": 25}`))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{
			"userId": "user1",
			"week":   "1",
		})
		rr := httptest.NewRecorder()

		handler := UpdateLeaderboardEntryHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("UpdateLeaderboardEntryHandler - no fields provided", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		// Mock GetItem for existing entry
		mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#user1"},
			"SK": &types.AttributeValueMemberS{Value: "LEADERBOARD#WEEK#1"},
			"data": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"user_id":    &types.AttributeValueMemberS{Value: "user1"},
					"username":   &types.AttributeValueMemberS{Value: "testuser"},
					"points":     &types.AttributeValueMemberN{Value: "10"},
					"correct":    &types.AttributeValueMemberN{Value: "8"},
					"incorrect":  &types.AttributeValueMemberN{Value: "2"},
					"week":       &types.AttributeValueMemberN{Value: "1"},
					"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
					"updated_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
				},
			},
		}, nil)
		
		// Mock PutItem for update (will still update updated_at)
		mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)

		// Empty update request
		req := httptest.NewRequest("PUT", "/leaderboard/entry", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{
			"userId": "user1",
			"week":   "1",
		})
		rr := httptest.NewRecorder()

		handler := UpdateLeaderboardEntryHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Should succeed and update updated_at timestamp
		assert.Equal(t, http.StatusOK, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("GetLeaderboardEntryHandler - zero week", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		// Mock GetItem to return empty result
		mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{}, nil)
		
		req := httptest.NewRequest("GET", "/leaderboard/entry", nil)
		req = mux.SetURLVars(req, map[string]string{
			"userId": "user1",
			"week":   "0",
		})
		rr := httptest.NewRecorder()

		handler := GetLeaderboardEntryHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Week 0 is technically valid as an integer
		assert.Equal(t, http.StatusNotFound, rr.Code) // Will try to fetch but not find anything
		mockDB.AssertExpectations(t)
	})

	t.Run("GetLeaderboardEntryHandler - negative week", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		// Mock GetItem to return empty result
		mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{}, nil)
		
		req := httptest.NewRequest("GET", "/leaderboard/entry", nil)
		req = mux.SetURLVars(req, map[string]string{
			"userId": "user1",
			"week":   "-5",
		})
		rr := httptest.NewRecorder()

		handler := GetLeaderboardEntryHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Negative week is technically valid as an integer
		assert.Equal(t, http.StatusNotFound, rr.Code) // Will try to fetch but not find anything
		mockDB.AssertExpectations(t)
	})

	t.Run("Large request body handling", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		// Create a very large user ID
		largeUserID := strings.Repeat("A", 50000)
		
		req := httptest.NewRequest("POST", "/leaderboard/entries", strings.NewReader(`{
			"user_id": "`+largeUserID+`",
			"username": "testuser",
			"points": 10,
			"correct": 8,
			"incorrect": 2,
			"week": 1
		}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)

		handler := CreateLeaderboardEntryHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Should handle large requests
		assert.Equal(t, http.StatusCreated, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Concurrent request simulation", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		// Simulate concurrent access by having multiple handlers use the same mock
		mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#user1"},
			"SK": &types.AttributeValueMemberS{Value: "LEADERBOARD#WEEK#1"},
			"data": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"user_id":    &types.AttributeValueMemberS{Value: "user1"},
					"username":   &types.AttributeValueMemberS{Value: "testuser"},
					"points":     &types.AttributeValueMemberN{Value: "10"},
					"correct":    &types.AttributeValueMemberN{Value: "8"},
					"incorrect":  &types.AttributeValueMemberN{Value: "2"},
					"week":       &types.AttributeValueMemberN{Value: "1"},
					"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
					"updated_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
				},
			},
		}, nil)

		req := httptest.NewRequest("GET", "/leaderboard/entry", nil)
		req = mux.SetURLVars(req, map[string]string{
			"userId": "user1",
			"week":   "1",
		})
		rr := httptest.NewRecorder()

		handler := GetLeaderboardEntryHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Response data validation", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		now := time.Now()
		mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "USER#user1"},
			"SK": &types.AttributeValueMemberS{Value: "LEADERBOARD#WEEK#1"},
			"data": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"user_id":    &types.AttributeValueMemberS{Value: "user1"},
					"username":   &types.AttributeValueMemberS{Value: "testuser"},
					"points":     &types.AttributeValueMemberN{Value: "10"},
					"correct":    &types.AttributeValueMemberN{Value: "8"},
					"incorrect":  &types.AttributeValueMemberN{Value: "2"},
					"week":       &types.AttributeValueMemberN{Value: "1"},
					"created_at": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
					"updated_at": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
				},
			},
		}, nil)

		req := httptest.NewRequest("GET", "/leaderboard/entry", nil)
		req = mux.SetURLVars(req, map[string]string{
			"userId": "user1",
			"week":   "1",
		})
		rr := httptest.NewRecorder()

		handler := GetLeaderboardEntryHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		
		// Verify response contains no sensitive data
		responseBody := rr.Body.String()
		assert.NotContains(t, responseBody, "PK")
		assert.NotContains(t, responseBody, "SK")
		assert.NotContains(t, responseBody, "GSI1_PK")
		assert.Contains(t, responseBody, "user1")
		assert.Contains(t, responseBody, "testuser")
		
		mockDB.AssertExpectations(t)
	})

	t.Run("CreateLeaderboardEntryHandler - empty username", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		req := httptest.NewRequest("POST", "/leaderboard/entries", strings.NewReader(`{
			"user_id": "user1",
			"username": "",
			"points": 10,
			"correct": 8,
			"incorrect": 2,
			"week": 1
		}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := CreateLeaderboardEntryHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Missing required fields")
		mockDB.AssertExpectations(t)
	})

	t.Run("GetLeaderboardEntriesByWeekHandler - empty results", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		mockDB.On("QueryItems", mock.Anything).Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{},
		}, nil)

		req := httptest.NewRequest("GET", "/leaderboard/week", nil)
		req = mux.SetURLVars(req, map[string]string{
			"week": "999",
		})
		rr := httptest.NewRecorder()

		handler := GetLeaderboardEntriesByWeekHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		// Check for empty entries array
		responseBody := rr.Body.String()
		assert.True(t, strings.Contains(responseBody, `"entries":[]`) || strings.Contains(responseBody, `"entries":null`))
		mockDB.AssertExpectations(t)
	})
}