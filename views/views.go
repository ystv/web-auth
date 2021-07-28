package views

import (
	"encoding/gob"
	"encoding/hex"
	"html/template"
	"io/fs"
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
		Version        string
		DatabaseURL    string
		DomainName     string
		LogoutEndpoint string
		Mail           SMTPConfig
		Security       SecurityConfig
	}
	// SMTPConfig stores the SMTP mailer configuration
	SMTPConfig struct {
		Host     string
		Username string
		Password string
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
		ResetFunc(w http.ResponseWriter, r *http.Request)
		// internal
		InternalFunc(w http.ResponseWriter, r *http.Request)
		UsersFunc(w http.ResponseWriter, r *http.Request)
		UserFunc(w http.ResponseWriter, r *http.Request)
		// middleware
		RequiresLogin(h http.Handler) http.HandlerFunc
		// api
		ValidateToken(myToken string) (bool, *JWTClaims)
		SetTokenHandler(w http.ResponseWriter, r *http.Request)
		TestAPI(w http.ResponseWriter, r *http.Request)
		getJWTCookie(w http.ResponseWriter, r *http.Request) (http.ResponseWriter, error)
	}

	// Views encapsulates our view dependencies
	Views struct {
		conf     Config
		user     *user.Store
		cookie   *sessions.CookieStore
		tpl      *template.Template
		mail     *mail.Mail
		cache    *cache.Cache
		validate *validator.Validate
	}

	// Notifcation template for messages
	Notifcation struct {
		Title   string
		Type    string
		Message string
	}
)

// here to verify we are meeting the interface
var _ Repo = &Views{}

// New initialises connections, templates, and cookies
func New(conf Config, templates fs.FS) *Views {
	v := Views{}
	// Connecting to stores
	dbStore, err := db.NewStore(conf.DatabaseURL)
	if err != nil {
		log.Fatalf("NewStore failed: %+v", err)
	}
	v.user = user.NewUserRepo(dbStore)

	// Connecting to mail
	v.mail, err = mail.NewMailer(mail.Config{
		Host:       conf.Mail.Host,
		Port:       587,
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

	// Loading templates
	v.tpl = template.Must(template.ParseFS(templates, "*.tmpl"))

	v.conf = conf

	// Struct validator
	v.validate = validator.New()

	return &v
}

// Context is a struct that is applied to the templates.
type Context struct {
	Message  string
	MsgType  string
	Version  string
	Callback string
	User     user.User
}

func (v Views) getData(s *sessions.Session) *Context {
	val := s.Values["user"]
	u := user.User{}
	u, ok := val.(user.User)
	if !ok {
		u = user.User{Authenticated: false}
	}
	c := Context{Version: v.conf.Version,
		Callback: "/internal",
		User:     u,
	}
	return &c
}
