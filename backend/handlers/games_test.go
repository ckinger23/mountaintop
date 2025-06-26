package handlers

import (
	"encoding/json"
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

func TestGetAllGamesHandler(t *testing.T) {
	tests := []struct {
		name           string
		leagueID       string
		week           string
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:     "successful fetch all games",
			leagueID: "league1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("Query", mock.Anything, mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
					return *input.TableName == "FootballLeague"
				})).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{"id": &types.AttributeValueMemberS{Value: "game1"}},
						{"id": &types.AttributeValueMemberS{Value: "game2"}},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:     "filter by week",
			leagueID: "league1",
			week:     "1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("Query", mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{"id": &types.AttributeValueMemberS{Value: "game1"}},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req, _ := http.NewRequest("GET", "/games", nil)
			q := req.URL.Query()
			q.Add("league_id", tt.leagueID)
			if tt.week != "" {
				q.Add("week", tt.week)
			}
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()
			handler := GetAllGamesHandler(mockDB)

			// Execute
			handler.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedCount > 0 {
				var response GamesResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.Games, tt.expectedCount)
			}
			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetGameHandler(t *testing.T) {
	tests := []struct {
		name           string
		gameID         string
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:   "successful fetch",
			gameID: "game1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, mock.Anything).Return(map[string]types.AttributeValue{
					"id":           &types.AttributeValueMemberS{Value: "game1"},
					"league_id":    &types.AttributeValueMemberS{Value: "league1"},
					"week":         &types.AttributeValueMemberN{Value: "1"},
					"home_team_id": &types.AttributeValueMemberS{Value: "team1"},
					"away_team_id": &types.AttributeValueMemberS{Value: "team2"},
					"game_date":    &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
					"status":       &types.AttributeValueMemberS{Value: "pending"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing game id",
			gameID:         "",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req, _ := http.NewRequest("GET", "/game", nil)
			q := req.URL.Query()
			if tt.gameID != "" {
				q.Add("id", tt.gameID)
			}
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()
			handler := GetGameHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestCreateGameHandler(t *testing.T) {
	now := time.Now()
	validReq := CreateGameRequest{
		LeagueID:   "league1",
		Week:       1,
		HomeTeamID: "team1",
		AwayTeamID: "team2",
		GameDate:   now,
		Status:     "pending",
	}

	tests := []struct {
		name           string
		request        interface{}
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:    "successful create",
			request: validReq,
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).
					Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid request",
			request:        "invalid",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("POST", "/games", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := CreateGameHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestUpdateGameHandler(t *testing.T) {
	tests := []struct {
		name           string
		gameID         string
		request        interface{}
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:   "successful update",
			gameID: "game1",
			request: map[string]string{
				"status": "completed",
				"winner": "home",
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, mock.Anything).Return(map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "game1"},
				}, nil)
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing game id",
			gameID:         "",
			request:        map[string]string{},
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("PUT", "/game", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
			q := req.URL.Query()
			if tt.gameID != "" {
				q.Add("id", tt.gameID)
			}
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()
			handler := UpdateGameHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestDeleteGameHandler(t *testing.T) {
	tests := []struct {
		name           string
		gameID         string
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:   "successful delete",
			gameID: "game1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("DeleteItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "missing game id",
			gameID:         "",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req, _ := http.NewRequest("DELETE", "/game", nil)
			q := req.URL.Query()
			if tt.gameID != "" {
				q.Add("id", tt.gameID)
			}
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()
			handler := DeleteGameHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"valid pending", "pending", true},
		{"valid in_progress", "in_progress", true},
		{"valid completed", "completed", true},
		{"invalid status", "invalid", false},
		{"empty status", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}
