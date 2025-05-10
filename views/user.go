package views

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/ystv/web-auth/infrastructure/mail"
	"github.com/ystv/web-auth/infrastructure/permission"
	"github.com/ystv/web-auth/officership"
	"github.com/ystv/web-auth/permission/permissions"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
	"github.com/ystv/web-auth/utils"
)

type (
	// UsersTemplate represents the context for the user template
	UsersTemplate struct {
		Users    []user.StrippedUser
		CurPage  int
		NextPage int
		PrevPage int
		LastPage int
		Sort     Sort
		TemplateHelper
	}

	// Sort is the parameters for how to sort a users request
	Sort struct {
		Pages      int
		Size       int
		PageNumber int
		Column     string
		Direction  string
		Search     string
		Enabled    string
		Deleted    string
	}

	// UserTemplate is for the user front end
	UserTemplate struct {
		User user.DetailedUser
		TemplateHelper
	}
)

// UsersFunc handles a users request
func (v *Views) UsersFunc(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodGet:
		return v._usersGet(c)
	case http.MethodPost:
		return v._usersPost(c)
	}

	return v.invalidMethodUsed(c)
}

//gocyclo:ignore
func (v *Views) _usersGet(c echo.Context) error {
	c1 := v.getSessionData(c)
	column := c.QueryParam("column")
	direction := c.QueryParam("direction")
	search := c.QueryParam("search")

	search, err := url.QueryUnescape(search)
	if err != nil {
		return fmt.Errorf("failed to unescape query: %w", err)
	}

	enabled := c.QueryParam("enabled")
	deleted := c.QueryParam("deleted")

	var size, page int

	sizeRaw := c.QueryParam("size")

	if sizeRaw == "all" {
		size = 0
	} else if len(sizeRaw) != 0 {
		page, err = strconv.Atoi(c.QueryParam("page"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest,
				fmt.Errorf("unable to parse page for users: %w", err))
		}

		size, err = strconv.Atoi(sizeRaw)
		//nolint:gocritic
		if err != nil {
			size = 0
		} else if size <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, errors.New("invalid size, must be positive"))
		} else if size != 5 && size != 10 && size != 25 && size != 50 && size != 75 && size != 100 {
			size = 0
		}
	}

	switch column {
	case "userId", "name", "username", "email", "lastLogin":
		switch direction {
		case "asc":
		case "desc":
			break
		default:
			column = ""
			direction = ""
		}
	default:
		column = ""
		direction = ""
	}

	dbUsers, fullCount, err := v.user.GetUsers(c.Request().Context(), size, page, search, column, direction, enabled,
		deleted)
	if err != nil {
		return fmt.Errorf("failed to get users for users: %w", err)
	}

	if (len(dbUsers) == 0 || fullCount == 0) && size != 0 && page != 0 {
		return echo.NewHTTPError(http.StatusBadRequest, errors.New("size and page given is not valid"))
	}

	tplUsers := DBUsersToUsersTemplateFormat(dbUsers)

	var sum int

	if size == 0 {
		sum = 0
	} else {
		sum = int(math.Ceil(float64(fullCount) / float64(size)))
	}

	if page <= 0 {
		page = 25
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		return fmt.Errorf("failed to get user permissions for users: %w", err)
	}

	data := UsersTemplate{
		Users: tplUsers,
		TemplateHelper: TemplateHelper{
			UserPermissions: p1,
			ActivePage:      "users",
			Assumed:         c1.Assumed,
		},
		Sort: Sort{
			Pages:      sum,
			Size:       size,
			PageNumber: page,
			Column:     column,
			Direction:  direction,
			Search:     search,
			Enabled:    enabled,
			Deleted:    deleted,
		},
	}

	return v.template.RenderTemplate(c.Response(), data, templates.UsersTemplate, templates.PaginationType)
}

