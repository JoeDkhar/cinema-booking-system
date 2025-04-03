# Cinema Ticket Booking System

A robust cinema ticket booking and management system built with Go, featuring real-time seat booking, concurrency management, and an intuitive user interface.

## Features

- Browse movies and showtimes
- Real-time seat selection and booking
- Concurrency management to prevent double bookings
- JSON marshaling/unmarshaling for data storage
- Comprehensive test suite

## Technology Stack

- Backend: Go with Gorilla Mux for routing
- Database: SQLite (can be configured for PostgreSQL)
- ORM: GORM
- Frontend: HTML, CSS, JavaScript

## Getting Started

### Prerequisites

- Go 1.16+
- SQLite (or PostgreSQL)

### Installation

1. Clone the repository: git clone https://github.com/JoeDkhar/cinema-booking-system.git cd cinema-booking-system
2. Install dependencies: go mod download
3. Run the application: go run cmd/server/main.go

4. Open your browser and visit `http://localhost:8080`

## Project Structure

- `cmd/server`: Application entry point
- `internal/database`: Database configuration and interactions
- `internal/handlers`: HTTP request handlers
- `internal/models`: Data models
- `internal/utils`: Utility functions
- `templates`: HTML templates
- `static`: CSS, JavaScript, and images
- `tests`: Unit tests

## Testing

Run tests with: go test ./tests/...


## License

[MIT](LICENSE)