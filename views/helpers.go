package views

import (
	"context"
	// #nosec
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/user"
	"gopkg.in/guregu/null.v4"
)

type (
	// Context is a struct that is applied to the templates.
	Context struct {
		// Title is used for sending pages to the user with custom titles
		Title string
		// Message is used for sending a message back to the user trying to log in, might decide to move later as it may not be needed
		Message string
		// MsgType is the bulma.io class used to indicate what should be displayed
		MsgType string
		// Callback is the address to redirect the user to
		Callback string
		// User is the stored logged-in user
		User user.User
		// Version is the version that is running
		Version string
	}

	InternalContext struct {
		Title   string
		Message string
		MesType string
	}
)

func (v *Views) getSessionData(eC echo.Context) *Context {
	session, err := v.cookie.Get(eC.Request(), v.conf.SessionCookieName)
	if err != nil {
		log.Printf("error getting session: %+v", err)
		return nil
	}
	userValue := session.Values["user"]
	u, ok := userValue.(user.User)
	if !ok {
		u = user.User{Authenticated: false}
	}
	internalValue := session.Values["internalContext"]
	i, ok := internalValue.(InternalContext)
	if !ok {
		i = InternalContext{}
	}
	c := &Context{
		Title:    i.Title,
		Message:  i.Message,
		MsgType:  i.MesType,
		Callback: "/internal",
		User:     u,
		Version:  v.conf.Version,
	}
	return c
}

func (v *Views) setMessagesInSession(eC echo.Context, c *Context) error {
	session, err := v.cookie.Get(eC.Request(), v.conf.SessionCookieName)
	if err != nil {
		return fmt.Errorf("error getting session: %w", err)
	}
	session.Values["internalContext"] = InternalContext{
		Title:   c.Title,
		Message: c.Message,
		MesType: c.MsgType,
	}

	err = session.Save(eC.Request(), eC.Response())
	if err != nil {
		return fmt.Errorf("failed to save session for set message: %w", err)
	}
	return nil
}

func (v *Views) clearMessagesInSession(eC echo.Context) error {
	session, err := v.cookie.Get(eC.Request(), v.conf.SessionCookieName)
	if err != nil {
		return fmt.Errorf("error getting session: %w", err)
	}
	session.Values["internalContext"] = InternalContext{}

	err = session.Save(eC.Request(), eC.Response())
	if err != nil {
		return fmt.Errorf("failed to save session for clear message: %w", err)
	}
	return nil
}

// DBUsersToUsersTemplateFormat converts from the DB layer type to the user template type
func DBUsersToUsersTemplateFormat(dbUsers []user.User) []user.StrippedUser {
	tplUsers := make([]user.StrippedUser, 0, len(dbUsers))
	for _, dbUser := range dbUsers {
		var strippedUser user.StrippedUser
		strippedUser.UserID = dbUser.UserID
		strippedUser.Username = dbUser.Username
		if dbUser.Firstname != dbUser.Nickname {
			strippedUser.Name = fmt.Sprintf("%s (%s) %s", dbUser.Firstname, dbUser.Nickname, dbUser.Lastname)
		} else {
			strippedUser.Name = fmt.Sprintf("%s %s", dbUser.Firstname, dbUser.Lastname)
		}
		strippedUser.Email = dbUser.Email
		strippedUser.Enabled = dbUser.Enabled
		if dbUser.DeletedAt.Valid || dbUser.DeletedBy.Valid {
			strippedUser.Deleted = true
		} else {
			strippedUser.Deleted = false
		}
		if dbUser.LastLogin.Valid {
			strippedUser.LastLogin = dbUser.LastLogin.Time.Format("2006-01-02 15:04:05")
		} else {
			strippedUser.LastLogin = "-"
		}
		tplUsers = append(tplUsers, strippedUser)
	}
	return tplUsers
}

