package permission

import (
	"context"
	"fmt"
)

// getPermissions returns all permissions
func (s *Store) getPermissions(ctx context.Context) (p []Permission, err error) {
	err = s.db.SelectContext(ctx, &p, `SELECT *
		FROM people.permissions;`)
	if err != nil {
		return nil, fmt.Errorf("failed to get permissions: %w", err)
	}
	return p, nil
}
