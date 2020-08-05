package db

import (
	"context"

	"github.com/ystv/web-auth/types"
)

// UpdateUser will update a user record by ID
func (store *DB) UpdateUser(ctx context.Context, user *types.User) error {
	_, err := store.ExecContext(ctx,
		`UPDATE people.users
		SET password = $1,
			salt = $2,
			email = $3,
			reset_pw = $4
		WHERE user_id = $5;`, user.Password, user.Salt, user.Email, user.ResetPw)
	if err != nil {
		return err
	}
	return nil
}

// GetUser will get a user using any unique identity fields for a user
func (store *DB) GetUser(ctx context.Context, user *types.User) error {
	return store.GetContext(ctx, user,
		`SELECT user_id, username, email, salt, password
		FROM people.users
		WHERE username = $1 OR email = $1
		LIMIT 1;`)
}
