package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func (app *application) HandleInitiateReset(w http.ResponseWriter, r *http.Request) {
	var resetEmail string
	var resetToken string
	var resetIssued string
	var resetExpiry string
	var userIsDeactivated bool

	err := r.ParseForm()
	if err != nil {
		log.Print("Error HandleResetRequest 10: ", err.Error())
		http.Error(w, "Invalid input. Please check the form and try again.", http.StatusInternalServerError)
		return
	}

	resetEmail = r.FormValue("email")
	if resetEmail == "" {
		log.Print("Error HandleResetRequest 20: Empty email")
		http.Error(w, "Email is required.", http.StatusBadRequest)
		return
	}

	statement := "SELECT user_isdeactivated FROM USERS WHERE user_email = ?"
	err = app.db.QueryRow(statement, resetEmail).Scan(&userIsDeactivated)

	switch {
	case err == sql.ErrNoRows:
		log.Printf("Alert HandleResetRequest 30: No user found for %s", resetEmail)
		// Skip token and db actions, go to render template
	case err != nil:
		log.Print("Error HandleResetRequest 31: ", err.Error())
		http.Error(w, "Server error. Please try again later.", http.StatusInternalServerError)
		return
	case userIsDeactivated:
		log.Print("Alert HandleResetRequest 32: Deactivated user requested reset ", resetEmail)
		http.Error(w, "That account has been deactivated.", http.StatusBadRequest)
		return
	default:
		// User exists and is not deactivated
		// Generate reset token - 32 digit hexadecimal string
		tokenBytes := make([]byte, 16) // 16 bytes = 32 hex chars
		_, err = rand.Read(tokenBytes)
		if err != nil {
			log.Print("Error HandleResetRequest 40: ", err.Error())
			http.Error(w, "Server error. Please try again later.", http.StatusInternalServerError)
			return
		}
		resetToken = hex.EncodeToString(tokenBytes)

		// Set timestamps
		now := time.Now().UTC()
		resetIssued = now.Format("2006-01-02 15:04:05")
		resetExpiry = now.Add(24 * time.Hour).Format("2006-01-02 15:04:05")

		// Insert into database
		statement = "INSERT INTO RESETS (reset_email, reset_token, reset_issued, reset_expiry) VALUES (?, ?, ?, ?)"
		_, err = app.db.Exec(statement, resetEmail, resetToken, resetIssued, resetExpiry)
		if err != nil {
			log.Print("Error HandleResetRequest 50: ", err.Error())
			http.Error(w, "Server error. Please try again later.", http.StatusInternalServerError)
			return
		}

		//Send email
		subject := "Cannan.dev Password Reset"
		body := fmt.Sprintf(`Hello there,

You requested a password reset for cannan.dev. To do so, please visit the link below:
https://cannan.dev/reset/%s

This link will expire in 24 hours. If you didn't request this reset, please ignore this email.

Regards,

cannan.dev`, resetToken)

		err = sendEmail(resetEmail, subject, body)
		if err != nil {
			log.Print("Error HandleResetRequest 60: ", err.Error())
			http.Error(w, "Server error. Please try again later.", http.StatusInternalServerError)
			return
		}
		log.Print("Queued reset email to ", resetEmail)
	}

	data := struct {
		Email string
	}{
		Email: resetEmail,
	}

	tmpl, err := template.ParseFiles("../dynamic/reset_requested.html")
	if err != nil {
		log.Print("Error HandleResetRequest 70: ", err.Error())
		http.Error(w, "Server error. Please try again later.", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Print("Error HandleResetRequest 80: ", err.Error())
		http.Error(w, "Server error. Please try again later.", http.StatusInternalServerError)
	}
}

func (app *application) HandleGetResetForm(w http.ResponseWriter, r *http.Request) {
	var resetEmail string
	var resetExpiry string
	var resetUsed bool

	resetToken := r.PathValue("token")

	// Lookup reset record
	statement := "SELECT reset_email, reset_expiry, reset_used FROM RESETS WHERE reset_token = ?"
	err := app.db.QueryRow(statement, resetToken).Scan(&resetEmail, &resetExpiry, &resetUsed)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		log.Print("Error HandleGetResetForm 10: ", err.Error())
		http.Error(w, "Server error. Please try again later.", http.StatusInternalServerError)
		return
	}

	// Check expiry and usage
	expiryTime, err := time.Parse("2006-01-02 15:04:05", resetExpiry)
	expiryTime = expiryTime.UTC()
	if err != nil {
		log.Print("Error HandleGetResetForm 20: invalid expiry format: ", err.Error())
		http.Error(w, "Invalid reset link.", http.StatusInternalServerError)
		return
	}
	if time.Now().UTC().After(expiryTime) || resetUsed {
		http.Error(w, "Reset token has expired or already been used.", http.StatusBadRequest)
		return
	}

	// Render reset form
	data := struct {
		Email string
		Token string
	}{
		Email: resetEmail,
		Token: resetToken,
	}

	tmpl, err := template.ParseFiles("../dynamic/reset_confirm.html")
	if err != nil {
		log.Print("Error HandleGetResetForm 30: ", err.Error())
		http.Error(w, "Server error. Please try again later.", http.StatusInternalServerError)
		return
	}
	if err := tmpl.Execute(w, data); err != nil {
		log.Print("Error HandleGetResetForm 40: ", err.Error())
		http.Error(w, "Server error. Please try again later.", http.StatusInternalServerError)
	}
}

