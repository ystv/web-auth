package views

import (
	"fmt"
	"net/http"

	"github.com/dustin/go-humanize"
	"github.com/ystv/web-auth/helpers"
)

type InternalTemplate struct {
	Nickname      string
	LastLogin     string
	TotalUsers    int
	LoginsPastDay int
}

// InternalFunc handles a request to the internal template
func InternalFunc(w http.ResponseWriter, r *http.Request) {
	session, err := cStore.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c := getData(session)
	err = uStore.GetUser(r.Context(), &c.User)
	if err != nil {
		err = fmt.Errorf("failed to get user: %w", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	ctx := InternalTemplate{
		Nickname:      c.User.Nickname,
		LastLogin:     humanize.Time(c.User.LastLogin),
		TotalUsers:    2000,
		LoginsPastDay: 20,
	}
	err = tpl.ExecuteTemplate(w, "internal.gohtml", ctx)
}

// RequiresLogin is a middleware which will be used for each
// httpHandler to check if there is any active session
func RequiresLogin(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := cStore.Get(r, "session")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if !helpers.GetUser(session).Authenticated {
			// Not authenticated
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		h.ServeHTTP(w, r)
	}
}
