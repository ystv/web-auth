package role

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type (
	// Repo where all user data is stored
	Repo interface {
		GetRoles(ctx context.Context) ([]Role, error)
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
	}

	RolePermission struct {
		RoleID       int `db:"role_id" json:"role_id"`
		PermissionID int `db:"permission_id" json:"permission_id"`
	}

	RoleUser struct {
		RoleID int `db:"role_id" json:"role_id"`
		UserID int `db:"user_id" json:"user_id"`
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
