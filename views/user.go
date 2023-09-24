package views

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/infrastructure/mail"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
	"github.com/ystv/web-auth/utils"
	"gopkg.in/guregu/null.v4"
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
	c1 := v.getSessionData(c)

	var err error

	if c.Request().Method == "POST" {
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
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid size, must be positive"))
			} else if size != 5 && size != 10 && size != 25 && size != 50 && size != 75 && size != 100 {
				size = 25
			}
		}

		if enabled == "enabled" || enabled == "disabled" {
			q.Set("enabled", enabled)
		} else if enabled != "any" {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("enabled must be set to either \"any\", \"enabled\" or \"disabled\""))
		}

		if deleted == "deleted" || deleted == "not_deleted" {
			q.Set("deleted", deleted)
		} else if deleted != "any" {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("deleted must be set to either \"any\", \"deleted\" or \"not_deleted\""))
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

	column := c.QueryParam("column")
	direction := c.QueryParam("direction")
	search := c.QueryParam("search")
	search, err = url.QueryUnescape(search)
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
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("unable to parse page for users: %w", err))
		}
		size, err = strconv.Atoi(sizeRaw)
		//nolint:gocritic
		if err != nil {
			size = 0
		} else if size <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid size, must be positive"))
		} else if size != 5 && size != 10 && size != 25 && size != 50 && size != 75 && size != 100 {
			size = 0
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

	dbUsers, fullCount, err := v.user.GetUsers(c.Request().Context(), size, page, search, column, direction, enabled, deleted)
	if err != nil {
		return fmt.Errorf("failed to get users for users: %w", err)
	}

	if len(dbUsers) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("size and page given is not valid"))
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

	detailedUser := DBUserToDetailedUser(userFromDB, v.user)

	detailedUser.Permissions, err = v.user.GetPermissionsForUser(c.Request().Context(), user.User{UserID: detailedUser.UserID})
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

// UserAddFunc handles an add user request
func (v *Views) UserAddFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	if c.Request().Method == http.MethodGet {
		data := struct {
			UserID     int
			ActivePage string
		}{
			UserID:     c1.User.UserID,
			ActivePage: "useradd",
		}

		return v.template.RenderTemplate(c.Response(), data, templates.UserAddTemplate, templates.RegularType)
	} else if c.Request().Method == http.MethodPost {
		err := c.Request().ParseForm()
		if err != nil {
			return fmt.Errorf("failed to parse form for userAdd: %w", err)
		}

		firstName := c.Request().FormValue("firstname")
		lastName := c.Request().FormValue("lastname")
		username := c.Request().FormValue("username")
		universityUsername := c.Request().FormValue("universityusername")
		email := c.Request().FormValue("email")

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
			return fmt.Errorf("failed to add user for addUser: %w", err)
		}

		var message struct {
			Message string `json:"message"`
			Error   error  `json:"error"`
		}

		mailer := v.mailer.ConnectMailer()

		if mailer != nil {
			template, err := v.template.GetEmailTemplate(templates.SignupEmailTemplate)
			if err != nil {
				return fmt.Errorf("failed to get email in addUser: %w", err)
			}

			file := mail.Mail{
				Subject: "Welcome to YSTV!",
				Tpl:     template,
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
			message.Message = fmt.Sprintf("No mailer present\nPlease send the username and password to this email: %s, username: %s, password: %s", email, username, password)
			message.Error = fmt.Errorf("no mailer present")
			log.Printf("no Mailer present")
		}
		log.Printf("created user: %s", u.Username)

		var status int

		return c.JSON(status, message)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

// UserEditFunc handles an edit user request
func (v *Views) UserEditFunc(c echo.Context) error {
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

		err = c.Request().ParseForm()
		if err != nil {
			return fmt.Errorf("failed to parse form for userEdit: %w", err)
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

		err = v.user.EditUser(c.Request().Context(), user1, c1.User.UserID)
		if err != nil {
			return fmt.Errorf("failed to edit user for editUser: %w", err)
		}
		return c.Redirect(http.StatusOK, fmt.Sprintf("/internal/user/%d", userID))
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
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
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
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

		err = v.user.RemoveRoleUsers(c.Request().Context(), user1)
		if err != nil {
			return fmt.Errorf("failed to delete roleUsers for deleteUser: %w", err)
		}

		err = v.user.DeleteUser(c.Request().Context(), user1, c1.User.UserID)
		if err != nil {
			return fmt.Errorf("failed to delete user for deleteUser: %w", err)
		}
		return v.UsersFunc(c)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}
