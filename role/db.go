package role

import (
	"context"
	"fmt"
)

// getRoles returns all roles for a user
func (s *Store) getRoles(ctx context.Context) ([]Role, error) {
	var r []Role
	err := s.db.SelectContext(ctx, &r, `SELECT r.role_id, r.name, r.description, COUNT(DISTINCT rm.user_id) AS users, COUNT(DISTINCT rp.permission_id) AS permissions
		FROM people.roles r
		LEFT JOIN people.role_members rm ON r.role_id = rm.role_id
		LEFT JOIN people.role_permissions rp ON r.role_id = rp.role_id
		GROUP BY r.role_id, r.name, r.description
		ORDER BY r.name;`)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}
	return r, nil
}

func (s *Store) getRole(ctx context.Context, r1 Role) (Role, error) {
	var r Role
	err := s.db.GetContext(ctx, &r, `SELECT r.role_id, r.name, r.description, COUNT(DISTINCT rm.user_id) AS users, COUNT(DISTINCT rp.permission_id) AS permissions
		FROM people.roles r
		LEFT JOIN people.role_members rm ON r.role_id = rm.role_id
		LEFT JOIN people.role_permissions rp ON r.role_id = rp.role_id
		WHERE r.role_id = $1 OR (r.name = $2 AND r.name != '')
		GROUP BY r.role_id, r.name, r.description
		LIMIT 1;`, r1.RoleID, r1.Name)
	if err != nil {
		return Role{}, fmt.Errorf("failed to get role: %w", err)
	}
	return r, nil
}

func (s *Store) addRole(ctx context.Context, r1 Role) (Role, error) {
	var r Role
	stmt, err := s.db.PrepareNamedContext(ctx, "INSERT INTO people.roles (name, description) VALUES (:name, :description) RETURNING role_id, name, description")
	if err != nil {
		return Role{}, fmt.Errorf("failed to add role: %w", err)
	}
	err = stmt.Get(&r, r1)
	if err != nil {
		return Role{}, fmt.Errorf("failed to add role: %w", err)
	}
	return r, nil
}

func (s *Store) editRole(ctx context.Context, r Role) (Role, error) {
	stmt, err := s.db.NamedExecContext(ctx, `UPDATE people.roles
		SET name = :name,
			description = :description
		WHERE role_id = :role_id`, r)
	if err != nil {
		return Role{}, fmt.Errorf("failed to update role: %w", err)
	}
	rows, err := stmt.RowsAffected()
	if err != nil {
		return Role{}, fmt.Errorf("failed to update role: %w", err)
	}
	if rows < 1 {
		return Role{}, fmt.Errorf("failed to update role: invalid rows affected: %d", rows)
	}
	return r, nil
}

func (s *Store) deleteRole(ctx context.Context, r Role) error {
	_, err := s.db.NamedExecContext(ctx, `DELETE FROM people.roles WHERE role_id = :role_id`, r)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

func (s *Store) deleteRolePermission(ctx context.Context, r Role) error {
	_, err := s.db.NamedExecContext(ctx, `DELETE FROM people.role_permissions WHERE role_id = :role_id`, r)
	if err != nil {
		return fmt.Errorf("failed to delete rolePermission: %w", err)
	}
	return nil
}

func (s *Store) deleteRoleUser(ctx context.Context, r Role) error {
	_, err := s.db.NamedExecContext(ctx, `DELETE FROM people.role_members WHERE role_id = :role_id`, r)
	if err != nil {
		return fmt.Errorf("failed to delete roleUser: %w", err)
	}
	return nil
}
