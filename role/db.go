package role

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

// getRoles returns all roles for a user
func (s *Store) getRoles(ctx context.Context) ([]Role, error) {
	var r []Role
	builder := sq.Select("r.*", "COUNT(DISTINCT rm.user_id) AS users", "COUNT(DISTINCT rp.permission_id) AS permissions").
		From("people.roles r").
		LeftJoin("people.role_members rm ON r.role_id = rm.role_id").
		LeftJoin("people.role_permissions rp ON r.role_id = rp.role_id").
		GroupBy("r", "r.role_id", "name", "description").
		OrderBy("r.name")
	sql, _, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getRoles: %w", err))
	}
	err = s.db.SelectContext(ctx, &r, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}
	return r, nil
}

func (s *Store) getRole(ctx context.Context, r1 Role) (Role, error) {
	var r Role
	builder := sq.Select("r.*", "COUNT(DISTINCT rm.user_id) AS users", "COUNT(DISTINCT rp.permission_id) AS permissions").
		From("people.roles r").
		LeftJoin("people.role_members rm ON r.role_id = rm.role_id").
		LeftJoin("people.role_permissions rp ON r.role_id = rp.role_id").
		Where(sq.Or{sq.Eq{"r.role_id": r1.RoleID}, sq.And{sq.Eq{"r.name": r1.Name}, sq.NotEq{"r.name": ""}}}).
		GroupBy("r", "r.role_id", "name", "description").
		Limit(1).
		PlaceholderFormat(sq.Dollar)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getRole: %w", err))
	}
	err = s.db.GetContext(ctx, &r, sql, args...)
	if err != nil {
		return Role{}, fmt.Errorf("failed to get role: %w", err)
	}
	return r, nil
}

func (s *Store) addRole(ctx context.Context, r Role) (Role, error) {
	builder := sq.Insert("people.roles").
		Columns("name", "description").
		Values(r.Name, r.Description).
		Suffix("RETURNING role_id").
		PlaceholderFormat(sq.Dollar)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addRole: %w", err))
	}
	stmt, err := s.db.PrepareContext(ctx, sql)
	if err != nil {
		return Role{}, fmt.Errorf("failed to add role: %w", err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(args...).Scan(&r.RoleID)
	if err != nil {
		return Role{}, fmt.Errorf("failed to add role: %w", err)
	}
	return r, nil
}

func (s *Store) editRole(ctx context.Context, r1 Role) (Role, error) {
	_ = ctx
	_ = r1
	panic("editRole not implemented")
}

func (s *Store) deleteRole(ctx context.Context, r Role) error {
	builder := sq.Delete("people.roles").
		Where(sq.Eq{"role_id": r.RoleID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteRole: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

func (s *Store) deleteRolePermission(ctx context.Context, r Role) error {
	builder := sq.Delete("people.role_permissions").
		Where(sq.Eq{"role_id": r.RoleID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteRolePermission: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete rolePermission: %w", err)
	}
	return nil
}
