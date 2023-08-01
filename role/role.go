package role

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type (
	// Repo where all user data is stored
	Repo interface {
		GetRoles(ctx context.Context) ([]Role, error)
		GetRole(ctx context.Context, r Role) (Role, error)
		AddRole(ctx context.Context, r Role) (Role, error)
		EditRole(ctx context.Context, r Role) (Role, error)
		DeleteRole(ctx context.Context, r Role) error

		getRoles(ctx context.Context) ([]Role, error)
		getRole(ctx context.Context, r1 Role) (Role, error)
		addRole(ctx context.Context, r1 Role) (Role, error)
		editRole(ctx context.Context, r1 Role) (Role, error)
		deleteRole(ctx context.Context, r1 Role) error
	}

	// Store stores the dependencies
	Store struct {
		db *sqlx.DB
	}

	// Role represents relevant user fields
	Role struct {
		RoleID      int    `db:"role_id" json:"id"`
		Name        string `db:"name" json:"name" schema:"name"`
		Description string `db:"description" json:"description" schema:"description"`
		Users       int    `db:"users" json:"users"`
		Permissions int    `db:"permissions" json:"permissions"`
	}

	//RolePermission struct {
	//	RoleID       int `db:"role_id" json:"role_id"`
	//	PermissionID int `db:"permission_id" json:"permission_id"`
	//}
	//
	//RoleUser struct {
	//	RoleID int `db:"role_id" json:"role_id"`
	//	UserID int `db:"user_id" json:"user_id"`
	//}
)

var _ Repo = &Store{}

// NewRoleRepo stores our dependency
func NewRoleRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

// GetRoles returns all roles
func (s *Store) GetRoles(ctx context.Context) ([]Role, error) {
	return s.getRoles(ctx)
}

// GetRole returns a role
func (s *Store) GetRole(ctx context.Context, r Role) (Role, error) {
	return s.getRole(ctx, r)
}

// AddRole adds a role
func (s *Store) AddRole(ctx context.Context, r Role) (Role, error) {
	return s.addRole(ctx, r)
}

// EditRole edits a role
func (s *Store) EditRole(ctx context.Context, r Role) (Role, error) {
	return s.editRole(ctx, r)
}

// DeleteRole deletes a role
func (s *Store) DeleteRole(ctx context.Context, r Role) error {
	return s.deleteRole(ctx, r)
}

// DeleteRolePermission deletes a rolePermission
func (s *Store) DeleteRolePermission(ctx context.Context, r Role) error {
	return s.deleteRolePermission(ctx, r)
}
