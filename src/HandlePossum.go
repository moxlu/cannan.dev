package main

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"html"
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
		"flag7.txt": "flag7",
		"flag8.txt": "flag8",
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

func possumHash(s string) string {
	h := md5.Sum([]byte(s))
	first := fmt.Sprintf("%x", h[:])
	h2 := md5.Sum([]byte(first))
	second := fmt.Sprintf("%x", h2[:])
	h3 := md5.Sum([]byte(second))
	return fmt.Sprintf("%x", h3[:])
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
	hash := possumHash(password)

	// This is a deliberately vulnerable function to demonstrate SQL injection and should not be replicated in a real server.
	query := fmt.Sprintf("SELECT username, admin FROM users WHERE username = '%s' AND hash = '%s' LIMIT 1", username, hash)
	row := app.db.QueryRow(query)
	
	var dbUsername string
    var userIsAdmin bool
    err := row.Scan(&dbUsername, &userIsAdmin)
	
	if err != nil {
		if err == sql.ErrNoRows {
			// Authentication failed
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprintf(w, "Login failed\n\nQuery used:\n%s", query) // show query for CTF learning
			return
		}
		// Unexpected DB error
		http.Error(w, "Internal Server Error\n\n" + err.Error(), http.StatusInternalServerError)
		log.Printf("db error: %v", err)
		return
	}

	// User has "passed" authentication, check if admin
	if userIsAdmin {
		session, _ := app.store.Get(r, "possum-session")
		session.Values["authenticated"] = true
		session.Values["username"] = dbUsername
		session.Values["userIsAdmin"] = userIsAdmin
		session.Values["flag7"] = possumFlags["flag7"]
		session.Save(r, w)
		http.ServeFile(w, r, "../possum/admin.html")
	} else {
		// Successful login as normal user
		fmt.Fprintf(w, "Welcome, %s!\n\nYou appear to be a normal user.\n\nQuery used:\n%s\n\n%s", dbUsername, query, possumFlags["flag3"])
	}
}

func (app *application) HandlePostVerifyCookie(w http.ResponseWriter, r *http.Request) {
    // Grab session (do not ignore error completely)
    session, err := app.store.Get(r, "possum-session")
    if err != nil {
        // If the cookie is malformed or signature invalid, inform the user (for a CTF it's OK).
        http.Error(w, "Could not load possum-session cookie: "+html.EscapeString(err.Error()), http.StatusBadRequest)
        return
    }

    // Attempt to read the username value from the session.
    // Use a safe conversion to string to handle different underlying types.
    var username string
    if v, ok := session.Values["username"]; ok && v != nil {
        switch s := v.(type) {
        case string:
            username = s
        case []byte:
            username = string(s)
        default:
            username = fmt.Sprintf("%v", s)
        }
    } else {
        http.Error(w, "The server loaded your possum-session cookie but could not find a readable username. Did you encrypt it correctly?", http.StatusUnauthorized)
        return
    }

    // Check for the impersonation target
    if username == "erwin.rommel" {
        flag := possumFlags["flag8"] // ensure you set possumFlags["flag8"] in initPossum()
        if flag == "" {
            // fallback message if flag not present
            fmt.Fprintln(w, "Verified as erwin.rommel — but flag not found on server.")
            return
        }
        // Success — reveal the flag (CTF behaviour)
        fmt.Fprintf(w, "Cookie verified! Username: %s\n\nFlag: %s\n", html.EscapeString(username), html.EscapeString(flag))
        return
    }

    // Not the target username
    fmt.Fprintf(w, "Cookie verified. Username present in session: %s", html.EscapeString(username))
}
