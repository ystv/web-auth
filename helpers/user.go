package helpers

import (
	"github.com/gorilla/sessions"
	"github.com/ystv/web-auth/types"
)

// GetUser returns a user from a session on
// error returns an empty unauthenticated user
func GetUser(s *sessions.Session) types.User {
	val := s.Values["user"]
	var user = types.User{}
	user, ok := val.(types.User)
	if !ok {
		return types.User{Authenticated: false}
	}
	return user
}
