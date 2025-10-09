package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	"github.com/alexedwards/argon2id"
	_ "github.com/mattn/go-sqlite3"
)

type Notice struct {
	Notice_title   template.HTML
	Notice_content template.HTML
}

type Challenge struct {
	Challenge_id       int
	Challenge_title    string
	Challenge_tags     string
	Challenge_points   string
	Challenge_featured bool
	Challenge_hidden   bool
	Challenge_solves   int
}

type PageData struct {
	Notices            []Notice
	ChallengesFeatured []Challenge
	ChallengesOther    []Challenge
}

func (app *application) SendLoginSuccess(w http.ResponseWriter, r *http.Request, user_email string) {
	var notices []Notice
	var challenges_featured []Challenge
	var challenges_other []Challenge
	var user_id int
	var user_name string
	var user_isadmin bool

	statement := "SELECT user_id, user_name, user_isadmin FROM USERS WHERE user_email = ?;"
	err := app.db.QueryRow(statement, user_email).Scan(&user_id, &user_name, &user_isadmin)
	if err != nil {
		log.Print("Error SendLoginSuccess 10: ", err.Error())
		http.Error(w, "Error loading main page", http.StatusInternalServerError)
		return
	}

	session, _ := app.store.Get(r, "cannan-session")
	session.Values["authenticated"] = true
	session.Values["user_email"] = user_email
	session.Values["user_id"] = user_id
	session.Values["user_isadmin"] = user_isadmin
	session.Values["user_name"] = user_name
	session.Values["user_RemoteAddr"] = r.RemoteAddr
	session.Save(r, w)

	log.Print("Successful login: ", session.Values["user_name"], " (", session.Values["user_email"], ")")

	statement = "SELECT notice_title, notice_content FROM NOTICES;"
	rows, err := app.db.Query(statement)
	if err != nil {
		log.Print("Error SendLoginSuccess 20: ", err.Error())
		http.Error(w, "Error loading main page", http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		var n Notice
		err = rows.Scan(&n.Notice_title, &n.Notice_content)
		if err != nil {
			log.Print("Error SendLoginSuccess 30: ", err.Error())
			http.Error(w, "Error loading main page", http.StatusInternalServerError)
			return
		}
		notices = append(notices, n)
	}

	statement = `
		SELECT 
			challenge_id, 
			challenge_title, 
			challenge_tags, 
			challenge_points, 
			challenge_featured, 
			challenge_hidden 
		FROM CHALLENGES 
		ORDER BY challenge_points ASC, challenge_title ASC;
	`

	rows, err = app.db.Query(statement)
	if err != nil {
		log.Print("Error SendLoginSuccess 40: ", err.Error())
		http.Error(w, "Error loading main page", http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		var c Challenge
		err = rows.Scan(&c.Challenge_id, &c.Challenge_title, &c.Challenge_tags, &c.Challenge_points, &c.Challenge_featured, &c.Challenge_hidden)
		if err != nil {
			log.Print("Error SendLoginSuccess 50: ", err.Error())
			http.Error(w, "Error loading main page", http.StatusInternalServerError)
			return
		}

		statement = "SELECT COUNT(user_id) FROM SOLVES WHERE challenge_id = ?;"
		err = app.db.QueryRow(statement, c.Challenge_id).Scan(&c.Challenge_solves)
		if err != nil {
			log.Print("Error SendLoginSuccess 60: ", err.Error())
			http.Error(w, "Error loading main page", http.StatusInternalServerError)
			return
		}

		if c.Challenge_hidden {
			continue
		} else if c.Challenge_featured {
			challenges_featured = append(challenges_featured, c)
		} else {
			challenges_other = append(challenges_other, c)
		}
	}

	tmpl, err := template.ParseFiles("../dynamic/main.html")
	if err != nil {
		log.Print("Error SendLoginSuccess 70: ", err.Error())
		http.Error(w, "Error loading main page", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Notices:            notices,
		ChallengesFeatured: challenges_featured,
		ChallengesOther:    challenges_other,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Print("Error SendLoginSuccess 80: ", err.Error())
		http.Error(w, "Error loading main page", http.StatusInternalServerError)
	}
}

func (app *application) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var storedHash string
	var PasswordIsCorrect bool

	err := r.ParseForm()
	if err != nil {
		log.Print("Error HandleLogin 10: ", err.Error())
		http.Error(w, "Invalid input. Please check the form and try again.", http.StatusInternalServerError)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	err = app.db.QueryRow("SELECT user_hash FROM USERS WHERE user_email = ?", email).Scan(&storedHash)
	if err == sql.ErrNoRows {
		log.Print("Alert HandleLogin 20: Unknown user ", email, " (", r.RemoteAddr, ")")
		http.Error(w, "Authentication failed. Please check your details and try again.", http.StatusUnauthorized)
		return
	} else if err != nil {
		log.Print("Error HandleLogin 21: ", err.Error())
		http.Error(w, "Invalid input. Please check the form and try again.", http.StatusInternalServerError)
		return
	}

	PasswordIsCorrect, err = argon2id.ComparePasswordAndHash(password, storedHash)
	if err != nil {
		log.Print("Error HandleLogin 30: ", err.Error())
		http.Error(w, "Invalid input. Please check the form and try again.", http.StatusInternalServerError)
		return
	}

	if PasswordIsCorrect {
		app.SendLoginSuccess(w, r, email)
	} else {
		log.Print("Alert HandleLogin 40: Wrong password for ", email, " (", r.RemoteAddr, ")")
		http.Error(w, "Authentication failed. Please check your details and try again.", http.StatusUnauthorized)
		return
	}
}
