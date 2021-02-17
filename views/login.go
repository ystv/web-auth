package views

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/schema"
	"github.com/patrickmn/go-cache"
	"github.com/ystv/web-auth/types"
	"github.com/ystv/web-auth/utils"
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

	session.Values["user"] = types.User{}
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
		context := getData(session)

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
		u := types.User{}
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
		if v.user.VerifyUser(r.Context(), &u) != nil {
			log.Printf("Failed login for \"%s\"", u.Username)
			err := session.Save(r, w)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			ctx := getData(session)
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
		err = v.user.SetUserLoggedIn(r.Context(), &u)
		if err != nil {
			err = fmt.Errorf("failed to set user logged in: %w", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		u.Authenticated = true
		// Bit of a cheat, just so we can have the last login displayed for internal
		u.LastLogin = prevLogin
		session.Values["user"] = u

		if r.Form.Get("remember") != "on" {
			session.Options.MaxAge = 0
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
			log.Printf("%+v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		http.Redirect(w, r, callback, http.StatusFound)
		return
	}
}

// SignUpFunc will enable new users to sign up to our service
func (v *Views) SignUpFunc(w http.ResponseWriter, r *http.Request) {
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
		err := v.tpl.ExecuteTemplate(w, "signup.gohtml", nil)
		if err != nil {
			log.Print(err)
		}
	}
}

// ForgotFunc handles sending a reset email
func (v *Views) ForgotFunc(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		err := v.tpl.ExecuteTemplate(w, "forgot.gohtml", nil)
		if err != nil {
			err = fmt.Errorf("failed to exec tmpl: %w", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	case "POST":
		r.ParseForm()
		u := types.User{Email: r.Form.Get("email")}

		if u.Email == "" {
			err := v.tpl.ExecuteTemplate(w, "forgot.gohtml", nil)
			if err != nil {
				err = fmt.Errorf("failed to exec tmpl: %w", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		// Get user and check if it exists
		if v.user.GetUser(r.Context(), &u) != nil {
			// User doesn't exist
			// TODO send no user message
			v.tpl.ExecuteTemplate(w, "forgot.gohtml", nil)
		}
		code := utils.RandomString(10)
		v.cache.Set(code, u.UserID, cache.DefaultExpiration)

		// Valid request, send email with reset code
		if v.mail.Enabled {
			err := v.mail.SendEmail(u.Email, "Forgotten Password", string(code))
			if err != nil {
				log.Printf("SendEmail failed: %s, ", err)
				log.Printf("reset email: %s, code: %s", u.Email, code)
			}
		} else {
			log.Printf("no mailer present")
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}

}

// ResetFunc handles resetting the password
func (v *Views) ResetFunc(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	if code == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	id, found := v.cache.Get(code)
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
		v.tpl.ExecuteTemplate(w, "reset.gohtml", ctx)
	case "POST":
		r.ParseForm()
		p := r.Form.Get("password")
		if p != r.Form.Get("confirmpassword") || p == "" {
			v.tpl.ExecuteTemplate(w, "reset.gohtml", ctx)
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
		err := v.user.UpdateUserPassword(r.Context(), &u)
		if err != nil {
			log.Printf("Failed to reset user: %+v", err)
		}
		v.cache.Delete(code)
		log.Printf("updated user: %s", u.Username)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
