package main

import (
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

func HandleGetInvite(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	IP := r.RemoteAddr
	log.Print(IP, " GET /invite Token:", token)
	Handle404(w, r)
}

func HandlePostInvite(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		log.Print("Error HandlePostInvite() 100 - Failed parsing /invite form")
	}

	token := r.FormValue("single_use_token")
	IP := r.RemoteAddr

	log.Print(IP, " POST /invite", token)
	Handle404(w, r)
}
