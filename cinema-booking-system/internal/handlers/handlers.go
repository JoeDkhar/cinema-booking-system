package handlers

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/JoeDkhar/cinema-booking-system/internal/cache"
	"github.com/JoeDkhar/cinema-booking-system/internal/database"
	"github.com/JoeDkhar/cinema-booking-system/internal/models"
	"github.com/JoeDkhar/cinema-booking-system/internal/utils"
	"github.com/gorilla/mux"
)

var (
	templates *template.Template
	// Mutex for each show to prevent double bookings
	showMutexes = make(map[uint]*sync.Mutex)
	// Global mutex to protect the showMutexes map
	globalMutex sync.Mutex

	// Generic caches for movies and shows using Go generics
	movieCache = cache.NewCache[models.Movie]()
	showCache  = cache.NewCache[models.Show]()
)

// Initialize loads templates and sets up handlers
func Initialize() {
	// Define template functions
	funcMap := template.FuncMap{
		"currentYear": func() int {
			return time.Now().Year()
		},
		"formatCurrency": utils.FormatCurrency,
		"formatDateTime": utils.FormatDateTime,
		"formatDate":     utils.FormatDate,
		"formatTime":     utils.FormatTime,
	}

	// Parse templates with functions
	templates = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))
}

// getShowMutex returns a mutex for a specific show
func getShowMutex(showID uint) *sync.Mutex {
	globalMutex.Lock()
	defer globalMutex.Unlock()

	if _, exists := showMutexes[showID]; !exists {
		showMutexes[showID] = &sync.Mutex{}
	}
	return showMutexes[showID]
}

// HomeHandler renders the home page
func HomeHandler(w http.ResponseWriter, r *http.Request) {
	var movies []models.Movie
	database.DB.Find(&movies)

	data := struct {
		Movies []models.Movie
	}{
		Movies: movies,
	}

	templates.ExecuteTemplate(w, "home.html", data)
}

// MoviesHandler renders the movies listing page
func MoviesHandler(w http.ResponseWriter, r *http.Request) {
	var movies []models.Movie
	database.DB.Find(&movies)

	data := struct {
		Movies []models.Movie
	}{
		Movies: movies,
	}

	templates.ExecuteTemplate(w, "movies.html", data)
}

// MovieDetailHandler renders the details of a specific movie
func MovieDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	// Try to get movie from cache first
	cacheKey := "movie_" + vars["id"]
	movie, found := movieCache.Get(cacheKey)

	if !found {
		// If not in cache, get from database
		if err := database.DB.Preload("Shows").First(&movie, id).Error; err != nil {
			http.Error(w, "Movie not found", http.StatusNotFound)
			return
		}

		// Store in cache for 10 minutes
		movieCache.Set(cacheKey, movie, 10*time.Minute)
	}

	data := struct {
		Movie models.Movie
	}{
		Movie: movie,
	}

	templates.ExecuteTemplate(w, "movie_detail.html", data)
}

// ShowDetailHandler renders the page for selecting seats for a specific show
func ShowDetailHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid show ID", http.StatusBadRequest)
		return
	}

	var show models.Show
	if err := database.DB.Preload("Movie").First(&show, id).Error; err != nil {
		http.Error(w, "Show not found", http.StatusNotFound)
		return
	}

	// Get all bookings for this show to determine booked seats
	var bookings []models.Booking
	database.DB.Where("show_id = ? AND confirmed = ?", show.ID, true).Find(&bookings)

	// Create a map of booked seats
	bookedSeats := make(map[string]bool)
	for _, booking := range bookings {
		for _, seat := range booking.Seats {
			key := seat.Row + strconv.Itoa(seat.Number)
			bookedSeats[key] = true
		}
	}

	data := struct {
		Show        models.Show
		BookedSeats map[string]bool
	}{
		Show:        show,
		BookedSeats: bookedSeats,
	}

	templates.ExecuteTemplate(w, "booking.html", data)
}

// BookingHandler handles the seat booking process
func BookingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	showIDStr := r.FormValue("show_id")
	showID, err := strconv.Atoi(showIDStr)
	if err != nil {
		http.Error(w, "Invalid show ID", http.StatusBadRequest)
		return
	}

	customerName := r.FormValue("customer_name")
	email := r.FormValue("email")
	seatsJSON := r.FormValue("seats")

	if customerName == "" || email == "" || seatsJSON == "" {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	// Parse selected seats
	var seats models.Seats
	err = json.Unmarshal([]byte(seatsJSON), &seats)
	if err != nil {
		http.Error(w, "Invalid seat selection", http.StatusBadRequest)
		return
	}

	if len(seats) == 0 {
		http.Error(w, "No seats selected", http.StatusBadRequest)
		return
	}

	// Create a channel for the booking response
	responseChan := make(chan BookingResponse)

	// Create a booking request
	bookingRequest := BookingRequest{
		ShowID:       uint(showID),
		CustomerName: customerName,
		Email:        email,
		Seats:        seats,
		ResponseChan: responseChan,
	}

	// Submit the booking request asynchronously
	ProcessBookingAsync(bookingRequest)

	// Wait for the response
	response := <-responseChan
	close(responseChan)

	if !response.Success {
		http.Error(w, response.ErrorMessage, http.StatusConflict)
		return
	}

	// Redirect to confirmation page
	http.Redirect(w, r, "/booking/confirmation/"+strconv.Itoa(int(response.BookingID)), http.StatusSeeOther)
}

// BookingConfirmationHandler renders the booking confirmation page
func BookingConfirmationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid booking ID", http.StatusBadRequest)
		return
	}

	var booking models.Booking
	if err := database.DB.Preload("Show").Preload("Show.Movie").First(&booking, id).Error; err != nil {
		http.Error(w, "Booking not found", http.StatusNotFound)
		return
	}

	data := struct {
		Booking models.Booking
	}{
		Booking: booking,
	}

	templates.ExecuteTemplate(w, "confirmation.html", data)
}

// API Handlers for AJAX requests

// GetAvailableSeatsHandler returns JSON of available seats for a show
func GetAvailableSeatsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid show ID", http.StatusBadRequest)
		return
	}

	var show models.Show
	if err := database.DB.First(&show, id).Error; err != nil {
		http.Error(w, "Show not found", http.StatusNotFound)
		return
	}

	// Get all bookings for this show
	var bookings []models.Booking
	database.DB.Where("show_id = ? AND confirmed = ?", show.ID, true).Find(&bookings)

	// Create a map of booked seats
	bookedSeats := make(map[string]bool)
	for _, booking := range bookings {
		for _, seat := range booking.Seats {
			key := seat.Row + strconv.Itoa(seat.Number)
			bookedSeats[key] = true
		}
	}

	// Convert to a response format
	type SeatStatus struct {
		Row    string `json:"row"`
		Number int    `json:"number"`
		Booked bool   `json:"booked"`
	}

	// Generate all seats for the show
	var seatStatuses []SeatStatus
	rows := []string{"A", "B", "C", "D", "E", "F", "G", "H"}
	seatsPerRow := show.TotalSeats / len(rows)

	for _, row := range rows {
		for num := 1; num <= seatsPerRow; num++ {
			key := row + strconv.Itoa(num)
			seatStatuses = append(seatStatuses, SeatStatus{
				Row:    row,
				Number: num,
				Booked: bookedSeats[key],
			})
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(seatStatuses)
}
