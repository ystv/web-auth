package permission

import "github.com/ystv/web-auth/permission/permissions"

// SufficientPermissionsFor takes a permission for a task and returns that permission and higher permissions that would be acceptable
func SufficientPermissionsFor(p permissions.Permissions) (m map[string]bool) {
	m = make(map[string]bool)

	m[p.String()] = true

	switch p {
	case permissions.MenuDisabled:
		m[permissions.MenuDisabled.String()] = true
		return m
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
		m[permissions.ManageMembersMembersAdmin.String()] = true
	case permissions.ManageMembersPermissions:
	case permissions.ManageMembersMicsKeyList:
	case permissions.ManageMembersMiscUnpaidList:
	case permissions.ManageMembersOfficers:
	case permissions.ManageMembersGroup:
	case permissions.ManageMembersMembersAdmin:
		m[permissions.ManageMembersAdmin.String()] = true
		break
	case permissions.EmailAccess:
	case permissions.EmailAlumni:
	case permissions.EmailOfficers:
		m[permissions.EmailEveryone.String()] = true
		break
	case permissions.CalendarSocialCreator:
		m[permissions.CalendarSocialAdmin.String()] = true
	case permissions.CalendarSocialAdmin:
		m[permissions.CalendarAdmin.String()] = true
		break
	case permissions.CalendarShowCreator:
		m[permissions.CalendarShowAdmin.String()] = true
	case permissions.CalendarShowAdmin:
		m[permissions.CalendarAdmin.String()] = true
		break
	case permissions.CalendarMeetingCreator:
		m[permissions.CalendarMeetingAdmin.String()] = true
	case permissions.CalendarMeetingAdmin:
		m[permissions.CalendarAdmin.String()] = true
		break
	case permissions.CMSNewsItemCreator:
		m[permissions.CMSNewsItemAdmin.String()] = true
	case permissions.CMSNewsItemAdmin:
		m[permissions.CMSNewsAdmin.String()] = true
	case permissions.CMSEndboardAdmin:
	case permissions.CMSView:
	case permissions.CMSPermalinkAdmin:
	case permissions.CMSNewsAdmin:
		m[permissions.CMSAdmin.String()] = true
		break
	case permissions.CMSNewsCreator:
		m[permissions.CMSNewsAdmin.String()] = true
		m[permissions.CMSAdmin.String()] = true
		break
	case permissions.CMSPageCreator:
		m[permissions.CMSPageAdmin.String()] = true
	case permissions.CMSPageAdmin:
		m[permissions.CMSAdmin.String()] = true
		break
	case permissions.CMSSlideshowCreator:
		m[permissions.CMSSlideshowAdmin.String()] = true
	case permissions.CMSSlideshowAdmin:
		m[permissions.CMSAdmin.String()] = true
		break
	}

	m[permissions.SuperUser.String()] = true
	return
}
