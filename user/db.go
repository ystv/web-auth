package user

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jinzhu/copier"

	"github.com/ystv/web-auth/permission"
	"github.com/ystv/web-auth/role"
	"github.com/ystv/web-auth/utils"
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

// addUser will add a user
func (s *Store) addUser(ctx context.Context, u1 User) (User, error) {
	var u User
	stmt, err := s.db.PrepareNamedContext(ctx, "INSERT INTO people.users (username, university_username, email, first_name, last_name, nickname, login_type, password, salt, reset_pw, enabled, created_at, created_by) VALUES (:username, :university_username, :email, :first_name, :last_name, :nickname, :login_type, :password, :salt, :reset_pw, :enabled, :created_at, :created_by) RETURNING user_id, username, university_username, email, first_name, last_name, nickname, login_type, password, salt, reset_pw, enabled, created_at, created_by")
	if err != nil {
		return User{}, fmt.Errorf("failed to add user: %w", err)
	}
	defer stmt.Close()
	err = stmt.Get(&u, u1)
	if err != nil {
		return User{}, fmt.Errorf("failed to add user: %w", err)
	}
	return u, nil
}

// editUser will edit a user record by ID
func (s *Store) editUser(ctx context.Context, u User) error {
	builder := utils.PSQL().Update("people.users").
		SetMap(map[string]interface{}{"password": u.Password,
			"salt":                u.Salt,
			"email":               u.Email,
			"last_login":          u.LastLogin,
			"reset_pw":            u.ResetPw,
			"avatar":              u.Avatar,
			"use_gravatar":        u.UseGravatar,
			"first_name":          u.Firstname,
			"nickname":            u.Nickname,
			"last_name":           u.Lastname,
			"username":            u.Username,
			"university_username": u.UniversityUsername,
			"ldap_username":       u.LDAPUsername,
			"login_type":          u.LoginType,
			"enabled":             u.Enabled,
			"updated_by":          u.UpdatedBy,
			"updated_at":          u.UpdatedAt,
			"deleted_by":          u.DeletedBy,
			"deleted_at":          u.DeletedAt,
		}).
		Where(sq.Eq{"user_id": u.UserID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editUser: %w", err))
	}
	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to edit user: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to edit user: %w", err)
	}
	if rows < 1 {
		return fmt.Errorf("failed to edit user: invalid rows affected: %d, this user may not exist: %d", rows, u.UserID)
	}
	return nil
}

// getUser will get a user using any unique identity fields for a user
func (s *Store) getUser(ctx context.Context, u1 User) (User, error) {
	var u User
	builder := utils.PSQL().Select("*").
		From("people.users").
		Where(sq.Or{
			sq.And{sq.Eq{"username": u1.Username}, sq.NotEq{"username": ""}},
			sq.And{sq.Eq{"email": u1.Email}, sq.NotEq{"email": ""}},
			sq.And{sq.Eq{"ldap_username": u1.LDAPUsername}, sq.NotEq{"ldap_username": ""}},
			sq.Eq{"user_id": u1.UserID}}).
		Limit(1)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getUser: %w", err))
	}
	err = s.db.GetContext(ctx, &u, sql, args...)
	if err != nil {
		return u, fmt.Errorf("failed to get user from db: %w", err)
	}
	return u, nil
}

