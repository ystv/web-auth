package views

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
	"log"
	"net/http"
	"strconv"
)

type (
	PermissionsTemplate struct {
		Permissions     []permission.Permission
		UserPermissions []permission.Permission
		ActivePage      string
	}

	PermissionTemplate struct {
		Permission      user.PermissionTemplate
		UserPermissions []permission.Permission
		ActivePage      string
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
		log.Printf("failed to get permissions for permissions: %+v", err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get permissions for permissions: %w", err))
		}
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		log.Printf("failed to get user permissions for permissions: %+v", err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get user permissions for permissions: %+v", err))
		}
	}

	data := PermissionsTemplate{
		Permissions:     permissions,
		UserPermissions: p1,
		ActivePage:      "permissions",
	}

	return v.template.RenderTemplate(c.Response(), data, templates.PermissionsTemplate, templates.RegularType)
}

// PermissionFunc handles a permission request
func (v *Views) PermissionFunc(c echo.Context) error {
	c1 := v.getSessionData(c)

	permissionID, err := strconv.Atoi(c.Param("permissionid"))
	if err != nil {
		log.Printf("failed to get permissionid for permission: %+v", err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to parse permissionid for permission: %w", err))
		}
	}

	permission1, err := v.permission.GetPermission(c.Request().Context(), permission.Permission{PermissionID: permissionID})
	if err != nil {
		log.Printf("failed to get permission for permission: %+v", err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get permission for permission: %w", err))
		}
	}

	permissionTemplate := v.bindPermissionToTemplate(permission1)

	permissionTemplate.Roles, err = v.user.GetRolesForPermission(c.Request().Context(), permission1)
	if err != nil {
		log.Printf("failed to get roles for permission: %+v", err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get roles for permission: %w", err))
		}
	}

	p1, err := v.user.GetPermissionsForUser(c.Request().Context(), c1.User)
	if err != nil {
		log.Printf("failed to get user permissions for permission: %+v", err)
		if !v.conf.Debug {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("failed to get user permissions for permission: %+v", err))
		}
	}

	data := PermissionTemplate{
		Permission:      permissionTemplate,
		UserPermissions: p1,
		ActivePage:      "permission",
	}

	return v.template.RenderTemplate(c.Response(), data, templates.PermissionTemplate, templates.RegularType)
}

// PermissionAddFunc handles an add permission request
func (v *Views) PermissionAddFunc(c echo.Context) error {
	return nil
}

// PermissionEditFunc handles an edit permission request
func (v *Views) PermissionEditFunc(c echo.Context) error {
	return nil
}

// PermissionDeleteFunc handles a delete permission request
func (v *Views) PermissionDeleteFunc(c echo.Context) error {
	return nil
}
