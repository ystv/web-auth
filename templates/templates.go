package templates

import (
	"embed"
	"fmt"
	permission1 "github.com/ystv/web-auth/infrastructure/permission"
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
	Templater struct {
		Permission *permission.Store
		Role       *role.Store
		User       *user.Store
	}

	Template     string
	TemplateType int
)

const (
	ForgotTemplate       Template = "forgot.tmpl"
	NotFound404Template  Template = "404NotFound.tmpl"
	ForgotEmailTemplate  Template = "forgotEmail.tmpl" // generated by go generate
	InternalTemplate     Template = "internal.tmpl"
	LoginTemplate        Template = "login.tmpl"
	NotificationTemplate Template = "notification.tmpl"
	ResetTemplate        Template = "reset.tmpl"
	ErrorTemplate        Template = "error.tmpl"
	SettingsTemplate     Template = "settings.tmpl"
	SignupTemplate       Template = "signup.tmpl"
	UserTemplate         Template = "user.tmpl"
	UsersTemplate        Template = "users.tmpl"
	RolesTemplate        Template = "roles.tmpl"
	RoleTemplate         Template = "role.tmpl"
	ResetEmailTemplate   Template = "resetEmail.tmpl" // generated by go generate
	PermissionsTemplate  Template = "permissions.tmpl"
	PermissionTemplate   Template = "permission.tmpl"

	NoNavType      TemplateType = 0
	PaginationType TemplateType = 1
	RegularType    TemplateType = 2
)

func NewTemplate(p *permission.Store, r *role.Store, u *user.Store) *Templater {
	return &Templater{
		Permission: p,
		Role:       r,
		User:       u,
	}
}

func (t Template) String() string {
	return string(t)
}

func (t *Templater) RenderTemplate(w io.Writer, data interface{}, mainTmpl Template, templateType TemplateType) error {
	var err error

	t1 := template.New("_base.tmpl")

	t1.Funcs(t.getFuncMaps())

	switch templateType {
	case NoNavType:
		t1, err = t1.ParseFS(tmpls, "_base.tmpl", "_bodyNoNavs.tmpl", "_head.tmpl", "_footer.tmpl", mainTmpl.String())
	case PaginationType:
		t1, err = t1.ParseFS(tmpls, "_base.tmpl", "_body.tmpl", "_head.tmpl", "_footer.tmpl", "_navbar.tmpl", "_sidebar.tmpl", "_pagination.tmpl", mainTmpl.String())
	case RegularType:
		t1, err = t1.ParseFS(tmpls, "_base.tmpl", "_body.tmpl", "_head.tmpl", "_footer.tmpl", "_navbar.tmpl", "_sidebar.tmpl", mainTmpl.String())
	default:
		return fmt.Errorf("unable to parse template, invalid type: %d", templateType)
	}

	if err != nil {
		log.Printf("failed to get templates for template(RenderTemplate): %+v", err)
		return err
	}

	return t1.Execute(w, data)
}

func (t *Templater) RenderEmail(emailTemplate Template) *template.Template {
	return template.Must(template.New(emailTemplate.String()).ParseFS(tmpls, emailTemplate.String()))
}

