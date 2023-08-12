package router

import (
	"embed"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	middleware2 "github.com/labstack/echo/v4/middleware"
	"github.com/ystv/web-auth/middleware"
	"github.com/ystv/web-auth/views"
	"io/fs"
	"net/http"
)

// embeddedFiles are the storage of static files in the executable
//
//go:embed public/static/*
var embeddedFiles embed.FS

type (
	// Router is the main space for the config
	Router struct {
		config *views.Config
		port   string
		views  *views.Views
		router *echo.Echo
	}

	// NewRouter is what initialises the router
	NewRouter struct {
		Config *views.Config
		Views  *views.Views
	}
)

// New initialises everything needed
func New(conf *NewRouter) *Router {
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

// Start starts the web server
func (r *Router) Start() error {
	r.router.Logger.Error(r.router.Start(r.config.Port))
	return fmt.Errorf("failed to start router on port %s", r.config.Port)
}

// loadRoutes loads all the potential rouutes a user could take
func (r *Router) loadRoutes() {
	r.router.RouteNotFound("/*", r.views.Error404)

	r.router.Use(middleware2.BodyLimit("15M"))

	assetHandler := http.FileServer(getFileSystem())

	r.router.GET("/public/*", echo.WrapHandler(http.StripPrefix("/public/", assetHandler)))

	validMethods := []string{http.MethodGet, http.MethodPost}

	internal := r.router.Group("/internal")

	if !r.router.Debug {
		internal.Use(r.views.RequiresLogin)
	}
	{
		internal.GET("", r.views.InternalFunc)
		internal.Match(validMethods, "/settings", r.views.SettingsFunc)
		internal.Match(validMethods, "/changepassword", r.views.ChangePasswordFunc)

		if !r.config.Debug {
			internal.GET("/permissions", r.views.PermissionsFunc, r.views.RequiresMinimumPermissionMMP)
		} else {
			internal.GET("/permissions", r.views.PermissionsFunc)
		}
		permission := internal.Group("/permission")
		if !r.config.Debug {
			permission.Use(r.views.RequiresMinimumPermissionMMP)
		}
		{
			permission.Match(validMethods, "/add", r.views.PermissionAddFunc)
			permissionID := permission.Group("/:permissionid")
			{
				permissionID.Match(validMethods, "/edit", r.views.PermissionEditFunc)
				permissionID.Match(validMethods, "/delete", r.views.PermissionDeleteFunc)
				permissionID.Match(validMethods, "", r.views.PermissionFunc)
			}
		}

		if !r.config.Debug {
			internal.GET("/roles", r.views.RolesFunc, r.views.RequiresMinimumPermissionMMG)
		} else {
			internal.GET("/roles", r.views.RolesFunc)
		}

		role := internal.Group("/role")
		if !r.config.Debug {
			role.Use(r.views.RequiresMinimumPermissionMMG)
		}
		{
			role.Match(validMethods, "/add", r.views.RoleAddFunc)
			roleID := role.Group("/:roleid")
			{
				roleID.Match(validMethods, "/edit", r.views.RoleEditFunc)
				roleID.Match(validMethods, "/delete", r.views.RoleDeleteFunc)
				permission1 := roleID.Group("/permission")
				{
					permission1.Match(validMethods, "/add", r.views.RoleAddPermissionFunc)
					permission1.Match(validMethods, "/remove/:permissionid", r.views.RoleRemovePermissionFunc)
				}
				user1 := roleID.Group("/user")
				{
					user1.Match(validMethods, "/add", r.views.RoleAddUserFunc)
					user1.Match(validMethods, "/remove/:userid", r.views.RoleRemoveUserFunc)
				}
				roleID.Match(validMethods, "", r.views.RoleFunc)
			}
		}

		if !r.config.Debug {
			internal.Match(validMethods, "/users", r.views.UsersFunc, r.views.RequiresMinimumPermissionMMML)
			internal.Match(validMethods, "/user/add", r.views.UserAddFunc, r.views.RequiresMinimumPermissionMMAdd)
		} else {
			internal.Match(validMethods, "/users", r.views.UsersFunc)
			internal.Match(validMethods, "/user/add", r.views.UserAddFunc)
		}

		user := internal.Group("/user")
		if !r.config.Debug {
			user.Use(r.views.RequiresMinimumPermissionMMAdmin)
		}
		{
			userID := user.Group("/:userid")
			{
				userID.Match(validMethods, "/edit", r.views.UserEditFunc)
				userID.Match(validMethods, "/delete", r.views.UserDeleteFunc)
				userID.Match(validMethods, "/reset", r.views.ResetUserPasswordFunc)
				userID.Match(validMethods, "/toggle", r.views.UserToggleEnabledFunc)
				userID.Match(validMethods, "", r.views.UserFunc)
			}
		}

		api := internal.Group("/api")
		{
			api.Match(validMethods, "/set_token", r.views.SetTokenHandler)
			manage := api.Group("/manage")
			{
				manage.Match(validMethods, "/add", r.views.TokenAddFunc)
				manage.Match(validMethods, "/:tokenid/delete", r.views.TokenDeleteFunc)
				manage.Match(validMethods, "", r.views.ManageAPIFunc)
			}
		}
	}

	api := r.router.Group("/api")
	{
		api.GET("/test", r.views.TestAPI)
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
		loginAPI := api.Group("/")
		if !r.config.Debug {
			loginAPI.Use(r.views.RequiresLogin)
		}
		{
			loginAPI.Match(validMethods, "set_token", r.views.SetTokenHandler)
		}
	}

	base := r.router.Group("/")
	{
		base.GET("", r.views.IndexFunc)
		base.Match(validMethods, "login", r.views.LoginFunc)
		base.Match(validMethods, "logout", r.views.LogoutFunc)
		base.Match(validMethods, "signup", r.views.SignUpFunc)
		base.Match(validMethods, "forgot", r.views.ForgotFunc)
		base.Match(validMethods, "reset/:url", r.views.ResetURLFunc)
	}
}

// getFileSystem gets files from the operating system, in a certain folder that is
func getFileSystem() http.FileSystem {
	fsys, err := fs.Sub(embeddedFiles, "public/static")
	if err != nil {
		panic(err)
	}

	return http.FS(fsys)
}
