package views

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/ystv/web-auth/crowd"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/utils"
)

type CrowdAppTemplate struct {
	CrowdApps           []crowd.CrowdApp
	AddedCrowdApp       *crowd.CrowdApp
	CrowdAppsStatusSort string
	Error               string
	TemplateHelper
}

func (v *Views) CrowdAppsFunc(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodGet:
		c1 := v.getSessionData(c)

		status := c.QueryParam("status")

		var dbStatus crowd.CrowdAppStatus
		switch status {
		case "active", "":
			status = "active"
			dbStatus = crowd.Active
		case "inactive":
			dbStatus = crowd.Inactive
		case "any":
			dbStatus = crowd.Any
		default:
			return echo.NewHTTPError(http.StatusBadRequest,
				errors.New("status must be set to either \"any\", \"active\" or \"inactive\""))
		}

		crowdApps, err := v.crowd.GetCrowdApps(c.Request().Context(), dbStatus)
		if err != nil {
			return fmt.Errorf("failed to get crowd apps: %w", err)
		}

		p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			return fmt.Errorf("failed to get user permissions for crowd apps: %w", err)
		}

		data := CrowdAppTemplate{
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

func (v *Views) crowdAppsFunc(c echo.Context, addedCrowdApp crowd.CrowdApp) error {
	c1 := v.getSessionData(c)

	status := "any"
	dbStatus := crowd.Any

	crowdApps, err := v.crowd.GetCrowdApps(c.Request().Context(), dbStatus)
	if err != nil {
		return fmt.Errorf("failed to get crowd apps: %w", err)
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		return fmt.Errorf("failed to get user permissions for crowd apps: %w", err)
	}

	data := CrowdAppTemplate{
		CrowdApps:           crowdApps,
		AddedCrowdApp:       &addedCrowdApp,
		CrowdAppsStatusSort: status,
		Error:               c.QueryParam("error"),
		TemplateHelper: TemplateHelper{
			UserPermissions: p1,
			ActivePage:      "crowdapps",
			Assumed:         c1.Assumed,
		},
	}

	return v.template.RenderTemplate(c.Response(), data, templates.CrowdAppsTemplate, templates.RegularType)
}

func (v *Views) CrowdAppFunc(c echo.Context) error {
	if c.Request().Method == http.MethodGet {
		c1 := v.getSessionData(c)

		crowdAppID, err := strconv.Atoi(c.Param("crowdappid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse crowdappid for crowd app: %w", err))
		}

		crowd1, err := v.crowd.GetCrowdApp(c.Request().Context(), crowd.CrowdApp{AppID: crowdAppID})
		if err != nil {
			return fmt.Errorf("failed to get crowd app: %w", err)
		}

		p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			return fmt.Errorf("failed to get user permissions for corwd app: %w", err)
		}

		data := struct {
			CrowdApp crowd.CrowdApp
			TemplateHelper
		}{
			CrowdApp: crowd1,
			TemplateHelper: TemplateHelper{
				UserPermissions: p1,
				ActivePage:      "crowdapp",
				Assumed:         c1.Assumed,
			},
		}

		return v.template.RenderTemplate(c.Response(), data, templates.CrowdAppTemplate, templates.RegularType)
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) CrowdAppAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		username, err := utils.GenerateRandom(utils.GenerateUsername)
		if err != nil {
			return fmt.Errorf("error generating username: %w", err)
		}

		password, err := utils.GenerateRandomLength(20, utils.GeneratePassword)
		if err != nil {
			return fmt.Errorf("error generating password: %w", err)
		}

		salt, err := utils.GenerateRandom(utils.GenerateSalt)
		if err != nil {
			return fmt.Errorf("error generating salt: %w", err)
		}

		name := c.Request().FormValue("name")
		description := null.StringFrom(c.Request().FormValue("description"))
		activeTemp := c.FormValue("active")
		active := false

		if activeTemp == "on" {
			active = true
		}

		if name == "" {
			return c.Redirect(http.StatusFound, "/internal/crowdapps?error="+
				url.QueryEscape("Name must be filled"))
		}

		ca1, err := v.crowd.GetCrowdApp(c.Request().Context(),
			crowd.CrowdApp{AppID: 0, Username: username})
		if err == nil && ca1.AppID > 0 {
			return errors.New("crowd app already exists")
		}

		addedCrowdApp, err := v.crowd.AddCrowdApp(c.Request().Context(),
			crowd.CrowdApp{
				Name:        name,
				Username:    username,
				Description: description,
				Active:      active,
				Password:    null.StringFrom(password),
				Salt:        null.StringFrom(salt),
			})
		if err != nil {
			return fmt.Errorf("failed to add crowd app for addCrowdApp: %w", err)
		}

		addedCrowdApp.Password = null.StringFrom(password)

		c.Request().Method = http.MethodGet

		return v.crowdAppsFunc(c, addedCrowdApp)
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) CrowdAppEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		crowdAppID, err := strconv.Atoi(c.Param("crowdappid"))
		if err != nil {
			return fmt.Errorf("failed to get officershipid for editOfficership: %w", err)
		}

		crowd1, err := v.crowd.GetCrowdApp(c.Request().Context(),
			crowd.CrowdApp{AppID: crowdAppID})
		if err != nil {
			return fmt.Errorf("failed to get officership for editOfficership: %w", err)
		}

		name := c.FormValue("name")
		description := c.FormValue("description")
		activeTemp := c.FormValue("active")
		active := false

		if activeTemp == "on" {
			active = true
		}

		crowd1.Name = name
		crowd1.Description = null.StringFrom(description)
		crowd1.Active = active

		_, err = v.crowd.EditCrowdApp(c.Request().Context(), crowd1)
		if err != nil {
			return fmt.Errorf("failed to edit crowd app for editCrowdApp: %w", err)
		}

		return c.Redirect(http.StatusFound, fmt.Sprintf("/internal/crowdapp/%d", crowdAppID))
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) CrowdAppDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		crowdAppID, err := strconv.Atoi(c.Param("crowdappid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest,
				fmt.Errorf("failed to parse crowdappid for crowd app delete: %w", err))
		}

		crowd1, err := v.crowd.GetCrowdApp(c.Request().Context(),
			crowd.CrowdApp{AppID: crowdAppID})
		if err != nil {
			return fmt.Errorf("failed to get crowd app for crowd app delete: %w", err)
		}

		err = v.crowd.DeleteCrowdApp(c.Request().Context(), crowd1)
		if err != nil {
			return fmt.Errorf("failed to delete crowd app for corwd app delete: %w", err)
		}

		return c.Redirect(http.StatusFound, "/internal/crowdapps")
	}

	return v.invalidMethodUsed(c)
}