//gocyclo:ignore
func (v *Views) _usersPost(c echo.Context) error {
	u, err := url.Parse("/internal/users")
	if err != nil {
		panic(fmt.Errorf("invalid url: %w", err)) // this panics because if this errors then many other things will be wrong
	}

	q := u.Query()

	column := c.FormValue("column")
	direction := c.FormValue("direction")
	search := c.FormValue("search")
	enabled := c.FormValue("enabled")
	deleted := c.FormValue("deleted")

	var size int

	sizeRaw := c.FormValue("size")

	if sizeRaw == "all" {
		size = 0
	} else {
		size, err = strconv.Atoi(sizeRaw)
		//nolint:gocritic
		if err != nil {
			size = 0
		} else if size <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, errors.New("invalid size, must be positive"))
		} else if size != 5 && size != 10 && size != 25 && size != 50 && size != 75 && size != 100 {
			size = 25
		}
	}

	if enabled == "enabled" || enabled == "disabled" {
		q.Set("enabled", enabled)
	} else if enabled != "any" {
		return echo.NewHTTPError(http.StatusBadRequest,
			errors.New("enabled must be set to either \"any\", \"enabled\" or \"disabled\""))
	}

	if deleted == "deleted" || deleted == "not_deleted" {
		q.Set("deleted", deleted)
	} else if deleted != "any" {
		return echo.NewHTTPError(http.StatusBadRequest,
			errors.New("deleted must be set to either \"any\", \"deleted\" or \"not_deleted\""))
	}

	if column == "userId" || column == "name" || column == "username" || column == "email" || column == "lastLogin" {
		if direction == "asc" || direction == "desc" {
			q.Set("column", column)
			q.Set("direction", direction)
		}
	}

	c.Request().Method = "GET"

	if size > 0 {
		q.Set("size", strconv.FormatInt(int64(size), 10))
		q.Set("page", "1")
	}

	if len(search) > 0 {
		q.Set("search", url.QueryEscape(search))
	}

	u.RawQuery = q.Encode()

	return c.Redirect(http.StatusFound, u.String())
}

// UserFunc handles a users request
func (v *Views) UserFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	userID, err := strconv.Atoi(c.Param("userid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse userid for user: %w", err))
	}

	userFromDB, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
	if err != nil {
		return fmt.Errorf("failed to get user for user: %w", err)
	}

	officers, err := v.officership.GetOfficershipMembers(c.Request().Context(), nil, &userFromDB, officership.Any, officership.Any, false)
	if err != nil {
		return fmt.Errorf("failed to get officers for user: %w", err)
	}

	detailedUser := DBUserToDetailedUser(userFromDB, v.user, officers)

	detailedUser.Permissions, err = v.user.GetPermissionsForUser(c.Request().Context(),
		user.User{UserID: detailedUser.UserID})
	if err != nil {
		return fmt.Errorf("failed to get permissions for user: %w", err)
	}

	detailedUser.Permissions = removeDuplicate(detailedUser.Permissions)

	detailedUser.Roles, err = v.user.GetRolesForUser(c.Request().Context(), user.User{UserID: detailedUser.UserID})
	if err != nil {
		return fmt.Errorf("failed to get roles for user: %w", err)
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		return fmt.Errorf("failed to get user permissions for user: %w", err)
	}

	data := UserTemplate{
		User: detailedUser,
		TemplateHelper: TemplateHelper{
			UserPermissions: p1,
			ActivePage:      "user",
			Assumed:         c1.Assumed,
		},
	}

	return v.template.RenderTemplate(c.Response(), data, templates.UserTemplate, templates.RegularType)
}

func (v *Views) AssumeUserFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			return fmt.Errorf("failed to get session for assume user: %w", err)
		}

		c1 := v.getSessionData(c)

		if c1.Assumed {
			return c.Redirect(http.StatusFound, "/internal")
		}

		userID, err := strconv.Atoi(c.Param("userid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse userid for user: %w", err))
		}

		userFromDB, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
		if err != nil {
			return fmt.Errorf("failed to get user for user: %w", err)
		}

		userFromDB.Authenticated = true

		userFromDB.LastLogin = null.TimeFrom(time.Now())

		c1.User.AssumedUser = &userFromDB

		session.Values["user"] = c1.User

		err = session.Save(c.Request(), c.Response())
		if err != nil {
			return fmt.Errorf("failed to save user session for assume: %w", err)
		}

		return c.Redirect(http.StatusFound, "/internal")
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) ReleaseUserFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			return fmt.Errorf("failed to get session for release user: %w", err)
		}

		c1 := v.getSessionData(c)

		if !c1.Assumed {
			return c.Redirect(http.StatusFound, "/internal")
		}

		var u user.User

		userValue := session.Values["user"]

		u, ok := userValue.(user.User)
		if !ok {
			u = user.User{Authenticated: false}
		}

		u.AssumedUser = nil

		session.Values["user"] = u

		err = session.Save(c.Request(), c.Response())
		if err != nil {
			return fmt.Errorf("failed to save user session for release: %w", err)
		}

		return c.Redirect(http.StatusFound, "/internal")
	}

	return v.invalidMethodUsed(c)
}

