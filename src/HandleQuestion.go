package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"slices"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func (app *application) HandlePostQuestion(w http.ResponseWriter, r *http.Request) {
		normalise := func(s string) string {
		s = strings.ToLower(strings.TrimSpace(s))
		s = strings.ReplaceAll(s, "\n", "")
		s = strings.ReplaceAll(s, "\r", "")
		return s
	}

	session, err := app.store.Get(r, "cannan-session")
	if err != nil {
		log.Print("Error HandlePostQuestion 10: ", err.Error())
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	err = r.ParseForm()
	if err != nil {
		log.Print("Error HandlePostQuestion 20: ", err.Error())
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	questionIdStr := r.PathValue("id")
	questionId, err := strconv.Atoi(questionIdStr)
	if err != nil {
		log.Print("Error HandlePostQuestion 30: ", err.Error())
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	userId := session.Values["userId"].(int)
	userAnswer := normalise(r.FormValue("answer"))
	var dbAnswers string
	var postscript template.HTML
	var alreadySolved int
	var storyId int

	statement := "SELECT question_answers FROM QUESTIONS WHERE question_id = ?;"
	err = app.db.QueryRow(statement, questionId).Scan(&dbAnswers)
	if err != nil {
		log.Print("Error HandlePostQuestion 30: ", err.Error())
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	answers := strings.Split(normalise(dbAnswers), ",")
	answerIsCorrect := slices.Contains(answers, userAnswer)

	if answerIsCorrect {
		statement = "SELECT 1 FROM QUESTIONS_SOLVED WHERE user_id = ? AND question_id = ? LIMIT 1;"
		err = app.db.QueryRow(statement, userId, questionId).Scan(&alreadySolved)
		if err == sql.ErrNoRows {
			statement = "INSERT INTO QUESTIONS_SOLVED (user_id, question_id) VALUES (?, ?);"
			_, err = app.db.Exec(statement, userId, questionId)
			if err != nil {
				log.Print("Error HandlePostQuestion 40: ", err.Error())
				http.Error(w, "Something went wrong.", http.StatusInternalServerError)
				return
			}
			//
		} else if err != nil {
			log.Print("Error HandlePostQuestion 50: ", err.Error())
			http.Error(w, "Error processing submission", http.StatusInternalServerError)
			return
		}
	} else {
		postscript = template.HTML("<br><h3>" + r.FormValue("answer") + " is not correct.</h3>")
	}
	
	statement = "SELECT question_story_id FROM QUESTIONS WHERE question_id = ? LIMIT 1;"
	err = app.db.QueryRow(statement, questionId).Scan(&storyId)
	if err != nil {
		log.Print("Error HandlePostQuestion 60: ", err.Error())
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}

	type question struct {
		questionId			int
		questionOrder       int
		questionTextBefore  string
		questionPlaceholder string
		questionTextAfter   string
	}

		var storyContent template.HTML

	statement = `
			SELECT question_id, question_order, question_text_before, question_placeholder, question_text_after
			FROM QUESTIONS
			WHERE question_story_id = ?
			ORDER BY question_order ASC, question_id ASC;`

		dbQuestions, err := app.db.Query(statement, storyId)
		if err != nil {
			log.Print("Error HandlePostQuestion 70: ", err.Error())
			http.Error(w, "Something went wrong.", http.StatusInternalServerError)
			return
		}
		defer dbQuestions.Close()

		for dbQuestions.Next() {
			var q question
			var alreadySolved int

			err = dbQuestions.Scan(&q.questionId, &q.questionOrder, &q.questionTextBefore, &q.questionPlaceholder, &q.questionTextAfter)
			if err != nil {
				log.Print("Error HandlePostQuestion 80: ", err.Error())
				http.Error(w, "Something went wrong.", http.StatusInternalServerError)
				return
			}

			statement = "SELECT 1 FROM QUESTIONS_SOLVED WHERE user_id = ? AND question_id = ? LIMIT 1;"
			err = app.db.QueryRow(statement, userId, q.questionId).Scan(&alreadySolved)
			if err == sql.ErrNoRows {
				storyContent += template.HTML(q.questionTextBefore)
				
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
        			q.questionId,
        			storyId,
        			q.questionPlaceholder)

    			storyContent += template.HTML(formHTML)
				storyContent += postscript
    			break
			} else if alreadySolved == 1 {
				storyContent += template.HTML(q.questionTextAfter)
			} else if err != nil {
				log.Print("Error HandlePostQuestion 0: ", err.Error())
				http.Error(w, "Something went wrong.", http.StatusInternalServerError)
				return
			}
		}

		w.Write([]byte(storyContent))
}