func (t *Templater) getFuncMaps() template.FuncMap {
	return template.FuncMap{
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
		"checkPermission": func(perms []permission.Permission, p string) bool {
			return t.permissionsParser(perms, p)
		},
		"parsePermissionsIntoHTML": func(perms []permission.Permission) template.HTML {
			var output strings.Builder
			for _, p := range perms {
				output.WriteString("<tr>")
				output.WriteString(fmt.Sprintf("<th>%d</th>", p.PermissionID))
				output.WriteString(fmt.Sprintf("<td>%s</td>", p.Name))
				output.WriteString(fmt.Sprintf("<td>%s</td>", p.Description))
				output.WriteString(fmt.Sprintf("<td>%d</td>", p.Roles))
				output.WriteString(fmt.Sprintf("<td><a href='/internal/permission/%d'>View</a></td>", p.PermissionID))
				output.WriteString("</tr>")
			}
			return template.HTML(output.String())
		},
		"parsePermissionIntoHTML": func(p user.PermissionTemplate, perms []permission.Permission) template.HTML {
			roleAdmin := t.permissionsParser(perms, permissions.ManageMembersGroup.String())
			var output, roles strings.Builder
			if len(p.Roles) > 0 {
				roles.WriteString("Inherited by: <ol>")
				for _, r := range p.Roles {
					if roleAdmin {
						roles.WriteString("<li style='list-style-type: none;'><span class='tab'></span>")
						roles.WriteString(fmt.Sprintf("<a href='/internal/role/%d'>%s</a>", r.RoleID, r.Name))
						roles.WriteString("</li>")
					} else {
						roles.WriteString("<li style='list-style-type: none;'><span class='tab'></span>")
						roles.WriteString(fmt.Sprintf("%s", r.Name))
						roles.WriteString("</li>")
					}
				}
				roles.WriteString("</ol>")
			}
			output.WriteString("<p>")
			output.WriteString(fmt.Sprintf("Permission ID: %d<br>", p.PermissionID))
			output.WriteString(fmt.Sprintf("Name: %s<br>", p.Name))
			output.WriteString(fmt.Sprintf("Description: %s<br><br>", p.Description))
			output.WriteString(fmt.Sprintf("%s", roles.String()))
			output.WriteString("</p>")
			return template.HTML(output.String())
		},
		"parseRolesIntoHTML": func(roles []role.Role) template.HTML {
			var output strings.Builder
			for _, r := range roles {
				output.WriteString("<tr>")
				output.WriteString(fmt.Sprintf("<th>%d</th>", r.RoleID))
				output.WriteString(fmt.Sprintf("<td>%s</td>", r.Name))
				output.WriteString(fmt.Sprintf("<td>%s</td>", r.Description))
				output.WriteString(fmt.Sprintf("<td>%d</td>", r.Users))
				output.WriteString(fmt.Sprintf("<td>%d</td>", r.Permissions))
				output.WriteString(fmt.Sprintf("<td><a href='/internal/role/%d'>View</a></td>", r.RoleID))
				output.WriteString("</tr>")
			}
			return template.HTML(output.String())
		},
		"parseRoleIntoHTML": func(r user.RoleTemplate, p1 []permission.Permission) template.HTML {
			permissionAdmin := t.permissionsParser(p1, permissions.ManageMembersPermissions.String())
			membersList := t.permissionsParser(p1, permissions.ManageMembersMembersList.String())
			var output, perms, users strings.Builder
			if len(r.Permissions) > 0 {
				perms.WriteString("Permissions: <ol>")
				for _, p := range r.Permissions {
					if permissionAdmin {
						perms.WriteString("<li style='list-style-type: none;'><span class='tab'></span>")
						perms.WriteString(fmt.Sprintf("<a href='/internal/permission/%d'>%s</a>", p.PermissionID, p.Name))
						perms.WriteString("</li>")
					} else {
						perms.WriteString("<li style='list-style-type: none;'><span class='tab'></span>")
						perms.WriteString(fmt.Sprintf("%s", p.Name))
						perms.WriteString("</li>")
					}
				}
				perms.WriteString("</ol><br>")
			}
			if len(r.Users) > 0 {
				users.WriteString("Inherited by: <ol>")
				for _, u := range r.Users {
					if membersList {
						users.WriteString("<li style='list-style-type: none;'><span class='tab'></span>")
						users.WriteString(fmt.Sprintf("<a href='/internal/user/%d'>%s</a>", u.UserID, u.Firstname+" "+u.Lastname))
						users.WriteString("</li>")
					} else {
						users.WriteString("<li style='list-style-type: none;'><span class='tab'></span>")
						users.WriteString(fmt.Sprintf("%s", u.Firstname+" "+u.Lastname))
						users.WriteString("</li>")
					}
				}
				users.WriteString("</ol>")
			}
			output.WriteString("<p>")
			output.WriteString(fmt.Sprintf("Role ID: %d<br>", r.RoleID))
			output.WriteString(fmt.Sprintf("Name: %s<br>", r.Name))
			output.WriteString(fmt.Sprintf("Description: %s<br><br>", r.Description))
			output.WriteString(perms.String())
			output.WriteString(users.String())
			output.WriteString("</p>")
			return template.HTML(output.String())
		},
		"parseUsersIntoHTML": func(tmplUsers []user.StrippedUser, perms []permission.Permission) template.HTML {
			memberAdmin := t.permissionsParser(perms, permissions.ManageMembersMembersAdmin.String())
			var output strings.Builder
			for _, tmplUser := range tmplUsers {
				var ifView strings.Builder
				var enabled, deleted string
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
					ifView.WriteString("<td>")
					ifView.WriteString(fmt.Sprintf("<a href=\"/internal/user/%d\">View</a>", tmplUser.UserID))
					ifView.WriteString("</td>")
				}
				output.WriteString("<tr>")
				output.WriteString(fmt.Sprintf("<td>%d</td>", tmplUser.UserID))
				output.WriteString(fmt.Sprintf("<td>%s</td>", tmplUser.Name))
				output.WriteString(fmt.Sprintf("<td>%s</td>", tmplUser.Username))
				output.WriteString(fmt.Sprintf("<td>%s</td>", tmplUser.Email))
				output.WriteString(fmt.Sprintf("<td>%s</td>", enabled))
				output.WriteString(fmt.Sprintf("<td>%s</td>", deleted))
				output.WriteString(fmt.Sprintf("<td>%s</td>", tmplUser.LastLogin))
				output.WriteString(ifView.String())
				output.WriteString("</tr>")
			}
			return template.HTML(output.String())
		},
		"parseUserIntoHTML": func(u user.DetailedUser, p1 []permission.Permission) template.HTML {
			permissionAdmin := t.permissionsParser(p1, permissions.ManageMembersPermissions.String())
			roleAdmin := t.permissionsParser(p1, permissions.ManageMembersGroup.String())
			var output, perms, roles strings.Builder
			var deleted, enabled, ldap, avatar, lastLogin, created, updated, deletedBy string
			if len(u.Permissions) > 0 {
				perms.WriteString("Permissions: <ol>")
				for _, p := range u.Permissions {
					if permissionAdmin {
						perms.WriteString("<li style='list-style-type: none;'><span class='tab'></span>")
						perms.WriteString(fmt.Sprintf("<a href='/internal/permission/%d'>%s</a>", p.PermissionID, p.Name))
						perms.WriteString("</li>")
					} else {
						perms.WriteString("<li style='list-style-type: none;'><span class='tab'></span>")
						perms.WriteString(fmt.Sprintf("%s", p.Name))
						perms.WriteString("</li>")
					}
				}
				perms.WriteString("</ol><br>")
			}
			if len(u.Roles) > 0 {
				roles.WriteString("Roles: <ol>")
				for _, r := range u.Roles {
					if roleAdmin {
						roles.WriteString("<li style='list-style-type: none;'><span class='tab'></span>")
						roles.WriteString(fmt.Sprintf("<a href='/internal/role/%d'>%s</a>", r.RoleID, r.Name))
						roles.WriteString("</li>")
					} else {
						roles.WriteString("<li style='list-style-type: none;'><span class='tab'></span>")
						roles.WriteString(fmt.Sprintf("%s", r.Name))
						roles.WriteString("</li>")
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
					name := t.parseUserName(u.CreatedBy)
					created = fmt.Sprintf("Created by <a href='/internal/user/%d'>%s</a> at %s<br>", u.CreatedBy.UserID, name, u.CreatedAt.String)
				}
			} else if u.CreatedAt.Valid {
				created = fmt.Sprintf("Created by UNKNOWN at %s<br>", u.CreatedAt.String)
			}
			if u.UpdatedBy.UserID != -1 {
				if len(u.UpdatedBy.Firstname) == 0 && len(u.UpdatedBy.Nickname) == 0 && len(u.UpdatedBy.Lastname) == 0 {
					updated = fmt.Sprintf("Updated by UNKNOWN(%d) at %s<br>", u.UpdatedBy.UserID, u.UpdatedAt.String)
				} else {
					name := t.parseUserName(u.UpdatedBy)
					updated = fmt.Sprintf("Updated by <a href='/internal/user/%d'>%s</a> at %s<br>", u.UpdatedBy.UserID, name, u.UpdatedAt.String)
				}
			} else if u.UpdatedAt.Valid {
				updated = fmt.Sprintf("Updated by UNKNOWN at %s<br>", u.UpdatedAt.String)
			}
			if u.DeletedBy.UserID != -1 {
				if len(u.DeletedBy.Firstname) == 0 && len(u.DeletedBy.Nickname) == 0 && len(u.DeletedBy.Lastname) == 0 {
					deleted = fmt.Sprintf("Deleted by UNKNOWN(%d) at %s<br>", u.DeletedBy.UserID, u.DeletedAt.String)
				} else {
					name := t.parseUserName(u.DeletedBy)
					deleted = fmt.Sprintf("Deleted by <a href='/internal/user/%d'>%s</a> at %s<br>", u.DeletedBy.UserID, name, u.DeletedAt.String)
				}
			} else if u.DeletedAt.Valid {
				deleted = fmt.Sprintf("Deleted by UNKNOWN at %s<br>", u.DeletedAt.String)
			}
			output.WriteString("<p>")
			output.WriteString(deleted)
			output.WriteString(fmt.Sprintf("User ID: %d<br>", u.UserID))
			output.WriteString(fmt.Sprintf("First name: %s<br>", u.Firstname))
			output.WriteString(fmt.Sprintf("Nickname: %s<br>", u.Nickname))
			output.WriteString(fmt.Sprintf("Last name: %s<br>", u.Lastname))
			output.WriteString(fmt.Sprintf("Username: %s<br>", u.Username))
			output.WriteString(fmt.Sprintf("Email: %s<br><br>", u.Email))
			output.WriteString(fmt.Sprintf("Enabled: %s<br>", enabled))
			output.WriteString(fmt.Sprintf("Login type: %s<br>", u.LoginType))
			output.WriteString(ldap)
			output.WriteString(fmt.Sprintf("Avatar source: %s<br><br>", avatar))
			output.WriteString(perms.String())
			output.WriteString(roles.String())
			output.WriteString(lastLogin)
			output.WriteString(created)
			output.WriteString(updated)
			output.WriteString(deletedBy)
			output.WriteString("</p>")
			return template.HTML(output.String())
		},
	}
}

func (t *Templater) parseUserName(u user.User) (name string) {
	if u.Firstname != u.Nickname {
		name = fmt.Sprintf("%s (%s) %s", u.Firstname, u.Nickname, u.Lastname)
	} else {
		name = fmt.Sprintf("%s %s", u.Firstname, u.Lastname)
	}
	return name
}

func (t *Templater) permissionsParser(perms []permission.Permission, p string) bool {
	m := permission1.SufficientPermissionsFor(permissions.Permissions(p))

	for _, perm := range perms {
		if m[perm.Name] {
			return true
		}
	}
	return false
}
