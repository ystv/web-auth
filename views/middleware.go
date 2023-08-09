package views

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/permission/permissions"
	"github.com/ystv/web-auth/user"
	"log"
	"net/http"
)

// RequiresLogin is a middleware which will be used for each
// httpHandler to check if there is any active session
func (v *Views) RequiresLogin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			log.Println(err)
			return v.LoginFunc(c)
		}

		c1 := v.getData(session)

		user2, err := v.user.GetUser(c.Request().Context(), c1.User)
		if err != nil {
			log.Println(err)
			return err
		}
		if user2.DeletedBy.Valid || !c1.User.Enabled {
			session.Values["user"] = &user.User{}
			session.Options.MaxAge = -1
			err = session.Save(c.Request(), c.Response())
			if err != nil {
				//http.Error(w, err.Error(), http.StatusInternalServerError)
				return v.errorHandle(c, err)
			}
			return c.Redirect(http.StatusFound, "/")
		}
		if !c1.User.Authenticated {
			// Not authenticated
			return c.Redirect(http.StatusFound, "/")
		}
		return next(c)
	}
}

// RequiresMinimumPermission is a middleware that will
// ensure that the user has the given permission.
func (v *Views) RequiresMinimumPermission(next echo.HandlerFunc, p permissions.Permissions) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			log.Println(err)
			http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
			return v.LoginFunc(c)
		}

		c1 := v.getData(session)

		perms, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			log.Println(err)
			http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
			return v.LoginFunc(c)
		}

		acceptedPerms := GetValidPermissions(p)

		for _, perm := range perms {
			if acceptedPerms[perm.Name] {
				return next(c)
			}
		}

		c.Response().WriteHeader(http.StatusForbidden)
		return v.Error500(c)
	}
}

// RequiresMinimumPermissionMMP is a middleware that will
// ensure that the user has ManageMembersPermissions.
func (v *Views) RequiresMinimumPermissionMMP(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			log.Println(err)
			http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
			return v.LoginFunc(c)
		}

		c1 := v.getData(session)

		perms, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			log.Println(err)
			http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
			return v.LoginFunc(c)
		}

		acceptedPerms := GetValidPermissions(permissions.ManageMembersPermissions)

		for _, perm := range perms {
			if acceptedPerms[perm.Name] {
				return next(c)
			}
		}

		c.Response().WriteHeader(http.StatusForbidden)
		return v.Error500(c)
	}
}

// RequiresMinimumPermissionMMG is a middleware that will
// ensure that the user has ManageMembersGroup.
func (v *Views) RequiresMinimumPermissionMMG(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			log.Println(err)
			http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
			return v.LoginFunc(c)
		}

		c1 := v.getData(session)

		perms, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			log.Println(err)
			http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
			return v.LoginFunc(c)
		}

		acceptedPerms := GetValidPermissions(permissions.ManageMembersGroup)

		for _, perm := range perms {
			if acceptedPerms[perm.Name] {
				return next(c)
			}
		}

		c.Response().WriteHeader(http.StatusForbidden)
		return v.Error500(c)
	}
}

// RequiresMinimumPermissionMMML is a middleware that will
// ensure that the user has ManageMembersMembersList.
func (v *Views) RequiresMinimumPermissionMMML(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			log.Println(err)
			http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
			return v.LoginFunc(c)
		}

		c1 := v.getData(session)

		perms, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			log.Println(err)
			http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
			return v.LoginFunc(c)
		}

		acceptedPerms := GetValidPermissions(permissions.ManageMembersMembersList)

		for _, perm := range perms {
			if acceptedPerms[perm.Name] {
				return next(c)
			}
		}

		c.Response().WriteHeader(http.StatusForbidden)
		return v.Error500(c)
	}
}

// RequiresMinimumPermissionMMAdd is a middleware that will
// ensure that the user has ManageMembersMembersAdd.
func (v *Views) RequiresMinimumPermissionMMAdd(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			log.Println(err)
			http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
			return v.LoginFunc(c)
		}

		c1 := v.getData(session)

		perms, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			log.Println(err)
			http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
			return v.LoginFunc(c)
		}

		acceptedPerms := GetValidPermissions(permissions.ManageMembersMembersAdd)

		for _, perm := range perms {
			if acceptedPerms[perm.Name] {
				return next(c)
			}
		}

		c.Response().WriteHeader(http.StatusForbidden)
		return v.Error500(c)
	}
}

