package views

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/user"
)

// LogoutFunc Implements the logout functionality.
// Will delete the session information from the cookie store
func (v *Views) LogoutFunc(c echo.Context) error {
	session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
	if err != nil {
		return fmt.Errorf("failed to get session for logout: %w", err)
	}

	session.Values["user"] = user.User{}
	session.Options.MaxAge = -1
	err = session.Save(c.Request(), c.Response())
	if err != nil {
		return fmt.Errorf("failed to save session for logout: %w", err)
	}
	endpoint := v.conf.LogoutEndpoint
	if endpoint == "" {
		endpoint = "/"
	}
	return c.Redirect(http.StatusFound, endpoint)
}
