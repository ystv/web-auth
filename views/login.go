package views

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/schema"
	"github.com/ystv/web-auth/user"
)

var decoder = schema.NewDecoder()

// LogoutFunc Implements the logout functionality.
// Will delete the session information from the cookie store
func (v *Views) LogoutFunc(w http.ResponseWriter, r *http.Request) {
	session, err := v.cookie.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session.Values["user"] = user.User{}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Expires: time.Now(),
		Domain:  v.conf.DomainName,
		Path:    "/",
	})
	// TODO Don't call env in this function have an initialiser
	// then fetch from that store?
	endpoint := v.conf.LogoutEndpoint
	if endpoint == "" {
		endpoint = "/"
	}
	http.Redirect(w, r, endpoint, http.StatusFound)
}

// LoginFunc implements the login functionality, will
// add a cookie to the cookie store for managing authentication
func (v *Views) LoginFunc(w http.ResponseWriter, r *http.Request) {
	session, _ := v.cookie.Get(r, "session")
	// We're ignoring the error here since sometimes the cookies keys change and then we
	// can overwrite it instead

	switch r.Method {
	case "GET":
		// Data for our HTML template
		context := v.getData(session)

		// Check if there is a callback request
		callbackURL, err := url.Parse(r.URL.Query().Get("callback"))
		if err == nil && strings.HasSuffix(callbackURL.Host, v.conf.DomainName) && callbackURL.String() != "" {
			context.Callback = callbackURL.String()
		}
		// Check if authenticated
		if context.User.Authenticated {
			http.Redirect(w, r, context.Callback, http.StatusFound)
			return
		}
		err = v.tpl.ExecuteTemplate(w, "login", context)
		if err != nil {
			log.Printf("login failed to exec tmpl: %+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	case "POST":
		// Parsing form to struct
		r.ParseForm()
		u := user.User{}
		decoder.Decode(&u, r.PostForm)
		// Since we let users enter either an email or username, it's easier
		// to just let it both for the query
		u.Email = u.Username

		callback := "/internal"
		callbackURL, err := url.Parse(r.URL.Query().Get("callback"))
		if err == nil && strings.HasSuffix(callbackURL.Host, v.conf.DomainName) && callbackURL.String() != "" {
			callback = callbackURL.String()
		}
		// Authentication
		u, err = v.user.VerifyUser(r.Context(), u)
		if err != nil {
			log.Printf("failed login for \"%s\"", u.Username)
			err := session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			ctx := v.getData(session)
			ctx.Callback = callback
			ctx.Message = "Invalid username or password"
			ctx.MsgType = "is-danger"
			err = v.tpl.ExecuteTemplate(w, "login", ctx)
			if err != nil {
				log.Printf("login failed to exec tmpl: %+v", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		prevLogin := u.LastLogin
		// Update last logged in
		err = v.user.SetUserLoggedIn(r.Context(), u)
		if err != nil {
			err = fmt.Errorf("failed to set user logged in: %w", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		u.Authenticated = true
		// Bit of a cheat, just so we can have the last login displayed for internal
		u.LastLogin = prevLogin
		session.Values["user"] = u

		if r.Form.Get("remember") != "on" {
			session.Options.MaxAge = 86400 * 31
		}

		err = session.Save(r, w)
		if err != nil {
			err = fmt.Errorf("failed to save user session: %w", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("user \"%s\" is authenticated", u.Username)
		w, err = v.getJWTCookie(w, r)
		if err != nil {
			err = fmt.Errorf("login: failed to set cookie: %w", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		http.Redirect(w, r, callback, http.StatusFound)
		return
	}
}
