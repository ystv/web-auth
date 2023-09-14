package views

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
)

type (
	PermissionsTemplate struct {
		Permissions []permission.Permission
		TemplateHelper
	}

	PermissionTemplate struct {
		Permission user.PermissionTemplate
		TemplateHelper
	}
)

func (v *Views) bindPermissionToTemplate(p1 permission.Permission) user.PermissionTemplate {
	var p = user.PermissionTemplate{}
	p.PermissionID = p1.PermissionID
	p.Name = p1.Name
	p.Description = p1.Description
	return p
}

// PermissionsFunc handles a permissions request
func (v *Views) PermissionsFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	permissions, err := v.permission.GetPermissions(c.Request().Context())
	if err != nil {
		return fmt.Errorf("failed to get permissions for permissions: %w", err)
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		return fmt.Errorf("failed to get user permissions for permissions: %w", err)
	}

	data := PermissionsTemplate{
		Permissions: permissions,
		TemplateHelper: TemplateHelper{
			UserPermissions: p1,
			ActivePage:      "permissions",
		},
	}

	return v.template.RenderTemplate(c.Response(), data, templates.PermissionsTemplate, templates.RegularType)
}

// PermissionFunc handles a permission request
func (v *Views) PermissionFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	permissionID, err := strconv.Atoi(c.Param("permissionid"))
	if err != nil {
		return fmt.Errorf("failed to parse permissionid for permission: %w", err)
	}

	permission1, err := v.permission.GetPermission(c.Request().Context(), permission.Permission{PermissionID: permissionID})
	if err != nil {
		return fmt.Errorf("failed to get permission for permission: %w", err)
	}

	permissionTemplate := v.bindPermissionToTemplate(permission1)

	permissionTemplate.Roles, err = v.user.GetRolesForPermission(c.Request().Context(), permission1)
	if err != nil {
		return fmt.Errorf("failed to get roles for permission: %w", err)
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		return fmt.Errorf("failed to get user permissions for permission: %w", err)
	}

	data := PermissionTemplate{
		Permission: permissionTemplate,
		TemplateHelper: TemplateHelper{
			UserPermissions: p1,
			ActivePage:      "permission",
		},
	}

	return v.template.RenderTemplate(c.Response(), data, templates.PermissionTemplate, templates.RegularType)
}

// PermissionAddFunc handles an add permission request
func (v *Views) PermissionAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		name := c.Request().FormValue("name")
		description := c.Request().FormValue("description")

		p1, err := v.permission.GetPermission(c.Request().Context(), permission.Permission{PermissionID: 0, Name: name})
		if err == nil && p1.PermissionID > 0 {
			return fmt.Errorf("permission with name \"%s\" already exists", name)
		}

		_, err = v.permission.AddPermission(c.Request().Context(), permission.Permission{PermissionID: -1, Name: name, Description: description})
		if err != nil {
			return fmt.Errorf("failed to add permission for permissionadd: %w", err)
		}
		return v.PermissionsFunc(c)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}

// PermissionEditFunc handles an edit permission request
func (v *Views) PermissionEditFunc(c echo.Context) error {
	_ = c
	return nil
}

// PermissionDeleteFunc handles a delete permission request
func (v *Views) PermissionDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		permissionID, err := strconv.Atoi(c.Param("permissionid"))
		if err != nil {
			return fmt.Errorf("failed to get permissionid for permission: %w", err)
		}

		permission1, err := v.permission.GetPermission(c.Request().Context(), permission.Permission{PermissionID: permissionID})
		if err != nil {
			return fmt.Errorf("failed to get permission for deletePermission: %w", err)
		}

		roles, err := v.user.GetRolesForPermission(c.Request().Context(), permission1)
		if err != nil {
			return fmt.Errorf("failed to get roles for deletePermission: %w", err)
		}

		for _, role1 := range roles {
			err = v.role.RemovePermissionsForRole(c.Request().Context(), role1)
			if err != nil {
				return fmt.Errorf("failed to delete rolePermission for deletePermission: %w", err)
			}
		}

		err = v.permission.DeletePermission(c.Request().Context(), permission1)
		if err != nil {
			return fmt.Errorf("failed to delete permission for deletePermission: %w", err)
		}
		return v.PermissionsFunc(c)
	}
	return echo.NewHTTPError(http.StatusMethodNotAllowed, fmt.Errorf("invalid method used"))
}