func DBUserToUserTemplateFormat(dbUser user.User, store *user.Store) user.DetailedUser {
	var u user.DetailedUser
	var err error
	u.UserID = dbUser.UserID
	u.Username = dbUser.Username
	u.UniversityUsername = dbUser.UniversityUsername
	u.LDAPUsername = dbUser.LDAPUsername
	u.LoginType = dbUser.LoginType
	u.Nickname = dbUser.Nickname
	u.Firstname = dbUser.Firstname
	u.Lastname = dbUser.Lastname
	u.Email = dbUser.Email
	u.LastLogin = null.NewString(dbUser.LastLogin.Time.Format("2006-01-02 15:04:05"), dbUser.LastLogin.Valid)
	u.ResetPw = dbUser.ResetPw
	u.Enabled = dbUser.Enabled
	u.CreatedAt = null.StringFrom(dbUser.CreatedAt.Time.Format("2006-01-02 15:04:05"))
	if dbUser.CreatedBy.Valid {
		u.CreatedBy, err = store.GetUser(context.Background(), user.User{UserID: int(dbUser.CreatedBy.Int64)})
		if err != nil {
			log.Println(err)
			u.CreatedBy = user.User{
				UserID:    int(dbUser.CreatedBy.Int64),
				Firstname: "",
				Nickname:  "",
				Lastname:  "",
			}
		}
	} else {
		u.CreatedBy = user.User{
			UserID:    -1,
			Firstname: "",
			Nickname:  "",
			Lastname:  "",
		}
	}
	if dbUser.UpdatedAt.Valid {
		u.UpdatedAt = null.StringFrom(dbUser.UpdatedAt.Time.Format("2006-01-02 15:04:05"))
	} else {
		u.UpdatedAt = null.NewString("", false)
	}
	if dbUser.UpdatedBy.Valid {
		u.UpdatedBy, err = store.GetUser(context.Background(), user.User{UserID: int(dbUser.UpdatedBy.Int64)})
		if err != nil {
			log.Println(err)
			u.UpdatedBy = user.User{
				UserID:    int(dbUser.UpdatedBy.Int64),
				Firstname: "",
				Nickname:  "",
				Lastname:  "",
			}
		}
	} else {
		u.UpdatedBy = user.User{
			UserID:    -1,
			Firstname: "",
			Nickname:  "",
			Lastname:  "",
		}
	}
	if dbUser.DeletedAt.Valid {
		u.DeletedAt = null.StringFrom(dbUser.DeletedAt.Time.Format("2006-01-02 15:04:05"))
	} else {
		u.DeletedAt = null.NewString("", false)
	}
	if dbUser.DeletedBy.Valid {
		u.DeletedBy, err = store.GetUser(context.Background(), user.User{UserID: int(dbUser.DeletedBy.Int64)})
		if err != nil {
			log.Println(err)
			u.DeletedBy = user.User{
				UserID:    int(dbUser.DeletedBy.Int64),
				Firstname: "",
				Nickname:  "",
				Lastname:  "",
			}
		}
	} else {
		u.DeletedBy = user.User{
			UserID:    -1,
			Firstname: "",
			Nickname:  "",
			Lastname:  "",
		}
	}
	if dbUser.UseGravatar {
		u.UseGravatar = true
		// #nosec
		hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(u.Email))))
		u.Avatar = fmt.Sprintf("https://www.gravatar.com/avatar/%s", hex.EncodeToString(hash[:]))
	} else {
		u.UseGravatar = false
		if len(dbUser.Avatar) == 0 {
			u.Avatar = "https://placehold.it/128x128"
		} else {
			u.Avatar = fmt.Sprintf("/avatar/%s", dbUser.Avatar)
		}
	}
	return u
}

func (v *Views) removeDuplicate(strSlice []permission.Permission) []permission.Permission {
	allKeys := make(map[int]bool)
	var list []permission.Permission
	for _, item := range strSlice {
		if _, value := allKeys[item.PermissionID]; !value {
			allKeys[item.PermissionID] = true
			list = append(list, item)
		}
	}
	return list
}
