package views

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/role"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
	"log"
	"net/http"
	"strconv"
)

type (
	RolesTemplate struct {
		Roles           []role.Role
		UserPermissions []permission.Permission
		ActivePage      string
	}

	RoleTemplate struct {
		Role            user.RoleTemplate
		UserPermissions []permission.Permission
		ActivePage      string
	}
)

func (v *Views) bindRoleToTemplate(r1 role.Role) user.RoleTemplate {
	var r user.RoleTemplate
	r.RoleID = r1.RoleID
	r.Name = r1.Name
	r.Description = r1.Description
	return r
}

// RolesFunc handles a roles request
func (v *Views) RolesFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	roles, err := v.role.GetRoles(c.Request().Context())
	if err != nil {
		log.Printf("failed to get roles for roles: %+v", err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get roles for roles: %w", err))
		}
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		log.Printf("failed to get user permissions for roles: %+v", err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get user permissions for roles: %+v", err))
		}
	}

	data := RolesTemplate{
		Roles:           roles,
		UserPermissions: p1,
		ActivePage:      "roles",
	}

	return v.template.RenderTemplate(c.Response(), data, templates.RolesTemplate, templates.RegularType)
}

func (v *Views) RoleFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	roleID, err := strconv.Atoi(c.Param("roleid"))
	if err != nil {
		log.Printf("failed to get roleid for role: %+v", err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get roleid for role: %w", err))
		}
	}

	role1, err := v.role.GetRole(c.Request().Context(), role.Role{RoleID: roleID})
	if err != nil {
		log.Printf("failed to get role for role: %+v", err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get role for role: %w", err))
		}
	}

	roleTemplate := v.bindRoleToTemplate(role1)

	roleTemplate.Permissions, err = v.user.GetPermissionsForRole(c.Request().Context(), role1)
	if err != nil {
		log.Printf("failed to get permissions for role: %+v", err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to permissions for role: %w", err))
		}
	}

	roleTemplate.Users, err = v.user.GetUsersForRole(c.Request().Context(), role1)
	if err != nil {
		log.Printf("failed to get users for role: %+v", err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get users for role: %w", err))
		}
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		log.Printf("failed to get user permissions for role: %+v", err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get user permissions for role: %+v", err))
		}
	}

	data := RoleTemplate{
		Role:            roleTemplate,
		UserPermissions: p1,
		ActivePage:      "role",
	}

	return v.template.RenderTemplate(c.Response(), data, templates.RoleTemplate, templates.RegularType)
}

func (v *Views) RoleAddFunc(c echo.Context) error {
	return nil
}

func (v *Views) RoleEditFunc(c echo.Context) error {
	return nil
}

func (v *Views) RoleDeleteFunc(c echo.Context) error {
	return nil
}
