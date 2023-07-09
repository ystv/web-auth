package views

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
	"strings"
	"time"
)

type (
	SettingsTemplate struct {
		User       *user.User
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
	lastLogin := time.Now()
	if c1.User.LastLogin.Valid {
		lastLogin = c1.User.LastLogin.Time
	}

	var gravatar string

	if c1.User.UseGravatar {
		hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace("liam.burnand@bswdi.co.uk"))))
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
