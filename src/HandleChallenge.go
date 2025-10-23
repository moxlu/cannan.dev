package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

func (app *application) HandleGetChallenges(w http.ResponseWriter, r *http.Request) {

	type challenge struct {
		ChallengeId       int
		ChallengeSolved   string
		ChallengeTitle    string
		ChallengeTags     string
		ChallengePoints   string
		ChallengeFeatured bool
		ChallengeHidden   bool
		ChallengeSolves   int
	}

	type story struct {
		StoryId       int
		StoryTitle    string
		StoryPoints   string
		StoryIntro	  template.HTML
		StoryContent  template.HTML
		StoryFeatured bool
	}

	type question struct {
		QuestionId			int
		QuestionOrder       int
		QuestionTextBefore  string
		QuestionPlaceholder string
		QuestionTextAfter   string
	}

	var challengesFeatured []challenge
	var storiesFeatured    []story
	var challengesOther    []challenge
	var storiesOther       []story
	
	session, err := app.store.Get(r, "cannan-session")
	if err != nil {
		log.Print("Alert HandleGetChallenges 10: ", err.Error())
		http.Error(w, "Problem when authenticating session.", http.StatusInternalServerError)
		return
	}

	statement := `
		SELECT challenge_id, challenge_title, challenge_tags, challenge_points, challenge_featured
		FROM CHALLENGES 
		WHERE challenge_hidden = 0
		ORDER BY challenge_points ASC, challenge_title ASC;`

	dbChallenges, err := app.db.Query(statement)
	if err != nil {
		log.Print("Error HandleGetChallenges 20: ", err.Error())
		http.Error(w, "Error loading challenge page", http.StatusInternalServerError)
		return
	}

	for dbChallenges.Next() {
		var c challenge
		var alreadySolved int

		err = dbChallenges.Scan(&c.ChallengeId, &c.ChallengeTitle, &c.ChallengeTags, &c.ChallengePoints, &c.ChallengeFeatured)
		if err != nil {
			log.Print("Error HandleGetChallenges 30: ", err.Error())
			http.Error(w, "Error loading challenge page", http.StatusInternalServerError)
			return
		}

		statement = "SELECT 1 FROM SOLVES WHERE user_id = ? AND challenge_id = ? LIMIT 1;"
		err = app.db.QueryRow(statement, session.Values["userId"], c.ChallengeId).Scan(&alreadySolved)
		if err == sql.ErrNoRows {
			c.ChallengeSolved = ""
		} else if alreadySolved == 1 {
			c.ChallengeSolved = "solved"
		} else if err != nil {
		log.Print("Error HandleGetChallenges 40: ", err.Error())
		http.Error(w, "Could not check if challenge is solved", http.StatusInternalServerError)
		return
	}

		statement = "SELECT COUNT(user_id) FROM SOLVES WHERE challenge_id = ?;"
		err = app.db.QueryRow(statement, c.ChallengeId).Scan(&c.ChallengeSolves)
		if err != nil {
			log.Print("Error HandleGetChallenges 50: ", err.Error())
			http.Error(w, "Error loading challenge page", http.StatusInternalServerError)
			return
		}

		if c.ChallengeFeatured {
			challengesFeatured = append(challengesFeatured, c)
		} else {
			challengesOther = append(challengesOther, c)
		}
	}

	statement = `
		SELECT story_id, story_title, story_intro, story_featured
		FROM STORIES 
		WHERE story_hidden = 0
		ORDER BY story_title ASC;`

	dbStories, err := app.db.Query(statement)
	if err != nil {
		log.Print("Error HandleGetChallenges 60: ", err.Error())
		http.Error(w, "Error loading challenge page", http.StatusInternalServerError)
		return
	}

	for dbStories.Next() {
		var s story
		s.StoryContent = ""

		err = dbStories.Scan(&s.StoryId, &s.StoryTitle, &s.StoryIntro, &s.StoryFeatured)
		if err != nil {
			log.Print("Error HandleGetChallenges 70: ", err.Error())
			http.Error(w, "Error loading challenge page", http.StatusInternalServerError)
			return
		}

		statement = "SELECT COUNT(question_id) FROM QUESTIONS WHERE question_story_id = ?;"
		err = app.db.QueryRow(statement, s.StoryId).Scan(&s.StoryPoints)
		if err != nil {
			log.Print("Error HandleGetChallenges 80: ", err.Error())
			http.Error(w, "Error loading challenge page", http.StatusInternalServerError)
			return
		}

		statement = `
			SELECT question_id, question_order, question_text_before, question_placeholder, question_text_after
			FROM QUESTIONS
			WHERE question_story_id = ?
			ORDER BY question_order ASC, question_id ASC;`

		dbQuestions, err := app.db.Query(statement, s.StoryId)
		if err != nil {
			log.Print("Error HandleGetChallenges 80: ", err.Error())
			http.Error(w, "Error loading challenge page", http.StatusInternalServerError)
			return
		}

		for dbQuestions.Next() {
			var q question
			var alreadySolved int

			err = dbQuestions.Scan(&q.QuestionId, &q.QuestionOrder, &q.QuestionTextBefore, &q.QuestionPlaceholder, &q.QuestionTextAfter)
			if err != nil {
				log.Print("Error HandleGetChallenges 90: ", err.Error())
				http.Error(w, "Error loading challenge page", http.StatusInternalServerError)
				return
			}

			statement = "SELECT 1 FROM QUESTIONS_SOLVED WHERE user_id = ? AND question_id = ? LIMIT 1;"
			err = app.db.QueryRow(statement, session.Values["userId"], q.QuestionId).Scan(&alreadySolved)
			if err == sql.ErrNoRows {
				s.StoryContent += template.HTML(q.QuestionTextBefore)
				
				formHTML := fmt.Sprintf(`
					<form hx-post="/question/%d" hx-target="#story-content-%d">
					<div class="container-input-flag-and-button-flag">
    				<input
        				class="input-flag"
        				type="text"
        				placeholder="%s"
        				name="answer"
        				required
        				autocomplete="off"
        				autocorrect="off"
        				autocapitalize="off"
        				spellcheck="false">
    				<button class="button-flag" type="submit">Submit</button>
					</div></form>`,
        			q.QuestionId,
        			s.StoryId,
        			q.QuestionPlaceholder)

    			s.StoryContent += template.HTML(formHTML)
    			break
			} else if alreadySolved == 1 {
				s.StoryContent += template.HTML(q.QuestionTextAfter)
			} else if err != nil {
				log.Print("Error HandleGetChallenges 100: ", err.Error())
				http.Error(w, "Could not check if question was solved", http.StatusInternalServerError)
				return
			}
		}
		
		if s.StoryFeatured {
			storiesFeatured = append(storiesFeatured, s)
		} else {
			storiesOther = append(storiesOther, s)
		}
	}

	tmpl, err := template.ParseFiles("../dynamic/challenges.html")
	if err != nil {
		log.Print("Error HandleGetChallenges 110: ", err.Error())
		http.Error(w, "Error loading challenge page", http.StatusInternalServerError)
		return
	}

	type pageData struct {
		ChallengesFeatured 	[]challenge
		StoriesFeatured 	[]story
		ChallengesOther    	[]challenge
		StoriesOther 		[]story
	}

	data := pageData{
		ChallengesFeatured: challengesFeatured,
		StoriesFeatured: 	storiesFeatured,
		ChallengesOther:    challengesOther,
		StoriesOther: 		storiesOther,
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Print("Error HandleGetChallenges 120: ", err.Error())
		http.Error(w, "Error loading challenge page", http.StatusInternalServerError)
	}
}

