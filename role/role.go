package role

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type (
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

// RemovePermissionsForRole deletes a rolePermission
func (s *Store) RemovePermissionsForRole(ctx context.Context, r Role) error {
	return s.removePermissionsForRole(ctx, r)
}

// RemoveUsersForRole deletes a roleUser
func (s *Store) RemoveUsersForRole(ctx context.Context, r Role) error {
	return s.removeUsersForRole(ctx, r)
}
