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

func TestGetLeagueHandler(t *testing.T) {
	tests := []struct {
		name           string
		leagueID       string
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:     "successful fetch",
			leagueID: "league1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					return key["PK"].(*types.AttributeValueMemberS).Value == "LEAGUE#league1" &&
						   key["SK"].(*types.AttributeValueMemberS).Value == "METADATA#league1"
				})).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
					"SK": &types.AttributeValueMemberS{Value: "METADATA#league1"},
					"entity_type": &types.AttributeValueMemberS{Value: "LEAGUE"},
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
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing league id",
			leagueID:       "",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "league not found",
			leagueID: "nonexistent",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{}, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "database error",
			leagueID: "league1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:     "invalid data format",
			leagueID: "league1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
					"SK": &types.AttributeValueMemberS{Value: "METADATA#league1"},
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

			req, _ := http.NewRequest("GET", "/league", nil)
			q := req.URL.Query()
			if tt.leagueID != "" {
				q.Add("id", tt.leagueID)
			}
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()
			handler := GetLeagueHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetAllLeaguesHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "successful fetch all leagues",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("QueryItems", mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
					return *input.TableName == "FootballLeague" &&
						   *input.IndexName == "GSI-EntityType"
				})).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{
							"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
							"SK": &types.AttributeValueMemberS{Value: "METADATA#league1"},
							"entity_type": &types.AttributeValueMemberS{Value: "LEAGUE"},
							"data": &types.AttributeValueMemberM{
								Value: map[string]types.AttributeValue{
									"id":         &types.AttributeValueMemberS{Value: "league1"},
									"name":       &types.AttributeValueMemberS{Value: "Test League 1"},
									"admin_id":   &types.AttributeValueMemberS{Value: "admin1"},
									"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
									"updated_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
								},
							},
						},
						{
							"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league2"},
							"SK": &types.AttributeValueMemberS{Value: "METADATA#league2"},
							"entity_type": &types.AttributeValueMemberS{Value: "LEAGUE"},
							"data": &types.AttributeValueMemberM{
								Value: map[string]types.AttributeValue{
									"id":         &types.AttributeValueMemberS{Value: "league2"},
									"name":       &types.AttributeValueMemberS{Value: "Test League 2"},
									"admin_id":   &types.AttributeValueMemberS{Value: "admin2"},
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

			req, _ := http.NewRequest("GET", "/leagues", nil)
			rr := httptest.NewRecorder()
			handler := GetAllLeaguesHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedCount > 0 {
				var response LeaguesResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.Leagues, tt.expectedCount)
			}
			mockDB.AssertExpectations(t)
		})
	}
}

func TestCreateLeagueHandler(t *testing.T) {
	validReq := CreateLeagueRequest{
		Name:    "Test League",
		AdminID: "admin1",
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
					return item["PK"].(*types.AttributeValueMemberS).Value == "LEAGUE#TEST_LEAGUE" &&
						   item["SK"].(*types.AttributeValueMemberS).Value == "METADATA#TEST_LEAGUE" &&
						   item["entity_type"].(*types.AttributeValueMemberS).Value == "LEAGUE"
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
			request: CreateLeagueRequest{
				Name: "", // Missing name
				AdminID: "admin1",
			},
			method:         "POST",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing admin ID",
			request: CreateLeagueRequest{
				Name:    "Test League",
				AdminID: "", // Missing admin ID
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
		{
			name:           "method not allowed",
			request:        validReq,
			method:         "GET",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest(tt.method, "/leagues", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := CreateLeagueHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestUpdateLeagueHandler(t *testing.T) {
	tests := []struct {
		name           string
		leagueID       string
		request        interface{}
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:     "successful update",
			leagueID: "league1",
			request: UpdateLeagueRequest{
				Name: stringPtr("Updated League Name"),
			},
			setupMock: func(mockDB *MockDBClient) {
				// Mock GetItem for existing league
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					return key["PK"].(*types.AttributeValueMemberS).Value == "LEAGUE#league1" &&
						   key["SK"].(*types.AttributeValueMemberS).Value == "METADATA#league1"
				})).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
					"SK": &types.AttributeValueMemberS{Value: "METADATA#league1"},
					"admin_id": &types.AttributeValueMemberS{Value: "admin1"},
					"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"id":         &types.AttributeValueMemberS{Value: "league1"},
							"name":       &types.AttributeValueMemberS{Value: "Old League Name"},
							"admin_id":   &types.AttributeValueMemberS{Value: "admin1"},
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
			name:           "missing league id",
			leagueID:       "",
			request:        UpdateLeagueRequest{},
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "league not found",
			leagueID: "nonexistent",
			request: UpdateLeagueRequest{
				Name: stringPtr("Updated Name"),
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{}, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "database error on get",
			leagueID: "league1",
			request: UpdateLeagueRequest{
				Name: stringPtr("Updated Name"),
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:     "no fields to update",
			leagueID: "league1",
			request:  UpdateLeagueRequest{}, // Empty request
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
				}, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "invalid data format",
			leagueID: "league1",
			request: UpdateLeagueRequest{
				Name: stringPtr("Updated Name"),
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
					"SK": &types.AttributeValueMemberS{Value: "METADATA#league1"},
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
			req, _ := http.NewRequest("PUT", "/league", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
			// Use mux to set path variables instead of query parameters
			req = mux.SetURLVars(req, map[string]string{
				"id": tt.leagueID,
			})

			rr := httptest.NewRecorder()
			handler := UpdateLeagueHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestDeleteLeagueHandler(t *testing.T) {
	tests := []struct {
		name           string
		leagueID       string
		method         string
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:     "successful delete",
			leagueID: "league1",
			method:   "DELETE",
			setupMock: func(mockDB *MockDBClient) {
				// Mock GetItem to check if league exists
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					return key["PK"].(*types.AttributeValueMemberS).Value == "LEAGUE#league1" &&
						   key["SK"].(*types.AttributeValueMemberS).Value == "METADATA#league1"
				})).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
					"SK": &types.AttributeValueMemberS{Value: "METADATA#league1"},
				}, nil)
				// Mock DeleteItem
				mockDB.On("DeleteItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					return key["PK"].(*types.AttributeValueMemberS).Value == "LEAGUE#league1" &&
						   key["SK"].(*types.AttributeValueMemberS).Value == "METADATA#league1"
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing league id",
			leagueID:       "",
			method:         "DELETE",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:     "league not found",
			leagueID: "nonexistent",
			method:   "DELETE",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:     "database error on delete",
			leagueID: "league1",
			method:   "DELETE",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "LEAGUE#league1"},
				}, nil)
				mockDB.On("DeleteItem", mock.Anything, "FootballLeague", mock.Anything).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "method not allowed",
			leagueID:       "league1",
			method:         "GET",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req, _ := http.NewRequest(tt.method, "/league", nil)
			q := req.URL.Query()
			if tt.leagueID != "" {
				q.Add("id", tt.leagueID)
			}
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()
			handler := DeleteLeagueHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

