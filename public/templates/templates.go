package templates

import (
	"context"
	"embed"
	"fmt"
	infraPermission "github.com/ystv/web-auth/infrastructure/permission"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/permission/permissions"
	"github.com/ystv/web-auth/role"
	"github.com/ystv/web-auth/user"
	"html/template"
	"io"
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
	Template string
)

const (
	Forbidden500Template Template = "500Forbidden.tmpl"
	ForgotTemplate       Template = "forgot.tmpl"
	NotFound404Template  Template = "404NotFound.tmpl"
	ForgotEmailTemplate  Template = "forgotEmail.tmpl"
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
	PermissionsTemplate  Template = "permissions.tmpl"
)

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
	t1.Funcs(template.FuncMap{
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
		"checkPermission": func(id int, p string) bool {
			m := infraPermission.GetValidPermissions(permissions.Permissions(p))

			u, err := t.User.GetUser(context.Background(), user.User{UserID: id})
			if err != nil {
				fmt.Println(err)
				return false
			}

			p1, err := t.User.GetPermissionsForUser(context.Background(), u)
			if err != nil {
				fmt.Println(err)
				return false
			}

			for _, perm := range p1 {
				if m[perm] {
					return true
				}
			}
			return false
		},
	})

	t1, err = t1.ParseFS(tmpls, "_base.tmpl", "_head.tmpl", "_footer.tmpl", "_navbar.tmpl", "_sidebar.tmpl", mainTmpl.GetString())
	if err != nil {
		fmt.Println(err)
		return err
	}

	return t1.Execute(w, data)
}

func (t *Templater) RenderTemplatePagination(w io.Writer, data interface{}, mainTmpl Template) error {
	var err error

	t1 := template.New("_base.tmpl")
	t1.Funcs(template.FuncMap{
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
		"memberAddPermission": func(id int) bool {
			u, err := t.User.GetUser(context.Background(), user.User{UserID: id})
			if err != nil {
				fmt.Println(err)
				return false
			}

			p, err := t.User.GetPermissionsForUser(context.Background(), u)
			if err != nil {
				fmt.Println(err)
				return false
			}

			for _, perm := range p {
				if perm == permissions.ManageMembersMembersAdd.GetString() {
					return true
				}
			}
			return false
		},
		"checkPermission": func(id int, p string) bool {
			m := infraPermission.GetValidPermissions(permissions.Permissions(p))

			u, err := t.User.GetUser(context.Background(), user.User{UserID: id})
			if err != nil {
				fmt.Println(err)
				return false
			}

			p1, err := t.User.GetPermissionsForUser(context.Background(), u)
			if err != nil {
				fmt.Println(err)
				return false
			}

			for _, perm := range p1 {
				if m[perm] {
					return true
				}
			}
			return false
		},
	})

	t1, err = t1.ParseFS(tmpls, "_base.tmpl", "_head.tmpl", "_footer.tmpl", "_navbar.tmpl", "_sidebar.tmpl", "_pagination.tmpl", mainTmpl.GetString())
	if err != nil {
		fmt.Println(err)
		return err
	}

	return t1.Execute(w, data)
}

func (t *Templater) RenderNoNavsTemplate(w io.Writer, data interface{}, mainTmpl Template) error {
	var err error

	t1 := template.New("_baseNoNavs.tmpl")
	t1.Funcs(template.FuncMap{
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
		"even": func(a int) bool {
			return a%2 == 0
		},
		"incUInt64": func(a uint64) uint64 {
			return a + 1
		},
	})

	t1, err = t1.ParseFS(tmpls, "_baseNoNavs.tmpl", "_head.tmpl", "_footer.tmpl", string(mainTmpl))
	if err != nil {
		fmt.Println(err)
		return err
	}

	return t1.Execute(w, data)
}

func (t *Templater) RenderEmail(emailTemplate Template) *template.Template {
	return template.Must(template.New("forgotEmail.tmpl").ParseFS(tmpls, emailTemplate.GetString()))
}