func (app *application) HandlePostResetForm(w http.ResponseWriter, r *http.Request) {
	var resetEmail string
	var resetExpiry string
	var resetUsed bool

	resetToken := r.PathValue("token")

	// Lookup reset record
	statement := "SELECT reset_email, reset_expiry, reset_used FROM RESETS WHERE reset_token = ?"
	err := app.db.QueryRow(statement, resetToken).Scan(&resetEmail, &resetExpiry, &resetUsed)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	} else if err != nil {
		log.Print("Error HandlePostResetForm 10: ", err.Error())
		http.Error(w, "Server error. Please try again later.", http.StatusInternalServerError)
		return
	}

	// Check expiry and usage
	expiryTime, err := time.Parse("2006-01-02 15:04:05", resetExpiry)
	if err != nil {
		log.Print("Error HandlePostResetForm 20: invalid expiry format: ", err.Error())
		http.Error(w, "Invalid reset link.", http.StatusInternalServerError)
		return
	}
	if time.Now().UTC().After(expiryTime) || resetUsed {
		http.Error(w, "Reset token has expired or already been used.", http.StatusBadRequest)
		return
	}

	// Parse form
	if err := r.ParseForm(); err != nil {
		log.Print("Error HandlePostResetForm 30: ", err.Error())
		http.Error(w, "Invalid form submission.", http.StatusBadRequest)
		return
	}

	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	if password == "" || confirmPassword == "" {
		http.Error(w, "Password fields are required.", http.StatusBadRequest)
		return
	}
	if password != confirmPassword {
		http.Error(w, "Passwords do not match. Please go back and try again.", http.StatusBadRequest)
		return
	}

	// Hash password
	hash, err := hashPassword(password)
	if err != nil {
		log.Print("Error HandlePostResetForm 40: ", err.Error())
		http.Error(w, "Server error. Please try again later.", http.StatusInternalServerError)
		return
	}

	// Update user password
	_, err = app.db.Exec("UPDATE USERS SET user_hash = ? WHERE user_email = ?", hash, resetEmail)
	if err != nil {
		log.Print("Error HandlePostResetForm 50: ", err.Error())
		http.Error(w, "Server error. Please try again later.", http.StatusInternalServerError)
		return
	}

	// Mark reset token as used
	_, err = app.db.Exec("UPDATE RESETS SET reset_used = 1 WHERE reset_token = ?", resetToken)
	if err != nil {
		log.Print("Error HandlePostResetForm 60: ", err.Error())
		http.Error(w, "Server error. Please try again later.", http.StatusInternalServerError)
		return
	}

	// Redirect to login page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
