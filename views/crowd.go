package views

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
	"net/http"
	"net/url"
)

func (v *Views) CrowdAppsFunc(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodGet:
		c1 := v.getSessionData(c)

		status := c.QueryParam("status")

		var dbStatus user.CrowdAppStatus
		switch status {
		case "active", "":
			status = "active"
			dbStatus = user.Active
		case "inactive":
			dbStatus = user.Inactive
		case "any":
			dbStatus = user.Any
		default:
			return echo.NewHTTPError(http.StatusBadRequest,
				errors.New("status must be set to either \"any\", \"active\" or \"inactive\""))
		}

		crowdApps, err := v.user.GetCrowdApps(c.Request().Context(), dbStatus)
		if err != nil {
			return fmt.Errorf("failed to get crowd apps: %w", err)
		}

		p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			return fmt.Errorf("failed to get user permissions for crowd apps: %w", err)
		}

		data := struct {
			CrowdApps           []user.CrowdApp
			CrowdAppsStatusSort string
			Error               string
			TemplateHelper
		}{
			CrowdApps:           crowdApps,
			CrowdAppsStatusSort: status,
			Error:               c.QueryParam("error"),
			TemplateHelper: TemplateHelper{
				UserPermissions: p1,
				ActivePage:      "crowdapps",
				Assumed:         c1.Assumed,
			},
		}

		return v.template.RenderTemplate(c.Response(), data, templates.CrowdAppsTemplate, templates.RegularType)
	case http.MethodPost:
		o, err := url.Parse("/internal/crowdapps")
		if err != nil {
			panic(fmt.Errorf("invalid url: %w", err)) // this panics because if this errors then many other things will be wrong
		}

		q := o.Query()

		status := c.FormValue("status")

		if status == "inactive" || status == "any" {
			q.Set("status", status)
		} else if status != "active" {
			return echo.NewHTTPError(http.StatusBadRequest,
				errors.New("status must be set to either \"any\", \"active\" or \"inactive\""))
		}

		c.Request().Method = "GET"

		o.RawQuery = q.Encode()

		return c.Redirect(http.StatusFound, o.String())
	}

	return v.invalidMethodUsed(c)
}
