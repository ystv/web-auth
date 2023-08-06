package views

import (
	"context"
	"github.com/ystv/web-auth/infrastructure/permission"
	"github.com/ystv/web-auth/permission/permissions"
	"github.com/ystv/web-auth/public/templates"
	"github.com/ystv/web-auth/user"
	"log"
	"net/http"

	"github.com/ystv/web-auth/helpers"
)

// RequiresLogin is a middleware which will be used for each
// httpHandler to check if there is any active session
func (v *Views) RequiresLogin(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := v.cookie.Get(r, v.conf.SessionCookieName)
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if !helpers.GetUser(session).Authenticated {
			// Not authenticated
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		h.ServeHTTP(w, r)
	}
}

// RequiresMinimumPermission is a middleware that will
// ensure that the user has the given permission.
func (v *Views) RequiresMinimumPermission(h http.Handler, p permissions.Permissions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := v.cookie.Get(r, v.conf.SessionCookieName)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		u := helpers.GetUser(session)

		if v.RequiresMinimumPermissionNoHttp(u.UserID, p) {
			h.ServeHTTP(w, r)
			return
		}

		err = v.template.RenderNoNavsTemplate(w, nil, templates.Forbidden500Template)
		if err != nil {
			log.Println(err)
		}
		w.WriteHeader(http.StatusForbidden)
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

	m := permission.GetValidPermissions(p)

	for _, perm := range p1 {
		if m[perm] {
			return true
		}
	}

	return false
}
