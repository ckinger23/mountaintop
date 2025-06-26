package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// MockDBClient is a mock implementation of the DatabaseClient interface
type MockDBClient struct {
	mock.Mock
}

// NewMockDBClient creates a new instance of MockDBClient with properly initialized mock
func NewMockDBClient() *MockDBClient {
	return &MockDBClient{
		Mock: mock.Mock{},
	}
}

// AssertExpectations asserts that everything specified with On and Return was in fact called as expected.
// Calls may have occurred in any order.
func (m *MockDBClient) AssertExpectations(t mock.TestingT) bool {
	return m.Mock.AssertExpectations(t)
}

func (m *MockDBClient) PutItem(ctx context.Context, tableName string, item map[string]types.AttributeValue) error {
	args := m.Called(ctx, tableName, item)
	return args.Error(0)
}

func (m *MockDBClient) GetItem(ctx context.Context, tableName string, key map[string]types.AttributeValue) (map[string]types.AttributeValue, error) {
	args := m.Called(ctx, tableName, key)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]types.AttributeValue), args.Error(1)
}

func (m *MockDBClient) DeleteItem(ctx context.Context, tableName string, key map[string]types.AttributeValue) error {
	args := m.Called(ctx, tableName, key)
	return args.Error(0)
}

func (m *MockDBClient) QueryItems(ctx context.Context, input *dynamodb.QueryInput) (*dynamodb.QueryOutput, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.QueryOutput), args.Error(1)
}

func (m *MockDBClient) Scan(ctx context.Context, input *dynamodb.ScanInput) (*dynamodb.ScanOutput, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.ScanOutput), args.Error(1)
}

