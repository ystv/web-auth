package templates

import (
	"context"
	"embed"
	"fmt"
	"github.com/ystv/web-auth/api"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/permission/permissions"
	"github.com/ystv/web-auth/role"
	"github.com/ystv/web-auth/user"
	"html/template"
	"io"
	"log"
	"strings"
	"time"
)

//go:embed *.tmpl
var tmpls embed.FS

type (
	Repo interface {
		RenderTemplate(w io.Writer, data interface{}, mainTmpl Template) error
		RenderTemplatePagination(w io.Writer, data interface{}, mainTmpl Template) error
		RenderNoNavsTemplate(w io.Writer, data interface{}, mainTmpl Template) error
		RenderEmail(emailTemplate Template) *template.Template
		getFuncMaps() template.FuncMap
		permissionsParser(id int, p string) bool
	}

	Templater struct {
		Permission *permission.Store
		Role       *role.Store
		User       *user.Store
	}

	Template string
)

const (
	Forbidden500Template        Template = "500Forbidden.tmpl"
	ForgotTemplate              Template = "forgot.tmpl"
	NotFound404Template         Template = "404NotFound.tmpl"
	ForgotPasswordEmailTemplate Template = "forgotPasswordEmail.tmpl"
	InternalTemplate            Template = "internal.tmpl"
	LoginTemplate               Template = "login.tmpl"
	NotificationTemplate        Template = "notification.tmpl"
	ResetTemplate               Template = "reset.tmpl"
	ErrorTemplate               Template = "error.tmpl"
	SettingsTemplate            Template = "settings.tmpl"
	SignupTemplate              Template = "signup.tmpl"
	SignupEmailTemplate         Template = "signupEmail.tmpl"
	UserTemplate                Template = "user.tmpl"
	UsersTemplate               Template = "users.tmpl"
	RolesTemplate               Template = "roles.tmpl"
	RoleTemplate                Template = "role.tmpl"
	ResetPasswordEmailTemplate  Template = "resetPasswordEmail.tmpl"
	PermissionsTemplate         Template = "permissions.tmpl"
	PermissionTemplate          Template = "permission.tmpl"
	ManageAPITemplate           Template = "manageAPI.tmpl"
)

var _ Repo = &Templater{}

func NewTemplate(p *permission.Store, r *role.Store, u *user.Store) *Templater {
	return &Templater{
		Permission: p,
		Role:       r,
		User:       u,
	}
}

func (t Template) GetString() string {
	return string(t)
}

func (t *Templater) RenderTemplate(w io.Writer, data interface{}, mainTmpl Template) error {
	var err error

	t1 := template.New("_base.tmpl")
	t1.Funcs(t.getFuncMaps())

	t1, err = t1.ParseFS(tmpls, "_base.tmpl", "_head.tmpl", "_footer.tmpl", "_navbar.tmpl", "_sidebar.tmpl", mainTmpl.GetString())
	if err != nil {
		log.Printf("failed to get templates for template(RenderTemplate): %+v", err)
		return err
	}

	return t1.Execute(w, data)
}

func (t *Templater) RenderTemplatePagination(w io.Writer, data interface{}, mainTmpl Template) error {
	var err error

	t1 := template.New("_base.tmpl")
	t1.Funcs(t.getFuncMaps())

	t1, err = t1.ParseFS(tmpls, "_base.tmpl", "_head.tmpl", "_footer.tmpl", "_navbar.tmpl", "_sidebar.tmpl", "_pagination.tmpl", mainTmpl.GetString())
	if err != nil {
		log.Printf("failed to get templates for template(RenderTemplatePagination): %+v", err)
		return err
	}

	return t1.Execute(w, data)
}

func (t *Templater) RenderNoNavsTemplate(w io.Writer, data interface{}, mainTmpl Template) error {
	var err error

	t1 := template.New("_baseNoNavs.tmpl")
	t1.Funcs(t.getFuncMaps())

	t1, err = t1.ParseFS(tmpls, "_baseNoNavs.tmpl", "_head.tmpl", "_footer.tmpl", string(mainTmpl))
	if err != nil {
		log.Printf("failed to get templates for template(RenderNoNavsTemplate): %+v", err)
		return err
	}

	return t1.Execute(w, data)
}

func (t *Templater) RenderEmail(emailTemplate Template) *template.Template {
	return template.Must(template.New(emailTemplate.GetString()).ParseFS(tmpls, emailTemplate.GetString()))
}

