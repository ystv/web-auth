package role

import (
	"context"
	"fmt"
)

// getRoles returns all roles for a user
func (s *Store) getRoles(ctx context.Context) (r []Role, err error) {
	err = s.db.SelectContext(ctx, &r, `SELECT r.role_id, r.name, r.description
		FROM people.roles r;`)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}
	return r, nil
}
