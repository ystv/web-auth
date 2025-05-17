package views

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
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
			return errors.Errorf("failed to get officerships: %+v", err)
		}

		p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			return errors.Errorf("failed to get user permissions for officerships: %+v", err)
		}

		data := struct {
			Officerships          []officership.Officership
			OfficershipStatusSort string
			Error                 string
			TemplateHelper
		}{
			Officerships:          officerships,
			OfficershipStatusSort: status,
			Error:                 c.QueryParam("error"),
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
			panic(errors.Errorf("invalid url: %+v", err)) // this panics because if this errors then many other things will be wrong
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
			return echo.NewHTTPError(http.StatusBadRequest, errors.Errorf("failed to parse officershipid for officership: %+v", err))
		}

		o, err := v.officership.GetOfficership(c.Request().Context(), officership.Officership{OfficershipID: officershipID})
		if err != nil {
			return errors.Errorf("failed to get officership: %+v", err)
		}

		officers, err := v.officership.GetOfficershipMembers(c.Request().Context(), &o, nil, officership.Any, officership.Any, false)
		if err != nil {
			return errors.Errorf("failed to get officership members: %+v", err)
		}

		p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			return errors.Errorf("failed to get user permissions for officership: %+v", err)
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
	if c.Request().Method == http.MethodPost {
		name := c.Request().FormValue("name")
		emailAlias := c.Request().FormValue("emailAlias")
		description := c.Request().FormValue("description")
		historyWikiURL := c.Request().FormValue("historyWikiURL")
		isCurrentTemp := c.FormValue("isCurrent")
		isCurrent := false

		if isCurrentTemp == "on" {
			isCurrent = true
		}

		if name == "" || emailAlias == "" || description == "" {
			return c.Redirect(http.StatusFound, "/internal/officerships?error="+
				url.QueryEscape("Name, email alias and description must be filled"))
		}

		if historyWikiURL != "" {
			_, err := url.ParseRequestURI(historyWikiURL)
			if err != nil {
				return errors.Errorf("failed to parse historyWikiURL: %+v", err)
			}
		}

		o1, err := v.officership.GetOfficership(c.Request().Context(),
			officership.Officership{OfficershipID: 0, Name: name})
		if err == nil && o1.OfficershipID > 0 {
			return errors.New("officership with name \"" + name + "\" already exists")
		}

		_, err = v.officership.AddOfficership(c.Request().Context(),
			officership.Officership{
				Name:           name,
				EmailAlias:     emailAlias,
				Description:    description,
				HistoryWikiURL: historyWikiURL,
				IsCurrent:      isCurrent,
			})
		if err != nil {
			return errors.Errorf("failed to add officerships for addOfficership: %+v", err)
		}

		return c.Redirect(http.StatusFound, "/internal/officerships")
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficershipEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		officershipID, err := strconv.Atoi(c.Param("officershipid"))
		if err != nil {
			return errors.Errorf("failed to get officershipid for editOfficership: %+v", err)
		}

		officership1, err := v.officership.GetOfficership(c.Request().Context(),
			officership.Officership{OfficershipID: officershipID})
		if err != nil {
			return errors.Errorf("failed to get officership for editOfficership: %+v", err)
		}

		name := c.FormValue("name")
		emailAlias := c.FormValue("emailAlias")
		description := c.FormValue("description")
		historyWikiURL := c.FormValue("historyWikiURL")
		isCurrentTemp := c.FormValue("isCurrent")
		isCurrent := false

		if isCurrentTemp == "on" {
			isCurrent = true
		}

		if historyWikiURL != "" {
			_, err = url.ParseRequestURI(historyWikiURL)
			if err != nil {
				return errors.Errorf("failed to parse historyWikiURL: %+v", err)
			}
		}

		officership1.Name = name
		officership1.EmailAlias = emailAlias
		officership1.Description = description
		officership1.HistoryWikiURL = historyWikiURL
		officership1.IsCurrent = isCurrent

		_, err = v.officership.EditOfficership(c.Request().Context(), officership1)
		if err != nil {
			return errors.Errorf("failed to edit officership for editOfficership: %+v", err)
		}

		return c.Redirect(http.StatusFound, fmt.Sprintf("/internal/officership/%d", officershipID))
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficershipDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		officershipID, err := strconv.Atoi(c.Param("officershipid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest,
				errors.Errorf("failed to parse officershipid for officership delete: %+v", err))
		}

		o, err := v.officership.GetOfficership(c.Request().Context(),
			officership.Officership{OfficershipID: officershipID})
		if err != nil {
			return errors.Errorf("failed to get officership for officership delete: %+v", err)
		}

		err = v.officership.RemoveOfficershipForOfficershipMembers(c.Request().Context(), o)
		if err != nil {
			return errors.Errorf("failed to delete officers from officership for officership delete: %+v", err)
		}

		if o.TeamID.Valid {
			err = v.officership.DeleteOfficershipTeamMember(c.Request().Context(),
				officership.OfficershipTeamMember{OfficerID: officershipID})
			if err != nil {
				return errors.Errorf("failed to delete team from officership for officership delete: %+v", err)
			}
		}

		err = v.officership.DeleteOfficership(c.Request().Context(), o)
		if err != nil {
			return errors.Errorf("failed to delete officership for officership delete: %+v", err)
		}

		return c.Redirect(http.StatusFound, "/internal/officerships")
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficersFunc(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodGet:
		return v._officersGet(c)
	case http.MethodPost:
		return v._officersPost(c)
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) _officersGet(c echo.Context) error {
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
			errArr = append(errArr, errors.Errorf("failed to get officers: %+v", err))
		}
	}()

	go func() {
		defer wg.Done()

		var err error

		officerships, err = v.officership.GetOfficerships(c.Request().Context(), dbOfficershipStatus)
		if err != nil {
			errArr = append(errArr, errors.Errorf("failed to get officerships: %+v", err))
		}
	}()

	go func() {
		defer wg.Done()

		var err error

		users, _, err = v.user.GetUsers(c.Request().Context(), 0, 0, "", "", "", "", "not_deleted")
		if errArr != nil {
			errArr = append(errArr, errors.Errorf("failed to get users: %+v", err))
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
		return errors.Errorf("failed to get user permissions for officers: %+v", err)
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
}

func (v *Views) _officersPost(c echo.Context) error {
	o, err := url.Parse("/internal/officership/officers")
	if err != nil {
		panic(errors.Errorf("invalid url: %+v", err)) // this panics because if this errors then many other things will be wrong
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

func (v *Views) OfficerFunc(c echo.Context) error {
	if c.Request().Method == http.MethodGet {
		c1 := v.getSessionData(c)

		officerID, err := strconv.Atoi(c.Param("officerid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, errors.Errorf("failed to parse officerid for officer: %+v", err))
		}

		officer, err := v.officership.GetOfficershipMember(c.Request().Context(), officership.OfficershipMember{OfficershipMemberID: officerID})
		if err != nil {
			return errors.Errorf("failed to get officer: %+v", err)
		}

		officerships, err := v.officership.GetOfficerships(c.Request().Context(), officership.Any)
		if err != nil {
			return errors.Errorf("failed to get officerships: %+v", err)
		}

		users, _, err := v.user.GetUsers(c.Request().Context(), 0, 0, "", "", "", "", "not_deleted")
		if err != nil {
			return errors.Errorf("failed to get users: %+v", err)
		}

		p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			return errors.Errorf("failed to get user permissions for officer: %+v", err)
		}

		data := struct {
			Officer      officership.OfficershipMember
			Officerships []officership.Officership
			Users        []user.User
			TemplateHelper
		}{
			Officer:      officer,
			Officerships: officerships,
			Users:        users,
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
		tempEndDate := c.FormValue("endDate")

		parseStart, err := time.Parse("02/01/2006", tempStartDate)
		if err != nil {
			return errors.Errorf("failed to parse start date: %+v", err)
		}

		diffStart := time.Now().Compare(parseStart)
		if diffStart != 1 {
			return errors.New("start date must be before today")
		}

		// Add 19 hours to always be at the end of the day when adding vs the midnight for ending,
		// this takes into consideration daylight savings from the server side
		// Liam - changed to be the start of an Admin meeting
		parseStart = parseStart.Add(time.Hour * 19)

		endDate := null.NewTime(time.Time{}, false)

		if tempEndDate != "" {
			var parseEnd time.Time
			parseEnd, err = time.Parse("02/01/2006", tempEndDate)
			if err != nil {
				return errors.Errorf("failed to parse end date: %+v", err)
			}

			diffEnd := time.Now().Compare(parseEnd)
			if diffEnd != 1 {
				return errors.New("end date must be before today")
			}

			endDate = null.TimeFrom(parseEnd)
		}

		userID, err := strconv.Atoi(tempUserID)
		if err != nil {
			return errors.Errorf("failed to convert user id to int: %+v", err)
		}

		officershipID, err := strconv.Atoi(tempOfficershipID)
		if err != nil {
			return errors.Errorf("failed to convert officershipID to int: %+v", err)
		}

		u1, err := v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
		if err != nil {
			return errors.Errorf("failed to get user for officerAdd: %+v", err)
		}

		o1, err := v.officership.GetOfficership(c.Request().Context(), officership.Officership{OfficershipID: officershipID})
		if err != nil {
			return errors.Errorf("failed to get officership for officerAdd: %+v", err)
		}

		_, err = v.officership.AddOfficershipMember(c.Request().Context(), officership.OfficershipMember{
			UserID:    u1.UserID,
			OfficerID: o1.OfficershipID,
			StartDate: null.TimeFrom(parseStart),
			EndDate:   endDate,
		})
		if err != nil {
			return errors.Errorf("failed to add officer for officerAdd: %+v", err)
		}

		return c.Redirect(http.StatusFound, "/internal/officership/officers")
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficerEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		officerID, err := strconv.Atoi(c.Param("officerid"))
		if err != nil {
			return errors.Errorf("failed to get officerid for editOfficer: %+v", err)
		}

		officer1, err := v.officership.GetOfficershipMember(c.Request().Context(),
			officership.OfficershipMember{OfficershipMemberID: officerID})
		if err != nil {
			return errors.Errorf("failed to get officer for editOfficer: %+v", err)
		}

		userID, err := strconv.Atoi(c.FormValue("userID"))
		if err != nil {
			return errors.Errorf("failed to get userID form for editOfficer: %+v", err)
		}

		_, err = v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
		if err != nil {
			return errors.Errorf("failed to get user form for editOfficer: %+v", err)
		}

		officershipID, err := strconv.Atoi(c.FormValue("officershipID"))
		if err != nil {
			return errors.Errorf("failed to get officershipID form for editOfficer: %+v", err)
		}

		_, err = v.officership.GetOfficership(c.Request().Context(),
			officership.Officership{OfficershipID: officershipID})
		if err != nil {
			return errors.Errorf("failed to get officership for editOfficer: %+v", err)
		}

		tempStartDate := c.FormValue("startDate")
		tempEndDate := c.FormValue("endDate")

		if tempStartDate == "" {
			return errors.New("start date cannot be blank")
		}

		parsedStart, err := time.Parse("02/01/2006", tempStartDate)
		if err != nil {
			return errors.Errorf("failed to parse start date: %+v", err)
		}

		diff := time.Now().Compare(parsedStart)
		if diff != 1 {
			return errors.New("start date must be before today")
		}

		endDate := null.NewTime(time.Time{}, false)

		if tempEndDate != "" {
			var parsedEnd time.Time

			parsedEnd, err = time.Parse("02/01/2006", tempEndDate)
			if err != nil {
				return errors.Errorf("failed to parse end date: %+v", err)
			}

			endDate = null.TimeFrom(parsedEnd)
		}

		officer1.OfficerID = officershipID
		officer1.UserID = userID
		officer1.StartDate = null.TimeFrom(parsedStart)
		officer1.EndDate = endDate

		_, err = v.officership.EditOfficershipMember(c.Request().Context(), officer1)
		if err != nil {
			return errors.Errorf("failed to edit officer for editOfficer: %+v", err)
		}

		return c.Redirect(http.StatusFound, fmt.Sprintf("/internal/officership/officer/%d", officerID))
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficerDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		officerID, err := strconv.Atoi(c.Param("officerid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest,
				errors.Errorf("failed to parse officerid for officer delete: %+v", err))
		}

		officer, err := v.officership.GetOfficershipMember(c.Request().Context(),
			officership.OfficershipMember{OfficershipMemberID: officerID})
		if err != nil {
			return errors.Errorf("failed to get officer for officer delete: %+v", err)
		}

		err = v.officership.DeleteOfficershipMember(c.Request().Context(), officer)
		if err != nil {
			return errors.Errorf("failed to delete officer for officer delete: %+v", err)
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
			return errors.Errorf("failed to get officershipTeams: %+v", err)
		}

		p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			return errors.Errorf("failed to get user permissions for officershipTeams: %+v", err)
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

func (v *Views) OfficershipTeamAddOfficershipFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		teamID, err := strconv.Atoi(c.Param("officershipteamid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, errors.Errorf("failed to get team id for officership team add officership: %+v", err))
		}

		_, err = v.officership.GetOfficershipTeam(c.Request().Context(), officership.OfficershipTeam{TeamID: teamID})
		if err != nil {
			return errors.Errorf("failed to get team for officership team add officership: %+v", err)
		}

		officershipID, err := strconv.Atoi(c.Request().FormValue("officershipID"))
		if err != nil {
			return errors.Errorf("failed to get officershipid for officership team add officership: %+v", err)
		}

		_, err = v.officership.GetOfficership(c.Request().Context(), officership.Officership{OfficershipID: officershipID})
		if err != nil {
			return errors.Errorf("failed to get officership for officership team add officership: %+v", err)
		}

		memberLevel := c.FormValue("memberLevel")
		var isLeader, isDeputy bool

		switch memberLevel {
		case "leader":
			isLeader = true
		case "deputy":
			isDeputy = true
		}

		officershipTeamMember := officership.OfficershipTeamMember{
			TeamID:    teamID,
			OfficerID: officershipID,
			IsLeader:  isLeader,
			IsDeputy:  isDeputy,
		}

		_, err = v.officership.GetOfficershipTeamMember(c.Request().Context(), officershipTeamMember)
		if err == nil {
			return errors.New("failed to add officership team member for officership team add officership: row already exists")
		}

		_, err = v.officership.AddOfficershipTeamMember(c.Request().Context(), officershipTeamMember)
		if err != nil {
			return errors.Errorf("failed to add officership team member for officership team add officership: %+v", err)
		}

		return c.Redirect(http.StatusFound, fmt.Sprintf("/internal/officership/team/%d", teamID))
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficershipTeamRemoveOfficershipFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		teamID, err := strconv.Atoi(c.Param("officershipteamid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, errors.Errorf("failed to get teamid for officership team remove officership: %+v", err))
		}

		_, err = v.officership.GetOfficershipTeam(c.Request().Context(), officership.OfficershipTeam{TeamID: teamID})
		if err != nil {
			return errors.Errorf("failed to get team for officership team remove officership: %+v", err)
		}

		officershipID, err := strconv.Atoi(c.Param("officershipid"))
		if err != nil {
			return errors.Errorf("failed to get officershipid for officership team remove officership: %+v", err)
		}

		_, err = v.officership.GetOfficership(c.Request().Context(), officership.Officership{OfficershipID: officershipID})
		if err != nil {
			return errors.Errorf("failed to get officership for officership team remove officership: %+v", err)
		}

		officershipTeamMember := officership.OfficershipTeamMember{
			TeamID:    teamID,
			OfficerID: officershipID,
		}

		_, err = v.officership.GetOfficershipTeamMember(c.Request().Context(), officershipTeamMember)
		if err != nil {
			return errors.Errorf("failed to get officership team member for officership team remove officership: %+v", err)
		}

		err = v.officership.DeleteOfficershipTeamMember(c.Request().Context(), officershipTeamMember)
		if err != nil {
			return errors.Errorf("failed to remove officership team member for officership team remove officership: %+v", err)
		}

		return c.Redirect(http.StatusFound, fmt.Sprintf("/internal/officership/team/%d", teamID))
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficershipTeamFunc(c echo.Context) error {
	if c.Request().Method == http.MethodGet {
		c1 := v.getSessionData(c)

		officershipTeamID, err := strconv.Atoi(c.Param("officershipteamid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest,
				errors.Errorf("failed to parse officershipteamid for officership team: %+v", err))
		}

		officershipTeam, err := v.officership.GetOfficershipTeam(c.Request().Context(),
			officership.OfficershipTeam{TeamID: officershipTeamID})
		if err != nil {
			return errors.Errorf("failed to get officershipTeam: %+v", err)
		}

		permissions, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			return errors.Errorf("failed to get user permissions for officershipTeam: %+v", err)
		}

		teamMembers, err := v.officership.GetOfficershipTeamMembers(c.Request().Context(), &officershipTeam,
			officership.Any)
		if err != nil {
			return errors.Errorf("failed to get officership team members for officershipTeam: %+v", err)
		}

		officershipsNotInTeam, err := v.officership.GetOfficershipsNotInTeam(c.Request().Context(), officershipTeam)
		if err != nil {
			return errors.Errorf("failed to get officerships not in team: %+v", err)
		}

		data := struct {
			OfficershipTeam struct {
				TeamID                int
				Name                  string
				EmailAlias            string
				ShortDescription      string
				FullDescription       string
				TeamMembers           []officership.OfficershipTeamMember
				OfficershipsNotInTeam []officership.Officership
			}
			TemplateHelper
		}{
			OfficershipTeam: struct {
				TeamID                int
				Name                  string
				EmailAlias            string
				ShortDescription      string
				FullDescription       string
				TeamMembers           []officership.OfficershipTeamMember
				OfficershipsNotInTeam []officership.Officership
			}{
				TeamID:                officershipTeam.TeamID,
				Name:                  officershipTeam.Name,
				EmailAlias:            officershipTeam.EmailAlias,
				ShortDescription:      officershipTeam.ShortDescription,
				FullDescription:       officershipTeam.FullDescription,
				TeamMembers:           teamMembers,
				OfficershipsNotInTeam: officershipsNotInTeam,
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
	if c.Request().Method == http.MethodPost {
		name := c.Request().FormValue("name")
		emailAlias := c.Request().FormValue("emailAlias")
		shortDescription := c.Request().FormValue("shortDescription")
		fullDescription := c.Request().FormValue("fullDescription")

		t1, err := v.officership.GetOfficershipTeam(c.Request().Context(),
			officership.OfficershipTeam{TeamID: 0, Name: name})
		if err == nil && t1.TeamID > 0 {
			return errors.Errorf("officership team with name \"%s\" already exists", name)
		}

		_, err = v.officership.AddOfficershipTeam(c.Request().Context(),
			officership.OfficershipTeam{
				Name:             name,
				EmailAlias:       emailAlias,
				ShortDescription: shortDescription,
				FullDescription:  fullDescription,
			})
		if err != nil {
			return errors.Errorf("failed to add team for addOfficershipTeam: %+v", err)
		}

		return c.Redirect(http.StatusFound, "/internal/officership/teams")
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficershipTeamEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		officershipTeamID, err := strconv.Atoi(c.Param("officershipteamid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest,
				errors.Errorf("failed to parse officershipteamid for editOfficershipTeam: %+v", err))
		}

		team1, err := v.officership.GetOfficershipTeam(c.Request().Context(),
			officership.OfficershipTeam{TeamID: officershipTeamID})
		if err != nil {
			return errors.Errorf("failed to get team for editOfficershipTeam: %+v", err)
		}

		name := c.FormValue("name")
		emailAlias := c.FormValue("emailAlias")
		shortDescription := c.FormValue("shortDescription")
		fullDescription := c.FormValue("fullDescription")

		if len(name) > 0 {
			team1.Name = name
		}

		if len(emailAlias) > 0 {
			team1.EmailAlias = emailAlias
		}

		if len(shortDescription) > 0 {
			team1.ShortDescription = shortDescription
		}

		if len(fullDescription) > 0 {
			team1.FullDescription = fullDescription
		}

		_, err = v.officership.EditOfficershipTeam(c.Request().Context(), team1)
		if err != nil {
			return errors.Errorf("failed to edit team for editOfficershipTeam: %+v", err)
		}

		return c.Redirect(http.StatusFound, fmt.Sprintf("/internal/officership/team/%d", officershipTeamID))
	}

	return v.invalidMethodUsed(c)
}

func (v *Views) OfficershipTeamDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		teamID, err := strconv.Atoi(c.Param("officershipteamid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest,
				errors.Errorf("failed to parse teamid for officership team delete: %+v", err))
		}

		team, err := v.officership.GetOfficershipTeam(c.Request().Context(),
			officership.OfficershipTeam{TeamID: teamID})
		if err != nil {
			return errors.Errorf("failed to get team for officer team delete: %+v", err)
		}

		err = v.officership.RemoveTeamForOfficershipTeamMembers(c.Request().Context(), team)
		if err != nil {
			return errors.Errorf("failed to remove officerships from team for officership team delete: %+v", err)
		}

		err = v.officership.DeleteOfficershipTeam(c.Request().Context(), team)
		if err != nil {
			return errors.Errorf("failed to delete officership team for officership team delete: %+v", err)
		}

		return c.Redirect(http.StatusFound, "/internal/officership/teams")
	}

	return v.invalidMethodUsed(c)
}
