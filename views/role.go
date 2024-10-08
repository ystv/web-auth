package views

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/role"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
)

type (
	// RolesTemplate is for the roles front end
	RolesTemplate struct {
		Roles []role.Role
		TemplateHelper
	}

	// RoleTemplate is for the role front end
	RoleTemplate struct {
		Role                 user.RoleTemplate
		PermissionsNotInRole []permission.Permission
		UsersNotInRole       []user.User
		TemplateHelper
	}
)

// bindRoleToTemplate converts from role.Role to user.RoleTemplate
func (v *Views) bindRoleToTemplate(r1 role.Role) user.RoleTemplate {
	return user.RoleTemplate{
		RoleID:      r1.RoleID,
		Name:        r1.Name,
		Description: r1.Description,
	}
}

// RolesFunc handles a roles request
func (v *Views) RolesFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	roles, err := v.role.GetRoles(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get roles for roles: %w", err)
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		return fmt.Errorf("failed to get user permissions for roles: %w", err)
	}

	data := RolesTemplate{
		Roles: roles,
		TemplateHelper: TemplateHelper{
			UserPermissions: p1,
			ActivePage:      "roles",
			Assumed:         c1.Assumed,
		},
	}

	return v.template.RenderTemplate(c.Response(), data, templates.RolesTemplate, templates.RegularType)
}

// RoleFunc handles a role request
func (v *Views) RoleFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	roleID, err := strconv.Atoi(c.Param("roleid"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to get roleid for role: %w", err))
	}

	role1, err := v.role.GetRole(c.Request().Context(), role.Role{RoleID: roleID})
	if err != nil {
		return fmt.Errorf("failed to get role for role: %w", err)
	}

	roleTemplate := v.bindRoleToTemplate(role1)

	roleTemplate.Permissions, err = v.user.GetPermissionsForRole(c.Request().Context(), role1)
	if err != nil {
		return fmt.Errorf("failed to permissions for role: %w", err)
	}

	roleTemplate.Users, err = v.user.GetUsersForRole(c.Request().Context(), role1)
	if err != nil {
		return fmt.Errorf("failed to get users for role: %w", err)
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		return fmt.Errorf("failed to get user permissions for role: %w", err)
	}

	permissions, err := v.user.GetPermissionsNotInRole(c.Request().Context(), role1)
	if err != nil {
		return fmt.Errorf("failed to get permissions not in role for role: %w", err)
	}

	users, err := v.user.GetUsersNotInRole(c.Request().Context(), role1)
	if err != nil {
		return fmt.Errorf("failed to get users not in role for role: %w", err)
	}

	data := RoleTemplate{
		Role:                 roleTemplate,
		PermissionsNotInRole: permissions,
		UsersNotInRole:       users,
		TemplateHelper: TemplateHelper{
			UserPermissions: p1,
			ActivePage:      "role",
			Assumed:         c1.Assumed,
		},
	}

	return v.template.RenderTemplate(c.Response(), data, templates.RoleTemplate, templates.RegularType)
}

// RoleAddFunc handles a role add request
func (v *Views) RoleAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		name := c.Request().FormValue("name")
		description := c.Request().FormValue("description")

		r1, err := v.role.GetRole(c.Request().Context(), role.Role{RoleID: 0, Name: name})
		if err == nil && r1.RoleID > 0 {
			return fmt.Errorf("role with name \"%s\" already exists", name)
		}

		_, err = v.role.AddRole(c.Request().Context(), role.Role{RoleID: -1, Name: name, Description: description})
		if err != nil {
			return fmt.Errorf("failed to add role for addrole: %w", err)
		}

		return c.Redirect(http.StatusFound, "/internal/roles")
	}

	return v.invalidMethodUsed(c)
}

// RoleEditFunc handles a role edit request
func (v *Views) RoleEditFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		roleID, err := strconv.Atoi(c.Param("roleid"))
		if err != nil {
			return fmt.Errorf("failed to get roleid for editRole: %w", err)
		}

		role1, err := v.role.GetRole(c.Request().Context(), role.Role{RoleID: roleID})
		if err != nil {
			return fmt.Errorf("failed to get role for editRole: %w", err)
		}

		err = c.Request().ParseForm()
		if err != nil {
			return fmt.Errorf("failed to parse form for roleEdit: %w", err)
		}

		name := c.Request().FormValue("name")
		description := c.Request().FormValue("description")

		if name != role1.Name && len(name) > 0 {
			role1.Name = name
		}

		if description != role1.Description && len(description) > 0 {
			role1.Description = description
		}

		_, err = v.role.EditRole(c.Request().Context(), role1)
		if err != nil {
			return fmt.Errorf("failed to edit role for editRole: %w", err)
		}

		return c.Redirect(http.StatusFound, fmt.Sprintf("/internal/role/%d", roleID))
	}

	return v.invalidMethodUsed(c)
}

// RoleDeleteFunc handles a role delete request
func (v *Views) RoleDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		roleID, err := strconv.Atoi(c.Param("roleid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to get roleid for role: %w", err))
		}

		role1, err := v.role.GetRole(c.Request().Context(), role.Role{RoleID: roleID})
		if err != nil {
			return fmt.Errorf("failed to get role for deleteRole: %w", err)
		}

		err = v.role.RemoveRoleForPermissions(c.Request().Context(), role1)
		if err != nil {
			return fmt.Errorf("failed to delete rolePermission for deleteRole: %w", err)
		}

		err = v.role.RemoveRoleForUsers(c.Request().Context(), role1)
		if err != nil {
			return fmt.Errorf("failed to delete roleUser for deleteRole: %w", err)
		}

		err = v.role.DeleteRole(c.Request().Context(), role1)
		if err != nil {
			return fmt.Errorf("failed to delete role for deleteRole: %w", err)
		}

		return c.Redirect(http.StatusFound, "/internal/roles")
	}

	return v.invalidMethodUsed(c)
}

