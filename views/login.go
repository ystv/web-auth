package views

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/schema"
	"github.com/patrickmn/go-cache"
	"github.com/ystv/web-auth/types"
	"github.com/ystv/web-auth/utils"
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
	u.Email = u.Username
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// Authentication
	if uStore.VerifyUser(r.Context(), &u) != nil {
		log.Printf("Failed login for \"%s\"", u.Username)
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
	log.Printf("user \"%s\" is authenticated", u.Username)
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
			http.Error(w, "User doesn't ", http.StatusInternalServerError)
		} else {
			http.Redirect(w, r, "/", http.StatusFound)
		}
	} else if r.Method == "GET" {
		err := tpl.ExecuteTemplate(w, "signup.gohtml", nil)
		if err != nil {
			log.Print(err)
		}
	}
}

// ForgotFunc handles sending a reset email
func ForgotFunc(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		tpl.ExecuteTemplate(w, "forgot.gohtml", nil)
	case "POST":
		r.ParseForm()
		u := types.User{Email: r.Form.Get("email")}

		if u.Email == "" {
			tpl.ExecuteTemplate(w, "forgot.gohtml", nil)
		}
		// Get user and check if it exists
		if uStore.GetUser(r.Context(), &u) != nil {
			// User doesn't exist
			// TODO send no user message
			tpl.ExecuteTemplate(w, "forgot.gohtml", nil)
		}
		code := utils.RandomString(10)
		c.Set(code, u.UserID, cache.DefaultExpiration)

		// Valid request, send email with reset code
		log.Printf("reset email: %s, code: %s", u.Email, code)
		err := m.SendEmail(u.Email, "Forgotten Password", string(code))
		if err != nil {
			log.Printf("SendEmail failed: %s", err)
		}
		http.Redirect(w, r, "/", http.StatusFound)
	}

}

// ResetFunc handles resetting the password
func ResetFunc(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	if code == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	id, found := c.Get(code)
	if !found {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	ctx := struct {
		Code   string
		UserID int
	}{code, id.(int)}

	switch r.Method {
	case "GET":
		tpl.ExecuteTemplate(w, "reset.gohtml", ctx)
	case "POST":
		r.ParseForm()
		p := r.Form.Get("password")
		if p != r.Form.Get("confirmpassword") || p == "" {
			tpl.ExecuteTemplate(w, "reset.gohtml", ctx)
			return
		}
		// Good password
		formUserID := r.Form.Get("userid")
		// TODO error handling
		ctx.UserID, _ = strconv.Atoi(formUserID)
		if ctx.UserID != id.(int) {
			http.Error(w, "Incorrect user id", http.StatusBadRequest)
		}

		// Update record

		u := types.User{UserID: id.(int), Password: p}
		err := uStore.UpdateUserPassword(r.Context(), &u)
		if err != nil {
			log.Printf("Failed to reset user: %+v", err)
		}
		c.Delete(code)
		log.Printf("updated user: %s", u.Username)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
