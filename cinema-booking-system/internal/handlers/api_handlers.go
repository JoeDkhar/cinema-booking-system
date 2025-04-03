package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/JoeDkhar/cinema-booking-system/internal/database"
	"github.com/JoeDkhar/cinema-booking-system/internal/models"
	"github.com/gorilla/mux"
)

// APIResponse is a standard API response format
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// HealthCheckHandler returns health status of the service
func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	sqlDB, err := database.DB.DB()
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database connection error",
		})
		return
	}

	// Ping database
	err = sqlDB.Ping()
	if err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Database ping failed",
		})
		return
	}

	// All systems go
	sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		},
	})
}

// APIMoviesHandler returns a list of movies in JSON format
func APIMoviesHandler(w http.ResponseWriter, r *http.Request) {
	var movies []models.Movie
	if err := database.DB.Find(&movies).Error; err != nil {
		sendJSONResponse(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to retrieve movies",
		})
		return
	}

	sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    movies,
	})
}

// APIMovieDetailHandler returns details of a specific movie in JSON format
func APIMovieDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		sendJSONResponse(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid movie ID",
		})
		return
	}

	var movie models.Movie
	if err := database.DB.Preload("Shows").First(&movie, id).Error; err != nil {
		sendJSONResponse(w, http.StatusNotFound, APIResponse{
			Success: false,
			Error:   "Movie not found",
		})
		return
	}

	sendJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    movie,
	})
}

// sendJSONResponse sends a structured JSON response
func sendJSONResponse(w http.ResponseWriter, statusCode int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}
