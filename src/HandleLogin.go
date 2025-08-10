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
		log.Print(err.Error())
		log.Print("Error SendLoginSuccess() 50 - Failed looking up user_id")
		http.Error(w, "Login successful but error loading main page", http.StatusInternalServerError)
		return
	}

	session, _ := app.store.Get(r, "session-name")
	session.Values["authenticated"] = true
	session.Values["user_id"] = user_id
	session.Values["user_email"] = user_email
	session.Values["user_name"] = user_name
	session.Values["user_isadmin"] = user_isadmin
	session.Save(r, w)

	log.Print("Successful login: ", session.Values["user_name"], " (", session.Values["user_email"], ")")

	statement = "SELECT notice_title, notice_content FROM NOTICES;"
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

	statement = "SELECT challenge_id, challenge_title, challenge_tags, challenge_points, challenge_featured, challenge_hidden FROM CHALLENGES;"
	rows, err = app.db.Query(statement)
	if err != nil {
		log.Print(err.Error())
		log.Print("Error SendLoginSuccess() 300 - Failed fetching challenges from db")
		http.Error(w, "Login successful but error loading main page", http.StatusInternalServerError)
		return
	}

	for rows.Next() {
		var c Challenge
		err = rows.Scan(&c.Challenge_id, &c.Challenge_title, &c.Challenge_tags, &c.Challenge_points, &c.Challenge_featured, &c.Challenge_hidden)
		if err != nil {
			log.Print("Error SendLoginSuccess() 400 - Failed scanning challenges")
			http.Error(w, "Login successful but error loading main page", http.StatusInternalServerError)
			return
		}

		statement = "SELECT COUNT(user_id) FROM SOLVES WHERE challenge_id = ?;"
		err = app.db.QueryRow(statement, c.Challenge_id).Scan(&c.Challenge_solves)
		if err != nil {
			log.Print("Error SendLoginSuccess() 450 - Failed counting solves")
			http.Error(w, "Login successful but error loading main page", http.StatusInternalServerError)
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

	tmpl, err := template.ParseFiles("../dynamic/login_success.html")
	if err != nil {
		log.Print("Error SendLoginSuccess() 500 - Failed parsing template")
		http.Error(w, "Login successful but error loading main page", http.StatusInternalServerError)
		return
	}

	data := PageData{
		Notices:            notices,
		ChallengesFeatured: challenges_featured,
		ChallengesOther:    challenges_other,
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
