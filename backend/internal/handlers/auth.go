package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/ckinger23/mountaintop/internal/app"
	"github.com/ckinger23/mountaintop/internal/middleware"
	"github.com/ckinger23/mountaintop/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

// Register returns a handler for user registration
// handlers use a closure pattern to inject dependencies while maintaining the signature
func Register(a *app.App) http.HandlerFunc {
	// All http.HandlerFunc's must have this exact signature:
	// An http.ResponseWriter and an *http.Request
	// responseWriter to write the response body, set headers, set status code
	// Request to get HTTP method, get URL params(with Chi), get query params, get headers, read body, get Context
	return func(w http.ResponseWriter, r *http.Request) {
		var req RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Validate input
		if req.Username == "" || req.Email == "" || req.Password == "" {
			http.Error(w, "Username, email, and password are required", http.StatusBadRequest)
			return
		}

		// Hash password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Error processing password", http.StatusInternalServerError)
			return
		}

		// Create user
		user := models.User{
			Username:     req.Username,
			Email:        req.Email,
			PasswordHash: string(hashedPassword),
			DisplayName:  req.DisplayName,
			IsAdmin:      false, // First user could be admin, implement logic as needed
		}

		if err := a.DB.Create(&user).Error; err != nil {
			http.Error(w, "Username or email already exists", http.StatusConflict)
			return
		}

		// Generate token
		token, err := middleware.GenerateToken(user.ID, user.Email, user.IsAdmin)
		if err != nil {
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		// Return response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthResponse{
			Token: token,
			User:  user,
		})
	}
}

// Login returns a handler for user authentication
func Login(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Find user by email
		var user models.User
		if err := a.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Generate token
		token, err := middleware.GenerateToken(user.ID, user.Email, user.IsAdmin)
		if err != nil {
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		// Return response
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(AuthResponse{
			Token: token,
			User:  user,
		})
	}
}

// GetCurrentUser returns a handler that returns the authenticated user's information
func GetCurrentUser(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := middleware.GetUserFromContext(r)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var user models.User
		if err := a.DB.First(&user, claims.UserID).Error; err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Validate that token claims match current database state
		// This handles cases where user permissions changed after token was issued
		if user.Email != claims.Email || user.IsAdmin != claims.IsAdmin {
			http.Error(w, "Token claims outdated, please login again", http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}
