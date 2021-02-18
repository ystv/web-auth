package views

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/patrickmn/go-cache"
	"github.com/ystv/web-auth/types"
	"github.com/ystv/web-auth/utils"
)

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
