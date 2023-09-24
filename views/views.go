package views

import (
	"context"
	"encoding/gob"
	"encoding/hex"
	"log"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/patrickmn/go-cache"

	"github.com/ystv/web-auth/api"
	"github.com/ystv/web-auth/infrastructure/db"
	"github.com/ystv/web-auth/infrastructure/mail"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/role"
	"github.com/ystv/web-auth/templates"
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
		Host       string
		Username   string
		Password   string
		Port       int
		DomainName string
	}

	// SecurityConfig stores the security configuration
	SecurityConfig struct {
		EncryptionKey     string
		AuthenticationKey string
		SigningKey        string
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
		mailer     *mail.MailerInit
		validate   *validator.Validate
	}

	TemplateHelper struct {
		UserPermissions []permission.Permission
		ActivePage      string
	}
)

// New initialises connections, templates, and cookies
func New(conf *Config, host string) *Views {
	v := &Views{}
	// Connecting to stores
	dbStore := db.NewStore(conf.DatabaseURL, host)
	v.permission = permission.NewPermissionRepo(dbStore)
	v.role = role.NewRoleRepo(dbStore)
	v.user = user.NewUserRepo(dbStore)
	v.api = api.NewAPIRepo(dbStore)

	v.template = templates.NewTemplate(v.permission, v.role, v.user)

	// Initialising cache
	v.cache = cache.New(1*time.Hour, 1*time.Hour)

	// Initialise mailer
	v.mailer = mail.NewMailer(mail.Config{
		Host:       conf.Mail.Host,
		Port:       conf.Mail.Port,
		Username:   conf.Mail.Username,
		Password:   conf.Mail.Password,
		DomainName: conf.Mail.DomainName,
	})

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
	gob.Register(InternalContext{})

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
