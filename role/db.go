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
		WHERE r.role_id = $1
		GROUP BY r.role_id, r.name, r.description
		LIMIT 1;`, r1.RoleID)
	if err != nil {
		return Role{}, fmt.Errorf("failed to get role: %w", err)
	}
	return r, nil
}

func (s *Store) addRole(ctx context.Context, r1 Role) (Role, error) {
	panic("addRole not implemented")
}

func (s *Store) editRole(ctx context.Context, r1 Role) (Role, error) {
	panic("editRole not implemented")
}

func (s *Store) deleteRole(ctx context.Context, r1 Role) error {
	panic("deleteRole not implemented")
}
