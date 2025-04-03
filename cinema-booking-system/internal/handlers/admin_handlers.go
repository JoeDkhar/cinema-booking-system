package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/JoeDkhar/cinema-booking-system/internal/database"
	"github.com/JoeDkhar/cinema-booking-system/internal/models"
	"github.com/gorilla/mux"
)

// AdminDashboardHandler renders the admin dashboard
func AdminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Get statistics
	var movieCount, showCount, bookingCount, userCount int64

	database.DB.Model(&models.Movie{}).Count(&movieCount)
	database.DB.Model(&models.Show{}).Count(&showCount)
	database.DB.Model(&models.Booking{}).Count(&bookingCount)
	database.DB.Model(&models.User{}).Count(&userCount)

	// Get recent bookings
	var recentBookings []models.Booking
	database.DB.Preload("Show").Preload("Show.Movie").Order("created_at DESC").Limit(10).Find(&recentBookings)

	data := struct {
		MovieCount     int64
		ShowCount      int64
		BookingCount   int64
		UserCount      int64
		RecentBookings []models.Booking
		User           models.User
	}{
		MovieCount:     movieCount,
		ShowCount:      showCount,
		BookingCount:   bookingCount,
		UserCount:      userCount,
		RecentBookings: recentBookings,
		User:           r.Context().Value("user").(models.User),
	}

	templates.ExecuteTemplate(w, "admin_dashboard.html", data)
}

// AdminNewMovieHandler handles creation of new movies
func AdminNewMovieHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		templates.ExecuteTemplate(w, "admin_movie_form.html", map[string]interface{}{
			"Action": "Create",
			"User":   r.Context().Value("user").(models.User),
		})
		return
	}

	// Process form submission (POST)
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	genre := r.FormValue("genre")
	durationStr := r.FormValue("duration")
	imageURL := r.FormValue("image_url")

	// Validate input
	if title == "" || description == "" || genre == "" || durationStr == "" {
		templates.ExecuteTemplate(w, "admin_movie_form.html", map[string]interface{}{
			"Action": "Create",
			"Error":  "All fields are required",
			"User":   r.Context().Value("user").(models.User),
		})
		return
	}

	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration <= 0 {
		templates.ExecuteTemplate(w, "admin_movie_form.html", map[string]interface{}{
			"Action": "Create",
			"Error":  "Duration must be a positive number",
			"User":   r.Context().Value("user").(models.User),
		})
		return
	}

	// Create movie
	movie := models.Movie{
		Title:       title,
		Description: description,
		Genre:       genre,
		Duration:    duration,
		ImageURL:    imageURL,
	}

	if err := database.DB.Create(&movie).Error; err != nil {
		templates.ExecuteTemplate(w, "admin_movie_form.html", map[string]interface{}{
			"Action": "Create",
			"Error":  "Error creating movie: " + err.Error(),
			"User":   r.Context().Value("user").(models.User),
		})
		return
	}

	// Redirect to admin movies page
	http.Redirect(w, r, "/admin/movies", http.StatusSeeOther)
}

// AdminEditMovieHandler handles editing of existing movies
func AdminEditMovieHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	var movie models.Movie
	if err := database.DB.First(&movie, id).Error; err != nil {
		http.Error(w, "Movie not found", http.StatusNotFound)
		return
	}

	if r.Method == http.MethodGet {
		templates.ExecuteTemplate(w, "admin_movie_form.html", map[string]interface{}{
			"Action": "Edit",
			"Movie":  movie,
			"User":   r.Context().Value("user").(models.User),
		})
		return
	}

	// Process form submission (POST)
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	genre := r.FormValue("genre")
	durationStr := r.FormValue("duration")
	imageURL := r.FormValue("image_url")

	// Validate input
	if title == "" || description == "" || genre == "" || durationStr == "" {
		templates.ExecuteTemplate(w, "admin_movie_form.html", map[string]interface{}{
			"Action": "Edit",
			"Movie":  movie,
			"Error":  "All fields are required",
			"User":   r.Context().Value("user").(models.User),
		})
		return
	}

	duration, err := strconv.Atoi(durationStr)
	if err != nil || duration <= 0 {
		templates.ExecuteTemplate(w, "admin_movie_form.html", map[string]interface{}{
			"Action": "Edit",
			"Movie":  movie,
			"Error":  "Duration must be a positive number",
			"User":   r.Context().Value("user").(models.User),
		})
		return
	}

	// Update movie
	movie.Title = title
	movie.Description = description
	movie.Genre = genre
	movie.Duration = duration
	movie.ImageURL = imageURL

	if err := database.DB.Save(&movie).Error; err != nil {
		templates.ExecuteTemplate(w, "admin_movie_form.html", map[string]interface{}{
			"Action": "Edit",
			"Movie":  movie,
			"Error":  "Error updating movie: " + err.Error(),
			"User":   r.Context().Value("user").(models.User),
		})
		return
	}

	// Redirect to admin movies page
	http.Redirect(w, r, "/admin/movies", http.StatusSeeOther)
}

