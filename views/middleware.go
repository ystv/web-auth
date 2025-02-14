package views

import (
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/ystv/web-auth/crowd"
	"github.com/ystv/web-auth/infrastructure/permission"
	"github.com/ystv/web-auth/permission/permissions"
	"github.com/ystv/web-auth/user"
)

// RequiresLogin is a middleware which will be used for each
// httpHandler to check if there is any active session
func (v *Views) RequiresLogin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			log.Printf("failed to get session: %+v", err)

			session, err = v.cookie.New(c.Request(), v.conf.SessionCookieName)
			if err != nil {
				panic(fmt.Errorf("failed to make new session: %w", err))
			}

			err = session.Save(c.Request(), c.Response())
			if err != nil {
				log.Printf("failed to save session for logout: %+v", err)
			}

			return c.Redirect(http.StatusFound, "/")
		}

		c1 := v.getSessionData(c)

		if !c1.User.Authenticated {
			return c.Redirect(http.StatusFound, "/")
		}

		userFromDB, err := v.user.GetUser(c.Request().Context(), c1.User)
		if err != nil {
			log.Printf("failed to get user from db: %+v", err)

			return c.Redirect(http.StatusFound, "/")
		}

		if userFromDB.DeletedBy.Valid || !c1.User.Enabled {
			session.Values["user"] = &user.User{}
			session.Options.MaxAge = -1

			err = session.Save(c.Request(), c.Response())
			if err != nil {
				return fmt.Errorf("failed to save session for requiresLogin: %w", err)
			}

			return c.Redirect(http.StatusFound, "/")
		}

		return next(c)
	}
}

// RequiresLoginJSON is a middleware which is used for each
// httpHandler to check if there is any active session and returns json if not
func (v *Views) RequiresLoginJSON(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			log.Printf("failed to get session for requiresLoginJSON: %+v", err)

			data := struct {
				Error error `json:"error"`
			}{
				Error: err,
			}

			return c.JSON(http.StatusInternalServerError, data)
		}

		c1 := v.getSessionData(c)

		if !c1.User.Authenticated {
			data := struct {
				Error string `json:"error"`
			}{
				Error: "user not logged in",
			}

			return c.JSON(http.StatusUnauthorized, data)
		}

		userFromDB, err := v.user.GetUser(c.Request().Context(), c1.User)
		if err != nil {
			log.Printf("failed to get user for requiresLoginJSON: %+v", err)

			data := struct {
				Error error `json:"error"`
			}{
				Error: fmt.Errorf("failed to get user for requiresLoginJSON: %w", err),
			}

			return c.JSON(http.StatusInternalServerError, data)
		}

		if userFromDB.DeletedBy.Valid || !c1.User.Enabled {
			session.Values["user"] = &user.User{}
			session.Options.MaxAge = -1

			err = session.Save(c.Request(), c.Response())
			if err != nil {
				log.Printf("failed to save session for requiresLoginJSON: %+v", err)
				data := struct {
					Error error `json:"error"`
				}{
					Error: fmt.Errorf("failed to save session for requiresLoginJSON: %w", err),
				}

				return c.JSON(http.StatusInternalServerError, data)
			}

			data := struct {
				Error string `json:"error"`
			}{
				Error: "user deleted or not enabled",
			}

			return c.JSON(http.StatusUnauthorized, data)
		}

		return next(c)
	}
}

