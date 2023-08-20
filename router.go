package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	middleware2 "github.com/labstack/echo/v4/middleware"
	"github.com/ystv/web-auth/middleware"
	"github.com/ystv/web-auth/permission/permissions"
	"github.com/ystv/web-auth/views"
	"io/fs"
	"net/http"
)

//go:embed public/static/*
var embeddedFiles embed.FS

type (
	Router struct {
		config  *views.Config
		address string
		views   *views.Views
		router  *echo.Echo
	}
	NewRouter struct {
		Config *views.Config
		Views  *views.Views
	}
)

func NewRouterFunc(conf *NewRouter) *Router {
	r := &Router{
		config: conf.Config,
		router: echo.New(),
		views:  conf.Views,
	}
	r.router.HideBanner = true

	r.router.Debug = r.config.Debug

	middleware.New(r.router, r.config.DomainName)

	r.loadRoutes()

	return r
}

func (r *Router) Start() error {
	r.router.Logger.Error(r.router.Start(r.config.Address))
	return fmt.Errorf("failed to start router on address %s", r.config.Address)
}

func (r *Router) loadRoutes() {
	r.router.RouteNotFound("/*", r.views.Error404)

	r.router.Use(middleware2.BodyLimit("15M"))

	r.router.HTTPErrorHandler = r.views.CustomHTTPErrorHandler

	r.router.Use(middleware2.GzipWithConfig(middleware2.GzipConfig{
		Level: 5,
	}))

	assetHandler := http.FileServer(getFileSystem())

	r.router.GET("/public/*", echo.WrapHandler(http.StripPrefix("/public/", assetHandler)))

	validMethods := []string{http.MethodGet, http.MethodPost}

	internal := r.router.Group("/internal")
	// internal is all the methods behind the login
	if !r.router.Debug {
		internal.Use(r.views.RequiresLogin)
	}
	internal.GET("", r.views.InternalFunc)
	internal.Match(validMethods, "/settings", r.views.SettingsFunc)

	// permissions are for listing the permissions
	if !r.config.Debug {
		//internal.GET("/permissions", r.views.PermissionsFunc, r.views.RequiresManageMembersPermissions)
		internal.GET("/permissions", r.views.PermissionsFunc, r.views.RequirePermission(permissions.ManageMembersPermissions))
	} else {
		internal.GET("/permissions", r.views.PermissionsFunc)
	}

	permission := internal.Group("/permission")
	// permission is any function to do with a specific permission or new permission
	if !r.config.Debug {
		permission.Use(r.views.RequirePermission(permissions.ManageMembersPermissions))
	}
	permission.Match(validMethods, "/add", r.views.PermissionAddFunc)
	permissionID := permission.Group("/:permissionid")
	// permissionID is any function to do with a specific permission
	permissionID.Match(validMethods, "/edit", r.views.PermissionEditFunc)
	permissionID.Match(validMethods, "/delete", r.views.PermissionDeleteFunc)
	permissionID.Match(validMethods, "", r.views.PermissionFunc)

	// roles are for listing the roles
	if !r.config.Debug {
		internal.GET("/roles", r.views.RolesFunc, r.views.RequirePermission(permissions.ManageMembersGroup))
	} else {
		internal.GET("/roles", r.views.RolesFunc)
	}

	role := internal.Group("/role")
	// role is any function to do with a specific role or new role
	if !r.config.Debug {
		role.Use(r.views.RequirePermission(permissions.ManageMembersGroup))
	}
	role.Match(validMethods, "/add", r.views.RoleAddFunc)
	roleID := role.Group("/:roleid")
	// roleID is any function to do with a specific role
	roleID.Match(validMethods, "/edit", r.views.RoleEditFunc)
	roleID.Match(validMethods, "/delete", r.views.RoleDeleteFunc)
	roleID.Match(validMethods, "", r.views.RoleFunc)

	// this section of users is a bit weird, users is valid for anyone who can list users and user/add can be used by add users permission
	if !r.config.Debug {
		internal.Match(validMethods, "/users", r.views.UsersFunc, r.views.RequirePermission(permissions.ManageMembersMembersList))
		internal.Match(validMethods, "/user/add", r.views.UserAddFunc, r.views.RequirePermission(permissions.ManageMembersMembersAdd))
	} else {
		internal.Match(validMethods, "/users", r.views.UsersFunc)
		internal.Match(validMethods, "/user/add", r.views.UserAddFunc)
	}

	user := internal.Group("/user/:userid")
	// user is any function to do with a specific user
	if !r.config.Debug {
		user.Use(r.views.RequirePermission(permissions.ManageMembersMembersAdmin))
	}
	user.Match(validMethods, "/edit", r.views.UserEditFunc)
	user.Match(validMethods, "/delete", r.views.UserDeleteFunc)
	user.Match(validMethods, "/reset", r.views.ResetUserPasswordFunc)
	user.Match(validMethods, "", r.views.UserFunc)

	api := r.router.Group("/api")
	// api is all the methods that are used by the api interactions
	api.GET("/set_token", r.views.SetTokenHandler)
	api.GET("/test", r.views.TestAPITokenFunc)
	api.GET("/health", func(c echo.Context) error {
		marshal, err := json.Marshal(struct {
			Status int `json:"status"`
		}{
			Status: http.StatusOK,
		})
		if err != nil {
			fmt.Println(err)
			return &echo.HTTPError{
				Code:     http.StatusBadRequest,
				Message:  err.Error(),
				Internal: err,
			}
		}

		c.Response().Header().Set("Content-Type", "application/json")
		return c.JSON(http.StatusOK, marshal)
	})

	base := r.router.Group("/")
	// base is the functions that don't require being logged in
	base.GET("", r.views.IndexFunc)
	base.Match(validMethods, "login", r.views.LoginFunc)
	base.Match(validMethods, "logout", r.views.LogoutFunc)
	base.Match(validMethods, "signup", r.views.SignUpFunc)
	base.Match(validMethods, "forgot", r.views.ForgotFunc)
	base.Match(validMethods, "reset/:url", r.views.ResetURLFunc)
}

func getFileSystem() http.FileSystem {
	fsys, err := fs.Sub(embeddedFiles, "public/static")
	if err != nil {
		panic(err)
	}

	return http.FS(fsys)
}
