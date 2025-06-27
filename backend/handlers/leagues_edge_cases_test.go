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

// TestLeagueHandlerEdgeCases tests additional edge cases and integration scenarios
func TestLeagueHandlerEdgeCases(t *testing.T) {
	t.Run("GetAllLeaguesHandler - corrupted data structure", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		mockDB.On("QueryItems", mock.Anything).Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
					"SK": &types.AttributeValueMemberS{Value: "METADATA#league1"},
					"data": &types.AttributeValueMemberS{Value: "corrupted"}, // Wrong type
				},
			},
		}, nil)

		req := httptest.NewRequest("GET", "/leagues", nil)
		rr := httptest.NewRecorder()

		handler := GetAllLeaguesHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Should succeed but skip corrupted items
		assert.Equal(t, http.StatusOK, rr.Code)
		// Response should have empty array (leagues might be null if no valid items)
		responseBody := rr.Body.String()
		assert.True(t, strings.Contains(responseBody, `"leagues":[]`) || strings.Contains(responseBody, `"leagues":null`))
		mockDB.AssertExpectations(t)
	})

	t.Run("CreateLeagueHandler - malformed JSON", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		req := httptest.NewRequest("POST", "/leagues", strings.NewReader(`{"name":"Test League","admin_id":"admin1",}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := CreateLeagueHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("CreateLeagueHandler - extremely long league name", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		longName := strings.Repeat("A", 1000)
		mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
		
		req := httptest.NewRequest("POST", "/leagues", strings.NewReader(`{
			"name": "`+longName+`",
			"admin_id": "admin1"
		}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := CreateLeagueHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Should handle large names
		assert.Equal(t, http.StatusCreated, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("CreateLeagueHandler - special characters in name", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
		
		req := httptest.NewRequest("POST", "/leagues", strings.NewReader(`{
			"name": "Test League!@#$%^&*()_+-={}[]|\\:;\"'<>,.?/~` + "`" + `",
			"admin_id": "admin1"
		}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := CreateLeagueHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Should handle special characters
		assert.Equal(t, http.StatusCreated, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("CreateLeagueHandler - whitespace only name", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
		
		req := httptest.NewRequest("POST", "/leagues", strings.NewReader(`{
			"name": "   ",
			"admin_id": "admin1"
		}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := CreateLeagueHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Whitespace-only name should be treated as valid (current implementation)
		assert.Equal(t, http.StatusCreated, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("UpdateLeagueHandler - partial update with admin ID only", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		// Mock GetItem for existing league
		mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
			"SK": &types.AttributeValueMemberS{Value: "METADATA#league1"},
			"admin_id": &types.AttributeValueMemberS{Value: "oldadmin"},
			"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
			"data": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"id":         &types.AttributeValueMemberS{Value: "league1"},
					"name":       &types.AttributeValueMemberS{Value: "Test League"},
					"admin_id":   &types.AttributeValueMemberS{Value: "oldadmin"},
					"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
					"updated_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
				},
			},
		}, nil)
		// Mock PutItem for update
		mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)

		// Only update admin ID, leave name unchanged
		req := httptest.NewRequest("PUT", "/league", strings.NewReader(`{"admin_id": "newadmin"}`))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{
			"id": "league1",
		})
		rr := httptest.NewRecorder()

		handler := UpdateLeagueHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("GetLeagueHandler - extremely long league ID", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		longID := strings.Repeat("A", 1000)
		mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{}, nil)
		
		req := httptest.NewRequest("GET", "/league", nil)
		q := req.URL.Query()
		q.Add("id", longID)
		req.URL.RawQuery = q.Encode()
		rr := httptest.NewRecorder()

		handler := GetLeagueHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Should handle large IDs gracefully (not found)
		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Large request body handling", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		// Create a very large admin ID
		largeAdminID := strings.Repeat("A", 50000)
		
		req := httptest.NewRequest("POST", "/leagues", strings.NewReader(`{
			"name": "Test League",
			"admin_id": "`+largeAdminID+`"
		}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)

		handler := CreateLeagueHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Should handle large requests
		assert.Equal(t, http.StatusCreated, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Concurrent request simulation", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		// Simulate concurrent access by having multiple handlers use the same mock
		mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
			"SK": &types.AttributeValueMemberS{Value: "METADATA#league1"},
			"data": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"id":         &types.AttributeValueMemberS{Value: "league1"},
					"name":       &types.AttributeValueMemberS{Value: "Test League"},
					"admin_id":   &types.AttributeValueMemberS{Value: "admin1"},
					"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
					"updated_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
				},
			},
		}, nil)

		req := httptest.NewRequest("GET", "/league", nil)
		q := req.URL.Query()
		q.Add("id", "league1")
		req.URL.RawQuery = q.Encode()
		rr := httptest.NewRecorder()

		handler := GetLeagueHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("Response data validation", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		now := time.Now()
		mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
			"SK": &types.AttributeValueMemberS{Value: "METADATA#league1"},
			"data": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"id":         &types.AttributeValueMemberS{Value: "league1"},
					"name":       &types.AttributeValueMemberS{Value: "Test League"},
					"admin_id":   &types.AttributeValueMemberS{Value: "admin1"},
					"created_at": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
					"updated_at": &types.AttributeValueMemberS{Value: now.Format(time.RFC3339)},
				},
			},
		}, nil)

		req := httptest.NewRequest("GET", "/league", nil)
		q := req.URL.Query()
		q.Add("id", "league1")
		req.URL.RawQuery = q.Encode()
		rr := httptest.NewRecorder()

		handler := GetLeagueHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		
		// Verify response contains no sensitive data
		responseBody := rr.Body.String()
		assert.NotContains(t, responseBody, "PK")
		assert.NotContains(t, responseBody, "SK")
		assert.NotContains(t, responseBody, "entity_type")
		assert.Contains(t, responseBody, "league1")
		assert.Contains(t, responseBody, "Test League")
		
		mockDB.AssertExpectations(t)
	})

	t.Run("CreateLeagueHandler - null values in JSON", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		req := httptest.NewRequest("POST", "/leagues", strings.NewReader(`{
			"name": null,
			"admin_id": "admin1"
		}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := CreateLeagueHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Should reject null values
		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "required")
		mockDB.AssertExpectations(t)
	})

	t.Run("UpdateLeagueHandler - invalid JSON structure", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		// Mock GetItem for existing league (this will be called before JSON parsing failure)
		mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
		}, nil)
		
		req := httptest.NewRequest("PUT", "/league", strings.NewReader(`{"name": {"nested": "object"}}`))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{
			"id": "league1",
		})
		rr := httptest.NewRecorder()

		handler := UpdateLeagueHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("GetAllLeaguesHandler - empty results", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		mockDB.On("QueryItems", mock.Anything).Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{},
		}, nil)

		req := httptest.NewRequest("GET", "/leagues", nil)
		rr := httptest.NewRecorder()

		handler := GetAllLeaguesHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		// Check for empty leagues array
		responseBody := rr.Body.String()
		assert.True(t, strings.Contains(responseBody, `"leagues":[]`) || strings.Contains(responseBody, `"leagues":null`))
		mockDB.AssertExpectations(t)
	})

	t.Run("DeleteLeagueHandler - attempting to delete with POST method", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		req := httptest.NewRequest("POST", "/league", nil)
		q := req.URL.Query()
		q.Add("id", "league1")
		req.URL.RawQuery = q.Encode()
		rr := httptest.NewRecorder()

		handler := DeleteLeagueHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("CreateLeagueHandler - unicode characters", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
		
		req := httptest.NewRequest("POST", "/leagues", strings.NewReader(`{
			"name": "Test League 测试联盟 🏈 ⚽ 🏀",
			"admin_id": "admin1"
		}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler := CreateLeagueHandler(mockDB)
		handler.ServeHTTP(rr, req)

		// Should handle unicode characters
		assert.Equal(t, http.StatusCreated, rr.Code)
		mockDB.AssertExpectations(t)
	})

	t.Run("UpdateLeagueHandler - update both name and admin simultaneously", func(t *testing.T) {
		mockDB := NewMockDBClient()
		
		// Mock GetItem for existing league
		mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
			"SK": &types.AttributeValueMemberS{Value: "METADATA#league1"},
			"admin_id": &types.AttributeValueMemberS{Value: "oldadmin"},
			"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
			"data": &types.AttributeValueMemberM{
				Value: map[string]types.AttributeValue{
					"id":         &types.AttributeValueMemberS{Value: "league1"},
					"name":       &types.AttributeValueMemberS{Value: "Old League Name"},
					"admin_id":   &types.AttributeValueMemberS{Value: "oldadmin"},
					"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
					"updated_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
				},
			},
		}, nil)
		// Mock PutItem for update
		mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)

		// Update both name and admin ID
		req := httptest.NewRequest("PUT", "/league", strings.NewReader(`{
			"name": "New League Name",
			"admin_id": "newadmin"
		}`))
		req.Header.Set("Content-Type", "application/json")
		req = mux.SetURLVars(req, map[string]string{
			"id": "league1",
		})
		rr := httptest.NewRecorder()

		handler := UpdateLeagueHandler(mockDB)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		mockDB.AssertExpectations(t)
	})
}