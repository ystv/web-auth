package views

import (
	"context"
	"github.com/ystv/web-auth/permission/permissions"
	"github.com/ystv/web-auth/public/templates"
	"github.com/ystv/web-auth/user"
	"log"
	"net/http"

	"github.com/ystv/web-auth/helpers"
)

// RequiresLogin is a middleware which will be used for each
// httpHandler to check if there is any active session
func (v *Views) RequiresLogin(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := v.cookie.Get(r, v.conf.SessionCookieName)
		if err != nil {
			log.Println(err)
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		if !helpers.GetUser(session).Authenticated {
			// Not authenticated
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		h.ServeHTTP(w, r)
	}
}

// RequiresMinimumPermission is a middleware that will
// ensure that the user has the given permission.
func (v *Views) RequiresMinimumPermission(h http.Handler, p permissions.Permissions) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := v.cookie.Get(r, v.conf.SessionCookieName)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		u := helpers.GetUser(session)

		perms, err := v.user.GetPermissionsForUser(r.Context(), u)
		if err != nil {
			log.Println(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		acceptedPerms := GetValidPermissions(p)

		for _, perm := range perms {
			if acceptedPerms[perm] {
				h.ServeHTTP(w, r)
				return
			}
		}

		err = v.template.RenderNoNavsTemplate(w, nil, templates.Forbidden500Template)
		if err != nil {
			log.Println(err)
		}
		w.WriteHeader(http.StatusForbidden)
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
		if m[perm] {
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
