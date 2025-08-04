package main

import (
	"log"
	"net/http"
)

func HandleLasersharks(w http.ResponseWriter, r *http.Request) {
	log.Print(r.RemoteAddr + " " + r.Method + " " + r.URL.String())
	w.Write([]byte("The Lasersharks are currently undergoing retraining. Please check back later.\n\n"))
}
