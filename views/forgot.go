package views

import (
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"

	"github.com/ystv/web-auth/infrastructure/mail"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
)

type (
	// Notification template for messages
	Notification struct {
		Title   string
		MsgType string
		Message string
	}
)

var notification = Notification{
	Title:   "Reset code sent",
	MsgType: "",
	Message: `Cheers! If your account exists, you should receive a new email from "YSTV Security" with a link to reset
your password shortly.`,
}

// ForgotFunc handles sending a reset email
func (v *Views) ForgotFunc(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodGet:
		return v.template.RenderTemplate(c.Response().Writer, nil, templates.ForgotTemplate, templates.NoNavType)
	case http.MethodPost:
		u := user.User{Email: c.FormValue("email")}

		if u.Email == "" {
			return v.template.RenderTemplate(c.Response(), nil, templates.ForgotTemplate, templates.NoNavType)
		}
		// Get user and check if it exists
		userFromDB, err := v.user.GetUser(c.Request().Context(), u)
		if err != nil {
			// User doesn't exist, we'll pretend they've got an email
			log.Printf("request for reset on unknown email \"%s\"", u.Email)

			return v.template.RenderTemplate(c.Response(), notification, templates.NotificationTemplate,
				templates.NoNavType)
		}

		url := uuid.NewString()
		v.cache.Set(url, userFromDB.UserID, cache.DefaultExpiration)

		mailer := v.mailer.ConnectMailer()

		// Valid request, send email with reset code
		if mailer != nil {
			var emailTemplate *template.Template

			emailTemplate, err = v.template.GetEmailTemplate(templates.ForgotEmailTemplate)
			if err != nil {
				return fmt.Errorf("failed to render email for forgot: %w", err)
			}

			file := mail.Mail{
				Subject: "YSTV Security - Reset Password",
				Tpl:     emailTemplate,
				To:      userFromDB.Email,
				From:    "YSTV Security <no-reply@ystv.co.uk>",
				TplData: struct {
					Email string
					URL   string
				}{
					Email: userFromDB.Email,
					URL:   "https://" + v.conf.DomainName + "/reset/" + url,
				},
			}

			err = mailer.SendMail(file)
			if err != nil {
				log.Printf("failed to send mail for forgot: %+v", err)

				return fmt.Errorf("failed to send email for forgot: %w", err)
			}

			_ = mailer.Close()

			log.Printf("request for password reset email: \"%s\"", userFromDB.Email)
		} else {
			log.Printf("no Mailer present")
			log.Printf("forgot password requested for email: %s", userFromDB.Email)
		}

		// User doesn't exist, we'll pretend they've got an email
		return v.template.RenderTemplate(c.Response().Writer, notification, templates.NotificationTemplate,
			templates.NoNavType)
	}

	return v.invalidMethodUsed(c) // maybe nil
}
