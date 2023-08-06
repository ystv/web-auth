package views

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/ystv/web-auth/infrastructure/mail"
	"github.com/ystv/web-auth/public/templates"
	"github.com/ystv/web-auth/user"
	"log"
	"net/http"
)

type (
	// Notification template for messages
	Notification struct {
		Title   string
		Type    string
		Message string
	}
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
		err = v.template.RenderNoNavsTemplate(w, nil, templates.ForgotTemplate)
		if err != nil {
			err = fmt.Errorf("failed to exec tmpl: %w", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	case "POST":
		err = r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		u := user.User{Email: r.Form.Get("email")}

		if u.Email == "" {
			err = v.template.RenderNoNavsTemplate(w, nil, templates.ForgotTemplate)
			if err != nil {
				err = fmt.Errorf("failed to exec tmpl: %w", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		// Get user1 and check if it exists
		user1, err := v.user.GetUser(r.Context(), u)
		if err != nil {
			// UserStripped doesn't exist, we'll pretend they've got an email
			log.Printf("request for reset on unknown email \"%s\"", user1.Email)
			err = v.template.RenderNoNavsTemplate(w, notification, templates.NotificationTemplate)
			if err != nil {
				err = fmt.Errorf("failed to exec template: %w", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
		url := uuid.NewString()
		v.cache.Set(url, user1.UserID, cache.DefaultExpiration)

		// Valid request, send email with reset code
		if v.mailer.Enabled {
			v.mailer, err = mail.NewMailer(mail.Config{
				Host:       v.conf.Mail.Host,
				Port:       v.conf.Mail.Port,
				Username:   v.conf.Mail.Username,
				Password:   v.conf.Mail.Password,
				DomainName: v.conf.DomainName,
			})
			if err != nil {
				log.Printf("mailer failed: %+v", err)
			}

			//forgot := forgotPasswords.ForgotPassword{
			//	URL:    uuid.NewString(),
			//	UserID: user1.UserID,
			//}

			//err = v.forgot.InsertURL(r.Context(), forgot)
			//if err != nil {
			//	err = v.errorHandle(w, err)
			//	if err != nil {
			//		return
			//	}
			//	return
			//}

			file := mail.Mail{
				Subject: "YSTV Security - Reset Password",
				Tpl:     v.template.RenderEmail(templates.ForgotEmailTemplate),
				To:      user1.Email,
				From:    "YSTV Security <no-reply@ystv.co.uk>",
				TplData: struct {
					Email string
					URL   string
				}{
					Email: user1.Email,
					URL:   "https://" + v.conf.DomainName + "/forgot/" + url,
				},
			}

			err = v.mailer.SendMail(file)
			if err != nil {
				err = v.errorHandle(w, err)
				if err != nil {
					fmt.Println(err)
				}
			}

			//err = v.mailer.SendResetEmail(user1.Email, "Forgotten Password", code)
			//if err != nil {
			//	log.Printf("SendResetEmail failed: %s, ", err)
			//	log.Printf("request for password reset email \"%s\":reset code \"%s\"", user1.Email, code)
			//}
			log.Printf("request for password reset email: \"%s\"", user1.Email)
		} else {
			log.Printf("no mailer present")
			log.Printf("reset email: %s, code: %s, reset link: https://%s/reset?code=%s", user1.Email, url, v.conf.DomainName, url)
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

func (v *Views) ResetURLFunc(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	url1 := vars["url"]

	userID, found := v.cache.Get(url1)
	if !found {
		err := v.errorHandle(w, fmt.Errorf("failed to get url"))
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	user1, err := v.user.GetUser(r.Context(), user.User{UserID: userID.(int)})
	if err != nil {
		v.cache.Delete(url1)
		err = v.errorHandle(w, fmt.Errorf("url is invalid because this user doesn't exist"))
		if err != nil {
			fmt.Println(err)
		}
		return
	}

	switch r.Method {
	case "GET":
		err = v.template.RenderNoNavsTemplate(w, nil, templates.ResetTemplate)
		if err != nil {
			err = v.errorHandle(w, err)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
		return
	case "POST":
		err = r.ParseForm()
		if err != nil {
			err = v.errorHandle(w, err)
			if err != nil {
				fmt.Println(err)
			}
		}

		password := r.FormValue("password")
		if password != r.FormValue("confirmpassword") {
			err = v.template.RenderNoNavsTemplate(w, nil, templates.ResetTemplate)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		//fmt.Println("DEBUG - RESET POST")
		err = r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		p := r.Form.Get("password")
		if p != r.Form.Get("confirmpassword") || p == "" {
			err = v.template.RenderNoNavsTemplate(w, nil, templates.ResetTemplate)
			//err = v.tpl.ExecuteTemplate(w, "reset.tmpl", ctx)
			if err != nil {
				return
			}
			return
		}
		user1.Password = password

		user2, err := v.user.UpdateUserPassword(r.Context(), user1)
		if err != nil {
			log.Printf("failed to reset user: %+v", err)
		}
		v.cache.Delete(url1)
		log.Printf("updated user: %s", user2.Username)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func (v *Views) errorHandle(w http.ResponseWriter, err error) error {
	data := struct {
		Error string
	}{
		Error: err.Error(),
	}
	fmt.Println(data.Error)
	err = v.template.RenderNoNavsTemplate(w, data, templates.ErrorTemplate)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
