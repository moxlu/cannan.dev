package main

import (
	"html/template"
	"log"
	"net/http"

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

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var match bool
	var response *template.Template
	var err error
	var storedHash string

	log.Print(r.RemoteAddr + " " + r.Method + " " + r.URL.String())

	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")
	db.QueryRow("SELECT user_hash FROM USERS WHERE user_email = ?", email).Scan(&storedHash)
	match, err = argon2id.ComparePasswordAndHash(password, storedHash)
	if err != nil {
		log.Print(err.Error())
		response, err = template.ParseFiles("../dynamic/login fail.html")
		if err != nil {
			log.Print(err.Error())
		}

	}

	if match {
		log.Print("Successful login: ", email)
		response, err = template.ParseFiles("../dynamic/login success.html")
		if err != nil {
			log.Print(err.Error())
		}
	}

	response.Execute(w, nil)
}
