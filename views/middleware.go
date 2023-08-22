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
		c1 := v.getData(session)
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

func (v *Views) RequirePermission(p permissions.Permissions) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
			if err != nil {
				return fmt.Errorf("failed to get session for requirePermission: %w", err)
			}

			c1 := v.getData(session)

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
