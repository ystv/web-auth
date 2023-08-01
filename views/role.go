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
		Roles      []role.Role
		UserID     int
		ActivePage string
	}

	RoleTemplate struct {
		Role                 user.RoleTemplate
		UserID               int
		PermissionsNotInRole []permission.Permission
		UsersNotInRole       []user.User
		ActivePage           string
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
		log.Println(err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	data := RolesTemplate{
		Roles:      roles,
		UserID:     c1.User.UserID,
		ActivePage: "roles",
	}

	return v.template.RenderTemplate(c.Response(), data, templates.RolesTemplate)
}

func (v *Views) RoleFunc(c echo.Context) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	roleID, err := strconv.Atoi(c.Param("roleid"))
	if err != nil {
		log.Printf("failed to get roleid for role: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	role1, err := v.role.GetRole(c.Request().Context(), role.Role{RoleID: roleID})
	if err != nil {
		log.Printf("failed to get role for role: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	roleTemplate := v.bindRoleToTemplate(role1)

	roleTemplate.Permissions, err = v.user.GetPermissionsForRole(c.Request().Context(), role1)
	if err != nil {
		log.Printf("failed to get permissions for role: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	roleTemplate.Users, err = v.user.GetUsersForRole(c.Request().Context(), role1)
	if err != nil {
		log.Printf("failed to get users for role: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	permissions, err := v.user.GetPermissionsNotInRole(c.Request().Context(), role1)
	if err != nil {
		log.Printf("failed to get permissions not in role for role: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, fmt.Errorf("failed to get permissions not in role for role: %+v", err))
		}
	}

	users, err := v.user.GetUsersNotInRole(c.Request().Context(), role1)
	if err != nil {
		log.Printf("failed to get users not in role for role: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, fmt.Errorf("failed to get users not in role for role: %+v", err))
		}
	}

	data := RoleTemplate{
		Role:                 roleTemplate,
		UserID:               c1.User.UserID,
		PermissionsNotInRole: permissions,
		UsersNotInRole:       users,
		ActivePage:           "role",
	}

	return v.template.RenderTemplate(c.Response(), data, templates.RoleTemplate)
}

func (v *Views) roleFunc(c echo.Context, roleID int) error {
	session, _ := v.cookie.Get(c.Request(), v.conf.SessionCookieName)

	c1 := v.getData(session)

	role1, err := v.role.GetRole(c.Request().Context(), role.Role{RoleID: roleID})
	if err != nil {
		log.Printf("failed to get role for role: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	roleTemplate := v.bindRoleToTemplate(role1)

	roleTemplate.Permissions, err = v.user.GetPermissionsForRole(c.Request().Context(), role1)
	if err != nil {
		log.Printf("failed to get roles for role: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	roleTemplate.Users, err = v.user.GetUsersForRole(c.Request().Context(), role1)
	if err != nil {
		log.Printf("failed to get users for role: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, err)
		}
	}

	permissions, err := v.user.GetPermissionsNotInRole(c.Request().Context(), role1)
	if err != nil {
		log.Printf("failed to get permissions not in role for role: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, fmt.Errorf("failed to get permissions not in role for role: %+v", err))
		}
	}

	users, err := v.user.GetUsersNotInRole(c.Request().Context(), role1)
	if err != nil {
		log.Printf("failed to get users not in role for role: %+v", err)
		if !v.conf.Debug {
			return v.errorHandle(c, fmt.Errorf("failed to get users not in role for role: %+v", err))
		}
	}

	data := RoleTemplate{
		Role:                 roleTemplate,
		UserID:               c1.User.UserID,
		PermissionsNotInRole: permissions,
		UsersNotInRole:       users,
		ActivePage:           "role",
	}

	return v.template.RenderTemplate(c.Response(), data, templates.RoleTemplate)
}

func (v *Views) RoleAddFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		err := c.Request().ParseForm()
		if err != nil {
			fmt.Println(err)
			return err
		}

		name := c.Request().FormValue("name")
		description := c.Request().FormValue("description")

		r1, err := v.role.GetRole(c.Request().Context(), role.Role{RoleID: 0, Name: name})
		if err == nil && r1.RoleID > 0 {
			return v.errorHandle(c, fmt.Errorf("role with name \"%s\" already exists", name))
		}

		_, err = v.role.AddRole(c.Request().Context(), role.Role{RoleID: -1, Name: name, Description: description})
		if err != nil {
			return v.errorHandle(c, err)
		}
		return v.RolesFunc(c)
	} else {
		return v.errorHandle(c, fmt.Errorf("invalid method used"))
	}
}

func (v *Views) RoleEditFunc(c echo.Context) error {
	return nil
}

func (v *Views) RoleDeleteFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		roleID, err := strconv.Atoi(c.Param("roleid"))
		if err != nil {
			log.Printf("failed to get roleid for role: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get roleid for role: %+v", err))
			}
		}

		role1, err := v.role.GetRole(c.Request().Context(), role.Role{RoleID: roleID})
		if err != nil {
			log.Printf("failed to get role for deleteRole: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get role for deleteRole: %+v", err))
			}
		}

		permissions, err := v.user.GetPermissionsForRole(c.Request().Context(), role1)
		if err != nil {
			log.Printf("failed to get permissions for deleteRole: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get permissions for deleteRole: %+v", err))
			}
		}

		for _, p1 := range permissions {
			err = v.permission.DeleteRolePermission(c.Request().Context(), p1)
			if err != nil {
				log.Printf("failed to delete rolePermission for deleteRole: %+v", err)
				if !v.conf.Debug {
					return v.errorHandle(c, fmt.Errorf("failed to delete rolePermission for deleteRole: %+v", err))
				}
			}
		}

		err = v.role.DeleteRole(c.Request().Context(), role1)
		if err != nil {
			log.Printf("failed to delete role for deleteRole: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to delete role for deleteRole: %+v", err))
			}
		}
		return v.RolesFunc(c)
	} else {
		return v.errorHandle(c, fmt.Errorf("invalid method used"))
	}
}

