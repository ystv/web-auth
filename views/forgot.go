package views

import (
	"fmt"
	"github.com/ystv/web-auth/public/templates"
	"log"
	"net/http"
	"strconv"

	"github.com/ystv/web-auth/user"

	"github.com/patrickmn/go-cache"
	"github.com/ystv/web-auth/utils"
)

var notification = Notification{
	Title:   "Reset code sent",
	Type:    "",
	Message: "Cheers! If your account exists, you should receive a new email from \"YSTV Security\" with a link to reset your password shortly.",
}

// ForgotFunc handles sending a reset email
func (v *Views) ForgotFunc(w http.ResponseWriter, r *http.Request) {
	var err error
	switch r.Method {
	case "GET":
		fmt.Println("DEBUG - FORGOT GET")
		err = v.template.RenderNoNavsTemplate(w, nil, templates.ForgotTemplate)
		//err = v.tpl.ExecuteTemplate(w, "forgot.tmpl", nil)
		if err != nil {
			err = fmt.Errorf("failed to exec tmpl: %w", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	case "POST":
		fmt.Println("DEBUG - FORGOT POST")
		err = r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		u := user.User{Email: r.Form.Get("email")}

		if u.Email == "" {
			err = v.template.RenderNoNavsTemplate(w, nil, templates.ForgotTemplate)
			//err = v.tpl.ExecuteTemplate(w, "forgot.tmpl", nil)
			if err != nil {
				err = fmt.Errorf("failed to exec tmpl: %w", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		// Get user1 and check if it exists
		user1, err := v.user.GetUser(r.Context(), u)
		if err != nil {
			// User doesn't exist, we'll pretend they've got an email
			log.Printf("request for reset on unknown email \"%s\"", user1.Email)
			err = v.template.RenderNoNavsTemplate(w, notification, templates.NotificationTemplate)
			//err = v.tpl.ExecuteTemplate(w, "notification.tmpl", notification)
			if err != nil {
				err = fmt.Errorf("failed to exec template: %w", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		code := utils.RandomString(10)
		v.cache.Set(code, user1.UserID, cache.DefaultExpiration)

		// Valid request, send email with reset code
		if v.mail.Enabled {
			err = v.mail.SendEmail(user1.Email, "Forgotten Password", code)
			if err != nil {
				log.Printf("SendEmail failed: %s, ", err)
				log.Printf("request for password reset email \"%s\":reset code \"%s\"", user1.Email, code)
			}
			log.Printf("request for password reset email: \"%s\"", user1.Email)
		} else {
			log.Printf("no mailer present")
			log.Printf("reset email: %s, code: %s, reset link: https://auth.%s/reset?code=%s", user1.Email, code, v.conf.DomainName, code)
		}

		// User doesn't exist, we'll pretend they've got an email
		err = v.template.RenderNoNavsTemplate(w, notification, templates.NotificationTemplate)
		//err = v.tpl.ExecuteTemplate(w, "notification.tmpl", notification)
		if err != nil {
			err = fmt.Errorf("failed to exec template: %w", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// ResetFunc handles resetting the password
func (v *Views) ResetFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DEBUG - RESET")
	var err error

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
		fmt.Println("DEBUG - RESET GET")
		err = v.template.RenderNoNavsTemplate(w, ctx, templates.ResetTemplate)
		//err = v.tpl.ExecuteTemplate(w, "reset.tmpl", ctx)
		if err != nil {
			return
		}
	case "POST":
		fmt.Println("DEBUG - RESET POST")
		err = r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		p := r.Form.Get("password")
		if p != r.Form.Get("confirmpassword") || p == "" {
			err = v.template.RenderNoNavsTemplate(w, ctx, templates.ResetTemplate)
			//err = v.tpl.ExecuteTemplate(w, "reset.tmpl", ctx)
			if err != nil {
				return
			}
			return
		}
		// Good password
		formUserID := r.Form.Get("userid")
		// TODO error handling
		ctx.UserID, _ = strconv.Atoi(formUserID)
		if ctx.UserID != id.(int) {
			http.Error(w, "incorrect user id", http.StatusBadRequest)
		}

		// Update record

		u := user.User{UserID: id.(int), Password: p}
		user1, err := v.user.UpdateUserPassword(r.Context(), u)
		if err != nil {
			log.Printf("failed to reset user: %+v", err)
		}
		v.cache.Delete(code)
		log.Printf("updated user: %s", user1.Username)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
