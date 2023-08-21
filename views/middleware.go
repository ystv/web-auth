package views

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/infrastructure/permission"
	"github.com/ystv/web-auth/permission/permissions"
	"github.com/ystv/web-auth/user"
	"log"
	"net/http"
)

// RequiresLogin is a middleware which will be used for each
// httpHandler to check if there is any active session
func (v *Views) RequiresLogin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			log.Println(err)
			return c.Redirect(http.StatusFound, "/")
		}
		c1 := v.getData(session)
		if !c1.User.Authenticated {
			return c.Redirect(http.StatusFound, "/")
		}
		userFromDB, err := v.user.GetUser(c.Request().Context(), c1.User)
		if err != nil {
			log.Println(err)
			return c.Redirect(http.StatusFound, "/")
		}
		if userFromDB.DeletedBy.Valid || !c1.User.Enabled {
			session.Values["user"] = &user.User{}
			session.Options.MaxAge = -1
			err = session.Save(c.Request(), c.Response())
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to save session for requiresLogin: %w", err))
			}
			return c.Redirect(http.StatusFound, "/")
		}
		if !c1.User.Authenticated {
			// Not authenticated
			return c.Redirect(http.StatusFound, "/")
		}
		return next(c)
	}
}

func (v *Views) RequirePermission(p permissions.Permissions) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
			if err != nil {
				log.Println(err)
				http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
				return v.LoginFunc(c)
			}

			c1 := v.getData(session)

			perms, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
			if err != nil {
				log.Println(err)
				http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
				return v.LoginFunc(c)
			}

			acceptedPerms := permission.SufficientPermissionsFor(p)

			for _, perm := range perms {
				if acceptedPerms[perm.Name] {
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden, fmt.Errorf("you are not authorised for accessing this"))
		}
	}
}

func (v *Views) RequiresMinimumPermissionNoHttp(userID int, p permissions.Permissions) bool {
	u, err := v.user.GetUser(context.Background(), user.User{UserID: userID})
	if err != nil {
		log.Println(err)
		return false
	}

	p1, err := v.user.GetPermissionsForUser(context.Background(), u)
	if err != nil {
		log.Println(err)
		return false
	}

	m := permission.SufficientPermissionsFor(p)

	for _, perm := range p1 {
		if m[perm.Name] {
			return true
		}
	}

	return false
}
