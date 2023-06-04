package views

import (
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/public/templates"
	"log"
	"net/http"
)

// PermissionsFunc handles a permissions request
func (v *Views) PermissionsFunc(w http.ResponseWriter, r *http.Request) {
	session, _ := v.cookie.Get(r, v.conf.SessionCookieName)

	c := v.getData(session)

	permissions, err := v.permission.GetPermissions(r.Context())
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Permissions []permission.Permission
		UserID      int
		ActivePage  string
	}{
		Permissions: permissions,
		UserID:      c.User.UserID,
		ActivePage:  "permissions",
	}

	err = v.template.RenderTemplate(w, data, templates.PermissionsTemplate)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
