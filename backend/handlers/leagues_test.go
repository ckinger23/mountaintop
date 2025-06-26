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
				mockDB.On("GetItem", mock.Anything, mock.Anything).Return(map[string]types.AttributeValue{
					"id":         &types.AttributeValueMemberS{Value: "league1"},
					"name":       &types.AttributeValueMemberS{Value: "Test League"},
					"admin_id":   &types.AttributeValueMemberS{Value: "admin1"},
					"created_at": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
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
				mockDB.On("Query", mock.Anything, mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
					return *input.TableName == "FootballLeague"
				})).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{"id": &types.AttributeValueMemberS{Value: "league1"}},
						{"id": &types.AttributeValueMemberS{Value: "league2"}},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
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
			req, _ := http.NewRequest("POST", "/leagues", strings.NewReader(string(body)))
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
			request: map[string]string{
				"name": "Updated League Name",
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, mock.Anything).Return(map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "league1"},
				}, nil)
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing league id",
			leagueID:       "",
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
			req, _ := http.NewRequest("PUT", "/league", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
			q := req.URL.Query()
			if tt.leagueID != "" {
				q.Add("id", tt.leagueID)
			}
			req.URL.RawQuery = q.Encode()

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
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:     "successful delete",
			leagueID: "league1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("DeleteItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "missing league id",
			leagueID:       "",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req, _ := http.NewRequest("DELETE", "/league", nil)
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
