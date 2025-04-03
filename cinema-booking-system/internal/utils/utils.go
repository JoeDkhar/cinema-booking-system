package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"
)

// FormatCurrency formats a float as a currency string
func FormatCurrency(amount float64) string {
	return fmt.Sprintf("$%.2f", amount)
}

// FormatDateTime formats a time into a readable string
func FormatDateTime(t time.Time) string {
	return t.Format("Mon, Jan 2, 2006 at 3:04 PM")
}

// FormatDate formats a time into a date string
func FormatDate(t time.Time) string {
	return t.Format("Mon, Jan 2, 2006")
}

// FormatTime formats a time into a time string
func FormatTime(t time.Time) string {
	return t.Format("3:04 PM")
}

// GenerateBookingReference creates a unique reference number
func GenerateBookingReference() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return fmt.Sprintf("BKG-%s", hex.EncodeToString(bytes))
}

// ValidateEmail checks if an email is valid
func ValidateEmail(email string) bool {
	pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, _ := regexp.MatchString(pattern, email)
	return match
}
