package user

import (
	"context"
	"errors"

	"github.com/ystv/web-auth/types"
	"github.com/ystv/web-auth/utils"
)

type (
	// IStore where all user data is stored
	IStore interface {
		GetUser(ctx context.Context, u *types.User) error
		UpdateUser(ctx context.Context, u *types.User) error
		GetPermissions(ctx context.Context, u *types.User) error
	}
	// Store stores the dependencies
	Store struct {
		userStore IStore
	}
)

// NewUserStore stores a dependency
func NewUserStore(store IStore) *Store {
	return &Store{store}
}

// GetUser returns a user using any unique identity fields
func (s *Store) GetUser(ctx context.Context, u *types.User) error {
	return s.userStore.GetUser(ctx, u)
}

// GetPermissions returns all permissions of a user
func (s *Store) GetPermissions(ctx context.Context, u *types.User) error {
	return s.userStore.GetPermissions(ctx, u)
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
	return errors.New("Invalid credentials")
}

// UpdateUserPassword will update the password and set the reset_pw to false
func (s *Store) UpdateUserPassword(ctx context.Context, u *types.User) error {
	plaintext := u.Password
	s.GetUser(ctx, u)
	u.Password = utils.HashPass(u.Salt + plaintext)
	u.ResetPw = false
	return s.userStore.UpdateUser(ctx, u)
}
