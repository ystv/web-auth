package views

import (
	"github.com/ystv/web-auth/public/templates"
	"github.com/ystv/web-auth/role"
	"log"
	"net/http"
)

// RolesFunc handles a roles request
func (v *Views) RolesFunc(w http.ResponseWriter, r *http.Request) {
	session, _ := v.cookie.Get(r, v.conf.SessionCookieName)

	c := v.getData(session)

	roles, err := v.role.GetRoles(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Roles      []role.Role
		UserID     int
		ActivePage string
	}{
		Roles:      roles,
		UserID:     c.User.UserID,
		ActivePage: "roles",
	}

	err = v.template.RenderTemplate(w, data, templates.RolesTemplate)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
