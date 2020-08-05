package views

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/ystv/web-auth/types"
)

var decoder = schema.NewDecoder()

// LogoutFunc Implements the logout functionality.
// Will delete the session information from the cookie store
func LogoutFunc(w http.ResponseWriter, r *http.Request) {
	session, err := cStore.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["user"] = types.User{}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

// LoginFunc implements the login functionality, will
// add a cookie to the cookie store for managing authentication
func LoginFunc(w http.ResponseWriter, r *http.Request) {
	session, err := cStore.Get(r, "session")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Parsing form to struct
	r.ParseForm()
	u := types.User{}
	err = decoder.Decode(&u, r.PostForm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Authentication
	if s.db.VerifyUser(context.Background(), &u) != nil {
		log.Printf("Invalid user %d", u.UserID)
		err = session.Save(r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ctx := getData(session)
		ctx.Message = "Invalid username or password"
		tpl.ExecuteTemplate(w, "index.gohtml", ctx)
		return
	}
	u.Authenticated = true
	session.Values["user"] = u
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("user \"%d\" is authenticated", u.UserID)
	w = getJWTCookie(w, r)
	http.Redirect(w, r, "/internal", http.StatusFound)
	return
}

// SignUpFunc will enable new users to sign up to our service
func SignUpFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		// Parsing form to struct
		r.ParseForm()
		u := types.User{}
		err := decoder.Decode(&u, r.PostForm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, "Unable to sign user up", http.StatusInternalServerError)
		} else {
			http.Redirect(w, r, "/login/", http.StatusFound)
		}
	} else if r.Method == "GET" {
		err := tpl.ExecuteTemplate(w, "signup.gohtml", "")
		if err != nil {
			log.Print(err)
		}
	}
}
