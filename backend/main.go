package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/gorilla/mux"
	"football-picking-league/backend/db"
	"football-picking-league/backend/handlers"
)

func main() {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(os.Getenv("AWS_REGION")),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Create DB client
	dbClient := db.NewDBClient(cfg)

	// Initialize router
	r := mux.NewRouter()

	// API routes
	r.HandleFunc("/api/games/{leagueId}", handlers.GetGamesHandler(dbClient)).Methods("GET")
	r.HandleFunc("/api/picks", handlers.SubmitPickHandler(dbClient)).Methods("POST")
	r.HandleFunc("/api/leaderboard/{leagueId}", handlers.GetLeaderboardHandler(dbClient)).Methods("GET")
	r.HandleFunc("/api/team-stats", handlers.GetTeamStatsHandler(dbClient)).Methods("GET")

	// Start server
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(port, r))
}
