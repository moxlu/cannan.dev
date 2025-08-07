package main

import (
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
	Challenge_id     string
	Challenge_title  string
	Challenge_points string
}

type PageData struct {
	Notices    []Notice
	Challenges []Challenge
}

func (app *application) SendLoginSuccess(w http.ResponseWriter, r *http.Request, email string) {
	var notices []Notice
	var challenges []Challenge

	session, _ := app.store.Get(r, "session-name")
	session.Values["authenticated"] = true
	session.Values["email"] = email
	session.Save(r, w)

	log.Print("Successful login: ", email)

	statement := "SELECT notice_title, notice_content FROM NOTICES;"
	rows, err := app.db.Query(statement)
	if err != nil {
		log.Print("Error SendLoginSuccess() 100 - Failed fetching notices")
		http.Error(w, "Login successful but error loading main page", http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		var n Notice
		err = rows.Scan(&n.Notice_title, &n.Notice_content)
		if err != nil {
			log.Print("Error SendLoginSuccess() 200 - Failed processing notices")
			http.Error(w, "Login successful but error loading main page", http.StatusInternalServerError)
			return
		}
		notices = append(notices, n)
	}

	statement = "SELECT challenge_id, challenge_title, challenge_points FROM CHALLENGES;"
	rows, err = app.db.Query(statement)
	if err != nil {
		log.Print("Error SendLoginSuccess() 300 - Failed fetching challenges")
		http.Error(w, "Login successful but error loading main page", http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		var c Challenge
		err = rows.Scan(&c.Challenge_id, &c.Challenge_title, &c.Challenge_points)
		if err != nil {
			log.Print("Error SendLoginSuccess() 400 - Failed processing challenges")
			http.Error(w, "Login successful but error loading main page", http.StatusInternalServerError)
			return
		}
		challenges = append(challenges, c)
	}

	data := PageData{
		Notices:    notices,
		Challenges: challenges,
	}

	tmpl, err := template.ParseFiles("../dynamic/login_success.html")
	if err != nil {
		log.Print("Error SendLoginSuccess() 500 - Failed parsing template")
		http.Error(w, "Login successful but error loading main page", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Print("Error SendLoginSuccess() 600 - Failed merging and serving template")
		log.Print(err.Error())
		http.Error(w, "Execution error", http.StatusInternalServerError)
	}
}

func (app *application) SendLoginFail(w http.ResponseWriter, IP string, email string) {
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

func (app *application) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var storedHash string
	var PasswordIsCorrect bool

	IP := r.RemoteAddr
	err := r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")

	if err != nil {
		log.Print("Error HandleLogin() 100 - Failed parsing /login form")
		app.SendLoginFail(w, IP, email)
		return
	}

	err = app.db.QueryRow("SELECT user_hash FROM USERS WHERE user_email = ?", email).Scan(&storedHash)
	if err != nil {
		log.Print("Error HandleLogin() 200 - Wrong email or failed SQL query to find or get user hash based on provided email")
		app.SendLoginFail(w, IP, email)
		return
	}

	PasswordIsCorrect, err = argon2id.ComparePasswordAndHash(password, storedHash)
	if err != nil {
		log.Print("Error HandleLogin() 300 - Failed Argon2id comparing password and hash")
		app.SendLoginFail(w, IP, email)
		return
	}

	if PasswordIsCorrect {
		app.SendLoginSuccess(w, r, email)
	} else {
		log.Print("Alert HandleLogin() 400 - Wrong password")
		app.SendLoginFail(w, IP, email)
		return
	}
}
