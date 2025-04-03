package handlers

import (
	"fmt"
	"log"
	"time"

	"github.com/JoeDkhar/cinema-booking-system/internal/database"
	"github.com/JoeDkhar/cinema-booking-system/internal/models"
)

// BookingRequest represents a seat booking request
type BookingRequest struct {
	ShowID       uint
	CustomerName string
	Email        string
	Seats        models.Seats
	ResponseChan chan BookingResponse
}

// BookingResponse represents the result of a booking request
type BookingResponse struct {
	Success      bool
	BookingID    uint
	ErrorMessage string
}

var (
	// Channel for booking requests
	bookingRequests = make(chan BookingRequest, 100)

	// Channel for cleanup signals
	cleanupSignal = make(chan bool)
)

// StartBookingProcessor initializes the booking processor goroutine
func StartBookingProcessor() {
	go processBookings()
	go periodicCleanup()
}

// StopBookingProcessor stops the booking processor
func StopBookingProcessor() {
	cleanupSignal <- true
	close(bookingRequests)
}

// ProcessBookingAsync submits a booking request to be processed asynchronously
func ProcessBookingAsync(request BookingRequest) {
	bookingRequests <- request
}

// processBookings handles booking requests in a separate goroutine
func processBookings() {
	for request := range bookingRequests {
		// Process the booking request
		booking := processBooking(request)

		// Send the response back through the response channel
		request.ResponseChan <- booking
	}
}

// processBooking handles a single booking request with proper locking
func processBooking(request BookingRequest) BookingResponse {
	// Get the mutex for this show
	mutex := getShowMutex(request.ShowID)
	mutex.Lock()
	defer mutex.Unlock()

	// Get show details
	var show models.Show
	if err := database.DB.First(&show, request.ShowID).Error; err != nil {
		return BookingResponse{
			Success:      false,
			ErrorMessage: "Show not found",
		}
	}

	// Check if seats are already booked
	var existingBookings []models.Booking
	database.DB.Where("show_id = ? AND confirmed = ?", request.ShowID, true).Find(&existingBookings)

	// Create a map of already booked seats
	bookedSeats := make(map[string]bool)
	for _, booking := range existingBookings {
		for _, seat := range booking.Seats {
			key := fmt.Sprintf("%s%d", seat.Row, seat.Number)
			bookedSeats[key] = true
		}
	}

	// Check if any of the selected seats are already booked
	for _, seat := range request.Seats {
		key := fmt.Sprintf("%s%d", seat.Row, seat.Number)
		if bookedSeats[key] {
			return BookingResponse{
				Success:      false,
				ErrorMessage: "Some selected seats are already booked",
			}
		}
	}

	// Create booking
	booking := models.Booking{
		ShowID:       request.ShowID,
		CustomerName: request.CustomerName,
		Email:        request.Email,
		Seats:        request.Seats,
		BookingTime:  time.Now(),
		TotalAmount:  float64(len(request.Seats)) * show.TicketPrice,
		Confirmed:    true,
	}

	// Save booking to database
	if err := database.DB.Create(&booking).Error; err != nil {
		return BookingResponse{
			Success:      false,
			ErrorMessage: "Error saving booking: " + err.Error(),
		}
	}

	return BookingResponse{
		Success:   true,
		BookingID: booking.ID,
	}
}

// periodicCleanup runs booking cleanup tasks periodically
func periodicCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Cancel expired provisional bookings (not confirmed within 15 minutes)
			expiredTime := time.Now().Add(-15 * time.Minute)
			result := database.DB.Where("confirmed = ? AND booking_time < ?", false, expiredTime).Delete(&models.Booking{})
			if result.Error == nil && result.RowsAffected > 0 {
				log.Printf("Cleaned up %d expired bookings", result.RowsAffected)
			}

		case <-cleanupSignal:
			return
		}
	}
}
