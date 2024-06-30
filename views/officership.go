package views

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"gopkg.in/guregu/null.v4"

	"github.com/ystv/web-auth/officership"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
)

func (v *Views) OfficershipsFunc(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodGet:
		c1 := v.getSessionData(c)

		status := c.QueryParam("status")

		var dbStatus officership.OfficershipsStatus
		switch status {
		case "current", "":
			status = "current"
			dbStatus = officership.Current
		case "retired":
			dbStatus = officership.Retired
		case "any":
			dbStatus = officership.Any
		default:
			return echo.NewHTTPError(http.StatusBadRequest,
				errors.New("status must be set to either \"any\", \"current\" or \"retired\""))
		}

		officerships, err := v.officership.GetOfficerships(c.Request().Context(), dbStatus)
		if err != nil {
			return fmt.Errorf("failed to get officerships: %w", err)
		}

		p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			return fmt.Errorf("failed to get user permissions for officerships: %w", err)
		}

		data := struct {
			Officerships          []officership.Officership
			OfficershipStatusSort string
			TemplateHelper
		}{
			Officerships:          officerships,
			OfficershipStatusSort: status,
			TemplateHelper: TemplateHelper{
				UserPermissions: p1,
				ActivePage:      "officerships",
				Assumed:         c1.Assumed,
			},
		}

		return v.template.RenderTemplate(c.Response(), data, templates.OfficershipsTemplate, templates.RegularType)
	case http.MethodPost:
		o, err := url.Parse("/internal/officerships")
		if err != nil {
			panic(fmt.Errorf("invalid url: %w", err)) // this panics because if this errors then many other things will be wrong
		}

		q := o.Query()

		status := c.FormValue("status")

		if status == "retired" || status == "any" {
			q.Set("status", status)
		} else if status != "current" {
			return echo.NewHTTPError(http.StatusBadRequest,
				errors.New("status must be set to either \"any\", \"current\" or \"retired\""))
		}

		c.Request().Method = "GET"

		o.RawQuery = q.Encode()

		return c.Redirect(http.StatusFound, o.String())
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficershipFunc(c echo.Context) error {
	if c.Request().Method == http.MethodGet {
		c1 := v.getSessionData(c)

		officershipID, err := strconv.Atoi(c.Param("officershipid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse officershipid for officership: %w", err))
		}

		o, err := v.officership.GetOfficership(c.Request().Context(), officership.Officership{OfficershipID: officershipID})
		if err != nil {
			return fmt.Errorf("failed to get officership: %w", err)
		}

		officers, err := v.officership.GetOfficershipMembers(c.Request().Context(), &o, nil, officership.Any, officership.Any, false)
		if err != nil {
			return fmt.Errorf("failed to get officership members: %w", err)
		}

		p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			return fmt.Errorf("failed to get user permissions for officership: %w", err)
		}

		data := struct {
			Officership officership.Officership
			Officers    []officership.OfficershipMember
			TemplateHelper
		}{
			Officership: o,
			Officers:    officers,
			TemplateHelper: TemplateHelper{
				UserPermissions: p1,
				ActivePage:      "officership",
				Assumed:         c1.Assumed,
			},
		}

		return v.template.RenderTemplate(c.Response(), data, templates.OfficershipTemplate, templates.RegularType)
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficershipAddFunc(c echo.Context) error {
	return v.invalidMethodUsed(c)
}

func (v *Views) OfficershipEditFunc(c echo.Context) error {
	return v.invalidMethodUsed(c)
}

func (v *Views) OfficershipDeleteFunc(c echo.Context) error {
	return v.invalidMethodUsed(c)
}

