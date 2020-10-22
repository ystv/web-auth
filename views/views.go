package views

import (
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/patrickmn/go-cache"
	"github.com/ystv/web-auth/infrastructure/db"
	"github.com/ystv/web-auth/infrastructure/mail"
	"github.com/ystv/web-auth/types"
	"github.com/ystv/web-auth/user"
)

var (
	uStore *user.Store

	cStore *sessions.CookieStore

	tpl *template.Template

	m *mail.Mail

	c *cache.Cache
)

// New initialises connections, templates, and cookies
func New() {
	// Connecting to stores
	dbStore, err := db.NewStore(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("NewStore failed: %+v", err)
	}
	uStore = user.NewUserRepo(dbStore)

	// Connecting to mail
	m, err = mail.NewMailer(mail.Config{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     587,
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
	})
	if err != nil {
		log.Printf("mailer failed: %+v", err)
	}

	// Initialising cache
	c = cache.New(1*time.Hour, 1*time.Hour)

	// Initialising session cookie
	authKey, _ := hex.DecodeString(os.Getenv("AUTHENTICATION_KEY"))
	if len(authKey) == 0 {
		authKey = securecookie.GenerateRandomKey(64)
	}
	encryptionKey, _ := hex.DecodeString(os.Getenv("ENCRYPTION_KEY"))
	if len(encryptionKey) == 0 {
		encryptionKey = securecookie.GenerateRandomKey(32)
	}
	cStore = sessions.NewCookieStore(
		authKey,
		encryptionKey,
	)
	cStore.Options = &sessions.Options{
		MaxAge:   60 * 60 * 24,
		HttpOnly: true,
		Path:     "/",
	}

	// So we can use our struct in the cookie
	gob.Register(types.User{})

	// Loading templates
	tpl = template.Must(template.ParseGlob("public/templates/*.gohtml"))
}

// IndexFunc handles the index page.
func IndexFunc(w http.ResponseWriter, r *http.Request) {
	session, err := cStore.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Data for our HTML template
	context := getData(session)

	// Check if there is a callback request
	callback := r.URL.Query().Get("callback")
	if callback != "" && strings.HasSuffix(callback, os.Getenv("DOMAIN_NAME")) {
		context.Callback = callback
	}
	// Check if authenticated
	if context.User.Authenticated {
		http.Redirect(w, r, context.Callback, http.StatusFound)
		return
	}
	loginCallback := fmt.Sprintf("login?callback=%s", context.Callback)
	http.Redirect(w, r, loginCallback, http.StatusFound)
}

// Context is a struct that is applied to the templates.
type Context struct {
	Message  string
	MsgType  string
	Version  string
	Callback string
	User     types.User
}

func getData(s *sessions.Session) *Context {
	val := s.Values["user"]
	u := types.User{}
	u, ok := val.(types.User)
	if !ok {
		u = types.User{Authenticated: false}
	}
	c := Context{Version: "0.4.8",
		Message:  "News: Removed trailing slash & relative routing",
		Callback: "/internal",
		User:     u,
	}
	return &c
}
