package views

import (
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/templates"
	"log"
	"time"

	"github.com/dustin/go-humanize"
)

type (
	// InternalTemplate represents the context for the internal template
	InternalTemplate struct {
		UserID              int
		Nickname            string
		LastLogin           string
		TotalUsers          int
		TotalActiveUsers    int
		LoginsPast24Hours   int
		ActiveUsersPastYear int
		ActivePage          string
		Assumed             bool
	}
)

// InternalFunc handles a request to the internal template
func (v *Views) InternalFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)
	lastLogin := time.Now()
	if c1.User.LastLogin.Valid {
		lastLogin = c1.User.LastLogin.Time
	}
	count, err := v.user.CountUsers(c.Request().Context())
	if err != nil {
		log.Println(err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	totalActiveUsers, err := v.user.CountUsersActive(c.Request().Context())
	if err != nil {
		log.Println(err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	hours24, err := v.user.CountUsers24Hours(c.Request().Context())
	if err != nil {
		log.Println(err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	pastYear, err := v.user.CountUsersPastYear(c.Request().Context())
	if err != nil {
		log.Println(err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	ctx := InternalTemplate{
		UserID:              c1.User.UserID,
		Nickname:            c1.User.Nickname,
		LastLogin:           humanize.Time(lastLogin),
		TotalUsers:          count,
		TotalActiveUsers:    totalActiveUsers,
		LoginsPast24Hours:   hours24,
		ActiveUsersPastYear: pastYear,
		ActivePage:          "dashboard",
	}
	return v.template.RenderTemplate(c.Response(), ctx, templates.InternalTemplate)
}
