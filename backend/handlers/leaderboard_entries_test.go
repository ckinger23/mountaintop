package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAllLeaderboardEntriesHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "successful fetch all entries",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("QueryItems", mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
					return *input.TableName == "FootballLeague" &&
						   *input.IndexName == "GSI1-EntityLookup" &&
						   *input.KeyConditionExpression == "GSI1_PK = :pk"
				})).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{
							"PK": &types.AttributeValueMemberS{Value: "USER#user1"},
							"SK": &types.AttributeValueMemberS{Value: "LEADERBOARD#WEEK#1"},
							"GSI1_PK": &types.AttributeValueMemberS{Value: "LEADERBOARD_ENTRY"},
							"data": &types.AttributeValueMemberM{
								Value: map[string]types.AttributeValue{
									"user_id":    &types.AttributeValueMemberS{Value: "user1"},
									"username":   &types.AttributeValueMemberS{Value: "testuser1"},
									"points":     &types.AttributeValueMemberN{Value: "10"},
									"correct":    &types.AttributeValueMemberN{Value: "8"},
									"incorrect":  &types.AttributeValueMemberN{Value: "2"},
									"week":       &types.AttributeValueMemberN{Value: "1"},
									"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
									"updated_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
								},
							},
						},
						{
							"PK": &types.AttributeValueMemberS{Value: "USER#user2"},
							"SK": &types.AttributeValueMemberS{Value: "LEADERBOARD#WEEK#1"},
							"GSI1_PK": &types.AttributeValueMemberS{Value: "LEADERBOARD_ENTRY"},
							"data": &types.AttributeValueMemberM{
								Value: map[string]types.AttributeValue{
									"user_id":    &types.AttributeValueMemberS{Value: "user2"},
									"username":   &types.AttributeValueMemberS{Value: "testuser2"},
									"points":     &types.AttributeValueMemberN{Value: "15"},
									"correct":    &types.AttributeValueMemberN{Value: "10"},
									"incorrect":  &types.AttributeValueMemberN{Value: "5"},
									"week":       &types.AttributeValueMemberN{Value: "1"},
									"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
									"updated_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
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
			name: "database error",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("QueryItems", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
		{
			name: "empty results",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("QueryItems", mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req, _ := http.NewRequest("GET", "/leaderboard/entries", nil)
			rr := httptest.NewRecorder()
			handler := GetAllLeaderboardEntriesHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedCount > 0 {
				var response LeaderboardResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.Entries, tt.expectedCount)
			}
			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetLeaderboardEntryHandler(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		week           string
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:   "successful fetch",
			userID: "user1",
			week:   "1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					return key["PK"].(*types.AttributeValueMemberS).Value == "USER#user1" &&
						   key["SK"].(*types.AttributeValueMemberS).Value == "LEADERBOARD#WEEK#1"
				})).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "USER#user1"},
					"SK": &types.AttributeValueMemberS{Value: "LEADERBOARD#WEEK#1"},
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"user_id":    &types.AttributeValueMemberS{Value: "user1"},
							"username":   &types.AttributeValueMemberS{Value: "testuser1"},
							"points":     &types.AttributeValueMemberN{Value: "10"},
							"correct":    &types.AttributeValueMemberN{Value: "8"},
							"incorrect":  &types.AttributeValueMemberN{Value: "2"},
							"week":       &types.AttributeValueMemberN{Value: "1"},
							"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
							"updated_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid week format",
			userID: "user1",
			week:   "invalid",
			setupMock: func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "entry not found",
			userID: "user1",
			week:   "1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{}, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "database error",
			userID: "user1",
			week:   "1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "invalid data format",
			userID: "user1",
			week:   "1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "USER#user1"},
					"SK": &types.AttributeValueMemberS{Value: "LEADERBOARD#WEEK#1"},
					"data": &types.AttributeValueMemberS{Value: "invalid"}, // Wrong type
				}, nil)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req := httptest.NewRequest("GET", "/leaderboard/entry", nil)
			// Use mux to set path variables
			req = mux.SetURLVars(req, map[string]string{
				"userId": tt.userID,
				"week":   tt.week,
			})

			rr := httptest.NewRecorder()
			handler := GetLeaderboardEntryHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestCreateLeaderboardEntryHandler(t *testing.T) {
	validReq := CreateLeaderboardEntryRequest{
		UserID:    "user1",
		Username:  "testuser",
		Points:    10,
		Correct:   8,
		Incorrect: 2,
		Week:      1,
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
					return item["PK"].(*types.AttributeValueMemberS).Value == "USER#user1" &&
						   item["SK"].(*types.AttributeValueMemberS).Value == "LEADERBOARD#WEEK#1" &&
						   item["entity_type"].(*types.AttributeValueMemberS).Value == "LEADERBOARD_ENTRY"
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
			name: "missing required fields",
			request: CreateLeaderboardEntryRequest{
				UserID: "", // Missing user ID
				Week:   1,
			},
			method:         "POST",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing week",
			request: CreateLeaderboardEntryRequest{
				UserID:   "user1",
				Username: "testuser",
				Week:     0, // Invalid week
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
			req, _ := http.NewRequest(tt.method, "/leaderboard/entries", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := CreateLeaderboardEntryHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestUpdateLeaderboardEntryHandler(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		week           string
		request        interface{}
		method         string
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:   "successful update",
			userID: "user1",
			week:   "1",
			method: "PUT",
			request: UpdateLeaderboardEntryRequest{
				Points:    intPtr(15),
				Correct:   intPtr(10),
				Incorrect: intPtr(5),
			},
			setupMock: func(mockDB *MockDBClient) {
				// Mock GetItem for existing entry
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					return key["PK"].(*types.AttributeValueMemberS).Value == "USER#user1" &&
						   key["SK"].(*types.AttributeValueMemberS).Value == "LEADERBOARD#WEEK#1"
				})).Return(map[string]types.AttributeValue{
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
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid week format",
			userID:         "user1",
			week:           "invalid",
			method:         "PUT",
			request:        UpdateLeaderboardEntryRequest{},
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "entry not found",
			userID: "user1",
			week:   "1",
			method: "PUT",
			request: UpdateLeaderboardEntryRequest{
				Points: intPtr(15),
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{}, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "database error on get",
			userID: "user1",
			week:   "1",
			method: "PUT",
			request: UpdateLeaderboardEntryRequest{
				Points: intPtr(15),
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:   "invalid data format",
			userID: "user1",
			week:   "1",
			method: "PUT",
			request: UpdateLeaderboardEntryRequest{
				Points: intPtr(15),
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "USER#user1"},
					"SK": &types.AttributeValueMemberS{Value: "LEADERBOARD#WEEK#1"},
					"data": &types.AttributeValueMemberS{Value: "invalid"}, // Wrong type
				}, nil)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest(tt.method, "/leaderboard/entry", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
			// Use mux to set path variables
			req = mux.SetURLVars(req, map[string]string{
				"userId": tt.userID,
				"week":   tt.week,
			})

			rr := httptest.NewRecorder()
			handler := UpdateLeaderboardEntryHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestDeleteLeaderboardEntryHandler(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		week           string
		method         string
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:   "successful delete",
			userID: "user1",
			week:   "1",
			method: "DELETE",
			setupMock: func(mockDB *MockDBClient) {
				// Mock GetItem to check if entry exists
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					return key["PK"].(*types.AttributeValueMemberS).Value == "USER#user1" &&
						   key["SK"].(*types.AttributeValueMemberS).Value == "LEADERBOARD#WEEK#1"
				})).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "USER#user1"},
					"SK": &types.AttributeValueMemberS{Value: "LEADERBOARD#WEEK#1"},
				}, nil)
				// Mock DeleteItem
				mockDB.On("DeleteItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					return key["PK"].(*types.AttributeValueMemberS).Value == "USER#user1" &&
						   key["SK"].(*types.AttributeValueMemberS).Value == "LEADERBOARD#WEEK#1"
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid week format",
			userID:         "user1",
			week:           "invalid",
			method:         "DELETE",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "entry not found",
			userID: "user1",
			week:   "1",
			method: "DELETE",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{}, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "database error on check",
			userID: "user1",
			week:   "1",
			method: "DELETE",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "database error on delete",
			userID: "user1",
			week:   "1",
			method: "DELETE",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "USER#user1"},
				}, nil)
				mockDB.On("DeleteItem", mock.Anything, "FootballLeague", mock.Anything).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req, _ := http.NewRequest(tt.method, "/leaderboard/entry", nil)
			// Use mux to set path variables
			req = mux.SetURLVars(req, map[string]string{
				"userId": tt.userID,
				"week":   tt.week,
			})

			rr := httptest.NewRecorder()
			handler := DeleteLeaderboardEntryHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

