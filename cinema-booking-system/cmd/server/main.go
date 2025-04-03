package main

import (
	"log"
	"net/http"
	"os"

	"github.com/JoeDkhar/cinema-booking-system/internal/database"
	"github.com/JoeDkhar/cinema-booking-system/internal/handlers"
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

	// Create router
	r := mux.NewRouter()

	// Register routes
	r.HandleFunc("/", handlers.HomeHandler).Methods("GET")
	r.HandleFunc("/movies", handlers.MoviesHandler).Methods("GET")
	r.HandleFunc("/movies/{id:[0-9]+}", handlers.MovieDetailHandler).Methods("GET")
	r.HandleFunc("/shows/{id:[0-9]+}", handlers.ShowDetailHandler).Methods("GET")
	r.HandleFunc("/booking", handlers.BookingHandler).Methods("POST")
	r.HandleFunc("/booking/confirmation/{id:[0-9]+}", handlers.BookingConfirmationHandler).Methods("GET")

	// API routes
	r.HandleFunc("/api/shows/{id:[0-9]+}/seats", handlers.GetAvailableSeatsHandler).Methods("GET")

	// Serve static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Start server
	port := "8080"
	log.Printf("Server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
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
