package handlers

import (
	"net/http"
	"time"

	"github.com/JoeDkhar/cinema-booking-system/internal/database"
	"github.com/JoeDkhar/cinema-booking-system/internal/models"
	"github.com/JoeDkhar/cinema-booking-system/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// RegisterHandler handles user registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Render registration form
		templates.ExecuteTemplate(w, "register.html", nil)
		return
	}

	// Process form submission (POST)
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	passwordConfirm := r.FormValue("password_confirm")

	// Validate input
	if username == "" || email == "" || password == "" {
		templates.ExecuteTemplate(w, "register.html", map[string]interface{}{
			"Error": "All fields are required",
		})
		return
	}

	if password != passwordConfirm {
		templates.ExecuteTemplate(w, "register.html", map[string]interface{}{
			"Error": "Passwords do not match",
		})
		return
	}

	if !utils.ValidateEmail(email) {
		templates.ExecuteTemplate(w, "register.html", map[string]interface{}{
			"Error": "Invalid email address",
		})
		return
	}

	// Check if username or email already exists
	var count int64
	database.DB.Model(&models.User{}).Where("username = ? OR email = ?", username, email).Count(&count)
	if count > 0 {
		templates.ExecuteTemplate(w, "register.html", map[string]interface{}{
			"Error": "Username or email already exists",
		})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	// Create user
	user := models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}

	if err := database.DB.Create(&user).Error; err != nil {
		http.Error(w, "Error creating user", http.StatusInternalServerError)
		return
	}

	// Redirect to login page
	http.Redirect(w, r, "/login?registered=true", http.StatusSeeOther)
}

// LoginHandler handles user login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Render login form
		templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
			"Registered": r.URL.Query().Get("registered") == "true",
		})
		return
	}

	// Process form submission (POST)
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	redirect := r.FormValue("redirect")

	if redirect == "" {
		redirect = "/"
	}

	// Validate input
	if username == "" || password == "" {
		templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
			"Error": "Username and password are required",
		})
		return
	}

	// Find user
	var user models.User
	if err := database.DB.Where("username = ?", username).First(&user).Error; err != nil {
		templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
			"Error": "Invalid username or password",
		})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
			"Error": "Invalid username or password",
		})
		return
	}

	// Generate session token
	sessionToken := utils.GenerateSessionToken()

	// Update user with session token
	database.DB.Model(&user).Update("session_token", sessionToken)

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    sessionToken,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	// Redirect to requested page
	http.Redirect(w, r, redirect, http.StatusSeeOther)
}

// LogoutHandler handles user logout
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	})

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
