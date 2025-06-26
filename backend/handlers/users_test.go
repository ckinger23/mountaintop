package handlers

import (
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetAllUsersHandler(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectedCount  int
		expectError    bool
	}{
		{
			name: "successful fetch all users",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("QueryItems", mock.Anything, mock.MatchedBy(func(input *dynamodb.QueryInput) bool {
					return *input.TableName == "FootballLeague" &&
						*input.IndexName == "GSI1-EntityLookup" &&
						input.ExpressionAttributeValues[":pk"].(*types.AttributeValueMemberS).Value == "USER"
				})).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{
							"PK": &types.AttributeValueMemberS{Value: "USER#1"},
							"SK": &types.AttributeValueMemberS{Value: "PROFILE"},
							"data": &types.AttributeValueMemberM{
								Value: map[string]types.AttributeValue{
									"user_id":  &types.AttributeValueMemberS{Value: "1"},
									"username": &types.AttributeValueMemberS{Value: "testuser1"},
									"email":    &types.AttributeValueMemberS{Value: "test1@example.com"},
									"role":     &types.AttributeValueMemberS{Value: "player"},
								},
							},
						},
						{
							"PK": &types.AttributeValueMemberS{Value: "USER#2"},
							"SK": &types.AttributeValueMemberS{Value: "PROFILE"},
							"data": &types.AttributeValueMemberM{
								Value: map[string]types.AttributeValue{
									"user_id":  &types.AttributeValueMemberS{Value: "2"},
									"username": &types.AttributeValueMemberS{Value: "testuser2"},
									"email":    &types.AttributeValueMemberS{Value: "test2@example.com"},
									"role":     &types.AttributeValueMemberS{Value: "admin"},
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
				mockDB.On("QueryItems", mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{
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
				mockDB.On("QueryItems", mock.Anything, mock.Anything).Return(nil, errors.New("database connection failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
			expectError:    true,
		},
		{
			name: "malformed data attribute",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("QueryItems", mock.Anything, mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{
							"PK": &types.AttributeValueMemberS{Value: "USER#1"},
							"SK": &types.AttributeValueMemberS{Value: "PROFILE"},
							"data": &types.AttributeValueMemberS{Value: "invalid-data-format"},
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			req, _ := http.NewRequest("GET", "/users", nil)
			rr := httptest.NewRecorder()
			handler := GetAllUsersHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			
			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				var response UsersResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Len(t, response.Users, tt.expectedCount)
				
				// Verify structure of response when we have users
				if tt.expectedCount > 0 {
					user := response.Users[0]
					assert.NotEmpty(t, user.ID)
					assert.NotEmpty(t, user.Username)
					assert.NotEmpty(t, user.Email)
					assert.NotEmpty(t, user.Role)
				}
			}
			
			if tt.expectError {
				assert.Contains(t, rr.Body.String(), "error")
			}
			
			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetUserByIDHandler(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectError    bool
	}{
		{
			name:   "successful fetch by id",
			userID: "user1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					pk := key["PK"].(*types.AttributeValueMemberS).Value
					sk := key["SK"].(*types.AttributeValueMemberS).Value
					return pk == "USER#user1" && sk == "PROFILE"
				})).Return(map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "USER#user1"},
					"SK": &types.AttributeValueMemberS{Value: "PROFILE"},
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"user_id":  &types.AttributeValueMemberS{Value: "user1"},
							"username": &types.AttributeValueMemberS{Value: "testuser"},
							"email":    &types.AttributeValueMemberS{Value: "test@example.com"},
							"role":     &types.AttributeValueMemberS{Value: "player"},
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "missing user id",
			userID:         "",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{}, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:   "database error",
			userID: "user1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:   "invalid data format",
			userID: "user1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"data": &types.AttributeValueMemberS{Value: "invalid"},
				}, nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			url := "/users"
			if tt.userID != "" {
				url = fmt.Sprintf("/users?id=%s", tt.userID)
			}
			req, _ := http.NewRequest("GET", url, nil)
			rr := httptest.NewRecorder()
			handler := GetUserByIDHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			
			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				var response UserResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.userID, response.ID)
				assert.NotEmpty(t, response.Username)
				assert.NotEmpty(t, response.Email)
				assert.NotEmpty(t, response.Role)
			}
			
			if tt.expectError {
				assert.Contains(t, rr.Body.String(), "error")
			}
			
			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetUserByUsernameHandler(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectError    bool
	}{
		{
			name:     "successful fetch by username",
			username: "testuser",
			setupMock: func(mockDB *MockDBClient) {
				// Mock the username lookup
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					pk := key["PK"].(*types.AttributeValueMemberS).Value
					sk := key["SK"].(*types.AttributeValueMemberS).Value
					return pk == "USERNAME#testuser" && sk == "LOOKUP"
				})).Return(map[string]types.AttributeValue{
					"user_id": &types.AttributeValueMemberS{Value: "user1"},
				}, nil)

				// Mock the user data fetch
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					pk := key["PK"].(*types.AttributeValueMemberS).Value
					sk := key["SK"].(*types.AttributeValueMemberS).Value
					return pk == "USER#user1" && sk == "PROFILE"
				})).Return(map[string]types.AttributeValue{
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"user_id":  &types.AttributeValueMemberS{Value: "user1"},
							"username": &types.AttributeValueMemberS{Value: "testuser"},
							"email":    &types.AttributeValueMemberS{Value: "test@example.com"},
							"role":     &types.AttributeValueMemberS{Value: "player"},
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "missing username",
			username:       "",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:     "username not found",
			username: "nonexistent",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{}, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:     "username lookup error",
			username: "testuser",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil, errors.New("lookup error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
		{
			name:     "invalid lookup data",
			username: "testuser",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"invalid": &types.AttributeValueMemberS{Value: "data"},
				}, nil)
			},
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			url := "/users/username"
			if tt.username != "" {
				url = fmt.Sprintf("/users/username?username=%s", tt.username)
			}
			req, _ := http.NewRequest("GET", url, nil)
			rr := httptest.NewRecorder()
			handler := GetUserByUsernameHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			
			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				var response UserResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.username, response.Username)
			}
			
			if tt.expectError {
				assert.Contains(t, rr.Body.String(), "error")
			}
			
			mockDB.AssertExpectations(t)
		})
	}
}

