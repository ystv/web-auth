package user

import (
	"context"
	"errors"
	"time"

	"github.com/ystv/web-auth/types"
	"github.com/ystv/web-auth/utils"
	"gopkg.in/guregu/null.v4"
)

type (
	// Repo where all user data is stored
	Repo interface {
		GetUser(ctx context.Context, u *types.User) error
		GetUsers(ctx context.Context, u *[]types.User) error
		UpdateUser(ctx context.Context, u *types.User) error
		GetPermissions(ctx context.Context, u *types.User) error
	}
	// Store stores the dependencies
	Store struct {
		users Repo
	}
)

// NewUserRepo stores our dependency
func NewUserRepo(store Repo) *Store {
	return &Store{store}
}

// GetUser returns a user using any unique identity fields
func (s *Store) GetUser(ctx context.Context, u *types.User) error {
	return s.users.GetUser(ctx, u)
}

// GetUsers returns a group of users, used for administration
func (s *Store) GetUsers(ctx context.Context, u *[]types.User) error {
	return s.users.GetUsers(ctx, u)
}

// GetPermissions returns all permissions of a user
func (s *Store) GetPermissions(ctx context.Context, u *types.User) error {
	return s.users.GetPermissions(ctx, u)
}

// VerifyUser will check that that the password is correct with provided
// credentials
func (s *Store) VerifyUser(ctx context.Context, u *types.User) error {
	plaintext := u.Password
	err := s.GetUser(ctx, u)
	if err != nil {
		return err
	}
	if utils.HashPass(u.Salt+plaintext) == u.Password {
		return nil
	}
	return errors.New("invalid credentials")
}

// UpdateUserPassword will update the password and set the reset_pw to false
func (s *Store) UpdateUserPassword(ctx context.Context, u *types.User) error {
	plaintext := u.Password
	s.GetUser(ctx, u)
	u.Password = utils.HashPass(u.Salt + plaintext)
	u.ResetPw = false
	return s.users.UpdateUser(ctx, u)
}

// SetUserLoggedIn will set the last login date to now
func (s *Store) SetUserLoggedIn(ctx context.Context, u *types.User) error {
	u.LastLogin = null.TimeFrom(time.Now())
	return s.users.UpdateUser(ctx, u)
}
