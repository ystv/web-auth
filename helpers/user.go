package helpers

import (
	"github.com/ystv/web-auth/user"

	"github.com/gorilla/sessions"
)

// GetUser returns a user from a session on
// error returns an empty unauthenticated user
func GetUser(s *sessions.Session) user.User {
	val := s.Values["user"]
	//fmt.Println(val)
	var u = user.User{}
	u, ok := val.(user.User)
	if !ok {
		return user.User{Authenticated: false}
	}
	return u
}
