package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"github.com/alexedwards/argon2id"
	_ "github.com/mattn/go-sqlite3"
)

func (app *application) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var storedHash string
	var passwordIsCorrect bool
	var userId int
	var userName string

	err := r.ParseForm()
	if err != nil {
		log.Print("Error HandleLogin 10: ", err.Error())
		http.Error(w, "Invalid input. Please check the form and try again.", http.StatusInternalServerError)
		return
	}

	userEmail := r.FormValue("email")
	userPassword := r.FormValue("password")

	err = app.db.QueryRow("SELECT user_hash FROM USERS WHERE user_email = ?", userEmail).Scan(&storedHash)
	if err == sql.ErrNoRows {
		log.Print("Alert HandleLogin 20: Unknown user ", userEmail, " (", r.RemoteAddr, ")")
		http.Error(w, "Authentication failed. Please check your details and try again.", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Print("Error HandleLogin 21: ", err.Error())
		http.Error(w, "Invalid input. Please check the form and try again.", http.StatusInternalServerError)
		return
	}

	passwordIsCorrect, err = argon2id.ComparePasswordAndHash(userPassword, storedHash)
	if err != nil {
		log.Print("Error HandleLogin 30: ", err.Error())
		http.Error(w, "Invalid input. Please check the form and try again.", http.StatusInternalServerError)
		return
	}

	if passwordIsCorrect {
		statement := "SELECT user_id, user_name FROM USERS WHERE user_email = ?;"
		err := app.db.QueryRow(statement, userEmail).Scan(&userId, &userName)
		if err != nil {
			log.Print("Error HandleLogin 40: ", err.Error())
			http.Error(w, "Error loading main page", http.StatusInternalServerError)
			return
		}
		session, _ := app.store.Get(r, "cannan-session")
		session.Values["userId"] = userId
		session.Values["userName"] = userName
		session.Values["userEmail"] = userEmail
		session.Values["authenticated"] = true
		session.Save(r, w)
		
		log.Print("Successful login: ", session.Values["userName"], ", ", session.Values["userEmail"], ", ", r.RemoteAddr)	
		app.SendLoginSuccess(w, r, userEmail)
	} else {
		log.Print("Alert HandleLogin 50: Wrong password for ", userEmail, " (", r.RemoteAddr, ")")
		http.Error(w, "Authentication failed. Please check your details and try again.", http.StatusUnauthorized)
		return
	}
}

func (app *application) SendLoginSuccess(w http.ResponseWriter, r *http.Request, user_email string) {
	type notice struct {
		NoticeTitle   template.HTML
		NoticeContent template.HTML
	}
	
	var notices []notice

	statement := "SELECT notice_title, notice_content FROM NOTICES;"
	dbNotices, err := app.db.Query(statement)
	if err != nil {
		log.Print("Error SendLoginSuccess 10: ", err.Error())
		http.Error(w, "Error loading main page", http.StatusInternalServerError)
		return
	}

	for dbNotices.Next() {
		var n notice
		err = dbNotices.Scan(&n.NoticeTitle, &n.NoticeContent)
		if err != nil {
			log.Print("Error SendLoginSuccess 20: ", err.Error())
			http.Error(w, "Error loading main page", http.StatusInternalServerError)
			return
		}
		notices = append(notices, n)
	}

	tmpl, err := template.ParseFiles("../dynamic/main.html")
	if err != nil {
		log.Print("Error SendLoginSuccess 30: ", err.Error())
		http.Error(w, "Error loading main page", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, notices); err != nil {
		log.Print("Error SendLoginSuccess 40: ", err.Error())
		http.Error(w, "Error loading main page", http.StatusInternalServerError)
	}
}