func (v *Views) RoleAddPermissionFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		err := c.Request().ParseForm()
		if err != nil {
			fmt.Println(err)
			return err
		}

		roleID, err := strconv.Atoi(c.Param("roleid"))
		if err != nil {
			log.Printf("failed to get role for roleAddPermission: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get role for roleAddPermission: %+v", err))
			}
		}

		_, err = v.role.GetRole(c.Request().Context(), role.Role{RoleID: roleID})
		if err != nil {
			log.Printf("failed to get role for roleAddPermission: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get role for roleAddPermission: %+v", err))
			}
		}

		permissionID, err := strconv.Atoi(c.Request().FormValue("permission"))
		if err != nil {
			log.Printf("failed to get permissionid for roleAddPermission: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get permissionid for roleAddPermission: %+v", err))
			}
		}
		_, err = v.permission.GetPermission(c.Request().Context(), permission.Permission{PermissionID: permissionID})
		if err != nil {
			log.Printf("failed to get permission for roleAddPermission: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get permission for roleAddPermission: %+v", err))
			}
		}

		rolePermission := user.RolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
		}

		_, err = v.user.GetRolePermission(c.Request().Context(), rolePermission)
		if err == nil {
			log.Printf("failed to add rolePermisison for roleAddPermission: row already exists")
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to add rolePermission for roleAddPermission: row already exists"))
			}
		}

		_, err = v.user.AddRolePermission(c.Request().Context(), rolePermission)
		if err != nil {
			log.Printf("failed to add rolePermission for roleAddPermission: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to add rolePermission for roleAddPermission: %+v", err))
			}
		}

		return v.roleFunc(c, roleID)
	} else {
		return v.errorHandle(c, fmt.Errorf("invalid method used"))
	}
}

