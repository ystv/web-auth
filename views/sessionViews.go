package views

import (
	"encoding/base64"
	"log"
	"net/http"

	whirl "github.com/balacode/zr-whirl"
	"github.com/rmil/web-auth/db"
	"github.com/rmil/web-auth/sessions"
)

// LogoutFunc Implements the logout functionality.
// Will delete the session information from the cookie store
func LogoutFunc(w http.ResponseWriter, r *http.Request) {
	session, err := sessions.Store.Get(r, "session")
	if err == nil { // If there is no error, remove session
		if session.Values["loggedin"] != "false" {
			session.Values["loggedin"] = "false"
			session.Save(r, w)
		}
	}
	http.Redirect(w, r, "/login", http.StatusFound)
	// redirect to login irrespective of error or not
}

// LoginFunc implements the login functionality, will
// add a cookie to the cookie store for managing authentication
func LoginFunc(w http.ResponseWriter, r *http.Request) {
	session, err := sessions.Store.Get(r, "session")

	if err != nil {
		loginTemplate.Execute(w, nil)
		// In case of error during fetching
		// session info, execute login template
		return
	}
	if session.Values["loggedin"] == "true" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	switch r.Method {
	case "GET":
		loginTemplate.Execute(w, nil)
	case "POST":
		r.ParseForm()
		username := r.Form.Get("username")
		password := r.Form.Get("password")

		password = hashPassword(password)

		if (username != "" && password != "") && db.ValidUser(username, password) {
			session.Values["loggedin"] = "true"
			session.Values["userID"], _ = db.GetUserID(username)
			session.Values["username"] = username
			session.Save(r, w)
			log.Printf("user \"%s\" is authenticated", username)
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		log.Printf("Invalid user %s", username)
		loginTemplate.Execute(w, nil)
	}
}

// SignUpFunc will enable new users to sign up to our service
func SignUpFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		email := r.Form.Get("email")

		password = hashPassword(password)

		err := db.CreateUser(username, password, email)
		if err != nil {
			http.Error(w, "Unable to sign user up", http.StatusInternalServerError)
		} else {
			http.Redirect(w, r, "/login/", http.StatusFound)
		}
	} else if r.Method == "GET" {
		signupTemplate.Execute(w, nil)
	}
}

func hashPassword(password string) string {
	bytePass := []byte(password)
	iter := 1000
	var next []byte
	for i := 0; i < iter; i++ {
		next = append(next, bytePass...)
		next = whirl.HashOfBytes(bytePass, []byte(next))
	}
	return base64.StdEncoding.EncodeToString(next)
}
