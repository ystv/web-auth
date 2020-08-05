package types

// User represents relevant user fields
type User struct {
	UserID        int    `db:"user_id" json:"id"`
	Username      string `db:"username" json:"username" schema:"username"`
	Password      string `db:"password" json:"-" schema:"password"`
	Salt          string `db:"salt" json:"-"`
	Email         string `db:"email" json:"email"`
	ResetPw       bool   `db:"reset_pw" json:"-"`
	Authenticated bool
}
