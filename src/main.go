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
	certFile := flag.String("certFile", "../run/fullchain.pem", "TLS certificate file")
	keyFile := flag.String("keyFile", "../run/privkey.pem", "TLS private key file")
	flag.Parse()

	// Start Databases
	dbCannan, err := sql.Open("sqlite3", "../run/cannan.db?_loc=auto&parseTime=true&_busy_timeout=5000")
	if err != nil {
		log.Fatalf("Fatal: failed to open database: %v", err)
	}
	defer dbCannan.Close()

	pragmaStatements := []string{
		"PRAGMA journal_mode = WAL;",
		"PRAGMA synchronous = NORMAL;",
		"PRAGMA foreign_keys = ON;",
	}

	for _, stmt := range pragmaStatements {
		if _, err := dbCannan.Exec(stmt); err != nil {
			log.Fatalf("Fatal: failed to execute %s: %v", stmt, err)
		}
	}
	log.Print("Cannan database loaded successfully.")

	dbPossum, err := sql.Open("sqlite3", "../possum/possum.db?_loc=auto&parseTime=true")
	if err != nil {
		log.Fatalf("Error (Fatal) Main 30: %v", err.Error())
		return
	}
	defer dbPossum.Close()

	if _, err := dbPossum.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Fatalf("Error (Fatal) Main 40: %v", err.Error())
		return
	}
	log.Print("Possum Database loaded OK")

	// Load session keys
	sessionKeyCannan, err := os.ReadFile("../run/session.key")
	if err != nil {
		log.Fatalf("Error (Fatal) Main 50: %v", err.Error())
		return
	}

	sessionKeyPossum, err := os.ReadFile("../possum/possum_session.key")
	if err != nil {
		log.Fatalf("Error (Fatal) Main 60: %v", err.Error())
		return
	}

	appCannan := &application{
		db:    dbCannan,
		store: sessions.NewCookieStore(sessionKeyCannan),
	}

	appPossum := &application{
		db:    dbPossum,
		store: sessions.NewCookieStore(sessionKeyPossum),
	}

	initPossum() // Load Possum flags

	// Routes
	MuxPrimary := http.NewServeMux()
	MuxPrimary.HandleFunc("GET /{$}", appCannan.HandleIndex)
	MuxPrimary.HandleFunc("GET /invite/{token}", appCannan.HandleGetInvite)
	MuxPrimary.HandleFunc("POST /invite/{token}", appCannan.HandlePostInvite)
	MuxPrimary.HandleFunc("POST /login", appCannan.HandleLogin)
	MuxPrimary.HandleFunc("POST /reset_initiate", appCannan.HandleInitiateReset)
	MuxPrimary.HandleFunc("GET /reset/{token}", appCannan.HandleGetResetForm)
	MuxPrimary.HandleFunc("POST /reset/{token}", appCannan.HandlePostResetForm)

	MuxPrimary.HandleFunc("GET /challenges", appCannan.HandleGetChallenges)
	MuxPrimary.HandleFunc("GET /challenge/{id}", appCannan.HandleGetChallenge)
	MuxPrimary.HandleFunc("POST /challenge/{id}", appCannan.HandlePostChallenge)
	MuxPrimary.HandleFunc("GET /scoreboard", appCannan.HandleGetScoreboard)
	MuxPrimary.HandleFunc("POST /question/{id}", appCannan.HandlePostQuestion)

	// PossumAI, uses different DB and session key
	MuxPrimary.Handle("GET /possumAI/static/", http.StripPrefix("/possumAI/static/", http.FileServer(http.Dir("../possum/static/"))))
	MuxPrimary.HandleFunc("GET /possumAI", appPossum.HandleGetPossumIndex)
	MuxPrimary.HandleFunc("POST /possumAI/login", appPossum.HandlePostPossumLogin)
	MuxPrimary.HandleFunc("POST /possumAI/enquiry", appPossum.HandlePostPossumEnquiry)
	MuxPrimary.HandleFunc("POST /possumAI/verifycookie", appPossum.HandlePostVerifyCookie)

	// Static files
	MuxPrimary.Handle("GET /cannan.css", http.FileServer(http.Dir("../static/")))
	MuxPrimary.Handle("GET /cannan.js", http.FileServer(http.Dir("../static/")))
	MuxPrimary.Handle("GET /favicon.ico", http.FileServer(http.Dir("../static/")))
	MuxPrimary.Handle("GET /robots.txt", http.FileServer(http.Dir("../static/")))
	
	// Disable challengeFiles directory listing
	MuxPrimary.HandleFunc("GET /challengeFiles/{$}", Handle404)
	// but serve individual challengeFiles
	MuxPrimary.Handle("GET /challengeFiles/", http.StripPrefix("/challengeFiles/", http.FileServer(http.Dir("../challengeFiles/"))))
	// TODO: Handle404 /challengeFiles/doesnotexist

	// Catch-all 404 (includes logging)
	MuxPrimary.HandleFunc("/", Handle404)

	// Start HTTPS server
	log.Printf("Starting HTTPS server on %s", *addr)
	err = http.ListenAndServeTLS(*addr, *certFile, *keyFile, MuxPrimary)
	if err != nil {
		log.Fatalf("Error (Fatal) Main 70: %v", err.Error())
	}
}