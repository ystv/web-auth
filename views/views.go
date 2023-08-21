package views

import (
	"encoding/gob"
	"encoding/hex"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/role"
	"github.com/ystv/web-auth/templates"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/patrickmn/go-cache"
	"github.com/ystv/web-auth/infrastructure/db"
	"github.com/ystv/web-auth/user"
)

type (
	// Config the global web-auth configuration
	Config struct {
		Version           string
		Debug             bool
		Address           string
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
		Host     string
		Username string
		Password string
		Port     int
	}

	// SecurityConfig stores the security configuration
	SecurityConfig struct {
		EncryptionKey     string
		AuthenticationKey string
		SigningKey        string
	}

	// Views encapsulates our view dependencies
	Views struct {
		conf       *Config
		permission *permission.Store
		role       *role.Store
		user       *user.Store
		cookie     *sessions.CookieStore
		cache      *cache.Cache
		validate   *validator.Validate
		template   *templates.Templater
	}
)

// New initialises connections, templates, and cookies
func New(conf *Config, host string) *Views {
	v := &Views{}
	// Connecting to stores
	dbStore := db.NewStore(conf.DatabaseURL, host, conf.Debug)
	v.permission = permission.NewPermissionRepo(dbStore)
	v.role = role.NewRoleRepo(dbStore)
	v.user = user.NewUserRepo(dbStore)

	v.template = templates.NewTemplate(v.permission, v.role, v.user)

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

	return v
}
