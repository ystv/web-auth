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
	case permissions.ManageMembersMembersList:
		fallthrough
	case permissions.ManageMembersMembersAdd:
		m[permissions.ManageMembersMembersAdmin.String()] = true
		fallthrough
	case permissions.ManageMembersPermissions:
		fallthrough
	case permissions.ManageMembersMicsKeyList:
		fallthrough
	case permissions.ManageMembersMiscUnpaidList:
		fallthrough
	case permissions.ManageMembersOfficers:
		fallthrough
	case permissions.ManageMembersGroup:
		fallthrough
	case permissions.ManageMembersMembersAdmin:
		m[permissions.ManageMembersAdmin.String()] = true
		break
	case permissions.EmailAccess:
		fallthrough
	case permissions.EmailAlumni:
		fallthrough
	case permissions.EmailOfficers:
		m[permissions.EmailEveryone.String()] = true
		break
	case permissions.CalendarSocialCreator:
		m[permissions.CalendarSocialAdmin.String()] = true
		fallthrough
	case permissions.CalendarSocialAdmin:
		m[permissions.CalendarAdmin.String()] = true
		break
	case permissions.CalendarShowCreator:
		m[permissions.CalendarShowAdmin.String()] = true
		fallthrough
	case permissions.CalendarShowAdmin:
		m[permissions.CalendarAdmin.String()] = true
		break
	case permissions.CalendarMeetingCreator:
		m[permissions.CalendarMeetingAdmin.String()] = true
		fallthrough
	case permissions.CalendarMeetingAdmin:
		m[permissions.CalendarAdmin.String()] = true
		break
	case permissions.CMSNewsItemCreator:
		m[permissions.CMSNewsItemAdmin.String()] = true
		fallthrough
	case permissions.CMSNewsItemAdmin:
		m[permissions.CMSNewsAdmin.String()] = true
		fallthrough
	case permissions.CMSEndboardAdmin:
		fallthrough
	case permissions.CMSView:
		fallthrough
	case permissions.CMSPermalinkAdmin:
		fallthrough
	case permissions.CMSNewsAdmin:
		m[permissions.CMSAdmin.String()] = true
		break
	case permissions.CMSNewsCreator:
		m[permissions.CMSNewsAdmin.String()] = true
		m[permissions.CMSAdmin.String()] = true
		break
	case permissions.CMSPageCreator:
		m[permissions.CMSPageAdmin.String()] = true
		fallthrough
	case permissions.CMSPageAdmin:
		m[permissions.CMSAdmin.String()] = true
		break
	case permissions.CMSSlideshowCreator:
		m[permissions.CMSSlideshowAdmin.String()] = true
		fallthrough
	case permissions.CMSSlideshowAdmin:
		m[permissions.CMSAdmin.String()] = true
		break
	}

	m[permissions.SuperUser.String()] = true
	return
}