// RequiresMinimumPermissionMMAdmin is a middleware that will
// ensure that the user has ManageMembersMembersAdmin.
func (v *Views) RequiresMinimumPermissionMMAdmin(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		session, err := v.cookie.Get(c.Request(), v.conf.SessionCookieName)
		if err != nil {
			log.Println(err)
			http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
			return v.LoginFunc(c)
		}

		c1 := v.getData(session)

		perms, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
		if err != nil {
			log.Println(err)
			http.Error(c.Response(), err.Error(), http.StatusInternalServerError)
			return v.LoginFunc(c)
		}

		acceptedPerms := GetValidPermissions(permissions.ManageMembersMembersAdmin)

		for _, perm := range perms {
			if acceptedPerms[perm.Name] {
				return next(c)
			}
		}

		c.Response().WriteHeader(http.StatusForbidden)
		return v.Error500(c)
	}
}

func (v *Views) RequiresMinimumPermissionNoHttp(userID int, p permissions.Permissions) bool {
	u, err := v.user.GetUser(context.Background(), user.User{UserID: userID})
	if err != nil {
		log.Println(err)
		return false
	}

	p1, err := v.user.GetPermissionsForUser(context.Background(), u)
	if err != nil {
		log.Println(err)
		return false
	}

	m := GetValidPermissions(p)

	for _, perm := range p1 {
		if m[perm.Name] {
			return true
		}
	}

	return false
}

func GetValidPermissions(p permissions.Permissions) (m map[string]bool) {
	m = make(map[string]bool)
	if p.GetString() == permissions.MenuDisabled.GetString() {
		m[p.GetString()] = true
		return
	}

	m[p.GetString()] = true

	switch p {
	case permissions.ManageMembersAdmin:
	case permissions.KeyCardAccess:
	case permissions.BookingsAdmin:
	case permissions.CalendarAdmin:
	case permissions.CMSAdmin:
	case permissions.Cobra:
	case permissions.Director:
	case permissions.EditNetStats:
	case permissions.EmailEveryone:
	case permissions.EquipmentAdmin:
	case permissions.HiresAdmin:
	case permissions.Inform:
	case permissions.KeyListManage:
	case permissions.MailingListAdmin:
	case permissions.OfficerReports:
	case permissions.Streamer:
	case permissions.TechieTodo:
	case permissions.VideoStats:
	case permissions.WatchAdmin:
		break
	case permissions.ManageMembersMembersList:
	case permissions.ManageMembersMembersAdd:
		m[permissions.ManageMembersMembersAdmin.GetString()] = true
	case permissions.ManageMembersPermissions:
	case permissions.ManageMembersMicsKeyList:
	case permissions.ManageMembersMiscUnpaidList:
	case permissions.ManageMembersOfficers:
	case permissions.ManageMembersGroup:
	case permissions.ManageMembersMembersAdmin:
		m[permissions.ManageMembersAdmin.GetString()] = true
		break
	case permissions.EmailAccess:
	case permissions.EmailAlumni:
	case permissions.EmailOfficers:
		m[permissions.EmailEveryone.GetString()] = true
		break
	case permissions.CalendarSocialCreator:
		m[permissions.CalendarSocialAdmin.GetString()] = true
	case permissions.CalendarSocialAdmin:
		m[permissions.CalendarAdmin.GetString()] = true
		break
	case permissions.CalendarShowCreator:
		m[permissions.CalendarShowAdmin.GetString()] = true
	case permissions.CalendarShowAdmin:
		m[permissions.CalendarAdmin.GetString()] = true
		break
	case permissions.CalendarMeetingCreator:
		m[permissions.CalendarMeetingAdmin.GetString()] = true
	case permissions.CalendarMeetingAdmin:
		m[permissions.CalendarAdmin.GetString()] = true
		break
	case permissions.CMSNewsItemCreator:
		m[permissions.CMSNewsItemAdmin.GetString()] = true
	case permissions.CMSNewsItemAdmin:
		m[permissions.CMSNewsAdmin.GetString()] = true
	case permissions.CMSEndboardAdmin:
	case permissions.CMSView:
	case permissions.CMSPermalinkAdmin:
	case permissions.CMSNewsAdmin:
		m[permissions.CMSAdmin.GetString()] = true
		break
	case permissions.CMSNewsCreator:
		m[permissions.CMSNewsAdmin.GetString()] = true
		m[permissions.CMSAdmin.GetString()] = true
		break
	case permissions.CMSPageCreator:
		m[permissions.CMSPageAdmin.GetString()] = true
	case permissions.CMSPageAdmin:
		m[permissions.CMSAdmin.GetString()] = true
		break
	case permissions.CMSSlideshowCreator:
		m[permissions.CMSSlideshowAdmin.GetString()] = true
	case permissions.CMSSlideshowAdmin:
		m[permissions.CMSAdmin.GetString()] = true
		break
	}

	m[permissions.SuperUser.GetString()] = true
	return
}