// RequiresLoginCrowd is a middleware which is used for crowd auth sites like wiki
func (v *Views) RequiresLoginCrowd(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			log.Printf("failed to get session for requiresLoginCrowd: %+v", err)
			data := struct {
				XMLName xml.Name `xml:"errors"`
				Error   error    `xml:"error"`
			}{
				Error: fmt.Errorf("failed to get session for requiresLoginCrowd: %w", err),
			}

			return c.XML(http.StatusInternalServerError, data)
		}

		c1 := v.getSessionData(c)

		if !c1.User.Authenticated {
			data := struct {
				XMLName xml.Name `xml:"errors"`
				Error   error    `xml:"error"`
			}{
				Error: errors.New("user not logged in"),
			}

			return c.XML(http.StatusUnauthorized, data)
		}

		userFromDB, err := v.user.GetUser(c.Request().Context(), c1.User)
		if err != nil {
			log.Printf("failed to get user for requiresLoginCrowd: %+v", err)
			data := struct {
				XMLName xml.Name `xml:"errors"`
				Error   error    `xml:"error"`
			}{
				Error: fmt.Errorf("failed to get roles for user: %w", err),
			}

			return c.XML(http.StatusInternalServerError, data)
		}

		if userFromDB.DeletedBy.Valid || !c1.User.Enabled {
			session.Values["user"] = &user.User{}
			session.Options.MaxAge = -1

			err = session.Save(c.Request(), c.Response())
			if err != nil {
				log.Printf("failed to save session for requiresLoginCrowd: %+v", err)
				data := struct {
					XMLName xml.Name `xml:"errors"`
					Error   error    `xml:"error"`
				}{
					Error: fmt.Errorf("failed to get roles for user: %w", err),
				}

				return c.XML(http.StatusInternalServerError, data)
			}

			data := struct {
				XMLName xml.Name `xml:"errors"`
				Error   error    `xml:"error"`
			}{
				Error: errors.New("user deleted or not enabled"),
			}

			return c.XML(http.StatusUnauthorized, data)
		}

		auth := c.Request().Header.Get(echo.HeaderAuthorization)
		l := len("basic")

		if len(auth) > l+1 && strings.EqualFold(auth[:l], "basic") {
			var b []byte
			// Invalid base64 shouldn't be treated as error
			// instead should be treated as invalid client input
			b, err = base64.StdEncoding.DecodeString(auth[l+1:])
			if err != nil {
				log.Printf("failed to decode basic auth header: %+v", err)
				data := struct {
					XMLName xml.Name `xml:"errors"`
					Error   error    `xml:"error"`
				}{
					Error: errors.New("failed to decode basic auth header"),
				}

				return c.XML(http.StatusBadRequest, data)
			}

			cred := string(b)
			for i := 0; i < len(cred); i++ {
				if cred[i] == ':' {
					var valid bool
					// Verify credentials
					valid, err = func(username, password string, c echo.Context) (bool, error) {
						crowd1 := crowd.CrowdApp{
							Name:     username,
							Password: null.StringFrom(password),
						}
						var crowd2 crowd.CrowdApp
						crowd2, err = v.crowd.VerifyCrowd(c.Request().Context(), crowd1)
						if err != nil {
							return false, err
						}

						if crowd2.AppID > 0 {
							return true, nil
						}

						return false, echo.NewHTTPError(http.StatusUnauthorized).SetInternal(fmt.Errorf("invalid credential"))
					}(cred[:i], cred[i+1:], c)
					if err != nil {
						log.Printf("invalid app credentials: %+v", err)
						data := struct {
							XMLName xml.Name `xml:"errors"`
							Error   error    `xml:"error"`
						}{
							Error: errors.New("invalid app credentials"),
						}

						return c.XML(http.StatusUnauthorized, data)
					} else if valid {
						return next(c)
					}
					break
				}
			}
		}

		log.Printf("app not logged in")
		data := struct {
			XMLName xml.Name `xml:"errors"`
			Error   error    `xml:"error"`
		}{
			Error: errors.New("app not logged in"),
		}

		return c.XML(http.StatusUnauthorized, data)
	}
}

func (v *Views) RequirePermission(p permissions.Permissions) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c1 := v.getSessionData(c)
			if c1 == nil {
				return errors.New("failed to get session data")
			}

			perms, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
			if err != nil {
				return fmt.Errorf("failed to get permissions for requirePermission: %w", err)
			}

			acceptedPerms := permission.SufficientPermissionsFor(p)

			for _, perm := range perms {
				if acceptedPerms[perm.Name] {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, errors.New("you are not authorised for accessing this"))
		}
	}
}
