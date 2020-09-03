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
			last_login = $4,
			reset_pw = $5
		WHERE user_id = $6;`, user.Password, user.Salt, user.Email, user.LastLogin, user.ResetPw, user.UserID)
	if err != nil {
		return err
	}
	return nil
}

// GetUser will get a user using any unique identity fields for a user
func (store *DB) GetUser(ctx context.Context, user *types.User) error {
	return store.GetContext(ctx, user,
		`SELECT user_id, username, nickname, email, last_login, salt, password
		FROM people.users
		WHERE username = $1 AND username != '' OR email = $2 AND email != '' OR user_id = $3
		LIMIT 1;`, user.Username, user.Email, user.UserID)
}

// GetPermissions returns all permissions for a user
func (store *DB) GetPermissions(ctx context.Context, u *types.User) error {
	return store.SelectContext(ctx, &u.Permissions, `SELECT p.permission_id, p.name
	FROM people.permissions p
	INNER JOIN people.role_permissions rp ON rp.permission_id = p.permission_id
	INNER JOIN people.role_members rm ON rm.role_id = rp.role_id
	WHERE rm.user_id = $1;`, u.UserID)
}
