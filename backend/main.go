package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"

	"football-picking-league/backend/db"
	"football-picking-league/backend/handlers"
)

// safeSubstring safely returns the last n characters of a string
// If the string is shorter than n, returns the entire string
func safeSubstring(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Set up environment
	env := os.Getenv("ENV")
	if env == "local" {
		log.Println("Running in LOCAL mode with LocalStack")
	} else {
		log.Println("Running in PRODUCTION mode with DynamoDB")
	}

	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}

	// Log AWS configuration details
	creds, err := cfg.Credentials.Retrieve(context.TODO())
	if err != nil {
		log.Printf("Warning: Unable to retrieve AWS credentials: %v", err)
	} else {
		log.Printf("AWS Region: %s", cfg.Region)
		log.Printf("AWS Credentials Source: %s", creds.Source)
		log.Printf("AWS Credentials Session Token: %s", creds.SessionToken)
		log.Printf("AWS Access Key ID: %s (last 4 chars)", safeSubstring(creds.AccessKeyID, 4))
	}

	// Create DB client - NewDBClient will choose the appropriate implementation based on ENV
	dbClient := db.NewDBClient(cfg)

	// Initialize router
	r := mux.NewRouter()

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// API routes
	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Server is healthy"))
	}).Methods("GET")

	// Register handlers with the database client
	r.HandleFunc("/api/games/{leagueId}", handlers.GetGamesHandler(dbClient)).Methods("GET")
	r.HandleFunc("/api/picks", handlers.SubmitPickHandler(dbClient)).Methods("POST")
	r.HandleFunc("/api/leaderboard/{leagueId}", handlers.GetLeaderboardHandler(dbClient)).Methods("GET")
	r.HandleFunc("/api/team-records", handlers.GetTeamsRecordsHandler(dbClient)).Methods("GET")

	// Start server
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(port, r))
}
