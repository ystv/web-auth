package db

import (
	"context"

	"github.com/ystv/web-auth/types"
)

// GetPermissions returns all permissions for a user
func (store *DB) GetPermissions(ctx context.Context, u *types.User) error {
	return store.SelectContext(ctx, &u.Permissions, `SELECT p.permission_id, p.name
	FROM people.permissions p
	INNER JOIN people.role_permissions rp ON rp.permission_id = p.permission_id
	INNER JOIN people.role_members rm ON rm.role_id = rp.role_id
	WHERE rm.user_id = $1;`, u.UserID)
}
