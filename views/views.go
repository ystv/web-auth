package views

import (
	"encoding/gob"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"os"
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
	uStore = user.NewUserStore(dbStore)

	// Connecting to mail
	m, err = mail.NewMailer(mail.Config{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     587,
		Username: os.Getenv("SMTP_USERNAME"),
		Password: os.Getenv("SMTP_PASSWORD"),
	})
	if err != nil {
		log.Fatal(err)
	}

	// Initialising cache
	c = cache.New(1*time.Hour, 1*time.Hour)

	// Initialising session cookie
	authKey, err := hex.DecodeString(os.Getenv("AUTHENTICATION_KEY"))
	if err != nil {
		authKey = securecookie.GenerateRandomKey(64)
	}
	encryptionKey, err := hex.DecodeString(os.Getenv("ENCRYPTION_KEY"))
	if err != nil {
		encryptionKey = securecookie.GenerateRandomKey(32)
	}
	cStore = sessions.NewCookieStore(
		authKey,
		encryptionKey,
	)
	cStore.Options = &sessions.Options{
		MaxAge:   60 * 15,
		HttpOnly: true,
		Path:     "/",
	}

	// So we can use our struct in the cookie
	gob.Register(types.User{})

	// Loading templates
	tpl = template.Must(template.ParseGlob("public/templates/*.gohtml"))
}

// IndexFunc handles the welcome/index page.
func IndexFunc(w http.ResponseWriter, r *http.Request) {
	session, err := cStore.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	context := getData(session)
	if r.Method == "GET" {
		callback := r.URL.Query().Get("callback")
		if context.User.Authenticated {
			http.Redirect(w, r, context.Callback, http.StatusFound)
		}
		context.Callback = callback
		tpl.ExecuteTemplate(w, "index.gohtml", context)
	}
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
	c := Context{Version: "0.4.6",
		Message:  "News: now with callback support!",
		Callback: "/internal",
		User:     u,
	}
	return &c
}
