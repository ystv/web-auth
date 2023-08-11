package views

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/user"
	"gopkg.in/guregu/null.v4"
	"log"
	"regexp"
	"strings"
	"time"
)

type (
	// Context is a struct that is applied to the templates.
	Context struct {
		Message  string
		MsgType  string
		Version  string
		Callback string
		User     user.User
	}
)

func (v *Views) getData(s *sessions.Session) *Context {
	val := s.Values["user"]
	var u user.User
	u, ok := val.(user.User)
	if !ok {
		u = user.User{Authenticated: false}
	}
	c := Context{
		Version:  v.conf.Version,
		Callback: "/internal",
		User:     u,
	}
	return &c
}

// DBToTemplateType converts from the DB layer type to the user template type
func DBToTemplateType(dbUsers []user.User) []user.StrippedUser {
	var tplUsers []user.StrippedUser
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		fmt.Println(err)
	}
	for _, dbUser := range dbUsers {
		var user1 user.StrippedUser
		user1.UserID = dbUser.UserID
		user1.Username = dbUser.Username
		if dbUser.Firstname != dbUser.Nickname {
			user1.Name = fmt.Sprintf("%s (%s) %s", dbUser.Firstname, dbUser.Nickname, dbUser.Lastname)
		} else {
			user1.Name = fmt.Sprintf("%s %s", dbUser.Firstname, dbUser.Lastname)
		}
		user1.Email = dbUser.Email
		user1.Enabled = dbUser.Enabled
		if dbUser.DeletedAt.Valid || dbUser.DeletedBy.Valid {
			user1.Deleted = true
		} else {
			user1.Deleted = false
		}
		if dbUser.LastLogin.Valid {
			user1.LastLogin = dbUser.LastLogin.Time.In(location).Format("2006-01-02 15:04:05 MST")
		} else {
			user1.LastLogin = "-"
		}
		tplUsers = append(tplUsers, user1)
	}
	return tplUsers
}

func DBUserToDetailedUser(dbUser user.User, store *user.Store) user.DetailedUser {
	var u user.DetailedUser
	var err error
	location, err := time.LoadLocation("Europe/London")
	if err != nil {
		fmt.Println(err)
	}
	u.UserID = dbUser.UserID
	u.Username = dbUser.Username
	u.UniversityUsername = dbUser.UniversityUsername
	u.LDAPUsername = dbUser.LDAPUsername
	u.LoginType = dbUser.LoginType
	u.Nickname = dbUser.Nickname
	u.Firstname = dbUser.Firstname
	u.Lastname = dbUser.Lastname
	u.Email = dbUser.Email
	u.LastLogin = null.NewString(dbUser.LastLogin.Time.In(location).Format("2006-01-02 15:04:05 MST"), dbUser.LastLogin.Valid)
	u.ResetPw = dbUser.ResetPw
	u.Enabled = dbUser.Enabled
	u.CreatedAt = null.StringFrom(dbUser.CreatedAt.Time.In(location).Format("2006-01-02 15:04:05 MST"))
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
		u.UpdatedAt = null.StringFrom(dbUser.UpdatedAt.Time.In(location).Format("2006-01-02 15:04:05 MST"))
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
		u.DeletedAt = null.StringFrom(dbUser.DeletedAt.Time.In(location).Format("2006-01-02 15:04:05 MST"))
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
		hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(dbUser.Email))))
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

func removeDuplicate(strSlice []permission.Permission) []permission.Permission {
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

func minRequirementsMet(password string) (errString string) {
	var match bool
	match, err := regexp.MatchString("^.*[a-z].*$", password)
	if err != nil || !match {
		errString = "password must contain at least 1 lower case letter"
	}
	match, err = regexp.MatchString("^.*[A-Z].*$", password)
	if err != nil || !match {
		if len(errString) > 0 {
			errString += " and password must contain at least 1 upper case letter"
		} else {
			errString = "password must contain at least 1 upper case letter"
		}
	}
	match, err = regexp.MatchString("^.*\\d.*$", password)
	if err != nil || !match {
		if len(errString) > 0 {
			errString += " and password must contain at least 1 number"
		} else {
			errString = "password must contain at least 1 number"
		}
	}
	match, err = regexp.MatchString("^.*[@$!%*?&|^£;:/.,<>()_=+~§±#{}-].*$", password)
	if err != nil || !match {
		if len(errString) > 0 {
			errString += " and password must contain at least 1 special character"
		} else {
			errString = "password must contain at least 1 special character"
		}
	}
	if len(password) <= 8 {
		if len(errString) > 0 {
			errString += " and password must be at least 8 characters long"
		} else {
			errString = "password must be at least 8 characters long"
		}
	}
	return errString
}
