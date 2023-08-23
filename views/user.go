package views

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/infrastructure/mail"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
	"github.com/ystv/web-auth/utils"
	"gopkg.in/guregu/null.v4"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
)

type (
	// UsersTemplate represents the context for the user template
	UsersTemplate struct {
		Users      []user.StrippedUser
		UserID     int
		CurPage    int
		NextPage   int
		PrevPage   int
		LastPage   int
		ActivePage string
		Sort       Sort
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
		User       user.DetailedUser
		UserID     int
		ActivePage string
	}
)

// UsersFunc handles a users request
func (v *Views) UsersFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	var err error

	if c.Request().Method == "POST" {
		err = c.Request().ParseForm()
		if err != nil {
			return v.errorHandle(c, err)
		}

		u, err := url.Parse("/internal/users")
		if err != nil {
			return v.errorHandle(c, fmt.Errorf("invlaid url: %w", err))
		}

		q := u.Query()

		column := c.Request().FormValue("column")
		direction := c.Request().FormValue("direction")
		search := c.Request().FormValue("search")

		var size int
		sizeRaw := c.Request().FormValue("size")
		if sizeRaw == "all" {
			size = 0
		} else {
			size, err = strconv.Atoi(sizeRaw)
			if err != nil {
				size = 0
			} else if size <= 0 {
				return v.errorHandle(c, fmt.Errorf("invalid size, must be positive"))
			} else if size != 5 && size != 10 && size != 25 && size != 50 && size != 75 && size != 100 {
				size = 25
			}
		}

		enabled := c.Request().FormValue("enabled")
		if enabled == "enabled" || enabled == "disabled" {
			q.Set("enabled", enabled)
		} else if enabled != "any" {
			return v.errorHandle(c, fmt.Errorf("enabled must be set to either \"any\", \"enabled\" or \"disabled\""))
		}

		deleted := c.Request().FormValue("deleted")
		if deleted == "deleted" || deleted == "not_deleted" {
			q.Set("deleted", deleted)
		} else if deleted != "any" {
			return v.errorHandle(c, fmt.Errorf("deleted must be set to either \"any\", \"deleted\" or \"not_deleted\""))
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

	column := c.Request().URL.Query().Get("column")
	direction := c.Request().URL.Query().Get("direction")
	search := c.Request().URL.Query().Get("search")
	if len(search) > 0 {
		search, err = url.QueryUnescape(search)
		if err != nil {
			log.Printf("failed to parse search in users: %+v", err)
		}
	}
	enabled := c.Request().URL.Query().Get("enabled")
	deleted := c.Request().URL.Query().Get("deleted")
	var size, page, count int
	sizeRaw := c.Request().URL.Query().Get("size")
	if sizeRaw == "all" {
		size = 0
	} else if len(sizeRaw) != 0 {
		page, err = strconv.Atoi(c.Request().URL.Query().Get("page"))
		if err != nil {
			page = 1
			log.Println(err)
			return v.errorHandle(c, err)
		}
		size, err = strconv.Atoi(sizeRaw)
		if err != nil {
			size = 0
		} else if size <= 0 {
			err = v.errorHandle(c, fmt.Errorf("invalid size, must be positive"))
		} else if size != 5 && size != 10 && size != 25 && size != 50 && size != 75 && size != 100 {
			size = 0
		}

		count, err = v.user.CountUsers(c.Request().Context())
		if err != nil {
			log.Println(err)
			return v.errorHandle(c, err)
		}

		if count <= size*(page-1) {
			log.Println("size and page given is not valid")
			return v.errorHandle(c, fmt.Errorf("size and page given is not valid"))
		}
	}

	switch column {
	case "userId":
	case "name":
	case "username":
	case "email":
	case "lastLogin":
		switch direction {
		case "asc":
		case "desc":
			break
		default:
			column = ""
			direction = ""
		}
		break
	default:
		column = ""
		direction = ""
	}
	var dbUsers []user.User

	sort := len(column) > 0 && len(direction) > 0
	searchBool := len(search) > 0

	if sort && searchBool {
		dbUsers, err = v.user.GetUsersSearchOrder(c.Request().Context(), size, page, search, column, direction, enabled, deleted)
	} else if sort && !searchBool {
		dbUsers, err = v.user.GetUsersOrderNoSearch(c.Request().Context(), size, page, column, direction, enabled, deleted)
	} else if !sort && searchBool {
		dbUsers, err = v.user.GetUsersSearchNoOrder(c.Request().Context(), size, page, search, enabled, deleted)
	} else {
		dbUsers, err = v.user.GetUsers(c.Request().Context(), size, page, enabled, deleted)
	}

	if err != nil {
		log.Println(err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}
	tplUsers := DBToTemplateType(dbUsers)

	var sum int

	if size == 0 {
		sum = 0
	} else {
		sum = int(math.Ceil(float64(count) / float64(size)))
	}

	if page <= 0 {
		page = 25
	}

	data := UsersTemplate{
		Users:      tplUsers,
		UserID:     c1.User.UserID,
		ActivePage: "users",
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
	return v.template.RenderTemplatePagination(c.Response(), data, templates.UsersTemplate)
}

// UserFunc handles a users request
func (v *Views) UserFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	userID, err := strconv.Atoi(c.Param("userid"))
	if err != nil {
		//http.Error(c.Response(), err.Error(), http.StatusBadRequest)
		return v.errorHandle(c, err)
	}
	user1, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
	if err != nil {
		log.Printf("failed to get user in user: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	user2 := DBUserToDetailedUser(user1, v.user)

	user2.Permissions, err = v.user.GetPermissionsForUser(c.Request().Context(), user.User{UserID: user2.UserID})
	if err != nil {
		log.Println(err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	user2.Permissions = removeDuplicate(user2.Permissions)

	user2.Roles, err = v.user.GetRolesForUser(c.Request().Context(), user.User{UserID: user2.UserID})
	if err != nil {
		log.Println(err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	data := UserTemplate{
		User:       user2,
		UserID:     c1.User.UserID,
		ActivePage: "user",
	}

	return v.template.RenderTemplate(c.Response(), data, templates.UserTemplate)
}

// userFunc handles a users request internal
func (v *Views) userFunc(c echo.Context, userID int) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	user1, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
	if err != nil {
		log.Printf("failed to get user in user: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	user2 := DBUserToDetailedUser(user1, v.user)

	user2.Permissions, err = v.user.GetPermissionsForUser(c.Request().Context(), user.User{UserID: user2.UserID})
	if err != nil {
		log.Println(err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	user2.Permissions = removeDuplicate(user2.Permissions)

	user2.Roles, err = v.user.GetRolesForUser(c.Request().Context(), user.User{UserID: user2.UserID})
	if err != nil {
		log.Println(err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	data := UserTemplate{
		User:       user2,
		UserID:     c1.User.UserID,
		ActivePage: "user",
	}

	return v.template.RenderTemplate(c.Response(), data, templates.UserTemplate)
}

// UserAddFunc handles an add user request
func (v *Views) UserAddFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	if c.Request().Method == http.MethodGet {
		data := struct {
			UserID     int
			ActivePage string
		}{
			UserID:     c1.User.UserID,
			ActivePage: "useradd",
		}

		return v.template.RenderTemplate(c.Response(), data, templates.UserAddTemplate)
	} else if c.Request().Method == http.MethodPost {
		err := c.Request().ParseForm()
		if err != nil {
			return v.errorHandle(c, fmt.Errorf("failed to parse form for userAdd: %+v", err))
		}

		firstName := c.Request().FormValue("firstname")
		lastName := c.Request().FormValue("lastname")
		username := c.Request().FormValue("username")
		universityUsername := c.Request().FormValue("universityusername")
		email := c.Request().FormValue("email")

		password := utils.GeneratePassword()
		salt := utils.GenerateSalt()
		u := user.User{
			UserID:             0,
			Username:           username,
			UniversityUsername: null.StringFrom(universityUsername),
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
			log.Printf("failed to add user for addUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to add user for addUser: %+v", err))
			}
		}

		var message struct {
			Message string `json:"message"`
			Error   error  `json:"error"`
		}

		if v.Mailer.Enabled {
			v.Mailer, err = mail.NewMailer(mail.Config{
				Host:       v.conf.Mail.Host,
				Port:       v.conf.Mail.Port,
				Username:   v.conf.Mail.Username,
				Password:   v.conf.Mail.Password,
				DomainName: v.conf.DomainName,
			})
			if err != nil {
				log.Printf("Mailer failed: %+v", err)
			}

			file := mail.Mail{
				Subject: "Welcome to YSTV!",
				Tpl:     v.template.RenderEmail(templates.SignupEmailTemplate),
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

			err = v.Mailer.SendMail(file)
			if err != nil {
				log.Printf("failed to send email in addUser: %+v", err)
				return v.errorHandle(c, fmt.Errorf("failed to send email in addUser: %+v", err))
			}

			message.Message = fmt.Sprintf("Successfully sent user email to: \"%s\"", email)
		} else {
			message.Message = fmt.Sprintf("No mailer present\nPlease send the username and password to this email: %s, username: %s, password: %s", email, username, password)
			message.Error = fmt.Errorf("no mailer present")
			log.Printf("no Mailer present")
		}
		log.Printf("created user: %s", u.Username)

		var status int

		return c.JSON(status, message)
	} else {
		return v.errorHandle(c, fmt.Errorf("invalid method used"))
	}
}

// UserEditFunc handles an edit user request
func (v *Views) UserEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

		c1 := v.getData(session)

		userID, err := strconv.Atoi(c.Param("userid"))
		if err != nil {
			log.Printf("failed to get userid for toggleUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get userid for toggleUser: %+v", err))
			}
		}

		user1, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
		if err != nil {
			log.Printf("failed to get user for toggleUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get user for toggleUser: %+v", err))
			}
		}

		err = c.Request().ParseForm()
		if err != nil {
			return v.errorHandle(c, fmt.Errorf("failed to parse form for userEdit: %+v", err))
		}

		firstName := c.Request().FormValue("firstname")
		nickname := c.Request().FormValue("nickname")
		lastName := c.Request().FormValue("lastname")
		username := c.Request().FormValue("username")
		universityUsername := c.Request().FormValue("universityusername")
		LDAPUsername := c.Request().FormValue("ldapusername")
		email := c.Request().FormValue("email")
		// login type can't be changed yet but the infrastructure is in
		loginType := c.Request().FormValue("logintype")
		_ = loginType

		if firstName != user1.Firstname && len(firstName) > 0 {
			user1.Firstname = firstName
		}
		if nickname != user1.Nickname && len(nickname) > 0 {
			user1.Nickname = nickname
		}
		if lastName != user1.Lastname && len(lastName) > 0 {
			user1.Lastname = lastName
		}
		if username != user1.Username && len(username) > 0 {
			user1.Username = username
		}
		if universityUsername != user1.UniversityUsername.String && len(universityUsername) > 0 {
			user1.UniversityUsername = null.StringFrom(universityUsername)
		}
		if LDAPUsername != user1.LDAPUsername.String && len(LDAPUsername) > 0 {
			user1.LDAPUsername = null.StringFrom(LDAPUsername)
		}
		if email != user1.Email && len(email) > 0 {
			user1.Email = email
		}

		_, err = v.user.EditUser(c.Request().Context(), user1, c1.User.UserID)
		if err != nil {
			log.Printf("failed to edit user for editUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to edit user for editUser: %+v", err))
			}
		}
		return v.userFunc(c, userID)
	} else {
		return v.errorHandle(c, fmt.Errorf("invalid method used"))
	}
}

// UserToggleEnabledFunc handles an toggle enable user request
func (v *Views) UserToggleEnabledFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

		c1 := v.getData(session)

		userID, err := strconv.Atoi(c.Param("userid"))
		if err != nil {
			log.Printf("failed to get userid for toggleUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get userid for toggleUser: %+v", err))
			}
		}

		user1, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
		if err != nil {
			log.Printf("failed to get user for toggleUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get user for toggleUser: %+v", err))
			}
		}

		user1.Enabled = !user1.Enabled

		_, err = v.user.EditUser(c.Request().Context(), user1, c1.User.UserID)
		if err != nil {
			log.Printf("failed to edit user for toggleUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to edit user for toggleUser: %+v", err))
			}
		}
		return v.userFunc(c, userID)
	} else {
		return v.errorHandle(c, fmt.Errorf("invalid method used"))
	}
}

// UserDeleteFunc handles an delete user request
func (v *Views) UserDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

		c1 := v.getData(session)

		userID, err := strconv.Atoi(c.Param("userid"))
		if err != nil {
			log.Printf("failed to get userid for deleteUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get userid for deleteUser: %+v", err))
			}
		}

		user1, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
		if err != nil {
			log.Printf("failed to get user for deleteUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get user for deleteUser: %+v", err))
			}
		}

		err = v.user.RemoveRoleUsers(c.Request().Context(), user1)
		if err != nil {
			log.Printf("failed to delete roleUsers for deleteUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to delete roleUsers for deleteUser: %+v", err))
			}
		}

		_, err = v.user.DeleteUser(c.Request().Context(), user1, c1.User.UserID)
		if err != nil {
			log.Printf("failed to delete user for deleteUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to delete user for deleteUser: %+v", err))
			}
		}
		return v.UsersFunc(c)
	} else {
		return v.errorHandle(c, fmt.Errorf("invalid method used"))
	}
}