func (v *Views) OfficersFunc(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodGet:
		c1 := v.getSessionData(c)

		officershipStatus := c.QueryParam("officershipStatus")
		officerStatus := c.QueryParam("officerStatus")

		var dbOfficershipStatus, dbOfficerStatus officership.OfficershipsStatus
		switch officershipStatus {
		case "current", "":
			officershipStatus = "current"
			dbOfficershipStatus = officership.Current
		case "retired":
			dbOfficershipStatus = officership.Retired
		case "any":
			dbOfficershipStatus = officership.Any
		default:
			return echo.NewHTTPError(http.StatusBadRequest,
				errors.New("officershipStatus must be set to either \"any\", \"current\" or \"retired\""))
		}

		switch officerStatus {
		case "current", "":
			officerStatus = "current"
			dbOfficerStatus = officership.Current
		case "retired":
			dbOfficerStatus = officership.Retired
		case "any":
			dbOfficerStatus = officership.Any
		default:
			return echo.NewHTTPError(http.StatusBadRequest,
				errors.New("officerStatus must be set to either \"any\", \"current\" or \"retired\""))
		}

		wg := sync.WaitGroup{}
		wg.Add(3)

		var errArr []error

		var officers []officership.OfficershipMember

		var officerships []officership.Officership

		var users []user.User

		go func() {
			defer wg.Done()

			var err error

			officers, err = v.officership.GetOfficershipMembers(c.Request().Context(), nil, nil, dbOfficershipStatus,
				dbOfficerStatus, true)
			if err != nil {
				errArr = append(errArr, fmt.Errorf("failed to get officers: %w", err))
			}
		}()

		go func() {
			defer wg.Done()

			var err error

			officerships, err = v.officership.GetOfficerships(c.Request().Context(), dbOfficershipStatus)
			if err != nil {
				errArr = append(errArr, fmt.Errorf("failed to get officerships: %w", err))
			}
		}()

		go func() {
			defer wg.Done()

			var err error

			users, _, err = v.user.GetUsers(c.Request().Context(), 0, 0, "", "", "", "enabled", "not_deleted")
			if errArr != nil {
				errArr = append(errArr, fmt.Errorf("failed to get users: %w", err))
			}
		}()
		wg.Wait()

		if len(errArr) != 0 {
			var sb strings.Builder

			for _, err := range errArr {
				sb.WriteString(err.Error())
			}

			return errors.New(sb.String())
		}

		p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			return fmt.Errorf("failed to get user permissions for officers: %w", err)
		}

		data := struct {
			Officers              []officership.OfficershipMember
			Officerships          []officership.Officership
			Users                 []user.User
			OfficershipStatusSort string
			OfficerStatusSort     string
			TemplateHelper
		}{
			Officers:              officers,
			Officerships:          officerships,
			Users:                 users,
			OfficershipStatusSort: officershipStatus,
			OfficerStatusSort:     officerStatus,
			TemplateHelper: TemplateHelper{
				UserPermissions: p1,
				ActivePage:      "officers",
				Assumed:         c1.Assumed,
			},
		}

		return v.template.RenderTemplate(c.Response(), data, templates.OfficersTemplate, templates.RegularType)
	case http.MethodPost:
		o, err := url.Parse("/internal/officership/officers")
		if err != nil {
			panic(fmt.Errorf("invalid url: %w", err)) // this panics because if this errors then many other things will be wrong
		}

		q := o.Query()

		officershipStatus := c.FormValue("officershipStatus")
		officerStatus := c.FormValue("officerStatus")

		if officershipStatus == "retired" || officershipStatus == "any" {
			q.Set("officershipStatus", officershipStatus)
		} else if officershipStatus != "current" {
			return echo.NewHTTPError(http.StatusBadRequest,
				errors.New("officershipStatus must be set to either \"any\", \"current\" or \"retired\""))
		}

		if officerStatus == "retired" || officerStatus == "any" {
			q.Set("officerStatus", officerStatus)
		} else if officerStatus != "current" {
			return echo.NewHTTPError(http.StatusBadRequest,
				errors.New("officerStatus must be set to either \"any\", \"current\" or \"retired\""))
		}

		c.Request().Method = "GET"

		o.RawQuery = q.Encode()

		return c.Redirect(http.StatusFound, o.String())
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficerFunc(c echo.Context) error {
	if c.Request().Method == http.MethodGet {
		c1 := v.getSessionData(c)

		officerID, err := strconv.Atoi(c.Param("officerid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to parse officerid for officer: %w", err))
		}

		officer, err := v.officership.GetOfficershipMember(c.Request().Context(), officership.OfficershipMember{OfficershipMemberID: officerID})
		if err != nil {
			return fmt.Errorf("failed to get officer: %w", err)
		}

		p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			return fmt.Errorf("failed to get user permissions for officer: %w", err)
		}

		data := struct {
			Officer officership.OfficershipMember
			// Officerships []officership.Officership
			// Users        []user.User
			TemplateHelper
		}{
			Officer: officer,
			// Officerships: officerships,
			// Users:        users,
			TemplateHelper: TemplateHelper{
				UserPermissions: p1,
				ActivePage:      "officer",
				Assumed:         c1.Assumed,
			},
		}

		return v.template.RenderTemplate(c.Response(), data, templates.OfficerTemplate, templates.RegularType)
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficerAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		tempUserID := c.FormValue("userID")
		tempOfficershipID := c.FormValue("officershipID")
		tempStartDate := c.FormValue("startDate")

		parse, err := time.Parse("02/01/2006", tempStartDate)
		if err != nil {
			return fmt.Errorf("failed to parse start date: %w", err)
		}

		diff := time.Now().Compare(parse)
		if diff != 1 {
			return errors.New("start date must be before today")
		}

		userID, err := strconv.Atoi(tempUserID)
		if err != nil {
			return fmt.Errorf("failed to convert user id to int: %w", err)
		}

		officershipID, err := strconv.Atoi(tempOfficershipID)
		if err != nil {
			return fmt.Errorf("failed to convert officershipID to int: %w", err)
		}

		u1, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
		if err != nil {
			return fmt.Errorf("failed to get user for officerAdd: %w", err)
		}

		o1, err := v.officership.GetOfficership(c.Request().Context(), officership.Officership{OfficershipID: officershipID})
		if err != nil {
			return fmt.Errorf("failed to get officership for officerAdd: %w", err)
		}

		_, err = v.officership.AddOfficershipMember(c.Request().Context(), officership.OfficershipMember{
			UserID:    u1.UserID,
			OfficerID: o1.OfficershipID,
			StartDate: null.TimeFrom(parse),
		})
		if err != nil {
			return fmt.Errorf("failed to add officer for officerAdd: %w", err)
		}

		return c.Redirect(http.StatusFound, "/internal/officership/officers")
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficerEditFunc(c echo.Context) error {
	return v.invalidMethodUsed(c)
}

