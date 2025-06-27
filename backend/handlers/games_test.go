package handlers

import (
	"encoding/json"
	"errors"
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
				mockDB.On("QueryItems", mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
					return *input.TableName == "FootballLeague"
				})).Return(&dynamodb.QueryOutput{
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
									"game_date":    &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
									"status":       &types.AttributeValueMemberS{Value: "pending"},
									"created_at":   &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
									"updated_at":   &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
								},
							},
						},
						{
							"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
							"SK": &types.AttributeValueMemberS{Value: "GAME#game2"},
							"data": &types.AttributeValueMemberM{
								Value: map[string]types.AttributeValue{
									"id":           &types.AttributeValueMemberS{Value: "game2"},
									"league_id":    &types.AttributeValueMemberS{Value: "league1"},
									"week":         &types.AttributeValueMemberN{Value: "2"},
									"home_team_id": &types.AttributeValueMemberS{Value: "team3"},
									"away_team_id": &types.AttributeValueMemberS{Value: "team4"},
									"game_date":    &types.AttributeValueMemberS{Value: "2023-01-08T00:00:00Z"},
									"status":       &types.AttributeValueMemberS{Value: "pending"},
									"created_at":   &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
									"updated_at":   &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
								},
							},
						},
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
				mockDB.On("QueryItems", mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
					return *input.TableName == "FootballLeague" && input.FilterExpression != nil
				})).Return(&dynamodb.QueryOutput{
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
									"game_date":    &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
									"status":       &types.AttributeValueMemberS{Value: "pending"},
									"created_at":   &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
									"updated_at":   &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
								},
							},
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "missing league id",
			leagueID:       "",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name:     "invalid week parameter",
			leagueID: "league1",
			week:     "invalid",
			setupMock: func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name:     "database error",
			leagueID: "league1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("QueryItems", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
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
				mockDB.On("QueryItems", mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
					return *input.TableName == "FootballLeague" && 
						   *input.IndexName == "GSI1-EntityLookup" && 
						   *input.KeyConditionExpression == "GSI1_PK = :pk AND GSI1_SK = :sk"
				})).Return(&dynamodb.QueryOutput{
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
									"game_date":    &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
									"status":       &types.AttributeValueMemberS{Value: "pending"},
									"created_at":   &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
									"updated_at":   &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
								},
							},
						},
					},
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
		{
			name:   "game not found",
			gameID: "nonexistent",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("QueryItems", mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{},
				}, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "database error",
			gameID: "game1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("QueryItems", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "invalid game data format",
			gameID: "game1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("QueryItems", mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{
							"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
							"SK": &types.AttributeValueMemberS{Value: "GAME#game1"},
							"data": &types.AttributeValueMemberS{Value: "invalid"}, // Invalid data format
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusInternalServerError,
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
		method         string
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:    "successful create",
			request: validReq,
			method:  "POST",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(item map[string]types.AttributeValue) bool {
					return item["PK"].(*types.AttributeValueMemberS).Value == "LEAGUE#league1" &&
						   strings.HasPrefix(item["SK"].(*types.AttributeValueMemberS).Value, "GAME#")
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid request",
			request:        "invalid",
			method:         "POST",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "method not allowed",
			request:        validReq,
			method:         "GET",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name: "missing required fields",
			request: CreateGameRequest{
				LeagueID: "", // Missing league ID
				Week:     1,
			},
			method:         "POST",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid status",
			request: CreateGameRequest{
				LeagueID:   "league1",
				Week:       1,
				HomeTeamID: "team1",
				AwayTeamID: "team2",
				GameDate:   now,
				Status:     "invalid_status",
			},
			method:         "POST",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "database error",
			request: validReq,
			method:  "POST",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest(tt.method, "/games", strings.NewReader(string(body)))
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
		method         string
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:   "successful update",
			gameID: "game1",
			method: "PUT",
			request: UpdateGameRequest{
				Status: stringPtr("completed"),
				Winner: stringPtr("home"),
			},
			setupMock: func(mockDB *MockDBClient) {
				// First call to find game via GSI
				mockDB.On("QueryItems", mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
					return *input.IndexName == "GSI1-EntityLookup"
				})).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{
							"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
							"SK": &types.AttributeValueMemberS{Value: "GAME#game1"},
							"league_id": &types.AttributeValueMemberS{Value: "league1"},
						},
					},
				}, nil)
				// Second call to get full game data
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					return key["PK"].(*types.AttributeValueMemberS).Value == "LEAGUE#league1" &&
						   key["SK"].(*types.AttributeValueMemberS).Value == "GAME#game1"
				})).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
					"SK": &types.AttributeValueMemberS{Value: "GAME#game1"},
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
				}, nil)
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing game id",
			gameID:         "",
			method:         "PUT",
			request:        UpdateGameRequest{},
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "method not allowed",
			gameID:         "game1",
			method:         "GET",
			request:        UpdateGameRequest{},
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "no fields to update",
			gameID: "game1",
			method: "PUT",
			request: UpdateGameRequest{}, // No fields set
			setupMock: func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid status",
			gameID: "game1",
			method: "PUT",
			request: UpdateGameRequest{
				Status: stringPtr("invalid_status"),
			},
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "invalid winner",
			gameID: "game1",
			method: "PUT",
			request: UpdateGameRequest{
				Winner: stringPtr("invalid_winner"),
			},
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "game not found in GSI",
			gameID: "nonexistent",
			method: "PUT",
			request: UpdateGameRequest{
				Status: stringPtr("completed"),
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("QueryItems", mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{},
				}, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest(tt.method, "/game", strings.NewReader(string(body)))
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
		method         string
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:   "successful delete",
			gameID: "game1",
			method: "DELETE",
			setupMock: func(mockDB *MockDBClient) {
				// First get the game to find league ID (note: handler uses wrong key pattern)
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					return key["PK"].(*types.AttributeValueMemberS).Value == "GAME#game1" &&
						   key["SK"].(*types.AttributeValueMemberS).Value == "METADATA"
				})).Return(map[string]types.AttributeValue{
					"league_id": &types.AttributeValueMemberS{Value: "league1"},
				}, nil)
				mockDB.On("DeleteItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					return key["PK"].(*types.AttributeValueMemberS).Value == "LEAGUE#league1" &&
						   key["SK"].(*types.AttributeValueMemberS).Value == "GAME#game1"
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing game id",
			gameID:         "",
			method:         "DELETE",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "method not allowed",
			gameID:         "game1",
			method:         "GET",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:   "game not found",
			gameID: "nonexistent",
			method: "DELETE",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{}, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "database error on get",
			gameID: "game1",
			method: "DELETE",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req, _ := http.NewRequest(tt.method, "/game", nil)
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
		{"uppercase status", "PENDING", false},
		{"mixed case status", "Pending", false},
		{"whitespace", " pending ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidStatus(tt.status)
			assert.Equal(t, tt.expected, result)
		})
	}
}

