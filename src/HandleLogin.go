package main

import (
	"html/template"
	"log"
	"net/http"

	"github.com/alexedwards/argon2id"
	_ "github.com/mattn/go-sqlite3"
)

func SendLoginFail(w http.ResponseWriter, IP string, email string) {
	log.Print("Failed login: ", IP, " ", email)
	response, err := template.ParseFiles("../dynamic/login_fail.html")
	if err != nil {
		log.Print("Error SendLoginFail() 100 - Failed parsing login_fail.html")
		return
	}
	err = response.Execute(w, nil)
	if err != nil {
		log.Print("Error SendLoginFail() 200 - Failed serving login_fail.html")
		return
	}
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var storedHash string
	var PasswordIsCorrect bool
	var response *template.Template

	IP := r.RemoteAddr
	err := r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")

	if err != nil {
		log.Print("Error HandleLogin() 100 - Failed parsing /login form")
		SendLoginFail(w, IP, email)
		return
	}

	err = db.QueryRow("SELECT user_hash FROM USERS WHERE user_email = ?", email).Scan(&storedHash)
	if err != nil {
		log.Print("Error HandleLogin() 200 - Wrong email or failed SQL query to find or get user hash based on provided email")
		SendLoginFail(w, IP, email)
		return
	}

	PasswordIsCorrect, err = argon2id.ComparePasswordAndHash(password, storedHash)
	if err != nil {
		log.Print("Error HandleLogin() 300 - Failed Argon2id comparing password and hash")
		SendLoginFail(w, IP, email)
		return
	}

	if PasswordIsCorrect {
		log.Print("Successful login: ", IP, " ", email)
		response, err = template.ParseFiles("../dynamic/login_success.html")
		if err != nil {
			log.Print("Error HandleLogin() 400 - Failed parsing login_success.html")
			SendLoginFail(w, IP, email)
			return
		}
		err = response.Execute(w, nil)
		if err != nil {
			log.Print("Error HandleLogin() 500 - Failed serving login_success.html")
			SendLoginFail(w, IP, email)
			return
		}
	} else {
		log.Print("Alert HandleLogin() 600 - Wrong password")
		SendLoginFail(w, IP, email)
		return
	}
}
