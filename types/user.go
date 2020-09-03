package types

import "time"

type (
	// User represents relevant user fields
	User struct {
		UserID        int       `db:"user_id" json:"id"`
		Username      string    `db:"username" json:"username" schema:"username"`
		Nickname      string    `db:"nickname" schema:"nickname"`
		Firstname     string    `db:"first_name" json:"firstName" schema:"firstname"`
		Lastname      string    `db:"last_name" json:"lastName" schema:"lastname"`
		Password      string    `db:"password" json:"-" schema:"password"`
		Salt          string    `db:"salt" json:"-"`
		Email         string    `db:"email" json:"email" schema:"email"`
		LastLogin     time.Time `db:"last_login"`
		ResetPw       bool      `db:"reset_pw" json:"-"`
		Authenticated bool
		Permissions   []Permission `json:"permissions"`
	}

	// Permission represents an individual permission. Attempting to implement some RBAC here.
	Permission struct {
		ID   int    `db:"permission_id" json:"id"`
		Name string `db:"name" json:"name"`
	}
)
