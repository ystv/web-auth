package views

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"github.com/ystv/web-auth/infrastructure/mail"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
	"gopkg.in/guregu/null.v4"
)

func (v *Views) ResetURLFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	url := c.Param("url")

	userID, found := v.cache.Get(url)
	if !found {
		return fmt.Errorf("failed to get url for reset")
	}

	originalUser, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID.(int)})
	if err != nil {
		v.cache.Delete(url)
		return fmt.Errorf("url is invalid, failed to get user : %w", err)
	}

	switch c.Request().Method {
	case "GET":
		return v.template.RenderTemplate(c.Response(), c1, templates.ResetTemplate, templates.NoNavType)
	case "POST":
		password := c.FormValue("password")
		if password != c.FormValue("confirmpassword") {
			return v.template.RenderTemplate(c.Response(), nil, templates.ResetTemplate, templates.NoNavType)
		}

		originalUser.Password = null.StringFrom(password)

		updatedUser, err := v.user.UpdateUserPassword(c.Request().Context(), originalUser)
		if err != nil {
			log.Printf("failed to reset user: %+v", err)
		}
		v.cache.Delete(url)
		log.Printf("updated user: %s", updatedUser.Username)
		return c.Redirect(http.StatusFound, "/")
	}
	return nil
}

func (v *Views) ResetUserPasswordFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	userID, err := strconv.Atoi(c.Param("userid"))
	if err != nil {
		return fmt.Errorf("failed to parse userid for reset: %w", err)
	}

	userFromDB, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
	if err != nil {
		return fmt.Errorf("failed to get user for reset: %w", err)
	}

	userFromDB.ResetPw = true

	_, err = v.user.UpdateUser(c.Request().Context(), userFromDB, c1.User.UserID)
	if err != nil {
		return fmt.Errorf("failed to update user for reset: %w", err)
	}

	url := uuid.NewString()
	v.cache.Set(url, userFromDB.UserID, cache.DefaultExpiration)

	var status int

	var message struct {
		Message string `json:"message"`
		Error   error  `json:"error"`
	}

	mailer := v.mailer.ConnectMailer()

	// Valid request, send email with reset code
	if mailer != nil {
		emailTemplate, err := v.template.GetEmailTemplate(templates.ResetEmailTemplate)
		if err != nil {
			return fmt.Errorf("failed to render email for reset: %w", err)
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
				URL:   fmt.Sprintf("https://%s/reset/%s", v.conf.DomainName, url),
			},
		}

		err = mailer.SendMail(file)
		if err != nil {
			message.Message = fmt.Sprintf("Please forward the link to this email: %s, reset link: https://%s/reset/%s", userFromDB.Email, v.conf.DomainName, url)
			message.Error = fmt.Errorf("failed to send mail: %w", err)
			log.Printf("failed to send mail: %+v", err)
			log.Printf("password reset requested for email: %s by user: %d", userFromDB.Email, c1.User.UserID)
			return c.JSON(http.StatusInternalServerError, message)
		}
		_ = mailer.Close()

		log.Printf("password reset requested for email: %s by user: %d", userFromDB.Email, c1.User.UserID)
		message.Message = fmt.Sprintf("Reset email sent to: \"%s\"", userFromDB.Email)
	} else {
		message.Message = fmt.Sprintf("No mailer present\nPlease forward the link to this email: %s, reset link: https://%s/reset/%s", userFromDB.Email, v.conf.DomainName, url)
		message.Error = fmt.Errorf("no mailer present")
		log.Printf("no Mailer present")
		log.Printf("password reset requested for email: %s by user: %d", userFromDB.Email, c1.User.UserID)
	}
	log.Printf("reset for %d (%s) requested by %d (%s)", userFromDB.UserID, userFromDB.Firstname+" "+userFromDB.Lastname, c1.User.UserID, c1.User.Firstname+" "+c1.User.Lastname)
	return c.JSON(status, message)
}
