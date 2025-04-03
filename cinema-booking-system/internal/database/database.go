package database

import (
	"log"
	"time"

	"github.com/JoeDkhar/cinema-booking-system/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Initialize sets up the database connection and migrates the schema
func Initialize(dbPath string) error {
	var err error

	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	DB, err = gorm.Open(sqlite.Open(dbPath), config)
	if err != nil {
		return err
	}

	// Migrate the schema
	err = DB.AutoMigrate(
		&models.Movie{},
		&models.Show{},
		&models.Booking{},
		&models.User{},
	)
	if err != nil {
		return err
	}

	log.Println("Database initialized successfully")
	return nil
}

// SeedInitialData populates the database with sample data if it's empty
func SeedInitialData() error {
	// Check if movies already exist
	var count int64
	DB.Model(&models.Movie{}).Count(&count)
	if count > 0 {
		return nil // data already exists
	}

	// Create sample movies
	movies := []models.Movie{
		{
			Title:       "Inception",
			Description: "A thief who steals corporate secrets through the use of dream-sharing technology.",
			Duration:    148,
			Genre:       "Sci-Fi",
			ImageURL:    "/static/images/inception.jpg",
		},
		{
			Title:       "The Dark Knight",
			Description: "Batman fights the menace known as the Joker.",
			Duration:    152,
			Genre:       "Action",
			ImageURL:    "/static/images/dark_knight.jpg",
		},
		{
			Title:       "Interstellar",
			Description: "A team of explorers travel through a wormhole in space.",
			Duration:    169,
			Genre:       "Sci-Fi",
			ImageURL:    "/static/images/interstellar.jpg",
		},
	}

	for i := range movies {
		if err := DB.Create(&movies[i]).Error; err != nil {
			return err
		}

		// Create sample shows for each movie
		shows := []models.Show{
			{
				MovieID:     movies[i].ID,
				DateTime:    time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour).Add(15 * time.Hour),
				HallNumber:  i + 1,
				TotalSeats:  100,
				TicketPrice: 12.50,
			},
			{
				MovieID:     movies[i].ID,
				DateTime:    time.Now().AddDate(0, 0, 1).Truncate(24 * time.Hour).Add(19 * time.Hour),
				HallNumber:  i + 1,
				TotalSeats:  100,
				TicketPrice: 15.00,
			},
			{
				MovieID:     movies[i].ID,
				DateTime:    time.Now().AddDate(0, 0, 2).Truncate(24 * time.Hour).Add(17 * time.Hour),
				HallNumber:  i + 1,
				TotalSeats:  100,
				TicketPrice: 12.50,
			},
		}

		for _, show := range shows {
			if err := DB.Create(&show).Error; err != nil {
				return err
			}
		}
	}

	log.Println("Initial data seeded successfully")
	return nil
}
