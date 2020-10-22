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
func InternalFunc(w http.ResponseWriter, r *http.Request) {
	session, err := cStore.Get(r, "session")
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
	err = tpl.ExecuteTemplate(w, "internal.gohtml", ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UsersFunc handles a users request
func UsersFunc(w http.ResponseWriter, r *http.Request) {
	dbUsers := &[]types.User{}
	err := uStore.GetUsers(r.Context(), dbUsers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tplUsers := DBToTemplateType(dbUsers)

	ctx := UsersTemplate{
		Users: tplUsers,
	}
	err = tpl.ExecuteTemplate(w, "users.gohtml", ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UserFunc handles a users request
func UserFunc(w http.ResponseWriter, r *http.Request) {
	userString := mux.Vars(r)
	dbUser := &types.User{}
	var err error
	dbUser.UserID, err = strconv.Atoi(userString["userid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = uStore.GetUser(r.Context(), dbUser)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tpl.ExecuteTemplate(w, "user.gohtml", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// RequiresLogin is a middleware which will be used for each
// httpHandler to check if there is any active session
func RequiresLogin(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := cStore.Get(r, "session")
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
