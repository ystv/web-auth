package views

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/permission"
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
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	var err error

	if c.Request().Method == "POST" {
		err = c.Request().ParseForm()
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("unable to parse form for users: %w", err))
		}

		column := c.FormValue("column")
		direction := c.FormValue("direction")
		search := c.FormValue("search")
		var size int
		sizeRaw := c.FormValue("size")
		if sizeRaw == "all" {
			size = 0
		} else {
			size, err = strconv.Atoi(sizeRaw)
			if err != nil {
				size = 0
			} else if size <= 0 {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid size, must be positive"))
			} else if size != 5 && size != 10 && size != 25 && size != 50 && size != 75 && size != 100 {
				size = 25
			}
		}
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
		c.Request().Method = "GET"
		var redirect string
		if valid {
			if len(search) > 0 {
				if size > 0 {
					redirect = fmt.Sprintf("/internal/users?column=%s&direction=%s&search=%s&size=%d&page=1", column, direction, url.QueryEscape(search), size)
				} else {
					redirect = fmt.Sprintf("/internal/users?column=%s&direction=%s&search=%s", column, direction, url.QueryEscape(search))
				}
			} else {
				if size > 0 {
					redirect = fmt.Sprintf("/internal/users?column=%s&direction=%s&size=%d&page=1", column, direction, size)
				} else {
					redirect = fmt.Sprintf("/internal/users?column=%s&direction=%s", column, direction)
				}
			}
		} else if len(search) > 0 {
			if size > 0 {
				redirect = fmt.Sprintf("/internal/users?search=%s&size=%d&page=1", url.QueryEscape(search), size)
			} else {
				redirect = fmt.Sprintf("/internal/users?search=%s", url.QueryEscape(search))
			}
		} else {
			if size > 0 {
				redirect = fmt.Sprintf("/internal/users?size=%d&page=1", size)
			} else {
				c.Request().URL.Query().Del("*")
				redirect = "/internal/users"
			}
		}
		return c.Redirect(http.StatusFound, redirect)
	}

	column := c.QueryParam("column")
	direction := c.QueryParam("direction")
	search := c.QueryParam("search")
	var size, page, count int
	var countAll user.CountUsers
	sizeRaw := c.QueryParam("size")
	if sizeRaw == "all" {
		size = 0
	} else if len(sizeRaw) != 0 {
		page, err = strconv.Atoi(c.QueryParam("page"))
		if err != nil {
			page = 1
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("unable to parse page for users: %w", err))
		}
		size, err = strconv.Atoi(sizeRaw)
		if err != nil {
			size = 0
		} else if size <= 0 {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid size, must be positive"))
		} else if size != 5 && size != 10 && size != 25 && size != 50 && size != 75 && size != 100 {
			size = 0
		}

		countAll, err = v.user.CountUsersAll(c.Request().Context())
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get total users for users: %w", err))
		}

		count = countAll.TotalUsers

		if count <= size*(page-1) {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("size and page given is not valid"))
		}
	}

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
	if valid {
		if len(search) > 0 {
			if size > 0 && page > 0 {
				dbUsers, err = v.user.GetUsersSortedSearchSizePage(c.Request().Context(), column, direction, search, size, page)
				if err != nil {
					log.Println(err)
					if !v.conf.Debug {
						return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get users for users: %w", err))
					}
				}
				tmp, err := v.user.GetUsersSortedSearch(c.Request().Context(), column, direction, search)
				if err != nil {
					log.Println(err)
					if !v.conf.Debug {
						return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get users for users: %w", err))
					}
				}
				count = len(tmp)
			} else {
				dbUsers, err = v.user.GetUsersSortedSearch(c.Request().Context(), column, direction, search)
			}
		} else {
			if size > 0 && page > 0 {
				dbUsers, err = v.user.GetUsersSortedSizePage(c.Request().Context(), column, direction, size, page)
			} else {
				dbUsers, err = v.user.GetUsersSorted(c.Request().Context(), column, direction)
			}
		}
	} else if len(search) > 0 {
		if size > 0 && page > 0 {
			dbUsers, err = v.user.GetUsersSearchSizePage(c.Request().Context(), search, size, page)
			tmp, err := v.user.GetUsersSearch(c.Request().Context(), search)
			if err != nil {
				log.Println(err)
				if !v.conf.Debug {
					return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get users for users: %w", err))
				}
			}
			count = len(tmp)
		} else {
			dbUsers, err = v.user.GetUsersSearch(c.Request().Context(), search)
		}
	} else {
		if size > 0 && page > 0 {
			dbUsers, err = v.user.GetUsersSizePage(c.Request().Context(), size, page)
		} else {
			dbUsers, err = v.user.GetUsers(c.Request().Context())
		}
	}

	if err != nil {
		log.Println(err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get users for users: %w", err))
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

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		log.Printf("failed to get user permissions for users: %+v", err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get user permissions for users: %+v", err))
		}
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
			Enabled:    "",
			Deleted:    "",
		},
	}
	return v.template.RenderTemplate(c.Response(), data, templates.UsersTemplate, templates.PaginationType)
}

// UserFunc handles a users request
func (v *Views) UserFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	userID, err := strconv.Atoi(c.Param("userid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse userid for user: %w", err))
	}
	userFromDB, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
	if err != nil {
		log.Printf("failed to get user in user: %+v", err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get user for user: %w", err))
		}
	}

	detailedUser := DBUserToDetailedUser(userFromDB, v.user)

	detailedUser.Permissions, err = v.user.GetPermissionsForUser(c.Request().Context(), user.User{UserID: detailedUser.UserID})
	if err != nil {
		log.Println(err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get permissions for user: %w", err))
		}
	}

	detailedUser.Permissions = v.removeDuplicate(detailedUser.Permissions)

	detailedUser.Roles, err = v.user.GetRolesForUser(c.Request().Context(), user.User{UserID: detailedUser.UserID})
	if err != nil {
		log.Println(err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get roles for user: %w", err))
		}
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		log.Printf("failed to get user permissions for user: %+v", err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get user permissions for user: %+v", err))
		}
	}

	data := UserTemplate{
		User:            detailedUser,
		UserPermissions: p1,
		ActivePage:      "user",
	}

	return v.template.RenderTemplate(c.Response(), data, templates.UserTemplate, templates.RegularType)
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
