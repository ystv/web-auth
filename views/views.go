package views

import (
	"context"
	"encoding/gob"
	"encoding/hex"
	"github.com/labstack/echo/v4"
	"github.com/ystv/web-auth/api"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/permission/permissions"
	"github.com/ystv/web-auth/role"
	"github.com/ystv/web-auth/templates"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/patrickmn/go-cache"
	"github.com/ystv/web-auth/infrastructure/db"
	"github.com/ystv/web-auth/infrastructure/mail"
	"github.com/ystv/web-auth/user"
)

type (
	// Config the global web-auth configuration
	Config struct {
		Version           string
		Debug             bool
		Port              string
		DatabaseURL       string
		BaseDomainName    string
		DomainName        string
		LogoutEndpoint    string
		SessionCookieName string
		Mail              SMTPConfig
		Security          SecurityConfig
	}

	// SMTPConfig stores the SMTP Mailer configuration
	SMTPConfig struct {
		Host          string
		Username      string
		Password      string
		Port          int
		DefaultMailTo string
	}

	// SecurityConfig stores the security configuration
	SecurityConfig struct {
		EncryptionKey     string
		AuthenticationKey string
		SigningKey        string
	}

	// Repo defines all view interactions
	Repo interface {
		// Error404 is the error handler for 404 errors
		Error404(c echo.Context) error
		// Error500 is the error handler for 500 errors
		Error500(c echo.Context) error

		// IndexFunc is the index function for the root url
		IndexFunc(c echo.Context) error

		// LoginFunc handles logins
		LoginFunc(c echo.Context) error
		// LogoutFunc handles logouts
		LogoutFunc(c echo.Context) error
		// SignUpFunc handles signups
		SignUpFunc(c echo.Context) error
		ForgotFunc(c echo.Context) error
		ResetURLFunc(c echo.Context) error
		ResetUserPasswordFunc(c echo.Context) error

		ChangePasswordFunc(c echo.Context) error

		// InternalFunc is the internal dashboard
		InternalFunc(c echo.Context) error
		SettingsFunc(c echo.Context) error

		PermissionsFunc(c echo.Context) error
		PermissionFunc(c echo.Context) error
		permissionFunc(c echo.Context, permissionID int) error
		PermissionAddFunc(c echo.Context) error
		PermissionEditFunc(c echo.Context) error
		PermissionDeleteFunc(c echo.Context) error
		bindPermissionToTemplate(p1 permission.Permission) user.PermissionTemplate

		RolesFunc(c echo.Context) error
		RoleFunc(c echo.Context) error
		roleFunc(c echo.Context, roleID int) error
		RoleAddFunc(c echo.Context) error
		RoleEditFunc(c echo.Context) error
		RoleDeleteFunc(c echo.Context) error
		RoleAddPermissionFunc(c echo.Context) error
		RoleRemovePermissionFunc(c echo.Context) error
		RoleAddUserFunc(c echo.Context) error
		RoleRemoveUserFunc(c echo.Context) error
		bindRoleToTemplate(r1 role.Role) user.RoleTemplate

		UsersFunc(c echo.Context) error
		UserFunc(c echo.Context) error
		UserAddFunc(c echo.Context) error
		UserEditFunc(c echo.Context) error
		UserToggleEnabledFunc(c echo.Context) error
		UserDeleteFunc(c echo.Context) error

		// RequiresLogin ensures the user is logged in
		RequiresLogin(next echo.HandlerFunc) echo.HandlerFunc

		RequiresMinimumPermission(next echo.HandlerFunc, p permissions.Permissions) echo.HandlerFunc
		RequiresMinimumPermissionMMP(next echo.HandlerFunc) echo.HandlerFunc
		RequiresMinimumPermissionMMG(next echo.HandlerFunc) echo.HandlerFunc
		RequiresMinimumPermissionMMML(next echo.HandlerFunc) echo.HandlerFunc
		RequiresMinimumPermissionMMAdd(next echo.HandlerFunc) echo.HandlerFunc
		RequiresMinimumPermissionMMAdmin(next echo.HandlerFunc) echo.HandlerFunc
		RequiresMinimumPermissionNoHttp(userID int, p permissions.Permissions) bool

		ManageAPIFunc(c echo.Context) error
		manageAPIFunc(c echo.Context, addedJWT string) error
		TokenAddFunc(c echo.Context) error
		TokenDeleteFunc(c echo.Context) error
		// SetTokenHandler is
		SetTokenHandler(c echo.Context) error
		// ValidateToken is
		ValidateToken(myToken string) (bool, *JWTClaims)
		// newJWT is
		newJWT(u user.User) (string, error)
		// TestAPI is
		TestAPI(c echo.Context) error

		getData(s *sessions.Session) *Context

		errorHandle(c echo.Context, err error) error
	}

	// Views encapsulates our view dependencies
	Views struct {
		api        *api.Store
		cache      *cache.Cache
		conf       *Config
		cookie     *sessions.CookieStore
		Mailer     *mail.Mailer
		permission *permission.Store
		role       *role.Store
		template   *templates.Templater
		user       *user.Store
		validate   *validator.Validate
	}
)