// UserAddFunc handles an add user request
func (v *Views) UserAddFunc(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodGet:
		return v._userAddGet(c)
	case http.MethodPost:
		return v._userAddPost(c)
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) _userAddGet(c echo.Context) error {
	c1 := v.getSessionData(c)

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		return fmt.Errorf("failed to get user permissions for user: %w", err)
	}

	data := TemplateHelper{
		UserPermissions: p1,
		ActivePage:      "useradd",
		Assumed:         c1.Assumed,
	}

	return v.template.RenderTemplate(c.Response(), data, templates.UserAddTemplate, templates.RegularType)
}

func (v *Views) _userAddPost(c echo.Context) error {
	c1 := v.getSessionData(c)

	err := c.Request().ParseForm()
	if err != nil {
		return fmt.Errorf("failed to parse form for userAdd: %w", err)
	}

	firstName := c.FormValue("firstname")
	lastName := c.FormValue("lastname")
	username := c.FormValue("username")
	universityUsername := c.FormValue("universityusername")
	email := c.FormValue("email")
	tempDisableSendEmail := c.FormValue("disablesendemail")
	var sendEmail = true

	if tempDisableSendEmail == "on" && func() bool {
		m := permission.SufficientPermissionsFor(permissions.SuperUser)

		for _, perm := range c1.User.Permissions {
			if m[perm.Name] {
				return true
			}
		}

		return false
	}() {
		sendEmail = false
	}

	password, err := utils.GenerateRandom(utils.GeneratePassword)
	if err != nil {
		return fmt.Errorf("error generating password: %w", err)
	}

	salt, err := utils.GenerateRandom(utils.GenerateSalt)
	if err != nil {
		return fmt.Errorf("error generating salt: %w", err)
	}

	u := user.User{
		UserID:             0,
		Username:           username,
		UniversityUsername: universityUsername,
		LoginType:          "internal",
		Firstname:          firstName,
		Nickname:           firstName,
		Lastname:           lastName,
		Password:           null.StringFrom(password),
		Salt:               null.StringFrom(salt),
		Email:              email,
		ResetPw:            true,
		Enabled:            true,
	}

	_, err = v.user.AddUser(c.Request().Context(), u, c1.User.UserID)
	if err != nil {
		return fmt.Errorf("failed to add user for addUser: %w", err)
	}

	var message struct {
		Message string `json:"message"`
		Error   error  `json:"error"`
	}

	mailer := v.mailer.ConnectMailer()

	if mailer != nil && sendEmail {
		var tmpl *template.Template

		tmpl, err = v.template.GetEmailTemplate(templates.SignupEmailTemplate)
		if err != nil {
			return fmt.Errorf("failed to get email in addUser: %w", err)
		}

		file := mail.Mail{
			Subject: "Welcome to YSTV!",
			Tpl:     tmpl,
			To:      u.Email,
			From:    "YSTV No-Reply <no-reply@ystv.co.uk>",
			TplData: struct {
				Name     string
				Username string
				Password string
			}{
				Name:     firstName,
				Username: username,
				Password: password,
			},
		}

		err = mailer.SendMail(file)
		if err != nil {
			return fmt.Errorf("failed to send email in addUser: %w", err)
		}

		message.Message = fmt.Sprintf("Successfully sent user email to: \"%s\"", email)
	} else {
		message.Message = fmt.Sprintf(`No mailer present<br>Please send the username and password to this email: 
%s, username: %s, password: %s`, email, username, password)
		message.Error = errors.New("no mailer present")
		log.Printf("no Mailer present")
	}

	log.Printf("created user: %s", u.Username)

	var status int

	return c.JSON(status, message)
}

