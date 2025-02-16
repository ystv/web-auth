package views

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	emailParser "github.com/mcnijman/go-emailaddress"
	"github.com/patrickmn/go-cache"
	"gopkg.in/guregu/null.v4"

	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
)

// LoginFunc implements the login functionality, will
// add a cookie to the cookie store for managing authentication
func (v *Views) LoginFunc(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodGet:
		return v._loginGet(c)
	case http.MethodPost:
		return v._loginPost(c)
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) _loginGet(c echo.Context) error {
	// Data for our HTML template
	context := v.getSessionData(c)

	// Check if there is a callback request
	callbackURL, err := url.Parse(c.QueryParam("callback"))
	if err == nil && strings.HasSuffix(callbackURL.Host, v.conf.BaseDomainName) && callbackURL.String() != "" {
		context.Callback = callbackURL.String()
	}
	// Check if authenticated
	if context.User.Authenticated {
		return c.Redirect(http.StatusFound, context.Callback)
	}

	return v.template.RenderTemplate(c.Response(), context, templates.LoginTemplate, templates.NoNavType)
}

func (v *Views) _loginPost(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
	// We're ignoring the error here since sometimes the cookies keys change, and then we
	// can overwrite it instead, it does need to stay as it is written to here

	// Parsing form to struct
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
	if err != nil {
		log.Printf("failed to parse callback url: %+v", err)
	}
	if err == nil && strings.HasSuffix(callbackURL.Host, v.conf.BaseDomainName) && callbackURL.String() != "" {
		callback = callbackURL.String()
	}
	// Authentication
	u, resetPw, err := v.user.VerifyUser(c.Request().Context(), u)
	if err != nil {
		address, _ := emailParser.Parse(username)
		if address != nil {
			u.LDAPUsername = null.StringFrom(address.LocalPart)
			_, err = v.user.GetUser(c.Request().Context(), u)
		}
		log.Printf("failed login for \"%s\": %v", u.Username, err)

		err = session.Save(c.Request(), c.Response())
		if err != nil {
			return fmt.Errorf("failed to save session for login: %w", err)
		}

		if resetPw {
			ctx := v.getSessionData(c)
			ctx.Callback = callback
			ctx.Message = "Password reset required"
			ctx.MsgType = "is-danger"

			err = v.setMessagesInSession(c, ctx)
			if err != nil {
				return fmt.Errorf("failed to set message for login: %w", err)
			}

			url1 := uuid.NewString()
			v.cache.Set(url1, u.UserID, cache.DefaultExpiration)

			return c.Redirect(http.StatusFound, fmt.Sprintf("https://%s/reset/%s", v.conf.DomainName, url1))
		}

		ctx := v.getSessionData(c)
		ctx.Callback = callback
		ctx.Message = "Invalid username or password"
		ctx.MsgType = "is-danger"

		err = v.setMessagesInSession(c, ctx)
		if err != nil {
			return fmt.Errorf("failed to set message for login: %w", err)
		}

		return c.Redirect(http.StatusFound, "/login")
	}

	prevLogin := u.LastLogin
	// Update last logged in
	err = v.user.SetUserLoggedIn(c.Request().Context(), u)
	if err != nil {
		return fmt.Errorf("failed to set user logged in for login: %w", err)
	}

	u.Authenticated = true
	// This is a bit of a cheat, just so we can have the last login displayed for internal
	u.LastLogin = prevLogin

	err = v.clearMessagesInSession(c)
	if err != nil {
		return fmt.Errorf("failed to clear message: %w", err)
	}

	session.Values["user"] = u

	eightySixFourHundred := 86400
	thirtyOne := 31

	if c.FormValue("remember") != "on" {
		session.Options.MaxAge = eightySixFourHundred * thirtyOne
	}

	err = session.Save(c.Request(), c.Response())
	if err != nil {
		return fmt.Errorf("failed to save user session for login: %w", err)
	}

	log.Printf("user \"%s\" is authenticated", u.Username)

	return c.Redirect(http.StatusFound, callback)
}

//func (v *Views) LDAPFunc(username, password string) (bool, error) {
//	config := &auth.Config{
//		Server:   v.conf.AD.Server,
//		Port:     v.conf.AD.Port,
//		BaseDN:   v.conf.AD.BaseDN,
//		Security: auth.SecurityType(v.conf.AD.Security),
//	}
//
//	conn, err := config.Connect()
//	if err != nil {
//		return false, echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("error connecting to server: %w", err))
//	}
//	defer func(Conn *ldap.Conn) {
//		err = Conn.Close()
//		if err != nil {
//			log.Printf("failed to close to LDAP server: %+v", err)
//		}
//	}(conn.Conn)
//
//	status, err := conn.Bind(v.conf.AD.Bind.Username, v.conf.AD.Bind.Password)
//	if err != nil {
//		return false, echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("error binding to server: %w", err))
//	}
//
//	if !status {
//		return false, echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("error binding to server: invalid credentials"))
//	}
//
//	status1, err := auth.Authenticate(config, username, password)
//	if err != nil {
//		return false, echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("unable to authenticate %s with error: %w", username, err))
//	}
//
//	if status1 {
//		var entry *ldap.Entry
//		if _, err = mail.ParseAddress(username); err == nil {
//			entry, err = conn.GetAttributes("userPrincipalName", username, []string{"memberOf"})
//		} else {
//			entry, err = conn.GetAttributes("samAccountName", username, []string{"memberOf"})
//		}
//		if err != nil {
//			return false, echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("error getting user groups: %w", err))
//		}
//
//		dnGroups := entry.GetAttributeValues("memberOf")
//
//		if len(dnGroups) == 0 {
//			return false, echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("BIND_SAM user not member of any groups"))
//		}
//
//		//stv := false
//
//		for _, group := range dnGroups {
//			if group == "CN=STV Admin,CN=Users,DC=ystv,DC=local" {
//				//stv = true
//				return true, nil
//			}
//		}
//
//		//if !stv {
//		//	return false, echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("STV not allowed for %s!\n", username))
//		//}
//		log.Printf("%s is authenticated", username)
//		return true, nil
//	}
//	return false, echo.NewHTTPError(http.StatusUnauthorized, fmt.Errorf("user not authenticated: %s", username))
//}