// here to verify we are meeting the interface
var _ Repo = &Views{}

// New initialises connections, templates, and cookies
func New(conf *Config, host, port string) *Views {
	v := &Views{}
	// Connecting to stores
	dbStore, err := db.NewStore(conf.DatabaseURL)
	if err != nil {
		if conf.Debug {
			log.Printf("db failed: %+v", err)
		} else {
			log.Fatalf("db failed: %+v", err)
		}
	} else {
		log.Printf("connected to db: %s:%s", host, port)
	}

	v.permission = permission.NewPermissionRepo(dbStore)
	v.role = role.NewRoleRepo(dbStore)
	v.user = user.NewUserRepo(dbStore)
	v.api = api.NewAPIRepo(dbStore)

	v.template = templates.NewTemplate(v.permission, v.role, v.user)

	// Connecting to mail
	v.Mailer, err = mail.NewMailer(mail.Config{
		Host:       conf.Mail.Host,
		Port:       conf.Mail.Port,
		Username:   conf.Mail.Username,
		Password:   conf.Mail.Password,
		DomainName: conf.DomainName,
	})
	if err != nil {
		log.Printf("mailer failed: %+v", err)
	} else {
		log.Printf("connected to mailer: %s:%d", conf.Mail.Host, conf.Mail.Port)
		v.Mailer.KeepAlive = true

		v.Mailer.AddDefaults(mail.Defaults{
			DefaultTo:   conf.Mail.DefaultMailTo,
			DefaultFrom: "YSTV Web Auth <wauth@ystv.co.uk>",
		})
	}

	// Initialising cache
	v.cache = cache.New(1*time.Hour, 1*time.Hour)

	// Initialising session cookie
	authKey, _ := hex.DecodeString(conf.Security.AuthenticationKey)
	if len(authKey) == 0 {
		authKey = securecookie.GenerateRandomKey(64)
	}
	encryptionKey, _ := hex.DecodeString(conf.Security.EncryptionKey)
	if len(encryptionKey) == 0 {
		encryptionKey = securecookie.GenerateRandomKey(32)
	}
	v.cookie = sessions.NewCookieStore(
		authKey,
		encryptionKey,
	)
	v.cookie.Options = &sessions.Options{
		MaxAge:   60 * 60 * 24,
		HttpOnly: true,
		Path:     "/",
	}

	// So we can use our struct in the cookie
	gob.Register(user.User{})

	v.conf = conf

	// Struct validator
	v.validate = validator.New()

	go func() {
		for {
			err = v.api.DeleteOldToken(context.Background())
			if err != nil {
				log.Printf("failed to delete old token func: %+v", err)
			}
			time.Sleep(30 * time.Second)
		}
	}()

	return v
}

// errorHandle is for handling errors and presenting them to the user in a nice format
func (v *Views) errorHandle(c echo.Context, err error) error {
	data := struct {
		Error string
	}{
		Error: err.Error(),
	}
	log.Println(data.Error)
	return v.template.RenderNoNavsTemplate(c.Response().Writer, data, templates.ErrorTemplate)
}