func (t *Templater) getFuncMaps() template.FuncMap {
	return template.FuncMap{
		"formatDuration": func(d time.Duration) string {
			days := int64(d.Hours()) / 24
			hours := int64(d.Hours()) % 24
			minutes := int64(d.Minutes()) % 60
			seconds := int64(d.Seconds()) % 60

			segments := []struct {
				name  string
				value int64
			}{
				{"Day", days},
				{"Hour", hours},
				{"Min", minutes},
				{"Sec", seconds},
			}

			var parts []string

			for _, s := range segments {
				if s.value == 0 {
					continue
				}
				plural := ""
				if s.value != 1 {
					plural = "s"
				}

				parts = append(parts, fmt.Sprintf("%d %s%s", s.value, s.name, plural))
			}
			return strings.Join(parts, " ")
		},
		"formatTime": func(fmt string, t time.Time) string {
			return t.Format(fmt)
		},
		"now": func() time.Time {
			return time.Now()
		},
		"thisYear": func() int {
			return time.Now().Year()
		},
		"add": func(a, b int) int {
			return a + b
		},
		"inc": func(a int) int {
			return a + 1
		},
		"dec": func(a int) int {
			return a - 1
		},
		"even": func(a int) bool {
			return a%2 == 0
		},
		"incUInt64": func(a uint64) uint64 {
			return a + 1
		},
		"len": func(a string) int {
			return len(a)
		},
		"lenA": func(a []string) int {
			return len(a)
		},
		"lenP": func(a []permission.Permission) int {
			return len(a)
		},
		"lenR": func(a []role.Role) int {
			return len(a)
		},
		"lenU": func(a []user.User) int {
			return len(a)
		},
		"checkPermission": func(id int, p string) bool {
			return t.permissionsParser(id, p)
		},
		"parseHTMLManageAPI": func(tokens []api.Token) template.HTML {
			var output, tokenBuilder strings.Builder
			if len(tokens) > 0 {
				tokenBuilder.WriteString("Current tokens: <div class=\"toolbar\"><ol>")
				for _, token := range tokens {
					if len(token.Description) > 0 {
						tokenBuilder.WriteString(fmt.Sprintf("<li style='list-style-type: none;'><span class='tab'></span>%[1]s - %[2]s&emsp;<a class=\"button is-danger is-outlined\" onclick=\"removeToken%[3]sFromAPIModal()\"><span class=\"mdi mdi-key-minus\"></span>&ensp;Remove token</a></li>", token.Name, token.Description, strings.ReplaceAll(token.TokenID, "-", "")))
					} else {
						tokenBuilder.WriteString(fmt.Sprintf("<li style='list-style-type: none;'><span class='tab'></span>%[1]s&emsp;<a class=\"button is-danger is-outlined\" onclick=\"removeToken%[2]sFromAPIModal()\"><span class=\"mdi mdi-key-minus\"></span>&ensp;Remove token</a></li>", token.Name, strings.ReplaceAll(token.TokenID, "-", "")))
					}
					tokenBuilder.WriteString(fmt.Sprintf("<div id=\"removeToken%[1]sFromAPIModal\" class=\"modal\">\n        <div class=\"modal-background\"></div>\n        <div class=\"modal-content\">\n            <div class=\"box\">\n                <article class=\"media\">\n                    <div class=\"media-content\">\n                        <div class=\"content\">\n                            <p class=\"title\">Are you sure you want to remove \"%[2]s\" token?</p>\n                            <p><strong>This cannot be undone</strong></p>\n                            <form action=\"/internal/api/manage/%[3]s/delete\" method=\"post\">\n                                <button class=\"button is-danger\">Remove token</button>\n                            </form>\n                        </div>\n                    </div>\n                </article>\n            </div>\n        </div>\n        <button class=\"modal-close is-large\" aria-label=\"close\"></button>\n    </div><script>function removeToken%[1]sFromAPIModal() {\n            document.getElementById(\"removeToken%[1]sFromAPIModal\").classList.add(\"is-active\");\n        }</script>", strings.ReplaceAll(token.TokenID, "-", ""), token.Name, token.TokenID))
				}
				tokenBuilder.WriteString("</ol></div>")
			}
			output.WriteString(fmt.Sprintf("<p>%s</p>", tokenBuilder.String()))
			return template.HTML(output.String())
		},
		"parseHTMLPermissions": func(perms []permission.Permission) template.HTML {
			var output strings.Builder
			for _, p := range perms {
				output.WriteString(fmt.Sprintf("<tr><th>%d</th><td>%s</td><td>%s</td><td>%d</td><td><a href='/internal/permission/%d'>View</a></td></tr>", p.PermissionID, p.Name, p.Description, p.Roles, p.PermissionID))
			}
			return template.HTML(output.String())
		},
		"parseHTMLPermission": func(p user.PermissionTemplate, userID int) template.HTML {
			roleAdmin := t.permissionsParser(userID, permissions.ManageMembersGroup.GetString())
			var output, roles strings.Builder
			if len(p.Roles) > 0 {
				roles.WriteString("Inherited by: <ol>")
				for _, r := range p.Roles {
					if roleAdmin {
						roles.WriteString(fmt.Sprintf("<li style='list-style-type: none;'><span class='tab'></span><a href='/internal/role/%d'>%s</a></li>", r.RoleID, r.Name))
					} else {
						roles.WriteString(fmt.Sprintf("<li style='list-style-type: none;'><span class='tab'></span>%s</li>", r.Name))
					}
				}
				roles.WriteString("</ol>")
			}
			output.WriteString(fmt.Sprintf("<p>Permission ID: %d<br>Name: %s<br>Description: %s<br><br>%s</p>", p.PermissionID, p.Name, p.Description, roles.String()))
			return template.HTML(output.String())
		},
		"parseHTMLRoles": func(roles []role.Role) template.HTML {
			var output strings.Builder
			for _, r := range roles {
				output.WriteString(fmt.Sprintf("<tr><th>%d</th><td>%s</td><td>%s</td><td>%d</td><td>%d</td><td><a href='/internal/role/%d'>View</a></td></tr>", r.RoleID, r.Name, r.Description, r.Users, r.Permissions, r.RoleID))
			}
			return template.HTML(output.String())
		},
		"parseHTMLRole": func(r user.RoleTemplate, permissionsNotInRole []permission.Permission, usersNotInRole []user.User, userID int) template.HTML {
			permissionAdmin := t.permissionsParser(userID, permissions.ManageMembersPermissions.GetString())
			membersList := t.permissionsParser(userID, permissions.ManageMembersMembersList.GetString())
			var output, perms, users, permissionsToAdd, usersToAdd strings.Builder
			if len(r.Permissions) > 0 {
				perms.WriteString("Permissions: <div class=\"toolbar\"><ol>")
				for _, p := range r.Permissions {
					if permissionAdmin {
						perms.WriteString(fmt.Sprintf("<li style='list-style-type: none;'><span class='tab'></span><a href='/internal/permission/%[1]d'>%[2]s</a>&emsp;<a class=\"button is-danger is-outlined\" onclick=\"removePermission%[1]dFromRoleModal()\"><span class=\"mdi mdi-key-minus\"></span>&ensp;Remove permission</a></li>", p.PermissionID, p.Name))
					} else {
						perms.WriteString(fmt.Sprintf("<li style='list-style-type: none;'><span class='tab'></span>%[1]s&emsp;<a class=\"button is-danger is-outlined\" onclick=\"removePermission%[2]dFromRoleModal()\"><span class=\"mdi mdi-key-minus\"></span>&ensp;Remove permission</a><</li>", p.Name, p.PermissionID))
					}
					perms.WriteString(fmt.Sprintf("<div id=\"removePermission%[1]dFromRoleModal\" class=\"modal\">\n        <div class=\"modal-background\"></div>\n        <div class=\"modal-content\">\n            <div class=\"box\">\n                <article class=\"media\">\n                    <div class=\"media-content\">\n                        <div class=\"content\">\n                            <p class=\"title\">Are you sure you want to remove \"%[2]s\" from this role?</p>\n                            <p><strong>This can be undone</strong></p>\n                            <form action=\"/internal/role/%[3]d/permission/remove/%[1]d\" method=\"post\">\n                                <button class=\"button is-danger\">Remove permission</button>\n                            </form>\n                        </div>\n                    </div>\n                </article>\n            </div>\n        </div>\n        <button class=\"modal-close is-large\" aria-label=\"close\"></button>\n    </div><script>function removePermission%[1]dFromRoleModal() {\n            document.getElementById(\"removePermission%[1]dFromRoleModal\").classList.add(\"is-active\");\n        }</script>", p.PermissionID, p.Name, r.RoleID))
				}
				perms.WriteString("</ol></div><br>")
			}
			if len(r.Users) > 0 {
				users.WriteString("<br>Inherited by: <div class=\"toolbar\"><ol>")
				for _, u := range r.Users {
					var name string
					if u.Firstname != u.Nickname && len(u.Nickname) > 0 {
						name = fmt.Sprintf("%s (%s) %s", u.Firstname, u.Nickname, u.Lastname)
					} else {
						name = fmt.Sprintf("%s %s", u.Firstname, u.Lastname)
					}
					if membersList {
						users.WriteString(fmt.Sprintf("<li style='list-style-type: none;'><span class='tab'></span><a href='/internal/user/%[1]d'>%[2]s</a>&emsp;<a class=\"button is-danger is-outlined\" onclick=\"removeUser%[1]dFromRoleModal()\"><span class=\"mdi mdi-account-minus\"></span>&ensp;Remove user</a></li>", u.UserID, name))
					} else {
						users.WriteString(fmt.Sprintf("<li style='list-style-type: none;'><span class='tab'></span>%[1]s&emsp;<a class=\"button is-danger is-outlined\" onclick=\"removeUser%[2]dFromRoleModal()\"><span class=\"mdi mdi-account-minus\"></span>&ensp;Remove user</a></li>", name, u.UserID))
					}
					users.WriteString(fmt.Sprintf("<div id=\"removeUser%[1]dFromRoleModal\" class=\"modal\">\n        <div class=\"modal-background\"></div>\n        <div class=\"modal-content\">\n            <div class=\"box\">\n                <article class=\"media\">\n                    <div class=\"media-content\">\n                        <div class=\"content\">\n                            <p class=\"title\">Are you sure you want to remove \"%[2]s\" from this role?</p>\n                            <p><strong>This can be undone</strong></p>\n                            <form action=\"/internal/role/%[3]d/user/remove/%[1]d\" method=\"post\">\n                                <button class=\"button is-danger\">Remove user</button>\n                            </form>\n                        </div>\n                    </div>\n                </article>\n            </div>\n        </div>\n        <button class=\"modal-close is-large\" aria-label=\"close\"></button>\n    </div><script>function removeUser%[1]dFromRoleModal() {\n            document.getElementById(\"removeUser%[1]dFromRoleModal\").classList.add(\"is-active\");\n        }</script>", u.UserID, name, r.RoleID))
				}
				users.WriteString("</ol></div><br>")
			}
			if len(permissionsNotInRole) > 0 {
				permissionsToAdd.WriteString("Use the drop down below to add more permissions to this role.<br>")
				permissionsToAdd.WriteString(fmt.Sprintf("<form method=\"post\" action=\"/internal/role/%d/permission/add\">", r.RoleID))
				permissionsToAdd.WriteString("<div class=\"select\"><select id=\"permission\" name=\"permission\">")
				permissionsToAdd.WriteString("<option value disabled selected>Please select</option>")
				for _, p := range permissionsNotInRole {
					permissionsToAdd.WriteString(fmt.Sprintf("<option value=\"%d\">%s</option>", p.PermissionID, p.Name))
				}
				permissionsToAdd.WriteString("</select></div><br>")
				permissionsToAdd.WriteString("<button class=\"button is-info\" style=\"margin-top: 10px\"><span class=\"mdi mdi-key-plus\"></span>&ensp;Add permission</button></form>")
			}
			if len(usersNotInRole) > 0 {
				usersToAdd.WriteString("Use the drop down below to add more users to this role.<br>")
				usersToAdd.WriteString(fmt.Sprintf("<form method=\"post\" action=\"/internal/role/%d/user/add\">", r.RoleID))
				usersToAdd.WriteString("<div class=\"select\"><select id=\"user\" name=\"user\">")
				usersToAdd.WriteString("<option value disabled selected>Please select</option>")
				for _, u := range usersNotInRole {
					var name string
					if u.Firstname != u.Nickname && len(u.Nickname) > 0 {
						name = fmt.Sprintf("%s (%s) %s", u.Firstname, u.Nickname, u.Lastname)
					} else {
						name = fmt.Sprintf("%s %s", u.Firstname, u.Lastname)
					}
					usersToAdd.WriteString(fmt.Sprintf("<option value=\"%d\">%s</option>", u.UserID, name))
				}
				usersToAdd.WriteString("</select></div><br>")
				usersToAdd.WriteString("<button class=\"button is-info\" style=\"margin-top: 10px\"><span class=\"mdi mdi-account-plus\"></span>&ensp;Add user</button></form>")
			}
			output.WriteString(fmt.Sprintf("<p>Role ID: %d<br>Name: %s<br>Description: %s<br><br>%s%s%s%s</p>", r.RoleID, r.Name, r.Description, perms.String(), permissionsToAdd.String(), users.String(), usersToAdd.String()))
			return template.HTML(output.String())
		},
		"parseHTMLUsers": func(tmplUsers []user.StrippedUser, userID int) template.HTML {
			//defer t.timer("tmplUser")()
			memberAdmin := t.permissionsParser(userID, permissions.ManageMembersMembersAdmin.GetString())
			var output strings.Builder
			for _, tmplUser := range tmplUsers {
				var enabled, deleted, ifView string
				if tmplUser.Enabled {
					enabled = "Enabled"
				} else {
					enabled = "Disabled"
				}
				if tmplUser.Deleted {
					deleted = "Deleted"
				} else {
					deleted = "-"
				}
				if memberAdmin {
					ifView = fmt.Sprintf("<td><a href=\"/internal/user/%d\">View</a></td>", tmplUser.UserID)
				}
				output.WriteString(fmt.Sprintf("<tr><td>%d</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td>%s</tr>", tmplUser.UserID, tmplUser.Name, tmplUser.Username, tmplUser.Email, enabled, deleted, tmplUser.LastLogin, ifView))
			}
			return template.HTML(output.String())
		},
		"parseHTMLUser": func(u user.DetailedUser, userID int) template.HTML {
			permissionAdmin := t.permissionsParser(userID, permissions.ManageMembersPermissions.GetString())
			roleAdmin := t.permissionsParser(userID, permissions.ManageMembersGroup.GetString())
			var output, perms, roles strings.Builder
			var deleted, enabled, ldap, avatar, lastLogin, created, updated, deletedBy string
			if len(u.Permissions) > 0 {
				perms.WriteString("Permissions: <ol>")
				for _, p := range u.Permissions {
					if permissionAdmin {
						perms.WriteString(fmt.Sprintf("<li style='list-style-type: none;'><span class='tab'></span><a href='/internal/permission/%d'>%s</a></li>", p.PermissionID, p.Name))
					} else {
						perms.WriteString(fmt.Sprintf("<li style='list-style-type: none;'><span class='tab'></span>%s</li>", p.Name))
					}
				}
				perms.WriteString("</ol><br>")
			}
			if len(u.Roles) > 0 {
				roles.WriteString("Roles: <ol>")
				for _, r := range u.Roles {
					if roleAdmin {
						roles.WriteString(fmt.Sprintf("<li style='list-style-type: none;'><span class='tab'></span><a href='/internal/role/%d'>%s</a></li>", r.RoleID, r.Name))
					} else {
						roles.WriteString(fmt.Sprintf("<li style='list-style-type: none;'><span class='tab'></span>%s</li>", r.Name))
					}
				}
				roles.WriteString("</ol><br>")
			}
			if u.DeletedBy.UserID != -1 {
				deleted = "<h2 class='subtitle'><strong>Deleted!</strong></h2>"
			}
			if u.Enabled {
				enabled = "enabled"
			} else {
				enabled = "disabled"
			}
			if u.LDAPUsername.Valid {
				ldap = fmt.Sprintf("LDAP (Active Directory) username: %s<br>", u.LDAPUsername.String)
			}
			if u.UseGravatar {
				avatar = "Using gravatar"
			} else {
				avatar = "Using local file"
			}
			if u.LastLogin.Valid {
				lastLogin = fmt.Sprintf("Last login at %s<br>", u.LastLogin.String)
			}
			if u.CreatedBy.UserID != -1 {
				if len(u.CreatedBy.Firstname) == 0 && len(u.CreatedBy.Nickname) == 0 && len(u.CreatedBy.Lastname) == 0 {
					created = fmt.Sprintf("Created by UNKNOWN(%d) at %s<br>", u.CreatedBy.UserID, u.CreatedAt.String)
				} else {
					var name string
					if u.CreatedBy.Firstname != u.CreatedBy.Nickname && len(u.CreatedBy.Nickname) > 0 {
						name = fmt.Sprintf("%s (%s) %s", u.CreatedBy.Firstname, u.CreatedBy.Nickname, u.CreatedBy.Lastname)
					} else {
						name = fmt.Sprintf("%s %s", u.CreatedBy.Firstname, u.CreatedBy.Lastname)
					}
					created = fmt.Sprintf("Created by <a href='/internal/user/%d'>%s</a> at %s<br>", u.CreatedBy.UserID, name, u.CreatedAt.String)
				}
			} else if u.CreatedAt.Valid {
				created = fmt.Sprintf("Created by UNKNOWN at %s<br>", u.CreatedAt.String)
			}
			if u.UpdatedBy.UserID != -1 {
				if len(u.UpdatedBy.Firstname) == 0 && len(u.UpdatedBy.Nickname) == 0 && len(u.UpdatedBy.Lastname) == 0 {
					updated = fmt.Sprintf("Updated by UNKNOWN(%d) at %s<br>", u.UpdatedBy.UserID, u.UpdatedAt.String)
				} else {
					var name string
					if u.UpdatedBy.Firstname != u.UpdatedBy.Nickname && len(u.UpdatedBy.Nickname) > 0 {
						name = fmt.Sprintf("%s (%s) %s", u.UpdatedBy.Firstname, u.UpdatedBy.Nickname, u.UpdatedBy.Lastname)
					} else {
						name = fmt.Sprintf("%s %s", u.UpdatedBy.Firstname, u.UpdatedBy.Lastname)
					}
					updated = fmt.Sprintf("Updated by <a href='/internal/user/%d'>%s</a> at %s<br>", u.UpdatedBy.UserID, name, u.UpdatedAt.String)
				}
			} else if u.UpdatedAt.Valid {
				updated = fmt.Sprintf("Updated by UNKNOWN at %s<br>", u.UpdatedAt.String)
			}
			if u.DeletedBy.UserID != -1 {
				if len(u.DeletedBy.Firstname) == 0 && len(u.DeletedBy.Nickname) == 0 && len(u.DeletedBy.Lastname) == 0 {
					deleted = fmt.Sprintf("Deleted by UNKNOWN(%d) at %s<br>", u.DeletedBy.UserID, u.DeletedAt.String)
				} else {
					var name string
					if u.DeletedBy.Firstname != u.DeletedBy.Nickname && len(u.DeletedBy.Nickname) > 0 {
						name = fmt.Sprintf("%s (%s) %s", u.DeletedBy.Firstname, u.DeletedBy.Nickname, u.DeletedBy.Lastname)
					} else {
						name = fmt.Sprintf("%s %s", u.DeletedBy.Firstname, u.DeletedBy.Lastname)
					}
					deleted = fmt.Sprintf("Deleted by <a href='/internal/user/%d'>%s</a> at %s<br>", u.DeletedBy.UserID, name, u.DeletedAt.String)
				}
			} else if u.DeletedAt.Valid {
				deleted = fmt.Sprintf("Deleted by UNKNOWN at %s<br>", u.DeletedAt.String)
			}
			output.WriteString(fmt.Sprintf("<p>%sUser ID: %d<br>First name: %s<br>Nickname: %s<br>Last name: %s<br>Username: %s<br>Email: %s<br><br>Enabled: %s<br>Login type: %s<br>%sAvartar source: %s<br><br>%s%s%s%s%s%s</p>", deleted, u.UserID, u.Firstname, u.Nickname, u.Lastname, u.Username, u.Email, enabled, u.LoginType, ldap, avatar, perms.String(), roles.String(), lastLogin, created, updated, deletedBy))
			return template.HTML(output.String())
		},
	}
}

func (t *Templater) permissionsParser(id int, p string) bool {
	m := GetValidPermissions(permissions.Permissions(p))

	u, err := t.User.GetUser(context.Background(), user.User{UserID: id})
	if err != nil {
		log.Printf("failed to get user for template(permissionParser): %+v", err)
		return false
	}

	p1, err := t.User.GetPermissionsForUser(context.Background(), u)
	if err != nil {
		log.Printf("failed to get permission for template(permissionParser): %+v", err)
		return false
	}

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