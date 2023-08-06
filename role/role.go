package role

import (
	"context"
	"fmt"
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
		DeleteRolePermission(ctx context.Context, r Role) error
		DeleteRoleUser(ctx context.Context, r Role) error

		getRoles(ctx context.Context) ([]Role, error)
		getRole(ctx context.Context, r1 Role) (Role, error)
		addRole(ctx context.Context, r1 Role) (Role, error)
		editRole(ctx context.Context, r1 Role) (Role, error)
		deleteRole(ctx context.Context, r1 Role) error
		deleteRolePermission(ctx context.Context, r Role) error
		deleteRoleUser(ctx context.Context, r Role) error
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
	role, err := s.GetRole(ctx, r)
	if err != nil {
		return r, fmt.Errorf("failed to get role: %w", err)
	}
	if r.Name != role.Name && len(r.Name) > 0 {
		role.Name = r.Name
	}
	if r.Description != role.Description && len(r.Description) > 0 {
		role.Description = r.Description
	}
	return s.editRole(ctx, role)
}

// DeleteRole deletes a role
func (s *Store) DeleteRole(ctx context.Context, r Role) error {
	return s.deleteRole(ctx, r)
}

// DeleteRolePermission deletes a rolePermission
func (s *Store) DeleteRolePermission(ctx context.Context, r Role) error {
	return s.deleteRolePermission(ctx, r)
}

// DeleteRoleUser deletes a roleUser
func (s *Store) DeleteRoleUser(ctx context.Context, r Role) error {
	return s.deleteRoleUser(ctx, r)
}
