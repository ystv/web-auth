package views

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/infrastructure/permission"
	"github.com/ystv/web-auth/permission/permissions"
	"github.com/ystv/web-auth/user"
	"log"
	"net/http"

	"github.com/ystv/web-auth/helpers"
)

// RequiresLogin is a middleware which will be used for each
// httpHandler to check if there is any active session
func (v *Views) RequiresLogin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			log.Println(err)
			return v.LoginFunc(c)
		}
		user1 := helpers.GetUser(session)
		user2, err := v.user.GetUser(c.Request().Context(), user1)
		if err != nil {
			log.Println(err)
			return err
		}
		if user2.DeletedBy.Valid || !user1.Enabled {
			session.Values["user"] = &user.User{}
			session.Options.MaxAge = -1
			err = session.Save(c.Request(), c.Response())
			if err != nil {
				//http.Error(w, err.Error(), http.StatusInternalServerError)
				return v.errorHandle(c, err)
			}
			return c.Redirect(http.StatusFound, "/")
		}
		if !user1.Authenticated {
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

			u := helpers.GetUser(session)

			perms, err := v.user.GetPermissionsForUser(c.Request().Context(), u)
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

			c.Response().WriteHeader(http.StatusForbidden)
			return v.Error500(c)
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