func TestGetConferenceHandler(t *testing.T) {
	tests := []struct {
		name           string
		conferenceID   string
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectError    bool
	}{
		{
			name:         "successful get conference by id",
			conferenceID: "SEC",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					pk := key["PK"].(*types.AttributeValueMemberS).Value
					sk := key["SK"].(*types.AttributeValueMemberS).Value
					return pk == "CONFERENCE#SEC" && sk == "METADATA"
				})).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "CONFERENCE#SEC"},
					"SK": &types.AttributeValueMemberS{Value: "METADATA"},
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"id":         &types.AttributeValueMemberS{Value: "SEC"},
							"name":       &types.AttributeValueMemberS{Value: "SEC"},
							"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
							"updated_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
						},
					},
					"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
					"updated_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "missing conference id",
			conferenceID:   "",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:         "conference not found",
			conferenceID: "NONEXISTENT",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{}, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:         "database error",
			conferenceID: "SEC",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:         "invalid data format",
			conferenceID: "SEC",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"data": &types.AttributeValueMemberS{Value: "invalid-format"},
				}, nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:         "malformed data attribute",
			conferenceID: "SEC",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"id":         &types.AttributeValueMemberS{Value: "SEC"},
							"name":       &types.AttributeValueMemberN{Value: "123"}, // Wrong type but attributevalue.UnmarshalMap handles it
							"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
							"updated_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK, // attributevalue.UnmarshalMap converts number to string
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			url := "/conferences"
			if tt.conferenceID != "" {
				url = fmt.Sprintf("/conferences?id=%s", tt.conferenceID)
			}
			req := httptest.NewRequest("GET", url, nil)
			rr := httptest.NewRecorder()

			handler := GetConferenceHandler(mockDB)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				var response ConferenceResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.conferenceID, response.ID)
				assert.NotEmpty(t, response.Name)
			}

			if tt.expectError {
				assert.Contains(t, rr.Body.String(), "error")
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetAllConferencesHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectedCount  int
		expectError    bool
	}{
		{
			name: "successful fetch all conferences",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("QueryItems", mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
					return *input.TableName == "FootballLeague" &&
						*input.IndexName == "GSI1-EntityLookup"
				})).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{
							"PK": &types.AttributeValueMemberS{Value: "CONFERENCE#SEC"},
							"SK": &types.AttributeValueMemberS{Value: "METADATA"},
							"data": &types.AttributeValueMemberM{
								Value: map[string]types.AttributeValue{
									"id":         &types.AttributeValueMemberS{Value: "SEC"},
									"name":       &types.AttributeValueMemberS{Value: "SEC"},
									"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
									"updated_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
								},
							},
						},
						{
							"PK": &types.AttributeValueMemberS{Value: "CONFERENCE#B1G"},
							"SK": &types.AttributeValueMemberS{Value: "METADATA"},
							"data": &types.AttributeValueMemberM{
								Value: map[string]types.AttributeValue{
									"id":         &types.AttributeValueMemberS{Value: "B1G"},
									"name":       &types.AttributeValueMemberS{Value: "Big Ten"},
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
			expectError:    false,
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
			expectError:    false,
		},
		{
			name: "database error",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("QueryItems", mock.Anything).Return(nil, errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
			expectError:    true,
		},
		{
			name: "partial data corruption",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("QueryItems", mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{
							"PK": &types.AttributeValueMemberS{Value: "CONFERENCE#SEC"},
							"SK": &types.AttributeValueMemberS{Value: "METADATA"},
							"data": &types.AttributeValueMemberM{
								Value: map[string]types.AttributeValue{
									"id":         &types.AttributeValueMemberS{Value: "SEC"},
									"name":       &types.AttributeValueMemberS{Value: "SEC"},
									"created_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
									"updated_at": &types.AttributeValueMemberS{Value: "2023-01-01T00:00:00Z"},
								},
							},
						},
						{
							"PK": &types.AttributeValueMemberS{Value: "CONFERENCE#CORRUPT"},
							"SK": &types.AttributeValueMemberS{Value: "METADATA"},
							"data": &types.AttributeValueMemberS{Value: "invalid-format"},
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1, // Only valid record should be returned
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req := httptest.NewRequest("GET", "/conferences", nil)
			rr := httptest.NewRecorder()

			handler := GetAllConferencesHandler(mockDB)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				var response ConferencesResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.Conferences, tt.expectedCount)

				// Verify structure when we have conferences
				if tt.expectedCount > 0 {
					conf := response.Conferences[0]
					assert.NotEmpty(t, conf.ID)
					assert.NotEmpty(t, conf.Name)
				}
			}

			if tt.expectError {
				assert.Contains(t, rr.Body.String(), "error")
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestCreateConferenceHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		request        interface{}
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectError    bool
	}{
		{
			name:   "successful create",
			method: "POST",
			request: CreateConferenceRequest{
				Name: "Pac-12",
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(item map[string]types.AttributeValue) bool {
					pk := item["PK"].(*types.AttributeValueMemberS).Value
					sk := item["SK"].(*types.AttributeValueMemberS).Value
					entityType := item["entity_type"].(*types.AttributeValueMemberS).Value
					return strings.HasPrefix(pk, "CONFERENCE#") && sk == "METADATA" && entityType == "CONFERENCE"
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "wrong method",
			method:         "GET",
			request:        CreateConferenceRequest{Name: "Test"},
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
		},
		{
			name:           "invalid JSON",
			method:         "POST",
			request:        "invalid json",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:   "missing name",
			method: "POST",
			request: CreateConferenceRequest{
				Name: "",
			},
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:   "whitespace only name",
			method: "POST",
			request: CreateConferenceRequest{
				Name: "   ",
			},
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:   "database error on create",
			method: "POST",
			request: CreateConferenceRequest{
				Name: "Test Conference",
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:   "special characters in name",
			method: "POST",
			request: CreateConferenceRequest{
				Name: "Test & Special @ Conference",
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			var body string
			if str, ok := tt.request.(string); ok {
				body = str
			} else {
				bodyBytes, _ := json.Marshal(tt.request)
				body = string(bodyBytes)
			}

			req := httptest.NewRequest(tt.method, "/conferences", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler := CreateConferenceHandler(mockDB)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectError && tt.expectedStatus == http.StatusCreated {
				var response ConferenceResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.ID)

				if createReq, ok := tt.request.(CreateConferenceRequest); ok {
					assert.Equal(t, createReq.Name, response.Name)
				}
				assert.NotZero(t, response.CreatedAt)
				assert.NotZero(t, response.UpdatedAt)
			}

			if tt.expectError {
				assert.Contains(t, rr.Body.String(), "error")
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestUpdateConferenceHandler(t *testing.T) {
	fixedTime := time.Now()
	tests := []struct {
		name           string
		conferenceID   string
		request        interface{}
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectError    bool
	}{
		{
			name:         "successful update",
			conferenceID: "SEC",
			request: CreateConferenceRequest{
				Name: "Updated SEC",
			},
			setupMock: func(mockDB *MockDBClient) {
				// Mock existing conference fetch
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "CONFERENCE#SEC"},
					"SK": &types.AttributeValueMemberS{Value: "METADATA"},
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"name": &types.AttributeValueMemberS{Value: "SEC"},
						},
					},
					"created_at": &types.AttributeValueMemberS{Value: fixedTime.Format(time.RFC3339)},
				}, nil)

				// Mock update
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "missing conference id",
			conferenceID:   "",
			request:        CreateConferenceRequest{Name: "Test"},
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:         "conference not found",
			conferenceID: "NONEXISTENT",
			request:      CreateConferenceRequest{Name: "Test"},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:         "invalid JSON",
			conferenceID: "SEC",
			request:      "invalid json",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"name": &types.AttributeValueMemberS{Value: "SEC"},
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:         "missing name in update",
			conferenceID: "SEC",
			request: CreateConferenceRequest{
				Name: "",
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"name": &types.AttributeValueMemberS{Value: "SEC"},
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:         "invalid existing data format",
			conferenceID: "SEC",
			request:      CreateConferenceRequest{Name: "Updated"},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"data": &types.AttributeValueMemberS{Value: "invalid-format"},
				}, nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:         "database error on update",
			conferenceID: "SEC",
			request:      CreateConferenceRequest{Name: "Updated"},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"name": &types.AttributeValueMemberS{Value: "SEC"},
						},
					},
					"created_at": &types.AttributeValueMemberS{Value: fixedTime.Format(time.RFC3339)},
				}, nil)
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			var body string
			if str, ok := tt.request.(string); ok {
				body = str
			} else {
				bodyBytes, _ := json.Marshal(tt.request)
				body = string(bodyBytes)
			}

			req := httptest.NewRequest("PUT", "/conferences/"+tt.conferenceID, strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			
			// Set up mux vars
			req = mux.SetURLVars(req, map[string]string{"id": tt.conferenceID})
			rr := httptest.NewRecorder()

			handler := UpdateConferenceHandler(mockDB)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				var response ConferenceResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.conferenceID, response.ID)

				if updateReq, ok := tt.request.(CreateConferenceRequest); ok {
					assert.Equal(t, updateReq.Name, response.Name)
				}
			}

			if tt.expectError {
				assert.Contains(t, rr.Body.String(), "error")
			}

			mockDB.AssertExpectations(t)
		})
	}
}

func TestDeleteConferenceHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		conferenceID   string
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectError    bool
	}{
		{
			name:         "successful delete",
			method:       "DELETE",
			conferenceID: "SEC",
			setupMock: func(mockDB *MockDBClient) {
				// Mock existence check
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "CONFERENCE#SEC"},
				}, nil)

				// Mock delete
				mockDB.On("DeleteItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "wrong method",
			method:         "GET",
			conferenceID:   "SEC",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusMethodNotAllowed,
			expectError:    true,
		},
		{
			name:           "missing conference id",
			method:         "DELETE",
			conferenceID:   "",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:         "conference not found",
			method:       "DELETE",
			conferenceID: "NONEXISTENT",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:         "database error on delete",
			method:       "DELETE",
			conferenceID: "SEC",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "CONFERENCE#SEC"},
				}, nil)
				mockDB.On("DeleteItem", mock.Anything, "FootballLeague", mock.Anything).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			url := "/conferences"
			if tt.conferenceID != "" {
				url = fmt.Sprintf("/conferences?id=%s", tt.conferenceID)
			}
			req := httptest.NewRequest(tt.method, url, nil)
			rr := httptest.NewRecorder()

			handler := DeleteConferenceHandler(mockDB)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				var response map[string]string
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response["message"], "deleted successfully")
			}

			if tt.expectError {
				assert.Contains(t, rr.Body.String(), "error")
			}

			mockDB.AssertExpectations(t)
		})
	}
}