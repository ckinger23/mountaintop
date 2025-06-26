package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
				mockDB.On("Scan", mock.Anything, &dynamodb.ScanInput{
					TableName: aws.String("LeaderboardEntries"),
				}).Return(&dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{
						{"user_id": &types.AttributeValueMemberS{Value: "user1"}},
						{"user_id": &types.AttributeValueMemberS{Value: "user2"}},
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
				mockDB.On("GetItem", mock.Anything, mock.Anything).Return(map[string]types.AttributeValue{
					"user_id": &types.AttributeValueMemberS{Value: "user1"},
					"week":    &types.AttributeValueMemberN{Value: "1"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user id",
			userID:         "",
			week:           "1",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req, _ := http.NewRequest("GET", "/leaderboard/entry", nil)
			q := req.URL.Query()
			if tt.userID != "" {
				q.Add("user_id", tt.userID)
			}
			if tt.week != "" {
				q.Add("week", tt.week)
			}
			req.URL.RawQuery = q.Encode()

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
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:    "successful create",
			request: validReq,
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("PutItem", mock.Anything, "LeaderboardEntries", mock.Anything).
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
			req, _ := http.NewRequest("POST", "/leaderboard/entries", strings.NewReader(string(body)))
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
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:   "successful update",
			userID: "user1",
			week:   "1",
			request: map[string]int{
				"points":    15,
				"correct":   10,
				"incorrect": 5,
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, mock.Anything).Return(map[string]types.AttributeValue{
					"user_id": &types.AttributeValueMemberS{Value: "user1"},
					"week":    &types.AttributeValueMemberN{Value: "1"},
				}, nil)
				mockDB.On("PutItem", mock.Anything, "LeaderboardEntries", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user id",
			userID:         "",
			week:           "1",
			request:        map[string]int{},
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			body, _ := json.Marshal(tt.request)
			req, _ := http.NewRequest("PUT", "/leaderboard/entry", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
			q := req.URL.Query()
			if tt.userID != "" {
				q.Add("user_id", tt.userID)
			}
			if tt.week != "" {
				q.Add("week", tt.week)
			}
			req.URL.RawQuery = q.Encode()

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
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:   "successful delete",
			userID: "user1",
			week:   "1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("DeleteItem", mock.Anything, "LeaderboardEntries", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "missing user id",
			userID:         "",
			week:           "1",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req, _ := http.NewRequest("DELETE", "/leaderboard/entry", nil)
			q := req.URL.Query()
			if tt.userID != "" {
				q.Add("user_id", tt.userID)
			}
			if tt.week != "" {
				q.Add("week", tt.week)
			}
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()
			handler := DeleteLeaderboardEntryHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}
