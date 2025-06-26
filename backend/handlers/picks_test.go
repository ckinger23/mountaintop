package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSubmitPickHandler(t *testing.T) {
	validReq := PickRequest{
		UserID: "user1",
		GameID: "game1",
		Week:   1,
		Pick:   "home",
	}

	tests := []struct {
		name           string
		request        interface{}
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:    "successful submit",
			request: validReq,
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("Query", mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{},
				}, nil)
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
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
			req, _ := http.NewRequest("POST", "/picks", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := SubmitPickHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetPickHandler(t *testing.T) {
	tests := []struct {
		name           string
		pickID         string
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:   "successful fetch",
			pickID: "pick1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, mock.Anything).Return(map[string]types.AttributeValue{
					"id":      &types.AttributeValueMemberS{Value: "pick1"},
					"user_id": &types.AttributeValueMemberS{Value: "user1"},
					"game_id": &types.AttributeValueMemberS{Value: "game1"},
					"week":    &types.AttributeValueMemberN{Value: "1"},
					"pick":    &types.AttributeValueMemberS{Value: "home"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing pick id",
			pickID:         "",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req, _ := http.NewRequest("GET", "/pick", nil)
			q := req.URL.Query()
			if tt.pickID != "" {
				q.Add("id", tt.pickID)
			}
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()
			handler := GetPickHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetAllPicksByUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectedCount  int
	}{
		{
			name:   "successful fetch by user",
			userID: "user1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("Query", mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{"id": &types.AttributeValueMemberS{Value: "pick1"}},
						{"id": &types.AttributeValueMemberS{Value: "pick2"}},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name:           "missing user id",
			userID:         "",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req, _ := http.NewRequest("GET", "/picks/user", nil)
			q := req.URL.Query()
			if tt.userID != "" {
				q.Add("user_id", tt.userID)
			}
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()
			handler := GetAllPicksByUserHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedCount > 0 {
				var response PicksResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.Picks, tt.expectedCount)
			}
			mockDB.AssertExpectations(t)
		})
	}
}

func TestUpdatePickHandler(t *testing.T) {
	tests := []struct {
		name           string
		pickID         string
		request        interface{}
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:   "successful update",
			pickID: "pick1",
			request: map[string]string{
				"pick": "away",
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, mock.Anything).Return(map[string]types.AttributeValue{
					"id": &types.AttributeValueMemberS{Value: "pick1"},
				}, nil)
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing pick id",
			pickID:         "",
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
			req, _ := http.NewRequest("PUT", "/pick", strings.NewReader(string(body)))
			req.Header.Set("Content-Type", "application/json")
			q := req.URL.Query()
			if tt.pickID != "" {
				q.Add("id", tt.pickID)
			}
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()
			handler := UpdatePickHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}

func TestDeletePickHandler(t *testing.T) {
	tests := []struct {
		name           string
		pickID         string
		setupMock      func(*MockDBClient)
		expectedStatus int
	}{
		{
			name:   "successful delete",
			pickID: "pick1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("DeleteItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "missing pick id",
			pickID:         "",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req, _ := http.NewRequest("DELETE", "/pick", nil)
			q := req.URL.Query()
			if tt.pickID != "" {
				q.Add("id", tt.pickID)
			}
			req.URL.RawQuery = q.Encode()

			rr := httptest.NewRecorder()
			handler := DeletePickHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			mockDB.AssertExpectations(t)
		})
	}
}
