package user

import (
	"context"
	"encoding/hex"
	"errors"

	whirl "github.com/balacode/zr-whirl"
	"github.com/ystv/web-auth/types"
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
	return s.GetUser(ctx, u)
}

// GetPermissions returns all permissions of a user
func (s *Store) GetPermissions(ctx context.Context, u *types.User) error {
	return s.GetPermissions(ctx, u)
}

// VerifyUser will check that that the password is correct with provided
// credentials
func (s *Store) VerifyUser(ctx context.Context, u *types.User) error {
	plaintext := u.Password
	s.GetUser(ctx, u)
	if hashPass(u.Salt+plaintext) == u.Password {
		return nil
	}
	return errors.New("Invalid credentials")
}

func hashPass(password string) string {
	iter := 1000
	var next string
	for i := 0; i < iter; i++ {
		next += password
		tmp := whirl.HashOfBytes([]byte(next), []byte(""))
		next = hex.EncodeToString(tmp)
	}
	return next
}

// func hashPass(pass []byte) ([]byte, error) {
// 	pass, err := bcrypt.GenerateFromPassword(pass, 10)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return pass, nil
// }

// func checkPassHash(hash, pass []byte) error {
// 	return bcrypt.CompareHashAndPassword(hash, pass)
// }