// RoleAddPermissionFunc handles a rolePermission add request
func (v *Views) RoleAddPermissionFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		roleID, err := strconv.Atoi(c.Param("roleid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to get role for roleAddPermission: %w", err))
		}

		_, err = v.role.GetRole(c.Request().Context(), role.Role{RoleID: roleID})
		if err != nil {
			return fmt.Errorf("failed to get role for roleAddPermission: %w", err)
		}

		permissionID, err := strconv.Atoi(c.Request().FormValue("permissionID"))
		if err != nil {
			return fmt.Errorf("failed to get permissionid for roleAddPermission: %w", err)
		}

		_, err = v.permission.GetPermission(c.Request().Context(), permission.Permission{PermissionID: permissionID})
		if err != nil {
			return fmt.Errorf("failed to get permission for roleAddPermission: %w", err)
		}

		rolePermission := user.RolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
		}

		_, err = v.user.GetRolePermission(c.Request().Context(), rolePermission)
		if err == nil {
			return errors.New("failed to add rolePermission for roleAddPermission: row already exists")
		}

		_, err = v.user.AddRolePermission(c.Request().Context(), rolePermission)
		if err != nil {
			return fmt.Errorf("failed to add rolePermission for roleAddPermission: %w", err)
		}

		return c.Redirect(http.StatusFound, fmt.Sprintf("/internal/role/%d", roleID))
	}

	return v.invalidMethodUsed(c)
}

// RoleRemovePermissionFunc handles a rolePermission remove request
func (v *Views) RoleRemovePermissionFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		roleID, err := strconv.Atoi(c.Param("roleid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to get roleid for roleRemovePermission: %w", err))
		}

		_, err = v.role.GetRole(c.Request().Context(), role.Role{RoleID: roleID})
		if err != nil {
			return fmt.Errorf("failed to get role for roleRemovePermission: %w", err)
		}

		permissionID, err := strconv.Atoi(c.Param("permissionid"))
		if err != nil {
			return fmt.Errorf("failed to get permissionid for roleRemovePermission: %w", err)
		}

		_, err = v.permission.GetPermission(c.Request().Context(), permission.Permission{PermissionID: permissionID})
		if err != nil {
			return fmt.Errorf("failed to get permission for roleRemovePermission: %w", err)
		}

		rolePermission := user.RolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
		}

		_, err = v.user.GetRolePermission(c.Request().Context(), rolePermission)
		if err != nil {
			return fmt.Errorf("failed to get rolePermisison for roleRemovePermission: %w", err)
		}

		err = v.user.RemoveRolePermission(c.Request().Context(), rolePermission)
		if err != nil {
			return fmt.Errorf("failed to remove rolePermission for roleRemoveRole: %w", err)
		}

		return c.Redirect(http.StatusFound, fmt.Sprintf("/internal/role/%d", roleID))
	}

	return v.invalidMethodUsed(c)
}

// RoleAddUserFunc handles a roleUser add request
func (v *Views) RoleAddUserFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		roleID, err := strconv.Atoi(c.Param("roleid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to get role for roleAddUser: %w", err))
		}

		_, err = v.role.GetRole(c.Request().Context(), role.Role{RoleID: roleID})
		if err != nil {
			return fmt.Errorf("failed to get user for roleAddUser: %w", err)
		}

		userID, err := strconv.Atoi(c.Request().FormValue("userID"))
		if err != nil {
			return fmt.Errorf("failed to get userID for roleAddUser: %w", err)
		}

		_, err = v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
		if err != nil {
			return fmt.Errorf("failed to get user for roleAddUser: %w", err)
		}

		roleUser := user.RoleUser{
			RoleID: roleID,
			UserID: userID,
		}

		_, err = v.user.GetRoleUser(c.Request().Context(), roleUser)
		if err == nil {
			return errors.New("failed to add roleUser for roleAddUser: row already exists")
		}

		_, err = v.user.AddRoleUser(c.Request().Context(), roleUser)
		if err != nil {
			return fmt.Errorf("failed to add roleUser for roleAddUser: %w", err)
		}

		return c.Redirect(http.StatusFound, fmt.Sprintf("/internal/role/%d", roleID))
	}

	return v.invalidMethodUsed(c)
}

// RoleRemoveUserFunc handles a roleUser remove request
func (v *Views) RoleRemoveUserFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		roleID, err := strconv.Atoi(c.Param("roleid"))
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("failed to get roleid for roleRemoveUser: %w", err))
		}

		_, err = v.role.GetRole(c.Request().Context(), role.Role{RoleID: roleID})
		if err != nil {
			return fmt.Errorf("failed to get role for roleRemoveUser: %w", err)
		}

		userID, err := strconv.Atoi(c.Param("userid"))
		if err != nil {
			return fmt.Errorf("failed to get userID for roleRemoveUser: %w", err)
		}

		_, err = v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
		if err != nil {
			return fmt.Errorf("failed to get user for roleRemoveUser: %w", err)
		}

		roleUser := user.RoleUser{
			RoleID: roleID,
			UserID: userID,
		}

		_, err = v.user.GetRoleUser(c.Request().Context(), roleUser)
		if err != nil {
			return fmt.Errorf("failed to get roleUser for roleRemoveUser: %w", err)
		}

		err = v.user.RemoveRoleUser(c.Request().Context(), roleUser)
		if err != nil {
			return fmt.Errorf("failed to remove roleUser for roleRemoveUser: %w", err)
		}

		return c.Redirect(http.StatusFound, fmt.Sprintf("/internal/role/%d", roleID))
	}

	return v.invalidMethodUsed(c)
}
