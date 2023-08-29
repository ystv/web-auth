package views

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
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
		fmt.Println(enabled, len(enabled))
		if enabled == "enabled" || enabled == "disabled" {
			q.Set("enabled", enabled)
		} else if enabled != "any" {
			return v.errorHandle(c, fmt.Errorf("enabled must be set to either \"any\", \"enabled\" or \"disabled\""))
		}

		deleted := c.Request().FormValue("deleted")
		fmt.Println(deleted, len(deleted))
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

		fmt.Println(q.Encode())

		u.RawQuery = q.Encode()
		return c.Redirect(http.StatusFound, u.String())
	}

	column := c.Request().URL.Query().Get("column")
	direction := c.Request().URL.Query().Get("direction")
	search := c.Request().URL.Query().Get("search")
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

func (v *Views) UserAddFunc(c echo.Context) error {
	return nil
}

func (v *Views) UserEditFunc(c echo.Context) error {
	return nil
}

func (v *Views) UserDeleteFunc(c echo.Context) error {
	return nil
}
