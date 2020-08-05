package db

import (
	"context"
	"os"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ystv/web-auth/types"
)

var store Store

// Store interface, all functions offered by the db
type Store interface {
	VerifyUser(ctx context.Context, user *types.User) error
	UpdateUser(ctx context.Context, user *types.User) error
}

// DB is the connection pool
type DB struct {
	*pgxpool.Pool
}

// NewStore initialises the store
func NewStore(dataSourceName string) (*DB, error) {
	dbpool, err := pgxpool.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, err
	}
	return &DB{dbpool}, nil
}
