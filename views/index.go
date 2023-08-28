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
	loginCallback := fmt.Sprintf("login?callback=%s", c1.Callback)
	return c.Redirect(http.StatusFound, loginCallback)
}
