package tests

import (
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/JoeDkhar/cinema-booking-system/internal/database"
	"github.com/JoeDkhar/cinema-booking-system/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestMain sets up the test database
func TestMain(m *testing.M) {
	// Set up a test database
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}

	database.DB = db

	// Migrate the schema
	err = db.AutoMigrate(
		&models.Movie{},
		&models.Show{},
		&models.Booking{},
		&models.User{},
	)
	if err != nil {
		panic("failed to migrate test database")
	}

	// Run tests
	os.Exit(m.Run())
}

// setupTestShow creates a test show with 100 seats
func setupTestShow(t *testing.T) models.Show {
	movie := models.Movie{
		Title:       "Test Movie",
		Description: "A test movie",
		Duration:    120,
		Genre:       "Action",
	}

	if err := database.DB.Create(&movie).Error; err != nil {
		t.Fatalf("Error creating test movie: %v", err)
	}

	show := models.Show{
		MovieID:     movie.ID,
		DateTime:    time.Now().Add(24 * time.Hour),
		HallNumber:  1,
		TotalSeats:  100,
		TicketPrice: 10.0,
	}

	if err := database.DB.Create(&show).Error; err != nil {
		t.Fatalf("Error creating test show: %v", err)
	}

	return show
}

// Test single booking
func TestCreateBooking(t *testing.T) {
	// Setup
	database.DB.Exec("DELETE FROM bookings")
	database.DB.Exec("DELETE FROM shows")
	database.DB.Exec("DELETE FROM movies")

	show := setupTestShow(t)

	// Create a booking
	seats := models.Seats{
		{Row: "A", Number: 1},
		{Row: "A", Number: 2},
	}

	booking := models.Booking{
		ShowID:       show.ID,
		CustomerName: "John Doe",
		Email:        "john@example.com",
		Seats:        seats,
		BookingTime:  time.Now(),
		TotalAmount:  float64(len(seats)) * show.TicketPrice,
		Confirmed:    true,
	}

	if err := database.DB.Create(&booking).Error; err != nil {
		t.Fatalf("Error creating booking: %v", err)
	}

	// Verify booking was created
	var savedBooking models.Booking
	if err := database.DB.First(&savedBooking, booking.ID).Error; err != nil {
		t.Fatalf("Error retrieving saved booking: %v", err)
	}

	if savedBooking.CustomerName != "John Doe" {
		t.Errorf("Expected customer name 'John Doe', got '%s'", savedBooking.CustomerName)
	}

	if len(savedBooking.Seats) != 2 {
		t.Errorf("Expected 2 seats, got %d", len(savedBooking.Seats))
	}

	if savedBooking.TotalAmount != 20.0 {
		t.Errorf("Expected total amount 20.0, got %.2f", savedBooking.TotalAmount)
	}
}

// Test concurrent bookings for the same seat
func TestConcurrentBooking(t *testing.T) {
	// Setup
	database.DB.Exec("DELETE FROM bookings")
	database.DB.Exec("DELETE FROM shows")
	database.DB.Exec("DELETE FROM movies")

	show := setupTestShow(t)

	// Prepare concurrent booking attempts for the same seat
	const numBookings = 5
	var wg sync.WaitGroup
	wg.Add(numBookings)

	// Create a mutex to synchronize seat checking
	mutex := &sync.Mutex{}

	successCount := 0
	var successMutex sync.Mutex

	for i := 0; i < numBookings; i++ {
		go func(i int) {
			defer wg.Done()

			customerName := "User" + string(rune(i+65)) // A, B, C, etc.

			// Generate the same seat for all concurrent bookings
			seats := models.Seats{
				{Row: "B", Number: 10},
			}

			// Check if seat is already booked
			mutex.Lock()
			var existingBookings []models.Booking
			database.DB.Where("show_id = ? AND confirmed = ?", show.ID, true).Find(&existingBookings)

			// Check if any seats are already taken
			seatBooked := false
			for _, booking := range existingBookings {
				for _, seat := range booking.Seats {
					if seat.Row == "B" && seat.Number == 10 {
						seatBooked = true
						break
					}
				}
				if seatBooked {
					break
				}
			}

			// Create booking if seat is available
			if !seatBooked {
				booking := models.Booking{
					ShowID:       show.ID,
					CustomerName: customerName,
					Email:        customerName + "@example.com",
					Seats:        seats,
					BookingTime:  time.Now(),
					TotalAmount:  float64(len(seats)) * show.TicketPrice,
					Confirmed:    true,
				}

				if err := database.DB.Create(&booking).Error; err == nil {
					successMutex.Lock()
					successCount++
					successMutex.Unlock()
				}
			}
			mutex.Unlock()
		}(i)
	}

	wg.Wait()

	// Verify only one booking succeeded
	if successCount != 1 {
		t.Errorf("Expected 1 successful booking, got %d", successCount)
	}

	// Check the database to confirm only one booking exists for the seat
	var bookings []models.Booking
	database.DB.Where("show_id = ? AND confirmed = ?", show.ID, true).Find(&bookings)

	seatBCount := 0
	for _, booking := range bookings {
		for _, seat := range booking.Seats {
			if seat.Row == "B" && seat.Number == 10 {
				seatBCount++
			}
		}
	}

	if seatBCount != 1 {
		t.Errorf("Expected seat B10 to be booked exactly once, got %d bookings", seatBCount)
	}
}

// Test JSON marshaling/unmarshaling
func TestSeatsMarshalingUnmarshaling(t *testing.T) {
	// Create a booking with seats
	seats := models.Seats{
		{Row: "C", Number: 5},
		{Row: "C", Number: 6},
		{Row: "C", Number: 7},
	}

	// Marshal seats to JSON
	seatsData, err := json.Marshal(seats)
	if err != nil {
		t.Fatalf("Error marshaling seats: %v", err)
	}

	// Unmarshal JSON back to seats
	var unmarshaledSeats models.Seats
	err = json.Unmarshal(seatsData, &unmarshaledSeats)
	if err != nil {
		t.Fatalf("Error unmarshaling seats: %v", err)
	}

	// Verify seats are unmarshaled correctly
	if len(unmarshaledSeats) != 3 {
		t.Errorf("Expected 3 seats, got %d", len(unmarshaledSeats))
	}

	if unmarshaledSeats[0].Row != "C" || unmarshaledSeats[0].Number != 5 {
		t.Errorf("Expected seat C5, got %s%d", unmarshaledSeats[0].Row, unmarshaledSeats[0].Number)
	}

	if unmarshaledSeats[1].Row != "C" || unmarshaledSeats[1].Number != 6 {
		t.Errorf("Expected seat C6, got %s%d", unmarshaledSeats[1].Row, unmarshaledSeats[1].Number)
	}

	if unmarshaledSeats[2].Row != "C" || unmarshaledSeats[2].Number != 7 {
		t.Errorf("Expected seat C7, got %s%d", unmarshaledSeats[2].Row, unmarshaledSeats[2].Number)
	}
}
