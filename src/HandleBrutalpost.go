package main

import (
	"log"
	"net/http"
)

func HandleBrutalpost(w http.ResponseWriter, r *http.Request) {
	log.Print(r.RemoteAddr + " " + r.Method + " " + r.URL.String())
	w.Write([]byte("Brutalpost is still recovering. Please check back later.\n\n"))
}
