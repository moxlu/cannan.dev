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
	MuxPrimary.HandleFunc("GET /{$}", HandleIndex)
	MuxPrimary.HandleFunc("POST /login", HandleLogin)
	MuxPrimary.HandleFunc("GET /invite/{token}", HandleGetInvite)
	MuxPrimary.HandleFunc("POST /invite", HandlePostInvite)
	MuxPrimary.HandleFunc("GET POST /brutalpost", HandleBrutalpost)
	MuxPrimary.HandleFunc("GET /lasersharks", HandleLasersharks)

	// Fileserver for specified static files only
	fileServer := http.FileServer(http.Dir("../static/"))
	MuxPrimary.Handle("GET /favicon.ico", fileServer)
	MuxPrimary.Handle("GET /main.css", fileServer)
	MuxPrimary.Handle("GET /robots.txt", fileServer)

	// Everything else should go to 404
	MuxPrimary.HandleFunc("/", Handle404)

	log.Print("Starting server on :4000")
	err = http.ListenAndServe(":4000", MuxPrimary)
	log.Print(err.Error())
}
