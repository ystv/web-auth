package views

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"github.com/ystv/web-auth/infrastructure/mail"
	"github.com/ystv/web-auth/templates"
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
func (v *Views) ForgotFunc(c echo.Context) error {
	var err error
	switch c.Request().Method {
	case "GET":
		return v.template.RenderNoNavsTemplate(c.Response().Writer, nil, templates.ForgotTemplate)
	case "POST":
		err = c.Request().ParseForm()
		if err != nil {
			http.Error(c.Response(), err.Error(), http.StatusBadRequest)
		}
		u := user.User{Email: c.Request().Form.Get("email")}

		if u.Email == "" {
			return v.template.RenderNoNavsTemplate(c.Response(), nil, templates.ForgotTemplate)
		}
		// Get user and check if it exists
		user1, err := v.user.GetUser(c.Request().Context(), u)
		if err != nil {
			// User doesn't exist, we'll pretend they've got an email
			log.Printf("request for reset on unknown email \"%s\"", user1.Email)
			return v.template.RenderNoNavsTemplate(c.Response(), notification, templates.NotificationTemplate)
		}
		url := uuid.NewString()
		v.cache.Set(url, user1.UserID, cache.DefaultExpiration)

		// Valid request, send email with reset code
		if v.Mailer.Enabled {
			v.Mailer = mail.NewMailer(mail.Config{
				Host:       v.conf.Mail.Host,
				Port:       v.conf.Mail.Port,
				Username:   v.conf.Mail.Username,
				Password:   v.conf.Mail.Password,
				DomainName: v.conf.DomainName,
			})

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

			err = v.Mailer.SendMail(file)
			if err != nil {
				return v.errorHandle(c, err)
			}
			log.Printf("request for password reset email: \"%s\"", user1.Email)
		} else {
			log.Printf("no Mailer present")
			log.Printf("reset email: %s, code: %s, reset link: https://%s/reset?code=%s", user1.Email, url, v.conf.DomainName, url)
		}

		// User doesn't exist, we'll pretend they've got an email
		return v.template.RenderNoNavsTemplate(c.Response().Writer, notification, templates.NotificationTemplate)
	}
	return nil
}
