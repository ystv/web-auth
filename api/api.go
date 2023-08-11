package api

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type (
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

	Token struct {
		TokenID     string `db:"token_id" json:"tokenID"`
		Name        string `db:"name" json:"name,omitempty"`
		Description string `db:"description" json:"description,omitempty"`
		Expiry      int64  `db:"expiry" json:"expiry"`
		UserID      int    `db:"user_id" json:"userID"`
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

func (s *Store) GetTokens(ctx context.Context, userID int) ([]Token, error) {
	return s.getTokens(ctx, userID)
}

func (s *Store) GetToken(ctx context.Context, t Token) (Token, error) {
	return s.getToken(ctx, t)
}

func (s *Store) AddToken(ctx context.Context, t Token) (Token, error) {
	return s.addToken(ctx, t)
}

func (s *Store) DeleteToken(ctx context.Context, t Token) error {
	return s.deleteToken(ctx, t)
}

func (s *Store) DeleteOldToken(ctx context.Context) error {
	return s.deleteOldToken(ctx)
}
