package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JoeDkhar/cinema-booking-system/internal/database"
	"github.com/JoeDkhar/cinema-booking-system/internal/handlers"
	"github.com/JoeDkhar/cinema-booking-system/internal/middleware"
	"github.com/gorilla/mux"
)

func main() {
	// Ensure directories exist
	ensureDir("static")
	ensureDir("static/css")
	ensureDir("static/js")
	ensureDir("static/images")
	ensureDir("templates")

	// Initialize database
	dbPath := "cinema.db"
	err := database.Initialize(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Seed initial data
	err = database.SeedInitialData()
	if err != nil {
		log.Fatalf("Failed to seed database: %v", err)
	}

	// Initialize handlers
	handlers.Initialize()

	// Start the booking processor
	handlers.StartBookingProcessor()

	// Defer stopping the booking processor
	defer handlers.StopBookingProcessor()

	// Create router
	r := mux.NewRouter()

	// Apply middlewares
	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.RecoveryMiddleware)
	r.Use(middleware.CORSMiddleware)

	// API routes with version prefix
	api := r.PathPrefix("/api/v1").Subrouter()

	// Public routes
	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/movies", handlers.MoviesHandler).Methods("GET")
	r.HandleFunc("/movies/{id:[0-9]+}", handlers.MovieDetailHandler).Methods("GET")
	r.HandleFunc("/shows/{id:[0-9]+}", handlers.ShowDetailHandler).Methods("GET")
	r.HandleFunc("/booking", handlers.BookingHandler).Methods("POST")
	r.HandleFunc("/booking/confirmation/{id:[0-9]+}", handlers.BookingConfirmationHandler).Methods("GET")

	// User authentication routes
	r.HandleFunc("/register", handlers.RegisterHandler).Methods("GET", "POST")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("GET", "POST")
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("POST")

	// API routes
	api.HandleFunc("/shows/{id:[0-9]+}/seats", handlers.GetAvailableSeatsHandler).Methods("GET")
	api.HandleFunc("/health", handlers.HealthCheckHandler).Methods("GET")
	api.HandleFunc("/movies", handlers.APIMoviesHandler).Methods("GET")
	api.HandleFunc("/movies/{id:[0-9]+}", handlers.APIMovieDetailHandler).Methods("GET")

	// Admin routes (protected)
	admin := r.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.AuthMiddleware)
	admin.HandleFunc("/dashboard", handlers.AdminDashboardHandler).Methods("GET")
	admin.HandleFunc("/movies/new", handlers.AdminNewMovieHandler).Methods("GET", "POST")
	admin.HandleFunc("/movies/{id:[0-9]+}/edit", handlers.AdminEditMovieHandler).Methods("GET", "POST")
	admin.HandleFunc("/shows/new", handlers.AdminNewShowHandler).Methods("GET", "POST")
	admin.HandleFunc("/bookings", handlers.AdminBookingsHandler).Methods("GET")

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Create server with timeouts
	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port 8080...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

// ensureDir makes sure a directory exists, creating it if necessary
func ensureDir(dir string) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
}
