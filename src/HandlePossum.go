package main

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var possumFlags = map[string]string{}

func initPossum() {
	files := map[string]string{
		"flag2.txt": "flag2",
		"flag3.txt": "flag3",
		"flag6.txt": "flag6",
	}

	for filename, key := range files {
		path := filepath.Join(".", "../possum/", filename)
		b, err := os.ReadFile(path)
		if err != nil {
			log.Printf("flag file not found at %s: %v", path, err)
			continue
		}
		possumFlags[key] = string(b)
	}
}

func doubleMD5(s string) string {
	h := md5.Sum([]byte(s))
	first := fmt.Sprintf("%x", h[:])
	h2 := md5.Sum([]byte(first))
	return fmt.Sprintf("%x", h2[:])
}

func (app *application) HandleGetPossumIndex(w http.ResponseWriter, r *http.Request) {
	path := "../possum/index.html"

	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	} else if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("stat %s: %v", path, err)
		return
	}
	http.ServeFile(w, r, path)
}

func (app *application) HandlePostPossumEnquiry(w http.ResponseWriter, r *http.Request) {
	log.Print(r.RemoteAddr + " " + r.Method + " " + r.URL.String())
	
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad request, which in this case, I suppose is a good request? However, it's not the kind of bad request we are looking for! Please try a different method. Keep it simple - how many fields is the server expecting?", http.StatusBadRequest)
		return
	}

	fieldCount := len(r.PostForm)
	if fieldCount <= 3 {
		fmt.Fprintf(w, "Enquiry submitted successfully! We will get back to you soon.")
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	fmt.Fprintf(w, "Internal Server Error. Something went wrong with your submission. %s", possumFlags["flag2"])
	}

func (app *application) HandlePostPossumLogin(w http.ResponseWriter, r *http.Request) {
	log.Print(r.RemoteAddr + " " + r.Method + " " + r.URL.String())

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	hash := doubleMD5(password)

	// This is a vulnerable function to demonstrate SQL injection and should not be replicated in a real server.
	query := fmt.Sprintf("SELECT username FROM users WHERE username = '%s' AND hash = '%s' LIMIT 1", username, hash)
	row := app.db.QueryRow(query)

	var user string
	err := row.Scan(&user)
	if err != nil {
		if err == sql.ErrNoRows {
			// Authentication failed
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Login failed\n\nQuery used:\n%s", query) // show query for teaching
			return
		}
		// Unexpected DB error
		http.Error(w, "Internal Server Error\n\n" + err.Error(), http.StatusInternalServerError)
		log.Printf("db error: %v", err)
		return
	}

	// Successful login
	fmt.Fprintf(w, "Welcome, %s!\n\nQuery used:\n%s\n\n%s", user, query, possumFlags["flag3"])
}