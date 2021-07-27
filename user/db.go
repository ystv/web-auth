package user

import (
	"context"
	"fmt"
)

// updateUser will update a user record by ID
func (s *Store) updateUser(ctx context.Context, user User) error {
	_, err := s.db.ExecContext(ctx,
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

// getUser will get a user using any unique identity fields for a user
func (s *Store) getUser(ctx context.Context, user User) (User, error) {
	u := User{}
	err := s.db.GetContext(ctx, &u,
		`SELECT user_id, username, nickname, email, last_login, salt, password
		FROM people.users
		WHERE username = $1 AND username != '' OR email = $2 AND email != '' OR user_id = $3
		LIMIT 1;`, user.Username, user.Email, user.UserID)
	if err != nil {
		return u, fmt.Errorf("failed to get user from db: %w", err)
	}
	return u, nil
}

// getUsers will get a group of users
func (s *Store) getUsers(ctx context.Context) ([]User, error) {
	u := []User{}
	err := s.db.SelectContext(ctx, &u,
		`s.dbuser_id, username, nickname, first_name, last_name, email, last_login
	FROM people.users;`)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getPermissions returns all permissions for a user
func (s *Store) getPermissions(ctx context.Context, u User) ([]Permission, error) {
	p := []Permission{}
	err := s.db.SelectContext(ctx, &p, `SELECT p.permission_id, p.name
	FROM people.permissions p
	INNER JOIN people.role_permissions rp ON rp.permission_id = p.permission_id
	INNER JOIN people.role_members rm ON rm.role_id = rp.role_id
	WHERE rm.user_id = $1;`, u.UserID)
	if err != nil {
		return nil, err
	}
	return p, nil
}
