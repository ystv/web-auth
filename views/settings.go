package views

import (
	// #nosec
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v4"

	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
)

type (
	// SettingsTemplate is for the settings front end
	SettingsTemplate struct {
		User      user.User
		LastLogin string
		Gravatar  string
		TemplateHelper
	}
)

// SettingsFunc handles a request to the internal template
func (v *Views) SettingsFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getSessionData(c)

	if c.Request().Method == http.MethodPost {
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

		err := v.user.EditUser(c.Request().Context(), c1.User, c1.User.UserID)
		if err != nil {
			return fmt.Errorf("failed to edit user for settings: %w", err)
		}

		session.Values["user"] = c1.User

		err = session.Save(c.Request(), c.Response())
		if err != nil {
			return fmt.Errorf("failed to save user session in settings: %w", err)
		}

		return c.Redirect(http.StatusFound, "/internal/settings")
	}

	lastLogin := time.Now()
	if c1.User.LastLogin.Valid {
		lastLogin = c1.User.LastLogin.Time
	}

	var gravatar string

	if c1.User.UseGravatar {
		// #nosec
		hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(c1.User.Email))))
		gravatar = fmt.Sprintf("https://www.gravatar.com/avatar/%s", hex.EncodeToString(hash[:]))
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		return fmt.Errorf("failed to get user permissions for settings: %w", err)
	}

	ctx := SettingsTemplate{
		User:      c1.User,
		LastLogin: humanize.Time(lastLogin),
		Gravatar:  gravatar,
		TemplateHelper: TemplateHelper{
			UserPermissions: p1,
			ActivePage:      "settings",
			Assumed:         c1.Assumed,
		},
	}

	return v.template.RenderTemplate(c.Response(), ctx, templates.SettingsTemplate, templates.RegularType)
}
