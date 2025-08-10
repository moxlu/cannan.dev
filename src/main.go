package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	_ "github.com/mattn/go-sqlite3"
)

type application struct {
	db    *sql.DB
	store *sessions.CookieStore
}

func (app *application) HandleIndex(w http.ResponseWriter, r *http.Request) {
	log.Print(r.RemoteAddr + " " + r.Method + " " + r.URL.String())

	response, err := template.ParseFiles("../dynamic/index.html")
	if err != nil {
		log.Print("Error HandleIndex() 100 - Failed parsing index.html")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = response.Execute(w, nil)
	if err != nil {
		log.Print("Error HandleIndex() 200 - Failed serving index.html")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func Handle404(w http.ResponseWriter, r *http.Request) {
	log.Print("Alert 404 - " + r.RemoteAddr + " " + r.Method + " " + r.URL.String())
	http.Error(w, "Page not found", http.StatusNotFound)
}

func main() {
	// Start Database
	db, err := sql.Open("sqlite3", "../run/cannan.db")
	if err != nil {
		log.Print(err.Error())
		return
	}

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Print(err.Error())
		return
	} else {
		log.Print("Database loaded OK")
	}
	defer db.Close()

	key1, err := os.ReadFile("../run/keys.txt")
	if err != nil {
		log.Print(err.Error())
		return
	}

	app := &application{
		db:    db,
		store: sessions.NewCookieStore(key1),
	}

	// Start mux and handlers
	MuxPrimary := http.NewServeMux()
	MuxPrimary.HandleFunc("GET /{$}", app.HandleIndex)
	MuxPrimary.HandleFunc("POST /login", app.HandleLogin)
	MuxPrimary.HandleFunc("GET /challenge/{id}", app.HandleChallengeGet)
	MuxPrimary.HandleFunc("POST /challenge/{id}", app.HandleChallengePost)
	MuxPrimary.HandleFunc("GET /invite/{token}", app.HandleInviteGet)
	MuxPrimary.HandleFunc("POST /invite/{token}", app.HandleInvitePost)
	MuxPrimary.HandleFunc("GET /brutalpost", HandleBrutalpost)
	MuxPrimary.HandleFunc("GET /lasersharks", HandleLasersharks)

	// Fileserver for specified static files only
	MuxPrimary.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../static/"))))
	MuxPrimary.Handle("GET /challengeFiles/", http.StripPrefix("/challengeFiles/", http.FileServer(http.Dir("../challengeFiles/"))))
	MuxPrimary.Handle("GET /favicon.ico", http.FileServer(http.Dir("../static/"))) // For default browser grab

	// Everything else should go to 404
	MuxPrimary.HandleFunc("/", Handle404)

	log.Print("Starting server on :4000")
	err = http.ListenAndServe(":4000", MuxPrimary)
	log.Print(err.Error())
}