// UserEditFunc handles an edit user request
func (v *Views) UserEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		userID, err := strconv.Atoi(c.Param("userid"))
		if err != nil {
			return fmt.Errorf("failed to get userid for editUser: %w", err)
		}

		user1, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
		if err != nil {
			return fmt.Errorf("failed to get user for editUser: %w", err)
		}

		err = c.Request().ParseForm()
		if err != nil {
			return fmt.Errorf("failed to parse form for userEdit: %w", err)
		}

		firstName := c.FormValue("firstname")
		nickname := c.FormValue("nickname")
		lastName := c.FormValue("lastname")
		username := c.FormValue("username")
		universityUsername := c.FormValue("universityusername")
		LDAPUsername := c.FormValue("ldapusername")
		email := c.FormValue("email")
		// login type can't be changed yet but the infrastructure is in
		loginType := c.FormValue("logintype")
		_ = loginType

		if len(firstName) > 0 {
			user1.Firstname = firstName
		}

		if len(nickname) > 0 {
			user1.Nickname = nickname
		}

		if len(lastName) > 0 {
			user1.Lastname = lastName
		}

		if len(username) > 0 {
			user1.Username = username
		}

		if len(universityUsername) > 0 {
			user1.UniversityUsername = universityUsername
		}

		if len(LDAPUsername) > 0 {
			user1.LDAPUsername = null.StringFrom(LDAPUsername)
		}

		if len(email) > 0 {
			user1.Email = email
		}

		err = v.user.EditUser(c.Request().Context(), user1, c1.User.UserID)
		if err != nil {
			return fmt.Errorf("failed to edit user for editUser: %w", err)
		}

		return c.Redirect(http.StatusFound, fmt.Sprintf("/internal/user/%d", userID))
	}

	return v.invalidMethodUsed(c)
}

// UserToggleEnabledFunc handles an toggle enable user request
func (v *Views) UserToggleEnabledFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		userID, err := strconv.Atoi(c.Param("userid"))
		if err != nil {
			return fmt.Errorf("failed to get userid for toggleUser: %w", err)
		}

		user1, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
		if err != nil {
			return fmt.Errorf("failed to get user for toggleUser: %w", err)
		}

		user1.Enabled = !user1.Enabled

		err = v.user.EditUser(c.Request().Context(), user1, c1.User.UserID)
		if err != nil {
			return fmt.Errorf("failed to edit user for toggleUser: %w", err)
		}

		return c.Redirect(http.StatusFound, fmt.Sprintf("/internal/user/%d", userID))
	}

	return v.invalidMethodUsed(c)
}

