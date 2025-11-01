package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtSecret = []byte("your-secret-key-change-this") // TODO: Move to environment variable

type Claims struct {
	UserID        uint   `json:"user_id"`
	Email         string `json:"email"`
	IsAdmin       bool   `json:"is_admin"`        // Legacy field, kept for backward compatibility
	IsGlobalAdmin bool   `json:"is_global_admin"` // Superuser with all permissions
	jwt.RegisteredClaims
}

type contextKey string

const UserContextKey contextKey = "user"

// GenerateToken creates a JWT token for a user
func GenerateToken(userID uint, email string, isAdmin bool, isGlobalAdmin bool) (string, error) {
	claims := Claims{
		UserID:        userID,
		Email:         email,
		IsAdmin:       isAdmin,
		IsGlobalAdmin: isGlobalAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// AuthMiddleware validates JWT tokens
// Chi's r.Use() automatically provides the next http.Handler
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		// Context in Go helps with passsing request-scoped values,
		// cancellation signals and deadlines across API/function boundaries
		// backpack that travels with http request through each handler chain
		// grab the context that is a part of the request r
		// WithValue creates a new context with the added value, contexts are immutable
		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		// pass the new context to the next handler
		// r.WithContext creates shallow copy of the request with the New context
		// next.ServeHttp call the next handler in the chain
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminMiddleware ensures the user is a global admin or league owner
// Note: This middleware only checks if user has SOME admin privileges
// Individual handlers should verify league-specific permissions
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*Claims)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Allow global admins OR league owners (IsAdmin is set for league owners)
		if !claims.IsAdmin && !claims.IsGlobalAdmin {
			http.Error(w, "Admin access required", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext extracts user claims from the request context
func GetUserFromContext(r *http.Request) (*Claims, bool) {
	claims, ok := r.Context().Value(UserContextKey).(*Claims)
	return claims, ok
}
