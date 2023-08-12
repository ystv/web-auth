package user

import (
	"context"
	"fmt"
	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/role"
	"time"
)

// countUsers will get the number of total users
func (s *Store) countUsers(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM people.users;`)
	if err != nil {
		return count, fmt.Errorf("failed to count users from db: %w", err)
	}
	return count, nil
}

// countUsersActive will get the number of total active users
func (s *Store) countUsersActive(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, `SELECT COUNT(*)
		FROM people.users
		WHERE enabled = true AND deleted_by IS NULL AND deleted_at IS NULL;`)
	if err != nil {
		return count, fmt.Errorf("failed to count users from db: %w", err)
	}
	return count, nil
}

// countUsers24Hours will get the number of users in the last 24 hours
func (s *Store) countUsers24Hours(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, `SELECT COUNT(*)
		FROM people.users
		WHERE last_login > TO_TIMESTAMP($1, 'YYYY-MM-DD HH24:MI:SS');`, time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04:05"))
	if err != nil {
		return count, fmt.Errorf("failed to count users 24 hours from db: %w", err)
	}
	return count, nil
}

// countUsersPastYear will get the number of users in the last 24 hours
func (s *Store) countUsersPastYear(ctx context.Context) (int, error) {
	var count int
	err := s.db.GetContext(ctx, &count, `SELECT COUNT(*)
		FROM people.users
		WHERE last_login > TO_TIMESTAMP($1, 'YYYY-MM-DD HH24:MI:SS');`, time.Now().AddDate(-1, 0, 0).Format("2006-01-02 15:04:05"))
	if err != nil {
		return count, fmt.Errorf("failed to count users past year from db: %w", err)
	}
	return count, nil
}

// addUser will add a user
func (s *Store) addUser(ctx context.Context, u1 User) (User, error) {
	var u User
	stmt, err := s.db.PrepareNamedContext(ctx, "INSERT INTO people.users (username, university_username, email, first_name, last_name, nickname, login_type, password, salt, reset_pw, enabled, created_at, created_by) VALUES (:username, :university_username, :email, :first_name, :last_name, :nickname, :login_type, :password, :salt, :reset_pw, :enabled, :created_at, :created_by) RETURNING user_id, username, university_username, email, first_name, last_name, nickname, login_type, password, salt, reset_pw, enabled, created_at, created_by")
	if err != nil {
		return User{}, fmt.Errorf("failed to add user: %w", err)
	}
	err = stmt.Get(&u, u1)
	if err != nil {
		return User{}, fmt.Errorf("failed to add user: %w", err)
	}
	return u, nil
}

// editUser will update a user record by ID
func (s *Store) editUser(ctx context.Context, u User) (User, error) {
	stmt, err := s.db.NamedExecContext(ctx, `UPDATE people.users
		SET password = :password,
			salt = :salt,
			email = :email,
			last_login = :last_login,
			reset_pw = :reset_pw,
			avatar = :avatar,
			username = :username,
			use_gravatar = :use_gravatar,
			first_name = :first_name,
			last_name = :last_name,
			nickname = :nickname,
			university_username = :university_username,
			ldap_username = :ldap_username,
			login_type = :login_type,
			enabled = :enabled,
			updated_by = :updated_by,
			updated_at = :updated_at,
			deleted_by = :deleted_by,
			deleted_at = :deleted_at
		WHERE user_id = :user_id;`, u)
	if err != nil {
		return User{}, fmt.Errorf("failed to edit user: %w", err)
	}
	rows, err := stmt.RowsAffected()
	if err != nil {
		return User{}, fmt.Errorf("failed to edit user: %w", err)
	}
	if rows < 1 {
		return User{}, fmt.Errorf("failed to edit user: invalid rows affected: %d", rows)
	}
	return u, nil
}

// getUser will get a user using any unique identity fields for a user
func (s *Store) getUser(ctx context.Context, u1 User) (User, error) {
	var u User
	err := s.db.GetContext(ctx, &u, `SELECT *
		FROM people.users
		WHERE (username = $1 AND username != '') OR (email = $2 AND email != '') OR (ldap_username = $3 AND ldap_username != '') OR user_id = $4
		LIMIT 1;`, u1.Username, u1.Email, u1.LDAPUsername, u1.UserID)
	if err != nil {
		return u, fmt.Errorf("error at getUser: %w", err)
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
		return nil, fmt.Errorf("error at getUsers: %w", err)
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
		    (LOWER(CAST(user_id AS TEXT)) LIKE LOWER('%%' || $1 || '%%')
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
		return nil, fmt.Errorf("error at getUsersSearchNoOrder: %w", err)
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
		return nil, fmt.Errorf("error at getUsersOrderNoSearch: %w", err)
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
		FROM people.users
		WHERE
		    ((LOWER(CAST(user_id AS TEXT)) LIKE LOWER('%%' || $1 || '%%'))
			OR (LOWER(username) LIKE LOWER('%%' || $1 || '%%'))
			OR (LOWER(nickname) LIKE LOWER('%%' || $1 || '%%'))
			OR (LOWER(first_name) LIKE LOWER('%%' || $1 || '%%'))
			or (LOWER(last_name) LIKE LOWER('%%' || $1 || '%%'))
			OR (LOWER(email) LIKE LOWER('%%' || $1 || '%%'))
			OR (LOWER(first_name || ' ' || last_name) LIKE LOWER('%%' || $1 || '%%')))
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
		return nil, fmt.Errorf("error at getUsersSearchOrder: %w", err)
	}
	return u, nil
}

// parseDirection parses the string for asc/desc and returns the SQL equivalent for it
// No parameter is outputted into SQL for prevention of SQL injections
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

// parseEnabled parses the string for users enabled and returns the SQL equivalent for it
// No parameter is outputted into SQL for prevention of SQL injections
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

// parseDeleted parses the string for user deleted and returns the SQL equivalent for it
// No parameter is outputted into SQL for prevention of SQL injections
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

// parsePageSize parses the page and size for pagination and returns the SQL equivalent for it
// No parameter is outputted into SQL without conditioning for prevention of SQL injections
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

// getRoleUser returns a role user - moved here for cycle import reasons
func (s *Store) getRoleUser(ctx context.Context, ru1 RoleUser) (RoleUser, error) {
	var ru RoleUser
	err := s.db.GetContext(ctx, &ru, `SELECT *
		FROM people.role_members
		WHERE role_id = $1 AND user_id = $2
		LIMIT 1`, ru1.RoleID, ru1.UserID)
	if err != nil {
		return RoleUser{}, fmt.Errorf("failed to get roleUser: %w", err)
	}
	return ru, nil
}

// getUsersNotInRole returns all the users not currently in the role.Role to be added
func (s *Store) getUsersNotInRole(ctx context.Context, r role.Role) ([]User, error) {
	var u []User
	err := s.db.SelectContext(ctx, &u, `SELECT DISTINCT u.*
		FROM people.users u
        WHERE user_id NOT IN
        (SELECT u.user_id
		FROM people.users u
		LEFT JOIN people.role_members ru on u.user_id = ru.user_id
		WHERE ru.role_id = $1) AND deleted_by IS NOT NULL
		ORDER BY first_name, last_name`, r.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get users not in role: %w", err)
	}
	return u, nil
}

// addRoleUser creates a link between a role.Role and User
func (s *Store) addRoleUser(ctx context.Context, ru1 RoleUser) (RoleUser, error) {
	var ru RoleUser
	stmt, err := s.db.PrepareNamedContext(ctx, "INSERT INTO people.role_members (role_id, user_id) VALUES (:role_id, :user_id) RETURNING role_id, user_id")
	if err != nil {
		return RoleUser{}, fmt.Errorf("failed to add roleUser: %w", err)
	}
	err = stmt.Get(&ru, ru1)
	if err != nil {
		return RoleUser{}, fmt.Errorf("failed to add roleUser: %w", err)
	}
	return ru, nil
}

// removeRoleUser removes a link between a role.Role and User
func (s *Store) removeRoleUser(ctx context.Context, ru RoleUser) error {
	_, err := s.db.NamedExecContext(ctx, `DELETE FROM people.role_members WHERE role_id = :role_id AND user_id = :user_id`, ru)
	if err != nil {
		return fmt.Errorf("failed to remove roleUser: %w", err)
	}
	return nil
}

// removeRoleUser removes all links between role.Role and a User
func (s *Store) removeRoleUsers(ctx context.Context, u User) error {
	_, err := s.db.NamedExecContext(ctx, `DELETE FROM people.role_members WHERE user_id = :user_id`, u)
	if err != nil {
		return fmt.Errorf("failed to remove roleUsers: %w", err)
	}
	return nil
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

// getRolePermission returns a role permission - moved here for cycle import reasons
func (s *Store) getRolePermission(ctx context.Context, rp1 RolePermission) (RolePermission, error) {
	var rp RolePermission
	err := s.db.GetContext(ctx, &rp, `SELECT *
		FROM people.role_permissions
		WHERE role_id = $1 AND permission_id = $2
		LIMIT 1`, rp1.RoleID, rp1.PermissionID)
	if err != nil {
		return RolePermission{}, fmt.Errorf("failed to get rolePermisison: %w", err)
	}
	return rp, nil
}

// getPermissionsNotInRole returns all the permissions not currently in the role.Role to be added
func (s *Store) getPermissionsNotInRole(ctx context.Context, r role.Role) ([]permission.Permission, error) {
	var p []permission.Permission
	err := s.db.SelectContext(ctx, &p, `SELECT DISTINCT p.*
		FROM people.permissions p
        WHERE permission_id NOT IN
        (SELECT p.permission_id
		FROM people.permissions p
		LEFT JOIN people.role_permissions rp on p.permission_id = rp.permission_id
		WHERE rp.role_id = $1)
		ORDER BY name`, r.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions not in role: %w", err)
	}
	return p, nil
}

// addRolePermission creates a link between a role.Role and permission.Permission
func (s *Store) addRolePermission(ctx context.Context, rp1 RolePermission) (RolePermission, error) {
	var rp RolePermission
	stmt, err := s.db.PrepareNamedContext(ctx, "INSERT INTO people.role_permissions (role_id, permission_id) VALUES (:role_id, :permission_id) RETURNING role_id, permission_id")
	if err != nil {
		return RolePermission{}, fmt.Errorf("failed to add rolePermission: %w", err)
	}
	err = stmt.Get(&rp, rp1)
	if err != nil {
		return RolePermission{}, fmt.Errorf("failed to add rolePermission: %w", err)
	}
	return rp, nil
}

// removeRolePermission removes a link between a role.Role and permission.Permission
func (s *Store) removeRolePermission(ctx context.Context, rp RolePermission) error {
	_, err := s.db.NamedExecContext(ctx, `DELETE FROM people.role_permissions WHERE role_id = :role_id AND permission_id = :permission_id`, rp)
	if err != nil {
		return fmt.Errorf("failed to remove rolePermission: %w", err)
	}
	return nil
}
