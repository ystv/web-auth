package permission

import (
	"context"
	"fmt"
)

// getPermissions returns all permissions
func (s *Store) getPermissions(ctx context.Context) ([]Permission, error) {
	var p []Permission
	err := s.db.SelectContext(ctx, &p, `SELECT p.*, COUNT(rp.role_id) AS roles
		FROM people.permissions p
		LEFT JOIN people.role_permissions rp on p.permission_id = rp.permission_id
		GROUP BY p, p.permission_id, name, description
		ORDER BY p.name;`)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}
	return p, nil
}

// getPermissions returns all permissions
func (s *Store) getPermission(ctx context.Context, p1 Permission) (Permission, error) {
	var p Permission
	err := s.db.GetContext(ctx, &p, `SELECT p.*, COUNT(rp.role_id) AS roles
		FROM people.permissions p
		LEFT JOIN people.role_permissions rp on p.permission_id = rp.permission_id
		WHERE p.permission_id = $1
		GROUP BY p, p.permission_id, name, description
		LIMIT 1;`, p1.PermissionID)
	if err != nil {
		return Permission{}, fmt.Errorf("failed to get permission: %w", err)
	}
	return p, nil
}

func (s *Store) addPermission(ctx context.Context, p1 Permission) (Permission, error) {
	var p Permission
	stmt, err := s.db.PrepareNamedContext(ctx, "INSERT INTO people.permissions (name, description) VALUES (:name, :description) RETURNING permission_id, name, description")
	if err != nil {
		return Permission{}, fmt.Errorf("failed to add permission: %w", err)
	}
	err = stmt.Get(&p, p1)
	if err != nil {
		return Permission{}, fmt.Errorf("failed to add permission: %w", err)
	}
	return p, nil
}

func (s *Store) editPermission(ctx context.Context, p1 Permission) (Permission, error) {
	return Permission{}, nil
}

func (s *Store) deletePermission(ctx context.Context, p1 Permission) error {
	_, err := s.db.NamedExecContext(ctx, `DELETE FROM people.permissions WHERE permission_id = :permission_id`, p1)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}
	return nil
}

func (s *Store) deleteRolePermission(ctx context.Context, p1 Permission) error {
	_, err := s.db.NamedExecContext(ctx, `DELETE FROM people.role_permissions WHERE permission_id = :permission_id`, p1)
	if err != nil {
		return fmt.Errorf("failed to delete rolePermission: %w", err)
	}
	return nil
}
