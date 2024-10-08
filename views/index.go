package views

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo/v4"
)

// IndexFunc handles the index page.
func (v *Views) IndexFunc(c echo.Context) error {
	// Data for our HTML template
	c1 := v.getSessionData(c)

	// Check if there is a callback request
	callbackURL, err := url.Parse(c.QueryParam("callback"))
	if err == nil && strings.HasSuffix(callbackURL.Host, v.conf.BaseDomainName) && callbackURL.String() != "" {
		c1.Callback = callbackURL.String()
	}

	// Check if authenticated
	if c1.User.Authenticated {
		return c.Redirect(http.StatusFound, c1.Callback)
	}

	loginCallback := "login?callback=" + c1.Callback

	return c.Redirect(http.StatusFound, loginCallback)
}
