package views

import (
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/public/templates"
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

// RequiresPermission is a middleware that will
// ensure that the user has the given permission.
func (v *Views) RequiresPermission(h http.Handler, p permission.Permission) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := v.cookie.Get(r, v.conf.SessionCookieName)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		u := helpers.GetUser(session)

		perms, err := v.user.GetPermissionsForUser(r.Context(), u)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, perm := range perms {
			if perm.Name == p.Name {
				h.ServeHTTP(w, r)
				return
			}
		}

		err = v.template.RenderNoNavsTemplate(w, nil, templates.Forbidden500Template)
		if err != nil {
			log.Println(err)
		}
		w.WriteHeader(http.StatusForbidden)
	}
}
