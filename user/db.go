package user

import (
	"context"
	"fmt"
	"github.com/ystv/web-auth/role"
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
func (s *Store) countUsers24Hours(ctx context.Context) (int, error) {
	count := 0
	err := s.db.GetContext(ctx, &count,
		`SELECT COUNT(*)
		FROM people.users
		WHERE last_login > TO_TIMESTAMP($1, 'YYYY-MM-DD HH24:MI:SS');`, time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04:05"))
	if err != nil {
		return count, fmt.Errorf("failed to count users 24 hours from db: %w", err)
	}
	return count, nil
}

// countUsersPastYear will get the number of users in the last 24 hours
func (s *Store) countUsersPastYear(ctx context.Context) (int, error) {
	count := 0
	err := s.db.GetContext(ctx, &count,
		`SELECT COUNT(*)
		FROM people.users
		WHERE last_login > TO_TIMESTAMP($1, 'YYYY-MM-DD HH24:MI:SS');`, time.Now().AddDate(-1, 0, 0).Format("2006-01-02 15:04:05"))
	if err != nil {
		return count, fmt.Errorf("failed to count users past year from db: %w", err)
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
		`SELECT *
		FROM people.users
		WHERE (username = $1 AND username != '') OR (email = $2 AND email != '') OR (ldap_username = $3 AND ldap_username != '') OR user_id = $4
		LIMIT 1;`, user.Username, user.Email, user.LDAPUsername, user.UserID)
	if err != nil {
		return u, fmt.Errorf("failed to get user from db: %w", err)
	}
	return u, nil
}

// getUsers will get users
func (s *Store) getUsers(ctx context.Context) ([]User, error) {
	var u []User
	err := s.db.SelectContext(ctx, &u,
		`SELECT *
		FROM people.users;`)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getUsersSizePage will get users with page size
func (s *Store) getUsersSizePage(ctx context.Context, size, page int) ([]User, error) {
	var u []User
	// SELECT u.*, COUNT(*) / $1 AS pages
	err := s.db.SelectContext(ctx, &u,
		`SELECT u.*
		FROM people.users u
-- 		GROUP BY u, user_id, username, university_username, email, first_name, last_name, nickname, login_type, password, salt, avatar, last_login, reset_pw, enabled, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by, use_gravatar, ldap_username
		LIMIT $1
		OFFSET $2;`, size, size*(page-1))
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getUsersSearch will get users search
func (s *Store) getUsersSearch(ctx context.Context, search string) ([]User, error) {
	var u []User
	err := s.db.SelectContext(ctx, &u,
		`SELECT *
		FROM people.users
		WHERE (CAST(user_id AS TEXT) LIKE '%' || $1 || '%')
		   OR (username LIKE '%' || $1 || '%')
		   OR (nickname LIKE '%' || $1 || '%')
		   OR (first_name LIKE '%' || $1 || '%')
		   or (last_name LIKE '%' || $1 || '%')
		   OR (email LIKE '%' || $1 || '%')
		   OR (first_name || ' ' || last_name LIKE '%' || $1 || '%');`, search)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getUsersSearchSizePage will get users search with size and page
func (s *Store) getUsersSearchSizePage(ctx context.Context, search string, size, page int) ([]User, error) {
	var u []User
	err := s.db.SelectContext(ctx, &u,
		`SELECT *
		FROM people.users
		WHERE (CAST(user_id AS TEXT) LIKE '%' || $1 || '%')
		   OR (username LIKE '%' || $1 || '%')
		   OR (nickname LIKE '%' || $1 || '%')
		   OR (first_name LIKE '%' || $1 || '%')
		   or (last_name LIKE '%' || $1 || '%')
		   OR (email LIKE '%' || $1 || '%')
		   OR (first_name || ' ' || last_name LIKE '%' || $1 || '%')
-- 		GROUP BY u, user_id, username, university_username, email, first_name, last_name, nickname, login_type, password, salt, avatar, last_login, reset_pw, enabled, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by, use_gravatar, ldap_username
		LIMIT $2
		OFFSET $3;`, search, size, size*(page-1))
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getUsersOptionsDesc will get users sorting asc
func (s *Store) getUsersOptionsAsc(ctx context.Context, sortBy string) ([]User, error) {
	var u []User
	err := s.db.SelectContext(ctx, &u,
		`SELECT *
		FROM people.users
		ORDER BY
		    CASE WHEN $1 = 'userId' THEN user_id END ASC,
		    CASE WHEN $1 = 'name' THEN first_name END ASC,
			CASE WHEN $1 = 'name' THEN last_name END ASC,
		    CASE WHEN $1 = 'username' THEN username END ASC,
		    CASE WHEN $1 = 'email' THEN email END ASC,
		    CASE WHEN $1 = 'lastLogin' THEN last_login END ASC NULLS FIRST;`, sortBy)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getUsersOptionsDescSizePage will get users sorting asc with size and page
func (s *Store) getUsersOptionsAscSizePage(ctx context.Context, sortBy string, size, page int) ([]User, error) {
	var u []User
	err := s.db.SelectContext(ctx, &u,
		`SELECT *
		FROM people.users
		ORDER BY
		    CASE WHEN $1 = 'userId' THEN user_id END ASC,
		    CASE WHEN $1 = 'name' THEN first_name END ASC,
			CASE WHEN $1 = 'name' THEN last_name END ASC,
		    CASE WHEN $1 = 'username' THEN username END ASC,
		    CASE WHEN $1 = 'email' THEN email END ASC,
		    CASE WHEN $1 = 'lastLogin' THEN last_login END ASC NULLS FIRST
		LIMIT $2
		OFFSET $3;`, sortBy, size, size*(page-1))
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getUsersSearchOptionsAsc will get users search with sorting asc
func (s *Store) getUsersSearchOptionsAsc(ctx context.Context, search, sortBy string) ([]User, error) {
	var u []User
	err := s.db.SelectContext(ctx, &u,
		`SELECT *
		FROM people.users
		WHERE (CAST(user_id AS TEXT) LIKE '%' || $1 || '%')
		   OR (username LIKE '%' || $1 || '%')
		   OR (nickname LIKE '%' || $1 || '%')
		   OR (first_name LIKE '%' || $1 || '%')
		   or (last_name LIKE '%' || $1 || '%')
		   OR (email LIKE '%' || $1 || '%')
		   OR (first_name || ' ' || last_name LIKE '%' || $1 || '%')
		ORDER BY
		    CASE WHEN $2 = 'userId' THEN user_id END ASC,
		    CASE WHEN $2 = 'name' THEN first_name END ASC,
			CASE WHEN $2 = 'name' THEN last_name END ASC,
		    CASE WHEN $2 = 'username' THEN username END ASC,
		    CASE WHEN $2 = 'email' THEN email END ASC,
		    CASE WHEN $2 = 'lastLogin' THEN last_login END ASC NULLS FIRST;`, search, sortBy)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getUsersSearchOptionsAscSizePage will get users search with sorting asc with size and page
func (s *Store) getUsersSearchOptionsAscSizePage(ctx context.Context, search, sortBy string, size, page int) ([]User, error) {
	var u []User
	err := s.db.SelectContext(ctx, &u,
		`SELECT *
		FROM people.users
		WHERE (CAST(user_id AS TEXT) LIKE '%' || $1 || '%')
		   OR (username LIKE '%' || $1 || '%')
		   OR (nickname LIKE '%' || $1 || '%')
		   OR (first_name LIKE '%' || $1 || '%')
		   or (last_name LIKE '%' || $1 || '%')
		   OR (email LIKE '%' || $1 || '%')
		   OR (first_name || ' ' || last_name LIKE '%' || $1 || '%')
		ORDER BY
		    CASE WHEN $2 = 'userId' THEN user_id END ASC,
		    CASE WHEN $2 = 'name' THEN first_name END ASC,
			CASE WHEN $2 = 'name' THEN last_name END ASC,
		    CASE WHEN $2 = 'username' THEN username END ASC,
		    CASE WHEN $2 = 'email' THEN email END ASC,
		    CASE WHEN $2 = 'lastLogin' THEN last_login END ASC NULLS FIRST
		LIMIT $3
		OFFSET $4;`, search, sortBy, size, size*(page-1))
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getUsersOptionsDesc will get users sorting desc
func (s *Store) getUsersOptionsDesc(ctx context.Context, sortBy string) ([]User, error) {
	var u []User
	err := s.db.SelectContext(ctx, &u,
		`SELECT *
		FROM people.users
		ORDER BY
		    CASE WHEN $1 = 'userId' THEN user_id END DESC,
		    CASE WHEN $1 = 'name' THEN first_name END DESC,
			CASE WHEN $1 = 'name' THEN last_name END DESC,
		    CASE WHEN $1 = 'username' THEN username END DESC,
		    CASE WHEN $1 = 'email' THEN email END DESC,
		    CASE WHEN $1 = 'lastLogin' THEN last_login END DESC NULLS LAST;`, sortBy)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getUsersOptionsDescSizePage will get users sorting desc with size and page
func (s *Store) getUsersOptionsDescSizePage(ctx context.Context, sortBy string, size, page int) ([]User, error) {
	var u []User
	err := s.db.SelectContext(ctx, &u,
		`SELECT *
		FROM people.users
		ORDER BY
		    CASE WHEN $1 = 'userId' THEN user_id END DESC,
		    CASE WHEN $1 = 'name' THEN first_name END DESC,
			CASE WHEN $1 = 'name' THEN last_name END DESC,
		    CASE WHEN $1 = 'username' THEN username END DESC,
		    CASE WHEN $1 = 'email' THEN email END DESC,
		    CASE WHEN $1 = 'lastLogin' THEN last_login END DESC NULLS LAST
		LIMIT $2
		OFFSET $3;`, sortBy, size, size*(page-1))
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getUsersSearchOptionsDesc will get users search with sorting desc
func (s *Store) getUsersSearchOptionsDesc(ctx context.Context, search, sortBy string) ([]User, error) {
	var u []User
	err := s.db.SelectContext(ctx, &u,
		`SELECT *
		FROM people.users
		WHERE (CAST(user_id AS TEXT) LIKE '%' || $1 || '%')
		   OR (username LIKE '%' || $1 || '%')
		   OR (nickname LIKE '%' || $1 || '%')
		   OR (first_name LIKE '%' || $1 || '%')
		   or (last_name LIKE '%' || $1 || '%')
		   OR (email LIKE '%' || $1 || '%')
		   OR (first_name || ' ' || last_name LIKE '%' || $1 || '%')
		ORDER BY
		    CASE WHEN $2 = 'userId' THEN user_id END DESC,
		    CASE WHEN $2 = 'name' THEN first_name END DESC,
			CASE WHEN $2 = 'name' THEN last_name END DESC,
		    CASE WHEN $2 = 'username' THEN username END DESC,
		    CASE WHEN $2 = 'email' THEN email END DESC,
		    CASE WHEN $2 = 'lastLogin' THEN last_login END DESC NULLS LAST;`, search, sortBy)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getUsersSearchOptionsDescSizePage will get users search with sorting desc with size and page
func (s *Store) getUsersSearchOptionsDescSizePage(ctx context.Context, search, sortBy string, size, page int) ([]User, error) {
	var u []User
	err := s.db.SelectContext(ctx, &u,
		`SELECT *
		FROM people.users
		WHERE (CAST(user_id AS TEXT) LIKE '%' || $1 || '%')
		   OR (username LIKE '%' || $1 || '%')
		   OR (nickname LIKE '%' || $1 || '%')
		   OR (first_name LIKE '%' || $1 || '%')
		   or (last_name LIKE '%' || $1 || '%')
		   OR (email LIKE '%' || $1 || '%')
		   OR (first_name || ' ' || last_name LIKE '%' || $1 || '%')
		ORDER BY
		    CASE WHEN $2 = 'userId' THEN user_id END DESC,
		    CASE WHEN $2 = 'name' THEN first_name END DESC,
			CASE WHEN $2 = 'name' THEN last_name END DESC,
		    CASE WHEN $2 = 'username' THEN username END DESC,
		    CASE WHEN $2 = 'email' THEN email END DESC,
		    CASE WHEN $2 = 'lastLogin' THEN last_login END DESC NULLS LAST
		LIMIT $3
		OFFSET $4;`, search, sortBy, size, size*(page-1))
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getPermissionsForUser returns all permissions for a user
func (s *Store) getPermissionsForUser(ctx context.Context, u User) (p []string, err error) {
	err = s.db.SelectContext(ctx, &p, `SELECT p.name
		FROM people.permissions p
		INNER JOIN people.role_permissions rp ON rp.permission_id = p.permission_id
		INNER JOIN people.role_members rm ON rm.role_id = rp.role_id
		WHERE rm.user_id = $1;`, u.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	return p, nil
}

// getRolesForUser returns all roles for a user
func (s *Store) getRolesForUser(ctx context.Context, u User) (r []role.Role, err error) {
	err = s.db.SelectContext(ctx, &r, `SELECT r.*
		FROM people.roles r
		INNER JOIN people.role_members rm ON rm.role_id = r.role_id
		WHERE rm.user_id = $1;`, u.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return r, nil
}
