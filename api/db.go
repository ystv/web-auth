package api

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"gopkg.in/guregu/null.v4"

	"github.com/ystv/web-auth/utils"
)

// getTokens will get the tokens for a user
func (s *Store) getTokens(ctx context.Context, userID int) ([]Token, error) {
	var t []Token

	builder := utils.PSQL().Select("*").
		From("web_auth.api_tokens").
		Where(sq.Eq{"user_id": userID})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getTokens: %w", err))
	}

	err = s.db.SelectContext(ctx, &t, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tokens: %w", err)
	}

	return t, nil
}

// getToken will get a specific token
func (s *Store) getToken(ctx context.Context, t1 Token) (Token, error) {
	var t Token

	builder := utils.PSQL().Select("*").
		From("web_auth.api_tokens").
		Where(sq.Eq{"token_id": t1.TokenID}).
		Limit(1)

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for getToken: %w", err))
	}

	err = s.db.GetContext(ctx, &t, sql, args...)
	if err != nil {
		return t, fmt.Errorf("failed to get token: %w", err)
	}

	return t, nil
}

// addToken will add a token specified in the parameter
func (s *Store) addToken(ctx context.Context, t1 Token) (Token, error) {
	var t Token

	builder := utils.PSQL().Insert("web_auth.api_tokens").
		Columns("token_id", "name", "description", "expiry", "user_id").
		Values(t1.TokenID, t1.Name, t1.Description, t1.Expiry, t1.UserID).
		Suffix("RETURNING token_id, name, description, expiry, user_id")

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for addToken: %w", err))
	}

	stmt, err := s.db.PrepareContext(ctx, sql)
	if err != nil {
		return Token{}, fmt.Errorf("failed to add token: %w", err)
	}

	defer stmt.Close()

	err = stmt.QueryRow(args...).Scan(&t)
	if err != nil {
		return Token{}, fmt.Errorf("failed to add token: %w", err)
	}

	return t, nil
}

// deleteToken will delete a token and prevent it from being used
func (s *Store) deleteToken(ctx context.Context, t Token) error {
	builder := utils.PSQL().Delete("web_auth.api_tokens").
		Where(sq.Eq{"token_id": t.TokenID})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deletToken: %w", err))
	}

	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete token: %w", err)
	}

	return nil
}

// deleteOldToken will delete all expired tokens
func (s *Store) deleteOldToken(ctx context.Context) error {
	builder := utils.PSQL().Delete("web_auth.api_tokens").
		Where(sq.LtOrEq{"expiry": null.TimeFrom(time.Now())})

	sql, args, err := builder.ToSql()
	if err != nil {
		panic(fmt.Errorf("failed to build sql for deleteOldToken: %w", err))
	}

	_, err = s.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete old token: %w", err)
	}

	return nil
}
