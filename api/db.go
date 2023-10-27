package api

import (
	"context"
	"fmt"
	"time"

	"gopkg.in/guregu/null.v4"
)

// getTokens will get the tokens for a user
func (s *Store) getTokens(ctx context.Context, userID int) ([]Token, error) {
	var t []Token
	err := s.db.SelectContext(ctx, &t, `SELECT *
		FROM web_auth.api_tokens
		WHERE user_id = $1;`, userID)
	if err != nil {
		return nil, fmt.Errorf("error at getTokens: %w", err)
	}
	return t, nil
}

// getToken will get a specific token
func (s *Store) getToken(ctx context.Context, t1 Token) (Token, error) {
	var t Token
	err := s.db.GetContext(ctx, &t, `SELECT *
		FROM web_auth.api_tokens
		WHERE token_id = $1
		LIMIT 1;`, t1.TokenID)
	if err != nil {
		return t, fmt.Errorf("error at getToken: %w", err)
	}
	return t, nil
}

// addToken will add a token specified in the parameter
func (s *Store) addToken(ctx context.Context, t1 Token) (Token, error) {
	var t Token
	stmt, err := s.db.PrepareNamedContext(ctx, "INSERT INTO web_auth.api_tokens (token_id, name, description, expiry, user_id) VALUES (:token_id, :name, :description, :expiry, :user_id) RETURNING token_id, name, description, expiry, user_id")
	if err != nil {
		return Token{}, fmt.Errorf("failed to add token: %w", err)
	}
	defer stmt.Close()
	err = stmt.Get(&t, t1)
	if err != nil {
		return Token{}, fmt.Errorf("failed to add token: %w", err)
	}
	return t, nil
}

// deleteToken will delete a token and prevent it from being used
func (s *Store) deleteToken(ctx context.Context, t Token) error {
	_, err := s.db.NamedExecContext(ctx, `DELETE FROM web_auth.api_tokens WHERE token_id = :token_id`, t)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}
	return nil
}

// deleteOldToken will delete all expired tokens
func (s *Store) deleteOldToken(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM web_auth.api_tokens WHERE expiry <= $1`, null.TimeFrom(time.Now()))
	if err != nil {
		return fmt.Errorf("failed to delete old token: %w", err)
	}
	return nil
}
