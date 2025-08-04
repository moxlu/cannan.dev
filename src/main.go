package main

import (
	"html/template"
	"log"
	"net/http"
)

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
	MuxPrimary := http.NewServeMux()

	// Main Handlers
	MuxPrimary.HandleFunc("/", HandleIndex)

	// Challenge Specific Handlers
	MuxPrimary.HandleFunc("/brutalpost", HandleBrutalpost)
	MuxPrimary.HandleFunc("/lasersharks", HandleLasersharks)

	// Fileserver for static files only
	fileServer := http.FileServer(http.Dir("../static/"))
	MuxPrimary.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	MuxPrimary.Handle("GET /robots.txt", fileServer)

	log.Print("Starting server on :8080")
	err := http.ListenAndServe(":8080", MuxPrimary)
	log.Fatal(err)
}
