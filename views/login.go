package views

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/patrickmn/go-cache"
	"github.com/ystv/web-auth/templates"
	"gopkg.in/guregu/null.v4"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/ystv/web-auth/user"
)

type LoginTemplate struct {
	Version string
	Message string
	MsgType string
}

// LoginFunc implements the login functionality, will
// add a cookie to the cookie store for managing authentication
func (v *Views) LoginFunc(c echo.Context) error {
	var err error

	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
	// We're ignoring the error here since sometimes the cookies keys change, and then we
	// can overwrite it instead

	switch c.Request().Method {
	case "GET":
		// Data for our HTML template
		context := v.getData(session)

		// Check if there is a callback request
		callbackURL, err := url.Parse(c.QueryParam("callback"))
		if err == nil && strings.HasSuffix(callbackURL.Host, v.conf.BaseDomainName) && callbackURL.String() != "" {
			context.Callback = callbackURL.String()
		}
		// Check if authenticated
		if context.User.Authenticated {
			//http.Redirect(w, r, context.Callback, http.StatusFound)
			return c.Redirect(http.StatusFound, context.Callback)
		}

		data := LoginTemplate{
			Version: v.conf.Version,
			Message: "",
			MsgType: "",
		}

		return v.template.RenderNoNavsTemplate(c.Response(), data, templates.LoginTemplate)
	case "POST":
		// Parsing form to struct
		err = c.Request().ParseForm()
		if err != nil {
			log.Printf("parse form fail: %v", err)
			return v.errorHandle(c, err)
		}
		username := c.FormValue("username")
		password := c.FormValue("password")
		var u user.User
		// Since we let users enter either an email or username, it's easier
		// to just let it both for the query
		u.Username = username
		u.Email = username
		u.LDAPUsername = null.StringFrom(username)
		u.Password = null.StringFrom(password)

		callback := "/internal"
		callbackURL, err := url.Parse(c.QueryParam("callback"))
		if err == nil && strings.HasSuffix(callbackURL.Host, v.conf.BaseDomainName) && callbackURL.String() != "" {
			callback = callbackURL.String()
		}
		// Authentication
		u, resetPw, err := v.user.VerifyUser(c.Request().Context(), u)
		if err != nil {
			log.Printf("failed login for \"%s\": %v", u.Username, err)
			err = session.Save(c.Request(), c.Response())
			if err != nil {
				//http.Error(w, err.Error(), http.StatusInternalServerError)
				return v.errorHandle(c, err)
			}

			if resetPw {
				ctx := v.getData(session)
				ctx.Callback = callback
				ctx.Message = "Password reset required"
				ctx.MsgType = "is-danger"

				url1 := uuid.NewString()
				v.cache.Set(url1, u.UserID, cache.DefaultExpiration)

				return c.Redirect(http.StatusFound, fmt.Sprintf("https://%s/forgot/%s", v.conf.DomainName, url1))
			}
			ctx := v.getData(session)
			ctx.Callback = callback
			ctx.Message = "Invalid username or password"
			ctx.MsgType = "is-danger"
			return v.template.RenderNoNavsTemplate(c.Response(), ctx, templates.LoginTemplate)
		}
		prevLogin := u.LastLogin
		// Update last logged in
		err = v.user.SetUserLoggedIn(c.Request().Context(), u)
		if err != nil {
			err = fmt.Errorf("failed to set user logged in: %w", err)
			return v.errorHandle(c, err)
		}
		u.Authenticated = true
		// This is a bit of a cheat, just so we can have the last login displayed for internal
		u.LastLogin = prevLogin
		session.Values["user"] = u

		if c.Request().Form.Get("remember") != "on" {
			session.Options.MaxAge = 86400 * 31
		}

		err = session.Save(c.Request(), c.Response())
		if err != nil {
			err = fmt.Errorf("failed to save user session: %w", err)
			return v.errorHandle(c, err)
		}

		log.Printf("user \"%s\" is authenticated", u.Username)
		return c.Redirect(http.StatusFound, callback)
	}
	return fmt.Errorf("failed to parse method")
}
