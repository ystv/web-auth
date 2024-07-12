package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/ystv/web-auth/permission/permissions"
	"github.com/ystv/web-auth/views"
)

//go:embed public/static/*
var embeddedFiles embed.FS

type (
	Router struct {
		config *views.Config
		views  *views.Views
		router *echo.Echo
	}
	RouterConf struct {
		Config *views.Config
		Views  *views.Views
	}
)

func NewRouter(conf *RouterConf) *Router {
	r := &Router{
		config: conf.Config,
		router: echo.New(),
		views:  conf.Views,
	}
	r.router.HideBanner = true

	r.router.Debug = r.config.Debug

	r.middleware()

	r.loadRoutes()

	return r
}

//nolint:staticcheck
func (r *Router) Start() error {
	r.router.Logger.Error(r.router.Start(r.config.Address))

	return fmt.Errorf("failed to start router on address %s", r.config.Address)
}

// middleware initialises web server middleware
func (r *Router) middleware() {
	config := middleware.CORSConfig{
		AllowCredentials: true,
		Skipper:          middleware.DefaultSkipper,
		AllowOrigins: []string{
			"https://" + r.config.DomainName,
			"https://" + r.config.BaseDomainName,
			"https://creator." + r.config.BaseDomainName,
			"https://my." + r.config.BaseDomainName,
			"https://cobra." + r.config.BaseDomainName,
			"https://api." + r.config.BaseDomainName,
			"https://internal." + r.config.BaseDomainName,
			"https://docs." + r.config.BaseDomainName,
			"https://history." + r.config.BaseDomainName,
		},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept,
			echo.HeaderAccessControlAllowCredentials, echo.HeaderAccessControlAllowOrigin, echo.HeaderAuthorization},
		AllowMethods: []string{http.MethodGet, http.MethodPost},
	}

	r.router.Pre(middleware.RemoveTrailingSlash())
	r.router.Use(middleware.Recover())
	r.router.Use(middleware.CORSWithConfig(config))
	r.router.Use(middleware.BodyLimit("15M"))
	r.router.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Skipper:   middleware.DefaultSkipper,
		Level:     5,
		MinLength: 0,
	}))
}

func (r *Router) loadRoutes() {
	r.router.RouteNotFound("/*", r.views.Error404)

	r.router.HTTPErrorHandler = r.views.CustomHTTPErrorHandler

	assetHandler := http.FileServer(http.FS(echo.MustSubFS(embeddedFiles, "public/static")))

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
		internal.GET("/permissions", r.views.PermissionsFunc,
			r.views.RequirePermission(permissions.ManageMembersPermissions))
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
	rolePermission := roleID.Group("/permission")
	rolePermission.Match(validMethods, "/add", r.views.RoleAddPermissionFunc)
	rolePermission.Match(validMethods, "/remove/:permissionid", r.views.RoleRemovePermissionFunc)

	roleUser := roleID.Group("/user")
	roleUser.Match(validMethods, "/add", r.views.RoleAddUserFunc)
	roleUser.Match(validMethods, "/remove/:userid", r.views.RoleRemoveUserFunc)
	roleID.Match(validMethods, "", r.views.RoleFunc)

	// this section of users is a bit weird, users is valid for anyone who can list users and user/add can be used by
	// add users permission
	if !r.config.Debug {
		internal.Match(validMethods, "/users", r.views.UsersFunc,
			r.views.RequirePermission(permissions.ManageMembersMembersList))
		internal.Match(validMethods, "/user/add", r.views.UserAddFunc,
			r.views.RequirePermission(permissions.ManageMembersMembersAdd))
	} else {
		internal.Match(validMethods, "/users", r.views.UsersFunc)
		internal.Match(validMethods, "/user/add", r.views.UserAddFunc)
	}

	internal.Match(validMethods, "/user/release", r.views.ReleaseUserFunc)
	user := internal.Group("/user/:userid")
	// user is any function to do with a specific user
	if !r.config.Debug {
		user.Use(r.views.RequirePermission(permissions.ManageMembersMembersAdmin))
	}

	user.Match(validMethods, "/edit", r.views.UserEditFunc)
	user.Match(validMethods, "/delete", r.views.UserDeleteFunc)
	user.Match(validMethods, "/reset", r.views.ResetUserPasswordFunc)
	user.Match(validMethods, "/toggle", r.views.UserToggleEnabledFunc)
	user.Match(validMethods, "/assume", r.views.AssumeUserFunc, r.views.RequirePermission(permissions.SuperUser))
	user.Match(validMethods, "", r.views.UserFunc)

	internal.Match(validMethods, "/officerships", r.views.OfficershipsFunc,
		r.views.RequirePermission(permissions.ManageMembersOfficers))

	officershipsRoute := internal.Group("/officership")

	if !r.config.Debug {
		officershipsRoute.Use(r.views.RequirePermission(permissions.ManageMembersOfficers))
	}

	officershipsRoute.Match(validMethods, "/add", r.views.OfficershipAddFunc)

	officershipsRoute.Match(validMethods, "/officers", r.views.OfficersFunc)
	officershipsRoute.Match(validMethods, "/officer/add", r.views.OfficerAddFunc)

	officer := officershipsRoute.Group("/officer/:officerid")
	officer.Match(validMethods, "/edit", r.views.OfficerEditFunc)
	officer.Match(validMethods, "/delete", r.views.OfficerDeleteFunc)
	officer.Match(validMethods, "", r.views.OfficerFunc)

	officershipsRoute.Match(validMethods, "/teams", r.views.OfficershipTeamsFunc)
	officershipsRoute.Match(validMethods, "/team/add", r.views.OfficershipTeamAddFunc)

	officershipTeam := officershipsRoute.Group("/team/:officershipteamid")
	officershipTeam.Match(validMethods, "/edit", r.views.OfficershipTeamEditFunc)
	officershipTeam.Match(validMethods, "/delete", r.views.OfficershipTeamDeleteFunc)
	officershipTeam.Match(validMethods, "", r.views.OfficershipTeamFunc)

	officership := officershipsRoute.Group("/:officershipid")
	officership.Match(validMethods, "/edit", r.views.OfficershipEditFunc)
	officership.Match(validMethods, "/delete", r.views.OfficershipDeleteFunc)
	officership.Match(validMethods, "", r.views.OfficershipFunc)

	internalAPI := internal.Group("/api")
	internalAPI.Match(validMethods, "/set_token", r.views.SetTokenHandler)
	manage := internalAPI.Group("/manage")
	manage.Match(validMethods, "/add", r.views.TokenAddFunc)
	manage.Match(validMethods, "/:tokenid/delete", r.views.TokenDeleteFunc)
	manage.Match(validMethods, "", r.views.ManageAPIFunc)

	api := r.router.Group("/api")
	// api is all the methods that are used by the api interactions
	api.GET("/set_token", r.views.SetTokenHandler, r.views.RequiresLoginJSON)
	api.GET("/test", r.views.TestAPITokenFunc)
	api.GET("/health", func(c echo.Context) error {
		marshal, err := json.Marshal(struct {
			Status int `json:"status"`
		}{
			Status: http.StatusOK,
		})
		if err != nil {
			log.Printf("failed to marshal health json: %+v", err)

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
	base.Match(validMethods, "logout", r.views.LogoutFunc, r.views.RequiresLogin)
	base.Match(validMethods, "signup", r.views.SignUpFunc)
	base.Match(validMethods, "forgot", r.views.ForgotFunc)
	base.Match(validMethods, "reset/:url", r.views.ResetURLFunc)
}
