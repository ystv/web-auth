package views

import (
	"net/http"

	"github.com/rmil/web-auth/sessions"
)

// InternalFunc handles a request to the internal template
func InternalFunc(w http.ResponseWriter, r *http.Request) {
	c := getData(r)
	internalTemplate.Execute(w, c)
}

//RequiresLogin is a middleware which will be used for each
//httpHandler to check if there is any active session
func RequiresLogin(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !sessions.IsLoggedIn(r) {
			http.Redirect(w, r, "/login/", 302)
			return
		}
		handler(w, r)
	}
}
