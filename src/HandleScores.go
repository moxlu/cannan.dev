package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"sort"

	_ "github.com/mattn/go-sqlite3"
)

type ScoreboardEntry struct {
	UserRank              int
	UserID                int
	UserName              string
	UserDescription       sql.NullString
	UserScoreFeatured     int
	UserScoreOverall      int
	UserLastSolveFeatured sql.NullTime // for future use
	UserLastSolveOverall  sql.NullTime // for future use
}

type ScoreData struct {
	ScoreboardOverall  []ScoreboardEntry
	ScoreboardFeatured []ScoreboardEntry
}

func (app *application) HandleScoresGet(w http.ResponseWriter, r *http.Request) {
	var score_overall []ScoreboardEntry
	var score_featured []ScoreboardEntry

	query := `
    SELECT 
        u.user_id,
        u.user_name,
		u.user_description,
        COALESCE(SUM(c.challenge_points), 0) AS overall_score,
		COALESCE(SUM(CASE WHEN c.challenge_featured = 1 THEN c.challenge_points ELSE 0 END), 0) AS featured_score
    FROM USERS u
    LEFT JOIN SOLVES s ON u.user_id = s.user_id
    LEFT JOIN CHALLENGES c ON s.challenge_id = c.challenge_id
    WHERE u.user_isdeactivated = 0
      AND u.user_ishidden = 0
    GROUP BY u.user_id, u.user_name
    ORDER BY overall_score DESC;
    `
	rows, err := app.db.Query(query)
	if err != nil {
		log.Print(err.Error())
		log.Print("Error HandleScoresGet 100 - Failed querying db for scores")
		return
	}
	defer rows.Close()

	rank := 1
	for rows.Next() {
		var entry ScoreboardEntry
		err := rows.Scan(
			&entry.UserID,
			&entry.UserName,
			&entry.UserDescription,
			&entry.UserScoreOverall,
			&entry.UserScoreFeatured,
		)
		if err != nil {
			log.Print(err.Error())
			log.Print("Error HandleScoresGet 200 - Failed scanning db score values")
			return
		}
		entry.UserRank = rank
		rank++
		score_overall = append(score_overall, entry)
	}

	score_featured = make([]ScoreboardEntry, len(score_overall))
	copy(score_featured, score_overall)
	// because otherwise both objects point to the same underlying array

	// Sort featured scoreboard descending by UserScoreFeatured
	sort.Slice(score_featured, func(i, j int) bool {
		return score_featured[i].UserScoreFeatured > score_featured[j].UserScoreFeatured
	})

	// Re-assign ranks based on new order
	for i := range score_featured {
		score_featured[i].UserRank = i + 1
	}

	data := ScoreData{
		ScoreboardOverall:  score_overall,
		ScoreboardFeatured: score_featured,
	}

	tmpl, err := template.ParseFiles("../dynamic/scoreboards.html")
	if err != nil {
		log.Print("Error HandleScoresGet() 300 - Failed parsing template")
		http.Error(w, "Login successful but error loading main page", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Print("Error HandleScoresGet() 400 - Failed merging and serving template")
		log.Print(err.Error())
		http.Error(w, "Execution error", http.StatusInternalServerError)
	}
}
