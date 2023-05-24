package views

import (
	"fmt"
	"github.com/ystv/web-auth/public/templates"
	"net/http"

	"github.com/ystv/web-auth/helpers"
	"github.com/ystv/web-auth/user"
)

// RequiresLogin is a middleware which will be used for each
// httpHandler to check if there is any active session
func (v *Views) RequiresLogin(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//fmt.Println("DEBUG - REQUIRE LOGIN")
		session, err := v.cookie.Get(r, v.conf.SessionCookieName)
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
		//fmt.Println("DEBUG - REQUIRE PERMISSION")
		session, err := v.cookie.Get(r, v.conf.SessionCookieName)
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//fmt.Println(session, session.Options, session.Values, v.conf.SessionCookieName)
		u := helpers.GetUser(session)
		//fmt.Println(u)
		perms, err := v.user.GetPermissions(r.Context(), u)
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		//fmt.Println(perms)
		for _, perm := range perms {
			fmt.Println(perm, p, p.Name)
			if perm == p.Name {
				h.ServeHTTP(w, r)
				return
			}
		}
		err = v.template.RenderNoNavsTemplate(w, nil, templates.ForbiddenTemplate)
		//err = v.tpl.ExecuteTemplate(w, "forbidden.tmpl", nil)
		if err != nil {
			//fmt.Println(3, err)
			fmt.Println(err)
		}
		//fmt.Println(4)
		w.WriteHeader(http.StatusForbidden)
	}
}
