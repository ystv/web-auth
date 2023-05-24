package views

import (
	"fmt"
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
		Nickname            string
		LastLogin           string
		TotalUsers          int
		LoginsPast24Hours   int
		ActiveUsersPastYear int
		ActivePage          string
	}
	SettingsTemplate struct {
		User       User
		LastLogin  string
		ActivePage string
	}
	// UsersTemplate represents the context for the user template
	UsersTemplate struct {
		Users                                 []User
		CurPage, NextPage, PrevPage, LastPage int
		ActivePage                            string
		Sort                                  struct {
			Column    string
			Direction string
		}
	}
	// User represents user information, an administrator can view
	User struct {
		UserID      int
		Username    string
		Nickname    string
		Name        string
		Email       string
		LastLogin   string
		Avatar      string
		UseGravatar bool
	}
)

// DBToTemplateType converts from the DB layer type to the user template type
func DBToTemplateType(dbUser *[]user.User) []User {
	var tplUsers []User
	for i := range *dbUser {
		user1 := User{}
		user1.UserID = (*dbUser)[i].UserID
		user1.Username = (*dbUser)[i].Username
		user1.Nickname = (*dbUser)[i].Nickname
		user1.Name = (*dbUser)[i].Firstname + " " + (*dbUser)[i].Lastname
		user1.Email = (*dbUser)[i].Email
		user1.Avatar = (*dbUser)[i].Avatar
		user1.UseGravatar = (*dbUser)[i].UseGravatar
		if (*dbUser)[i].LastLogin.Valid {
			user1.LastLogin = (*dbUser)[i].LastLogin.Time.Format("2006-01-02 15:04:05")
		} else {
			user1.LastLogin = "-"
		}
		tplUsers = append(tplUsers, user1)
	}
	return tplUsers
}

// DBToTemplateTypeSingle converts from the DB layer type to the user template type single
func DBToTemplateTypeSingle(dbUser user.User) User {
	var tplUsers User
	tplUsers.UserID = dbUser.UserID
	tplUsers.Username = dbUser.Username
	tplUsers.Nickname = dbUser.Nickname
	tplUsers.Name = dbUser.Firstname + " " + dbUser.Lastname
	tplUsers.Email = dbUser.Email
	tplUsers.Avatar = dbUser.Avatar
	tplUsers.UseGravatar = dbUser.UseGravatar
	if dbUser.LastLogin.Valid {
		tplUsers.LastLogin = dbUser.LastLogin.Time.Format("2006-01-02 15:04:05")
	} else {
		tplUsers.LastLogin = "-"
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
	count, err := v.user.CountUsers(r.Context())
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hours24, err := v.user.CountUsers24Hours(r.Context())
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pastYear, err := v.user.CountUsersPastYear(r.Context())
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	ctx := InternalTemplate{
		Nickname:            c.User.Nickname,
		LastLogin:           humanize.Time(lastLogin),
		TotalUsers:          count,
		LoginsPast24Hours:   hours24,
		ActiveUsersPastYear: pastYear,
		ActivePage:          "dashboard",
	}
	//err := v.tpl.ExecuteTemplate(w, "internal.tmpl", ctx)
	err = v.template.RenderTemplate(w, ctx, templates.InternalTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// SettingsFunc handles a request to the internal template
func (v *Views) SettingsFunc(w http.ResponseWriter, r *http.Request) {
	session, _ := v.cookie.Get(r, v.conf.SessionCookieName)

	c := v.getData(session)
	lastLogin := time.Now()
	if c.User.LastLogin.Valid {
		lastLogin = c.User.LastLogin.Time
	}

	tplUser := DBToTemplateTypeSingle(c.User)

	ctx := SettingsTemplate{
		User:       tplUser,
		LastLogin:  humanize.Time(lastLogin),
		ActivePage: "settings",
	}
	//err := v.tpl.ExecuteTemplate(w, "internal.tmpl", ctx)
	err := v.template.RenderTemplate(w, ctx, templates.SettingsTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// UsersFunc handles a users request
func (v *Views) UsersFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			err = v.errorHandle(w, err)
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		column := r.FormValue("column")
		direction := r.FormValue("direction")
		valid := true
		switch column {
		case "userId":
		case "name":
		case "username":
		case "email":
		case "lastLogin":
			break
		default:
			valid = false
		}
		switch direction {
		case "asc":
		case "desc":
			break
		default:
			valid = false
		}
		if valid {
			http.Redirect(w, r, fmt.Sprintf("/internal/users?column=%s&direction=%s", column, direction), http.StatusFound)
		}
	}
	column := r.URL.Query().Get("column")
	direction := r.URL.Query().Get("direction")
	fmt.Println(column, direction)
	valid := true
	switch column {
	case "userId":
	case "name":
	case "username":
	case "email":
	case "lastLogin":
		break
	default:
		valid = false
	}
	switch direction {
	case "asc":
	case "desc":
		break
	default:
		valid = false
	}
	var dbUsers []user.User
	var err error
	if valid {
		dbUsers, err = v.user.GetUsersSorted(r.Context(), column, direction)
	} else {
		dbUsers, err = v.user.GetUsers(r.Context())
	}
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tplUsers := DBToTemplateType(&dbUsers)

	ctx := UsersTemplate{
		Users:      tplUsers,
		ActivePage: "users",
		Sort: struct {
			Column    string
			Direction string
		}{Column: column, Direction: direction},
	}
	err = v.template.RenderTemplate(w, ctx, templates.UsersTemplate)
	//err = v.tpl.ExecuteTemplate(w, "users.tmpl", ctx)
	if err != nil {
		fmt.Println(err)
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
