package user

import (
	"context"
	//nolint:goimports
	"fmt"
	//nolint:goimports
	sq "github.com/Masterminds/squirrel"
	//nolint:goimports
	"github.com/jinzhu/copier"
	//nolint:goimports
	"github.com/ystv/web-auth/permission"
	//nolint:goimports
	"github.com/ystv/web-auth/role"
	//nolint:goimports
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
func (s *Store) updateUser(ctx context.Context, u User) error {
	builder := sq.Update("people.users").
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
		Where(sq.Eq{"user_id": u.UserID}).
		PlaceholderFormat(sq.Dollar)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for updateUser: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

// getUser will get a user using any unique identity fields for a user
func (s *Store) getUser(ctx context.Context, u1 User) (User, error) {
	var u User
	builder := sq.Select("*").
		From("people.users").
		Where(sq.Or{
			sq.And{sq.Eq{"username": u1.Username}, sq.NotEq{"username": ""}},
			sq.And{sq.Eq{"email": u1.Email}, sq.NotEq{"email": ""}},
			sq.And{sq.Eq{"ldap_username": u1.LDAPUsername}, sq.NotEq{"ldap_username": ""}},
			sq.Eq{"user_id": u1.UserID}}).
		Limit(1).
		PlaceholderFormat(sq.Dollar)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for updateUser: %w", err))
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
	builder := sq.Select(
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
			return nil, -1, fmt.Errorf("invalid sorting direction, entered \"%s\" of length %d, but expected either \"direction\" or \"desc\"", direction, len(direction))
		}
	}
	if page >= 1 && size >= 5 && size <= 100 {
		builder = builder.Limit(uint64(size)).Offset(uint64(size * (page - 1)))
	}
	builder = builder.PlaceholderFormat(sq.Dollar)
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
