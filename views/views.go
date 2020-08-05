package views

import (
	"encoding/gob"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/ystv/web-auth/db"
	"github.com/ystv/web-auth/types"
	"github.com/ystv/web-auth/user"
)

var (
	uStore *user.Store

	cStore *sessions.CookieStore

	tpl *template.Template
)

func init() {
	dbStore, err := db.NewStore(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("NewStore failed: %+v", dbStore)
	}
	uStore = user.NewUserStore(dbStore)

	authKeyOne := securecookie.GenerateRandomKey(64)
	encryptionKeyOne := securecookie.GenerateRandomKey(32)

	cStore = sessions.NewCookieStore(
		authKeyOne,
		encryptionKeyOne,
	)

	cStore.Options = &sessions.Options{
		MaxAge:   60 * 15,
		HttpOnly: true,
		Path:     "/",
	}

	gob.Register(types.User{})

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
		tpl.ExecuteTemplate(w, "index.gohtml", context)
	}
}

// Context is a struct that is applied to the templates.
type Context struct {
	Message string
	Version string
	User    types.User
}

func getData(s *sessions.Session) *Context {
	val := s.Values["user"]
	u := types.User{}
	u, ok := val.(types.User)
	if !ok {
		u = types.User{Authenticated: false}
	}
	c := Context{Version: "0.4.1",
		Message: "Auth service: now with postgres support",
		User:    u,
	}
	return &c
}
