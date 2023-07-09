package views

import (
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/templates"
	"github.com/ystv/web-auth/user"
	"log"
	"strconv"
)

type (
	PermissionsTemplate struct {
		Permissions []permission.Permission
		UserID      int
		ActivePage  string
	}

	PermissionTemplate struct {
		Permission user.PermissionTemplate
		UserID     int
		ActivePage string
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
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	permissions, err := v.permission.GetPermissions(c.Request().Context())
	if err != nil {
		log.Printf("failed to get permissions for permissions: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	data := PermissionsTemplate{
		Permissions: permissions,
		UserID:      c1.User.UserID,
		ActivePage:  "permissions",
	}

	return v.template.RenderTemplate(c.Response(), data, templates.PermissionsTemplate)
}

// PermissionFunc handles a permission request
func (v *Views) PermissionFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	permissionID, err := strconv.Atoi(c.Param("permissionid"))
	if err != nil {
		log.Printf("failed to get permissionid for permission: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}
	permission1, err := v.permission.GetPermission(c.Request().Context(), permission.Permission{PermissionID: permissionID})
	if err != nil {
		log.Printf("failed to get permission for permission: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	permissionTemplate := v.bindPermissionToTemplate(permission1)

	permissionTemplate.Roles, err = v.user.GetRolesForPermission(c.Request().Context(), permission1)

	data := PermissionTemplate{
		Permission: permissionTemplate,
		UserID:     c1.User.UserID,
		ActivePage: "permission",
	}

	return v.template.RenderTemplate(c.Response(), data, templates.PermissionTemplate)
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
