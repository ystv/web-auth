package views

import (
	"log"
	"net/http"

	"github.com/ystv/web-auth/helpers"
)

// InternalFunc handles a request to the internal template
func InternalFunc(w http.ResponseWriter, r *http.Request) {
	session, err := cStore.Get(r, "session")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c := getData(session)
	err = tpl.ExecuteTemplate(w, "internal.gohtml", c)
	log.Print(err)
}

//RequiresLogin is a middleware which will be used for each
//httpHandler to check if there is any active session
func RequiresLogin(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
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
		handler(w, r)
	}
}
