package user

import (
	"context"
	"fmt"
	"time"
)

// countUsers will get the number of total users
func (s *Store) countUsers(ctx context.Context) (int, error) {
	count := 0
	err := s.db.GetContext(ctx, &count,
		`SELECT COUNT(*)
		FROM people.users;`)
	if err != nil {
		return count, fmt.Errorf("failed to count users from db: %w", err)
	}
	return count, nil
}

// countUsers24Hours will get the number of users in the last 24 hours
// TODO make this a lot less bad, it just is bad
func (s *Store) countUsers24Hours(ctx context.Context) (int, error) {
	count := 0
	err := s.db.GetContext(ctx, &count,
		`SELECT COUNT(*)
		FROM people.users
		WHERE last_login > timestamp '`+time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04:05")+`';`)
	if err != nil {
		return count, fmt.Errorf("failed to count users 24 hours from db: %w", err)
	}
	return count, nil
}

// countUsersPastYear will get the number of users in the last 24 hours
// TODO make this a lot less bad, it just is bad
func (s *Store) countUsersPastYear(ctx context.Context) (int, error) {
	count := 0
	err := s.db.GetContext(ctx, &count,
		`SELECT COUNT(*)
		FROM people.users
		WHERE last_login > timestamp '`+time.Now().AddDate(-1, 0, 0).Format("2006-01-02 15:04:05")+`';`)
	if err != nil {
		return count, fmt.Errorf("failed to count users 24 hours from db: %w", err)
	}
	return count, nil
}

// updateUser will update a user record by ID
func (s *Store) updateUser(ctx context.Context, user User) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE people.users
		SET password = $1,
			salt = $2,
			email = $3,
			last_login = $4,
			reset_pw = $5,
			avatar = $6,
			use_gravatar = $7,
			first_name = $8,
			last_name = $9,
			nickname = $10
		WHERE user_id = $11;`, user.Password, user.Salt, user.Email, user.LastLogin, user.ResetPw, user.Avatar, user.UseGravatar, user.Firstname, user.Lastname, user.Nickname, user.UserID)
	if err != nil {
		return err
	}
	return nil
}

// getUser will get a user using any unique identity fields for a user
func (s *Store) getUser(ctx context.Context, user User) (User, error) {
	u := User{}
	err := s.db.GetContext(ctx, &u,
		`SELECT user_id, username, nickname, first_name, last_name, email, last_login, salt, password, avatar, use_gravatar
		FROM people.users
		WHERE (username = $1 AND username != '') OR (email = $2 AND email != '') OR user_id = $3
		LIMIT 1;`, user.Username, user.Email, user.UserID)
	if err != nil {
		return u, fmt.Errorf("failed to get user from db: %w", err)
	}
	return u, nil
}

// getUsers will get a group of users
func (s *Store) getUsers(ctx context.Context) ([]User, error) {
	var u []User
	err := s.db.SelectContext(ctx, &u,
		`SELECT user_id, username, nickname, first_name, last_name, email, last_login, avatar, use_gravatar
	FROM people.users;`)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getPermissions returns all permissions for a user
func (s *Store) getPermissions(ctx context.Context, u User) ([]string, error) {
	var p []string
	err := s.db.SelectContext(ctx, &p, `SELECT p.name
	FROM people.permissions p
	INNER JOIN people.role_permissions rp ON rp.permission_id = p.permission_id
	INNER JOIN people.role_members rm ON rm.role_id = rp.role_id
	WHERE rm.user_id = $1;`, u.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	return p, nil
}
