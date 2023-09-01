package views

import (
	"fmt"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"

	"github.com/dustin/go-humanize"
)

type (
	// InternalTemplate represents the context for the internal template
	InternalTemplate struct {
		Nickname        string
		LastLogin       string
		CountAll        user.CountUsers
		UserPermissions []permission.Permission
		ActivePage      string
	}
)

// InternalFunc handles a request to the internal template
func (v *Views) InternalFunc(c echo.Context) error {
	c1 := v.getSessionData(c)
	lastLogin := time.Now()
	if c1.User.LastLogin.Valid {
		lastLogin = c1.User.LastLogin.Time
	}

	countAll, err := v.user.CountUsersAll(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get count users all for interal: %w", err)
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		return fmt.Errorf("failed to get permissions for internal: %w", err)
	}

	ctx := InternalTemplate{
		Nickname:        c1.User.Nickname,
		LastLogin:       humanize.Time(lastLogin),
		CountAll:        countAll,
		UserPermissions: p1,
		ActivePage:      "dashboard",
	}
	return v.template.RenderTemplate(c.Response(), ctx, templates.InternalTemplate, templates.RegularType)
}
