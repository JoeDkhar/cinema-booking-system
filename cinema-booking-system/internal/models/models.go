package models

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
)

// Movie represents a film showing in the cinema
type Movie struct {
	gorm.Model
	Title       string `json:"title"`
	Description string `json:"description"`
	Duration    int    `json:"duration_minutes"`
	Genre       string `json:"genre"`
	ImageURL    string `json:"image_url"`
	Shows       []Show `json:"shows" gorm:"foreignKey:MovieID"`
}

// Show represents a specific screening of a movie
type Show struct {
	gorm.Model
	MovieID     uint      `json:"movie_id"`
	DateTime    time.Time `json:"date_time"`
	HallNumber  int       `json:"hall_number"`
	TotalSeats  int       `json:"total_seats"`
	TicketPrice float64   `json:"ticket_price"`
	Bookings    []Booking `json:"bookings" gorm:"foreignKey:ShowID"`
}

// Seat represents a specific seat in the cinema hall
type Seat struct {
	Row    string `json:"row"`
	Number int    `json:"number"`
}

// Seats represents a collection of seats
type Seats []Seat

// MarshalJSON custom JSON marshaler for Seats
func (s Seats) MarshalJSON() ([]byte, error) {
	return json.Marshal([]Seat(s))
}

// UnmarshalJSON custom JSON unmarshaler for Seats
func (s *Seats) UnmarshalJSON(data []byte) error {
	var seats []Seat
	if err := json.Unmarshal(data, &seats); err != nil {
		return err
	}
	*s = Seats(seats)
	return nil
}

// Booking represents a ticket booking
type Booking struct {
	gorm.Model
	ShowID       uint      `json:"show_id"`
	CustomerName string    `json:"customer_name"`
	Email        string    `json:"email"`
	SeatsJSON    string    `json:"-"` // Stored as JSON string in database
	Seats        Seats     `json:"seats" gorm:"-"`
	BookingTime  time.Time `json:"booking_time"`
	TotalAmount  float64   `json:"total_amount"`
	Confirmed    bool      `json:"confirmed" gorm:"default:false"`
}

// BeforeSave handles JSON marshaling of seats before saving to the database
func (b *Booking) BeforeSave(tx *gorm.DB) error {
	if len(b.Seats) == 0 {
		return errors.New("booking must have at least one seat")
	}

	seatsData, err := json.Marshal(b.Seats)
	if err != nil {
		return err
	}
	b.SeatsJSON = string(seatsData)
	return nil
}

// AfterFind handles JSON unmarshaling of seats after retrieving from the database
func (b *Booking) AfterFind(tx *gorm.DB) error {
	if b.SeatsJSON == "" {
		return nil
	}

	return json.Unmarshal([]byte(b.SeatsJSON), &b.Seats)
}

// User represents a registered user of the system
type User struct {
	gorm.Model
	Username     string `json:"username" gorm:"unique"`
	Email        string `json:"email" gorm:"unique"`
	PasswordHash string `json:"-"`
	SessionToken string `json:"-"`
	IsAdmin      bool   `json:"is_admin" gorm:"default:false"`
}
