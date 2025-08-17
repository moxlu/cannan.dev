package main

import (
	"database/sql"
	"flag"
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

func main() {
	// Default flags for dev, prod flags are set in cannan.service
	addr := flag.String("addr", ":4000", "HTTPS network address")
	dsn := flag.String("dsn", "../run/cannan.db?_loc=auto&parseTime=true", "SQLite database file")
	certFile := flag.String("certFile", "../run/fullchain.pem", "TLS certificate file")
	keyFile := flag.String("keyFile", "../run/privkey.pem", "TLS private key file")
	flag.Parse()

	// Start Database
	db, err := sql.Open("sqlite3", *dsn)
	if err != nil {
		log.Fatalf("Error (Fatal) Main 10: %v", err.Error())
		return
	}
	defer db.Close()

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatalf("Error (Fatal) Main 20: %v", err.Error())
		return
	}
	log.Print("Database loaded OK")

	// Session key
	sessionKey, err := os.ReadFile("../run/session.key")
	if err != nil {
		log.Fatalf("Error (Fatal) Main 30: %v", err.Error())
		return
	}

	app := &application{
		db:    db,
		store: sessions.NewCookieStore(sessionKey),
	}

	// Routes
	MuxPrimary := http.NewServeMux()
	MuxPrimary.HandleFunc("GET /{$}", app.HandleIndex)
	MuxPrimary.HandleFunc("GET /invite/{token}", app.HandleGetInvite)
	MuxPrimary.HandleFunc("POST /invite/{token}", app.HandlePostInvite)
	MuxPrimary.HandleFunc("POST /login", app.HandleLogin)
	MuxPrimary.HandleFunc("POST /reset_initiate", app.HandleInitiateReset)
	MuxPrimary.HandleFunc("GET /reset/{token}", app.HandleGetResetForm)
	MuxPrimary.HandleFunc("POST /reset/{token}", app.HandlePostResetForm)

	MuxPrimary.HandleFunc("GET /challenge/{id}", app.HandleGetChallenge)
	MuxPrimary.HandleFunc("POST /challenge/{id}", app.HandlePostChallenge)
	MuxPrimary.HandleFunc("GET /scores", app.HandleGetScores)
	MuxPrimary.HandleFunc("GET /brutalpost", HandleGetBrutalpost)
	MuxPrimary.HandleFunc("GET /lasersharks", HandleGetLasersharks)

	// Static files - note users can lookup directory
	MuxPrimary.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("../static/"))))
	MuxPrimary.Handle("GET /challengeFiles/", http.StripPrefix("/challengeFiles/", http.FileServer(http.Dir("../challengeFiles/"))))
	MuxPrimary.Handle("GET /favicon.ico", http.FileServer(http.Dir("../static/"))) // For default browser grab

	// Catch-all
	MuxPrimary.HandleFunc("/", Handle404)

	//Start HTTPS server
	log.Printf("Starting HTTPS server on %s", *addr)
	err = http.ListenAndServeTLS(*addr, *certFile, *keyFile, MuxPrimary)
	if err != nil {
		log.Fatalf("Error (Fatal) Main 40: %v", err.Error())
	}
}
