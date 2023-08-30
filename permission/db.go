package permission

import (
	"context"
	//nolint:goimports
	"fmt"
	sq "github.com/Masterminds/squirrel"
)

// getPermissions returns all permissions
func (s *Store) getPermissions(ctx context.Context) ([]Permission, error) {
	var p []Permission
	builder := sq.Select("p.*", "COUNT(rp.role_id) AS roles").
		From("people.permissions p").
		LeftJoin("people.role_permissions rp on p.permission_id = rp.permission_id").
		GroupBy("p", "p.permission_id", "name", "description").
		OrderBy("p.name")
	sql, _, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getPermissions: %w", err))
	}
	err = s.db.SelectContext(ctx, &p, sql)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}
	return p, nil
}

// getPermissions returns all permissions
func (s *Store) getPermission(ctx context.Context, p1 Permission) (Permission, error) {
	var p Permission
	builder := sq.Select("p.*", "COUNT(rp.role_id) AS roles").
		From("people.permissions p").
		LeftJoin("people.role_permissions rp on p.permission_id = rp.permission_id").
		Where(sq.Eq{"p.permission_id": p1.PermissionID}).
		GroupBy("p", "p.permission_id", "name", "description").
		Limit(1).
		PlaceholderFormat(sq.Dollar)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getPermission: %w", err))
	}
	err = s.db.GetContext(ctx, &p, sql, args...)
	if err != nil {
		return Permission{}, fmt.Errorf("failed to get permission: %w", err)
	}
	return p, nil
}

func (s *Store) addPermission(ctx context.Context, p Permission) (Permission, error) {
	builder := sq.Insert("people.permissions").
		Columns("name", "description").
		Values(p.Name, p.Description).
		Suffix("RETURNING permission_id").
		PlaceholderFormat(sq.Dollar)
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addPermission: %w", err))
	}
	stmt, err := s.db.PrepareContext(ctx, sql)
	if err != nil {
		return Permission{}, fmt.Errorf("failed to add permission: %w", err)
	}
	defer stmt.Close()
	err = stmt.QueryRow(args...).Scan(&p.PermissionID)
	if err != nil {
		return Permission{}, fmt.Errorf("failed to add permission: %w", err)
	}
	return p, nil
}

func (s *Store) editPermission(ctx context.Context, p1 Permission) (Permission, error) {
	_ = ctx
	_ = p1
	panic("editPermission not implemented")
}

func (s *Store) deletePermission(ctx context.Context, p Permission) error {
	builder := sq.Delete("people.permissions").
		Where(sq.Eq{"permission_id": p.PermissionID})
	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deletePermission: %w", err))
	}
	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}
	return nil
}

func (s *Store) deleteRolePermission(ctx context.Context, p Permission) error {
	builder := sq.Delete("people.role_permissions").
		Where(sq.Eq{"permission_id": p.PermissionID})
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
