package views

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"github.com/ystv/web-auth/infrastructure/mail"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
	"gopkg.in/guregu/null.v4"
	"log"
	"net/http"
	"strconv"
)

func (v *Views) ResetURLFunc(c echo.Context) error {
	url := c.Param("url")

	userID, found := v.cache.Get(url)
	if !found {
		return v.errorHandle(c, fmt.Errorf("failed to get url"))
	}

	user1, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID.(int)})
	if err != nil {
		v.cache.Delete(url)
		return v.errorHandle(c, fmt.Errorf("url is invalid because this user doesn't exist"))
	}

	switch c.Request().Method {
	case "GET":
		return v.template.RenderNoNavsTemplate(c.Response(), nil, templates.ResetTemplate)
	case "POST":
		err = c.Request().ParseForm()
		if err != nil {
			return v.errorHandle(c, err)
		}

		password := c.Request().FormValue("password")
		if password != c.Request().FormValue("confirmpassword") {
			data := struct{ Error string }{Error: "passwords do not match"}
			return v.template.RenderNoNavsTemplate(c.Response(), data, templates.ResetTemplate)
		}

		errString := minRequirementsMet(password)
		if len(errString) > 0 {
			data := struct{ Error string }{Error: errString}
			return v.template.RenderNoNavsTemplate(c.Response(), data, templates.ResetTemplate)
		}

		user1.Password = null.StringFrom(password)

		user2, err := v.user.EditUserPassword(c.Request().Context(), user1)
		if err != nil {
			log.Printf("failed to reset user: %+v", err)
		}
		v.cache.Delete(url)
		log.Printf("updated user: %s", user2.Username)
		return c.Redirect(http.StatusFound, "/")
	}
	return nil
}

func (v *Views) ResetUserPasswordFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	userID, err := strconv.Atoi(c.Param("userid"))
	if err != nil {
		return v.errorHandle(c, err)
	}

	user1, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
	if err != nil {
		log.Println(err)
		return v.errorHandle(c, err)
	}

	user1.ResetPw = true

	_, err = v.user.EditUser(c.Request().Context(), user1, c1.User.UserID)
	if err != nil {
		return v.errorHandle(c, err)
	}

	url := uuid.NewString()
	v.cache.Set(url, user1.UserID, cache.DefaultExpiration)

	var message struct {
		Message string `json:"message"`
		Error   error  `json:"error"`
	}

	// Valid request, send email with reset code
	if v.Mailer.Enabled {
		v.Mailer, err = mail.NewMailer(mail.Config{
			Host:       v.conf.Mail.Host,
			Port:       v.conf.Mail.Port,
			Username:   v.conf.Mail.Username,
			Password:   v.conf.Mail.Password,
			DomainName: v.conf.DomainName,
		})
		if err != nil {
			log.Printf("Mailer failed: %+v", err)
		}

		file := mail.Mail{
			Subject: "YSTV Security - Reset Password",
			Tpl:     v.template.RenderEmail(templates.ResetPasswordEmailTemplate),
			To:      user1.Email,
			From:    "YSTV Security <no-reply@ystv.co.uk>",
			TplData: struct {
				Email string
				URL   string
			}{
				Email: user1.Email,
				URL:   fmt.Sprintf("https://%s/reset/%s", v.conf.DomainName, url),
			},
		}

		err = v.Mailer.SendMail(file)
		if err != nil {
			return v.errorHandle(c, err)
		}

		log.Printf("request for password reset email: \"%s\"", user1.Email)
		message.Message = fmt.Sprintf("Reset email sent to: \"%s\"", user1.Email)
	} else {
		message.Message = fmt.Sprintf("No mailer present\nPlease forward the link to this email: %s, reset link: https://%s/reset/%s", user1.Email, v.conf.DomainName, url)
		message.Error = fmt.Errorf("no mailer present")
		log.Printf("no Mailer present")
		log.Printf("reset email: %s, url: %s, reset link: https://%s/reset/%s", user1.Email, url, v.conf.DomainName, url)
	}
	log.Printf("reset for %d (%s) requested by %d (%s)", user1.UserID, user1.Firstname+" "+user1.Lastname, c1.User.UserID, c1.User.Firstname+" "+c1.User.Lastname)
	var status int
	return c.JSON(status, message)
}
