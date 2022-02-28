package views

import (
	"net/http"

	"github.com/ystv/web-auth/helpers"
	"github.com/ystv/web-auth/user"
)

// RequiresLogin is a middleware which will be used for each
// httpHandler to check if there is any active session
func (v *Views) RequiresLogin(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := v.cookie.Get(r, "session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
func (v *Views) RequiresPermission(h http.Handler, p user.Permission) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := v.cookie.Get(r, "session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		u := helpers.GetUser(session)
		perms, err := v.user.GetPermissions(r.Context(), u)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, perm := range perms {
			if perm == p.Name {
				h.ServeHTTP(w, r)
				return
			}
		}
		v.tpl.ExecuteTemplate(w, "forbidden.tmpl", nil)
		w.WriteHeader(http.StatusForbidden)
	}
}