// AdminNewShowHandler handles creation of new shows
func AdminNewShowHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		var movies []models.Movie
		database.DB.Find(&movies)

		templates.ExecuteTemplate(w, "admin_show_form.html", map[string]interface{}{
			"Action": "Create",
			"Movies": movies,
			"User":   r.Context().Value("user").(models.User),
		})
		return
	}

	// Process form submission (POST)
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	movieIDStr := r.FormValue("movie_id")
	dateStr := r.FormValue("date")
	timeStr := r.FormValue("time")
	hallNumberStr := r.FormValue("hall_number")
	totalSeatsStr := r.FormValue("total_seats")
	ticketPriceStr := r.FormValue("ticket_price")

	// Validate input
	if movieIDStr == "" || dateStr == "" || timeStr == "" || hallNumberStr == "" || totalSeatsStr == "" || ticketPriceStr == "" {
		var movies []models.Movie
		database.DB.Find(&movies)

		templates.ExecuteTemplate(w, "admin_show_form.html", map[string]interface{}{
			"Action": "Create",
			"Movies": movies,
			"Error":  "All fields are required",
			"User":   r.Context().Value("user").(models.User),
		})
		return
	}

	// Parse values
	movieID, err := strconv.Atoi(movieIDStr)
	if err != nil {
		http.Error(w, "Invalid movie ID", http.StatusBadRequest)
		return
	}

	// Parse date and time
	dateTime, err := time.Parse("2006-01-02 15:04", dateStr+" "+timeStr)
	if err != nil {
		http.Error(w, "Invalid date or time format", http.StatusBadRequest)
		return
	}

	hallNumber, err := strconv.Atoi(hallNumberStr)
	if err != nil || hallNumber <= 0 {
		http.Error(w, "Invalid hall number", http.StatusBadRequest)
		return
	}

	totalSeats, err := strconv.Atoi(totalSeatsStr)
	if err != nil || totalSeats <= 0 {
		http.Error(w, "Invalid total seats", http.StatusBadRequest)
		return
	}

	ticketPrice, err := strconv.ParseFloat(ticketPriceStr, 64)
	if err != nil || ticketPrice <= 0 {
		http.Error(w, "Invalid ticket price", http.StatusBadRequest)
		return
	}

	// Create show
	show := models.Show{
		MovieID:     uint(movieID),
		DateTime:    dateTime,
		HallNumber:  hallNumber,
		TotalSeats:  totalSeats,
		TicketPrice: ticketPrice,
	}

	if err := database.DB.Create(&show).Error; err != nil {
		http.Error(w, "Error creating show: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Redirect to admin shows page
	http.Redirect(w, r, "/admin/shows", http.StatusSeeOther)
}

// AdminBookingsHandler displays all bookings
func AdminBookingsHandler(w http.ResponseWriter, r *http.Request) {
	var bookings []models.Booking
	database.DB.Preload("Show").Preload("Show.Movie").Find(&bookings)

	data := struct {
		Bookings []models.Booking
		User     models.User
	}{
		Bookings: bookings,
		User:     r.Context().Value("user").(models.User),
	}

	templates.ExecuteTemplate(w, "admin_bookings.html", data)
}
