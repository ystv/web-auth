package user

import (
	"context"
	"strconv"

	//nolint:gosec
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
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
		(SELECT COUNT(*) FROM people.users WHERE enabled = true AND deleted_by IS NULL AND deleted_at IS NULL)
		    AS active_users,
		(SELECT COUNT(*) FROM people.users WHERE last_login > TO_TIMESTAMP($1, 'YYYY-MM-DD HH24:MI:SS'))
		    AS active_users_past_24_hours,
		(SELECT COUNT(*) FROM people.users WHERE last_login > TO_TIMESTAMP($2, 'YYYY-MM-DD HH24:MI:SS'))
		    AS active_users_past_year;`,
		time.Now().AddDate(0, 0, -1).Format("2006-01-02 15:04:05"),
		time.Now().AddDate(-1, 0, 0).Format("2006-01-02 15:04:05"))

	if err != nil {
		return countUsers, fmt.Errorf("failed to count users all from db: %w", err)
	}

	return countUsers, nil
}

// addUser will add a user
func (s *Store) addUser(ctx context.Context, u User) (User, error) {
	builder := utils.PSQL().Insert("people.users").
		Columns("username", "university_username", "email", "first_name", "last_name", "nickname",
			"login_type", "password", "salt", "reset_pw", "enabled", "created_at", "created_by").
		Values(u.Username, u.UniversityUsername, u.Email, u.Firstname, u.Lastname, u.Nickname, u.LoginType, u.Password,
			u.Salt, u.ResetPw, u.Enabled, u.CreatedAt, u.CreatedBy).
		Suffix("RETURNING user_id")

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addUser: %w", err))
	}

	stmt, err := s.db.PrepareContext(ctx, sql)
	if err != nil {
		return User{}, fmt.Errorf("failed to add user: %w", err)
	}

	defer stmt.Close()

	err = stmt.QueryRow(args...).Scan(&u.UserID)
	if err != nil {
		return User{}, fmt.Errorf("failed to add user: %w", err)
	}

	return u, nil
}

// editUser will edit a user record by ID
func (s *Store) editUser(ctx context.Context, u User) error {
	builder := utils.PSQL().Update("people.users").
		SetMap(map[string]interface{}{
			"password":            u.Password,
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
		return fmt.Errorf("failed to edit user: invalid rows affected: %d, this user may not exist: %d",
			rows, u.UserID)
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

	//nolint:musttag
	err = s.db.GetContext(ctx, &u, sql, args...)
	if err != nil {
		return u, fmt.Errorf("failed to get user from db: %w", err)
	}

	switch avatar := u.Avatar; {
	case u.UseGravatar:
		//nolint:gosec
		hash := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(u.Email))))
		u.Avatar = "https://www.gravatar.com/avatar/" + hex.EncodeToString(hash[:])
	case avatar == "":
		u.Avatar = "https://placehold.it/128x128"
	case strings.Contains(avatar, s.cdnEndpoint):
	case strings.Contains(avatar, fmt.Sprintf("%d.", u.UserID)):
		u.Avatar = "https://ystv.co.uk/static/images/members/thumb/" + avatar
	default:
		log.Printf("unknown avatar, user id: %d, length: %d, db string: %s, continuing", u.UserID, len(u.Avatar), u.Avatar)
		u.Avatar = ""
	}

	return u, nil
}

// getUsers will get users search with sorting with size and page, enabled and deleted
// Use the parameter direction for determining of the sorting will be ascending(asc) or descending(desc)
func (s *Store) getUsers(ctx context.Context, size, page int, search, sortBy, direction, enabled,
	deleted string) ([]User, int, error) {
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

	defer func() {
		_ = rows.Close()
	}()

	type tempStruct struct {
		User
		Count int `db:"full_count" json:"fullCount"`
	}

	for rows.Next() {
		var u1 User

		var temp tempStruct

		//nolint:musttag
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

func (s *Store) _getUsersBuilder(size, page int, search, sortBy, direction, enabled,
	deleted string) (*sq.SelectBuilder, error) {
	builder := utils.PSQL().Select("*", "count(*) OVER() AS full_count").
		From("people.users")

	if len(search) > 0 {
		builder = builder.Where(
			"(CAST(user_id AS TEXT) LIKE '%' || ? || '%' "+
				"OR LOWER(username) LIKE LOWER('%' || ? || '%') "+
				"OR LOWER(nickname) LIKE LOWER('%' || ? || '%') "+
				"OR LOWER(first_name) LIKE LOWER('%' || ? || '%') "+
				"OR LOWER(last_name) LIKE LOWER('%' || ? || '%') "+
				"OR LOWER(email) LIKE LOWER('%' || ? || '%') "+
				"OR LOWER(first_name || ' ' || last_name) LIKE LOWER('%' || ? || '%'))",
			search, search, search, search, search, search, search)
	}

	switch enabled {
	case "enabled":
		builder = builder.Where(sq.Eq{"enabled": true})
	case "disabled":
		builder = builder.Where(sq.Eq{"enabled": false})
	}

	switch deleted {
	case "not_deleted":
		builder = builder.Where(sq.Eq{"deleted_by": nil})
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
					"CASE WHEN ? = 'lastLogin' THEN last_login END ASC NULLS FIRST",
				sortBy, sortBy, sortBy, sortBy, sortBy, sortBy)
		case "desc":
			builder = builder.OrderByClause(
				"CASE WHEN ? = 'userId' THEN user_id END DESC, "+
					"CASE WHEN ? = 'name' THEN first_name END DESC, "+
					"CASE WHEN ? = 'name' THEN last_name END DESC, "+
					"CASE WHEN ? = 'username' THEN username END DESC, "+
					"CASE WHEN ? = 'email' THEN email END DESC, "+
					"CASE WHEN ? = 'lastLogin' THEN last_login END DESC NULLS LAST",
				sortBy, sortBy, sortBy, sortBy, sortBy, sortBy)
		default:
			return nil, fmt.Errorf(`invalid sorting direction, entered "%s" of length %d, but expected either 
"direction" or "desc"`, direction, len(direction))
		}
	}

	if page >= 1 && size >= 5 && size <= 100 {
		parsed1, err := strconv.ParseUint(strconv.Itoa(size), 10, 64)
		if err != nil {
			return nil, fmt.Errorf(`invalid value for size in direction "%s"`, direction)
		}
		parsed2, err := strconv.ParseUint(strconv.Itoa(size*(page-1)), 10, 64)
		if err != nil {
			return nil, fmt.Errorf(`invalid value for page in direction "%s"`, direction)
		}
		builder = builder.Limit(parsed1).Offset(parsed2)
	}

	return &builder, nil
}

// getPermissionsForUser returns all permissions for a user
func (s *Store) getPermissionsForUser(ctx context.Context, u User) ([]permission.Permission, error) {
	var p []permission.Permission

	builder := utils.PSQL().Select("p.*").
		From("people.permissions p").
		LeftJoin("people.role_permissions rp ON rp.permission_id = p.permission_id").
		LeftJoin("people.role_members rm ON rm.role_id = rp.role_id").
		Where(sq.Eq{"rm.user_id": u.UserID})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getPermissionsForUser: %w", err))
	}

	err = s.db.SelectContext(ctx, &p, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions for user: %w", err)
	}

	return p, nil
}

// getRolesForUser returns all roles for a user
func (s *Store) getRolesForUser(ctx context.Context, u User) ([]role.Role, error) {
	var r []role.Role

	builder := utils.PSQL().Select("r.*").
		From("people.roles r").
		LeftJoin("people.role_members rm ON rm.role_id = r.role_id").
		Where(sq.Eq{"rm.user_id": u.UserID})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getRolesForUser: %w", err))
	}

	err = s.db.SelectContext(ctx, &r, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles for user: %w", err)
	}

	return r, nil
}

// getUsersForRole returns all users for a role - moved here for cycle import reasons
func (s *Store) getUsersForRole(ctx context.Context, r role.Role) ([]User, error) {
	var u []User

	builder := utils.PSQL().Select("u.*").
		From("people.users u").
		LeftJoin("people.role_members rm ON rm.user_id = u.user_id").
		Where(sq.Eq{"rm.role_id": r.RoleID})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getUsersForRole: %w", err))
	}

	//nolint:musttag
	err = s.db.SelectContext(ctx, &u, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users for role: %w", err)
	}

	return u, nil
}

// getRoleUser returns a role user - moved here for cycle import reasons
func (s *Store) getRoleUser(ctx context.Context, ru1 RoleUser) (RoleUser, error) {
	var ru RoleUser

	builder := utils.PSQL().Select("*").
		From("people.role_members").
		Where(sq.And{
			sq.Eq{"role_id": ru1.RoleID},
			sq.Eq{"user_id": ru1.UserID},
		}).
		Limit(1)

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getRoleUser: %w", err))
	}

	err = s.db.GetContext(ctx, &ru, sql, args...)
	if err != nil {
		return RoleUser{}, fmt.Errorf("failed to get role user: %w", err)
	}

	return ru, nil
}

// getUsersNotInRole returns all the users not currently in the role.Role to be added
func (s *Store) getUsersNotInRole(ctx context.Context, r role.Role) ([]User, error) {
	var u []User

	subQuery := utils.PSQL().Select("u.user_id").
		From("people.users u").
		LeftJoin("people.role_members ru on u.user_id = ru.user_id").
		Where(sq.Eq{"ru.role_id": r.RoleID})

	builder := utils.PSQL().Select("u.*").
		Distinct().
		From("people.users u").
		Where(sq.And{
			utils.NotIn("user_id", subQuery),
			sq.Eq{"deleted_by": nil},
		}).
		OrderBy("first_name", "last_name")

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getRoles: %w", err))
	}

	//nolint:musttag
	err = s.db.SelectContext(ctx, &u, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	return u, nil
}

// addRoleUser creates a link between a role.Role and User
func (s *Store) addRoleUser(ctx context.Context, ru1 RoleUser) (RoleUser, error) {
	var ru RoleUser

	builder := utils.PSQL().Insert("people.role_members").
		Columns("role_id", "user_id").
		Values(ru1.RoleID, ru1.UserID).
		Suffix("RETURNING role_id, user_id")

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addRoleUser: %w", err))
	}

	stmt, err := s.db.PrepareContext(ctx, sql)
	if err != nil {
		return RoleUser{}, fmt.Errorf("failed to add role user: %w", err)
	}

	defer stmt.Close()

	err = stmt.QueryRow(args...).Scan(&ru.RoleID, &ru.UserID)
	if err != nil {
		return RoleUser{}, fmt.Errorf("failed to add role user: %w", err)
	}

	return ru, nil
}

// removeRoleUser removes a link between a role.Role and User
func (s *Store) removeRoleUser(ctx context.Context, ru RoleUser) error {
	builder := utils.PSQL().Delete("people.role_members").
		Where(sq.And{
			sq.Eq{"role_id": ru.RoleID},
			sq.Eq{"user_id": ru.UserID},
		})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for removeRoleUser: %w", err))
	}

	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to remove role user: %w", err)
	}

	return nil
}

// removeUserForRoles removes all links between role.Role and a User
func (s *Store) removeUserForRoles(ctx context.Context, u User) error {
	builder := utils.PSQL().Delete("people.role_members").
		Where(sq.Eq{"user_id": u.UserID})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for removeUserForRoles: %w", err))
	}

	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to remove user for roles: %w", err)
	}

	return nil
}

// getPermissionsForRole returns all permissions for a role - moved here for cycle import reasons
func (s *Store) getPermissionsForRole(ctx context.Context, r role.Role) ([]permission.Permission, error) {
	var p []permission.Permission

	builder := utils.PSQL().Select("p.*").
		From("people.permissions p").
		LeftJoin("people.role_permissions rp ON p.permission_id = rp.permission_id").
		Where(sq.Eq{"rp.role_id": r.RoleID})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getPermissionsForRole: %w", err))
	}

	err = s.db.SelectContext(ctx, &p, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions for role: %w", err)
	}

	return p, nil
}

// getRolesForPermission returns all roles for a permission - moved here for cycle import reasons
func (s *Store) getRolesForPermission(ctx context.Context, p permission.Permission) ([]role.Role, error) {
	var r []role.Role

	builder := utils.PSQL().Select("r.*").
		From("people.roles r").
		LeftJoin("people.role_permissions rp ON r.role_id = rp.role_id").
		Where(sq.Eq{"rp.permission_id": p.PermissionID})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getRolesForPermission: %w", err))
	}

	err = s.db.SelectContext(ctx, &r, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles for permission: %w", err)
	}

	return r, nil
}

// getRolePermission returns a role permission - moved here for cycle import reasons
func (s *Store) getRolePermission(ctx context.Context, rp1 RolePermission) (RolePermission, error) {
	var rp RolePermission

	builder := utils.PSQL().Select("*").
		From("people.role_permissions").
		Where(sq.And{
			sq.Eq{"role_id": rp1.RoleID},
			sq.Eq{"permission_id": rp1.PermissionID},
		}).
		Limit(1)

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getRolePermission: %w", err))
	}

	err = s.db.GetContext(ctx, &rp, sql, args...)
	if err != nil {
		return RolePermission{}, fmt.Errorf("failed to get role permission: %w", err)
	}

	return rp, nil
}

// getPermissionsNotInRole returns all the permissions not currently in the role.Role to be added
func (s *Store) getPermissionsNotInRole(ctx context.Context, r role.Role) ([]permission.Permission, error) {
	var p []permission.Permission

	subQuery := utils.PSQL().Select("p.permission_id").
		From("people.permissions p").
		LeftJoin("people.role_permissions rp on p.permission_id = rp.permission_id").
		Where(sq.Eq{"rp.role_id": r.RoleID})

	builder := utils.PSQL().Select("p.*").
		Distinct().
		From("people.permissions p").
		Where(utils.NotIn("permission_id", subQuery)).
		OrderBy("name")

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getPermissionsNotInRole: %w", err))
	}

	//nolint:asasalint
	err = s.db.SelectContext(ctx, &p, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions not in role: %w", err)
	}

	return p, nil
}

// addRolePermission creates a link between a role.Role and permission.Permission
func (s *Store) addRolePermission(ctx context.Context, rp1 RolePermission) (RolePermission, error) {
	var rp RolePermission

	builder := utils.PSQL().Insert("people.role_permissions").
		Columns("role_id ", "permission_id").
		Values(rp1.RoleID, rp1.PermissionID).
		Suffix("RETURNING role_id, permission_id")

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addRolePermission: %w", err))
	}

	stmt, err := s.db.PrepareContext(ctx, sql)
	if err != nil {
		return RolePermission{}, fmt.Errorf("failed to add rolePermission: %w", err)
	}

	defer stmt.Close()

	err = stmt.QueryRow(args...).Scan(&rp.RoleID, &rp.PermissionID)
	if err != nil {
		return RolePermission{}, fmt.Errorf("failed to add rolePermission: %w", err)
	}

	return rp, nil
}

// removeRolePermission removes a link between a role.Role and permission.Permission
func (s *Store) removeRolePermission(ctx context.Context, rp RolePermission) error {
	builder := utils.PSQL().Delete("people.role_permissions").
		Where(sq.And{sq.Eq{"role_id": rp.RoleID}, sq.Eq{"permission_id": rp.PermissionID}})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for removeRolePermission: %w", err))
	}

	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete rolePermission: %w", err)
	}

	return nil
}

func (s *Store) getCrowdApp(ctx context.Context, c1 CrowdApp) (CrowdApp, error) {
	var c CrowdApp

	builder := utils.PSQL().Select("*").
		From("web_auth.crowd_apps").
		Where(sq.Or{
			sq.Eq{"app_id": c1.AppID},
			sq.Eq{"name": c1.Name}}).
		Limit(1)

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getCrowdApp: %w", err))
	}

	//nolint:musttag
	err = s.db.GetContext(ctx, &c, sql, args...)
	if err != nil {
		return c, fmt.Errorf("failed to get crowd app from db: %w", err)
	}

	return c, nil
}

func (s *Store) getCrowdApps(ctx context.Context, crowdAppStatus CrowdAppStatus) ([]CrowdApp, error) {
	var c []CrowdApp

	builder := utils.PSQL().Select("app_id", "name", "description", "active").
		From("web_auth.crowd_apps")

	switch crowdAppStatus {
	case Any:
	case Active:
		builder = builder.Where("active = true")
	case Inactive:
		builder = builder.Where("active = false")
	}

	builder = builder.OrderBy("name")

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getCrowdApps: %w", err))
	}

	err = s.db.SelectContext(ctx, &c, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get crowd apps: %w", err)
	}

	return c, nil
}