func (app *application) HandleGetChallenge(w http.ResponseWriter, r *http.Request) {
	
	type ChallengeDetails struct {
		Challenge_id          string
		Challenge_title       string
		Challenge_description template.HTML
		Challenge_points      string
		Challenge_result      template.HTML
	}
	
	var c ChallengeDetails
	var alreadySolved int

	session, err := app.store.Get(r, "cannan-session")
	if err != nil {
		log.Printf("invalid session: %v", err)
		session = sessions.NewSession(app.store, "cannan-session")
	}
	user_id := session.Values["userId"]
	c.Challenge_id = r.PathValue("id")

	statement := "SELECT challenge_title, challenge_description, challenge_points FROM CHALLENGES WHERE challenge_id = ?;"
	row := app.db.QueryRow(statement, c.Challenge_id)

	err = row.Scan(&c.Challenge_title, &c.Challenge_description, &c.Challenge_points)
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

func (app *application) HandlePostChallenge(w http.ResponseWriter, r *http.Request) {
	normalise := func(s string) string {
		s = strings.ToLower(strings.TrimSpace(s))
		s = strings.ReplaceAll(s, "\n", "")
		s = strings.ReplaceAll(s, "\r", "")
		return s
	}

	var dbFlags string
	var alreadySolved bool

	session, _ := app.store.Get(r, "cannan-session")
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

	flagList := strings.Split(normalise(dbFlags), ",")

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
