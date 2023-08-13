package views

import (
	"github.com/gorilla/sessions"
	"github.com/ystv/web-auth/user"
)

type (
	// Context is a struct that is applied to the templates.
	Context struct {
		// Message is used for sending a message back to the user trying to log in, might decide to move later as it may not be needed
		Message string
		// MsgType is the bulma.io class used to indicate what should be displayed
		MsgType string
		// Callback is the address to redirect the user to
		Callback string
		// User is the stored logged-in user
		User user.User
	}
)

func (v *Views) getData(s *sessions.Session) *Context {
	val := s.Values["user"]
	u := user.User{}
	u, ok := val.(user.User)
	if !ok {
		u = user.User{Authenticated: false}
	}
	c := Context{
		Callback: "/internal",
		User:     u,
	}
	return &c
}

// DBToTemplateType converts from the DB layer type to the user template type
func DBToTemplateType(dbUser *[]user.User) []UserStripped {
	var tplUsers []UserStripped
	for i := range *dbUser {
		user1 := UserStripped{}
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

//// DBToTemplateTypeSingle converts from the DB layer type to the user template type single
//func DBToTemplateTypeSingle(dbUser user.UserStripped) UserStripped {
//	var tplUsers UserStripped
//	tplUsers.UserID = dbUser.UserID
//	tplUsers.Username = dbUser.Username
//	tplUsers.Nickname = dbUser.Nickname
//	tplUsers.Name = dbUser.Firstname + " " + dbUser.Lastname
//	tplUsers.Email = dbUser.Email
//	tplUsers.Avatar = dbUser.Avatar
//	tplUsers.UseGravatar = dbUser.UseGravatar
//	if dbUser.LastLogin.Valid {
//		tplUsers.LastLogin = dbUser.LastLogin.Time.Format("2006-01-02 15:04:05")
//	} else {
//		tplUsers.LastLogin = "-"
//	}
//	return tplUsers
//}

func (v *Views) removeDuplicate(strSlice []string) []string {
	allKeys := make(map[string]bool)
	var list []string
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
