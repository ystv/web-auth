package role

import (
	"context"
	"fmt"
)

// getRoles returns all roles for a user
func (s *Store) getRoles(ctx context.Context) (r []Role, err error) {
	err = s.db.SelectContext(ctx, &r, `SELECT r.role_id, r.name, r.description, COUNT(rm.user_id) AS users
		FROM people.roles r
		INNER JOIN people.role_members rm on r.role_id = rm.role_id
		GROUP BY r.role_id, r.name, r.description
		ORDER BY name;`)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}
	return r, nil
}

//func (s *Store) getUsersCountForRole(ctx context.Context, r Role) (i int, err error) {
//	err = s.db.GetContext(ctx, &i, `SELECT COUNT(*)
//		FROM people.role_members
//		WHERE role_id = $1`, r.RoleID)
//	if err != nil {
//		return -1, fmt.Errorf("failed to get role users count: %w", err)
//	}
//	return i, nil
//}

//func (s *Store) GetUsersForRole(ctx context.Context, r Role) (i User, err error) {
//	err = s.db.GetContext(ctx, &i, `SELECT COUNT(*)
//		FROM people.role_members
//		WHERE role_id = $1`, r.RoleID)
//	if err != nil {
//		return -1, fmt.Errorf("failed to get role users count: %w", err)
//	}
//	return i, nil
//}
