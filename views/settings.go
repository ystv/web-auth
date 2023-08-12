package views

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
	"log"
	"net/http"
	"strings"
	"time"
)

type (
	// SettingsTemplate is for the settings front end
	SettingsTemplate struct {
		User       user.User
		UserID     int
		LastLogin  string
		ActivePage string
		Gravatar   string
	}
)

// SettingsFunc handles a request to the internal template
func (v *Views) SettingsFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	if c.Request().Method == http.MethodPost {
		err := c.Request().ParseForm()
		if err != nil {
			return v.errorHandle(c, fmt.Errorf("failed to parse form for settings: %+v", err))
		}

		firstName := c.Request().FormValue("firstname")
		nickname := c.Request().FormValue("nickname")
		lastName := c.Request().FormValue("lastname")
		// avatar type can't be changed yet but the infrastructure is in
		avatar := c.Request().FormValue("avatar")
		_ = avatar

		if firstName != c1.User.Firstname && len(firstName) > 0 {
			c1.User.Firstname = firstName
		}
		if nickname != c1.User.Nickname && len(nickname) > 0 {
			c1.User.Nickname = nickname
		}
		if lastName != c1.User.Lastname && len(lastName) > 0 {
			c1.User.Lastname = lastName
		}

		_, err = v.user.EditUser(c.Request().Context(), c1.User, c1.User.UserID)
		if err != nil {
			log.Printf("failed to edit user for settings: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to edit user for settings: %+v", err))
			}
		}

		session.Values["user"] = c1.User

		err = session.Save(c.Request(), c.Response())
		if err != nil {
			err = fmt.Errorf("failed to save user session in settings: %w", err)
			return v.errorHandle(c, err)
		}

		c.Request().Method = http.MethodGet
		return v.SettingsFunc(c)
	}

	lastLogin := time.Now()
	if c1.User.LastLogin.Valid {
		lastLogin = c1.User.LastLogin.Time
	}

	var gravatar string

	if c1.User.UseGravatar {
		hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(c1.User.Email))))
		gravatar = fmt.Sprintf("https://www.gravatar.com/avatar/%s", hex.EncodeToString(hash[:]))
	}

	ctx := SettingsTemplate{
		User:       c1.User,
		UserID:     c1.User.UserID,
		LastLogin:  humanize.Time(lastLogin),
		ActivePage: "settings",
		Gravatar:   gravatar,
	}

	return v.template.RenderTemplate(c.Response(), ctx, templates.SettingsTemplate)
}
