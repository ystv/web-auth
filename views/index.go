package views

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"net/url"
	"strings"
)

// IndexFunc handles the index page.
func (v *Views) IndexFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	// Data for our HTML template
	context := v.getData(session)

	// Check if there is a callback request
	callbackURL, err := url.Parse(c.Request().URL.Query().Get("callback"))
	if err == nil && strings.HasSuffix(callbackURL.Host, v.conf.BaseDomainName) && callbackURL.String() != "" {
		context.Callback = callbackURL.String()
	}

	// Check if authenticated
	if context.User.Authenticated {
		return c.Redirect(http.StatusFound, context.Callback)
	}
	loginCallback := fmt.Sprintf("login?callback=%s", context.Callback)
	return c.Redirect(http.StatusFound, loginCallback)
}
