package views

import (
	"context"
	// #nosec
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	// importing time zones in case the system doesn't have them
	_ "time/tzdata"

	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/user"
	"gopkg.in/guregu/null.v4"
)

type (
	// Context is a struct that is applied to the templates.
	Context struct {
		// TitleText is used for sending pages to the user with custom titles
		TitleText string
		// Message is used for sending a message back to the user trying to log in, might decide to move later as it may not be needed
		Message string
		// MsgType is the bulma.io class used to indicate what should be displayed
		MsgType string
		// Callback is the address to redirect the user to
		Callback string
		// User is the stored logged-in user
		User user.User
		// Version is the version that is running
		Version    string
		Assumed    bool
		actualUser user.User
	}

	InternalContext struct {
		TitleText string
		Message   string
		MesType   string
	}
)

func (v *Views) getSessionData(eC echo.Context) *Context {
	session, err := v.cookie.Get(eC.Request(), v.conf.SessionCookieName)
	if err != nil {
		log.Printf("error getting session: %+v", err)
		err = session.Save(eC.Request(), eC.Response())
		if err != nil {
			panic(fmt.Errorf("failed to save user session for getSessionData: %w", err))
		}
		i := InternalContext{}
		c := &Context{
			TitleText: i.TitleText,
			Message:   i.Message,
			MsgType:   i.MesType,
			Callback:  "/internal",
			Version:   v.conf.Version,
		}
		return c
	}

	var u, actual user.User
	userValue := session.Values["user"]
	u, ok := userValue.(user.User)
	if !ok {
		u = user.User{Authenticated: false}
	}
	var assumed bool
	if u.AssumedUser != nil {
		assumed = true
		actual = u
		u = *u.AssumedUser
	}

	internalValue := session.Values["internalContext"]
	i, ok := internalValue.(InternalContext)
	if !ok {
		i = InternalContext{}
	}
	c := &Context{
		TitleText:  i.TitleText,
		Message:    i.Message,
		MsgType:    i.MesType,
		Callback:   "/internal",
		User:       u,
		Version:    v.conf.Version,
		Assumed:    assumed,
		actualUser: actual,
	}
	return c
}

func (v *Views) setMessagesInSession(eC echo.Context, c *Context) error {
	session, err := v.cookie.Get(eC.Request(), v.conf.SessionCookieName)
	if err != nil {
		return fmt.Errorf("error getting session: %w", err)
	}
	session.Values["internalContext"] = InternalContext{
		TitleText: c.TitleText,
		Message:   c.Message,
		MesType:   c.MsgType,
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

// DBUserToDetailedUser handles all the little details for the users front end
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

// removeDuplicates removes all duplicate permissions
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

// minRequirementsMet tests if the password meets the minimum requirements
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
