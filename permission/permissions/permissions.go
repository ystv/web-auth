package permissions

type Permissions string

//nolint:gochecknoglobals
var (
	KeyCardAccess               Permissions = "Access.Keycard.Station"
	BookingsAdmin               Permissions = "BookingsAdmin"
	CalendarAdmin               Permissions = "Calendar.Admin"
	CalendarMeetingAdmin        Permissions = "Calendar.Meeting.Admin"
	CalendarMeetingCreator      Permissions = "Calendar.Meeting.Creator"
	CalendarShowAdmin           Permissions = "Calendar.Show.Admin"
	CalendarShowCreator         Permissions = "Calendar.Show.Creator"
	CalendarSocialAdmin         Permissions = "Calendar.Social.Admin"
	CalendarSocialCreator       Permissions = "Calendar.Social.Creator"
	CMSAdmin                    Permissions = "CMS.Admin"
	CMSEndboardAdmin            Permissions = "CMS.EndboardAdmin"
	CMSNewsAdmin                Permissions = "CMS.News.Admin"
	CMSNewsCreator              Permissions = "CMS.News.Creator"
	CMSNewsItemAdmin            Permissions = "CMS.News.Item.Admin"
	CMSNewsItemCreator          Permissions = "CMS.News.Item.Creator"
	CMSPageAdmin                Permissions = "CMS.Page.Admin"
	CMSPageCreator              Permissions = "CMS.Page.Creator"
	CMSPermalinkAdmin           Permissions = "CMS.Permalink.Admin"
	CMSSlideshowAdmin           Permissions = "CMS.Slideshow.Admin"
	CMSSlideshowCreator         Permissions = "CMS.Slideshow.Creator"
	CMSView                     Permissions = "CMS.View"
	Cobra                       Permissions = "COBRA"
	Director                    Permissions = "Director"
	EditNetStats                Permissions = "EditNetStats"
	EmailAccess                 Permissions = "Email.Access"
	EmailAlumni                 Permissions = "Email.Alumni"
	EmailEveryone               Permissions = "Email.Everyone"
	EmailOfficers               Permissions = "Email.Officers"
	EquipmentAdmin              Permissions = "EquipmentAdmin"
	HiresAdmin                  Permissions = "HiresAdmin"
	Inform                      Permissions = "Inform"
	KeyListManage               Permissions = "KeyList.Manage"
	MailingListAdmin            Permissions = "MailingList.Admin"
	ManageMembersAdmin          Permissions = "ManageMembers.Admin"
	ManageMembersGroup          Permissions = "ManageMembers.Groups"
	ManageMembersMembersAdd     Permissions = "ManageMembers.Members.Add"
	ManageMembersMembersAdmin   Permissions = "ManageMembers.Members.Admin"
	ManageMembersMembersList    Permissions = "ManageMembers.Members.List"
	ManageMembersMicsKeyList    Permissions = "ManageMembers.Misc.KeyList"
	ManageMembersMiscUnpaidList Permissions = "ManageMembers.Misc.UnpaidList"
	ManageMembersOfficers       Permissions = "ManageMembers.Officers"
	ManageMembersPermissions    Permissions = "ManageMembers.Permissions"
	MenuDisabled                Permissions = "Menu.Disabled"
	OfficerReports              Permissions = "OfficerReports"
	Streamer                    Permissions = "Streamer"
	SuperUser                   Permissions = "SuperUser"
	TechieTodo                  Permissions = "TechieTodo"
	VideoStats                  Permissions = "VideoStats"
	WatchAdmin                  Permissions = "Watch.Admin"
)

// String gets the string value for a Permission
func (p Permissions) String() string {
	return string(p)
}
