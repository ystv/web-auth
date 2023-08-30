package views

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
	"net/http"
	"strings"
	"time"
)

type (
	SettingsTemplate struct {
		User            user.User
		UserPermissions []permission.Permission
		LastLogin       string
		ActivePage      string
		Gravatar        string
	}
)

// SettingsFunc handles a request to the internal template
func (v *Views) SettingsFunc(c echo.Context) error {
	c1 := v.getSessionData(c)
	lastLogin := time.Now()
	if c1.User.LastLogin.Valid {
		lastLogin = c1.User.LastLogin.Time
	}

	var gravatar string

	if c1.User.UseGravatar {
		hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(c1.User.Email))))
		gravatar = fmt.Sprintf("https://www.gravatar.com/avatar/%s", hex.EncodeToString(hash[:]))
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get user permissions for settings: %+v", err))
	}

	ctx := SettingsTemplate{
		User:            c1.User,
		UserPermissions: p1,
		LastLogin:       humanize.Time(lastLogin),
		ActivePage:      "settings",
		Gravatar:        gravatar,
	}

	return v.template.RenderTemplate(c.Response(), ctx, templates.SettingsTemplate, templates.RegularType)
}
