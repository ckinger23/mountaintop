package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ckinger23/mountaintop/internal/app"
	"github.com/ckinger23/mountaintop/internal/database"
	"github.com/ckinger23/mountaintop/internal/handlers"
	"github.com/ckinger23/mountaintop/internal/middleware"
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

	db, err := database.Connect(dbPath)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Seed data (only runs if database is empty)
	if err := database.SeedData(db); err != nil {
		log.Fatal("Failed to seed data:", err)
	}

	// Initialize application with dependencies
	application := app.NewApp(db)

	// Initialize router
	// returns a *chi.Mux which implements http.Handler
	// define routes, URL params, add middleware (logging auth, recover)
	// create Groups that all use same middleware
	// mounta as HTTP Server with ListenAndServe
	r := chi.NewRouter()

	// Middleware
	r.Use(chimiddleware.Logger)
	// catch panics and prevent server from crashing
	// panic caught, HTTP 500 fro request, server keeps running
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
	r.Post("/api/auth/register", handlers.Register(application))
	r.Post("/api/auth/login", handlers.Login(application))

	// Public read-only routes (no auth required)
	r.Get("/api/teams", handlers.GetTeams(application))
	r.Get("/api/seasons", handlers.GetSeasons(application))
	r.Get("/api/leaderboard", handlers.GetLeaderboard(application))

	// Protected routes (authentication required)
	r.Group(func(r chi.Router) {
		// r.Use provides the http.Handler argument to middleware automatically
		r.Use(middleware.AuthMiddleware)

		// User routes
		r.Get("/api/auth/me", handlers.GetCurrentUser(application))

		// League management
		r.Post("/api/leagues", handlers.CreateLeague(application))
		r.Get("/api/leagues", handlers.GetMyLeagues(application))
		r.Get("/api/leagues/{id}", handlers.GetLeague(application))
		r.Put("/api/leagues/{id}", handlers.UpdateLeague(application))
		r.Delete("/api/leagues/{id}/leave", handlers.LeaveLeague(application))
		r.Post("/api/leagues/join", handlers.JoinLeague(application))
		r.Get("/api/leagues/browse", handlers.BrowsePublicLeagues(application))

		// Games
		r.Get("/api/games", handlers.GetGames(application))
		r.Get("/api/games/{id}", handlers.GetGame(application))
		r.Get("/api/weeks", handlers.GetWeeks(application))
		r.Get("/api/weeks/current", handlers.GetCurrentWeek(application))

		// Picks
		r.Post("/api/picks", handlers.SubmitPick(application))
		r.Get("/api/picks/me", handlers.GetMyPicks(application))
		r.Get("/api/picks/user/{userId}", handlers.GetPicksForUser(application))
		r.Get("/api/picks/week/{weekId}", handlers.GetAllPicksForWeek(application))
		r.Get("/api/picks/stats/{userId}", handlers.GetPickStats(application))
	})

	// Admin routes (authentication + admin required)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware)
		r.Use(middleware.AdminMiddleware)

		// Game management
		r.Post("/api/admin/games", handlers.CreateGame(application))
		r.Put("/api/admin/games/{id}", handlers.UpdateGame(application))
		r.Delete("/api/admin/games/{id}", handlers.DeleteGame(application))
		r.Put("/api/admin/games/{id}/result", handlers.UpdateGameResult(application))

		// Season management
		r.Post("/api/admin/seasons", handlers.CreateSeason(application))

		// Week management
		r.Post("/api/admin/weeks", handlers.CreateWeek(application))
		r.Put("/api/admin/weeks/{id}", handlers.UpdateWeek(application))
		r.Put("/api/admin/weeks/{id}/open", handlers.OpenWeekForPicks(application))
		r.Put("/api/admin/weeks/{id}/lock", handlers.LockWeek(application))
		r.Put("/api/admin/weeks/{id}/complete", handlers.CompleteWeek(application))
	})

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
