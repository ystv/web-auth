package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Clarilab/gocloaksession"
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
		db    *sqlx.DB
		cloak *gocloaksession.GoCloakSession
	}
	// User represents relevant user fields
	User struct {
		UserID        int       `db:"user_id" json:"id"`
		Username      string    `db:"username" json:"username" schema:"username"`
		Nickname      string    `db:"nickname" json:"nickname" schema:"nickname"`
		Firstname     string    `db:"first_name" json:"firstName" schema:"firstname"`
		Lastname      string    `db:"last_name" json:"lastName" schema:"lastname"`
		Password      string    `db:"password" json:"-" schema:"password"`
		Salt          string    `db:"salt" json:"-"`
		Avatar        string    `db:"avatar" json:"avatar" schema:"avatar"`
		Email         string    `db:"email" json:"email" schema:"email"`
		LastLogin     null.Time `db:"last_login"`
		ResetPw       bool      `db:"reset_pw" json:"-"`
		Authenticated bool
		Permissions   []Permission `json:"permissions"`
		UseGravatar   bool         `db:"use_gravatar" json:"useGravatar" schema:"useGravatar"`
	}

	// Permission represents an individual permission. Attempting to implement some RBAC here.
	Permission struct {
		Name string `db:"name" json:"name"`
	}
)

// NewUserRepo stores our dependency
func NewUserRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

// GetUser returns a user using any unique identity fields
func (s *Store) GetUser(ctx context.Context, u User) (User, error) {
	return s.getUser(ctx, u)
}

// GetUsers returns a group of users, used for administration
func (s *Store) GetUsers(ctx context.Context) ([]User, error) {
	return s.getUsers(ctx)
}

// GetPermissions returns all permissions of a user
func (s *Store) GetPermissions(ctx context.Context, u User) ([]string, error) {
	return s.getPermissions(ctx, u)
}

// VerifyUser will check that that the password is correct with provided
// credentials and if verified will return the User object
func (s *Store) VerifyUser(ctx context.Context, u User) (User, error) {
	user, err := s.getUser(ctx, u)
	if err != nil {
		return u, fmt.Errorf("failed to get user: %w", err)
	}
	if utils.HashPass(user.Salt+u.Password) == user.Password {
		return user, nil
	}
	return u, errors.New("invalid credentials")
}

// UpdateUserPassword will update the password and set the reset_pw to false
func (s *Store) UpdateUserPassword(ctx context.Context, u User) (User, error) {
	user, err := s.GetUser(ctx, u)
	if err != nil {
		return u, fmt.Errorf("failed to get user: %w", err)
	}
	user.Password = utils.HashPass(user.Salt + u.Password)
	user.ResetPw = false
	err = s.updateUser(ctx, user)
	if err != nil {
		return u, fmt.Errorf("failed to update user: %w", err)
	}
	return user, nil
}

// SetUserLoggedIn will set the last login date to now
func (s *Store) SetUserLoggedIn(ctx context.Context, u User) error {
	u.LastLogin = null.TimeFrom(time.Now())
	return s.updateUser(ctx, u)
}

func (s *Store) CheckUserType(ctx context.Context, u User) error {
	return nil
}
