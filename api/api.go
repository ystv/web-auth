package api

import (
	"context"
	"github.com/jmoiron/sqlx"
	"gopkg.in/guregu/null.v4"
)

type (
	// Repo is used for navigating a package
	Repo interface {
		GetTokens(ctx context.Context, userID int) ([]Token, error)
		GetToken(ctx context.Context, t Token) (Token, error)
		AddToken(ctx context.Context, t Token) (Token, error)
		DeleteToken(ctx context.Context, t Token) error
		DeleteOldToken(ctx context.Context) error

		getTokens(ctx context.Context, userID int) ([]Token, error)
		getToken(ctx context.Context, t Token) (Token, error)
		addToken(ctx context.Context, t Token) (Token, error)
		deleteToken(ctx context.Context, t Token) error
		deleteOldToken(ctx context.Context) error
	}

	// Store stores the dependencies
	Store struct {
		db *sqlx.DB
	}

	// Token is the struct for a token to be stored
	Token struct {
		TokenID     string    `db:"token_id" json:"tokenID"`
		Name        string    `db:"name" json:"name,omitempty"`
		Description string    `db:"description" json:"description,omitempty"`
		Expiry      null.Time `db:"expiry" json:"expiry"`
		UserID      int       `db:"user_id" json:"userID"`
	}
)

// here to verify we are meeting the interface
var _ Repo = &Store{}

// NewAPIRepo stores the dependency
func NewAPIRepo(db *sqlx.DB) *Store {
	return &Store{
		db: db,
	}
}

// GetTokens returns all the tokens that a user has
func (s *Store) GetTokens(ctx context.Context, userID int) ([]Token, error) {
	return s.getTokens(ctx, userID)
}

// GetToken returns a specific token
func (s *Store) GetToken(ctx context.Context, t Token) (Token, error) {
	return s.getToken(ctx, t)
}

// AddToken adds a token id, name, description, user id and expiration, the actual jwt token is not stored
func (s *Store) AddToken(ctx context.Context, t Token) (Token, error) {
	return s.addToken(ctx, t)
}

// DeleteToken deletes a specific token
func (s *Store) DeleteToken(ctx context.Context, t Token) error {
	return s.deleteToken(ctx, t)
}

// DeleteOldToken deletes all old tokens by the subroutine
func (s *Store) DeleteOldToken(ctx context.Context) error {
	return s.deleteOldToken(ctx)
}
