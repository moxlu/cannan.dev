package main

import (
	"html/template"
	"log"
	"net/http"
)

func Handle404(w http.ResponseWriter, r *http.Request) {
	log.Print("Alert Handle404: ", r.Method, " ", r.URL.String(), " (", r.RemoteAddr, ")")
	http.Error(w, "Page not found", http.StatusNotFound)
}

func (app *application) HandleIndex(w http.ResponseWriter, r *http.Request) {
	response, err := template.ParseFiles("../dynamic/index.html")
	if err != nil {
		log.Print("Error HandleIndex() 10 - Failed parsing index.html")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = response.Execute(w, nil)
	if err != nil {
		log.Print("Error HandleIndex() 20 - Failed serving index.html")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
