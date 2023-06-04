package permission

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/ystv/web-auth/role"
)

type (
	// Repo where all user data is stored
	Repo interface {
		GetPermissions(ctx context.Context) ([]Permission, error)
		GetPermissionsForRole(ctx context.Context, r role.Role) ([]Permission, error)
		GetPermission(ctx context.Context, p Permission) (Permission, error)
		AddPermission(ctx context.Context, p Permission) (Permission, error)
	}
	// Store stores the dependencies
	Store struct {
		db *sqlx.DB
	}
	// Permission represents relevant permission fields
	Permission struct {
		PermissionID int    `db:"permission_id" json:"id"`
		Name         string `db:"name" json:"name"`
		Description  string `db:"description" json:"description"`
		//ParentID     int    `db:"parent_id" json:"parentID"`
	}
)

// NewPermissionRepo stores our dependency
func NewPermissionRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

// GetPermissions returns all permissions of a user
func (s *Store) GetPermissions(ctx context.Context) ([]Permission, error) {
	return s.getPermissions(ctx)
}

// GetPermission returns all permissions of a user
func (s *Store) GetPermission(ctx context.Context, id int) (Permission, error) {
	return s.getPermission(ctx, id)
}
