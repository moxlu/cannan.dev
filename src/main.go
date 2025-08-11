package main

import (
	"database/sql"
	"flag"
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
	addr := flag.String("addr", ":4000", "HTTPS network address")
	dsn := flag.String("dsn", "../run/cannan.db", "SQLite database file")
	certFile := flag.String("cert", "../run/fullchain.pem", "TLS certificate file")
	keyFile := flag.String("key", "../run/privkey.pem", "TLS private key file")
	flag.Parse()

	// Start Database
	db, err := sql.Open("sqlite3", *dsn)
	if err != nil {
		log.Fatalf("Database open error: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatalf("Failed to enable foreign keys: %v", err)
	}
	log.Print("Database loaded OK")

	// Session key
	sessionKey, err := os.ReadFile("../run/session.key")
	if err != nil {
		log.Fatalf("Session key read error: %v", err)
		return
	}

	app := &application{
		db:    db,
		store: sessions.NewCookieStore(sessionKey),
	}

	// Routes
	MuxPrimary := http.NewServeMux()
	MuxPrimary.HandleFunc("GET /{$}", app.HandleIndex)
	MuxPrimary.HandleFunc("POST /login", app.HandleLogin)
	MuxPrimary.HandleFunc("GET /challenge/{id}", app.HandleChallengeGet)
	MuxPrimary.HandleFunc("POST /challenge/{id}", app.HandleChallengePost)
	MuxPrimary.HandleFunc("GET /invite/{token}", app.HandleInviteGet)
	MuxPrimary.HandleFunc("POST /invite/{token}", app.HandleInvitePost)
	MuxPrimary.HandleFunc("GET /scores", app.HandleScoresGet)
	MuxPrimary.HandleFunc("GET /brutalpost", HandleBrutalpost)
	MuxPrimary.HandleFunc("GET /lasersharks", HandleLasersharks)

	// Static files
	MuxPrimary.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../static/"))))
	MuxPrimary.Handle("GET /challengeFiles/", http.StripPrefix("/challengeFiles/", http.FileServer(http.Dir("../challengeFiles/"))))
	MuxPrimary.Handle("GET /favicon.ico", http.FileServer(http.Dir("../static/"))) // For default browser grab

	// Catch-all 404
	MuxPrimary.HandleFunc("/", Handle404)

	//Start HTTPS server
	log.Printf("Starting HTTPS server on %s", *addr)
	err = http.ListenAndServeTLS(*addr, *certFile, *keyFile, MuxPrimary)
	if err != nil {
		log.Fatalf("HTTPS server failed: %v", err)
	}
}
