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
	if err != nil {
		log.Printf("failed to get roles for permission: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, fmt.Errorf("failed to get roles for permission: %+v", err))
		}
	}

	data := PermissionTemplate{
		Permission: permissionTemplate,
		UserID:     c1.User.UserID,
		ActivePage: "permission",
	}

	return v.template.RenderTemplate(c.Response(), data, templates.PermissionTemplate)
}

// PermissionAddFunc handles an add permission request
func (v *Views) PermissionAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		err := c.Request().ParseForm()
		if err != nil {
			fmt.Println(err)
			return err
		}

		name := c.Request().FormValue("name")
		description := c.Request().FormValue("description")

		p1, err := v.permission.GetPermission(c.Request().Context(), permission.Permission{PermissionID: 0, Name: name})
		if err == nil && p1.PermissionID > 0 {
			return v.errorHandle(c, fmt.Errorf("permission with name \"%s\" already exists", name))
		}

		_, err = v.permission.AddPermission(c.Request().Context(), permission.Permission{PermissionID: -1, Name: name, Description: description})
		if err != nil {
			return v.errorHandle(c, err)
		}
		return v.PermissionsFunc(c)
	} else {
		return v.errorHandle(c, fmt.Errorf("invalid method used"))
	}
}

// PermissionEditFunc handles an edit permission request
func (v *Views) PermissionEditFunc(c echo.Context) error {
	return nil
}

// PermissionDeleteFunc handles a delete permission request
func (v *Views) PermissionDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		permissionID, err := strconv.Atoi(c.Param("permissionid"))
		if err != nil {
			log.Printf("failed to get permissionid for permission: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get permissionid for permission: %+v", err))
			}
		}

		permission1, err := v.permission.GetPermission(c.Request().Context(), permission.Permission{PermissionID: permissionID})
		if err != nil {
			log.Printf("failed to get permission for deletePermission: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get permission for deletePermission: %+v", err))
			}
		}

		roles, err := v.user.GetRolesForPermission(c.Request().Context(), permission1)
		if err != nil {
			log.Printf("failed to get roles for deletePermission: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get roles for deletePermission: %+v", err))
			}
		}

		for _, role1 := range roles {
			err = v.role.DeleteRolePermission(c.Request().Context(), role1)
			if err != nil {
				log.Printf("failed to delete rolePermission for deletePermission: %+v", err)
				if !v.conf.Debug {
					return v.errorHandle(c, fmt.Errorf("failed to delete rolePermission for deletePermission: %+v", err))
				}
			}
		}

		err = v.permission.DeletePermission(c.Request().Context(), permission1)
		if err != nil {
			log.Printf("failed to delete permission for deletePermission: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to delete permission for deletePermission: %+v", err))
			}
		}
		return v.PermissionsFunc(c)
	} else {
		return v.errorHandle(c, fmt.Errorf("invalid method used"))
	}
}
