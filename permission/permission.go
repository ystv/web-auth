package permission

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type (
	// Store stores the dependencies
	Store struct {
		db *sqlx.DB
	}

	// Permission represents relevant permission fields
	Permission struct {
		PermissionID int    `db:"permission_id" json:"id"`
		Name         string `db:"name" json:"name"`
		Description  string `db:"description" json:"description"`
		Roles        int    `db:"roles" json:"roles"`
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
func (s *Store) GetPermission(ctx context.Context, p Permission) (Permission, error) {
	return s.getPermission(ctx, p)
}

// AddPermission returns all permissions of a user
func (s *Store) AddPermission(ctx context.Context, p Permission) (Permission, error) {
	return s.addPermission(ctx, p)
}

// EditPermission returns all permissions of a user
func (s *Store) EditPermission(ctx context.Context, p Permission) (Permission, error) {
	return s.editPermission(ctx, p)
}

// DeletePermission deletes a permission
func (s *Store) DeletePermission(ctx context.Context, p Permission) error {
	return s.deletePermission(ctx, p)
}

// DeleteRolePermission deletes a rolePermission
func (s *Store) DeleteRolePermission(ctx context.Context, p Permission) error {
	return s.deleteRolePermission(ctx, p)
}
