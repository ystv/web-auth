package templates

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"log"
	"time"

	permission1 "github.com/ystv/web-auth/infrastructure/permission"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/permission/permissions"
	"github.com/ystv/web-auth/role"
	"github.com/ystv/web-auth/user"
	"gopkg.in/guregu/null.v4"
)

// tmpls are the storage of templates in the executable
//
//go:embed *.tmpl
var tmpls embed.FS

type Templater struct {
	Permission *permission.Store
	Role       *role.Store
	User       *user.Store
}

type Template string

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
	SignupEmailTemplate  Template = "signupEmail.tmpl"
	PermissionTemplate   Template = "permission.tmpl"
	ManageAPITemplate    Template = "manageAPI.tmpl"
	UserAddTemplate      Template = "userAdd.tmpl"
)

type TemplateType int

const (
	NoNavType TemplateType = iota
	PaginationType
	RegularType
)

// NewTemplate returns the template format to be used
func NewTemplate(p *permission.Store, r *role.Store, u *user.Store) *Templater {
	return &Templater{
		Permission: p,
		Role:       r,
		User:       u,
	}
}

// String returns the string equivalent of Template
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

func (t *Templater) GetEmailTemplate(emailTemplate Template) (*template.Template, error) {
	return template.New(emailTemplate.String()).ParseFS(tmpls, emailTemplate.String())
}

// getFuncMaps returns all the in built functions that templates can use
func (t *Templater) getFuncMaps() template.FuncMap {
	return template.FuncMap{
		"thisYear": func() int {
			return time.Now().Year()
		},
		"inc": func(a int) int {
			return a + 1
		},
		"dec": func(a int) int {
			return a - 1
		},
		"checkPermission": func(perms []permission.Permission, p string) bool {
			m := permission1.SufficientPermissionsFor(permissions.Permissions(p))

			for _, perm := range perms {
				if m[perm.Name] {
					return true
				}
			}
			return false
		},
		"getUserModifierField": func(u user.User, atTime null.String, prefix string) template.HTML {
			var s string
			if u.UserID != -1 {
				if len(u.Firstname) == 0 && len(u.Nickname) == 0 && len(u.Lastname) == 0 {
					s = fmt.Sprintf("%s by UNKNOWN(%d) at %s<br>", template.HTMLEscapeString(prefix), u.UserID, template.HTMLEscapeString(atTime.String))
				} else {
					name := formatUserName(u)
					s = fmt.Sprintf("%s by <a href=\"/internal/user/%d\">%s</a> at %s<br>", template.HTMLEscapeString(prefix), u.UserID, template.HTMLEscapeString(name), template.HTMLEscapeString(atTime.String))
				}
			} else if atTime.Valid {
				s = fmt.Sprintf("%s by UNKNOWN at %s<br>", prefix, atTime.String)
			}
			// #nosec
			return template.HTML(s)
		},
		"formatUserName": func(u user.DetailedUser) (name string) {
			return formatUserName(user.User{
				Firstname: u.Firstname,
				Nickname:  u.Nickname,
				Lastname:  u.Lastname,
			})
		},
		"formatUserNameUserStruct": formatUserName,
	}
}

func formatUserName(u user.User) (name string) {
	if u.Firstname != u.Nickname {
		name = fmt.Sprintf("%s (%s) %s", u.Firstname, u.Nickname, u.Lastname)
	} else {
		name = fmt.Sprintf("%s %s", u.Firstname, u.Lastname)
	}
	return name
}
