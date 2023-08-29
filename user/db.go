package user

import (
	"context"
	"fmt"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/role"
	"time"
)

// countUsersAll will get the number of total users
func (s *Store) countUsersAll(ctx context.Context) (CountUsers, error) {
	var countUsers CountUsers
	err := s.db.GetContext(ctx, &countUsers,
		`SELECT
		(SELECT COUNT(*) FROM people.users) as total_users,
		(SELECT COUNT(*) FROM people.users WHERE enabled = true AND deleted_by IS NULL AND deleted_at IS NULL) as active_users,
		(SELECT COUNT(*) FROM people.users WHERE last_login > TO_TIMESTAMP($1, 'YYYY-MM-DD HH24:MI:SS')) as active_users_past_24_hours,
		(SELECT COUNT(*) FROM people.users WHERE last_login > TO_TIMESTAMP($2, 'YYYY-MM-DD HH24:MI:SS')) as active_users_past_year;`,
		time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04:05"),
		time.Now().AddDate(-1, 0, 0).Format("2006-01-02 15:04:05"))
	if err != nil {
		return countUsers, fmt.Errorf("failed to count users all from db: %w", err)
	}
	return countUsers, nil
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
			nickname = $10,
			university_username = $11,
			ldap_username = $12,
			login_type = $13,
			enabled = $14,
			updated_by = $15,
			updated_at = $16,
			deleted_by = $17,
			deleted_at = $18
		WHERE user_id = $19;`, user.Password, user.Salt, user.Email, user.LastLogin, user.ResetPw, user.Avatar, user.UseGravatar, user.Firstname, user.Lastname, user.Nickname, user.UniversityUsername, user.LDAPUsername, user.LoginType, user.Enabled, user.UpdatedBy, user.UpdatedAt, user.DeletedBy, user.DeletedAt, user.UserID)
	if err != nil {
		return err
	}
	return nil
}

// getUser will get a user using any unique identity fields for a user
func (s *Store) getUser(ctx context.Context, user User) (User, error) {
	var u User
	err := s.db.GetContext(ctx, &u, `SELECT *
		FROM people.users
		WHERE (username = $1 AND username != '') OR (email = $2 AND email != '') OR (ldap_username = $3 AND ldap_username != '') OR user_id = $4
		LIMIT 1;`, user.Username, user.Email, user.LDAPUsername, user.UserID)
	if err != nil {
		return u, fmt.Errorf("failed to get user from db: %w", err)
	}
	return u, nil
}

// getUsers will get users with page size
func (s *Store) getUsers(ctx context.Context, size, page int, enabled, deleted string) ([]User, error) {
	var u []User
	enabledSQL := s.parseEnabled(enabled, false)
	deletedSQL := s.parseDeleted(deleted, len(enabledSQL) > 0)
	var where string
	if len(enabledSQL) > 0 || len(deletedSQL) > 0 {
		where = `WHERE`
	}
	pageSize := s.parsePageSize(page, size)
	err := s.db.SelectContext(ctx, &u, fmt.Sprintf(`SELECT u.*
		FROM people.users u
		%[1]s
		%[2]s
		%[3]s
		%[4]s;`, where, enabledSQL, deletedSQL, pageSize))
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getUsersSearchNoOrder will get users search with size and page
func (s *Store) getUsersSearchNoOrder(ctx context.Context, size, page int, search, enabled, deleted string) ([]User, error) {
	var u []User
	enabledSQL := s.parseEnabled(enabled, true)
	deletedSQL := s.parseDeleted(deleted, true)
	pageSize := s.parsePageSize(page, size)
	err := s.db.SelectContext(ctx, &u, fmt.Sprintf(`SELECT *
		FROM people.users
		WHERE
		    (CAST(user_id AS TEXT) LIKE '%%' || $1 || '%%'
			OR LOWER(username) LIKE LOWER('%%' || $1 || '%%')
			OR LOWER(nickname) LIKE LOWER('%%' || $1 || '%%')
			OR LOWER(first_name) LIKE LOWER('%%' || $1 || '%%')
			or LOWER(last_name) LIKE LOWER('%%' || $1 || '%%')
			OR LOWER(email) LIKE LOWER('%%' || $1 || '%%')
			OR LOWER(first_name || ' ' || last_name) LIKE LOWER('%%' || $1 || '%%'))
			%[1]s
			%[2]s
		%[3]s;`, enabledSQL, deletedSQL, pageSize), search)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getUsersOrderNoSearch will get users sorting with size and page
// Use the parameter direction for determining of the sorting will be ascending(asc) or descending(desc)
func (s *Store) getUsersOrderNoSearch(ctx context.Context, size, page int, sortBy, direction, enabled, deleted string) ([]User, error) {
	var u []User
	dir, nulls, err := s.parseDirection(direction)
	if err != nil {
		return nil, err
	}
	enabledSQL := s.parseEnabled(enabled, false)
	deletedSQL := s.parseDeleted(deleted, len(enabledSQL) > 0)
	var where string
	if len(enabledSQL) > 0 || len(deletedSQL) > 0 {
		where = `WHERE`
	}
	pageSize := s.parsePageSize(page, size)
	err = s.db.SelectContext(ctx, &u, fmt.Sprintf(`SELECT *
		FROM people.users
		%[3]s
		%[4]s
		%[5]s
		ORDER BY
		    CASE WHEN $1 = 'userId' THEN user_id END %[1]s,
		    CASE WHEN $1 = 'name' THEN first_name END %[1]s,
			CASE WHEN $1 = 'name' THEN last_name END %[1]s,
		    CASE WHEN $1 = 'username' THEN username END %[1]s,
		    CASE WHEN $1 = 'email' THEN email END %[1]s,
		    CASE WHEN $1 = 'lastLogin' THEN last_login END %[1]s NULLS %[2]s
		%[6]s;`, dir, nulls, where, enabledSQL, deletedSQL, pageSize), sortBy)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// getUsersSearchOrder will get users search with sorting with size and page, enabled and deleted
// Use the parameter direction for determining of the sorting will be ascending(asc) or descending(desc)
func (s *Store) getUsersSearchOrder(ctx context.Context, size, page int, search, sortBy, direction, enabled, deleted string) ([]User, error) {
	var u []User
	dir, nulls, err := s.parseDirection(direction)
	if err != nil {
		return nil, err
	}
	enabledSQL := s.parseEnabled(enabled, true)
	deletedSQL := s.parseDeleted(deleted, true)
	pageSize := s.parsePageSize(page, size)
	err = s.db.SelectContext(ctx, &u, fmt.Sprintf(`SELECT *
		WHERE
		    (CAST(user_id AS TEXT) LIKE '%%' || $1 || '%%'
			OR LOWER(username) LIKE LOWER('%%' || $1 || '%%')
			OR LOWER(nickname) LIKE LOWER('%%' || $1 || '%%')
			OR LOWER(first_name) LIKE LOWER('%%' || $1 || '%%')
			or LOWER(last_name) LIKE LOWER('%%' || $1 || '%%')
			OR LOWER(email) LIKE LOWER('%%' || $1 || '%%')
			OR LOWER(first_name || ' ' || last_name) LIKE LOWER('%%' || $1 || '%%'))
			%[3]s
			%[4]s
		ORDER BY
		    CASE WHEN $2 = 'userId' THEN user_id END %[1]s,
		    CASE WHEN $2 = 'name' THEN first_name END %[1]s,
			CASE WHEN $2 = 'name' THEN last_name END %[1]s,
		    CASE WHEN $2 = 'username' THEN username END %[1]s,
		    CASE WHEN $2 = 'email' THEN email END %[1]s,
		    CASE WHEN $2 = 'lastLogin' THEN last_login END %[1]s NULLS %[2]s
	    %[5]s;`, dir, nulls, enabledSQL, deletedSQL, pageSize), search, sortBy)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (s *Store) parseDirection(direction string) (string, string, error) {
	var dir, nulls string
	if direction == "asc" {
		dir = `ASC`
		nulls = `FIRST`
	} else if direction == "desc" {
		dir = `DESC`
		nulls = `LAST`
	} else {
		return ``, ``, fmt.Errorf("invalid sorting direction, entered \"%s\" of length %d, but expected either \"direction\" or \"desc\"", direction, len(direction))
	}
	return dir, nulls, nil
}

func (s *Store) parseEnabled(enabled string, includeAND bool) string {
	if enabled == "enabled" {
		if includeAND {
			return `AND enabled`
		} else {
			return `enabled`
		}
	} else if enabled == "disabled" {
		if includeAND {
			return `AND NOT enabled`
		} else {
			return `NOT enabled`
		}
	}
	return ``
}

func (s *Store) parseDeleted(deleted string, includeAND bool) string {
	if deleted == "deleted" {
		if includeAND {
			return `AND deleted_by IS NOT NULL`
		} else {
			return `deleted_by IS NOT NULL`
		}
	} else if deleted == "not_deleted" {
		if includeAND {
			return `AND deleted_by IS NULL`
		} else {
			return `deleted_by IS NULL`
		}
	}
	return ``
}

func (s *Store) parsePageSize(page, size int) string {
	if page < 1 || size < 5 || size > 100 {
		return ``
	} else {
		return fmt.Sprintf(`LIMIT %d OFFSET %d`, size, size*(page-1))
	}
}

// getPermissionsForUser returns all permissions for a user
func (s *Store) getPermissionsForUser(ctx context.Context, u User) ([]permission.Permission, error) {
	var p []permission.Permission
	err := s.db.SelectContext(ctx, &p, `SELECT p.*
		FROM people.permissions p
		LEFT JOIN people.role_permissions rp ON rp.permission_id = p.permission_id
		LEFT JOIN people.role_members rm ON rm.role_id = rp.role_id
		WHERE rm.user_id = $1;`, u.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}
	return p, nil
}

// getRolesForUser returns all roles for a user
func (s *Store) getRolesForUser(ctx context.Context, u User) ([]role.Role, error) {
	var r []role.Role
	err := s.db.SelectContext(ctx, &r, `SELECT r.*
		FROM people.roles r
		LEFT JOIN people.role_members rm ON rm.role_id = r.role_id
		WHERE rm.user_id = $1;`, u.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}
	return r, nil
}

// getUsersForRole returns all users for a role - moved here for cycle import reasons
func (s *Store) getUsersForRole(ctx context.Context, r role.Role) ([]User, error) {
	var u []User
	err := s.db.SelectContext(ctx, &u, `SELECT u.*
		FROM people.users u
		LEFT JOIN people.role_members rm ON rm.user_id = u.user_id
		WHERE rm.role_id = $1;`, r.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role users: %w", err)
	}
	return u, nil
}

// getPermissionsForRole returns all permissions for a role - moved here for cycle import reasons
func (s *Store) getPermissionsForRole(ctx context.Context, r role.Role) ([]permission.Permission, error) {
	var p []permission.Permission
	err := s.db.SelectContext(ctx, &p, `SELECT p.*
		FROM people.permissions p
		LEFT JOIN people.role_permissions rp on p.permission_id = rp.permission_id
		WHERE rp.role_id = $1`, r.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role permissions: %w", err)
	}
	return p, nil
}

// getRolesForPermission returns all roles for a permission - moved here for cycle import reasons
func (s *Store) getRolesForPermission(ctx context.Context, p permission.Permission) ([]role.Role, error) {
	var r []role.Role
	err := s.db.SelectContext(ctx, &r, `SELECT r.*
		FROM people.roles r
		LEFT JOIN people.role_permissions rp on r.role_id = rp.role_id
		WHERE rp.permission_id = $1`, p.PermissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission roles: %w", err)
	}
	return r, nil
}
