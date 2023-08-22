package views

import (
	"fmt"
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
	switch c.Request().Method {
	case "GET":
		return v.template.RenderTemplate(c.Response().Writer, nil, templates.ForgotTemplate, templates.NoNavType)
	case "POST":
		u := user.User{Email: c.FormValue("email")}

		if u.Email == "" {
			return v.template.RenderTemplate(c.Response(), nil, templates.ForgotTemplate, templates.NoNavType)
		}
		// Get user and check if it exists
		userFromDB, err := v.user.GetUser(c.Request().Context(), u)
		if err != nil {
			// User doesn't exist, we'll pretend they've got an email
			log.Printf("request for reset on unknown email \"%s\"", userFromDB.Email)
			return v.template.RenderTemplate(c.Response(), notification, templates.NotificationTemplate, templates.NoNavType)
		}
		url := uuid.NewString()
		v.cache.Set(url, userFromDB.UserID, cache.DefaultExpiration)

		mailer := mail.NewMailer(mail.Config{
			Host:       v.conf.Mail.Host,
			Port:       v.conf.Mail.Port,
			Username:   v.conf.Mail.Username,
			Password:   v.conf.Mail.Password,
			DomainName: v.conf.Mail.DomainName,
		})

		// Valid request, send email with reset code
		if mailer != nil {
			emailTemplate, err := v.template.RenderEmail(templates.ForgotEmailTemplate)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to render email for forgot: %w", err))
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
				return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to send email for forgot: %w", err))
			}
			_ = mailer.Close()

			log.Printf("request for password reset email: \"%s\"", userFromDB.Email)
		} else {
			log.Printf("no Mailer present")
			log.Printf("reset email: %s, code: %s, reset link: https://%s/reset/%s", userFromDB.Email, url, v.conf.DomainName, url)
		}

		// User doesn't exist, we'll pretend they've got an email
		return v.template.RenderTemplate(c.Response().Writer, notification, templates.NotificationTemplate, templates.NoNavType)
	}
	return nil
}
