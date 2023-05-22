package views

import (
	"github.com/ystv/web-auth/public/templates"
	"net/http"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gorilla/mux"
	"github.com/ystv/web-auth/user"
)

type (
	// InternalTemplate represents the context for the internal template
	InternalTemplate struct {
		Nickname      string
		LastLogin     string
		TotalUsers    int
		LoginsPastDay int
		ActivePage    string
	}
	// UsersTemplate represents the context for the user template
	UsersTemplate struct {
		Users                                 []User
		CurPage, NextPage, PrevPage, LastPage int
		ActivePage                            string
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
func DBToTemplateType(dbUser *[]user.User) []User {
	var tplUsers []User
	for i := range *dbUser {
		user1 := User{}
		user1.UserID = (*dbUser)[i].UserID
		user1.Username = (*dbUser)[i].Username
		user1.Name = (*dbUser)[i].Firstname + " " + (*dbUser)[i].Lastname
		user1.Email = (*dbUser)[i].Email
		if (*dbUser)[i].LastLogin.Valid {
			user1.LastLogin = (*dbUser)[i].LastLogin.Time.Format("2006-01-02 15:04:05")
		} else {
			user1.LastLogin = "-"
		}
		tplUsers = append(tplUsers, user1)
	}
	return tplUsers
}

// InternalFunc handles a request to the internal template
func (v *Views) InternalFunc(w http.ResponseWriter, r *http.Request) {
	session, _ := v.cookie.Get(r, v.conf.SessionCookieName)

	c := v.getData(session)
	lastLogin := time.Now()
	if c.User.LastLogin.Valid {
		lastLogin = c.User.LastLogin.Time
	}
	ctx := InternalTemplate{
		Nickname:      c.User.Nickname,
		LastLogin:     humanize.Time(lastLogin),
		TotalUsers:    2000,
		LoginsPastDay: 20,
		ActivePage:    "dashboard",
	}
	//err := v.tpl.ExecuteTemplate(w, "internal.tmpl", ctx)
	err := v.template.RenderTemplate(w, ctx, templates.InternalTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UsersFunc handles a users request
func (v *Views) UsersFunc(w http.ResponseWriter, r *http.Request) {

	dbUsers, err := v.user.GetUsers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tplUsers := DBToTemplateType(&dbUsers)

	ctx := UsersTemplate{
		Users:      tplUsers,
		ActivePage: "users",
	}
	err = v.template.RenderTemplate(w, ctx, templates.UsersTemplate)
	//err = v.tpl.ExecuteTemplate(w, "users.tmpl", ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UserFunc handles a users request
func (v *Views) UserFunc(w http.ResponseWriter, r *http.Request) {
	userString := mux.Vars(r)
	userID, err := strconv.Atoi(userString["userid"])
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user1, err := v.user.GetUser(r.Context(), user.User{UserID: userID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		User       user.User
		ActivePage string
	}{
		ActivePage: "user",
		User:       user1,
	}

	err = v.template.RenderTemplate(w, data, templates.UserTemplate)
	//err = v.tpl.ExecuteTemplate(w, "user.tmpl", struct {ActivePage string}{ActivePage: "user"})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
