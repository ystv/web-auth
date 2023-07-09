package views

import (
	"encoding/gob"
	"encoding/hex"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/public/templates"
	"github.com/ystv/web-auth/role"
	"log"
	"net/http"
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

	// Repo defines all view interactions
	Repo interface {
		// index
		IndexFunc(w http.ResponseWriter, r *http.Request)
		// login
		LogoutFunc(w http.ResponseWriter, r *http.Request)
		LoginFunc(w http.ResponseWriter, r *http.Request)
		SignUpFunc(w http.ResponseWriter, r *http.Request)
		ForgotFunc(w http.ResponseWriter, r *http.Request)
		ResetURLFunc(w http.ResponseWriter, r *http.Request)
		//ResetFunc(w http.ResponseWriter, r *http.Request)
		// internal
		InternalFunc(w http.ResponseWriter, r *http.Request)
		UsersFunc(w http.ResponseWriter, r *http.Request)
		UserFunc(w http.ResponseWriter, r *http.Request)
		// middleware
		RequiresLogin(h http.Handler) http.HandlerFunc
		// api
		SetTokenHandler(w http.ResponseWriter, r *http.Request)
		ValidateToken(myToken string) (bool, *JWTClaims)
		newJWT(u user.User) (string, error)
		TestAPI(w http.ResponseWriter, r *http.Request)
	}

	// Views encapsulates our view dependencies
	Views struct {
		conf       Config
		permission *permission.Store
		role       *role.Store
		user       *user.Store
		cookie     *sessions.CookieStore
		mailer     *mail.Mailer
		cache      *cache.Cache
		validate   *validator.Validate
		template   *templates.Templater
	}
)

// here to verify we are meeting the interface
var _ Repo = &Views{}

// New initialises connections, templates, and cookies
func New(conf Config) *Views {
	v := Views{}
	// Connecting to stores
	dbStore, err := db.NewStore(conf.DatabaseURL)
	if err != nil {
		log.Fatalf("NewStore failed: %+v", err)
	}

	v.permission = permission.NewPermissionRepo(dbStore)
	v.role = role.NewRoleRepo(dbStore)
	v.user = user.NewUserRepo(dbStore)

	v.template = templates.NewTemplate(v.permission, v.role, v.user)

	// Connecting to mail
	v.mailer, err = mail.NewMailer(mail.Config{
		Host:       conf.Mail.Host,
		Port:       conf.Mail.Port,
		Username:   conf.Mail.Username,
		Password:   conf.Mail.Password,
		DomainName: conf.DomainName,
	})
	if err != nil {
		log.Printf("mailer failed: %+v", err)
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

	return &v
}
