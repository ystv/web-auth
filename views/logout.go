package views

import (
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/user"
	"net/http"
)

// LogoutFunc Implements the logout functionality.
// Will delete the session information from the cookie store
func (v *Views) LogoutFunc(c echo.Context) error {
	session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
	if err != nil {
		return v.errorHandle(c, err)
	}

	session.Values["user"] = user.User{}
	session.Options.MaxAge = -1
	err = session.Save(c.Request(), c.Response())
	if err != nil {
		return v.errorHandle(c, err)
	}
	// TODO Don't call env in this function have an initialiser
	// then fetch from that store?
	endpoint := v.conf.LogoutEndpoint
	if endpoint == "" {
		endpoint = "/"
	}
	return c.Redirect(http.StatusFound, endpoint)
}