package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

// Global variables
var db *sql.DB

func HandleIndex(w http.ResponseWriter, r *http.Request) {
	log.Print(r.RemoteAddr + " " + r.Method + " " + r.URL.String())

	response, err := template.ParseFiles("../dynamic/index.html")
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = response.Execute(w, nil)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func main() {
	// Start Database
	var err error
	db, err = sql.Open("sqlite3", "../run/cannan.db")
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer db.Close()

	_, err = db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Print(err.Error())
		return
	} else {
		log.Print("Database loaded OK")
	}

	// Start mux and handlers
	MuxPrimary := http.NewServeMux()
	MuxPrimary.HandleFunc("/", HandleIndex)
	MuxPrimary.HandleFunc("POST /login", HandleLogin)
	MuxPrimary.HandleFunc("/brutalpost", HandleBrutalpost)
	MuxPrimary.HandleFunc("/lasersharks", HandleLasersharks)

	// Fileserver for static files only
	fileServer := http.FileServer(http.Dir("../static/"))
	MuxPrimary.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	MuxPrimary.Handle("GET /robots.txt", fileServer)

	log.Print("Starting server on :4000")
	err = http.ListenAndServe(":4000", MuxPrimary)
	log.Print(err.Error())
}
