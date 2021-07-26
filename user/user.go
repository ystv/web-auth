package user

import (
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/ystv/web-auth/utils"
	"gopkg.in/guregu/null.v4"
)

type (
	// Repo where all user data is stored
	Repo interface {
		GetUser(ctx context.Context, u User) error
		GetUsers(ctx context.Context, u User) error
		UpdateUser(ctx context.Context, u User) error
		GetPermissions(ctx context.Context, u User) error
		CheckUserType(ctx context.Context, u User) error
	}
	// Store stores the dependencies
	Store struct {
		db *sqlx.DB
	}
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
		LastLogin     null.Time `db:"last_login"`
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

// NewUserRepo stores our dependency
func NewUserRepo(db *sqlx.DB) *Store {
	return &Store{db}
}

// GetUser returns a user using any unique identity fields
func (s *Store) GetUser(ctx context.Context, u User) (User, error) {
	return s.GetUser(ctx, u)
}

// GetUsers returns a group of users, used for administration
func (s *Store) GetUsers(ctx context.Context) ([]User, error) {
	return s.getUsers(ctx)
}

// GetPermissions returns all permissions of a user
func (s *Store) GetPermissions(ctx context.Context, u User) ([]Permission, error) {
	return s.getPermissions(ctx, u)
}

// VerifyUser will check that that the password is correct with provided
// credentials
func (s *Store) VerifyUser(ctx context.Context, u User) error {
	plaintext := u.Password
	user, err := s.getUser(ctx, u)
	if err != nil {
		return err
	}
	if utils.HashPass(user.Salt+plaintext) == user.Password {
		return nil
	}
	return errors.New("invalid credentials")
}

// UpdateUserPassword will update the password and set the reset_pw to false
func (s *Store) UpdateUserPassword(ctx context.Context, u User) error {
	plaintext := u.Password
	s.GetUser(ctx, u)
	u.Password = utils.HashPass(u.Salt + plaintext)
	u.ResetPw = false
	return s.updateUser(ctx, u)
}

// SetUserLoggedIn will set the last login date to now
func (s *Store) SetUserLoggedIn(ctx context.Context, u User) error {
	u.LastLogin = null.TimeFrom(time.Now())
	return s.updateUser(ctx, u)
}

func (s *Store) CheckUserType(ctx context.Context, u User) error {
	return nil
}
