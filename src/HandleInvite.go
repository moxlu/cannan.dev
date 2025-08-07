package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/alexedwards/argon2id"
	_ "github.com/mattn/go-sqlite3"
)

func hashPassword(plainPwd string) (string, error) {
	params := &argon2id.Params{
		Memory:      32 * 1024, // 32 MB
		Iterations:  2,
		Parallelism: 1,
		SaltLength:  16,
		KeyLength:   32,
	}

	hashedPwd, err := argon2id.CreateHash(plainPwd, params)
	return hashedPwd, err
}

func (app *application) HandleInviteGet(w http.ResponseWriter, r *http.Request) {

	var invite_id int
	var invite_claimed sql.NullTime
	invite_token := r.PathValue("token")
	data := map[string]string{
		"Token": invite_token,
	}

	// fmt.Printf("Token value: '%s'\n", invite_token)
	// fmt.Printf("Token is of type %T\n", invite_token)
	// fmt.Printf("Token length: %d\n", len(invite_token))

	statement := "SELECT invite_id, invite_claimed_time FROM INVITES WHERE invite_token = ?;"
	err := app.db.QueryRow(statement, invite_token).Scan(&invite_id, &invite_claimed)

	if err != nil {
		log.Print("Error HandleInviteGet() 100 - Could not find token")
		w.Write([]byte("Token does not exist or is already claimed."))
		return
	}

	if invite_claimed.Valid { // Checks if Not Null
		log.Print("Alert HandleInviteGet() 200 - Token already claimed")
		w.Write([]byte("Token does not exist or is already claimed."))
		return
	}

	// TODO: check if token has expired

	response, err := template.ParseFiles("../dynamic/invite.html")
	if err != nil {
		log.Print("Error HandleInviteGet() 300 - Failed parsing invite.html")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = response.Execute(w, data)
	if err != nil {
		log.Print("Error HandleInviteGet() 400 - Failed serving invite.html")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (app *application) HandleInvitePost(w http.ResponseWriter, r *http.Request) {

	var invite_id int
	var invite_issued sql.NullTime
	var invite_claimed sql.NullTime
	invite_token := r.PathValue("token")

	statement := "SELECT invite_id, invite_issued, invite_claimed_time FROM INVITES WHERE invite_token = ?;"
	err := app.db.QueryRow(statement, invite_token).Scan(&invite_id, &invite_issued, &invite_claimed)
	if err != nil {
		log.Print("Error HandleInvitePost() 100 - Could not find token")
		w.Write([]byte("Token does not exist or is already claimed."))
		return
	}

	if invite_claimed.Valid { // Checks if Not Null
		log.Print("Alert HandleInvitePost() 200 - Token already claimed")
		w.Write([]byte("Token does not exist or is already claimed."))
		return
	}

	// TODO: check if token has expired

	err = r.ParseForm()
	if err != nil {
		log.Print("Error HandlePostInvite() 300 - Failed parsing /invite form")
		return
	}
	email := r.FormValue("email")
	name := r.FormValue("name")
	hash, err := hashPassword(r.FormValue("password"))
	if err != nil {
		log.Print("Error HandlePostInvite() 400 - Failed during hash of new user password")
		w.Write([]byte("Error :("))
		return
	}

	time_now := time.Now()
	statement = "INSERT INTO USERS (user_email, user_name, user_hash, user_invited, user_joined) VALUES (?, ?, ?, ?, ?);"
	_, err = app.db.Exec(statement, email, name, hash, invite_issued, time_now)
	if err != nil {
		log.Print("Error HandlePostInvite() 500 - Failed adding new user to db")
		w.Write([]byte("Error :("))
		return
	}

	statement = "UPDATE INVITES SET invite_claimed_by = ?, invite_claimed_time = ? WHERE invite_id = ?;"
	_, err = app.db.Exec(statement, email, time_now, invite_id)
	if err != nil {
		log.Print("Error HandlePostInvite() 600 - Failed marking invite claimed")
		w.Write([]byte("Error :("))
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther) // returns user to login
}
