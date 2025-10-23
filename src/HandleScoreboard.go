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

func (app *application) HandleGetScoreboard(w http.ResponseWriter, r *http.Request) {
	var score_overall []ScoreboardEntry
	var score_featured []ScoreboardEntry

	query := `
	SELECT 
		u.user_id,
		u.user_name,
		u.user_description,
		COALESCE(SUM(points), 0) AS overall_score,
		COALESCE(SUM(CASE WHEN is_featured = 1 THEN points ELSE 0 END), 0) AS featured_score,
		MAX(last_solve_datetime) AS last_solve_overall,
		MAX(CASE WHEN is_featured = 1 THEN last_solve_datetime ELSE NULL END) AS last_solve_featured
	FROM USERS u
	LEFT JOIN (
		SELECT 
			s.user_id,
			c.challenge_points AS points,
			c.challenge_featured AS is_featured,
			s.solve_datetime AS last_solve_datetime
		FROM SOLVES s
		JOIN CHALLENGES c ON s.challenge_id = c.challenge_id

		UNION ALL

		SELECT
			qs.user_id,
			1 AS points,
			st.story_featured AS is_featured,
			qs.solve_datetime AS last_solve_datetime
		FROM QUESTIONS_SOLVED qs
		JOIN QUESTIONS q ON qs.question_id = q.question_id
		JOIN STORIES st ON q.question_story_id = st.story_id
	) AS combined ON combined.user_id = u.user_id
	WHERE u.user_isdeactivated = 0
	AND u.user_ishidden = 0
	GROUP BY u.user_id, u.user_name, u.user_description
	ORDER BY overall_score DESC, last_solve_overall ASC;
	`

	rows, err := app.db.Query(query)
	if err != nil {
		log.Print("Error HandleGetScoreboard 10: ", err.Error())
		http.Error(w, "Error processing scores", http.StatusInternalServerError)
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
			log.Print("Error HandleGetScoreboard 20: ", err.Error())
			http.Error(w, "Error processing scores", http.StatusInternalServerError)
			return
		}
		entry.UserRank = rank
		rank++
		score_overall = append(score_overall, entry)
	}

	// Copy overall scoreboard to featured scoreboard
	score_featured = make([]ScoreboardEntry, 0, len(score_overall)) // Start with 0 length but reserve capacity

	// Only include users with featured points > 0
	for _, entry := range score_overall {
		if entry.UserScoreFeatured > 0 {
			score_featured = append(score_featured, entry)
		}
	}

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
		log.Print("Error HandleGetScoreboard 30: ", err.Error())
		http.Error(w, "Error displaying scores", http.StatusInternalServerError)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Print("Error HandleGetScroeboard 40: ", err.Error())
		http.Error(w, "Error displaying scores", http.StatusInternalServerError)
	}
}