// UserDeleteFunc handles an delete user request
func (v *Views) UserDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		c1 := v.getSessionData(c)

		userID, err := strconv.Atoi(c.Param("userid"))
		if err != nil {
			return fmt.Errorf("failed to get userid for deleteUser: %w", err)
		}

		user1, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
		if err != nil {
			return fmt.Errorf("failed to get user for deleteUser: %w", err)
		}

		err = v.user.RemoveUserForRoles(c.Request().Context(), user1)
		if err != nil {
			return fmt.Errorf("failed to delete roleUsers for deleteUser: %w", err)
		}

		err = v.user.DeleteUser(c.Request().Context(), user1, c1.User.UserID)
		if err != nil {
			return fmt.Errorf("failed to delete user for deleteUser: %w", err)
		}

		return c.Redirect(http.StatusFound, "/internal/users")
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) UploadAvatarUserFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		c1 := v.getSessionData(c)

		userID, err := strconv.Atoi(c.Param("userid"))
		if err != nil {
			return fmt.Errorf("failed to get userid for deleteUser: %w", err)
		}

		user1, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
		if err != nil {
			return fmt.Errorf("failed to get user for deleteUser: %w", err)
		}

		data := struct {
			Error string `json:"error"`
		}{}

		useGravatarTemp := c.FormValue("useGravatar")
		useGravatar := false

		if useGravatarTemp == "on" {
			useGravatar = true
		}

		if !useGravatar {
			var file *multipart.FileHeader
			file, err = c.FormFile("upload")
			if err != nil {
				// return fmt.Errorf("failed to get file for uploadAvatar: %w", err)
				log.Printf("failed to get file for uploadAvatar, user id: %d, error: %+v", user1.UserID, err)
				data.Error = fmt.Sprintf("failed to get file for uploadAvatar: %+v", err)
				return c.JSON(http.StatusOK, data)
			}

			var fileName string
			var fileBytes []byte
			fileName, fileBytes, err = v.fileUpload(file)
			if err != nil {
				// return fmt.Errorf("failed to upload file for uploadAvatar: %w", err)
				log.Printf("failed to upload file for uploadAvatar, user id: %d, error: %+v", user1.UserID, err)
				data.Error = fmt.Sprintf("failed to upload file for uploadAvatar: %+v", err)
				return c.JSON(http.StatusOK, data)
			}

			buf := bytes.NewReader(fileBytes)

			// This uploads the contents of the buffer to S3
			_, err = v.cdn.PutObjectWithContext(c.Request().Context(), &s3.PutObjectInput{
				Bucket: aws.String("avatars"),
				Key:    aws.String(fileName),
				Body:   buf,
			})
			if err != nil {
				// return fmt.Errorf("failed to upload file to cdn for uploadAvatar: %w", err)
				log.Printf("failed to upload file to cdn for uploadAvatar, user id: %d, error: %+v", user1.UserID, err)
				data.Error = fmt.Sprintf("failed to upload file to cdn for uploadAvatar: %+v", err)
				return c.JSON(http.StatusOK, data)
			}

			user1.Avatar = fmt.Sprintf("%s/avatars/%s", v.conf.CDNEndpoint, fileName)
		}

		user1.UseGravatar = useGravatar

		err = v.user.EditUserAvatarUser(c.Request().Context(), user1, c1.User.UserID)
		if err != nil {
			// return fmt.Errorf("failed to edit user for uploadAvatar: %w", err)
			log.Printf("failed to edit user for uploadAvatar, user id: %d, error: %+v", user1.UserID, err)
			data.Error = fmt.Sprintf("failed to edit user for uploadAvatar: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = "successfully uploaded avatar"
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for uploadAvatar, user id: %d, error: %+v", c1.User.UserID, err)
		}

		err = session.Save(c.Request(), c.Response())
		if err != nil {
			return fmt.Errorf("failed to save user session in settings: %w", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
}

func (v *Views) RemoveAvatarUserFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		c1 := v.getSessionData(c)

		userID, err := strconv.Atoi(c.Param("userid"))
		if err != nil {
			return fmt.Errorf("failed to get userid for deleteUser: %w", err)
		}

		user1, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
		if err != nil {
			return fmt.Errorf("failed to get user for deleteUser: %w", err)
		}

		data := struct {
			Error string `json:"error"`
		}{}

		if len(user1.Avatar) > 0 && strings.Contains(user1.Avatar, v.conf.CDNEndpoint) {
			split := strings.Split(user1.Avatar, "/")
			key := split[len(split)-1]
			_, err = v.cdn.DeleteObject(&s3.DeleteObjectInput{
				Bucket: aws.String("avatars"),
				Key:    aws.String(key),
			})
			if err != nil {
				// return fmt.Errorf("failed to delete file from cdn for removeAvatar: %w", err)
				log.Printf("failed to delete file from cdn for removeAvatar, user id: %d, error: %+v", c1.User.UserID, err)
				data.Error = fmt.Sprintf("failed to delete file from cdn for removeAvatar: %+v", err)
				return c.JSON(http.StatusOK, data)
			}
		}

		user1.Avatar = ""

		err = v.user.EditUserAvatarUser(c.Request().Context(), user1, c1.User.UserID)
		if err != nil {
			// return fmt.Errorf("failed to edit user for removeAvatar: %w", err)
			log.Printf("failed to edit user for removeAvatar, user id: %d, error: %+v", c1.User.UserID, err)
			data.Error = fmt.Sprintf("failed to edit user for removeAvatar: %+v", err)
			return c.JSON(http.StatusOK, data)
		}

		c1.Message = "successfully removed image"
		c1.MsgType = "is-success"
		err = v.setMessagesInSession(c, c1)
		if err != nil {
			log.Printf("failed to set data for removedAvatar, user id: %d, error: %+v", c1.User.UserID, err)
		}

		err = session.Save(c.Request(), c.Response())
		if err != nil {
			return fmt.Errorf("failed to save user session in settings: %w", err)
		}

		return c.JSON(http.StatusOK, data)
	}
	return v.invalidMethodUsed(c)
}
