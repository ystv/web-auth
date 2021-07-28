package views

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// IndexFunc handles the index page.
func (v Views) IndexFunc(w http.ResponseWriter, r *http.Request) {
	session, _ := v.cookie.Get(r, "session")

	// Data for our HTML template
	context := v.getData(session)

	// Check if there is a callback request
	callbackURL, err := url.Parse(r.URL.Query().Get("callback"))
	if err == nil && strings.HasSuffix(callbackURL.Host, v.conf.DomainName) && callbackURL.String() != "" {
		context.Callback = callbackURL.String()
	}

	// Check if authenticated
	if context.User.Authenticated {
		http.Redirect(w, r, context.Callback, http.StatusFound)
		return
	}
	loginCallback := fmt.Sprintf("login?callback=%s", context.Callback)
	http.Redirect(w, r, loginCallback, http.StatusFound)
}
