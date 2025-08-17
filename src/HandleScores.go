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
	UserLastSolveFeatured sql.NullString
	UserLastSolveOverall  sql.NullString
}

type ScoreData struct {
	ScoreboardOverall  []ScoreboardEntry
	ScoreboardFeatured []ScoreboardEntry
}

func (app *application) HandleScoresGet(w http.ResponseWriter, r *http.Request) {
	var score_overall []ScoreboardEntry
	var score_featured []ScoreboardEntry

	// Intent:
	query := `
    SELECT 
        u.user_id,
        u.user_name,
		u.user_description,
        COALESCE(SUM(c.challenge_points), 0) AS overall_score,
		COALESCE(SUM(CASE WHEN c.challenge_featured = 1 THEN c.challenge_points ELSE 0 END), 0) AS featured_score,
		MAX(s.solve_datetime) AS last_solve_overall,
    	MAX(CASE WHEN c.challenge_featured = 1 THEN s.solve_datetime ELSE NULL END) AS last_solve_featured
    FROM USERS u
    LEFT JOIN SOLVES s ON u.user_id = s.user_id
    LEFT JOIN CHALLENGES c ON s.challenge_id = c.challenge_id
    WHERE u.user_isdeactivated = 0
      AND u.user_ishidden = 0
    GROUP BY u.user_id, u.user_name
    ORDER BY overall_score DESC, last_solve_overall ASC;
    `
	rows, err := app.db.Query(query)
	if err != nil {
		log.Print("Error HandleScoresGet 100: ", err.Error())
		http.Error(w, "Error processing scores", http.StatusInternalServerError)
		// Do I need rows.Close() here?
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
			&entry.UserLastSolveOverall,
			&entry.UserLastSolveFeatured,
		)
		if err != nil {
			log.Print("Error HandleScoresGet 200: ", err.Error())
			http.Error(w, "Error processing scores", http.StatusInternalServerError)
			return
		}
		entry.UserRank = rank
		rank++
		score_overall = append(score_overall, entry)
	}

	// Copy overall scoreboard to featured scoreboard
	score_featured = make([]ScoreboardEntry, len(score_overall))
	copy(score_featured, score_overall)

	// Sort featured scoreboard by UserScoreFeatured (descending) and UserLastSolveFeatured (ascending)
	sort.Slice(score_featured, func(i, j int) bool {
		if score_featured[i].UserScoreFeatured != score_featured[j].UserScoreFeatured {
			return score_featured[i].UserScoreFeatured > score_featured[j].UserScoreFeatured
		}

		// For equal scores, sort by earlier last solve time (ascending)
		// Since datetime strings in SQLite format sort lexicographically in chronological order
		if score_featured[i].UserLastSolveFeatured.Valid && score_featured[j].UserLastSolveFeatured.Valid {
			return score_featured[i].UserLastSolveFeatured.String < score_featured[j].UserLastSolveFeatured.String
		}

		// Handle cases where one or both timestamps are NULL
		if !score_featured[i].UserLastSolveFeatured.Valid && !score_featured[j].UserLastSolveFeatured.Valid {
			return false // Maintain order if both are NULL
		}
		return score_featured[i].UserLastSolveFeatured.Valid // NULLs last
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
		log.Print("Error HandleScoresGet 300: ", err.Error())
		http.Error(w, "Error displaying scores", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Print("Error HandleScoresGet 400: ", err.Error())
		http.Error(w, "Error displaying scores", http.StatusInternalServerError)
	}
}