// getUsers will get users search with sorting with size and page, enabled and deleted
// Use the parameter direction for determining of the sorting will be ascending(asc) or descending(desc)
func (s *Store) getUsers(ctx context.Context, size, page int, search, sortBy, direction, enabled, deleted string) ([]User, int, error) {
	var u []User
	var count int
	builder, err := s._getUsersBuilder(size, page, search, sortBy, direction, enabled, deleted)
	if err != nil {
		return nil, -1, fmt.Errorf("failed to build sql for getUsers: %w", err)
	}
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getUsers: %w", err))
	}
	rows, err := s.db.QueryxContext(ctx, sql, args...)
	if err != nil {
		return nil, -1, fmt.Errorf("failed to get db users: %w", err)
	}

	defer rows.Close()

	type tempStruct struct {
		User
		Count int `db:"full_count" json:"fullCount"`
	}

	for rows.Next() {
		var u1 User
		var temp tempStruct
		err = rows.StructScan(&temp)
		if err != nil {
			return nil, -1, fmt.Errorf("failed to get db users: %w", err)
		}
		count = temp.Count
		err = copier.Copy(&u1, &temp)
		if err != nil {
			return nil, -1, fmt.Errorf("failed to copy struct: %w", err)
		}
		u = append(u, u1)
	}
	return u, count, nil
}

func (s *Store) _getUsersBuilder(size, page int, search, sortBy, direction, enabled, deleted string) (*sq.SelectBuilder, error) {
	builder := utils.PSQL().Select(
		"*",
		"count(*) OVER() AS full_count",
	).
		From("people.users")
	if len(search) > 0 {
		builder = builder.Where(
			"(CAST(user_id AS TEXT) LIKE '%' || ? || '%' "+
				"OR LOWER(username) LIKE LOWER('%' || ? || '%') "+
				"OR LOWER(nickname) LIKE LOWER('%' || ? || '%') "+
				"OR LOWER(first_name) LIKE LOWER('%' || ? || '%') "+
				"OR LOWER(last_name) LIKE LOWER('%' || ? || '%') "+
				"OR LOWER(email) LIKE LOWER('%' || ? || '%') "+
				"OR LOWER(first_name || ' ' || last_name) LIKE LOWER('%' || ? || '%'))", search, search, search, search, search, search, search)
	}
	switch enabled {
	case "enabled":
		builder = builder.Where(sq.Eq{"enabled": true})
		break
	case "disabled":
		builder = builder.Where(sq.Eq{"enabled": false})
		break
	}
	switch deleted {
	case "not_deleted":
		builder = builder.Where(sq.Eq{"deleted_by": nil})
		break
	case "deleted":
		builder = builder.Where(sq.NotEq{"deleted_by": nil})
	}
	if len(sortBy) > 0 && len(direction) > 0 {
		switch direction {
		case "asc":
			builder = builder.OrderByClause(
				"CASE WHEN ? = 'userId' THEN user_id END ASC, "+
					"CASE WHEN ? = 'name' THEN first_name END ASC, "+
					"CASE WHEN ? = 'name' THEN last_name END ASC, "+
					"CASE WHEN ? = 'username' THEN username END ASC, "+
					"CASE WHEN ? = 'email' THEN email END ASC, "+
					"CASE WHEN ? = 'lastLogin' THEN last_login END ASC NULLS FIRST", sortBy, sortBy, sortBy, sortBy, sortBy, sortBy)
			break
		case "desc":
			builder = builder.OrderByClause(
				"CASE WHEN ? = 'userId' THEN user_id END DESC, "+
					"CASE WHEN ? = 'name' THEN first_name END DESC, "+
					"CASE WHEN ? = 'name' THEN last_name END DESC, "+
					"CASE WHEN ? = 'username' THEN username END DESC, "+
					"CASE WHEN ? = 'email' THEN email END DESC, "+
					"CASE WHEN ? = 'lastLogin' THEN last_login END DESC NULLS LAST", sortBy, sortBy, sortBy, sortBy, sortBy, sortBy)
			break
		default:
			return nil, fmt.Errorf("invalid sorting direction, entered \"%s\" of length %d, but expected either \"direction\" or \"desc\"", direction, len(direction))
		}
	}
	if page >= 1 && size >= 5 && size <= 100 {
		builder = builder.Limit(uint64(size)).Offset(uint64(size * (page - 1)))
	}

	return &builder, nil
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
		WHERE ru.role_id = $1) AND deleted_by IS NULL
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
	defer stmt.Close()
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
	defer stmt.Close()
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
