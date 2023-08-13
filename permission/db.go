package permission

import (
	"context"
	"fmt"
)

// getPermissions returns all permissions
func (s *Store) getPermissions(ctx context.Context) (p []Permission, err error) {
	err = s.db.SelectContext(ctx, &p, `SELECT *
		FROM people.permissions
		ORDER BY name;`)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}
	return p, nil
}

// getPermissions returns all permissions
func (s *Store) getPermission(ctx context.Context, id int) (Permission, error) {
	p := Permission{}
	err := s.db.GetContext(ctx, &p, `SELECT *
		FROM people.permissions
		WHERE permission_id = $1
		LIMIT 1;`, id)
	if err != nil {
		return Permission{}, fmt.Errorf("failed to get permission: %w", err)
	}
	return p, nil
}