func TestGetUserByEmailHandler(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectError    bool
	}{
		{
			name:  "successful fetch by email",
			email: "test@example.com",
			setupMock: func(mockDB *MockDBClient) {
				// Mock the email lookup
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					pk := key["PK"].(*types.AttributeValueMemberS).Value
					sk := key["SK"].(*types.AttributeValueMemberS).Value
					return pk == "EMAIL#test@example.com" && sk == "LOOKUP"
				})).Return(map[string]types.AttributeValue{
					"user_id": &types.AttributeValueMemberS{Value: "user1"},
				}, nil)

				// Mock the user data fetch
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					pk := key["PK"].(*types.AttributeValueMemberS).Value
					sk := key["SK"].(*types.AttributeValueMemberS).Value
					return pk == "USER#user1" && sk == "PROFILE"
				})).Return(map[string]types.AttributeValue{
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"user_id":  &types.AttributeValueMemberS{Value: "user1"},
							"username": &types.AttributeValueMemberS{Value: "testuser"},
							"email":    &types.AttributeValueMemberS{Value: "test@example.com"},
							"role":     &types.AttributeValueMemberS{Value: "player"},
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "missing email",
			email:          "",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:  "email not found",
			email: "nonexistent@example.com",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{}, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:  "case insensitive email",
			email: "Test@Example.COM",
			setupMock: func(mockDB *MockDBClient) {
				// Should lookup with lowercase email
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					pk := key["PK"].(*types.AttributeValueMemberS).Value
					return pk == "EMAIL#test@example.com"
				})).Return(map[string]types.AttributeValue{
					"user_id": &types.AttributeValueMemberS{Value: "user1"},
				}, nil)

				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(key map[string]types.AttributeValue) bool {
					pk := key["PK"].(*types.AttributeValueMemberS).Value
					sk := key["SK"].(*types.AttributeValueMemberS).Value
					return pk == "USER#user1" && sk == "PROFILE"
				})).Return(map[string]types.AttributeValue{
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"user_id":  &types.AttributeValueMemberS{Value: "user1"},
							"username": &types.AttributeValueMemberS{Value: "testuser"},
							"email":    &types.AttributeValueMemberS{Value: "test@example.com"},
							"role":     &types.AttributeValueMemberS{Value: "player"},
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			url := "/users/email"
			if tt.email != "" {
				url = fmt.Sprintf("/users/email?email=%s", tt.email)
			}
			req, _ := http.NewRequest("GET", url, nil)
			rr := httptest.NewRecorder()
			handler := GetUserByEmailHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			
			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				var response UserResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, strings.ToLower(tt.email), response.Email)
			}
			
			if tt.expectError {
				assert.Contains(t, rr.Body.String(), "error")
			}
			
			mockDB.AssertExpectations(t)
		})
	}
}

func TestCreateUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		request        interface{}
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectError    bool
	}{
		{
			name: "successful create",
			request: CreateUserRequest{
				Username: "newuser",
				Email:    "new@example.com",
				Role:     "player",
			},
			setupMock: func(mockDB *MockDBClient) {
				// Mock email uniqueness check
				mockDB.On("Scan", mock.Anything, mock.Anything).Return(&dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{},
				}, nil)

				// Mock user creation (main record)
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(item map[string]types.AttributeValue) bool {
					pk := item["PK"].(*types.AttributeValueMemberS).Value
					sk := item["SK"].(*types.AttributeValueMemberS).Value
					return strings.HasPrefix(pk, "USER#") && sk == "PROFILE"
				})).Return(nil)

				// Mock username lookup creation
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(item map[string]types.AttributeValue) bool {
					pk := item["PK"].(*types.AttributeValueMemberS).Value
					return pk == "USERNAME#newuser"
				})).Return(nil)

				// Mock email lookup creation
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.MatchedBy(func(item map[string]types.AttributeValue) bool {
					pk := item["PK"].(*types.AttributeValueMemberS).Value
					return pk == "EMAIL#new@example.com"
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectError:    false,
		},
		{
			name:           "invalid JSON",
			request:        "invalid json",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "missing username",
			request: CreateUserRequest{
				Email: "test@example.com",
				Role:  "player",
			},
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "missing email",
			request: CreateUserRequest{
				Username: "testuser",
				Role:     "player",
			},
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "missing role",
			request: CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
			},
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "email already exists",
			request: CreateUserRequest{
				Username: "newuser",
				Email:    "existing@example.com",
				Role:     "player",
			},
			setupMock: func(mockDB *MockDBClient) {
				// Mock existing user found
				mockDB.On("Scan", mock.Anything, mock.Anything).Return(&dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{
						{
							"data": &types.AttributeValueMemberM{
								Value: map[string]types.AttributeValue{
									"email": &types.AttributeValueMemberS{Value: "existing@example.com"},
								},
							},
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusConflict,
			expectError:    true,
		},
		{
			name: "database error on create",
			request: CreateUserRequest{
				Username: "newuser",
				Email:    "new@example.com",
				Role:     "player",
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("Scan", mock.Anything, mock.Anything).Return(&dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{},
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

			req, _ := http.NewRequest("POST", "/users", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			handler := CreateUserHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			
			if !tt.expectError && tt.expectedStatus == http.StatusCreated {
				var response UserResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				
				if createReq, ok := tt.request.(CreateUserRequest); ok {
					assert.Equal(t, createReq.Username, response.Username)
					assert.Equal(t, strings.ToLower(createReq.Email), response.Email)
					assert.Equal(t, createReq.Role, response.Role)
				}
			}
			
			if tt.expectError {
				assert.Contains(t, rr.Body.String(), "error")
			}
			
			mockDB.AssertExpectations(t)
		})
	}
}

func TestUpdateUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		request        interface{}
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectError    bool
	}{
		{
			name:   "successful update username",
			userID: "user1",
			request: UpdateUserRequest{
				Username: stringPtr("updateduser"),
			},
			setupMock: func(mockDB *MockDBClient) {
				// Mock existing user fetch
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"created_at": &types.AttributeValueMemberS{Value: time.Now().Format(time.RFC3339)},
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"username": &types.AttributeValueMemberS{Value: "olduser"},
							"email":    &types.AttributeValueMemberS{Value: "test@example.com"},
							"role":     &types.AttributeValueMemberS{Value: "player"},
						},
					},
				}, nil)

				// Mock update
				mockDB.On("PutItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "missing user id",
			userID:         "",
			request:        UpdateUserRequest{},
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:   "no fields to update",
			userID: "user1",
			request: UpdateUserRequest{},
			setupMock: func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			request: UpdateUserRequest{
				Username: stringPtr("newname"),
			},
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{}, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:   "email conflict",
			userID: "user1",
			request: UpdateUserRequest{
				Email: stringPtr("taken@example.com"),
			},
			setupMock: func(mockDB *MockDBClient) {
				// Mock existing user fetch
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"username": &types.AttributeValueMemberS{Value: "testuser"},
							"email":    &types.AttributeValueMemberS{Value: "old@example.com"},
							"role":     &types.AttributeValueMemberS{Value: "player"},
						},
					},
				}, nil)

				// Mock email conflict check
				mockDB.On("Scan", mock.Anything, mock.Anything).Return(&dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{
						{
							"data": &types.AttributeValueMemberM{
								Value: map[string]types.AttributeValue{
									"email": &types.AttributeValueMemberS{Value: "taken@example.com"},
								},
							},
						},
					},
				}, nil)
			},
			expectedStatus: http.StatusConflict,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDB := NewMockDBClient()
			tt.setupMock(mockDB)

			bodyBytes, _ := json.Marshal(tt.request)
			url := fmt.Sprintf("/users/%s", tt.userID)
			if tt.userID != "" {
				url = fmt.Sprintf("/users?id=%s", tt.userID)
			}
			req, _ := http.NewRequest("PUT", url, strings.NewReader(string(bodyBytes)))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			handler := UpdateUserHandler(mockDB)

			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			
			if !tt.expectError && tt.expectedStatus == http.StatusOK {
				var response UserResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.userID, response.ID)
			}
			
			if tt.expectError {
				assert.Contains(t, rr.Body.String(), "error")
			}
			
			mockDB.AssertExpectations(t)
		})
	}
}

