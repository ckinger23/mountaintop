package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"football-picking-league/backend/db"
	"football-picking-league/backend/handlers"
)

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
	var cfg aws.Config
	var err error

	// Only load AWS config if not in local mode
	if env != "local" {
		cfg, err = config.LoadDefaultConfig(context.TODO(),
			config.WithRegion(os.Getenv("AWS_REGION")),
		)
		if err != nil {
			log.Fatalf("unable to load SDK config, %v", err)
		}
	}

	// Create DB client - NewDBClient will choose the appropriate implementation
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
	r.HandleFunc("/api/games/{leagueId}", handlers.GetGamesHandler(&dbClient)).Methods("GET")
	r.HandleFunc("/api/picks", handlers.SubmitPickHandler(&dbClient)).Methods("POST")
	r.HandleFunc("/api/leaderboard/{leagueId}", handlers.GetLeaderboardHandler(&dbClient)).Methods("GET")
	r.HandleFunc("/api/team-stats", handlers.GetTeamStatsHandler(&dbClient)).Methods("GET")

	// Start server
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(port, r))
}
