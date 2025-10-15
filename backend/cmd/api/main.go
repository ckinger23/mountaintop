package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ckinger23/cfb-picks-system/internal/database"
	"github.com/ckinger23/cfb-picks-system/internal/handlers"
	"github.com/ckinger23/cfb-picks-system/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func main() {
	// Initialize database
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./cfb-picks.db"
	}

	if err := database.Connect(dbPath); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := database.Migrate(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Seed data (only runs if database is empty)
	if err := database.SeedData(); err != nil {
		log.Fatal("Failed to seed data:", err)
	}

	// Initialize router
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"}, // React dev server
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes
	r.Post("/api/auth/register", handlers.Register)
	r.Post("/api/auth/login", handlers.Login)
	
	// Public read-only routes (no auth required)
	r.Get("/api/teams", handlers.GetTeams)
	r.Get("/api/seasons", handlers.GetSeasons)
	r.Get("/api/leaderboard", handlers.GetLeaderboard)

	// Protected routes (authentication required)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)

		// User routes
		r.Get("/api/auth/me", handlers.GetCurrentUser)

		// Games
		r.Get("/api/games", handlers.GetGames)
		r.Get("/api/games/{id}", handlers.GetGame)
		r.Get("/api/weeks", handlers.GetWeeks)
		r.Get("/api/weeks/current", handlers.GetCurrentWeek)

		// Picks
		r.Post("/api/picks", handlers.SubmitPick)
		r.Get("/api/picks/me", handlers.GetUserPicks)
		r.Get("/api/picks/user/{userId}", handlers.GetPicksByUser)
		r.Get("/api/picks/week/{weekId}", handlers.GetAllPicksForWeek)
		r.Get("/api/picks/stats/{userId}", handlers.GetPickStats)
	})

	// Admin routes (authentication + admin required)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Use(middleware.AdminMiddleware)

		// Game management
		r.Post("/api/admin/games", handlers.CreateGame)
		r.Put("/api/admin/games/{id}/result", handlers.UpdateGameResult)

		// TODO: Add more admin routes for creating weeks, seasons, teams, etc.
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