// Helper function for creating int pointers
func intPtr(i int) *int {
	return &i
}

// TestGetLeaderboardEntriesByWeekHandler tests the GetLeaderboardEntriesByWeekHandler
func TestGetLeaderboardEntriesByWeekHandler(t *testing.T) {
	tests := []struct {
		name           string
		week           string
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "successful fetch by week",
			week: "1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("QueryItems", mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
					return *input.TableName == "FootballLeague" &&
						   *input.IndexName == "GSI2-UserPicks" &&
						   *input.KeyConditionExpression == "GSI2_PK = :pk"
				})).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{
							"PK": &types.AttributeValueMemberS{Value: "USER#user1"},
							"SK": &types.AttributeValueMemberS{Value: "LEADERBOARD#WEEK#1"},
							"GSI2_PK": &types.AttributeValueMemberS{Value: "LEADERBOARD_WEEK#1"},
							"data": &types.AttributeValueMemberM{
								Value: map[string]types.AttributeValue{
									"user_id":    &types.AttributeValueMemberS{Value: "user1"},
									"username":   &types.AttributeValueMemberS{Value: "testuser1"},
									"points":     &types.AttributeValueMemberN{Value: "10"},
									"correct":    &types.AttributeValueMemberN{Value: "8"},
									"incorrect":  &types.AttributeValueMemberN{Value: "2"},
									"week":       &types.AttributeValueMemberN{Value: "1"},
									"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
									"updated_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
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
			name: "invalid week format",
			week: "invalid",
			setupMock: func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name: "database error",
			week: "1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("QueryItems", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req := httptest.NewRequest("GET", "/leaderboard/week", nil)
			// Use mux to set path variables
			req = mux.SetURLVars(req, map[string]string{
				"week": tt.week,
			})
			rr := httptest.NewRecorder()

			handler := GetLeaderboardEntriesByWeekHandler(mockDB)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedCount > 0 {
				var response LeaderboardResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.Entries, tt.expectedCount)
			}
			mockDB.AssertExpectations(t)
		})
	}
}

