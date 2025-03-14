package permission

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"

	"github.com/ystv/web-auth/utils"
)

// getPermissions returns all permissions
func (s *Store) getPermissions(ctx context.Context) ([]Permission, error) {
	var p []Permission

	builder := utils.PSQL().Select("p.*", "COUNT(rp.role_id) AS roles").
		From("people.permissions p").
		LeftJoin("people.role_permissions rp on p.permission_id = rp.permission_id").
		GroupBy("p", "p.permission_id", "name", "description").
		OrderBy("p.name")

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getPermissions: %w", err))
	}

	err = s.db.SelectContext(ctx, &p, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}

	return p, nil
}

// getPermission returns a permission
func (s *Store) getPermission(ctx context.Context, p1 Permission) (Permission, error) {
	var p Permission

	builder := utils.PSQL().Select("p.*", "COUNT(rp.role_id) AS roles").
		From("people.permissions p").
		LeftJoin("people.role_permissions rp on p.permission_id = rp.permission_id").
		Where(sq.Eq{"p.permission_id": p1.PermissionID}).
		GroupBy("p", "p.permission_id", "name", "description").
		Limit(1)

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

// addPermission adds a new permission
func (s *Store) addPermission(ctx context.Context, p Permission) (Permission, error) {
	builder := utils.PSQL().Insert("people.permissions").
		Columns("name", "description").
		Values(p.Name, p.Description).
		Suffix("RETURNING permission_id")

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

// editPermission edits an existing permission
func (s *Store) editPermission(ctx context.Context, p Permission) (Permission, error) {
	builder := utils.PSQL().Update("people.permissions").
		SetMap(map[string]interface{}{
			"name":        p.Name,
			"description": p.Description,
		}).
		Where(sq.Eq{"permission_id": p.PermissionID})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for editPermission: %w", err))
	}

	res, err := s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return Permission{}, fmt.Errorf("failed to edit permission: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return Permission{}, fmt.Errorf("failed to edit permission: %w", err)
	}

	if rows < 1 {
		return Permission{}, fmt.Errorf("failed to edit permissions: invalid rows affected: %d", rows)
	}

	return p, nil
}

// deletePermission deletes a specific permission
func (s *Store) deletePermission(ctx context.Context, p Permission) error {
	builder := utils.PSQL().Delete("people.permissions").
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

// removePermissionForRoles deletes the connection between multiple role.Role and a Permission
func (s *Store) removePermissionForRoles(ctx context.Context, p Permission) error {
	builder := utils.PSQL().Delete("people.role_permissions").
		Where(sq.Eq{"permission_id": p.PermissionID})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for removePermissionForRoles: %w", err))
	}

	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete rolePermission: %w", err)
	}

	return nil
}
