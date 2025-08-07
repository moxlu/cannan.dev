package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type ChallengeDetails struct {
	Challenge_id          string
	Challenge_title       string
	Challenge_description template.HTML
	Challenge_points      string
}

func (app *application) HandleChallengeGet(w http.ResponseWriter, r *http.Request) {
	var c ChallengeDetails

	c.Challenge_id = r.PathValue("id")
	log.Print("GET Challenge ", c.Challenge_id)

	statement := "SELECT challenge_title, challenge_description, challenge_points FROM CHALLENGES WHERE challenge_id = ?;"
	row := app.db.QueryRow(statement, c.Challenge_id)

	err := row.Scan(&c.Challenge_title, &c.Challenge_description, &c.Challenge_points)
	if err != nil {
		log.Print("Error HandleChallengeGet() 100 - Failed fetching challenge details")
		http.Error(w, "Login successful but error loading main page", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("../dynamic/challenge.html")
	if err != nil {
		log.Print("Error HandleChallengeGet() 200 - Failed parsing template")
		http.Error(w, "Login successful but error loading main page", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, c); err != nil {
		log.Print("Error HandleChallengeGet() 300 - Failed merging and serving template")
		log.Print(err.Error())
		http.Error(w, "Execution error", http.StatusInternalServerError)
	}

}

func (app *application) HandleChallengePost(w http.ResponseWriter, r *http.Request) {
	var dbFlags string
	var user_id int
	var alreadySolved bool
	var email string

	session, _ := app.store.Get(r, "session-name")

	auth, authok := session.Values["authenticated"].(bool)
	email, emailok := session.Values["email"].(string)
	email = strings.TrimSpace(email)

	if !auth || !authok || !emailok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	userFlag := r.FormValue("flag")
	if err != nil {
		log.Print("Error HandleChallengePost() 100 - Couldn't parse user flag")
		return
	}

	challenge_id := r.PathValue("id")
	statement := "SELECT challenge_flags FROM CHALLENGES WHERE challenge_id = ?;"
	err = app.db.QueryRow(statement, challenge_id).Scan(&dbFlags)
	if err != nil {
		log.Print("Error HandleChallengePost() 200 - Failed fetching flags from db")
		http.Error(w, "Login successful but error loading main page", http.StatusInternalServerError)
		return
	}

	flagList := strings.Split(dbFlags, ",")

	flagIsCorrect := false
	for _, f := range flagList {
		if strings.TrimSpace(f) == userFlag {
			flagIsCorrect = true
			break
		}
	}

	if flagIsCorrect {
		statement := "SELECT user_id FROM USERS WHERE user_email = ?;"
		err = app.db.QueryRow(statement, email).Scan(&user_id)
		log.Print("Email appears to be ", email)
		log.Print("User id appears to be ", user_id)
		if err != nil {
			log.Print("Error HandleChallengePost() 300 - Couldn't find user_id in db")
			w.Write([]byte("Error :("))
			return
		}

		statement = "SELECT 1 FROM SOLVES WHERE user_id = ? AND challenge_id = ? LIMIT 1;"
		err = app.db.QueryRow(statement, user_id, challenge_id).Scan(&alreadySolved)
		if err == sql.ErrNoRows {
			statement = "INSERT INTO SOLVES (user_id, challenge_id) VALUES (?, ?);"
			_, err = app.db.Exec(statement, user_id, challenge_id)
			if err != nil {
				log.Print("Error HandleChallengePost() 400 - Problem recording solved flag")
				w.Write([]byte("Error :("))
				return
			}
			log.Print(email, " solved Challenge ", challenge_id)
			w.Write([]byte("<br>Solved! Well done!"))

		} else if err != nil {
			log.Print("Error HandleChallengePost() 500 - Problem recording solved flag")
			w.Write([]byte("Error :("))
			return
		} else {
			log.Print(email, " tried to re-solve Challenge ", challenge_id)
			w.Write([]byte("<br>You have already solved this challenge."))
		}
	} else {
		log.Print(email, " did not solve Challenge ", challenge_id)
		w.Write([]byte("<br>Incorrect. Keep trying!"))
	}

}
