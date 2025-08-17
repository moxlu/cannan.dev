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
	Challenge_result      template.HTML
}

func (app *application) HandleChallengeGet(w http.ResponseWriter, r *http.Request) {
	var c ChallengeDetails
	var alreadySolved int

	session, _ := app.store.Get(r, "session-name")
	user_id := session.Values["user_id"]
	c.Challenge_id = r.PathValue("id")

	statement := "SELECT challenge_title, challenge_description, challenge_points FROM CHALLENGES WHERE challenge_id = ?;"
	row := app.db.QueryRow(statement, c.Challenge_id)

	err := row.Scan(&c.Challenge_title, &c.Challenge_description, &c.Challenge_points)
	if err != nil {
		log.Print("Error HandleChallengeGet 10: ", err.Error())
		http.Error(w, "Could not fetch challenge details", http.StatusInternalServerError)
		return
	}

	statement = "SELECT 1 FROM SOLVES WHERE user_id = ? AND challenge_id = ? LIMIT 1;"
	err = app.db.QueryRow(statement, user_id, c.Challenge_id).Scan(&alreadySolved)
	if err == sql.ErrNoRows {
		c.Challenge_result = ""
	} else if alreadySolved == 1 {
		c.Challenge_result = "<h3>You have already solved this challenge.</h3>"
	} else if err != nil {
		log.Print("Error HandleChallengeGet 20: ", err.Error())
		http.Error(w, "Could not check if challenge is solved", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("../dynamic/challenge.html")
	if err != nil {
		log.Print("Error HandleChallengeGet 30: ", err.Error())
		http.Error(w, "Could not display challenge details", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, c); err != nil {
		log.Print("Error HandleChallengeGet 40: ", err.Error())
		http.Error(w, "Could not display challenge details", http.StatusInternalServerError)
	}
}

func (app *application) HandleChallengePost(w http.ResponseWriter, r *http.Request) {
	normalise := func(s string) string {
		s = strings.ToLower(strings.TrimSpace(s))
		s = strings.ReplaceAll(s, "\n", "")
		s = strings.ReplaceAll(s, "\r", "")
		return s
	}

	var dbFlags string
	var alreadySolved bool

	session, _ := app.store.Get(r, "session-name")
	user_id := session.Values["user_id"]
	user_email := session.Values["user_email"]

	if !session.Values["authenticated"].(bool) {
		log.Print("Alert HandleChallengePost 10: Submitter not authorised")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	err := r.ParseForm()
	if err != nil {
		log.Print("Error HandleChallengePost 20: ", err.Error())
		http.Error(w, "Error processing submission", http.StatusInternalServerError)
		return
	}

	userFlag := normalise(r.FormValue("flag"))

	challenge_id := r.PathValue("id")
	statement := "SELECT challenge_flags FROM CHALLENGES WHERE challenge_id = ?;"
	err = app.db.QueryRow(statement, challenge_id).Scan(&dbFlags)
	if err != nil {
		log.Print("Error HandleChallengePost 30: ", err.Error())
		http.Error(w, "Error processing submission", http.StatusInternalServerError)
		return
	}

	flagList := strings.Split(dbFlags, ",")

	flagIsCorrect := false
	for _, f := range flagList {
		if normalise(f) == userFlag {
			flagIsCorrect = true
			break
		}
	}

	if flagIsCorrect {
		statement = "SELECT 1 FROM SOLVES WHERE user_id = ? AND challenge_id = ? LIMIT 1;"
		err = app.db.QueryRow(statement, user_id, challenge_id).Scan(&alreadySolved)
		if err == sql.ErrNoRows {
			statement = "INSERT INTO SOLVES (user_id, challenge_id) VALUES (?, ?);"
			_, err = app.db.Exec(statement, user_id, challenge_id)
			if err != nil {
				log.Print("Error HandleChallengePost 40: ", err.Error())
				http.Error(w, "Error processing submission", http.StatusInternalServerError)
				return
			}
			log.Print(user_email, " solved Challenge ", challenge_id)
			w.Write([]byte("<h3>Solved! Well done!</h3>"))

		} else if err != nil {
			log.Print("Error HandleChallengePost 50: ", err.Error())
			http.Error(w, "Error processing submission", http.StatusInternalServerError)
			return
		} else {
			log.Print(user_email, " tried to re-solve Challenge ", challenge_id)
			w.Write([]byte("<h3>You have already solved this challenge.</h3>"))
		}
	} else {
		log.Print(user_email, " did not solve Challenge ", challenge_id)
		w.Write([]byte("<h3>Incorrect. Keep trying!</h3>"))
	}
}
