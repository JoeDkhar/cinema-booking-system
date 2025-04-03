module github.com/JoeDkhar/cinema-booking-system

go 1.23.0

toolchain go1.23.5

replace github.com/JoeDkhar/cinema-booking-system => ./

require (
	github.com/gorilla/mux v1.8.1
	golang.org/x/crypto v0.36.0
	gorm.io/driver/sqlite v1.5.7
	gorm.io/gorm v1.25.12
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/mattn/go-sqlite3 v1.14.22 // indirect
	golang.org/x/text v0.23.0 // indirect
)
