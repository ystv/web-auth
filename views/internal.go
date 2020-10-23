package views

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"github.com/ystv/web-auth/helpers"
	"github.com/ystv/web-auth/types"
)

type (
	// InternalTemplate represents the context for the internal template
	InternalTemplate struct {
		Nickname      string
		LastLogin     string
		TotalUsers    int
		LoginsPastDay int
	}
	// UsersTemplate represents the context for the user template
	UsersTemplate struct {
		Users                                 []User
		CurPage, NextPage, PrevPage, LastPage int
	}
	// User represents user information, an administrator can view
	User struct {
		UserID    int
		Username  string
		Name      string
		Email     string
		LastLogin string
	}
)

// DBToTemplateType converts from the DB layer type to the user template type
func DBToTemplateType(dbUser *[]types.User) []User {
	tplUsers := []User{}
	user := User{}
	for i := range *dbUser {
		user.UserID = (*dbUser)[i].UserID
		user.Username = (*dbUser)[i].Username
		user.Name = (*dbUser)[i].Firstname + " " + (*dbUser)[i].Lastname
		user.Email = (*dbUser)[i].Email
		if (*dbUser)[i].LastLogin.Valid {
			user.LastLogin = (*dbUser)[i].LastLogin.Time.Format("2006-01-02 15:04:05")
		} else {
			user.LastLogin = "-"
		}
		tplUsers = append(tplUsers, user)
	}
	return tplUsers
}

// InternalFunc handles a request to the internal template
func (v *Views) InternalFunc(w http.ResponseWriter, r *http.Request) {
	session, err := v.cookie.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c := getData(session)
	lastLogin := time.Now()
	if c.User.LastLogin.Valid {
		lastLogin = c.User.LastLogin.Time
	}
	ctx := InternalTemplate{
		Nickname:      c.User.Nickname,
		LastLogin:     humanize.Time(lastLogin),
		TotalUsers:    2000,
		LoginsPastDay: 20,
	}
	err = v.tpl.ExecuteTemplate(w, "internal.gohtml", ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UsersFunc handles a users request
func (v *Views) UsersFunc(w http.ResponseWriter, r *http.Request) {
	dbUsers := &[]types.User{}
	err := v.user.GetUsers(r.Context(), dbUsers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tplUsers := DBToTemplateType(dbUsers)

	ctx := UsersTemplate{
		Users: tplUsers,
	}
	err = v.tpl.ExecuteTemplate(w, "users.gohtml", ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UserFunc handles a users request
func (v *Views) UserFunc(w http.ResponseWriter, r *http.Request) {
	userString := mux.Vars(r)
	dbUser := &types.User{}
	var err error
	dbUser.UserID, err = strconv.Atoi(userString["userid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = v.user.GetUser(r.Context(), dbUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = v.tpl.ExecuteTemplate(w, "user.gohtml", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

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
func (v *Views) RequiresPermission(h http.Handler, p types.Permission) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := v.cookie.Get(r, "session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		u := helpers.GetUser(session)
		err = v.user.GetPermissions(r.Context(), &u)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, perm := range u.Permissions {
			if perm == p {
				h.ServeHTTP(w, r)
				return
			}
		}
		v.tpl.ExecuteTemplate(w, "forbidden.gohtml", nil)
		w.WriteHeader(http.StatusForbidden)
	}
}