func (v *Views) OfficerDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		officerID, err := strconv.Atoi(c.Param("officerid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest,
				fmt.Errorf("failed to parse officerid for officer delete: %w", err))
		}

		officer, err := v.officership.GetOfficershipMember(c.Request().Context(),
			officership.OfficershipMember{OfficershipMemberID: officerID})
		if err != nil {
			return fmt.Errorf("failed to get officer for officer delete: %w", err)
		}

		err = v.officership.DeleteOfficershipMember(c.Request().Context(), officer)
		if err != nil {
			return fmt.Errorf("failed to delete officer for officer delete: %w", err)
		}

		return c.Redirect(http.StatusFound, "/internal/officership/officers")
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficershipTeamsFunc(c echo.Context) error {
	if c.Request().Method == http.MethodGet {
		c1 := v.getSessionData(c)

		officers, err := v.officership.GetOfficershipTeams(c.Request().Context())
		if err != nil {
			return fmt.Errorf("failed to get officershipTeams: %w", err)
		}

		p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			return fmt.Errorf("failed to get user permissions for officershipTeams: %w", err)
		}

		data := struct {
			OfficershipTeams []officership.OfficershipTeam
			TemplateHelper
		}{
			OfficershipTeams: officers,
			TemplateHelper: TemplateHelper{
				UserPermissions: p1,
				ActivePage:      "officershipTeams",
				Assumed:         c1.Assumed,
			},
		}

		return v.template.RenderTemplate(c.Response(), data, templates.OfficershipTeamsTemplate, templates.RegularType)
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficershipTeamFunc(c echo.Context) error {
	if c.Request().Method == http.MethodGet {
		c1 := v.getSessionData(c)

		officershipTeamID, err := strconv.Atoi(c.Param("officershipteamid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest,
				fmt.Errorf("failed to parse officershipteamid for officership team: %w", err))
		}

		officershipTeam, err := v.officership.GetOfficershipTeam(c.Request().Context(),
			officership.OfficershipTeam{TeamID: officershipTeamID})
		if err != nil {
			return fmt.Errorf("failed to get officershipTeam: %w", err)
		}

		permissions, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			return fmt.Errorf("failed to get user permissions for officershipTeam: %w", err)
		}

		teamMembers, err := v.officership.GetOfficershipTeamMembers(c.Request().Context(), &officershipTeam,
			officership.Any)
		if err != nil {
			return fmt.Errorf("failed to get officership team members for officershipTeam: %w", err)
		}

		data := struct {
			OfficershipTeam struct {
				TeamID           int
				Name             string
				EmailAlias       string
				ShortDescription string
				FullDescription  string
				TeamMembers      []officership.OfficershipTeamMember
			}
			TemplateHelper
		}{
			OfficershipTeam: struct {
				TeamID           int
				Name             string
				EmailAlias       string
				ShortDescription string
				FullDescription  string
				TeamMembers      []officership.OfficershipTeamMember
			}{
				TeamID:           officershipTeam.TeamID,
				Name:             officershipTeam.Name,
				EmailAlias:       officershipTeam.EmailAlias,
				ShortDescription: officershipTeam.ShortDescription,
				FullDescription:  officershipTeam.FullDescription,
				TeamMembers:      teamMembers,
			},
			TemplateHelper: TemplateHelper{
				UserPermissions: permissions,
				ActivePage:      "officershipTeam",
				Assumed:         c1.Assumed,
			},
		}

		return v.template.RenderTemplate(c.Response(), data, templates.OfficershipTeamTemplate, templates.RegularType)
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficershipTeamAddFunc(c echo.Context) error {
	return v.invalidMethodUsed(c)
}

func (v *Views) OfficershipTeamEditFunc(c echo.Context) error {
	return v.invalidMethodUsed(c)
}

func (v *Views) OfficershipTeamDeleteFunc(c echo.Context) error {
	return v.invalidMethodUsed(c)
}
