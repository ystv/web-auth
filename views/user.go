package views

import (
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
)

type (
	// UsersTemplate represents the context for the user template
	UsersTemplate struct {
		Users           []user.StrippedUser
		UserPermissions []permission.Permission
		CurPage         int
		NextPage        int
		PrevPage        int
		LastPage        int
		ActivePage      string
		Sort            Sort
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
		User            user.DetailedUser
		UserPermissions []permission.Permission
		ActivePage      string
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
		Users:           tplUsers,
		UserPermissions: p1,
		ActivePage:      "users",
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

	detailedUser := DBUserToUserTemplateFormat(userFromDB, v.user)

	detailedUser.Permissions, err = v.user.GetPermissionsForUser(c.Request().Context(), user.User{UserID: detailedUser.UserID})
	if err != nil {
		return fmt.Errorf("failed to get permissions for user: %w", err)
	}

	detailedUser.Permissions = v.removeDuplicate(detailedUser.Permissions)

	detailedUser.Roles, err = v.user.GetRolesForUser(c.Request().Context(), user.User{UserID: detailedUser.UserID})
	if err != nil {
		return fmt.Errorf("failed to get roles for user: %w", err)
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		return fmt.Errorf("failed to get user permissions for user: %w", err)
	}

	data := UserTemplate{
		User:            detailedUser,
		UserPermissions: p1,
		ActivePage:      "user",
	}

	return v.template.RenderTemplate(c.Response(), data, templates.UserTemplate, templates.RegularType)
}

func (v *Views) UserAddFunc(c echo.Context) error {
	_ = c
	return nil
}

func (v *Views) UserEditFunc(c echo.Context) error {
	_ = c
	return nil
}

func (v *Views) UserDeleteFunc(c echo.Context) error {
	_ = c
	return nil
}