func TestDeleteUserHandler(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMock      func(*MockDBClient)
		expectedStatus int
		expectError    bool
	}{
		{
			name:   "successful delete",
			userID: "user1",
			setupMock: func(mockDB *MockDBClient) {
				// Mock user existence check
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"email": &types.AttributeValueMemberS{Value: "test@example.com"},
						},
					},
				}, nil)

				// Mock delete
				mockDB.On("DeleteItem", mock.Anything, "FootballLeague", mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "missing user id",
			userID:         "",
			setupMock:      func(mockDB *MockDBClient) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:   "user not found",
			userID: "nonexistent",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{}, nil)
			},
			expectedStatus: http.StatusNotFound,
			expectError:    true,
		},
		{
			name:   "database error on delete",
			userID: "user1",
			setupMock: func(mockDB *MockDBClient) {
				mockDB.On("GetItem", mock.Anything, "FootballLeague", mock.Anything).Return(map[string]types.AttributeValue{
					"data": &types.AttributeValueMemberM{
						Value: map[string]types.AttributeValue{
							"email": &types.AttributeValueMemberS{Value: "test@example.com"},
						},
					},
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

			url := fmt.Sprintf("/users/%s", tt.userID)
			if tt.userID != "" {
				url = fmt.Sprintf("/users?id=%s", tt.userID)
			}
			req, _ := http.NewRequest("DELETE", url, nil)
			rr := httptest.NewRecorder()
			handler := DeleteUserHandler(mockDB)

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

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}