func (v *Views) RoleRemovePermissionFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		roleID, err := strconv.Atoi(c.Param("roleid"))
		if err != nil {
			log.Printf("failed to get roleid for roleRemovePermission: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get roleid for roleRemovePermission: %+v", err))
			}
		}

		_, err = v.role.GetRole(c.Request().Context(), role.Role{RoleID: roleID})
		if err != nil {
			log.Printf("failed to get role for roleRemovePermission: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get role for roleRemovePermission: %+v", err))
			}
		}

		permissionID, err := strconv.Atoi(c.Param("permissionid"))
		if err != nil {
			log.Printf("failed to get permissionid for roleRemovePermission: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get permissionid for roleRemovePermission: %+v", err))
			}
		}
		_, err = v.permission.GetPermission(c.Request().Context(), permission.Permission{PermissionID: permissionID})
		if err != nil {
			log.Printf("failed to get permission for roleRemovePermission: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get permission for roleRemovePermission: %+v", err))
			}
		}

		rolePermission := user.RolePermission{
			RoleID:       roleID,
			PermissionID: permissionID,
		}

		_, err = v.user.GetRolePermission(c.Request().Context(), rolePermission)
		if err != nil {
			log.Printf("failed to get rolePermisison for roleRemovePermission: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get rolePermission for roleRemovePermission: %+v", err))
			}
		}

		err = v.user.RemoveRolePermission(c.Request().Context(), rolePermission)
		if err != nil {
			log.Printf("failed to remove rolePermission for roleRemoveRole: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to remove rolePermission for roleRemovePermission: %+v", err))
			}
		}

		return v.roleFunc(c, roleID)
	} else {
		return v.errorHandle(c, fmt.Errorf("invalid method used"))
	}
}

func (v *Views) RoleAddUserFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		err := c.Request().ParseForm()
		if err != nil {
			fmt.Println(err)
			return err
		}

		roleID, err := strconv.Atoi(c.Param("roleid"))
		if err != nil {
			log.Printf("failed to get role for roleAddUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get role for roleAddUser: %+v", err))
			}
		}

		_, err = v.role.GetRole(c.Request().Context(), role.Role{RoleID: roleID})
		if err != nil {
			log.Printf("failed to get user for roleAddUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get user for roleAddUser: %+v", err))
			}
		}

		userID, err := strconv.Atoi(c.Request().FormValue("user"))
		if err != nil {
			log.Printf("failed to get userID for roleAddUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get userID for roleAddUser: %+v", err))
			}
		}
		_, err = v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
		if err != nil {
			log.Printf("failed to get user for roleAddUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get user for roleAddUser: %+v", err))
			}
		}

		roleUser := user.RoleUser{
			RoleID: roleID,
			UserID: userID,
		}

		_, err = v.user.GetRoleUser(c.Request().Context(), roleUser)
		if err == nil {
			log.Printf("failed to add roleUser for roleAddUser: row already exists")
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to add roleUser for roleAddUser: row already exists"))
			}
		}

		_, err = v.user.AddRoleUser(c.Request().Context(), roleUser)
		if err != nil {
			log.Printf("failed to add roleUser for roleAddUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to add roleUser for roleAddUser: %+v", err))
			}
		}

		return v.roleFunc(c, roleID)
	} else {
		return v.errorHandle(c, fmt.Errorf("invalid method used"))
	}
}

func (v *Views) RoleRemoveUserFunc(c echo.Context) error {
	if c.Request().Method == http.MethodPost {
		roleID, err := strconv.Atoi(c.Param("roleid"))
		if err != nil {
			log.Printf("failed to get roleid for roleRemoveUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get roleid for roleRemoveUser: %+v", err))
			}
		}

		_, err = v.role.GetRole(c.Request().Context(), role.Role{RoleID: roleID})
		if err != nil {
			log.Printf("failed to get role for roleRemoveUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get role for roleRemoveUser: %+v", err))
			}
		}

		userID, err := strconv.Atoi(c.Param("userid"))
		if err != nil {
			log.Printf("failed to get userID for roleRemoveUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get userID for roleRemoveUser: %+v", err))
			}
		}
		_, err = v.user.GetUser(c.Request().Context(), user.User{UserID: userID})
		if err != nil {
			log.Printf("failed to get user for roleRemoveUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get user for roleRemoveUser: %+v", err))
			}
		}

		roleUser := user.RoleUser{
			RoleID: roleID,
			UserID: userID,
		}

		_, err = v.user.GetRoleUser(c.Request().Context(), roleUser)
		if err != nil {
			log.Printf("failed to get roleUser for roleRemoveUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to get roleUser for roleRemoveUser: %+v", err))
			}
		}

		err = v.user.RemoveRoleUser(c.Request().Context(), roleUser)
		if err != nil {
			log.Printf("failed to remove roleUser for roleRemoveUser: %+v", err)
			if !v.conf.Debug {
				return v.errorHandle(c, fmt.Errorf("failed to remove roleUser for roleRemoveUser: %+v", err))
			}
		}

		return v.roleFunc(c, roleID)
	} else {
		return v.errorHandle(c, fmt.Errorf("invalid method used"))
	}
}
